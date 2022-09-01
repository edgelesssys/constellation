package trustedlaunch

import (
	"testing"

	"github.com/edgelesssys/constellation/internal/attestation/simulator"
	"github.com/edgelesssys/constellation/internal/attestation/vtpm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetSNPAttestation(t *testing.T) {
	testCases := map[string]struct {
		tpmFunc vtpm.TPMOpenFunc
		wantErr bool
	}{
		"success": {
			tpmFunc: simulator.OpenSimulatedTPM,
			wantErr: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			tpm, err := tc.tpmFunc()
			require.NoError(err)
			defer tpm.Close()

			_, err = getAttestation(tpm)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}
