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

const (
	// timeout is the maximum time given to the helm client.
	upgradeTimeout = 3 * time.Minute
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

	// TODO: Replace log.Printf with actual CLILogger during refactoring of upgrade cmd.
	actionConfig := &action.Configuration{}
	if err := actionConfig.Init(settings.RESTClientGetter(), helmNamespace, "secret", log.Debugf); err != nil {
		return nil, fmt.Errorf("initializing config: %w", err)
	}

	return &Client{config: actionConfig, crdClient: client, log: log}, nil
}

// CurrentVersion returns the version of the currently installed helm release.
func (c *Client) CurrentVersion(release string) (string, error) {
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

// Upgrade runs a helm-upgrade on all deployments that are managed via Helm.
// If the CLI receives an interrupt signal it will cancel the context.
// Canceling the context will prompt helm to abort and roll back the ongoing upgrade.
func (c *Client) Upgrade(ctx context.Context, config *config.Config) error {
	action := action.NewUpgrade(c.config)
	action.Atomic = true
	action.Namespace = constants.HelmNamespace
	action.ReuseValues = false
	action.Timeout = upgradeTimeout

	ciliumChart, err := loadChartsDir(helmFS, ciliumPath)
	if err != nil {
		return fmt.Errorf("loading cilium: %w", err)
	}
	certManagerChart, err := loadChartsDir(helmFS, certManagerPath)
	if err != nil {
		return fmt.Errorf("loading cert-manager: %w", err)
	}
	conOperatorChart, err := loadChartsDir(helmFS, conOperatorsPath)
	if err != nil {
		return fmt.Errorf("loading operators: %w", err)
	}
	conServicesChart, err := loadChartsDir(helmFS, conServicesPath)
	if err != nil {
		return fmt.Errorf("loading constellation-services chart: %w", err)
	}

	values, err := c.prepareValues(ciliumChart, ciliumReleaseName)
	if err != nil {
		return err
	}
	if _, err := action.RunWithContext(ctx, ciliumReleaseName, ciliumChart, values); err != nil {
		return fmt.Errorf("upgrading cilium: %w", err)
	}

	values, err = c.prepareValues(certManagerChart, certManagerReleaseName)
	if err != nil {
		return err
	}
	if _, err := action.RunWithContext(ctx, certManagerReleaseName, certManagerChart, values); err != nil {
		return fmt.Errorf("upgrading cert-manager: %w", err)
	}

	err = c.updateOperatorCRDs(ctx, conOperatorChart)
	if err != nil {
		return fmt.Errorf("updating operator CRDs: %w", err)
	}
	values, err = c.prepareValues(conOperatorChart, conOperatorsReleaseName)
	if err != nil {
		return err
	}
	if _, err := action.RunWithContext(ctx, conOperatorsReleaseName, conOperatorChart, values); err != nil {
		return fmt.Errorf("upgrading services: %w", err)
	}

	values, err = c.prepareValues(conServicesChart, conServicesReleaseName)
	if err != nil {
		return err
	}
	if _, err := action.RunWithContext(ctx, conServicesReleaseName, conServicesChart, values); err != nil {
		return fmt.Errorf("upgrading operators: %w", err)
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

// updateOperatorCRDs walks through the dependencies of the given chart and applies
// the files in the dependencie's 'crds' folder.
// This function is NOT recursive!
func (c *Client) updateOperatorCRDs(ctx context.Context, chart *chart.Chart) error {
	for _, dep := range chart.Dependencies() {
		for _, crdFile := range dep.Files {
			if strings.HasPrefix(crdFile.Name, "crds/") {
				c.log.Debugf("updating crd: %s", crdFile.Name)
				err := c.ApplyCRD(ctx, crdFile.Data)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// Using the type from cli/internal/cmd would introduce the following import cycle:
// cli/internal/cmd (upgradeexecute.go) -> cli/internal/cloudcmd (upgrade.go) ->
// -> cli/internal/cloudcmd/helm (client.go) -> cli/internal/cmd (log.go).
type debugLog interface {
	Debugf(format string, args ...any)
	Sync()
}
