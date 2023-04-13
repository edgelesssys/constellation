/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package helm

import (
	"context"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/compatibility"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
)

func TestShouldUpgrade(t *testing.T) {
	testCases := map[string]struct {
		version            string
		assertCorrectError func(t *testing.T, err error) bool
		wantError          bool
	}{
		"valid upgrade": {
			version: "1.9.0",
		},
		"not a valid upgrade": {
			version: "1.0.0",
			assertCorrectError: func(t *testing.T, err error) bool {
				target := &compatibility.InvalidUpgradeError{}
				return assert.ErrorAs(t, err, &target)
			},
			wantError: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			client := Client{kubectl: nil, actions: &stubActionWrapper{version: tc.version}, log: logger.NewTest(t)}

			chart, err := loadChartsDir(helmFS, certManagerInfo.path)
			require.NoError(err)
			err = client.shouldUpgrade(certManagerInfo.releaseName, chart.Metadata.Version)
			if tc.wantError {
				tc.assertCorrectError(t, err)
				return
			}
			assert.NoError(err)
		})
	}
}

func TestUpgradeRelease(t *testing.T) {
	testCases := map[string]struct {
		allowDestructive   bool
		version            string
		assertCorrectError func(t *testing.T, err error) bool
		wantError          bool
	}{
		"allow": {
			allowDestructive: true,
			version:          "1.9.0",
		},
		"deny": {
			allowDestructive: false,
			version:          "1.9.0",
			assertCorrectError: func(t *testing.T, err error) bool {
				return assert.ErrorIs(t, err, ErrConfirmationMissing)
			},
			wantError: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			client := Client{kubectl: nil, actions: &stubActionWrapper{version: tc.version}, log: logger.NewTest(t)}

			chart, err := loadChartsDir(helmFS, certManagerInfo.path)
			require.NoError(err)
			fileHandler := file.NewHandler(afero.NewMemMapFs())
			err = client.upgradeRelease(context.Background(), 0, config.Default(), chart, tc.allowDestructive, fileHandler)
			if tc.wantError {
				tc.assertCorrectError(t, err)
				return
			}
			assert.NoError(err)
		})
	}
}

type stubActionWrapper struct {
	version string
}

// listAction returns a list of len 1 with a release that has only it's version set.
func (a *stubActionWrapper) listAction(_ string) ([]*release.Release, error) {
	return []*release.Release{{Chart: &chart.Chart{Metadata: &chart.Metadata{Version: a.version}}}}, nil
}

func (a *stubActionWrapper) getValues(_ string) (map[string]any, error) {
	return nil, nil
}

func (a *stubActionWrapper) upgradeAction(_ context.Context, _ string, _ *chart.Chart, _ map[string]any, _ time.Duration) error {
	return nil
}
