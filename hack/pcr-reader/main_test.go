package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"testing"

	"github.com/edgelesssys/constellation/internal/attestation/vtpm"
	"github.com/edgelesssys/constellation/internal/oid"
	"github.com/google/go-tpm-tools/proto/attest"
	"github.com/google/go-tpm-tools/proto/tpm"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetVerifyPeerCertificateFunc(t *testing.T) {
	testCases := map[string]struct {
		rawCerts [][]byte
		wantErr  bool
	}{
		"no certificates": {
			rawCerts: nil,
			wantErr:  true,
		},
		"invalid certificate": {
			rawCerts: [][]byte{
				{0x1, 0x2, 0x3},
			},
			wantErr: true,
		},
		"no extension": {
			rawCerts: [][]byte{
				mustGenerateTestCert(t, &x509.Certificate{
					SerialNumber: big.NewInt(123),
				}),
			},
			wantErr: true,
		},
		"certificate with attestation": {
			rawCerts: [][]byte{
				mustGenerateTestCert(t, &x509.Certificate{
					SerialNumber: big.NewInt(123),
					ExtraExtensions: []pkix.Extension{
						{
							Id:       oid.GCP{}.OID(),
							Value:    []byte{0x1, 0x2, 0x3},
							Critical: true,
						},
					},
				}),
			},
			wantErr: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			attDoc := &[]byte{}
			verify := getVerifyPeerCertificateFunc(attDoc)

			err := verify(tc.rawCerts, nil)
			if tc.wantErr {
				assert.Error(err)
			} else {
				require.NoError(err)

				assert.NotNil(attDoc)
				cert, err := x509.ParseCertificate(tc.rawCerts[0])
				require.NoError(err)
				assert.Equal(cert.Extensions[0].Value, *attDoc)
			}
		})
	}
}

func mustGenerateTestCert(t *testing.T, template *x509.Certificate) []byte {
	require := require.New(t)
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(err)
	cert, err := x509.CreateCertificate(rand.Reader, template, template, priv.Public(), priv)
	require.NoError(err)
	return cert
}

func TestExportToFile(t *testing.T) {
	testCases := map[string]struct {
		pcrs    map[uint32][]byte
		fs      *afero.Afero
		wantErr bool
	}{
		"file not writeable": {
			pcrs: map[uint32][]byte{
				0: {0x1, 0x2, 0x3},
				1: {0x1, 0x2, 0x3},
				2: {0x1, 0x2, 0x3},
			},
			fs:      &afero.Afero{Fs: afero.NewReadOnlyFs(afero.NewMemMapFs())},
			wantErr: true,
		},
		"file writeable": {
			pcrs: map[uint32][]byte{
				0: {0x1, 0x2, 0x3},
				1: {0x1, 0x2, 0x3},
				2: {0x1, 0x2, 0x3},
			},
			fs:      &afero.Afero{Fs: afero.NewMemMapFs()},
			wantErr: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			path := "test-file"
			err := exportToFile(path, tc.pcrs, tc.fs)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				content, err := tc.fs.ReadFile(path)
				require.NoError(err)

				for _, pcr := range tc.pcrs {
					for _, register := range pcr {
						assert.Contains(string(content), fmt.Sprintf("%#02X", register))
					}
				}
			}
		})
	}
}

func TestValidatePCRAttDoc(t *testing.T) {
	testCases := map[string]struct {
		attDocRaw []byte
		wantErr   bool
	}{
		"invalid attestation document": {
			attDocRaw: []byte{0x1, 0x2, 0x3},
			wantErr:   true,
		},
		"nil attestation": {
			attDocRaw: mustMarshalAttDoc(t, vtpm.AttestationDocument{}),
			wantErr:   true,
		},
		"nil quotes": {
			attDocRaw: mustMarshalAttDoc(t, vtpm.AttestationDocument{
				Attestation: &attest.Attestation{},
			}),
			wantErr: true,
		},
		"invalid PCRs": {
			attDocRaw: mustMarshalAttDoc(t, vtpm.AttestationDocument{
				Attestation: &attest.Attestation{
					Quotes: []*tpm.Quote{
						{
							Pcrs: &tpm.PCRs{
								Hash: tpm.HashAlgo_SHA256,
								Pcrs: map[uint32][]byte{
									0: {0x1, 0x2, 0x3},
								},
							},
						},
					},
				},
			}),
			wantErr: true,
		},
		"valid PCRs": {
			attDocRaw: mustMarshalAttDoc(t, vtpm.AttestationDocument{
				Attestation: &attest.Attestation{
					Quotes: []*tpm.Quote{
						{
							Pcrs: &tpm.PCRs{
								Hash: tpm.HashAlgo_SHA256,
								Pcrs: map[uint32][]byte{
									0: []byte("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"),
								},
							},
						},
					},
				},
			}),
			wantErr: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			pcrs, err := validatePCRAttDoc(tc.attDocRaw)
			if tc.wantErr {
				assert.Error(err)
			} else {
				require.NoError(err)

				attDoc := vtpm.AttestationDocument{}
				require.NoError(json.Unmarshal(tc.attDocRaw, &attDoc))
				qIdx, err := vtpm.GetSHA256QuoteIndex(attDoc.Attestation.Quotes)
				require.NoError(err)
				assert.EqualValues(attDoc.Attestation.Quotes[qIdx].Pcrs.Pcrs, pcrs)
			}
		})
	}
}

func mustMarshalAttDoc(t *testing.T, attDoc vtpm.AttestationDocument) []byte {
	attDocRaw, err := json.Marshal(attDoc)
	require.NoError(t, err)
	return attDocRaw
}

func TestPrintPCRs(t *testing.T) {
	testCases := map[string]struct {
		pcrs   map[uint32][]byte
		format string
	}{
		"json": {
			pcrs: map[uint32][]byte{
				0: {0x1, 0x2, 0x3},
				1: {0x1, 0x2, 0x3},
				2: {0x1, 0x2, 0x3},
			},
			format: "json",
		},
		"empty format": {
			pcrs: map[uint32][]byte{
				0: {0x1, 0x2, 0x3},
				1: {0x1, 0x2, 0x3},
				2: {0x1, 0x2, 0x3},
			},
			format: "",
		},
		"yaml": {
			pcrs: map[uint32][]byte{
				0: {0x1, 0x2, 0x3},
				1: {0x1, 0x2, 0x3},
				2: {0x1, 0x2, 0x3},
			},
			format: "yaml",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			var out bytes.Buffer
			err := printPCRs(&out, tc.pcrs, tc.format)
			assert.NoError(err)

			for idx, pcr := range tc.pcrs {
				assert.Contains(out.String(), fmt.Sprintf("%d", idx))
				assert.Contains(out.String(), base64.StdEncoding.EncodeToString(pcr))
			}
		})
	}
}
