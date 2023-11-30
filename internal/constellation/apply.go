/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package constellation

import (
	"context"
	"fmt"

	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/crypto"
	"github.com/edgelesssys/constellation/v2/internal/kms/uri"
	"github.com/edgelesssys/constellation/v2/internal/license"
)

// An Applier handles applying a specific configuration to a Constellation cluster
// with existing Infrastructure.
// In Particular, this involves Initialization and Upgrading of the cluster.
type Applier struct {
	log            debugLog
	licenseChecker *license.Checker
	spinner        spinnerInterf
}

type debugLog interface {
	Debugf(format string, args ...any)
}

// NewApplier creates a new Applier.
func NewApplier(log debugLog, spinner spinnerInterf) *Applier {
	return &Applier{
		log:            log,
		spinner:        spinner,
		licenseChecker: license.NewChecker(license.NewClient()),
	}
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
