/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// aTLS provides config generation functions to bootstrap attested TLS connections.
package atls

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/crypto"
)

const attestationTimeout = 30 * time.Second

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
//
// ATTENTION: The tls.Config ensures freshness of the server's attestation only for the first connection it is used for.
// If freshness is required, you must create a new tls.Config for each connection or ensure freshness on the protocol level.
// If freshness is not required, you can reuse this tls.Config.
//
// If no validators are set, the server's attestation document will not be verified.
// If issuer is nil, the client will be unable to perform mutual aTLS.
func CreateAttestationClientTLSConfig(issuer Issuer, validators []Validator) (*tls.Config, error) {
	clientNonce, err := crypto.GenerateRandomBytes(crypto.RNGLengthDefault)
	if err != nil {
		return nil, err
	}
	clientConn := &clientConnection{
		issuer:      issuer,
		validators:  validators,
		clientNonce: clientNonce,
	}

	return &tls.Config{
		VerifyPeerCertificate: clientConn.verify,
		GetClientCertificate:  clientConn.getCertificate,                      // use custom certificate for mutual aTLS connections
		InsecureSkipVerify:    true,                                           // disable default verification because we use our own verify func
		ServerName:            base64.StdEncoding.EncodeToString(clientNonce), // abuse ServerName as a channel to transmit the nonce
		MinVersion:            tls.VersionTLS12,
	}, nil
}

// Issuer issues an attestation document.
type Issuer interface {
	variant.Getter
	Issue(ctx context.Context, userData []byte, nonce []byte) (quote []byte, err error)
}

// Validator is able to validate an attestation document.
type Validator interface {
	variant.Getter
	Validate(ctx context.Context, attDoc []byte, nonce []byte) ([]byte, error)
}

// getATLSConfigForClientFunc returns a config setup function that is called once for every client connecting to the server.
// This allows for different server configuration for every client.
// In aTLS this is used to generate unique nonces for every client.
func getATLSConfigForClientFunc(issuer Issuer, validators []Validator) (func(*tls.ClientHelloInfo) (*tls.Config, error), error) {
	// generate key for the server
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	// this function will be called once for every client
	return func(chi *tls.ClientHelloInfo) (*tls.Config, error) {
		// generate nonce for this connection
		serverNonce, err := crypto.GenerateRandomBytes(crypto.RNGLengthDefault)
		if err != nil {
			return nil, err
		}

		serverConn := &serverConnection{
			privKey:     priv,
			issuer:      issuer,
			validators:  validators,
			serverNonce: serverNonce,
		}

		cfg := &tls.Config{
			VerifyPeerCertificate: serverConn.verify,
			GetCertificate:        serverConn.getCertificate,
			MinVersion:            tls.VersionTLS12,
		}

		// enable mutual aTLS if any validators are set
		if len(validators) > 0 {
			cfg.ClientAuth = tls.RequireAnyClientCert // validity of certificate will be checked by our custom verify function

			// ugly hack: abuse acceptable client CAs as a channel to transmit the nonce
			if cfg.ClientCAs, err = encodeNonceToCertPool(serverNonce, priv); err != nil {
				return nil, fmt.Errorf("encode nonce: %w", err)
			}
		}

		return cfg, nil
	}, nil
}

// getCertificate creates a client or server certificate for aTLS connections.
// The certificate uses certificate extensions to embed an attestation document generated using nonce.
func getCertificate(ctx context.Context, issuer Issuer, priv, pub any, nonce []byte) (*tls.Certificate, error) {
	serialNumber, err := crypto.GenerateCertificateSerialNumber()
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
		attDoc, err := issuer.Issue(ctx, hash, nonce)
		if err != nil {
			return nil, err
		}

		extensions = append(extensions, pkix.Extension{Id: issuer.OID(), Value: attDoc})
	}

	// create certificate that includes the attestation document as extension
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
func processCertificate(rawCerts [][]byte, _ [][]*x509.Certificate) (*x509.Certificate, []byte, error) {
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
	var exts []string
	for _, ex := range cert.Extensions {
		for _, validator := range validators {
			if ex.Id.Equal(validator.OID()) {
				ctx, cancel := context.WithTimeout(context.Background(), attestationTimeout)
				defer cancel()

				userData, err := validator.Validate(ctx, ex.Value, nonce)
				if err != nil {
					return err
				}
				if !bytes.Equal(userData, hash) {
					return errors.New("certificate hash does not match user data")
				}
				return nil
			}
		}
		exts = append(exts, ex.Id.String())
	}

	return fmt.Errorf("certificate does not contain compatible attestation documents: got extension OIDs %#v", exts)
}

func hashPublicKey(pub any) ([]byte, error) {
	pubBytes, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return nil, err
	}
	result := sha256.Sum256(pubBytes)
	return result[:], nil
}

// encodeNonceToCertPool returns a cert pool that contains a certificate whose CN is the base64-encoded nonce.
func encodeNonceToCertPool(nonce []byte, privKey *ecdsa.PrivateKey) (*x509.CertPool, error) {
	template := &x509.Certificate{
		SerialNumber: &big.Int{},
		Subject:      pkix.Name{CommonName: base64.StdEncoding.EncodeToString(nonce)},
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &privKey.PublicKey, privKey)
	if err != nil {
		return nil, err
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		return nil, err
	}
	pool := x509.NewCertPool()
	pool.AddCert(cert)
	return pool, nil
}

// decodeNonceFromAcceptableCAs interprets the CN of acceptableCAs[0] as base64-encoded nonce and returns the decoded nonce.
// acceptableCAs should have been received by a client where the server used encodeNonceToCertPool to transmit the nonce.
func decodeNonceFromAcceptableCAs(acceptableCAs [][]byte) ([]byte, error) {
	if len(acceptableCAs) != 1 {
		return nil, errors.New("unexpected acceptableCAs length")
	}
	var rdnSeq pkix.RDNSequence
	if _, err := asn1.Unmarshal(acceptableCAs[0], &rdnSeq); err != nil {
		return nil, err
	}

	// https://github.com/golang/go/blob/19309779ac5e2f5a2fd3cbb34421dafb2855ac21/src/crypto/x509/pkix/pkix.go#L188
	oidCommonName := asn1.ObjectIdentifier{2, 5, 4, 3}

	for _, rdnSet := range rdnSeq {
		for _, rdn := range rdnSet {
			if rdn.Type.Equal(oidCommonName) {
				nonce, ok := rdn.Value.(string)
				if !ok {
					return nil, errors.New("unexpected RDN type")
				}
				return base64.StdEncoding.DecodeString(nonce)
			}
		}
	}

	return nil, errors.New("CN not found")
}

// clientConnection holds state for client to server connections.
type clientConnection struct {
	issuer      Issuer
	validators  []Validator
	clientNonce []byte
}

// verify the validity of an aTLS server certificate.
func (c *clientConnection) verify(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
	cert, hash, err := processCertificate(rawCerts, verifiedChains)
	if err != nil {
		return err
	}

	// don't perform verification of attestation document if no validators are set
	if len(c.validators) == 0 {
		return nil
	}

	return verifyEmbeddedReport(c.validators, cert, hash, c.clientNonce)
}

// getCertificate generates a client certificate for mutual aTLS connections.
func (c *clientConnection) getCertificate(cri *tls.CertificateRequestInfo) (*tls.Certificate, error) {
	// generate and hash key
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	// ugly hack: abuse acceptable client CAs as a channel to receive the nonce
	serverNonce, err := decodeNonceFromAcceptableCAs(cri.AcceptableCAs)
	if err != nil {
		return nil, fmt.Errorf("decode nonce: %w", err)
	}

	return getCertificate(cri.Context(), c.issuer, priv, &priv.PublicKey, serverNonce)
}

// serverConnection holds state for server to client connections.
type serverConnection struct {
	issuer      Issuer
	validators  []Validator
	privKey     *ecdsa.PrivateKey
	serverNonce []byte
}

// verify the validity of a clients aTLS certificate.
// Only needed for mutual aTLS.
func (c *serverConnection) verify(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
	cert, hash, err := processCertificate(rawCerts, verifiedChains)
	if err != nil {
		return err
	}

	return verifyEmbeddedReport(c.validators, cert, hash, c.serverNonce)
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
	return getCertificate(chi.Context(), c.issuer, c.privKey, &c.privKey.PublicKey, clientNonce)
}

// FakeIssuer fakes an issuer and can be used for tests.
type FakeIssuer struct {
	variant.Getter
}

// NewFakeIssuer creates a new FakeIssuer with the given OID.
func NewFakeIssuer(oid variant.Getter) *FakeIssuer {
	return &FakeIssuer{oid}
}

// Issue marshals the user data and returns it.
func (FakeIssuer) Issue(_ context.Context, userData []byte, nonce []byte) ([]byte, error) {
	return json.Marshal(FakeAttestationDoc{UserData: userData, Nonce: nonce})
}

// FakeValidator fakes a validator and can be used for tests.
type FakeValidator struct {
	variant.Getter
	err error // used for package internal testing only
}

// NewFakeValidator creates a new FakeValidator with the given OID.
func NewFakeValidator(oid variant.Getter) *FakeValidator {
	return &FakeValidator{oid, nil}
}

// NewFakeValidators returns a slice with a single FakeValidator.
func NewFakeValidators(oid variant.Getter) []Validator {
	return []Validator{NewFakeValidator(oid)}
}

// Validate unmarshals the attestation document and verifies the nonce.
func (v FakeValidator) Validate(_ context.Context, attDoc []byte, nonce []byte) ([]byte, error) {
	var doc FakeAttestationDoc
	if err := json.Unmarshal(attDoc, &doc); err != nil {
		return nil, err
	}

	if !bytes.Equal(doc.Nonce, nonce) {
		return nil, fmt.Errorf("invalid nonce: expected %x, got %x", doc.Nonce, nonce)
	}

	return doc.UserData, v.err
}

// FakeAttestationDoc is a fake attestation document used for testing.
type FakeAttestationDoc struct {
	UserData []byte
	Nonce    []byte
}

type fakeOID struct {
	asn1.ObjectIdentifier
}

func (o fakeOID) OID() asn1.ObjectIdentifier {
	return o.ObjectIdentifier
}
