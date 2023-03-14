/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/v2/cli/internal/cloudcmd"
	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
)

func TestUpgradeApply(t *testing.T) {
	someErr := errors.New("some error")
	testCases := map[string]struct {
		upgrader stubUpgrader
		wantErr  bool
	}{
		"success": {
			upgrader: stubUpgrader{},
		},
		"nodeVersion some error": {
			upgrader: stubUpgrader{nodeVersionErr: someErr},
			wantErr:  true,
		},
		"nodeVersion in progress error": {
			upgrader: stubUpgrader{nodeVersionErr: cloudcmd.ErrInProgress},
		},
		"helm other error": {
			upgrader: stubUpgrader{helmErr: someErr},
			wantErr:  true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)
			cmd := newUpgradeApplyCmd()
			cmd.Flags().String("config", constants.ConfigFilename, "") // register persistent flag manually
			cmd.Flags().Bool("force", true, "")                        // register persistent flag manually

			err := cmd.Flags().Set("yes", "true")
			require.NoError(err)

			handler := file.NewHandler(afero.NewMemMapFs())
			cfg := defaultConfigWithExpectedMeasurements(t, config.Default(), cloudprovider.Azure)
			require.NoError(handler.WriteYAML(constants.ConfigFilename, cfg))

			upgrader := upgradeApplyCmd{upgrader: tc.upgrader, log: logger.NewTest(t)}
			err = upgrader.upgradeApply(cmd, handler)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

type stubUpgrader struct {
	nodeVersionErr error
	helmErr        error
}

func (u stubUpgrader) UpgradeNodeVersion(_ context.Context, _ *config.Config) error {
	return u.nodeVersionErr
}

func (u stubUpgrader) UpgradeHelmServices(_ context.Context, _ *config.Config, _ time.Duration, _ bool) error {
	return u.helmErr
}

func (u stubUpgrader) UpdateMeasurements(_ context.Context, _ measurements.M) error {
	return nil
}

func (u stubUpgrader) GetClusterMeasurements(_ context.Context) (measurements.M, *corev1.ConfigMap, error) {
	return measurements.M{}, &corev1.ConfigMap{}, nil
}
