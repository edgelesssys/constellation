package kubernetesca

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/internal/file"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetCertificate(t *testing.T) {
	ecCert, ecKey := mustCreateCert(mustCreateECKey)
	rsaCert, rsaKey := mustCreateCert(mustCreateRSAKey)
	testCert, testKey := mustCreateCert(mustCreatePKCS8Key)
	unsupportedKey := []byte(`-----BEGIN SOME KEY-----
Q29uc3RlbGxhdGlvbg==
-----END SOME KEY-----`)
	invalidKey := []byte(`-----BEGIN PRIVATE KEY-----
Q29uc3RlbGxhdGlvbg==
-----END PRIVATE KEY-----`)
	invalidCert := []byte(`-----BEGIN CERTIFICATE-----
Q29uc3RlbGxhdGlvbg==
-----END CERTIFICATE-----`)

	testCases := map[string]struct {
		caCert  []byte
		caKey   []byte
		wantErr bool
	}{
		"success ec key": {
			caCert: ecCert,
			caKey:  ecKey,
		},
		"success rsa key": {
			caCert: rsaCert,
			caKey:  rsaKey,
		},
		"success any key": {
			caCert: testCert,
			caKey:  testKey,
		},
		"unsupported key": {
			caCert:  ecCert,
			caKey:   unsupportedKey,
			wantErr: true,
		},
		"invalid key": {
			caCert:  ecCert,
			caKey:   invalidKey,
			wantErr: true,
		},
		"invalid certificate": {
			caCert:  invalidCert,
			caKey:   ecKey,
			wantErr: true,
		},
		"no ca certificate": {
			caKey:   ecKey,
			wantErr: true,
		},
		"no ca key": {
			caCert:  ecCert,
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			file := file.NewHandler(afero.NewMemMapFs())

			if len(tc.caCert) > 0 {
				require.NoError(file.Write(caCertFilename, tc.caCert, 0o644))
			}
			if len(tc.caKey) > 0 {
				require.NoError(file.Write(caKeyFilename, tc.caKey, 0o644))
			}

			ca := New(file)

			nodeName := "test"
			kubeCert, kubeKey, err := ca.GetCertificate(nodeName)
			if tc.wantErr {
				assert.Error(err)
				return
			}

			assert.NoError(err)
			certPEM, _ := pem.Decode(kubeCert)
			require.NotNil(certPEM)
			cert, err := x509.ParseCertificate(certPEM.Bytes)
			require.NoError(err)
			assert.Equal("system:node:"+nodeName, cert.Subject.CommonName)
			assert.Equal("system:nodes", cert.Subject.Organization[0])
			assert.Equal(x509.KeyUsageDigitalSignature|x509.KeyUsageKeyEncipherment, cert.KeyUsage)
			assert.Equal(x509.ExtKeyUsageClientAuth, cert.ExtKeyUsage[0])
			assert.False(cert.IsCA)
			assert.True(cert.BasicConstraintsValid)

			keyPEM, _ := pem.Decode(kubeKey)
			require.NotNil(keyPEM)
			key, err := x509.ParseECPrivateKey(keyPEM.Bytes)
			require.NoError(err)
			require.IsType(&ecdsa.PublicKey{}, cert.PublicKey)
			assert.Equal(&key.PublicKey, cert.PublicKey.(*ecdsa.PublicKey))
		})
	}
}

func mustCreateCert(getKey func() (crypto.PrivateKey, []byte)) ([]byte, []byte) {
	caPriv, keyPEM := getKey()
	caTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "kubernetes",
		},
		NotBefore: time.Now().Add(-2 * time.Hour),
		IsCA:      true,
		KeyUsage:  x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	}
	caCert, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, publicKey(caPriv), caPriv)
	if err != nil {
		panic(err)
	}
	caCertPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caCert,
	})

	return caCertPEM, keyPEM
}

func mustCreateECKey() (crypto.PrivateKey, []byte) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(err)
	}
	keyBytes, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		panic(err)
	}
	return key, pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: keyBytes,
	})
}

func mustCreatePKCS8Key() (crypto.PrivateKey, []byte) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(err)
	}
	keyBytes, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		panic(err)
	}
	return key, pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: keyBytes,
	})
}

func mustCreateRSAKey() (crypto.PrivateKey, []byte) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	keyBytes := x509.MarshalPKCS1PrivateKey(key)
	return key, pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: keyBytes,
	})
}

func publicKey(priv crypto.PrivateKey) any {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	default:
		return nil
	}
}
