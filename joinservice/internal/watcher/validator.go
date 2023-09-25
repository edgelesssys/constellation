/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package watcher

import (
	"context"
	"crypto/x509"
	"encoding/asn1"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/edgelesssys/constellation/v2/internal/atls"
	"github.com/edgelesssys/constellation/v2/internal/attestation/choose"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/joinservice/internal/certcache"
)

// Updatable implements an updatable atls.Validator.
type Updatable struct {
	log         *logger.Logger
	mux         sync.Mutex
	fileHandler file.Handler
	variant     variant.Variant
	cachedCerts *certcache.CachedCerts
	atls.Validator
}

// NewValidator initializes a new updatable validator and performs an initial update (aka. initialization).
func NewValidator(log *logger.Logger, variant variant.Variant, fileHandler file.Handler, cachedCerts *certcache.CachedCerts) (*Updatable, error) {
	u := &Updatable{
		log:         log,
		fileHandler: fileHandler,
		variant:     variant,
		cachedCerts: cachedCerts,
	}
	err := u.Update()

	return u, err
}

// Validate calls the validators Validate method, and prevents any updates during the call.
func (u *Updatable) Validate(ctx context.Context, attDoc []byte, nonce []byte) ([]byte, error) {
	u.mux.Lock()
	defer u.mux.Unlock()
	return u.Validator.Validate(ctx, attDoc, nonce)
}

// OID returns the validators Object Identifier.
func (u *Updatable) OID() asn1.ObjectIdentifier {
	u.mux.Lock()
	defer u.mux.Unlock()
	return u.Validator.OID()
}

// Update switches out the underlying validator.
func (u *Updatable) Update() error {
	u.mux.Lock()
	defer u.mux.Unlock()

	u.log.Infof("Updating expected measurements")

	data, err := u.fileHandler.Read(filepath.Join(constants.ServiceBasePath, constants.AttestationConfigFilename))
	if err != nil {
		return err
	}
	cfg, err := config.UnmarshalAttestationConfig(data, u.variant)
	if err != nil {
		return fmt.Errorf("unmarshaling config: %w", err)
	}
	u.log.Debugf("New expected measurements: %+v", cfg.GetMeasurements())

	cfgWithCerts, err := u.configWithCerts(cfg)
	if err != nil {
		return fmt.Errorf("adding cached certificates: %w", err)
	}

	validator, err := choose.Validator(cfgWithCerts, u.log)
	if err != nil {
		return fmt.Errorf("choosing validator: %w", err)
	}
	u.Validator = validator

	return nil
}

// addCachedCerts adds the certificates cached by the validator to the attestation config, if applicable.
func (u *Updatable) configWithCerts(cfg config.AttestationCfg) (config.AttestationCfg, error) {
	switch c := cfg.(type) {
	case *config.AzureSEVSNP:
		ask, err := u.getCachedAskCert()
		if err != nil {
			return nil, fmt.Errorf("getting cached ASK certificate: %w", err)
		}
		c.AMDSigningKey = config.Certificate(ask)
		return c, nil
	}
	// TODO(derpsteb): Add AWS SEV-SNP

	return cfg, nil
}

// getCachedAskCert returns the cached SEV-SNP ASK certificate.
func (u *Updatable) getCachedAskCert() (x509.Certificate, error) {
	if u.cachedCerts == nil {
		return x509.Certificate{}, fmt.Errorf("no cached certs available")
	}
	ask, ark := u.cachedCerts.SevSnpCerts()
	if ask == nil {
		return x509.Certificate{}, fmt.Errorf("no ASK available")
	}
	if ark == nil {
		return x509.Certificate{}, fmt.Errorf("no ARK available")
	}
	return *ask, nil
}
