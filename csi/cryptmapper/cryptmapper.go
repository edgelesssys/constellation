/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package cryptmapper provides a wrapper around libcryptsetup to manage dm-crypt volumes for CSI drivers.
package cryptmapper

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/crypto"
	"github.com/edgelesssys/constellation/v2/internal/cryptsetup"
)

const (
	// LUKSHeaderSize is the amount of bytes taken up by the header of a LUKS2 partition.
	// The header is 16MiB (1048576 Bytes * 16).
	LUKSHeaderSize    = 16777216
	cryptPrefix       = "/dev/mapper/"
	integritySuffix   = "_dif"
	integrityFSSuffix = "-integrity"
	keySizeIntegrity  = 96
	keySizeCrypt      = 64
)

// CryptMapper manages dm-crypt volumes.
type CryptMapper struct {
	mapper        deviceMapper
	kms           keyCreator
	getDiskFormat func(disk string) (string, error)
}

// New initializes a new CryptMapper with the given kms client and key-encryption-key ID.
// kms is used to fetch data encryption keys for the dm-crypt volumes.
func New(kms keyCreator, mapper deviceMapper) *CryptMapper {
	return &CryptMapper{
		mapper:        mapper,
		kms:           kms,
		getDiskFormat: getDiskFormat,
	}
}

// CloseCryptDevice closes the crypt device mapped for volumeID.
// Returns nil if the volume does not exist.
func (c *CryptMapper) CloseCryptDevice(volumeID string) error {
	source, err := filepath.EvalSymlinks(cryptPrefix + volumeID)
	if err != nil {
		var pathErr *fs.PathError
		if errors.As(err, &pathErr) {
			return nil
		}
		return fmt.Errorf("getting device path for disk %q: %w", cryptPrefix+volumeID, err)
	}
	if err := c.closeCryptDevice(source, volumeID, "crypt"); err != nil {
		return fmt.Errorf("closing crypt device: %w", err)
	}

	integrity, err := filepath.EvalSymlinks(cryptPrefix + volumeID + integritySuffix)
	if err == nil {
		// If device was created with integrity, we need to also close the integrity device
		integrityErr := c.closeCryptDevice(integrity, volumeID+integritySuffix, "integrity")
		if integrityErr != nil {
			return integrityErr
		}
	}
	if err != nil {
		var pathErr *fs.PathError
		if errors.As(err, &pathErr) {
			// integrity device does not exist
			return nil
		}
		return fmt.Errorf("getting device path for disk %q: %w", cryptPrefix+volumeID, err)
	}

	return nil
}

// OpenCryptDevice maps the volume at source to the crypt device identified by volumeID.
// The key used to encrypt the volume is fetched using CryptMapper's kms client.
func (c *CryptMapper) OpenCryptDevice(ctx context.Context, source, volumeID string, integrity bool) (string, error) {
	// Initialize the block device
	free, err := c.mapper.Init(source)
	if err != nil {
		return "", fmt.Errorf("initializing dm-crypt to map device %q: %w", source, err)
	}
	defer free()

	var passphrase []byte
	// Try to load LUKS headers
	// If this fails, the device is either not formatted at all, or already formatted with a different FS
	if err := c.mapper.LoadLUKS2(); err != nil {
		passphrase, err = c.formatNewDevice(ctx, volumeID, source, integrity)
		if err != nil {
			return "", fmt.Errorf("formatting device: %w", err)
		}
	} else {
		uuid, err := c.mapper.GetUUID()
		if err != nil {
			return "", err
		}
		passphrase, err = c.kms.GetDEK(ctx, uuid, crypto.StateDiskKeyLength)
		if err != nil {
			return "", err
		}
		if len(passphrase) != crypto.StateDiskKeyLength {
			return "", fmt.Errorf("expected key length to be [%d] but got [%d]", crypto.StateDiskKeyLength, len(passphrase))
		}
	}

	if err := c.mapper.ActivateByPassphrase(volumeID, 0, string(passphrase), cryptsetup.ReadWriteQueueBypass); err != nil {
		return "", fmt.Errorf("trying to activate dm-crypt volume: %w", err)
	}

	return cryptPrefix + volumeID, nil
}

// ResizeCryptDevice resizes the underlying crypt device and returns the mapped device path.
func (c *CryptMapper) ResizeCryptDevice(ctx context.Context, volumeID string) (string, error) {
	free, err := c.mapper.InitByName(volumeID)
	if err != nil {
		return "", fmt.Errorf("initializing device: %w", err)
	}
	defer free()

	if err := c.mapper.LoadLUKS2(); err != nil {
		return "", fmt.Errorf("loading device: %w", err)
	}

	uuid, err := c.mapper.GetUUID()
	if err != nil {
		return "", err
	}
	passphrase, err := c.kms.GetDEK(ctx, uuid, crypto.StateDiskKeyLength)
	if err != nil {
		return "", fmt.Errorf("getting key: %w", err)
	}

	if err := c.mapper.ActivateByPassphrase("", 0, string(passphrase), resizeFlags); err != nil {
		return "", fmt.Errorf("activating keyring for crypt device %q with passphrase: %w", volumeID, err)
	}

	if err := c.mapper.Resize(volumeID, 0); err != nil {
		return "", fmt.Errorf("resizing device: %w", err)
	}

	return cryptPrefix + volumeID, nil
}

// GetDevicePath returns the device path of a mapped crypt device.
func (c *CryptMapper) GetDevicePath(volumeID string) (string, error) {
	name := strings.TrimPrefix(volumeID, cryptPrefix)
	free, err := c.mapper.InitByName(name)
	if err != nil {
		return "", fmt.Errorf("initializing device: %w", err)
	}
	defer free()

	deviceName := c.mapper.GetDeviceName()
	if deviceName == "" {
		return "", errors.New("unable to determine device name")
	}
	return deviceName, nil
}

// closeCryptDevice closes the crypt device mapped for volumeID.
func (c *CryptMapper) closeCryptDevice(source, volumeID, deviceType string) error {
	free, err := c.mapper.InitByName(volumeID)
	if err != nil {
		return fmt.Errorf("initializing dm-%s to unmap device %q: %w", deviceType, source, err)
	}
	defer free()

	if err := c.mapper.Deactivate(volumeID); err != nil {
		return fmt.Errorf("deactivating dm-%s volume %q for device %q: %w", deviceType, cryptPrefix+volumeID, source, err)
	}

	return nil
}

func (c *CryptMapper) formatNewDevice(ctx context.Context, volumeID, source string, integrity bool) ([]byte, error) {
	format, err := c.getDiskFormat(source)
	if err != nil {
		return nil, fmt.Errorf("determining if disk is formatted: %w", err)
	}
	if format != "" {
		return nil, fmt.Errorf("disk %q is already formatted as: %s", source, format)
	}

	// Device is not formatted, so we can safely create a new LUKS2 partition
	if err := c.mapper.Format(integrity); err != nil {
		return nil, fmt.Errorf("formatting device %q: %w", source, err)
	}

	uuid, err := c.mapper.GetUUID()
	if err != nil {
		return nil, err
	}
	passphrase, err := c.kms.GetDEK(ctx, uuid, crypto.StateDiskKeyLength)
	if err != nil {
		return nil, err
	}
	if len(passphrase) != crypto.StateDiskKeyLength {
		return nil, fmt.Errorf("expected key length to be [%d] but got [%d]", crypto.StateDiskKeyLength, len(passphrase))
	}

	// Add a new keyslot using the internal volume key
	if err := c.mapper.KeyslotAddByVolumeKey(0, "", string(passphrase)); err != nil {
		return nil, fmt.Errorf("adding keyslot: %w", err)
	}

	if integrity {
		logProgress := func(size, offset uint64) {
			prog := (float64(offset) / float64(size)) * 100
			fmt.Printf("Wipe in progress: %.2f%%\n", prog)
		}

		if err := c.mapper.Wipe(volumeID, 1024*1024, 0, logProgress, 30*time.Second); err != nil {
			return nil, fmt.Errorf("wiping device: %w", err)
		}
	}

	return passphrase, nil
}

// IsIntegrityFS checks if the fstype string contains an integrity suffix.
// If yes, returns the trimmed fstype and true, fstype and false otherwise.
func IsIntegrityFS(fstype string) (string, bool) {
	if strings.HasSuffix(fstype, integrityFSSuffix) {
		return strings.TrimSuffix(fstype, integrityFSSuffix), true
	}
	return fstype, false
}

// deviceMapper is an interface for device mapper methods.
type deviceMapper interface {
	Init(devicePath string) (func(), error)
	InitByName(name string) (func(), error)
	ActivateByPassphrase(deviceName string, keyslot int, passphrase string, flags int) error
	ActivateByVolumeKey(deviceName string, volumeKey string, volumeKeySize int, flags int) error
	Deactivate(deviceName string) error
	Format(integrity bool) error
	Free()
	GetDeviceName() string
	GetUUID() (string, error)
	LoadLUKS2() error
	KeyslotAddByVolumeKey(keyslot int, volumeKey string, passphrase string) error
	Wipe(name string, wipeBlockSize int, flags int, progress func(size, offset uint64), frequency time.Duration) error
	Resize(name string, newSize uint64) error
}

// keyCreator is an interface to create data encryption keys.
type keyCreator interface {
	GetDEK(ctx context.Context, dekID string, dekSize int) ([]byte, error)
}
