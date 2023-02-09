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

	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/versions/components"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpgradeApply(t *testing.T) {
	testCases := map[string]struct {
		upgrader     stubUpgrader
		imageFetcher stubImageFetcher
		wantErr      bool
	}{
		"success": {
			imageFetcher: stubImageFetcher{
				reference: "someReference",
			},
		},
		"fetch error": {
			imageFetcher: stubImageFetcher{
				fetchReferenceErr: errors.New("error"),
			},
			wantErr: true,
		},
		"upgrade error": {
			upgrader: stubUpgrader{imageErr: errors.New("error")},
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

			handler := file.NewHandler(afero.NewMemMapFs())
			cfg := defaultConfigWithExpectedMeasurements(t, config.Default(), cloudprovider.Azure)
			require.NoError(handler.WriteYAML(constants.ConfigFilename, cfg))

			upgrader := upgradeApplyCmd{upgrader: tc.upgrader, log: logger.NewTest(t)}
			err := upgrader.upgradeApply(cmd, &tc.imageFetcher, handler)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

type stubUpgrader struct {
	imageErr error
	helmErr  error
	k8sErr   error
}

func (u stubUpgrader) UpgradeImage(context.Context, string, string, measurements.M) error {
	return u.imageErr
}

func (u stubUpgrader) UpgradeHelmServices(ctx context.Context, config *config.Config, timeout time.Duration, allowDestructive bool) error {
	return u.helmErr
}

func (u stubUpgrader) UpgradeK8s(ctx context.Context, clusterVersion string, components components.Components) error {
	return u.k8sErr
}

type stubImageFetcher struct {
	reference         string
	fetchReferenceErr error
}

func (f *stubImageFetcher) FetchReference(_ context.Context, _ *config.Config) (string, error) {
	return f.reference, f.fetchReferenceErr
}
