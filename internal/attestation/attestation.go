/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
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
