/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package snp

import (
	"bytes"
	"context"
	"crypto"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/attestation"
	"github.com/edgelesssys/constellation/v2/internal/attestation/vtpm"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/google/go-tpm-tools/proto/attest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestSha512sum(t *testing.T) {
	testCases := map[string]struct {
		key   string
		hash  string
		match bool
	}{
		"success": {
			// Generated using: rsa.GenerateKey(rand.Reader, 1024).
			key:   "30819f300d06092a864886f70d010101050003818d0030818902818100d4b2f072a32fa98456eb7f5938e2ff361fb64d698ea91e003d34bfc5374b814c16ba9ae3ec392ef6d48cf79b63067e338aa941219a7bcdf18aa43cd38bbe5567504838a3b1dca482035458853c5a171709dfae9df551815010bdfbc6df733cde84c4f7a5b0591d9cda9db087fb411ee3e2a4f19ad50c8331712ecdc5dd7ce34b0203010001",
			hash:  "2d6fe5ec59d7240b8a4c27c2ff27ba1071105fa50d45543768fcbabf9ee3cb8f8fa0afa51e08e053af30f6d11066ebfd47e75bda5ccc085c115d7e1896f3c62f",
			match: true,
		},
		"mismatching hash": {
			key:   "30819f300d06092a864886f70d010101050003818d0030818902818100d4b2f072a32fa98456eb7f5938e2ff361fb64d698ea91e003d34bfc5374b814c16ba9ae3ec392ef6d48cf79b63067e338aa941219a7bcdf18aa43cd38bbe5567504838a3b1dca482035458853c5a171709dfae9df551815010bdfbc6df733cde84c4f7a5b0591d9cda9db087fb411ee3e2a4f19ad50c8331712ecdc5dd7ce34b0203010001",
			hash:  "00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
			match: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			newKey, err := loadKeyFromHex(tc.key)
			require.NoError(err)

			// Function under test:
			hash, err := sha512sum(newKey)
			assert.NoError(err)

			expected, err := hex.DecodeString(tc.hash)
			require.NoError(err)

			if tc.match {
				assert.True(bytes.Equal(expected, hash[:]), fmt.Sprintf("expected hash %x, got %x", expected, hash))
			} else {
				assert.False(bytes.Equal(expected, hash[:]), fmt.Sprintf("expected mismatching hashes, got %x", hash))
			}
		})
	}
}

func loadKeyFromHex(key string) (crypto.PublicKey, error) {
	decoded, err := hex.DecodeString(key)
	if err != nil {
		return nil, err
	}

	return x509.ParsePKIXPublicKey(decoded)
}

type stubGCPValidator struct{}

func (stubGCPValidator) validate(_ vtpm.AttestationDocument, _ *x509.Certificate, _ *x509.Certificate, _ [64]byte, _ *config.GCPSEVSNP, _ attestation.Logger) error {
	return nil
}
