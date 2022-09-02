package proto

import (
	"testing"

	"github.com/edgelesssys/constellation/internal/crypto/testvector"
	"github.com/stretchr/testify/assert"
)

func TestDeriveStateDiskKey(t *testing.T) {
	testCases := map[string]struct {
		masterSecret testvector.HKDF
	}{
		"all zero": {
			masterSecret: testvector.HKDFZero,
		},
		"all 0xff": {
			masterSecret: testvector.HKDF0xFF,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			stateDiskKey, err := deriveStateDiskKey(tc.masterSecret.Secret, tc.masterSecret.Salt, tc.masterSecret.Info)

			assert.NoError(err)
			assert.Equal(tc.masterSecret.Output, stateDiskKey)
		})
	}
}
