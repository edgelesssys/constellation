/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package helm

import (
	"context"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
)

func TestIsUpgrade(t *testing.T) {
	testCases := map[string]struct {
		currentVersion string
		newVersion     string
		wantUpgrade    bool
	}{
		"upgrade": {
			currentVersion: "0.1.0",
			newVersion:     "0.2.0",
			wantUpgrade:    true,
		},
		"downgrade": {
			currentVersion: "0.2.0",
			newVersion:     "0.1.0",
			wantUpgrade:    false,
		},
		"equal": {
			currentVersion: "0.1.0",
			newVersion:     "0.1.0",
			wantUpgrade:    false,
		},
		"invalid current version": {
			currentVersion: "asdf",
			newVersion:     "0.1.0",
			wantUpgrade:    false,
		},
		"invalid new version": {
			currentVersion: "0.1.0",
			newVersion:     "asdf",
			wantUpgrade:    false,
		},
		"patch version": {
			currentVersion: "0.1.0",
			newVersion:     "0.1.1",
			wantUpgrade:    true,
		},
		"pre-release version": {
			currentVersion: "0.1.0",
			newVersion:     "0.1.1-rc1",
			wantUpgrade:    true,
		},
		"pre-release version downgrade": {
			currentVersion: "0.1.1-rc1",
			newVersion:     "0.1.0",
			wantUpgrade:    false,
		},
		"pre-release of same version": {
			currentVersion: "0.1.0",
			newVersion:     "0.1.0-rc1",
			wantUpgrade:    false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			upgrade := isUpgrade(tc.currentVersion, tc.newVersion)
			assert.Equal(tc.wantUpgrade, upgrade)

			upgrade = isUpgrade("v"+tc.currentVersion, "v"+tc.newVersion)
			assert.Equal(tc.wantUpgrade, upgrade)
		})
	}
}

func TestUpgradeRelease(t *testing.T) {
	testCases := map[string]struct {
		allowDestructive bool
		wantError        bool
	}{
		"allow": {
			allowDestructive: true,
		},
		"deny": {
			allowDestructive: false,
			wantError:        true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			client := Client{kubectl: nil, actions: &stubActionWrapper{}, log: logger.NewTest(t)}
			err := client.upgradeRelease(context.Background(), 0, config.Default(), certManagerPath, certManagerReleaseName, false, tc.allowDestructive)
			if tc.wantError {
				assert.ErrorIs(err, ErrConfirmationMissing)
				return
			}
			assert.NoError(err)
		})
	}
}

type stubActionWrapper struct{}

// listAction returns a list of len 1 with a release that has only it's version set.
func (a *stubActionWrapper) listAction(_ string) ([]*release.Release, error) {
	return []*release.Release{{Chart: &chart.Chart{Metadata: &chart.Metadata{Version: "1.0.0"}}}}, nil
}

func (a *stubActionWrapper) getValues(release string) (map[string]any, error) {
	return nil, nil
}

func (a *stubActionWrapper) upgradeAction(ctx context.Context, releaseName string, chart *chart.Chart, values map[string]any, timeout time.Duration) error {
	return nil
}
