package client

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadProjectID(t *testing.T) {
	testCases := map[string]struct {
		rawConfig     string
		skipWrite     bool
		wantProjectID string
		wantErr       bool
	}{
		"valid config": {
			rawConfig:     `project-id = project-id`,
			wantProjectID: "project-id",
		},
		"invalid config": {
			rawConfig: `x = y`,
			wantErr:   true,
		},
		"config is empty": {
			wantErr: true,
		},
		"config does not exist": {
			skipWrite: true,
			wantErr:   true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			fs := afero.NewMemMapFs()
			if !tc.skipWrite {
				require.NoError(afero.WriteFile(fs, "gce.conf", []byte(tc.rawConfig), 0o644))
			}
			gotProjectID, err := loadProjectID(fs, "gce.conf")
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantProjectID, gotProjectID)
		})
	}
}
