package watcher

import (
	"encoding/asn1"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/edgelesssys/constellation/internal/atls"
	"github.com/edgelesssys/constellation/internal/attestation/azure"
	"github.com/edgelesssys/constellation/internal/attestation/gcp"
	"github.com/edgelesssys/constellation/internal/attestation/qemu"
	"github.com/edgelesssys/constellation/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/edgelesssys/constellation/internal/file"
	"github.com/edgelesssys/constellation/internal/logger"
)

// Updatable implements an updatable atls.Validator.
type Updatable struct {
	log          *logger.Logger
	mux          sync.Mutex
	newValidator newValidatorFunc
	fileHandler  file.Handler
	csp          cloudprovider.Provider
	atls.Validator
}

// NewValidator initializes a new updatable validator.
func NewValidator(log *logger.Logger, csp string, fileHandler file.Handler) (*Updatable, error) {
	var newValidator newValidatorFunc
	switch cloudprovider.FromString(csp) {
	case cloudprovider.Azure:
		newValidator = func(m map[uint32][]byte, e []uint32, idkeydigest []byte, enforceIdKeyDigest bool, log *logger.Logger) atls.Validator {
			return azure.NewValidator(m, e, idkeydigest, enforceIdKeyDigest, log)
		}
	case cloudprovider.GCP:
		newValidator = func(m map[uint32][]byte, e []uint32, _ []byte, _ bool, log *logger.Logger) atls.Validator {
			return gcp.NewValidator(m, e, log)
		}
	case cloudprovider.QEMU:
		newValidator = func(m map[uint32][]byte, e []uint32, _ []byte, _ bool, log *logger.Logger) atls.Validator {
			return qemu.NewValidator(m, e, log)
		}
	default:
		return nil, fmt.Errorf("unknown cloud service provider: %q", csp)
	}

	u := &Updatable{
		log:          log,
		newValidator: newValidator,
		fileHandler:  fileHandler,
		csp:          cloudprovider.FromString(csp),
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
	return u.Validator.OID()
}

// Update switches out the underlying validator.
func (u *Updatable) Update() error {
	u.mux.Lock()
	defer u.mux.Unlock()

	u.log.Infof("Updating expected measurements")

	var measurements map[uint32][]byte
	if err := u.fileHandler.ReadJSON(filepath.Join(constants.ServiceBasePath, constants.MeasurementsFilename), &measurements); err != nil {
		return err
	}
	u.log.Debugf("New measurements: %v", measurements)

	var enforced []uint32
	if err := u.fileHandler.ReadJSON(filepath.Join(constants.ServiceBasePath, constants.EnforcedPCRsFilename), &enforced); err != nil {
		return err
	}
	u.log.Debugf("Enforced PCRs: %v", enforced)

	var idkeydigest []byte
	var enforceIdKeyDigest bool
	if u.csp == cloudprovider.Azure {
		u.log.Infof("Updating encforceIdKeyDigest value")
		enforceRaw, err := u.fileHandler.Read(filepath.Join(constants.ServiceBasePath, constants.EnforceIdKeyDigestFilename))
		if err != nil {
			return err
		}
		enforceIdKeyDigest, err = strconv.ParseBool(string(enforceRaw))
		if err != nil {
			return fmt.Errorf("parsing content of EnforceIdKeyDigestFilename: %s: %w", enforceRaw, err)
		}
		u.log.Debugf("New encforceIdKeyDigest value: %v", enforceIdKeyDigest)

		u.log.Infof("Updating expected idkeydigest")
		idkeydigestRaw, err := u.fileHandler.Read(filepath.Join(constants.ServiceBasePath, constants.IdKeyDigestFilename))
		if err != nil {
			return err
		}
		idkeydigest, err = hex.DecodeString(string(idkeydigestRaw))
		if err != nil {
			return fmt.Errorf("parsing hexstring: %s: %w", idkeydigestRaw, err)
		}
		u.log.Debugf("New idkeydigest: %x", idkeydigest)
	}

	u.Validator = u.newValidator(measurements, enforced, idkeydigest, enforceIdKeyDigest, u.log)

	return nil
}

type newValidatorFunc func(measurements map[uint32][]byte, enforcedPCRs []uint32, idkeydigest []byte, enforceIdKeyDigest bool, log *logger.Logger) atls.Validator
