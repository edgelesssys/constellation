package attestation

import (
	"github.com/edgelesssys/constellation/internal/crypto"
)

const (
	// clusterIDContext is the value to use for info when deriving the cluster ID.
	clusterIDContext = "clusterID"
	// MeasurementSecretContext is the value to use for info
	// when deriving the measurement secret from the master secret.
	MeasurementSecretContext = "measurementSecret"
)

// DeriveClusterID derives the cluster ID from a salt and secret value.
func DeriveClusterID(salt, secret []byte) ([]byte, error) {
	return crypto.DeriveKey(secret, salt, []byte(crypto.HKDFInfoPrefix+clusterIDContext), crypto.DerivedKeyLengthDefault)
}

// DeriveMeasurementSecret derives the secret value needed to derive ClusterID.
func DeriveMeasurementSecret(masterSecret []byte) ([]byte, error) {
	// TODO: replace hard coded salt
	return crypto.DeriveKey(masterSecret, []byte("Constellation"), []byte(crypto.HKDFInfoPrefix+MeasurementSecretContext), crypto.DerivedKeyLengthDefault)
}
