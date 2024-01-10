/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package constellation

import (
	"context"
	"fmt"

	"github.com/edgelesssys/constellation/v2/internal/atls"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/constellation/helm"
	"github.com/edgelesssys/constellation/v2/internal/constellation/kubecmd"
	"github.com/edgelesssys/constellation/v2/internal/crypto"
	"github.com/edgelesssys/constellation/v2/internal/grpc/dialer"
	"github.com/edgelesssys/constellation/v2/internal/kms/uri"
	"github.com/edgelesssys/constellation/v2/internal/license"
)

// ApplyContext denotes the context in which the apply command is run.
type ApplyContext string

const (
	// ApplyContextCLI is used when the Applier is used by the CLI.
	ApplyContextCLI ApplyContext = "cli"
	// ApplyContextTerraform is used when the Applier is used by Terraform.
	ApplyContextTerraform ApplyContext = "terraform"
)

// An Applier handles applying a specific configuration to a Constellation cluster
// with existing Infrastructure.
// In Particular, this involves Initialization and Upgrading of the cluster.
type Applier struct {
	log            debugLog
	licenseChecker licenseChecker
	spinner        spinnerInterf

	applyContext ApplyContext

	// newDialer creates a new aTLS gRPC dialer.
	newDialer     func(validator atls.Validator) *dialer.Dialer
	kubecmdClient kubecmdClient
	helmClient    helmApplier
}

type licenseChecker interface {
	CheckLicense(ctx context.Context, csp cloudprovider.Provider, action license.Action, licenseID string) (int, error)
}

type debugLog interface {
	Debug(format string, args ...any)
}

// NewApplier creates a new Applier.
func NewApplier(
	log debugLog, spinner spinnerInterf,
	applyContext ApplyContext,
	newDialer func(validator atls.Validator) *dialer.Dialer,
) *Applier {
	return &Applier{
		log:            log,
		spinner:        spinner,
		licenseChecker: license.NewChecker(),
		applyContext:   applyContext,
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
func (a *Applier) CheckLicense(ctx context.Context, csp cloudprovider.Provider, initRequest bool, licenseID string) (int, error) {
	a.log.Debug("Contacting license server for license '%s'", licenseID)

	var action license.Action
	if initRequest {
		action = license.Init
	} else {
		action = license.Apply
	}
	if a.applyContext == ApplyContextTerraform {
		action += "-terraform"
	}

	quota, err := a.licenseChecker.CheckLicense(ctx, csp, action, licenseID)
	if err != nil {
		return 0, fmt.Errorf("checking license: %w", err)
	}
	a.log.Debug("Got response from license server for license '%s'", licenseID)

	return quota, nil
}

// GenerateMasterSecret generates a new master secret.
func (a *Applier) GenerateMasterSecret() (uri.MasterSecret, error) {
	a.log.Debug("Generating master secret")
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
	a.log.Debug("Generated master secret key and salt values")
	return secret, nil
}

// GenerateMeasurementSalt generates a new measurement salt.
func (a *Applier) GenerateMeasurementSalt() ([]byte, error) {
	a.log.Debug("Generating measurement salt")
	measurementSalt, err := crypto.GenerateRandomBytes(crypto.RNGLengthDefault)
	if err != nil {
		return nil, fmt.Errorf("generating measurement salt: %w", err)
	}
	a.log.Debug("Generated measurement salt")
	return measurementSalt, nil
}
