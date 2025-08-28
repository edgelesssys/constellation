/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: BUSL-1.1
*/

/*
Package cryptsetup provides a wrapper around libcryptsetup.
The package is used to manage encrypted disks for Constellation.

Since libcryptsetup is not thread safe, this package uses a global lock to prevent concurrent use.
There should only be one instance using this package per process.
*/
package cryptsetup

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"
)

const (
	// ConstellationStateDiskTokenID is the ID of Constellation's state disk token.
	ConstellationStateDiskTokenID = 0
	// SetDiskInitialized is a flag to set the Constellation state disk token to initialized.
	SetDiskInitialized = true
	// SetDiskNotInitialized is a flag to set the Constellation state disk token to not initialized.
	SetDiskNotInitialized = false

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
	nameInit func(name string) (deviceDetachedHeader, deviceAttachedHeader cryptDevice, headerDevice string, headerFile string, err error)
	pathInit func(path string) (deviceDetachedHeader, deviceAttachedHeader cryptDevice, headerDevice string, headerFile string, err error)
	// deviceWithDetachedHeader is the cryptsetup device with detached header we are working on.
	deviceWithDetachedHeader cryptDevice
	// deviceWithAttachedHeader is a cryptsetup device loaded without a separate, detached header.
	// If this is not a fresh disk, this device is purely used to write back changes that affect the header to the original disk.
	deviceWithAttachedHeader cryptDevice
	// headerDevice is the name of the loopback device containing the detached cryptsetup header.
	headerDevice string
	// headerFile is the path to the file containing the detached cryptsetup header.
	// The file is mounted as a loopback device on "headerDevice".
	headerFile string
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
	if c.hasDetachedHeaderDevice() || c.hasAttachedHeaderDevice() {
		return nil, errDeviceAlreadyOpen
	}
	deviceDetachedHeader, deviceAttachedHeader, headerDevice, headerFile, err := c.pathInit(devicePath)
	if err != nil {
		return nil, fmt.Errorf("init cryptsetup by device path %q: %w", devicePath, err)
	}
	c.deviceWithDetachedHeader = deviceDetachedHeader
	c.deviceWithAttachedHeader = deviceAttachedHeader
	c.headerDevice = headerDevice
	c.headerFile = headerFile
	return c.Free, nil
}

// InitByName opens an active crypt device using its mapped name.
func (c *CryptSetup) InitByName(name string) (free func(), err error) {
	packageLock.Lock()
	defer packageLock.Unlock()
	if c.hasDetachedHeaderDevice() || c.hasAttachedHeaderDevice() {
		return nil, errDeviceAlreadyOpen
	}
	deviceDetachedHeader, deviceAttachedHeader, headerDevice, headerFile, err := c.nameInit(name)
	if err != nil {
		return nil, fmt.Errorf("init cryptsetup by name %q: %w", name, err)
	}
	c.deviceWithDetachedHeader = deviceDetachedHeader
	c.deviceWithAttachedHeader = deviceAttachedHeader
	c.headerDevice = headerDevice
	c.headerFile = headerFile
	return c.Free, nil
}

// Free frees resources from a previously opened crypt device.
func (c *CryptSetup) Free() {
	packageLock.Lock()
	defer packageLock.Unlock()
	c.free()
}

// ActivateByPassphrase actives a crypt device using a passphrase.
func (c *CryptSetup) ActivateByPassphrase(deviceName string, keyslot int, passphrase string, flags int) error {
	packageLock.Lock()
	defer packageLock.Unlock()
	if !c.hasDetachedHeaderDevice() {
		if err := c.reload(); err != nil {
			return fmt.Errorf("re-loading crypt device for activation: %w", err)
		}
	}
	if err := c.deviceWithDetachedHeader.ActivateByPassphrase(deviceName, keyslot, passphrase, flags); err != nil {
		return fmt.Errorf("activating crypt device %q using passphrase: %w", deviceName, err)
	}
	return nil
}

// ActivateByVolumeKey activates a crypt device using a volume key.
// Set volumeKey to empty string to use the internal key.
func (c *CryptSetup) ActivateByVolumeKey(deviceName, volumeKey string, volumeKeySize, flags int) error {
	packageLock.Lock()
	defer packageLock.Unlock()
	if !c.hasDetachedHeaderDevice() {
		if err := c.reload(); err != nil {
			return fmt.Errorf("re-loading crypt device for activation: %w", err)
		}
	}
	if err := c.deviceWithDetachedHeader.ActivateByVolumeKey(deviceName, volumeKey, volumeKeySize, flags); err != nil {
		return fmt.Errorf("activating crypt device %q using volume key: %w", deviceName, err)
	}
	return nil
}

// Deactivate deactivates a crypt device, removing the mapped device.
func (c *CryptSetup) Deactivate(deviceName string) error {
	packageLock.Lock()
	defer packageLock.Unlock()
	if !c.hasDetachedHeaderDevice() {
		return errDeviceNotOpen
	}
	if err := c.deviceWithDetachedHeader.Deactivate(deviceName); err != nil {
		return fmt.Errorf("deactivating crypt device %q: %w", deviceName, err)
	}
	return nil
}

// Format formats a disk as a LUKS2 crypt device.
// Optionally set integrity to true to enable dm-integrity for the device.
func (c *CryptSetup) Format(integrity bool) error {
	packageLock.Lock()
	defer packageLock.Unlock()
	var device cryptDevice

	// If we are re-formatting an existing device, we start from scratch without a detached header
	if c.hasAttachedHeaderDevice() {
		if c.deviceWithDetachedHeader != nil {
			c.deviceWithDetachedHeader.Free()
			c.deviceWithDetachedHeader = nil
		}
		if c.headerDevice != "" {
			_ = detachLoopbackDevice(c.headerDevice)
			c.headerDevice = ""
		}
		c.headerFile = ""
		device = c.deviceWithAttachedHeader
	} else {
		return errDeviceNotOpen
	}

	if err := format(device, integrity); err != nil {
		return fmt.Errorf("formatting crypt device %q: %w", device.GetDeviceName(), err)
	}
	if err := c.createHeaderBackup(); err != nil {
		return err
	}
	return nil
}

// GetDeviceName gets the path to the underlying device.
func (c *CryptSetup) GetDeviceName() string {
	packageLock.Lock()
	defer packageLock.Unlock()
	device, err := c.getActiveDevice()
	if err != nil {
		return ""
	}
	return device.GetDeviceName()
}

// GetUUID gets the device's LUKS2 UUID.
// The UUID is returned in lowercase.
func (c *CryptSetup) GetUUID() (string, error) {
	packageLock.Lock()
	defer packageLock.Unlock()
	device, err := c.getActiveDevice()
	if err != nil {
		return "", err
	}
	uuid := device.GetUUID()
	if uuid == "" {
		return "", fmt.Errorf("unable to get UUID for device %q", device.GetDeviceName())
	}
	return strings.ToLower(uuid), nil
}

// KeyslotAddByVolumeKey adds a key slot to a device, allowing later activations using the chosen passphrase.
// Set volumeKey to empty string to use the internal key.
func (c *CryptSetup) KeyslotAddByVolumeKey(keyslot int, volumeKey string, passphrase string) error {
	packageLock.Lock()
	defer packageLock.Unlock()
	device, err := c.getActiveDevice()
	if err != nil {
		return err
	}
	if err := device.KeyslotAddByVolumeKey(keyslot, volumeKey, passphrase); err != nil {
		return fmt.Errorf("adding keyslot to device %q: %w", device.GetDeviceName(), err)
	}
	if err := c.createHeaderBackup(); err != nil {
		return err
	}
	return nil
}

// KeyslotChangeByPassphrase changes the passphrase for a keyslot.
func (c *CryptSetup) KeyslotChangeByPassphrase(currentKeyslot, newKeyslot int, currentPassphrase, newPassphrase string) error {
	packageLock.Lock()
	defer packageLock.Unlock()
	device, err := c.getActiveDevice()
	if err != nil {
		return err
	}
	if err := device.KeyslotChangeByPassphrase(currentKeyslot, newKeyslot, currentPassphrase, newPassphrase); err != nil {
		return fmt.Errorf("updating passphrase for device %q: %w", device.GetDeviceName(), err)
	}
	if err := c.createHeaderBackup(); err != nil {
		return err
	}
	return nil
}

// LoadLUKS2 loads the device as LUKS2 crypt device.
func (c *CryptSetup) LoadLUKS2() error {
	packageLock.Lock()
	defer packageLock.Unlock()
	if c.hasDetachedHeaderDevice() {
		if err := loadLUKS2(c.deviceWithDetachedHeader); err != nil {
			return fmt.Errorf("loading LUKS2 crypt device %q: %w", c.deviceWithDetachedHeader.GetDeviceName(), err)
		}
		return nil
	}
	return errors.New("cannot load LUKS2 on device with attached header")
}

// Resize resizes a device to the given size.
// name must be equal to the mapped device name.
// Set newSize to 0 to use the maximum available size.
func (c *CryptSetup) Resize(name string, newSize uint64) error {
	packageLock.Lock()
	defer packageLock.Unlock()
	device, err := c.getActiveDevice()
	if err != nil {
		return err
	}
	if err := device.Resize(name, newSize); err != nil {
		return fmt.Errorf("resizing crypt device %q: %w", device.GetDeviceName(), err)
	}
	if err := c.createHeaderBackup(); err != nil {
		return err
	}
	return nil
}

// TokenJSONGet gets the JSON data for a token.
func (c *CryptSetup) TokenJSONGet(token int) (string, error) {
	packageLock.Lock()
	defer packageLock.Unlock()
	device, err := c.getActiveDevice()
	if err != nil {
		return "", err
	}
	json, err := device.TokenJSONGet(token)
	if err != nil {
		return "", fmt.Errorf("getting JSON data for token %d: %w", token, err)
	}
	return json, nil
}

// TokenJSONSet sets the JSON data for a token.
// The JSON data must be a valid LUKS2 token.
// Required fields are:
//   - type [string] the token type (tokens with luks2- prefix are reserved)
//   - keyslots [array] the array of keyslot objects names that are assigned to the token
//
// Returns the allocated token ID on success.
func (c *CryptSetup) TokenJSONSet(token int, json string) (int, error) {
	packageLock.Lock()
	defer packageLock.Unlock()
	device, err := c.getActiveDevice()
	if err != nil {
		return -1, err
	}

	tokenID, err := device.TokenJSONSet(token, json)
	if err != nil {
		return -1, fmt.Errorf("setting JSON data for token %d: %w", token, err)
	}
	if err := c.createHeaderBackup(); err != nil {
		return -1, err
	}
	return tokenID, nil
}

// SetConstellationStateDiskToken sets the Constellation state disk token.
func (c *CryptSetup) SetConstellationStateDiskToken(diskIsInitialized bool) error {
	packageLock.Lock()
	defer packageLock.Unlock()
	token := constellationLUKS2Token{
		Type:              "constellation-state-disk",
		Keyslots:          []string{},
		DiskIsInitialized: diskIsInitialized,
	}
	json, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("marshaling token: %w", err)
	}

	device, err := c.getActiveDevice()
	if err != nil {
		return err
	}

	if _, err := device.TokenJSONSet(ConstellationStateDiskTokenID, string(json)); err != nil {
		return fmt.Errorf("setting token: %w", err)
	}
	if err := c.createHeaderBackup(); err != nil {
		return err
	}
	return nil
}

// ConstellationStateDiskTokenIsInitialized returns true if the Constellation state disk token is set to initialized.
func (c *CryptSetup) ConstellationStateDiskTokenIsInitialized() bool {
	packageLock.Lock()
	defer packageLock.Unlock()
	device, err := c.getActiveDevice()
	if err != nil {
		return false
	}

	stateDiskToken, err := device.TokenJSONGet(ConstellationStateDiskTokenID)
	if err != nil {
		return false
	}
	var token constellationLUKS2Token
	if err := json.Unmarshal([]byte(stateDiskToken), &token); err != nil {
		return false
	}
	return token.DiskIsInitialized
}

// Wipe overwrites the device with zeros to initialize integrity checksums.
func (c *CryptSetup) Wipe(
	name string, blockWipeSize int, flags int, logCallback func(size, offset uint64), logFrequency time.Duration,
) (err error) {
	packageLock.Lock()
	defer packageLock.Unlock()
	device, err := c.getActiveDevice()
	if err != nil {
		return err
	}

	// Active temporary device to perform wipe on
	tmpDevice := tmpDevicePrefix + name
	if err := device.ActivateByVolumeKey(tmpDevice, "", 0, wipeFlags); err != nil {
		return fmt.Errorf("trying to activate temporary dm-crypt volume: %w", err)
	}
	defer func() {
		if deactivateErr := device.Deactivate(tmpDevice); deactivateErr != nil {
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

	if err := device.Wipe(mappedDevicePath+tmpDevice, wipePattern, 0, 0, blockWipeSize, flags, progressCallback); err != nil {
		return fmt.Errorf("wiping disk of device %q: %w", device.GetDeviceName(), err)
	}
	if err := c.createHeaderBackup(); err != nil {
		return err
	}
	return nil
}

func (c *CryptSetup) free() {
	if c.hasDetachedHeaderDevice() {
		c.deviceWithDetachedHeader.Free()
		c.deviceWithDetachedHeader = nil
	}
	if c.hasAttachedHeaderDevice() {
		c.deviceWithAttachedHeader.Free()
		c.deviceWithAttachedHeader = nil
	}
	if c.headerDevice != "" {
		_ = detachLoopbackDevice(c.headerDevice)
	}
	if c.headerFile != "" {
		c.headerFile = ""
	}
}

func (c *CryptSetup) reload() error {
	if !c.hasAttachedHeaderDevice() {
		return errDeviceNotOpen
	}

	backingDevice := c.deviceWithAttachedHeader.GetDeviceName()
	c.free()
	var err error
	c.deviceWithDetachedHeader, c.deviceWithAttachedHeader, c.headerDevice, c.headerFile, err = c.pathInit(backingDevice)
	if err != nil {
		return fmt.Errorf("re-loading crypt device: %w", err)
	}

	if !c.hasDetachedHeaderDevice() {
		return errors.New("failed to reload device without detached header")
	}

	if err := loadLUKS2(c.deviceWithDetachedHeader); err != nil {
		return err
	}

	return nil
}

// getActiveDevice returns a handle to the active cryptsetup device with detached header if set,
// or one with attached header otherwise.
func (c *CryptSetup) getActiveDevice() (cryptDevice, error) {
	if c.hasDetachedHeaderDevice() {
		return c.deviceWithDetachedHeader, nil
	}
	if c.hasAttachedHeaderDevice() {
		return c.deviceWithAttachedHeader, nil
	}
	return nil, errDeviceNotOpen
}

// hasDetachedHeaderDevice checks if the value of the [CryptSetup.deviceWithDetachedHeader] interface is not nil.
func (c *CryptSetup) hasDetachedHeaderDevice() bool {
	return c.deviceWithDetachedHeader != nil && !reflect.ValueOf(c.deviceWithDetachedHeader).IsNil()
}

// hasAttachedHeaderDevice checks if the value of the [CryptSetup.deviceWithAttachedHeader] interface is not nil.
func (c *CryptSetup) hasAttachedHeaderDevice() bool {
	return c.deviceWithAttachedHeader != nil && !reflect.ValueOf(c.deviceWithAttachedHeader).IsNil()
}

// createHeaderBackup creates a backup of the detached header, and saves it back to the original device.
func (c *CryptSetup) createHeaderBackup() error {
	if c.hasDetachedHeaderDevice() && c.headerFile != "" {
		if err := headerBackup(c.deviceWithDetachedHeader, c.headerFile); err != nil {
			return fmt.Errorf("creating header backup for device %q: %w", c.deviceWithDetachedHeader.GetDeviceName(), err)
		}
	}
	if c.hasAttachedHeaderDevice() && c.headerFile != "" {
		if err := headerRestore(c.deviceWithAttachedHeader, c.headerFile); err != nil {
			return fmt.Errorf("restoring header for device %q (with attached header): %w", c.deviceWithDetachedHeader.GetDeviceName(), err)
		}
	}
	return nil
}

type constellationLUKS2Token struct {
	Type              string   `json:"type"`
	Keyslots          []string `json:"keyslots"`
	DiskIsInitialized bool     `json:"diskIsInitialized"`
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
	TokenJSONGet(token int) (string, error)
	TokenJSONSet(token int, json string) (int, error)
	Wipe(devicePath string, pattern int, offset, length uint64, wipeBlockSize int, flags int, progress func(size, offset uint64) int) error
}
