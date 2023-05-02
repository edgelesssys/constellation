/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package helm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/edgelesssys/constellation/v2/cli/internal/clusterid"
	"github.com/edgelesssys/constellation/v2/internal/attestation/idkeydigest"
	"github.com/edgelesssys/constellation/v2/internal/compatibility"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/deploy/helm"
	"github.com/edgelesssys/constellation/v2/internal/file"
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

// Client handles interaction with helm and the cluster.
type Client struct {
	config  *action.Configuration
	kubectl crdClient
	fs      file.Handler
	actions actionWrapper
	log     debugLog
}

// NewClient returns a new initializes client for the namespace Client.
func NewClient(client crdClient, kubeConfigPath, helmNamespace string, log debugLog) (*Client, error) {
	settings := cli.New()
	settings.KubeConfig = kubeConfigPath // constants.AdminConfFilename

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

	return &Client{kubectl: client, fs: fileHandler, actions: actions{config: actionConfig}, log: log}, nil
}

func (c *Client) shouldUpgrade(releaseName, newVersion string) error {
	currentVersion, err := c.currentVersion(releaseName)
	if err != nil {
		return fmt.Errorf("getting version for %s: %w", releaseName, err)
	}
	c.log.Debugf("Current %s version: %s", releaseName, currentVersion)
	c.log.Debugf("New %s version: %s", releaseName, newVersion)

	// This may break for cert-manager or cilium if we decide to upgrade more than one minor version at a time.
	// Leaving it as is since it is not clear to me what kind of sanity check we could do.
	if err := compatibility.IsValidUpgrade(currentVersion, newVersion); err != nil {
		return err
	}
	// at this point we conclude that the release should be upgraded. check that this CLI supports the upgrade.
	if releaseName == constellationOperatorsInfo.releaseName || releaseName == constellationServicesInfo.releaseName {
		if compatibility.EnsurePrefixV(constants.VersionInfo()) != compatibility.EnsurePrefixV(newVersion) {
			return fmt.Errorf("this CLI only supports microservice version %s for upgrading", constants.VersionInfo())
		}
	}
	c.log.Debugf("Upgrading %s from %s to %s", releaseName, currentVersion, newVersion)

	return nil
}

// Upgrade runs a helm-upgrade on all deployments that are managed via Helm.
// If the CLI receives an interrupt signal it will cancel the context.
// Canceling the context will prompt helm to abort and roll back the ongoing upgrade.
func (c *Client) Upgrade(ctx context.Context, config *config.Config, timeout time.Duration, allowDestructive bool) error {
	upgradeErrs := []error{}
	upgradeReleases := []*chart.Chart{}
	invalidUpgrade := &compatibility.InvalidUpgradeError{}

	for _, info := range []chartInfo{ciliumInfo, certManagerInfo, constellationOperatorsInfo, constellationServicesInfo} {
		chart, err := loadChartsDir(helmFS, info.path)
		if err != nil {
			return fmt.Errorf("loading chart: %w", err)
		}

		// define target version the chart is upgraded to
		var upgradeVersion string
		if info == constellationOperatorsInfo || info == constellationServicesInfo {
			// ensure that the services chart has the same version as the CLI
			updateVersions(chart, compatibility.EnsurePrefixV(constants.VersionInfo()))
			upgradeVersion = config.MicroserviceVersion
		} else {
			upgradeVersion = chart.Metadata.Version
		}

		err = c.shouldUpgrade(info.releaseName, upgradeVersion)
		switch {
		case errors.As(err, &invalidUpgrade):
			upgradeErrs = append(upgradeErrs, fmt.Errorf("skipping %s upgrade: %w", info.releaseName, err))
		case err != nil:
			return fmt.Errorf("should upgrade %s: %w", info.releaseName, err)
		case err == nil:
			upgradeReleases = append(upgradeReleases, chart)
		}
	}

	if len(upgradeReleases) == 0 {
		return errors.Join(upgradeErrs...)
	}

	crds, err := c.backupCRDs(ctx)
	if err != nil {
		return fmt.Errorf("creating CRD backup: %w", err)
	}
	if err := c.backupCRs(ctx, crds); err != nil {
		return fmt.Errorf("creating CR backup: %w", err)
	}

	fileHandler := file.NewHandler(afero.NewOsFs())
	for _, chart := range upgradeReleases {
		err = c.upgradeRelease(ctx, timeout, config, chart, allowDestructive, fileHandler)
		if err != nil {
			return fmt.Errorf("upgrading %s: %w", chart.Metadata.Name, err)
		}
	}

	return errors.Join(upgradeErrs...)
}

// Versions queries the cluster for running versions and returns a map of releaseName -> version.
func (c *Client) Versions() (ServiceVersions, error) {
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

	return ServiceVersions{
		cilium:                 compatibility.EnsurePrefixV(ciliumVersion),
		certManager:            compatibility.EnsurePrefixV(certManagerVersion),
		constellationOperators: compatibility.EnsurePrefixV(operatorsVersion),
		constellationServices:  compatibility.EnsurePrefixV(servicesVersion),
	}, nil
}

// currentVersion returns the version of the currently installed helm release.
func (c *Client) currentVersion(release string) (string, error) {
	rel, err := c.actions.listAction(release)
	if err != nil {
		return "", err
	}

	if len(rel) == 0 {
		return "", fmt.Errorf("release %s not found", release)
	}
	if len(rel) > 1 {
		return "", fmt.Errorf("multiple releases found for %s", release)
	}

	if rel[0] == nil || rel[0].Chart == nil || rel[0].Chart.Metadata == nil {
		return "", fmt.Errorf("received invalid release %s", release)
	}

	return rel[0].Chart.Metadata.Version, nil
}

// ServiceVersions bundles the versions of all services that are part of Constellation.
type ServiceVersions struct {
	cilium                 string
	certManager            string
	constellationOperators string
	constellationServices  string
}

// NewServiceVersions returns a new ServiceVersions struct.
func NewServiceVersions(cilium, certManager, constellationOperators, constellationServices string) ServiceVersions {
	return ServiceVersions{
		cilium:                 cilium,
		certManager:            certManager,
		constellationOperators: constellationOperators,
		constellationServices:  constellationServices,
	}
}

// Cilium returns the version of the Cilium release.
func (s ServiceVersions) Cilium() string {
	return s.cilium
}

// CertManager returns the version of the cert-manager release.
func (s ServiceVersions) CertManager() string {
	return s.certManager
}

// ConstellationOperators returns the version of the constellation-operators release.
func (s ServiceVersions) ConstellationOperators() string {
	return s.constellationOperators
}

// ConstellationServices returns the version of the constellation-services release.
func (s ServiceVersions) ConstellationServices() string {
	return s.constellationServices
}

func (c *Client) upgradeRelease(
	ctx context.Context, timeout time.Duration, conf *config.Config, chart *chart.Chart, allowDestructive bool, fileHandler file.Handler,
) error {
	// We need to load all values that can be statically loaded before merging them with the cluster
	// values. Otherwise the templates are not rendered correctly.
	k8sVersion, err := versions.NewValidK8sVersion(conf.KubernetesVersion)
	if err != nil {
		return fmt.Errorf("invalid k8s version: %w", err)
	}
	loader := NewLoader(conf.GetProvider(), k8sVersion)

	var values map[string]any
	var releaseName string

	switch chart.Metadata.Name {
	case ciliumInfo.chartName:
		releaseName = ciliumInfo.releaseName
		values, err = loader.loadCiliumValues()
		if err != nil {
			return fmt.Errorf("loading values: %w", err)
		}
	case certManagerInfo.chartName:
		releaseName = certManagerInfo.releaseName
		values = loader.loadCertManagerValues()

		if !allowDestructive {
			return ErrConfirmationMissing
		}
	case constellationOperatorsInfo.chartName:
		releaseName = constellationOperatorsInfo.releaseName
		values, err = loader.loadOperatorsValues()
		if err != nil {
			return fmt.Errorf("loading values: %w", err)
		}

		if err := c.updateCRDs(ctx, chart); err != nil {
			return fmt.Errorf("updating CRDs: %w", err)
		}
	case constellationServicesInfo.chartName:
		releaseName = constellationServicesInfo.releaseName
		values, err = loader.loadConstellationServicesValues()
		if err != nil {
			return fmt.Errorf("loading values: %w", err)
		}

		if err := c.applyMigrations(releaseName, values, conf, fileHandler); err != nil {
			return fmt.Errorf("applying migrations: %w", err)
		}
	default:
		return fmt.Errorf("unknown chart name: %s", chart.Metadata.Name)
	}

	values, err = c.prepareValues(values, releaseName)
	if err != nil {
		return fmt.Errorf("preparing values: %w", err)
	}

	err = c.actions.upgradeAction(ctx, releaseName, chart, values, timeout)
	if err != nil {
		return err
	}

	return nil
}

// applyMigrations checks the from version and applies the necessary migrations.
// The function assumes the caller has verified that our version drift restriction is not violated,
// Currently, this is done during config validation.
func (c *Client) applyMigrations(releaseName string, values map[string]any, conf *config.Config, fileHandler file.Handler) error {
	current, err := c.currentVersion(releaseName)
	if err != nil {
		return fmt.Errorf("getting %s version: %w", releaseName, err)
	}
	currentV, err := semver.New(current)
	if err != nil {
		return fmt.Errorf("parsing current version: %w", err)
	}

	if currentV.Major == 2 && currentV.Minor == 6 {
		return migrateFrom2_6(values, conf, fileHandler)
	}

	return nil
}

// migrateFrom2_6 applies the necessary migrations for upgrading from v2.6.x to v2.7.x.
// migrateFrom2_6 should be applied for v2.6.x --> v2.7.x.
// migrateFrom2_6 should NOT be applied for v2.7.0 --> v2.7.x.
// This function can be removed once we are sure that we will no longer provide backports for v2.6.
func migrateFrom2_6(values map[string]any, conf *config.Config, fileHandler file.Handler) error {
	// Manually setting attestationVariant is required here since upgrade normally isn't allowed to change this value.
	// However, to introduce the value into a 2.6 cluster for the first time we have to set it nevertheless.
	if err := setAttestationVariant(values, conf.AttestationVariant); err != nil {
		return fmt.Errorf("setting attestationVariant: %w", err)
	}

	// Manually setting idKeyConfig is required here since upgrade normally isn't allowed to change this value.
	// However, to introduce the value into a 2.6 cluster for the first time we have to set it nevertheless.
	var idFile clusterid.File
	if err := fileHandler.ReadJSON(constants.ClusterIDsFileName, &idFile); err != nil {
		return fmt.Errorf("reading cluster ID file: %w", err)
	}
	// Disallow users to set MAAFallback as ID key digest policy for upgrades, since it requires extra cloud resources.
	if conf.IDKeyDigestPolicy() == idkeydigest.MAAFallback {
		return fmt.Errorf("ID key digest policy %s is not supported for upgrades", conf.IDKeyDigestPolicy())
	}
	if err := setIdkeyConfig(values, conf, idFile.AttestationURL); err != nil {
		return fmt.Errorf("setting id key config: %w", err)
	}
	return nil
}

// prepareValues returns a values map as required for helm-upgrade.
// It imitates the behaviour of helm's reuse-values flag by fetching the current values from the cluster
// and merging the fetched values with the locally found values.
// This is done to ensure that new values (from upgrades of the local files) end up in the cluster.
// reuse-values does not ensure this.
func (c *Client) prepareValues(localValues map[string]any, releaseName string) (map[string]any, error) {
	// Ensure installCRDs is set for cert-manager chart.
	if releaseName == certManagerInfo.releaseName {
		localValues["installCRDs"] = true
	}
	clusterValues, err := c.actions.getValues(releaseName)
	if err != nil {
		return nil, fmt.Errorf("getting values for %s: %w", releaseName, err)
	}

	return helm.MergeMaps(clusterValues, localValues), nil
}

// GetValues queries the cluster for the values of the given release.
func (c *Client) GetValues(release string) (map[string]any, error) {
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
func (c *Client) updateCRDs(ctx context.Context, chart *chart.Chart) error {
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

// setAttestationVariant sets the attesationVariant value on verification-service and join-service value maps.
func setAttestationVariant(values map[string]any, variant string) error {
	joinServiceVals, ok := values["join-service"].(map[string]any)
	if !ok {
		return errors.New("invalid join-service values")
	}
	joinServiceVals["attestationVariant"] = variant

	verifyServiceVals, ok := values["verification-service"].(map[string]any)
	if !ok {
		return errors.New("invalid verification-service values")
	}
	verifyServiceVals["attestationVariant"] = variant

	return nil
}

// setIdkeyConfig sets the idkeyconfig value on the join-service value maps.
func setIdkeyConfig(values map[string]any, config *config.Config, maaURL string) error {
	joinServiceVals, ok := values["join-service"].(map[string]any)
	if !ok {
		return errors.New("invalid join-service values")
	}

	idKeyCfg := idkeydigest.Config{
		IDKeyDigests:      config.IDKeyDigests(),
		EnforcementPolicy: config.IDKeyDigestPolicy(),
		MAAURL:            maaURL,
	}
	marshalledCfg, err := json.Marshal(idKeyCfg)
	if err != nil {
		return fmt.Errorf("marshalling id key digest config: %w", err)
	}
	joinServiceVals["idKeyConfig"] = string(marshalledCfg)

	return nil
}

type debugLog interface {
	Debugf(format string, args ...any)
	Sync()
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
