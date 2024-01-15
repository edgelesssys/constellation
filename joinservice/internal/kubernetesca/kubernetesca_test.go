/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

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
	"log/slog"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
	kubeconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m, goleak.IgnoreAnyFunction("github.com/bazelbuild/rules_go/go/tools/bzltestutil.RegisterTimeoutHandler.func1"))
}

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
	defaultSigningRequestFunc := func() ([]byte, error) {
		privK, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return nil, err
		}
		csrTemplate := &x509.CertificateRequest{
			Subject: pkix.Name{
				Organization: []string{kubeconstants.NodesGroup},
				CommonName:   kubeconstants.NodesUserPrefix + "test-node",
			},
		}
		return x509.CreateCertificateRequest(rand.Reader, csrTemplate, privK)
	}

	testCases := map[string]struct {
		caCert               []byte
		caKey                []byte
		createSigningRequest func() ([]byte, error)
		wantErr              bool
	}{
		"success ec key": {
			caCert:               ecCert,
			caKey:                ecKey,
			createSigningRequest: defaultSigningRequestFunc,
		},
		"success rsa key": {
			caCert:               rsaCert,
			caKey:                rsaKey,
			createSigningRequest: defaultSigningRequestFunc,
		},
		"success any key": {
			caCert:               testCert,
			caKey:                testKey,
			createSigningRequest: defaultSigningRequestFunc,
		},
		"unsupported key": {
			caCert:               ecCert,
			caKey:                unsupportedKey,
			createSigningRequest: defaultSigningRequestFunc,
			wantErr:              true,
		},
		"invalid key": {
			caCert:               ecCert,
			caKey:                invalidKey,
			createSigningRequest: defaultSigningRequestFunc,
			wantErr:              true,
		},
		"invalid certificate": {
			caCert:               invalidCert,
			caKey:                ecKey,
			createSigningRequest: defaultSigningRequestFunc,
			wantErr:              true,
		},
		"no ca certificate": {
			caKey:                ecKey,
			createSigningRequest: defaultSigningRequestFunc,
			wantErr:              true,
		},
		"no ca key": {
			caCert:               ecCert,
			createSigningRequest: defaultSigningRequestFunc,
			wantErr:              true,
		},
		"no signing request": {
			caCert:               ecCert,
			caKey:                ecKey,
			createSigningRequest: func() ([]byte, error) { return nil, nil },
			wantErr:              true,
		},
		"incorrect common name format": {
			caCert: ecCert,
			caKey:  ecKey,
			createSigningRequest: func() ([]byte, error) {
				privK, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
				if err != nil {
					return nil, err
				}
				csrTemplate := &x509.CertificateRequest{
					Subject: pkix.Name{
						Organization: []string{kubeconstants.NodesGroup},
						CommonName:   "test-node",
					},
				}
				return x509.CreateCertificateRequest(rand.Reader, csrTemplate, privK)
			},
			wantErr: true,
		},
		"incorrect organization format": {
			caCert: ecCert,
			caKey:  ecKey,
			createSigningRequest: func() ([]byte, error) {
				privK, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
				if err != nil {
					return nil, err
				}
				csrTemplate := &x509.CertificateRequest{
					Subject: pkix.Name{
						Organization: []string{"test"},
						CommonName:   kubeconstants.NodesUserPrefix + "test-node",
					},
				}
				return x509.CreateCertificateRequest(rand.Reader, csrTemplate, privK)
			},
			wantErr: true,
		},
		"no organization": {
			caCert: ecCert,
			caKey:  ecKey,
			createSigningRequest: func() ([]byte, error) {
				privK, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
				if err != nil {
					return nil, err
				}
				csrTemplate := &x509.CertificateRequest{
					Subject: pkix.Name{
						CommonName: kubeconstants.NodesUserPrefix + "test-node",
					},
				}
				return x509.CreateCertificateRequest(rand.Reader, csrTemplate, privK)
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			fileHandler := file.NewHandler(afero.NewMemMapFs())

			if len(tc.caCert) > 0 {
				require.NoError(fileHandler.Write(caCertFilename, tc.caCert, file.OptNone))
			}
			if len(tc.caKey) > 0 {
				require.NoError(fileHandler.Write(caKeyFilename, tc.caKey, file.OptNone))
			}

			ca := New(
				slog.New(slog.NewTextHandler(logger.TestWriter{T: t}, nil)),
				fileHandler,
			)

			signingRequest, err := tc.createSigningRequest()
			require.NoError(err)
			kubeCert, err := ca.GetCertificate(signingRequest)
			if tc.wantErr {
				assert.Error(err)
				return
			}

			assert.NoError(err)
			certPEM, _ := pem.Decode(kubeCert)
			require.NotNil(certPEM)
			cert, err := x509.ParseCertificate(certPEM.Bytes)
			require.NoError(err)
			assert.True(strings.HasPrefix(cert.Subject.CommonName, kubeconstants.NodesUserPrefix))
			assert.Equal(kubeconstants.NodesGroup, cert.Subject.Organization[0])
			assert.Equal(x509.KeyUsageDigitalSignature|x509.KeyUsageKeyEncipherment, cert.KeyUsage)
			assert.Equal(x509.ExtKeyUsageClientAuth, cert.ExtKeyUsage[0])
			assert.False(cert.IsCA)
			assert.True(cert.BasicConstraintsValid)
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
