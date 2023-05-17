//go:build linux && cgo

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package mapper

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	ccryptsetup "github.com/edgelesssys/constellation/v2/internal/cryptsetup"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	cryptsetup "github.com/martinjungblut/go-cryptsetup"
	"go.uber.org/zap"
)

// packageLock is needed to block concurrent use of package functions, since libcryptsetup is not thread safe.
// See: https://gitlab.com/cryptsetup/cryptsetup/-/issues/710
//
//	https://stackoverflow.com/questions/30553386/cryptsetup-backend-safe-with-multithreading
var packageLock = sync.Mutex{}

// Mapper handles actions for formatting and mapping crypt devices.
type Mapper struct {
	device cryptDevice
	log    *logger.Logger
}

// New creates a new crypt device for the device at path.
func New(path string, log *logger.Logger) (*Mapper, error) {
	packageLock.Lock()
	device, err := cryptsetup.Init(path)
	if err != nil {
		return nil, fmt.Errorf("initializing crypt device for disk %q: %w", path, err)
	}
	return &Mapper{device: device, log: log}, nil
}

// Close closes and frees memory allocated for the crypt device.
func (m *Mapper) Close() error {
	defer packageLock.Unlock()
	if m.device.Free() {
		return nil
	}
	return errors.New("unable to close crypt device")
}

// IsLUKSDevice returns true if the device is formatted as a LUKS device.
func (m *Mapper) IsLUKSDevice() bool {
	return m.device.Load(cryptsetup.LUKS2{}) == nil
}

// DiskUUID gets the device's UUID.
func (m *Mapper) DiskUUID() string {
	return strings.ToLower(m.device.GetUUID())
}

// FormatDisk formats the disk and adds passphrase in keyslot 0.
func (m *Mapper) FormatDisk(passphrase string) error {
	luksParams := cryptsetup.LUKS2{
		SectorSize: 4096,
		Integrity:  "hmac(sha256)",
		PBKDFType: &cryptsetup.PbkdfType{
			// Use low memory recommendation from https://datatracker.ietf.org/doc/html/rfc9106#section-7
			Type:            "argon2id",
			TimeMs:          2000,
			Iterations:      3,
			ParallelThreads: 4,
			MaxMemoryKb:     65536, // ~64MiB
		},
	}

	genericParams := cryptsetup.GenericParams{
		Cipher:        "aes",
		CipherMode:    "xts-plain64",
		VolumeKeySize: 96, // 32*2 bytes for aes-xts-plain64 encryption, 32 bytes for hmac(sha256) integrity
	}

	if err := m.device.Format(luksParams, genericParams); err != nil {
		return fmt.Errorf("formatting disk: %w", err)
	}

	if err := m.device.KeyslotAddByVolumeKey(0, "", passphrase); err != nil {
		return fmt.Errorf("adding keyslot: %w", err)
	}

	// wipe using 64MiB block size
	if err := m.Wipe(67108864); err != nil {
		return fmt.Errorf("wiping disk: %w", err)
	}

	return nil
}

// MapDisk maps a crypt device to /dev/mapper/target using the provided passphrase.
func (m *Mapper) MapDisk(target, passphrase string) error {
	if err := m.device.ActivateByPassphrase(target, 0, passphrase, ccryptsetup.ReadWriteQueueBypass); err != nil {
		return fmt.Errorf("mapping disk as %q: %w", target, err)
	}
	return nil
}

// UnmapDisk removes the mapping of target.
func (m *Mapper) UnmapDisk(target string) error {
	return m.device.Deactivate(target)
}

// Wipe overwrites the device with zeros to initialize integrity checksums.
func (m *Mapper) Wipe(blockWipeSize int) error {
	// Activate as temporary device using the internal volume key
	tmpDevice := "tmp-cryptsetup-integrity"
	if err := m.device.ActivateByVolumeKey(tmpDevice, "", 0, (cryptsetup.CRYPT_ACTIVATE_PRIVATE | cryptsetup.CRYPT_ACTIVATE_NO_JOURNAL)); err != nil {
		return fmt.Errorf("activating as temporary device: %w", err)
	}

	// set progress logging callback once every 30 seconds
	ticker := time.NewTicker(30 * time.Second)
	firstReq := make(chan struct{}, 1)
	firstReq <- struct{}{}
	defer ticker.Stop()
	logProgress := func(size, offset uint64) {
		prog := (float64(offset) / float64(size)) * 100
		m.log.With(zap.String("progress", fmt.Sprintf("%.2f%%", prog))).Infof("Wiping disk")
	}

	progressCallback := func(size, offset uint64) int {
		select {
		case <-firstReq:
			logProgress(size, offset)
		case <-ticker.C:
			logProgress(size, offset)
		default:
		}

		return 0
	}

	start := time.Now()
	// wipe the device
	if err := m.device.Wipe("/dev/mapper/"+tmpDevice, cryptsetup.CRYPT_WIPE_ZERO, 0, 0, blockWipeSize, 0, progressCallback); err != nil {
		return fmt.Errorf("wiping disk: %w", err)
	}
	m.log.With(zap.Duration("duration", time.Since(start))).Infof("Wiping disk successful")

	// Deactivate the temporary device
	if err := m.device.Deactivate(tmpDevice); err != nil {
		return fmt.Errorf("deactivating temporary device: %w", err)
	}

	return nil
}
