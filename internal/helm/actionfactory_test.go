/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package helm

import (
	"errors"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/compatibility"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/semver"
	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
)

func TestAppendNewAction(t *testing.T) {
	assertUpgradeErr := func(assert *assert.Assertions, err error) {
		var invalidUpgrade *compatibility.InvalidUpgradeError
		assert.True(errors.As(err, &invalidUpgrade))
	}

	testCases := map[string]struct {
		lister              stubLister
		release             Release
		configTargetVersion semver.Semver
		force               bool
		allowDestructive    bool
		wantErr             bool
		assertErr           func(*assert.Assertions, error)
	}{
		"upgrade release": {
			lister: stubLister{version: semver.NewFromInt(1, 0, 0, "")},
			release: Release{
				ReleaseName: "test",
				Chart: &chart.Chart{
					Metadata: &chart.Metadata{
						Version: "1.1.0",
					},
				},
			},
			configTargetVersion: semver.NewFromInt(1, 1, 0, ""),
		},
		"upgrade to same version": {
			lister: stubLister{version: semver.NewFromInt(1, 0, 0, "")},
			release: Release{
				ReleaseName: "test",
				Chart: &chart.Chart{
					Metadata: &chart.Metadata{
						Version: "1.0.0",
					},
				},
			},
			configTargetVersion: semver.NewFromInt(1, 1, 0, ""),
			wantErr:             true,
			assertErr:           assertUpgradeErr,
		},
		"upgrade to older version": {
			lister: stubLister{version: semver.NewFromInt(1, 1, 0, "")},
			release: Release{
				ReleaseName: "test",
				Chart: &chart.Chart{
					Metadata: &chart.Metadata{
						Version: "1.0.0",
					},
				},
			},
			configTargetVersion: semver.NewFromInt(1, 0, 0, ""),
			wantErr:             true,
			assertErr:           assertUpgradeErr,
		},
		"upgrade to older version can be forced": {
			lister: stubLister{version: semver.NewFromInt(1, 1, 0, "")},
			release: Release{
				ReleaseName: "test",
				Chart: &chart.Chart{
					Metadata: &chart.Metadata{
						Version: "1.0.0",
					},
				},
			},
			configTargetVersion: semver.NewFromInt(1, 0, 0, ""),
			force:               true,
		},
		"non semver in chart metadata": {
			lister: stubLister{version: semver.NewFromInt(1, 0, 0, "")},
			release: Release{
				ReleaseName: "test",
				Chart: &chart.Chart{
					Metadata: &chart.Metadata{
						Version: "some-version",
					},
				},
			},
			wantErr: true,
		},
		"listing release fails": {
			lister: stubLister{err: assert.AnError},
			release: Release{
				ReleaseName: "test",
				Chart: &chart.Chart{
					Metadata: &chart.Metadata{
						Version: "1.1.0",
					},
				},
			},
			configTargetVersion: semver.NewFromInt(1, 1, 0, ""),
			wantErr:             true,
		},
		"release not installed": {
			lister: stubLister{err: errReleaseNotFound},
			release: Release{
				ReleaseName: "test",
				Chart: &chart.Chart{
					Metadata: &chart.Metadata{
						Version: "1.1.0",
					},
				},
			},
			configTargetVersion: semver.NewFromInt(1, 1, 0, ""),
		},
		"destructive release upgrade requires confirmation": {
			lister: stubLister{version: semver.NewFromInt(1, 0, 0, "")},
			release: Release{
				Chart: &chart.Chart{
					Metadata: &chart.Metadata{
						Version: "1.1.0",
					},
				},
				ReleaseName: certManagerInfo.releaseName,
			},
			configTargetVersion: semver.NewFromInt(1, 1, 0, ""),
			wantErr:             true,
			assertErr: func(assert *assert.Assertions, err error) {
				assert.ErrorIs(err, ErrConfirmationMissing)
			},
		},
		"destructive release upgrade can be accepted": {
			lister: stubLister{version: semver.NewFromInt(1, 0, 0, "")},
			release: Release{
				Chart: &chart.Chart{
					Metadata: &chart.Metadata{
						Version: "1.1.0",
					},
				},
				ReleaseName: certManagerInfo.releaseName,
			},
			configTargetVersion: semver.NewFromInt(1, 1, 0, ""),
			allowDestructive:    true,
		},
		"config version takes precedence over CLI version": {
			lister: stubLister{version: semver.NewFromInt(1, 0, 0, "")},
			release: Release{
				ReleaseName: constellationServicesInfo.releaseName,
				Chart: &chart.Chart{
					Metadata: &chart.Metadata{
						Version: "1.1.0",
					},
				},
			},
			configTargetVersion: semver.NewFromInt(1, 0, 0, ""),
			wantErr:             true,
			assertErr:           assertUpgradeErr,
		},
		"error if CLI version does not match config version on upgrade": {
			lister: stubLister{version: semver.NewFromInt(1, 0, 0, "")},
			release: Release{
				ReleaseName: constellationServicesInfo.releaseName,
				Chart: &chart.Chart{
					Metadata: &chart.Metadata{
						Version: "1.1.5",
					},
				},
			},
			configTargetVersion: semver.NewFromInt(1, 1, 0, ""),
			wantErr:             true,
			assertErr:           assertUpgradeErr,
		},
		"config version matches CLI version on upgrade": {
			lister: stubLister{version: semver.NewFromInt(1, 0, 0, "")},
			release: Release{
				ReleaseName: constellationServicesInfo.releaseName,
				Chart: &chart.Chart{
					Metadata: &chart.Metadata{
						Version: "1.1.5",
					},
				},
			},
			configTargetVersion: semver.NewFromInt(1, 1, 5, ""),
		},
		"config - CLI version mismatch can be forced through": {
			lister: stubLister{version: semver.NewFromInt(1, 0, 0, "")},
			release: Release{
				ReleaseName: constellationServicesInfo.releaseName,
				Chart: &chart.Chart{
					Metadata: &chart.Metadata{
						Version: "1.1.5",
					},
				},
			},
			configTargetVersion: semver.NewFromInt(1, 1, 0, ""),
			force:               true,
		},
		"installing new release requires matching config and CLI version": {
			lister: stubLister{err: errReleaseNotFound},
			release: Release{
				ReleaseName: constellationServicesInfo.releaseName,
				Chart: &chart.Chart{
					Metadata: &chart.Metadata{
						Version: "1.1.0",
					},
				},
			},
			configTargetVersion: semver.NewFromInt(1, 0, 0, ""),
			wantErr:             true,
			assertErr:           assertUpgradeErr,
		},
		"config - CLI version mismatch for new releases can be forced through": {
			lister: stubLister{err: errReleaseNotFound},
			release: Release{
				ReleaseName: constellationServicesInfo.releaseName,
				Chart: &chart.Chart{
					Metadata: &chart.Metadata{
						Version: "1.1.0",
					},
				},
			},
			configTargetVersion: semver.NewFromInt(1, 0, 0, ""),
			force:               true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			actions := []applyAction{}
			actionFactory := newActionFactory(nil, tc.lister, &action.Configuration{}, logger.NewTest(t))

			err := actionFactory.appendNewAction(tc.release, tc.configTargetVersion, tc.force, tc.allowDestructive, time.Second, &actions)
			if tc.wantErr {
				assert.Error(err)
				if tc.assertErr != nil {
					tc.assertErr(assert, err)
				}
				return
			}
			assert.NoError(err)
			assert.Len(actions, 1) // no error == actions gets appended
		})
	}
}

type stubLister struct {
	err     error
	version semver.Semver
}

func (s stubLister) currentVersion(_ string) (semver.Semver, error) {
	return s.version, s.err
}
