package aws

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidate(t *testing.T) {
	testCases := map[string]struct {
		nonce            []byte
		verifyResult     string
		verifyErr        error
		expectedUserData []byte
		expectErr        bool
	}{
		"valid": {
			nonce:            []byte{2, 3, 4},
			verifyResult:     `{"nonce":[2,3,4], "user_data":[5,6,7]}`,
			expectedUserData: []byte{5, 6, 7},
		},
		"invalid nonce": {
			nonce:        []byte{2, 3, 5},
			verifyResult: `{"nonce":[2,3,4], "user_data":[5,6,7]}`,
			expectErr:    true,
		},
		"nil nonce": {
			nonce:        nil,
			verifyResult: `{"nonce":[2,3,4], "user_data":[5,6,7]}`,
			expectErr:    true,
		},
		"verify error": {
			nonce:     []byte{2, 3, 4},
			verifyErr: errors.New("failed"),
			expectErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			doc := []byte("doc")

			verify := func(adBlob []byte, rootCertDer []byte, ts time.Time) (string, error) {
				assert.Equal(doc, adBlob)
				assert.Equal(awsNitroEnclavesRoot, rootCertDer)
				return tc.verifyResult, tc.verifyErr
			}

			userData, err := NewValidator(verify).Validate(doc, tc.nonce)
			if tc.expectErr {
				require.Error(err)
				return
			}

			require.NoError(err)
			assert.Equal(tc.expectedUserData, userData)
		})
	}
}
