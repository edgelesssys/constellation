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
	"github.com/edgelesssys/constellation/v2/internal/api/attestationconfigapi"
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
		fh            file.Handler
		iamUpgrader   *stubIamUpgrader
		configFetcher *stubConfigFetcher
		yesFlag       bool
		input         string
		wantErr       bool
	}{
		"success": {
			fh:            setupFs(true, true),
			configFetcher: &stubConfigFetcher{},
			iamUpgrader:   &stubIamUpgrader{},
		},
		"abort": {
			fh:            setupFs(true, true),
			iamUpgrader:   &stubIamUpgrader{},
			configFetcher: &stubConfigFetcher{},
			input:         "no",
			yesFlag:       true,
		},
		"config missing": {
			fh:            setupFs(false, true),
			iamUpgrader:   &stubIamUpgrader{},
			configFetcher: &stubConfigFetcher{},
			yesFlag:       true,
			wantErr:       true,
		},
		"iam vars missing": {
			fh:            setupFs(true, false),
			iamUpgrader:   &stubIamUpgrader{},
			configFetcher: &stubConfigFetcher{},
			yesFlag:       true,
			wantErr:       true,
		},
		"plan error": {
			fh: setupFs(true, true),
			iamUpgrader: &stubIamUpgrader{
				planErr: assert.AnError,
			},
			configFetcher: &stubConfigFetcher{},
			yesFlag:       true,
			wantErr:       true,
		},
		"apply error": {
			fh: setupFs(true, true),
			iamUpgrader: &stubIamUpgrader{
				hasDiff:  true,
				applyErr: assert.AnError,
			},
			configFetcher: &stubConfigFetcher{},
			yesFlag:       true,
			wantErr:       true,
		},
		"restore error": {
			fh: setupFs(true, true),
			iamUpgrader: &stubIamUpgrader{
				hasDiff:     true,
				rollbackErr: assert.AnError,
			},
			configFetcher: &stubConfigFetcher{},
			input:         "no\n",
			wantErr:       true,
		},
		"config fetcher err": {
			fh: setupFs(true, true),
			iamUpgrader: &stubIamUpgrader{
				rollbackErr: assert.AnError,
			},
			configFetcher: &stubConfigFetcher{
				fetchLatestErr: assert.AnError,
			},
			yesFlag: true,
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			cmd := newIAMUpgradeApplyCmd()
			cmd.SetIn(strings.NewReader(tc.input))

			iamUpgradeApplyCmd := &iamUpgradeApplyCmd{
				fileHandler:   tc.fh,
				log:           logger.NewTest(t),
				configFetcher: tc.configFetcher,
				flags: iamUpgradeApplyFlags{
					yes: tc.yesFlag,
				},
			}

			err := iamUpgradeApplyCmd.iamUpgradeApply(cmd, tc.iamUpgrader, "")
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

func (u *stubIamUpgrader) RestoreIAMWorkspace() error {
	return u.rollbackErr
}

type stubConfigFetcher struct {
	fetchLatestErr error
}

func (s *stubConfigFetcher) FetchAzureSEVSNPVersion(context.Context, attestationconfigapi.AzureSEVSNPVersionAPI) (attestationconfigapi.AzureSEVSNPVersionAPI, error) {
	panic("not implemented")
}

func (s *stubConfigFetcher) FetchAzureSEVSNPVersionList(context.Context, attestationconfigapi.AzureSEVSNPVersionList) (attestationconfigapi.AzureSEVSNPVersionList, error) {
	panic("not implemented")
}

func (s *stubConfigFetcher) FetchAzureSEVSNPVersionLatest(context.Context) (attestationconfigapi.AzureSEVSNPVersionAPI, error) {
	return attestationconfigapi.AzureSEVSNPVersionAPI{}, s.fetchLatestErr
}
