/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package helm

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/deploy/helm"
	"github.com/pkg/errors"
	"golang.org/x/mod/semver"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/cli"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
)

// Client handles interaction with helm.
type Client struct {
	config    *action.Configuration
	crdClient *apiextensionsclient.Clientset
	log       debugLog
}

// NewClient returns a new initializes client for the namespace Client.
func NewClient(kubeConfigPath, helmNamespace string, client *apiextensionsclient.Clientset, log debugLog) (*Client, error) {
	settings := cli.New()
	settings.KubeConfig = kubeConfigPath // constants.AdminConfFilename

	actionConfig := &action.Configuration{}
	if err := actionConfig.Init(settings.RESTClientGetter(), helmNamespace, "secret", log.Debugf); err != nil {
		return nil, fmt.Errorf("initializing config: %w", err)
	}

	return &Client{config: actionConfig, crdClient: client, log: log}, nil
}

// Upgrade runs a helm-upgrade on all deployments that are managed via Helm.
// If the CLI receives an interrupt signal it will cancel the context.
// Canceling the context will prompt helm to abort and roll back the ongoing upgrade.
func (c *Client) Upgrade(ctx context.Context, config *config.Config, timeout time.Duration) error {
	if err := c.upgradeRelease(ctx, timeout, ciliumPath, ciliumReleaseName, false); err != nil {
		return fmt.Errorf("upgrading cilium: %w", err)
	}

	if err := c.upgradeRelease(ctx, timeout, certManagerPath, certManagerReleaseName, false); err != nil {
		return fmt.Errorf("upgrading cert-manager: %w", err)
	}

	if err := c.upgradeRelease(ctx, timeout, conOperatorsPath, conOperatorsReleaseName, true); err != nil {
		return fmt.Errorf("upgrading constellation operators: %w", err)
	}

	if err := c.upgradeRelease(ctx, timeout, conServicesPath, conServicesReleaseName, false); err != nil {
		return fmt.Errorf("upgrading constellation-services: %w", err)
	}

	return nil
}

// currentVersion returns the version of the currently installed helm release.
func (c *Client) currentVersion(release string) (string, error) {
	client := action.NewList(c.config)
	client.Filter = release
	rel, err := client.Run()
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

func (c *Client) upgradeRelease(
	ctx context.Context, timeout time.Duration, chartPath, releaseName string, hasCRDs bool,
) error {
	chart, err := loadChartsDir(helmFS, chartPath)
	if err != nil {
		return fmt.Errorf("loading chart: %w", err)
	}
	currentVersion, err := c.currentVersion(releaseName)
	if err != nil {
		return fmt.Errorf("getting current version: %w", err)
	}
	c.log.Debugf("Current %s version: %s", releaseName, currentVersion)
	c.log.Debugf("New %s version: %s", releaseName, chart.Metadata.Version)

	if !isUpgrade(currentVersion, chart.Metadata.Version) {
		c.log.Debugf(
			"Skipping upgrade of %s: new version (%s) is not an upgrade for current version (%s)",
			releaseName, chart.Metadata.Version, currentVersion,
		)
		return nil
	}

	if hasCRDs {
		if err := c.updateCRDs(ctx, chart); err != nil {
			return fmt.Errorf("updating CRDs: %w", err)
		}
	}
	values, err := c.prepareValues(chart, releaseName)
	if err != nil {
		return fmt.Errorf("preparing values: %w", err)
	}

	c.log.Debugf("Upgrading %s from %s to %s", releaseName, currentVersion, chart.Metadata.Version)
	action := action.NewUpgrade(c.config)
	action.Atomic = true
	action.Namespace = constants.HelmNamespace
	action.ReuseValues = false
	action.Timeout = timeout
	if _, err := action.RunWithContext(ctx, releaseName, chart, values); err != nil {
		return fmt.Errorf("upgrading %s: %w", releaseName, err)
	}

	return nil
}

// prepareCertManagerValues returns a values map as required for helm-upgrade.
// It imitates the behaviour of helm's reuse-values flag by fetching the current values from the cluster
// and merging the fetched values with the locally found values.
// This is done to ensure that new values (from upgrades of the local files) end up in the cluster.
// reuse-values does not ensure this.
func (c *Client) prepareValues(chart *chart.Chart, releaseName string) (map[string]any, error) {
	// Ensure installCRDs is set for cert-manager chart.
	if releaseName == certManagerReleaseName {
		chart.Values["installCRDs"] = true
	}
	values, err := c.GetValues(releaseName)
	if err != nil {
		return nil, fmt.Errorf("getting values: %w", err)
	}
	return helm.MergeMaps(chart.Values, values), nil
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

// ApplyCRD updates the given CRD by parsing it, querying it's version from the cluster and finally updating it.
func (c *Client) ApplyCRD(ctx context.Context, rawCRD []byte) error {
	crd, err := parseCRD(rawCRD)
	if err != nil {
		return fmt.Errorf("parsing crds: %w", err)
	}

	clusterCRD, err := c.crdClient.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, crd.Name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("getting crd: %w", err)
	}
	crd.ResourceVersion = clusterCRD.ResourceVersion
	_, err = c.crdClient.ApiextensionsV1().CustomResourceDefinitions().Update(ctx, crd, metav1.UpdateOptions{})
	return err
}

// parseCRD takes a byte slice of data and tries to create a CustomResourceDefinition object from it.
func parseCRD(crdString []byte) (*v1.CustomResourceDefinition, error) {
	sch := runtime.NewScheme()
	_ = scheme.AddToScheme(sch)
	_ = v1.AddToScheme(sch)
	obj, groupVersionKind, err := serializer.NewCodecFactory(sch).UniversalDeserializer().Decode(crdString, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("decoding crd: %w", err)
	}
	if groupVersionKind.Kind == "CustomResourceDefinition" {
		return obj.(*v1.CustomResourceDefinition), nil
	}

	return nil, errors.New("parsed []byte, but did not find a CRD")
}

// updateCRDs walks through the dependencies of the given chart and applies
// the files in the dependencie's 'crds' folder.
// This function is NOT recursive!
func (c *Client) updateCRDs(ctx context.Context, chart *chart.Chart) error {
	for _, dep := range chart.Dependencies() {
		for _, crdFile := range dep.Files {
			if strings.HasPrefix(crdFile.Name, "crds/") {
				c.log.Debugf("Updating crd: %s", crdFile.Name)
				err := c.ApplyCRD(ctx, crdFile.Data)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// isUpgrade returns true if the new version is greater than the current version.
// Versions should adhere to the semver spec, but this function will prefix the versions with 'v' if they don't.
func isUpgrade(currentVersion, newVersion string) bool {
	if !strings.HasPrefix(currentVersion, "v") {
		currentVersion = "v" + currentVersion
	}
	if !strings.HasPrefix(newVersion, "v") {
		newVersion = "v" + newVersion
	}

	// If the current version is not a valid semver,
	// we cant compare it to the new version.
	// -> We can't upgrade.
	if !semver.IsValid(currentVersion) {
		return false
	}

	if semver.Compare(currentVersion, newVersion) < 0 {
		return true
	}
	return false
}

type debugLog interface {
	Debugf(format string, args ...any)
	Sync()
}
