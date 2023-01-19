/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
Package setup handles setting up rejoinclient and recoveryserver for the disk-mapper.

On success of either of these services, the state disk is decrypted and the node is tainted as initialized by updating it's PCRs.
*/
package setup

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"

	"github.com/edgelesssys/constellation/v2/disk-mapper/internal/systemd"
	"github.com/edgelesssys/constellation/v2/internal/attestation"
	"github.com/edgelesssys/constellation/v2/internal/attestation/vtpm"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/crypto"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/nodestate"
	"github.com/spf13/afero"
)

const (
	keyPath             = "/run/cryptsetup-keys.d"
	keyFile             = "state.key"
	stateDiskMappedName = "state"
	stateDiskMountPath  = "/var/run/state"
	cryptsetupOptions   = "cipher=aes-xts-plain64,integrity=hmac-sha256"
	stateInfoPath       = stateDiskMountPath + "/constellation/node_state.json"
)

// Manager handles formatting, mapping, mounting and unmounting of state disks.
type Manager struct {
	log      *logger.Logger
	csp      string
	diskPath string
	fs       afero.Afero
	mapper   DeviceMapper
	mounter  Mounter
	config   ConfigurationGenerator
	openTPM  vtpm.TPMOpenFunc
}

// New initializes a SetupManager with the given parameters.
func New(log *logger.Logger, csp string, diskPath string, fs afero.Afero,
	mapper DeviceMapper, mounter Mounter, openTPM vtpm.TPMOpenFunc,
) *Manager {
	return &Manager{
		log:      log,
		csp:      csp,
		diskPath: diskPath,
		fs:       fs,
		mapper:   mapper,
		mounter:  mounter,
		config:   systemd.New(fs),
		openTPM:  openTPM,
	}
}

// PrepareExistingDisk requests and waits for a decryption key to remap the encrypted state disk.
// Once the disk is mapped, the function taints the node as initialized by updating it's PCRs.
func (s *Manager) PrepareExistingDisk(recover RecoveryDoer) error {
	s.log.Infof("Preparing existing state disk")
	uuid := s.mapper.DiskUUID()

	endpoint := net.JoinHostPort("0.0.0.0", strconv.Itoa(constants.RecoveryPort))

	passphrase, measurementSecret, err := recover.Do(uuid, endpoint)
	if err != nil {
		return fmt.Errorf("failed to perform recovery: %w", err)
	}

	if err := s.mapper.MapDisk(stateDiskMappedName, string(passphrase)); err != nil {
		return err
	}

	if err := s.mounter.MkdirAll(stateDiskMountPath, os.ModePerm); err != nil {
		return err
	}
	// we do not care about cleaning up the mount point on error, since any errors returned here should cause a boot failure
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
func (s *Manager) PrepareNewDisk() error {
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

func (s *Manager) readMeasurementSalt(path string) ([]byte, error) {
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
func (s *Manager) saveConfiguration(passphrase []byte) error {
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

// RecoveryServer interface serves a recovery server.
type RecoveryServer interface {
	Serve(context.Context, net.Listener, string) (key, secret []byte, err error)
}

// RejoinClient interface starts a rejoin client.
type RejoinClient interface {
	Start(context.Context, string) (key, secret []byte)
}

// NodeRecoverer bundles a RecoveryServer and RejoinClient.
type NodeRecoverer struct {
	recoveryServer RecoveryServer
	rejoinClient   RejoinClient
}

// NewNodeRecoverer initializes a new nodeRecoverer.
func NewNodeRecoverer(recoveryServer RecoveryServer, rejoinClient RejoinClient) *NodeRecoverer {
	return &NodeRecoverer{
		recoveryServer: recoveryServer,
		rejoinClient:   rejoinClient,
	}
}

// Do performs a recovery procedure on the given state disk.
// The method starts a gRPC server to allow manual recovery by a user.
// At the same time it tries to request a decryption key from all available Constellation control-plane nodes.
func (r *NodeRecoverer) Do(uuid, endpoint string) (passphrase, measurementSecret []byte, err error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	lis, err := net.Listen("tcp", endpoint)
	if err != nil {
		return nil, nil, err
	}
	defer lis.Close()

	var once sync.Once
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		key, secret, serveErr := r.recoveryServer.Serve(ctx, lis, uuid)
		once.Do(func() {
			cancel()
			passphrase = key
			measurementSecret = secret
		})
		if serveErr != nil && !errors.Is(serveErr, context.Canceled) {
			err = serveErr
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		key, secret := r.rejoinClient.Start(ctx, uuid)
		once.Do(func() {
			cancel()
			passphrase = key
			measurementSecret = secret
		})
	}()

	wg.Wait()
	return passphrase, measurementSecret, err
}
