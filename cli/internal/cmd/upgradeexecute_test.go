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
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpgradeExecute(t *testing.T) {
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
			upgrader: stubUpgrader{err: errors.New("error")},
			wantErr:  true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)
			cmd := newUpgradeExecuteCmd()
			cmd.Flags().String("config", constants.ConfigFilename, "") // register persistent flag manually

			handler := file.NewHandler(afero.NewMemMapFs())
			cfg := defaultConfigWithExpectedMeasurements(t, config.Default(), cloudprovider.Azure)
			require.NoError(handler.WriteYAML(constants.ConfigFilename, cfg))

			err := upgradeExecute(cmd, &tc.imageFetcher, tc.upgrader, handler)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

type stubUpgrader struct {
	err     error
	helmErr error
}

func (u stubUpgrader) Upgrade(context.Context, string, string, measurements.M) error {
	return u.err
}

func (u stubUpgrader) UpgradeHelmServices(ctx context.Context, config *config.Config, timeout time.Duration) error {
	return u.helmErr
}

type stubImageFetcher struct {
	reference         string
	fetchReferenceErr error
}

func (f *stubImageFetcher) FetchReference(_ context.Context, _ *config.Config) (string, error) {
	return f.reference, f.fetchReferenceErr
}
