/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package constellation

import (
	"context"
	"fmt"

	"github.com/edgelesssys/constellation/v2/internal/atls"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constellation/kubecmd"
	"github.com/edgelesssys/constellation/v2/internal/crypto"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/grpc/dialer"
	"github.com/edgelesssys/constellation/v2/internal/helm"
	"github.com/edgelesssys/constellation/v2/internal/kms/uri"
	"github.com/edgelesssys/constellation/v2/internal/license"
	"github.com/edgelesssys/constellation/v2/internal/semver"
	"github.com/edgelesssys/constellation/v2/internal/state"
	"github.com/edgelesssys/constellation/v2/internal/versions"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

// An Applier handles applying a specific configuration to a Constellation cluster
// with existing Infrastructure.
// In Particular, this involves Initialization and Upgrading of the cluster.
type Applier struct {
	log            debugLog
	licenseChecker licenseChecker
	spinner        spinnerInterf

	// newDialer creates a new aTLS gRPC dialer.
	newDialer     func(validator atls.Validator) *dialer.Dialer
	kubecmdClient kubecmdClient
	helmClient    helmApplier
}

type licenseChecker interface {
	CheckLicense(context.Context, cloudprovider.Provider, string) (license.QuotaCheckResponse, error)
}

type debugLog interface {
	Debugf(format string, args ...any)
}

// NewApplier creates a new Applier.
func NewApplier(
	log debugLog,
	spinner spinnerInterf,
	newDialer func(validator atls.Validator) *dialer.Dialer,
) *Applier {
	return &Applier{
		log:            log,
		spinner:        spinner,
		licenseChecker: license.NewChecker(license.NewClient()),
		newDialer:      newDialer,
	}
}

// SetKubeConfig sets the config file to use for creating Kubernetes clients.
func (a *Applier) SetKubeConfig(kubeConfig []byte) error {
	kubecmdClient, err := kubecmd.New(kubeConfig, a.log)
	if err != nil {
		return err
	}
	helmClient, err := helm.NewClient(kubeConfig, a.log)
	if err != nil {
		return err
	}
	a.kubecmdClient = kubecmdClient
	a.helmClient = helmClient
	return nil
}

// CheckLicense checks the given Constellation license with the license server
// and returns the allowed quota for the license.
func (a *Applier) CheckLicense(ctx context.Context, csp cloudprovider.Provider, licenseID string) (int, error) {
	a.log.Debugf("Contacting license server for license '%s'", licenseID)
	quotaResp, err := a.licenseChecker.CheckLicense(ctx, csp, licenseID)
	if err != nil {
		return 0, fmt.Errorf("checking license: %w", err)
	}
	a.log.Debugf("Got response from license server for license '%s'", licenseID)

	return quotaResp.Quota, nil
}

// GenerateMasterSecret generates a new master secret.
func (a *Applier) GenerateMasterSecret() (uri.MasterSecret, error) {
	a.log.Debugf("Generating master secret")
	key, err := crypto.GenerateRandomBytes(crypto.MasterSecretLengthDefault)
	if err != nil {
		return uri.MasterSecret{}, err
	}
	salt, err := crypto.GenerateRandomBytes(crypto.RNGLengthDefault)
	if err != nil {
		return uri.MasterSecret{}, err
	}
	secret := uri.MasterSecret{
		Key:  key,
		Salt: salt,
	}
	a.log.Debugf("Generated master secret key and salt values")
	return secret, nil
}

// GenerateMeasurementSalt generates a new measurement salt.
func (a *Applier) GenerateMeasurementSalt() ([]byte, error) {
	a.log.Debugf("Generating measurement salt")
	measurementSalt, err := crypto.GenerateRandomBytes(crypto.RNGLengthDefault)
	if err != nil {
		return nil, fmt.Errorf("generating measurement salt: %w", err)
	}
	a.log.Debugf("Generated measurement salt")
	return measurementSalt, nil
}

type helmApplier interface {
	PrepareApply(
		csp cloudprovider.Provider, attestationVariant variant.Variant, k8sVersion versions.ValidK8sVersion, microserviceVersion semver.Semver, stateFile *state.State,
		flags helm.Options, serviceAccURI string, masterSecret uri.MasterSecret, openStackCfg *config.OpenStackConfig,
	) (
		helm.Applier, bool, error)
}

type kubecmdClient interface {
	UpgradeNodeImage(ctx context.Context, imageVersion semver.Semver, imageReference string, force bool) error
	UpgradeKubernetesVersion(ctx context.Context, kubernetesVersion versions.ValidK8sVersion, force bool) error
	ExtendClusterConfigCertSANs(ctx context.Context, alternativeNames []string) error
	GetClusterAttestationConfig(ctx context.Context, variant variant.Variant) (config.AttestationCfg, error)
	ApplyJoinConfig(ctx context.Context, newAttestConfig config.AttestationCfg, measurementSalt []byte) error
	BackupCRs(ctx context.Context, fileHandler file.Handler, crds []apiextensionsv1.CustomResourceDefinition, upgradeDir string) error
	BackupCRDs(ctx context.Context, fileHandler file.Handler, upgradeDir string) ([]apiextensionsv1.CustomResourceDefinition, error)
}
