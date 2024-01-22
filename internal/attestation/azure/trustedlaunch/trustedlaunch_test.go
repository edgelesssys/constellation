/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package trustedlaunch

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/attestation/simulator"
	"github.com/edgelesssys/constellation/v2/internal/attestation/vtpm"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/crypto"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	tpmclient "github.com/google/go-tpm-tools/client"
	"github.com/google/go-tpm-tools/proto/attest"
	"github.com/google/go-tpm/legacy/tpm2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAttestationCert(t *testing.T) {
	cgo := os.Getenv("CGO_ENABLED")
	if cgo == "0" {
		t.Skip("skipping test because CGO is disabled and tpm simulator requires it")
	}
	require := require.New(t)
	tpm, err := simulator.OpenSimulatedTPM()
	require.NoError(err)
	defer tpm.Close()

	// create key in TPM
	tpmAk, err := tpmclient.NewCachedKey(tpm, tpm2.HandleOwner, tpm2.Public{
		Type:       tpm2.AlgRSA,
		NameAlg:    tpm2.AlgSHA256,
		Attributes: tpm2.FlagFixedTPM | tpm2.FlagFixedParent | tpm2.FlagSensitiveDataOrigin | tpm2.FlagUserWithAuth | tpm2.FlagNoDA | tpm2.FlagRestricted | tpm2.FlagSign,
		RSAParameters: &tpm2.RSAParams{
			Sign: &tpm2.SigScheme{
				Alg:  tpm2.AlgRSASSA,
				Hash: tpm2.AlgSHA256,
			},
			KeyBits: 2048,
		},
	}, tpmAkIdx)
	require.NoError(err)
	defer tpmAk.Close()
	akPub, err := tpmAk.PublicArea().Encode()
	require.NoError(err)

	// root certificate
	rootKey, rootTemplate := fillCertTemplate(t, &x509.Certificate{
		Subject:               pkix.Name{CommonName: "root CA"},
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
	})
	rootCert := newTestCert(t, rootTemplate, rootTemplate, rootKey.Public(), rootKey)

	// intermediate certificate
	intermediateKey, intermediateTemplate := fillCertTemplate(t, &x509.Certificate{
		Subject:               pkix.Name{CommonName: "intermediate CA"},
		Issuer:                rootTemplate.Subject,
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
	})
	intermediateCert := newTestCert(t, intermediateTemplate, rootTemplate, intermediateKey.Public(), rootKey)

	// define NV index once to avoid the need for fancy error handling later
	require.NoError(tpm2.NVDefineSpace(
		tpm, tpm2.HandleOwner, tpmAkCertIdx, "", "", []byte{},
		tpm2.AttrOwnerWrite|tpm2.AttrOwnerRead|tpm2.AttrAuthRead|tpm2.AttrAuthWrite|tpm2.AttrNoDA, 1,
	))

	defaultAkCertFunc := func(*testing.T) *x509.Certificate {
		t.Helper()
		_, certTemplate := fillCertTemplate(t, &x509.Certificate{
			IssuingCertificateURL: []string{
				"192.0.2.1/ca.crt",
			},
			Subject: pkix.Name{CommonName: "AK Certificate"},
			Issuer:  intermediateCert.Subject,
		})
		return newTestCert(t, certTemplate, intermediateCert, tpmAk.PublicKey(), intermediateKey)
	}

	testCases := map[string]struct {
		crlServer       roundTripFunc
		getAkCert       func(*testing.T) *x509.Certificate
		wantIssueErr    bool
		wantValidateErr bool
	}{
		"success": {
			crlServer: func(req *http.Request) *http.Response {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader(intermediateCert.Raw)),
				}
			},
			getAkCert: defaultAkCertFunc,
		},
		"intermediate cert is fetched from multiple URLs": {
			crlServer: func(req *http.Request) *http.Response {
				if req.URL.String() == "192.0.2.1/ca.crt" {
					return &http.Response{StatusCode: http.StatusNotFound}
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader(intermediateCert.Raw)),
				}
			},
			getAkCert: func(*testing.T) *x509.Certificate {
				t.Helper()
				_, certTemplate := fillCertTemplate(t, &x509.Certificate{
					IssuingCertificateURL: []string{
						"192.0.2.1/ca.crt",
						"192.0.2.2/ca.crt",
					},
					Subject: pkix.Name{CommonName: "AK Certificate"},
					Issuer:  intermediateCert.Subject,
				})
				return newTestCert(t, certTemplate, intermediateCert, tpmAk.PublicKey(), intermediateKey)
			},
		},
		"intermediate cert cannot be fetched": {
			crlServer: func(req *http.Request) *http.Response {
				return &http.Response{StatusCode: http.StatusNotFound}
			},
			getAkCert:    defaultAkCertFunc,
			wantIssueErr: true,
		},
		"intermediate cert is not signed by root cert": {
			crlServer: func(req *http.Request) *http.Response {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader(rootCert.Raw)),
				}
			},
			getAkCert:       defaultAkCertFunc,
			wantValidateErr: true,
		},
		"ak does not match ak cert public key": {
			crlServer: func(req *http.Request) *http.Response {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader(intermediateCert.Raw)),
				}
			},
			getAkCert: func(*testing.T) *x509.Certificate {
				t.Helper()
				key, certTemplate := fillCertTemplate(t, &x509.Certificate{
					IssuingCertificateURL: []string{
						"192.0.2.1/ca.crt",
					},
					Subject: pkix.Name{CommonName: "AK Certificate"},
					Issuer:  intermediateCert.Subject,
				})
				return newTestCert(t, certTemplate, intermediateCert, key.Public(), intermediateKey)
			},
			wantValidateErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			akCert := tc.getAkCert(t).Raw

			// write akCert to TPM
			require.NoError(tpm2.NVUndefineSpace(tpm, "", tpm2.HandleOwner, tpmAkCertIdx))
			require.NoError(tpm2.NVDefineSpace(
				tpm, tpm2.HandleOwner, tpmAkCertIdx, "", "", []byte{},
				tpm2.AttrOwnerWrite|tpm2.AttrOwnerRead|tpm2.AttrAuthRead|tpm2.AttrAuthWrite|tpm2.AttrNoDA,
				uint16(len(akCert)),
			))
			require.NoError(tpm2.NVWrite(tpm, tpm2.HandleOwner, tpmAkCertIdx, "", akCert, 0))

			issuer := NewIssuer(logger.NewTest(t))
			issuer.hClient = newTestClient(tc.crlServer)

			certs, err := issuer.getAttestationCert(context.Background(), tpm, nil)
			if tc.wantIssueErr {
				assert.Error(err)
				return
			}
			require.NoError(err)

			attDoc := vtpm.AttestationDocument{
				InstanceInfo: certs,
				Attestation: &attest.Attestation{
					AkPub: akPub,
				},
			}

			validator := NewValidator(&config.AzureTrustedLaunch{Measurements: measurements.M{}}, nil)
			cert, err := x509.ParseCertificate(rootCert.Raw)
			require.NoError(err)
			roots := x509.NewCertPool()
			roots.AddCert(cert)
			validator.roots = roots

			key, err := validator.verifyAttestationKey(context.Background(), attDoc, nil)
			if tc.wantValidateErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			rsaKey, ok := key.(*rsa.PublicKey)
			require.True(ok)
			assert.True(rsaKey.Equal(tpmAk.PublicKey()))
		})
	}
}

type roundTripFunc func(req *http.Request) *http.Response

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

// newTestClient returns *http.Client with Transport replaced to avoid making real calls.
func newTestClient(fn roundTripFunc) *http.Client {
	return &http.Client{
		Transport: fn,
	}
}

func newTestCert(t *testing.T, template *x509.Certificate, parent *x509.Certificate, pub, priv any) *x509.Certificate {
	t.Helper()
	require := require.New(t)

	certDER, err := x509.CreateCertificate(rand.Reader, template, parent, pub, priv)
	require.NoError(err)
	cert, err := x509.ParseCertificate(certDER)
	require.NoError(err)
	return cert
}

func fillCertTemplate(t *testing.T, template *x509.Certificate) (*rsa.PrivateKey, *x509.Certificate) {
	t.Helper()
	require := require.New(t)

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(err, "generating root key failed")

	serialNumber, err := crypto.GenerateCertificateSerialNumber()
	require.NoError(err)
	now := time.Now()

	template.SerialNumber = serialNumber
	template.NotBefore = now.Add(-2 * time.Hour)
	template.NotAfter = now.Add(24 * 365 * time.Hour)
	return key, template
}
