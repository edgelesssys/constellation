package watcher

import (
	"encoding/asn1"
	"fmt"
	"path/filepath"
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
	atls.Validator
}

// NewValidator initializes a new updatable validator.
func NewValidator(log *logger.Logger, csp string, fileHandler file.Handler) (*Updatable, error) {
	var newValidator newValidatorFunc
	switch cloudprovider.FromString(csp) {
	case cloudprovider.Azure:
		newValidator = func(m map[uint32][]byte) atls.Validator { return azure.NewValidator(m) }
	case cloudprovider.GCP:
		newValidator = func(m map[uint32][]byte) atls.Validator { return gcp.NewValidator(m) }
	case cloudprovider.QEMU:
		newValidator = func(m map[uint32][]byte) atls.Validator { return qemu.NewValidator(m) }
	default:
		return nil, fmt.Errorf("unknown cloud service provider: %q", csp)
	}

	u := &Updatable{
		log:          log,
		newValidator: newValidator,
		fileHandler:  fileHandler,
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

	u.Validator = u.newValidator(measurements)

	return nil
}

type newValidatorFunc func(measurements map[uint32][]byte) atls.Validator
