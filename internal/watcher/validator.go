/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package watcher

import (
	"context"
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
)

// Updatable implements an updatable atls.Validator.
type Updatable struct {
	log         *logger.Logger
	mux         sync.Mutex
	fileHandler file.Handler
	variant     variant.Variant
	atls.Validator
}

// NewValidator initializes a new updatable validator.
func NewValidator(log *logger.Logger, variant variant.Variant, fileHandler file.Handler) (*Updatable, error) {
	u := &Updatable{
		log:         log,
		fileHandler: fileHandler,
		variant:     variant,
	}

	if err := u.Update(); err != nil {
		return nil, err
	}
	return u, nil
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

	validator, err := choose.Validator(cfg, u.log)
	if err != nil {
		return fmt.Errorf("updating validator: %w", err)
	}
	u.Validator = validator

	return nil
}
