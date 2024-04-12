/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package snp

import (
	"context"
	"crypto"
	"crypto/x509"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/attestation"
	"github.com/edgelesssys/constellation/v2/internal/attestation/vtpm"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/google/go-tpm-tools/proto/attest"
	"github.com/stretchr/testify/assert"
)

func TestGetTrustedKey(t *testing.T) {
	validator := func(ek []byte) *Validator {
		return &Validator{
			reportValidator: stubGCPValidator{},
			gceKeyGetter: func(_ context.Context, _ vtpm.AttestationDocument, _ []byte) (crypto.PublicKey, error) {
				return ek, nil
			},
		}
	}
	testCases := map[string]struct {
		akPub []byte
		ek    []byte
		info  []byte
	}{
		"success": {
			akPub: []byte("akPub"),
			ek:    []byte("ek"),
			info:  []byte("info"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			out, err := validator(tc.ek).getTrustedKey(
				context.Background(),
				vtpm.AttestationDocument{
					Attestation: &attest.Attestation{
						AkPub: tc.akPub,
					},
					InstanceInfo: tc.info,
				},
				nil,
			)

			assert.NoError(err)
			assert.Equal(tc.ek, out)
		})
	}
}

type stubGCPValidator struct{}

func (stubGCPValidator) validate(_ vtpm.AttestationDocument, _ *x509.Certificate, _ *x509.Certificate, _ [64]byte, _ *config.GCPSEVSNP, _ attestation.Logger) error {
	return nil
}
