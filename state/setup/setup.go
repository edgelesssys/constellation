package setup

import (
	"crypto/rand"
	"errors"
	"net"
	"os"
	"path/filepath"
	"syscall"

	"github.com/edgelesssys/constellation/bootstrapper/nodestate"
	"github.com/edgelesssys/constellation/internal/attestation"
	"github.com/edgelesssys/constellation/internal/attestation/vtpm"
	"github.com/edgelesssys/constellation/internal/crypto"
	"github.com/edgelesssys/constellation/internal/file"
	"github.com/edgelesssys/constellation/internal/logger"
	"github.com/spf13/afero"
	"go.uber.org/zap"
)

const (
	RecoveryPort        = "9000"
	keyPath             = "/run/cryptsetup-keys.d"
	keyFile             = "state.key"
	stateDiskMappedName = "state"
	stateDiskMountPath  = "/var/run/state"
	stateInfoPath       = stateDiskMountPath + "/constellation/node_state.json"
)

// SetupManager handles formating, mapping, mounting and unmounting of state disks.
type SetupManager struct {
	log       *logger.Logger
	csp       string
	fs        afero.Afero
	keyWaiter KeyWaiter
	mapper    DeviceMapper
	mounter   Mounter
	openTPM   vtpm.TPMOpenFunc
}

// New initializes a SetupManager with the given parameters.
func New(log *logger.Logger, csp string, fs afero.Afero, keyWaiter KeyWaiter, mapper DeviceMapper, mounter Mounter, openTPM vtpm.TPMOpenFunc) *SetupManager {
	return &SetupManager{
		log:       log,
		csp:       csp,
		fs:        fs,
		keyWaiter: keyWaiter,
		mapper:    mapper,
		mounter:   mounter,
		openTPM:   openTPM,
	}
}

// PrepareExistingDisk requests and waits for a decryption key to remap the encrypted state disk.
// Once the disk is mapped, the function taints the node as initialized by updating it's PCRs.
func (s *SetupManager) PrepareExistingDisk() error {
	s.log.Infof("Preparing existing state disk")
	uuid := s.mapper.DiskUUID()

getKey:
	passphrase, measurementSecret, err := s.keyWaiter.WaitForDecryptionKey(uuid, net.JoinHostPort("0.0.0.0", RecoveryPort))
	if err != nil {
		return err
	}

	if err := s.mapper.MapDisk(stateDiskMappedName, string(passphrase)); err != nil {
		// retry key fetching if disk mapping fails
		s.log.With(zap.Error(err)).Errorf("Failed to map state disk, retrying...")
		s.keyWaiter.ResetKey()
		goto getKey
	}

	if err := s.mounter.MkdirAll(stateDiskMountPath, os.ModePerm); err != nil {
		return err
	}
	// we do not care about cleaning up the mount point on error, since any errors returned here should result in a kernel panic in the main function
	if err := s.mounter.Mount(filepath.Join("/dev/mapper/", stateDiskMappedName), stateDiskMountPath, "ext4", syscall.MS_RDONLY, ""); err != nil {
		return err
	}

	measurementSalt, err := s.readMeasurementSalt(stateInfoPath)
	if err != nil {
		return err
	}
	clusterID, err := attestation.DeriveClusterID(measurementSalt, measurementSecret)
	if err != nil {
		return err
	}

	// taint the node as initialized
	if err := vtpm.MarkNodeAsBootstrapped(s.openTPM, clusterID); err != nil {
		return err
	}

	return s.mounter.Unmount(stateDiskMountPath, 0)
}

// PrepareNewDisk prepares an instances state disk by formatting the disk as a LUKS device using a random passphrase.
func (s *SetupManager) PrepareNewDisk() error {
	s.log.Infof("Preparing new state disk")

	// generate and save temporary passphrase
	if err := s.fs.MkdirAll(keyPath, os.ModePerm); err != nil {
		return err
	}

	passphrase := make([]byte, crypto.RNGLengthDefault)
	if _, err := rand.Read(passphrase); err != nil {
		return err
	}
	if err := s.fs.WriteFile(filepath.Join(keyPath, keyFile), passphrase, 0o400); err != nil {
		return err
	}

	if err := s.mapper.FormatDisk(string(passphrase)); err != nil {
		return err
	}

	return s.mapper.MapDisk(stateDiskMappedName, string(passphrase))
}

func (s *SetupManager) readMeasurementSalt(path string) ([]byte, error) {
	handler := file.NewHandler(s.fs)
	var state nodestate.NodeState
	if err := handler.ReadJSON(path, &state); err != nil {
		return nil, err
	}

	if len(state.MeasurementSalt) != crypto.RNGLengthDefault {
		return nil, errors.New("missing state information to retaint node")
	}

	return state.MeasurementSalt, nil
}
