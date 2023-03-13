/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package watcher

import (
	"encoding/asn1"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/edgelesssys/constellation/v2/internal/atls"
	"github.com/edgelesssys/constellation/v2/internal/attestation/choose"
	"github.com/edgelesssys/constellation/v2/internal/attestation/idkeydigest"
	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/oid"
)

// Updatable implements an updatable atls.Validator.
type Updatable struct {
	log         *logger.Logger
	mux         sync.Mutex
	fileHandler file.Handler
	variant     oid.Getter
	atls.Validator
}

// NewValidator initializes a new updatable validator.
func NewValidator(log *logger.Logger, variant oid.Getter, fileHandler file.Handler) (*Updatable, error) {
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
func (u *Updatable) Validate(attDoc []byte, nonce []byte) ([]byte, error) {
	u.mux.Lock()
	defer u.mux.Unlock()
	return u.Validator.Validate(attDoc, nonce)
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

	var measurements measurements.M
	if err := u.fileHandler.ReadJSON(filepath.Join(constants.ServiceBasePath, constants.MeasurementsFilename), &measurements); err != nil {
		return err
	}
	u.log.Debugf("New measurements: %+v", measurements)

	var digest idkeydigest.IDKeyDigests
	var enforceIDKeyDigest bool
	if u.variant.OID().Equal(oid.AzureSEVSNP{}.OID()) {
		u.log.Infof("Updating encforceIdKeyDigest value")
		enforceRaw, err := u.fileHandler.Read(filepath.Join(constants.ServiceBasePath, constants.EnforceIDKeyDigestFilename))
		if err != nil {
			return err
		}
		enforceIDKeyDigest, err = strconv.ParseBool(string(enforceRaw))
		if err != nil {
			return fmt.Errorf("parsing content of EnforceIdKeyDigestFilename: %s: %w", enforceRaw, err)
		}
		u.log.Debugf("New encforceIdKeyDigest value: %v", enforceIDKeyDigest)

		u.log.Infof("Updating expected idkeydigest")
		idkeydigestRaw, err := u.fileHandler.Read(filepath.Join(constants.ServiceBasePath, constants.IDKeyDigestFilename))
		if err != nil {
			return err
		}
		if err = json.Unmarshal(idkeydigestRaw, &digest); err != nil {
			return fmt.Errorf("unmarshaling content of IDKeyDigestFilename: %s: %w", idkeydigestRaw, err)
		}
		u.log.Debugf("New idkeydigest: %v", digest)
	}

	validator, err := choose.Validator(u.variant, measurements, digest, enforceIDKeyDigest, u.log)
	if err != nil {
		return fmt.Errorf("updating validator: %w", err)
	}
	u.Validator = validator

	return nil
}
