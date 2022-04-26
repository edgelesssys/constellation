package cmd

import (
	"bytes"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetGCPValidator(t *testing.T) {
	testCases := map[string]struct {
		ownerID   string
		clusterID string
		wantErr   bool
	}{
		"no input": {
			ownerID:   "",
			clusterID: "",
			wantErr:   true,
		},
		"unencoded secret ID": {
			ownerID:   "owner-id",
			clusterID: base64.StdEncoding.EncodeToString([]byte("unique-id")),
			wantErr:   true,
		},
		"unencoded cluster ID": {
			ownerID:   base64.StdEncoding.EncodeToString([]byte("owner-id")),
			clusterID: "unique-id",
			wantErr:   true,
		},
		"correct input": {
			ownerID:   base64.StdEncoding.EncodeToString([]byte("owner-id")),
			clusterID: base64.StdEncoding.EncodeToString([]byte("unique-id")),
			wantErr:   false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			cmd := newVerifyGCPCmd()
			cmd.Flags().String("owner-id", "", "")
			cmd.Flags().String("unique-id", "", "")
			require.NoError(cmd.Flags().Set("owner-id", tc.ownerID))
			require.NoError(cmd.Flags().Set("unique-id", tc.clusterID))
			var out bytes.Buffer
			cmd.SetOut(&out)
			var errOut bytes.Buffer
			cmd.SetErr(&errOut)

			_, err := getGCPValidator(cmd, map[uint32][]byte{
				0: []byte("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"),
				1: []byte("BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB"),
			})
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}
