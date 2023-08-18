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
	"time"

	"github.com/edgelesssys/constellation/v2/internal/compatibility"
	"github.com/edgelesssys/constellation/v2/internal/constants"
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
	cliVersion    semver.Semver
	log           debugLog
}

type crdClient interface {
	ApplyCRD(ctx context.Context, rawCRD []byte) error
}

// newActionFactory creates a new action factory for managing helm releases.
func newActionFactory(kubeClient crdClient, lister releaseVersionLister, actionConfig *action.Configuration, cliVersion semver.Semver, log debugLog) *actionFactory {
	return &actionFactory{
		cliVersion:    cliVersion,
		versionLister: lister,
		cfg:           actionConfig,
		kubeClient:    kubeClient,
		log:           log,
	}
}

// GetActions returns a list of actions to apply the given releases.
func (a actionFactory) GetActions(releases []Release, force, allowDestructive bool) (actions []applyAction, includesUpgrade bool, err error) {
	upgradeErrs := []error{}
	for _, release := range releases {
		err := a.appendNewAction(release, force, allowDestructive, &actions)
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

func (a actionFactory) appendNewAction(release Release, force, allowDestructive bool, actions *[]applyAction) error {
	newVersion, err := semver.New(release.Chart.Metadata.Version)
	if err != nil {
		return fmt.Errorf("parsing chart version: %w", err)
	}
	currentVersion, err := a.versionLister.currentVersion(release.ReleaseName)
	if errors.Is(err, errReleaseNotFound) {
		a.log.Debugf("Release %s not found, adding to new releases...", release.ReleaseName)
		*actions = append(*actions, a.newInstall(release))
		return nil
	}
	if err != nil {
		return fmt.Errorf("getting version for %s: %w", release.ReleaseName, err)
	}
	a.log.Debugf("Current %s version: %s", release.ReleaseName, currentVersion)
	a.log.Debugf("New %s version: %s", release.ReleaseName, newVersion)

	// This may break for cert-manager or cilium if we decide to upgrade more than one minor version at a time.
	// Leaving it as is since it is not clear to me what kind of sanity check we could do.
	if !force {
		if err := newVersion.IsUpgradeTo(currentVersion); err != nil {
			return fmt.Errorf("invalid upgrade for %s: %w", release.ReleaseName, err)
		}
	}

	// at this point we conclude that the release should be upgraded. check that this CLI supports the upgrade.
	if isCLIVersionedRelease(release.ReleaseName) && a.cliVersion.Compare(newVersion) != 0 {
		return fmt.Errorf("this CLI only supports microservice version %s for upgrading", a.cliVersion.String())
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
	if action.ReleaseName() == ciliumInfo.releaseName {
		action.postInstall = func(ctx context.Context) error {
			return ciliumPostInstall(ctx, a.log)
		}
	}
	return action
}

func ciliumPostInstall(ctx context.Context, log debugLog) error {
	log.Debugf("Waiting for Cilium to become ready")
	helper, err := newK8sCiliumHelper(constants.AdminConfFilename)
	if err != nil {
		return fmt.Errorf("creating Kubernetes client: %w", err)
	}
	timeToStartWaiting := time.Now()
	// TODO(3u13r): Reduce the timeout when we switched the package repository - this is only this high because we once
	// saw polling times of ~16 minutes when hitting a slow PoP from Fastly (GitHub's / ghcr.io CDN).
	if err := helper.WaitForDS(ctx, "kube-system", "cilium", log); err != nil {
		return fmt.Errorf("waiting for Cilium to become healthy: %w", err)
	}
	timeUntilFinishedWaiting := time.Since(timeToStartWaiting)
	log.Debugf("Cilium became healthy after %s", timeUntilFinishedWaiting.String())

	log.Debugf("Fix Cilium through restart")
	if err := helper.RestartDS("kube-system", "cilium"); err != nil {
		return fmt.Errorf("restarting Cilium: %w", err)
	}
	return nil
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
