/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
This package deals with the low level attestation and verification logic of Constellation nodes.

General tpm attestation code that is not subjective to a single platform should go into the vtpm package.
Since attestation capabilities can differ between platforms, the attestation code should go into a subpackage for that respective platform.

We commonly implement the following two interfaces for a platform:

	// Issuer issues an attestation document.
	type Issuer interface {
	    oid.Getter
	    Issue(userData []byte, nonce []byte) (quote []byte, err error)
	}

	// Validator is able to validate an attestation document.
	type Validator interface {
	    oid.Getter
	    Validate(attDoc []byte, nonce []byte) ([]byte, error)
	}

Attestation code for new platforms needs to implement these two interfaces.
*/
package attestation

import (
	"bytes"
	"crypto/sha256"

	"github.com/edgelesssys/constellation/v2/internal/crypto"
)

const (
	// clusterIDContext is the value to use for info when deriving the cluster ID.
	clusterIDContext = "clusterID"
	// MeasurementSecretContext is the value to use for info
	// when deriving the measurement secret from the master secret.
	MeasurementSecretContext = "measurementSecret"
)

// Logger is a logger used to print warnings and infos during attestation validation.
type Logger interface {
	Infof(format string, args ...any)
	Warnf(format string, args ...any)
}

// NOPLogger is a no-op implementation of [Logger].
type NOPLogger struct{}

// Infof is a no-op.
func (NOPLogger) Infof(string, ...interface{}) {}

// Warnf is a no-op.
func (NOPLogger) Warnf(string, ...interface{}) {}

// DeriveClusterID derives the cluster ID from a salt and secret value.
func DeriveClusterID(secret, salt []byte) ([]byte, error) {
	return crypto.DeriveKey(secret, salt, []byte(crypto.DEKPrefix+clusterIDContext), crypto.DerivedKeyLengthDefault)
}

// MakeExtraData binds userData to a random nonce used in attestation.
func MakeExtraData(userData []byte, nonce []byte) []byte {
	data := append([]byte{}, userData...)
	data = append(data, nonce...)
	digest := sha256.Sum256(data)
	return digest[:]
}

// CompareExtraData compares the extra data of a quote with the expected extra data.
// Returns true if the data from the quote matches the expected data.
// If the slices are not of equal length, the shorter slice is padded with zeros.
func CompareExtraData(quoteData, expectedData []byte) bool {
	if len(quoteData) != len(expectedData) {
		// If the lengths are not equal, pad the shorter slice with zeros.
		diff := len(quoteData) - len(expectedData)
		if diff < 0 {
			diff = -diff
			quoteData = append(quoteData, bytes.Repeat([]byte{0x00}, diff)...)
		} else {
			expectedData = append(expectedData, bytes.Repeat([]byte{0x00}, diff)...)
		}
	}
	return bytes.Equal(quoteData, expectedData)
}
