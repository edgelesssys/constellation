/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
Package diskencryption uses libcryptsetup to format and map crypt devices.

This is used by the disk-mapper to set up a node's state disk.

All interaction with libcryptsetup should be done here.
*/
package diskencryption

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/cryptsetup"
	"github.com/edgelesssys/constellation/v2/internal/logger"
)

// DiskEncryption handles actions for formatting and mapping crypt devices.
type DiskEncryption struct {
	device     cryptDevice
	devicePath string
	log        *logger.Logger
}

// New creates a new crypt device for the device at path.
func New(path string, log *logger.Logger) (*DiskEncryption, func(), error) {
	device := cryptsetup.New()
	_, err := device.Init(path)
	if err != nil {
		return nil, nil, fmt.Errorf("initializing crypt device for disk %q: %w", path, err)
	}
	d := &DiskEncryption{device: device, devicePath: path, log: log}
	return d, d.free, nil
}

// IsInitialized returns true if the device is formatted as a LUKS device,
// and has been successfully initialized before (successfully joined a cluster).
func (d *DiskEncryption) IsInitialized() bool {
	if err := d.device.LoadLUKS2(); err != nil {
		return false
	}

	return d.device.ConstellationStateDiskTokenIsInitialized()
}

// DiskUUID gets the device's UUID.
func (d *DiskEncryption) DiskUUID() (string, error) {
	return d.device.GetUUID()
}

// FormatDisk formats the disk and adds passphrase in keyslot 0.
func (d *DiskEncryption) FormatDisk(passphrase string) error {
	// Successfully calling LoadLUKS2() before FormatDisk() will cause format to fail.
	// To make sure format is idempotent, we need to run it on a freshly initialized device.
	// Therefore we free the device and reinitialize it.
	d.free()
	if _, err := d.device.Init(d.devicePath); err != nil {
		return fmt.Errorf("re-initializing crypt device for disk %q: %w", d.devicePath, err)
	}

	if err := d.device.Format(cryptsetup.FormatIntegrity); err != nil {
		return fmt.Errorf("formatting disk: %w", err)
	}

	if err := d.device.KeyslotAddByVolumeKey(0, "", passphrase); err != nil {
		return fmt.Errorf("adding keyslot: %w", err)
	}

	// wipe using 64MiB block size
	if err := d.Wipe(67108864); err != nil {
		return fmt.Errorf("wiping disk: %w", err)
	}

	if err := d.device.SetConstellationStateDiskToken(cryptsetup.SetDiskNotInitialized); err != nil {
		return fmt.Errorf("setting disk token: %w", err)
	}
	return nil
}

// MapDisk maps a crypt device to /dev/mapper/target using the provided passphrase.
func (d *DiskEncryption) MapDisk(target, passphrase string) error {
	if err := d.device.ActivateByPassphrase(target, 0, passphrase, cryptsetup.ReadWriteQueueBypass); err != nil {
		return fmt.Errorf("mapping disk as %q: %w", target, err)
	}
	return nil
}

// UnmapDisk removes the mapping of target.
func (d *DiskEncryption) UnmapDisk(target string) error {
	return d.device.Deactivate(target)
}

// Wipe overwrites the device with zeros to initialize integrity checksums.
func (d *DiskEncryption) Wipe(blockWipeSize int) error {
	logProgress := func(size, offset uint64) {
		prog := (float64(offset) / float64(size)) * 100
		d.log.With(slog.String("progress", fmt.Sprintf("%.2f%%", prog))).Infof("Wiping disk")
	}

	start := time.Now()
	// wipe the device
	if err := d.device.Wipe("integrity", blockWipeSize, 0, logProgress, 30*time.Second); err != nil {
		return fmt.Errorf("wiping disk: %w", err)
	}
	d.log.With(slog.Duration("duration", time.Since(start))).Infof("Wiping disk successful")

	return nil
}

func (d *DiskEncryption) free() {
	d.device.Free()
}

type cryptDevice interface {
	ActivateByPassphrase(deviceName string, keyslot int, passphrase string, flags int) error
	ActivateByVolumeKey(deviceName string, volumeKey string, volumeKeySize int, flags int) error
	Deactivate(deviceName string) error
	Format(integrity bool) error
	Free()
	GetUUID() (string, error)
	Init(path string) (func(), error)
	LoadLUKS2() error
	KeyslotAddByVolumeKey(keyslot int, volumeKey string, passphrase string) error
	SetConstellationStateDiskToken(diskIsInitialized bool) error
	ConstellationStateDiskTokenIsInitialized() bool
	Wipe(name string, wipeBlockSize int, flags int, logCallback func(size, offset uint64), logFrequency time.Duration) error
}
