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

	"github.com/edgelesssys/constellation/v2/cli/internal/clusterid"
	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/compatibility"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/kms/uri"
	"github.com/edgelesssys/constellation/v2/internal/semver"
	"github.com/edgelesssys/constellation/v2/internal/versions"
	"github.com/spf13/afero"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/release"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	// AllowDestructive is a named bool to signal that destructive actions have been confirmed by the user.
	AllowDestructive = true
	// DenyDestructive is a named bool to signal that destructive actions have not been confirmed by the user yet.
	DenyDestructive = false
)

// ErrConfirmationMissing signals that an action requires user confirmation.
var ErrConfirmationMissing = errors.New("action requires user confirmation")

var errReleaseNotFound = errors.New("release not found")

// UpgradeClient handles interaction with helm and the cluster.
type UpgradeClient struct {
	config           *action.Configuration
	kubectl          crdClient
	fs               file.Handler
	actions          actionWrapper
	upgradeWorkspace string
	log              debugLog
}

// NewUpgradeClient returns a newly initialized UpgradeClient for the given namespace.
func NewUpgradeClient(client crdClient, upgradeWorkspace, kubeConfigPath, helmNamespace string, log debugLog) (*UpgradeClient, error) {
	settings := cli.New()
	settings.KubeConfig = kubeConfigPath

	actionConfig := &action.Configuration{}
	if err := actionConfig.Init(settings.RESTClientGetter(), helmNamespace, "secret", log.Debugf); err != nil {
		return nil, fmt.Errorf("initializing config: %w", err)
	}

	fileHandler := file.NewHandler(afero.NewOsFs())

	kubeconfig, err := fileHandler.Read(kubeConfigPath)
	if err != nil {
		return nil, fmt.Errorf("reading gce config: %w", err)
	}

	if err := client.Initialize(kubeconfig); err != nil {
		return nil, fmt.Errorf("initializing kubectl: %w", err)
	}

	return &UpgradeClient{
		kubectl:          client,
		fs:               fileHandler,
		actions:          actions{config: actionConfig},
		upgradeWorkspace: upgradeWorkspace,
		log:              log,
	}, nil
}

func (c *UpgradeClient) shouldUpgrade(releaseName string, newVersion semver.Semver, force bool) error {
	currentVersion, err := c.currentVersion(releaseName)
	if err != nil {
		return fmt.Errorf("getting version for %s: %w", releaseName, err)
	}
	c.log.Debugf("Current %s version: %s", releaseName, currentVersion)
	c.log.Debugf("New %s version: %s", releaseName, newVersion)

	// This may break for cert-manager or cilium if we decide to upgrade more than one minor version at a time.
	// Leaving it as is since it is not clear to me what kind of sanity check we could do.
	if !force {
		if err := newVersion.IsUpgradeTo(currentVersion); err != nil {
			return err
		}
	}

	// at this point we conclude that the release should be upgraded. check that this CLI supports the upgrade.
	cliVersion := constants.BinaryVersion()
	if isCLIVersionedRelease(releaseName) && cliVersion.Compare(newVersion) != 0 {
		return fmt.Errorf("this CLI only supports microservice version %s for upgrading", cliVersion.String())
	}
	c.log.Debugf("Upgrading %s from %s to %s", releaseName, currentVersion, newVersion)

	return nil
}

// Upgrade runs a helm-upgrade on all deployments that are managed via Helm.
// If the CLI receives an interrupt signal it will cancel the context.
// Canceling the context will prompt helm to abort and roll back the ongoing upgrade.
func (c *UpgradeClient) Upgrade(ctx context.Context, config *config.Config, idFile clusterid.File, timeout time.Duration,
	allowDestructive, force bool, upgradeID string, conformance bool, helmWaitMode WaitMode, masterSecret uri.MasterSecret,
	serviceAccURI string, validK8sVersion versions.ValidK8sVersion, output terraform.ApplyOutput,
) error {
	upgradeErrs := []error{}
	upgradeReleases := []Release{}
	newReleases := []Release{}

	clusterName := clusterid.GetClusterName(config, idFile)
	helmLoader := NewLoader(config.GetProvider(), validK8sVersion, clusterName)
	c.log.Debugf("Created new Helm loader")
	releases, err := helmLoader.LoadReleases(config, conformance, helmWaitMode, masterSecret, serviceAccURI, idFile, output)
	if err != nil {
		return fmt.Errorf("loading releases: %w", err)
	}
	for _, release := range getManagedReleases(config, releases) {
		var invalidUpgrade *compatibility.InvalidUpgradeError
		// Get version of the chart embedded in the CLI
		// This is the version we are upgrading to
		// Since our bundled charts are embedded with version 0.0.0,
		// we need to update them to the same version as the CLI
		var upgradeVersion semver.Semver
		if isCLIVersionedRelease(release.ReleaseName) {
			updateVersions(release.Chart, constants.BinaryVersion())
			upgradeVersion = config.MicroserviceVersion
		} else {
			chartVersion, err := semver.New(release.Chart.Metadata.Version)
			if err != nil {
				return fmt.Errorf("parsing chart version: %w", err)
			}
			upgradeVersion = chartVersion
		}
		err = c.shouldUpgrade(release.ReleaseName, upgradeVersion, force)
		switch {
		case errors.Is(err, errReleaseNotFound):
			// if the release is not found, we need to install it
			c.log.Debugf("Release %s not found, adding to new releases...", release.ReleaseName)
			newReleases = append(newReleases, release)
		case errors.As(err, &invalidUpgrade):
			c.log.Debugf("Appending to %s upgrade: %s", release.ReleaseName, err)
			upgradeReleases = append(upgradeReleases, release)
		case err != nil:
			c.log.Debugf("Adding %s to upgrade releases...", release.ReleaseName)
			return fmt.Errorf("should upgrade %s: %w", release.ReleaseName, err)
		case err == nil:
			upgradeReleases = append(upgradeReleases, release)

			// Check if installing/upgrading the chart could be destructive
			// If so, we don't want to perform any actions,
			// unless the user confirms it to be OK.
			if !allowDestructive &&
				release.ReleaseName == certManagerInfo.releaseName {
				return ErrConfirmationMissing
			}
		}
	}

	// Backup CRDs and CRs if we are upgrading anything.
	if len(upgradeReleases) != 0 {
		c.log.Debugf("Creating backup of CRDs and CRs")
		crds, err := c.backupCRDs(ctx, upgradeID)
		if err != nil {
			return fmt.Errorf("creating CRD backup: %w", err)
		}
		if err := c.backupCRs(ctx, crds, upgradeID); err != nil {
			return fmt.Errorf("creating CR backup: %w", err)
		}
	}

	for _, release := range upgradeReleases {
		c.log.Debugf("Upgrading release %s", release.Chart.Metadata.Name)
		if release.ReleaseName == constellationOperatorsInfo.releaseName {
			if err := c.updateCRDs(ctx, release.Chart); err != nil {
				return fmt.Errorf("updating operator CRDs: %w", err)
			}
		}
		if err := c.upgradeRelease(ctx, timeout, release); err != nil {
			return fmt.Errorf("upgrading %s: %w", release.Chart.Metadata.Name, err)
		}
	}

	// Install new releases after upgrading existing ones.
	// This makes sure if a release was removed as a dependency from one chart,
	// and then added as a new standalone chart (or as a dependency of another chart),
	// that the new release is installed without creating naming conflicts.
	// If in the future, we require to install a new release before upgrading existing ones,
	// it should be done in a separate loop, instead of moving this one up.
	for _, release := range newReleases {
		c.log.Debugf("Installing new release %s", release.Chart.Metadata.Name)
		if err := c.installNewRelease(ctx, timeout, release); err != nil {
			return fmt.Errorf("upgrading %s: %w", release.Chart.Metadata.Name, err)
		}
	}

	return errors.Join(upgradeErrs...)
}

func getManagedReleases(config *config.Config, releases *Releases) []Release {
	res := []Release{releases.Cilium, releases.CertManager, releases.ConstellationOperators, releases.ConstellationServices}

	if config.GetProvider() == cloudprovider.AWS {
		res = append(res, *releases.AWSLoadBalancerController)
	}
	if config.DeployCSIDriver() {
		res = append(res, *releases.CSI)
	}
	return res
}

// Versions queries the cluster for running versions and returns a map of releaseName -> version.
func (c *UpgradeClient) Versions() (ServiceVersions, error) {
	ciliumVersion, err := c.currentVersion(ciliumInfo.releaseName)
	if err != nil {
		return ServiceVersions{}, fmt.Errorf("getting %s version: %w", ciliumInfo.releaseName, err)
	}
	certManagerVersion, err := c.currentVersion(certManagerInfo.releaseName)
	if err != nil {
		return ServiceVersions{}, fmt.Errorf("getting %s version: %w", certManagerInfo.releaseName, err)
	}
	operatorsVersion, err := c.currentVersion(constellationOperatorsInfo.releaseName)
	if err != nil {
		return ServiceVersions{}, fmt.Errorf("getting %s version: %w", constellationOperatorsInfo.releaseName, err)
	}
	servicesVersion, err := c.currentVersion(constellationServicesInfo.releaseName)
	if err != nil {
		return ServiceVersions{}, fmt.Errorf("getting %s version: %w", constellationServicesInfo.releaseName, err)
	}
	csiVersions, err := c.csiVersions()
	if err != nil {
		return ServiceVersions{}, fmt.Errorf("getting CSI versions: %w", err)
	}

	serviceVersions := ServiceVersions{
		cilium:                 ciliumVersion,
		certManager:            certManagerVersion,
		constellationOperators: operatorsVersion,
		constellationServices:  servicesVersion,
		csiVersions:            csiVersions,
	}

	if awsLBVersion, err := c.currentVersion(awsLBControllerInfo.releaseName); err == nil {
		serviceVersions.awsLBController = awsLBVersion
	} else if !errors.Is(err, errReleaseNotFound) {
		return ServiceVersions{}, fmt.Errorf("getting %s version: %w", awsLBControllerInfo.releaseName, err)
	}

	return serviceVersions, nil
}

// currentVersion returns the version of the currently installed helm release.
func (c *UpgradeClient) currentVersion(release string) (semver.Semver, error) {
	rel, err := c.actions.listAction(release)
	if err != nil {
		return semver.Semver{}, err
	}

	if len(rel) == 0 {
		return semver.Semver{}, errReleaseNotFound
	}
	if len(rel) > 1 {
		return semver.Semver{}, fmt.Errorf("multiple releases found for %s", release)
	}

	if rel[0] == nil || rel[0].Chart == nil || rel[0].Chart.Metadata == nil {
		return semver.Semver{}, fmt.Errorf("received invalid release %s", release)
	}

	return semver.New(rel[0].Chart.Metadata.Version)
}

func (c *UpgradeClient) csiVersions() (map[string]semver.Semver, error) {
	packedChartRelease, err := c.actions.listAction(csiInfo.releaseName)
	if err != nil {
		return nil, fmt.Errorf("listing %s: %w", csiInfo.releaseName, err)
	}

	csiVersions := make(map[string]semver.Semver)

	// No CSI driver installed
	if len(packedChartRelease) == 0 {
		return csiVersions, nil
	}

	if len(packedChartRelease) > 1 {
		return nil, fmt.Errorf("multiple releases found for %s", csiInfo.releaseName)
	}

	if packedChartRelease[0] == nil || packedChartRelease[0].Chart == nil {
		return nil, fmt.Errorf("received invalid release %s", csiInfo.releaseName)
	}

	dependencies := packedChartRelease[0].Chart.Metadata.Dependencies
	for _, dep := range dependencies {
		var err error
		csiVersions[dep.Name], err = semver.New(dep.Version)
		if err != nil {
			return nil, fmt.Errorf("parsing CSI version %q: %w", dep.Name, err)
		}
	}
	return csiVersions, nil
}

// installNewRelease installs a previously not installed release on the cluster.
func (c *UpgradeClient) installNewRelease(
	ctx context.Context, timeout time.Duration, release Release,
) error {
	return c.actions.installAction(ctx, release.ReleaseName, release.Chart, release.Values, timeout)
}

// upgradeRelease upgrades a release running on the cluster.
func (c *UpgradeClient) upgradeRelease(
	ctx context.Context, timeout time.Duration, release Release,
) error {
	return c.actions.upgradeAction(ctx, release.ReleaseName, release.Chart, release.Values, timeout)
}

// GetValues queries the cluster for the values of the given release.
func (c *UpgradeClient) GetValues(release string) (map[string]any, error) {
	client := action.NewGetValues(c.config)
	// Version corresponds to the releases revision. Specifying a Version <= 0 yields the latest release.
	client.Version = 0
	values, err := client.Run(release)
	if err != nil {
		return nil, fmt.Errorf("getting values for %s: %w", release, err)
	}
	return values, nil
}

// updateCRDs walks through the dependencies of the given chart and applies
// the files in the dependencie's 'crds' folder.
// This function is NOT recursive!
func (c *UpgradeClient) updateCRDs(ctx context.Context, chart *chart.Chart) error {
	for _, dep := range chart.Dependencies() {
		for _, crdFile := range dep.Files {
			if strings.HasPrefix(crdFile.Name, "crds/") {
				c.log.Debugf("Updating crd: %s", crdFile.Name)
				err := c.kubectl.ApplyCRD(ctx, crdFile.Data)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

type crdClient interface {
	Initialize(kubeconfig []byte) error
	ApplyCRD(ctx context.Context, rawCRD []byte) error
	GetCRDs(ctx context.Context) ([]apiextensionsv1.CustomResourceDefinition, error)
	GetCRs(ctx context.Context, gvr schema.GroupVersionResource) ([]unstructured.Unstructured, error)
}

type actionWrapper interface {
	listAction(release string) ([]*release.Release, error)
	getValues(release string) (map[string]any, error)
	installAction(ctx context.Context, releaseName string, chart *chart.Chart, values map[string]any, timeout time.Duration) error
	upgradeAction(ctx context.Context, releaseName string, chart *chart.Chart, values map[string]any, timeout time.Duration) error
}

type actions struct {
	config *action.Configuration
}

// listAction execute a List action by wrapping helm's action package.
// It creates the action, runs it at returns results and errors.
func (a actions) listAction(release string) ([]*release.Release, error) {
	action := action.NewList(a.config)
	action.Filter = release
	return action.Run()
}

func (a actions) getValues(release string) (map[string]any, error) {
	client := action.NewGetValues(a.config)
	// Version corresponds to the releases revision. Specifying a Version <= 0 yields the latest release.
	client.Version = 0
	return client.Run(release)
}

func (a actions) upgradeAction(ctx context.Context, releaseName string, chart *chart.Chart, values map[string]any, timeout time.Duration) error {
	action := action.NewUpgrade(a.config)
	action.Atomic = true
	action.Namespace = constants.HelmNamespace
	action.ReuseValues = false
	action.Timeout = timeout
	if _, err := action.RunWithContext(ctx, releaseName, chart, values); err != nil {
		return fmt.Errorf("upgrading %s: %w", releaseName, err)
	}
	return nil
}

func (a actions) installAction(ctx context.Context, releaseName string, chart *chart.Chart, values map[string]any, timeout time.Duration) error {
	action := action.NewInstall(a.config)
	action.Atomic = true
	action.Namespace = constants.HelmNamespace
	action.ReleaseName = releaseName
	action.Timeout = timeout
	if _, err := action.RunWithContext(ctx, chart, values); err != nil {
		return fmt.Errorf("installing previously not installed chart %s: %w", chart.Name(), err)
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
