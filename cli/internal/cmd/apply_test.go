/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/v2/cli/internal/helm"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/spf13/afero"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseApplyFlags(t *testing.T) {
	require := require.New(t)
	defaultFlags := func() *pflag.FlagSet {
		flags := NewApplyCmd().Flags()
		// Register persistent flags
		flags.String("workspace", "", "")
		flags.String("tf-log", "NONE", "")
		flags.Bool("force", false, "")
		flags.Bool("debug", false, "")
		return flags
	}

	testCases := map[string]struct {
		flags     *pflag.FlagSet
		wantFlags applyFlags
		wantErr   bool
	}{
		"default flags": {
			flags: defaultFlags(),
			wantFlags: applyFlags{
				helmWaitMode:   helm.WaitModeAtomic,
				upgradeTimeout: 5 * time.Minute,
			},
		},
		"skip phases": {
			flags: func() *pflag.FlagSet {
				flags := defaultFlags()
				require.NoError(flags.Set("skip-phases", fmt.Sprintf("%s,%s", skipHelmPhase, skipK8sPhase)))
				return flags
			}(),
			wantFlags: applyFlags{
				skipPhases:     skipPhases{skipHelmPhase: struct{}{}, skipK8sPhase: struct{}{}},
				helmWaitMode:   helm.WaitModeAtomic,
				upgradeTimeout: 5 * time.Minute,
			},
		},
		"skip helm wait": {
			flags: func() *pflag.FlagSet {
				flags := defaultFlags()
				require.NoError(flags.Set("skip-helm-wait", "true"))
				return flags
			}(),
			wantFlags: applyFlags{
				helmWaitMode:   helm.WaitModeNone,
				upgradeTimeout: 5 * time.Minute,
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			var flags applyFlags

			err := flags.parse(tc.flags)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			assert.Equal(tc.wantFlags, flags)
		})
	}
}

func TestBackupHelmCharts(t *testing.T) {
	testCases := map[string]struct {
		helmApplier      helm.Applier
		backupClient     *stubKubernetesUpgrader
		includesUpgrades bool
		wantErr          bool
	}{
		"success, no upgrades": {
			helmApplier:  &stubRunner{},
			backupClient: &stubKubernetesUpgrader{},
		},
		"success with upgrades": {
			helmApplier:      &stubRunner{},
			backupClient:     &stubKubernetesUpgrader{},
			includesUpgrades: true,
		},
		"saving charts fails": {
			helmApplier: &stubRunner{
				saveChartsErr: assert.AnError,
			},
			backupClient: &stubKubernetesUpgrader{},
			wantErr:      true,
		},
		"backup CRDs fails": {
			helmApplier: &stubRunner{},
			backupClient: &stubKubernetesUpgrader{
				backupCRDsErr: assert.AnError,
			},
			includesUpgrades: true,
			wantErr:          true,
		},
		"backup CRs fails": {
			helmApplier: &stubRunner{},
			backupClient: &stubKubernetesUpgrader{
				backupCRsErr: assert.AnError,
			},
			includesUpgrades: true,
			wantErr:          true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			a := applyCmd{
				fileHandler: file.NewHandler(afero.NewMemMapFs()),
				log:         logger.NewTest(t),
			}

			err := a.backupHelmCharts(context.Background(), tc.backupClient, tc.helmApplier, tc.includesUpgrades, "")
			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			if tc.includesUpgrades {
				assert.True(tc.backupClient.backupCRDsCalled)
				assert.True(tc.backupClient.backupCRsCalled)
			}
		})
	}
}
