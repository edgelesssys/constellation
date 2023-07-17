/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
Package cryptsetup provides a wrapper around libcryptsetup.
The package is used to manage encrypted disks for Constellation.

Since libcryptsetup is not thread safe, this package uses a global lock to prevent concurrent use.
There should only be one instance using this package per process.
*/
package cryptsetup

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

const (
	// FormatIntegrity is a flag to enable dm-integrity for a crypt device when formatting.
	FormatIntegrity = true
	// FormatNoIntegrity is a flag to disable dm-integrity for a crypt device when formatting.
	FormatNoIntegrity = false
	tmpDevicePrefix   = "tmp-cryptsetup-"
	mappedDevicePath  = "/dev/mapper/"
)

// packageLock is needed to block concurrent use of package functions, since libcryptsetup is not thread safe.
// See: https://gitlab.com/cryptsetup/cryptsetup/-/issues/710
//
//	https://stackoverflow.com/questions/30553386/cryptsetup-backend-safe-with-multithreading
var (
	packageLock          = sync.Mutex{}
	errDeviceNotOpen     = errors.New("crypt device not open")
	errDeviceAlreadyOpen = errors.New("crypt device already open")
)

// CryptSetup manages encrypted devices.
type CryptSetup struct {
	nameInit func(name string) (cryptDevice, error)
	pathInit func(path string) (cryptDevice, error)
	device   cryptDevice
}

// New creates a new CryptSetup.
// Before first use, call Init() or InitByName() to open a crypt device.
func New() *CryptSetup {
	return &CryptSetup{
		nameInit: initByName,
		pathInit: initByDevicePath,
	}
}

// Init opens a crypt device by device path.
func (c *CryptSetup) Init(devicePath string) (free func(), err error) {
	packageLock.Lock()
	defer packageLock.Unlock()
	if c.device != nil {
		return nil, errDeviceAlreadyOpen
	}
	device, err := c.pathInit(devicePath)
	if err != nil {
		return nil, fmt.Errorf("init cryptsetup by device path %q: %w", devicePath, err)
	}
	c.device = device
	return c.Free, nil
}

// InitByName opens an active crypt device using its mapped name.
func (c *CryptSetup) InitByName(name string) (free func(), err error) {
	packageLock.Lock()
	defer packageLock.Unlock()
	if c.device != nil {
		return nil, errDeviceAlreadyOpen
	}
	device, err := c.nameInit(name)
	if err != nil {
		return nil, fmt.Errorf("init cryptsetup by name %q: %w", name, err)
	}
	c.device = device
	return c.Free, nil
}

// Free frees resources from a previously opened crypt device.
func (c *CryptSetup) Free() {
	packageLock.Lock()
	defer packageLock.Unlock()
	if c.device != nil {
		c.device.Free()
		c.device = nil
	}
}

// ActivateByPassphrase actives a crypt device using a passphrase.
func (c *CryptSetup) ActivateByPassphrase(deviceName string, keyslot int, passphrase string, flags int) error {
	packageLock.Lock()
	defer packageLock.Unlock()
	if c.device == nil {
		return errDeviceNotOpen
	}
	if err := c.device.ActivateByPassphrase(deviceName, keyslot, passphrase, flags); err != nil {
		return fmt.Errorf("activating crypt device %q using passphrase: %w", deviceName, err)
	}
	return nil
}

// ActivateByVolumeKey activates a crypt device using a volume key.
// Set volumeKey to empty string to use the internal key.
func (c *CryptSetup) ActivateByVolumeKey(deviceName, volumeKey string, volumeKeySize, flags int) error {
	packageLock.Lock()
	defer packageLock.Unlock()
	if c.device == nil {
		return errDeviceNotOpen
	}
	if err := c.device.ActivateByVolumeKey(deviceName, volumeKey, volumeKeySize, flags); err != nil {
		return fmt.Errorf("activating crypt device %q using volume key: %w", deviceName, err)
	}
	return nil
}

// Deactivate deactivates a crypt device, removing the mapped device.
func (c *CryptSetup) Deactivate(deviceName string) error {
	packageLock.Lock()
	defer packageLock.Unlock()
	if c.device == nil {
		return errDeviceNotOpen
	}
	if err := c.device.Deactivate(deviceName); err != nil {
		return fmt.Errorf("deactivating crypt device %q: %w", deviceName, err)
	}
	return nil
}

// Format formats a disk as a LUKS2 crypt device.
// Optionally set integrity to true to enable dm-integrity for the device.
func (c *CryptSetup) Format(integrity bool) error {
	packageLock.Lock()
	defer packageLock.Unlock()
	if c.device == nil {
		return errDeviceNotOpen
	}
	if err := format(c.device, integrity); err != nil {
		return fmt.Errorf("formatting crypt device %q: %w", c.device.GetDeviceName(), err)
	}
	return nil
}

// GetDeviceName gets the path to the underlying device.
func (c *CryptSetup) GetDeviceName() string {
	return c.device.GetDeviceName()
}

// GetUUID gets the device's LUKS2 UUID.
// The UUID is returned in lowercase.
func (c *CryptSetup) GetUUID() (string, error) {
	packageLock.Lock()
	defer packageLock.Unlock()
	if c.device == nil {
		return "", errDeviceNotOpen
	}
	uuid := c.device.GetUUID()
	if uuid == "" {
		return "", fmt.Errorf("unable to get UUID for device %q", c.device.GetDeviceName())
	}
	return strings.ToLower(uuid), nil
}

// KeyslotAddByVolumeKey adds a key slot to a device, allowing later activations using the chosen passphrase.
// Set volumeKey to empty string to use the internal key.
func (c *CryptSetup) KeyslotAddByVolumeKey(keyslot int, volumeKey string, passphrase string) error {
	packageLock.Lock()
	defer packageLock.Unlock()
	if c.device == nil {
		return errDeviceNotOpen
	}
	if err := c.device.KeyslotAddByVolumeKey(keyslot, volumeKey, passphrase); err != nil {
		return fmt.Errorf("adding keyslot to device %q: %w", c.device.GetDeviceName(), err)
	}
	return nil
}

// KeyslotChangeByPassphrase changes the passphrase for a keyslot.
func (c *CryptSetup) KeyslotChangeByPassphrase(currentKeyslot, newKeyslot int, currentPassphrase, newPassphrase string) error {
	packageLock.Lock()
	defer packageLock.Unlock()
	if c.device == nil {
		return errDeviceNotOpen
	}
	if err := c.device.KeyslotChangeByPassphrase(currentKeyslot, newKeyslot, currentPassphrase, newPassphrase); err != nil {
		return fmt.Errorf("updating passphrase for device %q: %w", c.device.GetDeviceName(), err)
	}
	return nil
}

// LoadLUKS2 loads the device as LUKS2 crypt device.
func (c *CryptSetup) LoadLUKS2() error {
	if err := loadLUKS2(c.device); err != nil {
		return fmt.Errorf("loading LUKS2 crypt device %q: %w", c.device.GetDeviceName(), err)
	}
	return nil
}

// Resize resizes a device to the given size.
// name must be equal to the mapped device name.
// Set newSize to 0 to use the maximum available size.
func (c *CryptSetup) Resize(name string, newSize uint64) error {
	packageLock.Lock()
	defer packageLock.Unlock()
	if c.device == nil {
		return errDeviceNotOpen
	}
	if err := c.device.Resize(name, newSize); err != nil {
		return fmt.Errorf("resizing crypt device %q: %w", c.device.GetDeviceName(), err)
	}
	return nil
}

// Wipe overwrites the device with zeros to initialize integrity checksums.
func (c *CryptSetup) Wipe(
	name string, blockWipeSize int, flags int, logCallback func(size, offset uint64), logFrequency time.Duration,
) (err error) {
	packageLock.Lock()
	defer packageLock.Unlock()
	if c.device == nil {
		return errDeviceNotOpen
	}

	// Active temporary device to perform wipe on
	tmpDevice := tmpDevicePrefix + name
	if err := c.device.ActivateByVolumeKey(tmpDevice, "", 0, wipeFlags); err != nil {
		return fmt.Errorf("trying to activate temporary dm-crypt volume: %w", err)
	}
	defer func() {
		if deactivateErr := c.device.Deactivate(tmpDevice); deactivateErr != nil {
			err = errors.Join(err, fmt.Errorf("deactivating temporary device %q: %w", tmpDevice, deactivateErr))
		}
	}()

	// Set up non-blocking progress callback.
	ticker := time.NewTicker(logFrequency)
	firstReq := make(chan struct{}, 1)
	firstReq <- struct{}{}
	defer ticker.Stop()

	progressCallback := func(size, offset uint64) int {
		select {
		case <-firstReq:
			logCallback(size, offset)
		case <-ticker.C:
			logCallback(size, offset)
		default:
		}
		return 0
	}

	if err := c.device.Wipe(mappedDevicePath+tmpDevice, wipePattern, 0, 0, blockWipeSize, flags, progressCallback); err != nil {
		return fmt.Errorf("wiping disk of device %q: %w", c.device.GetDeviceName(), err)
	}
	return nil
}

type cryptDevice interface {
	ActivateByPassphrase(deviceName string, keyslot int, passphrase string, flags int) error
	ActivateByVolumeKey(deviceName, volumeKey string, volumeKeySize, flags int) error
	Deactivate(deviceName string) error
	GetDeviceName() string
	GetUUID() string
	Free() bool
	KeyslotAddByVolumeKey(keyslot int, volumeKey string, passphrase string) error
	KeyslotChangeByPassphrase(currentKeyslot, newKeyslot int, currentPassphrase, newPassphrase string) error
	Resize(name string, newSize uint64) error
	Wipe(devicePath string, pattern int, offset, length uint64, wipeBlockSize int, flags int, progress func(size, offset uint64) int) error
}
