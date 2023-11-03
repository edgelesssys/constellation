/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package helm

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/edgelesssys/constellation/v2/internal/compatibility"
	"github.com/edgelesssys/constellation/v2/internal/semver"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
)

// ErrConfirmationMissing signals that an action requires user confirmation.
var ErrConfirmationMissing = errors.New("action requires user confirmation")

var errReleaseNotFound = errors.New("release not found")

type actionFactory struct {
	versionLister releaseVersionLister
	cfg           *action.Configuration
	kubeClient    crdClient
	log           debugLog
}

type crdClient interface {
	ApplyCRD(ctx context.Context, rawCRD []byte) error
}

// newActionFactory creates a new action factory for managing helm releases.
func newActionFactory(kubeClient crdClient, lister releaseVersionLister, actionConfig *action.Configuration, log debugLog) *actionFactory {
	return &actionFactory{
		versionLister: lister,
		cfg:           actionConfig,
		kubeClient:    kubeClient,
		log:           log,
	}
}

// GetActions returns a list of actions to apply the given releases.
func (a actionFactory) GetActions(releases []Release, configTargetVersion semver.Semver, force, allowDestructive bool) (actions []applyAction, includesUpgrade bool, err error) {
	upgradeErrs := []error{}
	for _, release := range releases {
		err := a.appendNewAction(release, configTargetVersion, force, allowDestructive, &actions)
		var invalidUpgrade *compatibility.InvalidUpgradeError
		if errors.As(err, &invalidUpgrade) {
			upgradeErrs = append(upgradeErrs, err)
			continue
		}
		if err != nil {
			return actions, includesUpgrade, fmt.Errorf("creating action for %s: %w", release.ReleaseName, err)
		}
	}
	for _, action := range actions {
		if _, ok := action.(*upgradeAction); ok {
			includesUpgrade = true
			break
		}
	}
	return actions, includesUpgrade, errors.Join(upgradeErrs...)
}

func (a actionFactory) appendNewAction(release Release, configTargetVersion semver.Semver, force, allowDestructive bool, actions *[]applyAction) error {
	newVersion, err := semver.New(release.Chart.Metadata.Version)
	if err != nil {
		return fmt.Errorf("parsing chart version: %w", err)
	}
	cliSupportsConfigVersion := configTargetVersion.Compare(newVersion) != 0

	currentVersion, err := a.versionLister.currentVersion(release.ReleaseName)
	if errors.Is(err, errReleaseNotFound) {
		// Don't install a new release if the user's config specifies a different version than the CLI offers.
		if !force && isCLIVersionedRelease(release.ReleaseName) && cliSupportsConfigVersion {
			return compatibility.NewInvalidUpgradeError(
				currentVersion.String(),
				configTargetVersion.String(),
				fmt.Errorf("this CLI only supports installing microservice version %s", newVersion),
			)
		}

		a.log.Debugf("Release %s not found, adding to new releases...", release.ReleaseName)
		*actions = append(*actions, a.newInstall(release))
		return nil
	}
	if err != nil {
		return fmt.Errorf("getting version for %s: %w", release.ReleaseName, err)
	}
	a.log.Debugf("Current %s version: %s", release.ReleaseName, currentVersion)
	a.log.Debugf("New %s version: %s", release.ReleaseName, newVersion)

	if !force {
		// For charts we package ourselves, the version is equal to the CLI version (charts are embedded in the binary).
		// We need to make sure this matches with the version in a user's config, if an upgrade should be applied.
		if isCLIVersionedRelease(release.ReleaseName) {
			// If target version is not a valid upgrade, don't upgrade any charts.
			if err := configTargetVersion.IsUpgradeTo(currentVersion); err != nil {
				return fmt.Errorf("invalid upgrade for %s: %w", release.ReleaseName, err)
			}
			// Target version is newer than current version, so we should perform an upgrade.
			// Now make sure the target version is equal to the the CLI version.
			if cliSupportsConfigVersion {
				return compatibility.NewInvalidUpgradeError(
					currentVersion.String(),
					configTargetVersion.String(),
					fmt.Errorf("this CLI only supports upgrading to microservice version %s", newVersion),
				)
			}
		} else {
			// This may break for external chart dependencies if we decide to upgrade more than one minor version at a time.
			if err := newVersion.IsUpgradeTo(currentVersion); err != nil {
				return fmt.Errorf("invalid upgrade for %s: %w", release.ReleaseName, err)
			}
		}
	}

	if !allowDestructive &&
		release.ReleaseName == certManagerInfo.releaseName {
		return ErrConfirmationMissing
	}
	a.log.Debugf("Upgrading %s from %s to %s", release.ReleaseName, currentVersion, newVersion)
	*actions = append(*actions, a.newUpgrade(release))
	return nil
}

func (a actionFactory) newInstall(release Release) *installAction {
	action := &installAction{helmAction: newHelmInstallAction(a.cfg, release), release: release, log: a.log}
	return action
}

func (a actionFactory) newUpgrade(release Release) *upgradeAction {
	action := &upgradeAction{helmAction: newHelmUpgradeAction(a.cfg), release: release, log: a.log}
	if release.ReleaseName == constellationOperatorsInfo.releaseName {
		action.preUpgrade = func(ctx context.Context) error {
			if err := a.updateCRDs(ctx, release.Chart); err != nil {
				return fmt.Errorf("updating operator CRDs: %w", err)
			}
			return nil
		}
	}
	return action
}

// updateCRDs walks through the dependencies of the given chart and applies
// the files in the dependencie's 'crds' folder.
// This function is NOT recursive!
func (a actionFactory) updateCRDs(ctx context.Context, chart *chart.Chart) error {
	for _, dep := range chart.Dependencies() {
		for _, crdFile := range dep.Files {
			if strings.HasPrefix(crdFile.Name, "crds/") {
				a.log.Debugf("Updating crd: %s", crdFile.Name)
				err := a.kubeClient.ApplyCRD(ctx, crdFile.Data)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// isCLIVersionedRelease checks if the given release is versioned by the CLI,
// meaning that the version of the Helm release is equal to the version of the CLI that installed it.
func isCLIVersionedRelease(releaseName string) bool {
	return releaseName == constellationOperatorsInfo.releaseName ||
		releaseName == constellationServicesInfo.releaseName ||
		releaseName == csiInfo.releaseName
}

// releaseVersionLister can list the versions of a helm release.
type releaseVersionLister interface {
	currentVersion(release string) (semver.Semver, error)
}
