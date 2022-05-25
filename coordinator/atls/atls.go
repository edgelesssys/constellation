package atls

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"errors"
	"time"

	"github.com/edgelesssys/constellation/coordinator/config"
	"github.com/edgelesssys/constellation/coordinator/oid"
	"github.com/edgelesssys/constellation/coordinator/util"
)

// CreateAttestationServerTLSConfig creates a tls.Config object with a self-signed certificate and an embedded attestation document.
// Pass a list of validators to enable mutual aTLS.
// If issuer is nil, no attestation will be embedded.
func CreateAttestationServerTLSConfig(issuer Issuer, validators []Validator) (*tls.Config, error) {
	getConfigForClient, err := getATLSConfigForClientFunc(issuer, validators)
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		GetConfigForClient: getConfigForClient,
	}, nil
}

// CreateAttestationClientTLSConfig creates a tls.Config object that verifies a certificate with an embedded attestation document.
// If no validators are set, the server's attestation document will not be verified.
// If issuer is nil, the client will be unable to perform mutual aTLS.
func CreateAttestationClientTLSConfig(issuer Issuer, validators []Validator) (*tls.Config, error) {
	nonce, err := util.GenerateRandomBytes(config.RNGLengthDefault)
	if err != nil {
		return nil, err
	}
	clientConn := &clientConnection{
		issuer:      issuer,
		validators:  validators,
		clientNonce: nonce,
	}

	return &tls.Config{
		VerifyPeerCertificate: clientConn.verify,
		GetClientCertificate:  clientConn.getCertificate,                // use custom certificate for mutual aTLS connections
		InsecureSkipVerify:    true,                                     // disable default verification because we use our own verify func
		ServerName:            base64.StdEncoding.EncodeToString(nonce), // abuse ServerName as a channel to transmit the nonce
		MinVersion:            tls.VersionTLS12,
	}, nil
}

type Issuer interface {
	oid.Getter
	Issue(userData []byte, nonce []byte) (quote []byte, err error)
}

type Validator interface {
	oid.Getter
	Validate(attDoc []byte, nonce []byte) ([]byte, error)
}

// getATLSConfigForClientFunc returns a config setup function that is called once for every client connecting to the server.
// This allows for different server configuration for every client.
// In aTLS this is used to generate unique nonces for every client and embed them in the server's certificate.
func getATLSConfigForClientFunc(issuer Issuer, validators []Validator) (func(*tls.ClientHelloInfo) (*tls.Config, error), error) {
	// generate key for the server
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	// this function will be called once for every client
	return func(chi *tls.ClientHelloInfo) (*tls.Config, error) {
		// generate nonce for this connection
		nonce, err := util.GenerateRandomBytes(config.RNGLengthDefault)
		if err != nil {
			return nil, err
		}

		serverConn := &serverConnection{
			privKey:    priv,
			issuer:     issuer,
			validators: validators,
			nonce:      nonce,
		}

		clientAuth := tls.NoClientCert
		// enable mutual aTLS if any validators are set
		if len(validators) > 0 {
			clientAuth = tls.RequireAnyClientCert // validity of certificate will be checked by our custom verify function
		}

		return &tls.Config{
			ClientAuth:            clientAuth,
			VerifyPeerCertificate: serverConn.verify,
			GetCertificate:        serverConn.getCertificate,
			MinVersion:            tls.VersionTLS12,
		}, nil
	}, nil
}

// getCertificate creates a client or server certificate for aTLS connections.
// The certificate uses certificate extensions to embed an attestation document generated using remoteNonce.
// If localNonce is set, it is also embedded as a certificate extension.
func getCertificate(issuer Issuer, priv, pub any, remoteNonce, localNonce []byte) (*tls.Certificate, error) {
	serialNumber, err := util.GenerateCertificateSerialNumber()
	if err != nil {
		return nil, err
	}

	var extensions []pkix.Extension

	// create and embed attestation if quote Issuer is available
	if issuer != nil {
		hash, err := hashPublicKey(pub)
		if err != nil {
			return nil, err
		}

		// create attestation document using the nonce send by the remote party
		attDoc, err := issuer.Issue(hash, remoteNonce)
		if err != nil {
			return nil, err
		}

		extensions = append(extensions, pkix.Extension{Id: issuer.OID(), Value: attDoc})
	}

	// embed locally generated nonce in certificate
	if len(localNonce) > 0 {
		extensions = append(extensions, pkix.Extension{Id: oid.ATLSNonce, Value: localNonce})
	}

	// create certificate that includes the attestation document and the server nonce as extension
	now := time.Now()
	template := &x509.Certificate{
		SerialNumber:    serialNumber,
		Subject:         pkix.Name{CommonName: "Constellation"},
		NotBefore:       now.Add(-2 * time.Hour),
		NotAfter:        now.Add(2 * time.Hour),
		ExtraExtensions: extensions,
	}
	cert, err := x509.CreateCertificate(rand.Reader, template, template, pub, priv)
	if err != nil {
		return nil, err
	}

	return &tls.Certificate{Certificate: [][]byte{cert}, PrivateKey: priv}, nil
}

// processCertificate parses the certificate and verifies it.
// If successful returns the certificate and its hashed public key, an error otherwise.
func processCertificate(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) (*x509.Certificate, []byte, error) {
	// parse certificate
	if len(rawCerts) == 0 {
		return nil, nil, errors.New("rawCerts is empty")
	}
	cert, err := x509.ParseCertificate(rawCerts[0])
	if err != nil {
		return nil, nil, err
	}

	// verify self-signed certificate
	roots := x509.NewCertPool()
	roots.AddCert(cert)
	_, err = cert.Verify(x509.VerifyOptions{Roots: roots})
	if err != nil {
		return nil, nil, err
	}

	// hash of certificates public key is used as userData in the embedded attestation document
	hash, err := hashPublicKey(cert.PublicKey)
	return cert, hash, err
}

// verifyEmbeddedReport verifies an aTLS certificate by validating the attestation document embedded in the TLS certificate.
func verifyEmbeddedReport(validators []Validator, cert *x509.Certificate, hash, nonce []byte) error {
	for _, ex := range cert.Extensions {
		for _, validator := range validators {
			if ex.Id.Equal(validator.OID()) {
				userData, err := validator.Validate(ex.Value, nonce)
				if err != nil {
					return err
				}
				if !bytes.Equal(userData, hash) {
					return errors.New("certificate hash does not match user data")
				}
				return nil
			}
		}
	}

	return errors.New("certificate does not contain attestation document")
}

func hashPublicKey(pub any) ([]byte, error) {
	pubBytes, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return nil, err
	}
	result := sha256.Sum256(pubBytes)
	return result[:], nil
}

// clientConnection holds state for client to server connections.
type clientConnection struct {
	issuer      Issuer
	validators  []Validator
	clientNonce []byte
	serverNonce []byte
}

// verify the validity of an aTLS server certificate.
func (c *clientConnection) verify(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
	cert, hash, err := processCertificate(rawCerts, verifiedChains)
	if err != nil {
		return err
	}

	// get nonce send by server from cert extensions and save to connection state
	for _, ex := range cert.Extensions {
		if ex.Id.Equal(oid.ATLSNonce) {
			c.serverNonce = ex.Value
		}
	}

	// don't perform verification of attestation document if no validators are set
	if len(c.validators) == 0 {
		return nil
	}

	return verifyEmbeddedReport(c.validators, cert, hash, c.clientNonce)
}

// getCertificate generates a client certificate for mutual aTLS connections.
func (c *clientConnection) getCertificate(*tls.CertificateRequestInfo) (*tls.Certificate, error) {
	// generate and hash key
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	// create aTLS certificate using the server's nonce as read by clientConnection.verify
	// we do not pass a nonce because
	// 		1. we already received a certificate from the server
	//		2. we transmitted the client nonce as our server name in our client-hello message
	return getCertificate(c.issuer, priv, &priv.PublicKey, c.serverNonce, nil)
}

// serverConnection holds state for server to client connections.
type serverConnection struct {
	issuer     Issuer
	validators []Validator
	privKey    *ecdsa.PrivateKey
	nonce      []byte
}

// verify the validity of a clients aTLS certificate.
// Only needed for mutual aTLS.
func (c *serverConnection) verify(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
	cert, hash, err := processCertificate(rawCerts, verifiedChains)
	if err != nil {
		return err
	}

	return verifyEmbeddedReport(c.validators, cert, hash, c.nonce)
}

// getCertificate generates a client certificate for aTLS connections.
// Can be used for mutual as well as basic aTLS.
func (c *serverConnection) getCertificate(chi *tls.ClientHelloInfo) (*tls.Certificate, error) {
	// abuse ServerName as a channel to receive the nonce
	clientNonce, err := base64.StdEncoding.DecodeString(chi.ServerName)
	if err != nil {
		return nil, err
	}

	// create aTLS certificate using the nonce as extracted from the client-hello message
	// we also embed the nonce generated for this connection in case of mutual aTLS
	return getCertificate(c.issuer, c.privKey, &c.privKey.PublicKey, clientNonce, c.nonce)
}
