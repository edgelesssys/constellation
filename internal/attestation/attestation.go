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
	"github.com/edgelesssys/constellation/v2/internal/crypto"
)

const (
	// clusterIDContext is the value to use for info when deriving the cluster ID.
	clusterIDContext = "clusterID"
	// MeasurementSecretContext is the value to use for info
	// when deriving the measurement secret from the master secret.
	MeasurementSecretContext = "measurementSecret"
)

// DeriveClusterID derives the cluster ID from a salt and secret value.
func DeriveClusterID(secret, salt []byte) ([]byte, error) {
	return crypto.DeriveKey(secret, salt, []byte(crypto.DEKPrefix+clusterIDContext), crypto.DerivedKeyLengthDefault)
}
