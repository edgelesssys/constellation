package core

import (
	"fmt"
	"strings"
)

// GetDiskUUID gets the disk's UUID.
func (c *Core) GetDiskUUID() (string, error) {
	if err := c.encryptedDisk.Open(); err != nil {
		return "", fmt.Errorf("retrieving uuid of encrypted disk: cannot open disk: %w", err)
	}
	defer c.encryptedDisk.Close()
	uuid, err := c.encryptedDisk.UUID()
	if err != nil {
		return "", fmt.Errorf("cannot retrieve uuid of disk: %w", err)
	}
	return strings.ToLower(uuid), nil
}

// UpdateDiskPassphrase switches the initial random passphrase of the encrypted disk to a permanent passphrase.
func (c *Core) UpdateDiskPassphrase(passphrase string) error {
	if err := c.encryptedDisk.Open(); err != nil {
		return fmt.Errorf("updating passphrase of encrypted disk: cannot open disk: %w", err)
	}
	defer c.encryptedDisk.Close()
	return c.encryptedDisk.UpdatePassphrase(passphrase)
}

// EncryptedDisk manages the encrypted state disk.
type EncryptedDisk interface {
	// Open prepares the underlying device for disk operations.
	Open() error
	// Close closes the underlying device.
	Close() error
	// UUID gets the device's UUID.
	UUID() (string, error)
	// UpdatePassphrase switches the initial random passphrase of the encrypted disk to a permanent passphrase.
	UpdatePassphrase(passphrase string) error
}

type EncryptedDiskFake struct{}

func (f *EncryptedDiskFake) UUID() (string, error) {
	return "fake-disk-uuid", nil
}

func (f *EncryptedDiskFake) UpdatePassphrase(passphrase string) error {
	return nil
}

func (f *EncryptedDiskFake) Open() error {
	return nil
}

func (f *EncryptedDiskFake) Close() error {
	return nil
}
