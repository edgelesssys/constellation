/*
Copyright (c) Edgeless Systems GmbH
SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"context"
	"io"
	"path/filepath"
	"strings"
	"testing"

	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIamUpgradeApply(t *testing.T) {
	setupFs := func(createConfig, createTerraformVars bool) file.Handler {
		fs := afero.NewMemMapFs()
		fh := file.NewHandler(fs)
		if createConfig {
			cfg := defaultConfigWithExpectedMeasurements(t, config.Default(), cloudprovider.Azure)
			require.NoError(t, fh.WriteYAML(constants.ConfigFilename, cfg))
		}
		if createTerraformVars {
			require.NoError(t, fh.Write(
				filepath.Join(constants.TerraformIAMWorkingDir, "terraform.tfvars"),
				[]byte(
					"region = \"foo\"\n"+
						"resource_group_name = \"bar\"\n"+
						"service_principal_name = \"baz\"\n",
				),
			))
		}
		return fh
	}

	testCases := map[string]struct {
		fh          file.Handler
		iamUpgrader *stubIamUpgrader
		yesFlag     bool
		input       string
		wantErr     bool
	}{
		"success": {
			fh:          setupFs(true, true),
			iamUpgrader: &stubIamUpgrader{},
		},
		"abort": {
			fh:          setupFs(true, true),
			iamUpgrader: &stubIamUpgrader{},
			input:       "no",
			yesFlag:     true,
		},
		"config missing": {
			fh:          setupFs(false, true),
			iamUpgrader: &stubIamUpgrader{},
			yesFlag:     true,
			wantErr:     true,
		},
		"iam vars missing": {
			fh:          setupFs(true, false),
			iamUpgrader: &stubIamUpgrader{},
			yesFlag:     true,
			wantErr:     true,
		},
		"plan error": {
			fh: setupFs(true, true),
			iamUpgrader: &stubIamUpgrader{
				planErr: assert.AnError,
			},
			yesFlag: true,
			wantErr: true,
		},
		"apply error": {
			fh: setupFs(true, true),
			iamUpgrader: &stubIamUpgrader{
				hasDiff:  true,
				applyErr: assert.AnError,
			},
			yesFlag: true,
			wantErr: true,
		},
		"rollback error, log only": {
			fh: setupFs(true, true),
			iamUpgrader: &stubIamUpgrader{
				rollbackErr: assert.AnError,
			},
			input:   "no",
			yesFlag: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			cmd := newIAMUpgradeApplyCmd()
			cmd.SetIn(strings.NewReader(tc.input))

			iamUpgradeApplyCmd := &iamUpgradeApplyCmd{
				fileHandler: tc.fh,
				log:         logger.NewTest(t),
			}

			err := iamUpgradeApplyCmd.iamUpgradeApply(cmd, tc.iamUpgrader, "test", false, tc.yesFlag)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

type stubIamUpgrader struct {
	hasDiff     bool
	planErr     error
	applyErr    error
	rollbackErr error
}

func (u *stubIamUpgrader) PlanIAMUpgrade(context.Context, io.Writer, terraform.Variables, cloudprovider.Provider) (bool, error) {
	return u.hasDiff, u.planErr
}

func (u *stubIamUpgrader) ApplyIAMUpgrade(context.Context, cloudprovider.Provider) error {
	return u.applyErr
}

func (u *stubIamUpgrader) RollbackIAMWorkspace() error {
	return u.rollbackErr
}
