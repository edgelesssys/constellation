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
func CreateAttestationServerTLSConfig(issuer Issuer) (*tls.Config, error) {
	// generate and hash key
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	hash, err := hashPublicKey(&priv.PublicKey)
	if err != nil {
		return nil, err
	}

	getCertificate := func(chi *tls.ClientHelloInfo) (*tls.Certificate, error) {
		serialNumber, err := util.GenerateCertificateSerialNumber()
		if err != nil {
			return nil, err
		}

		// abuse ServerName as a channel to receive the nonce
		nonce, err := base64.StdEncoding.DecodeString(chi.ServerName)
		if err != nil {
			return nil, err
		}

		attDoc, err := issuer.Issue(hash, nonce)
		if err != nil {
			return nil, err
		}

		// create certficate that includes the attestation document as extension
		now := time.Now()
		template := &x509.Certificate{
			SerialNumber:    serialNumber,
			Subject:         pkix.Name{CommonName: "Constellation"},
			NotBefore:       now.Add(-2 * time.Hour),
			NotAfter:        now.Add(2 * time.Hour),
			ExtraExtensions: []pkix.Extension{{Id: issuer.OID(), Value: attDoc}},
		}
		cert, err := x509.CreateCertificate(rand.Reader, template, template, &priv.PublicKey, priv)
		if err != nil {
			return nil, err
		}

		return &tls.Certificate{Certificate: [][]byte{cert}, PrivateKey: priv}, nil
	}

	return &tls.Config{GetCertificate: getCertificate, MinVersion: tls.VersionTLS12}, nil
}

// CreateAttestationClientTLSConfig creates a tls.Config object that verifies a certificate with an embedded attestation document.
func CreateAttestationClientTLSConfig(validators []Validator) (*tls.Config, error) {
	nonce, err := util.GenerateRandomBytes(config.RNGLengthDefault)
	if err != nil {
		return nil, err
	}

	verify := func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
		// parse certificate
		if len(rawCerts) == 0 {
			return errors.New("rawCerts is empty")
		}
		cert, err := x509.ParseCertificate(rawCerts[0])
		if err != nil {
			return err
		}

		// verify self-signed certificate
		roots := x509.NewCertPool()
		roots.AddCert(cert)
		_, err = cert.Verify(x509.VerifyOptions{Roots: roots})
		if err != nil {
			return err
		}

		hash, err := hashPublicKey(cert.PublicKey)
		if err != nil {
			return err
		}

		// verify embedded report
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

	return &tls.Config{
		VerifyPeerCertificate: verify,
		InsecureSkipVerify:    true,                                     // disable default verification because we use our own verify func
		ServerName:            base64.StdEncoding.EncodeToString(nonce), // abuse ServerName as a channel to transmit the nonce
		MinVersion:            tls.VersionTLS12,
	}, nil
}

// CreateUnverifiedClientTLSConfig creates a tls.Config object that skips verification of a certificate with an embedded attestation document.
func CreateUnverifiedClientTLSConfig() (*tls.Config, error) {
	nonce, err := util.GenerateRandomBytes(config.RNGLengthDefault)
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		InsecureSkipVerify: true,                                     // disable certificate verification
		ServerName:         base64.StdEncoding.EncodeToString(nonce), // abuse ServerName as a channel to transmit the nonce
		MinVersion:         tls.VersionTLS12,
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

func hashPublicKey(pub any) ([]byte, error) {
	pubBytes, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return nil, err
	}
	result := sha256.Sum256(pubBytes)
	return result[:], nil
}
