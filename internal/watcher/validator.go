/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package watcher

import (
	"encoding/asn1"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/edgelesssys/constellation/v2/internal/atls"
	"github.com/edgelesssys/constellation/v2/internal/attestation/aws"
	"github.com/edgelesssys/constellation/v2/internal/attestation/azure/snp"
	"github.com/edgelesssys/constellation/v2/internal/attestation/azure/trustedlaunch"
	"github.com/edgelesssys/constellation/v2/internal/attestation/gcp"
	"github.com/edgelesssys/constellation/v2/internal/attestation/idkeydigest"
	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/attestation/qemu"
	"github.com/edgelesssys/constellation/v2/internal/attestation/tdx"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/logger"
)

// Updatable implements an updatable atls.Validator.
type Updatable struct {
	log          *logger.Logger
	mux          sync.Mutex
	newValidator newValidatorFunc
	fileHandler  file.Handler
	csp          cloudprovider.Provider
	azureCVM     bool
	atls.Validator
}

// NewValidator initializes a new updatable validator.
func NewValidator(log *logger.Logger, csp string, fileHandler file.Handler, azureCVM bool) (*Updatable, error) {
	var newValidator newValidatorFunc
	switch cloudprovider.FromString(csp) {
	case cloudprovider.AWS:
		newValidator = func(m measurements.M, _ idkeydigest.IDKeyDigests, _ bool, log *logger.Logger) atls.Validator {
			return aws.NewValidator(m, log)
		}
	case cloudprovider.Azure:
		if azureCVM {
			newValidator = func(m measurements.M, idkeydigest idkeydigest.IDKeyDigests, enforceIdKeyDigest bool, log *logger.Logger) atls.Validator {
				return snp.NewValidator(m, idkeydigest, enforceIdKeyDigest, log)
			}
		} else {
			newValidator = func(m measurements.M, idkeydigest idkeydigest.IDKeyDigests, enforceIdKeyDigest bool, log *logger.Logger) atls.Validator {
				return trustedlaunch.NewValidator(m, log)
			}
		}
	case cloudprovider.GCP:
		newValidator = func(m measurements.M, _ idkeydigest.IDKeyDigests, _ bool, log *logger.Logger) atls.Validator {
			return gcp.NewValidator(m, log)
		}
	case cloudprovider.QEMU:
		if tdx.Available() {
			newValidator = func(m measurements.M, _ idkeydigest.IDKeyDigests, _ bool, log *logger.Logger) atls.Validator {
				return tdx.NewValidator(m, log)
			}
		} else {
			newValidator = func(m measurements.M, _ idkeydigest.IDKeyDigests, _ bool, log *logger.Logger) atls.Validator {
				return qemu.NewValidator(m, log)
			}
		}
	default:
		return nil, fmt.Errorf("unknown cloud service provider: %q", csp)
	}

	u := &Updatable{
		log:          log,
		newValidator: newValidator,
		fileHandler:  fileHandler,
		csp:          cloudprovider.FromString(csp),
		azureCVM:     azureCVM,
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

	// handle legacy measurement format, where expected measurements and enforced measurements were stored in separate data structures
	// TODO: remove with v2.4.0
	var enforced []uint32
	if err := u.fileHandler.ReadJSON(filepath.Join(constants.ServiceBasePath, constants.EnforcedPCRsFilename), &enforced); err == nil {
		u.log.Debugf("Detected legacy format. Loading enforced PCRs...")
		if err := measurements.SetEnforced(enforced); err != nil {
			return err
		}
		u.log.Debugf("Merged measurements with enforced values: %+v", measurements)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	var digest idkeydigest.IDKeyDigests
	var enforceIDKeyDigest bool
	if u.csp == cloudprovider.Azure && u.azureCVM {
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

	u.Validator = u.newValidator(measurements, digest, enforceIDKeyDigest, u.log)

	return nil
}

type newValidatorFunc func(measurements measurements.M, idkeydigest idkeydigest.IDKeyDigests, enforceIdKeyDigest bool, log *logger.Logger) atls.Validator
