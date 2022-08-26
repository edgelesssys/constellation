package setup

import (
	"crypto/rand"
	"errors"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/edgelesssys/constellation/internal/attestation"
	"github.com/edgelesssys/constellation/internal/attestation/vtpm"
	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/edgelesssys/constellation/internal/crypto"
	"github.com/edgelesssys/constellation/internal/file"
	"github.com/edgelesssys/constellation/internal/logger"
	"github.com/edgelesssys/constellation/internal/nodestate"
	"github.com/edgelesssys/constellation/state/internal/systemd"
	"github.com/spf13/afero"
	"go.uber.org/zap"
)

const (
	keyPath             = "/run/cryptsetup-keys.d"
	keyFile             = "state.key"
	stateDiskMappedName = "state"
	stateDiskMountPath  = "/var/run/state"
	cryptsetupOptions   = "cipher=aes-xts-plain64,integrity=hmac-sha256"
	stateInfoPath       = stateDiskMountPath + "/constellation/node_state.json"
)

// SetupManager handles formatting, mapping, mounting and unmounting of state disks.
type SetupManager struct {
	log       *logger.Logger
	csp       string
	diskPath  string
	fs        afero.Afero
	keyWaiter KeyWaiter
	mapper    DeviceMapper
	mounter   Mounter
	config    ConfigurationGenerator
	openTPM   vtpm.TPMOpenFunc
}

// New initializes a SetupManager with the given parameters.
func New(log *logger.Logger, csp string, diskPath string, fs afero.Afero, keyWaiter KeyWaiter, mapper DeviceMapper, mounter Mounter, openTPM vtpm.TPMOpenFunc) *SetupManager {
	return &SetupManager{
		log:       log,
		csp:       csp,
		diskPath:  diskPath,
		fs:        fs,
		keyWaiter: keyWaiter,
		mapper:    mapper,
		mounter:   mounter,
		config:    systemd.New(fs),
		openTPM:   openTPM,
	}
}

// PrepareExistingDisk requests and waits for a decryption key to remap the encrypted state disk.
// Once the disk is mapped, the function taints the node as initialized by updating it's PCRs.
func (s *SetupManager) PrepareExistingDisk() error {
	s.log.Infof("Preparing existing state disk")
	uuid := s.mapper.DiskUUID()

	endpoint := net.JoinHostPort("0.0.0.0", strconv.Itoa(constants.RecoveryPort))
getKey:
	passphrase, measurementSecret, err := s.keyWaiter.WaitForDecryptionKey(uuid, endpoint)
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
	clusterID, err := attestation.DeriveClusterID(measurementSecret, measurementSalt)
	if err != nil {
		return err
	}

	// taint the node as initialized
	if err := vtpm.MarkNodeAsBootstrapped(s.openTPM, clusterID); err != nil {
		return err
	}

	if err := s.saveConfiguration(passphrase); err != nil {
		return err
	}

	return s.mounter.Unmount(stateDiskMountPath, 0)
}

// PrepareNewDisk prepares an instances state disk by formatting the disk as a LUKS device using a random passphrase.
func (s *SetupManager) PrepareNewDisk() error {
	s.log.Infof("Preparing new state disk")

	// generate and save temporary passphrase
	passphrase := make([]byte, crypto.RNGLengthDefault)
	if _, err := rand.Read(passphrase); err != nil {
		return err
	}
	if err := s.saveConfiguration(passphrase); err != nil {
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

// saveConfiguration saves the given passphrase and cryptsetup mapping configuration to disk.
func (s *SetupManager) saveConfiguration(passphrase []byte) error {
	// passphrase
	if err := s.fs.MkdirAll(keyPath, os.ModePerm); err != nil {
		return err
	}
	if err := s.fs.WriteFile(filepath.Join(keyPath, keyFile), passphrase, 0o400); err != nil {
		return err
	}

	// systemd cryptsetup unit
	return s.config.Generate(stateDiskMappedName, s.diskPath, filepath.Join(keyPath, keyFile), cryptsetupOptions)
}
