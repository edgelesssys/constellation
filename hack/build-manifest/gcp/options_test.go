package gcp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsGcpReleaseImage(t *testing.T) {
	testCases := map[string]struct {
		image       string
		wantVersion string
		wantError   bool
	}{
		"works for release image": {
			image:       "projects/constellation-images/global/images/constellation-v1-3-0",
			wantVersion: "v1.3.0",
		},
		"breaks for debug image": {
			image:     "projects/constellation-images/global/images/constellation-20220805151600",
			wantError: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			version, err := isGcpReleaseImage(tc.image)
			if tc.wantError {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			assert.Equal(tc.wantVersion, version)
		})
	}
}
