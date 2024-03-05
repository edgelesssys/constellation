/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package diskencryption handles interaction with a node's state disk.
package diskencryption

import (
	"fmt"

	"github.com/edgelesssys/constellation/v2/internal/cryptsetup"
	"github.com/spf13/afero"
)

const (
	stateMapperDevice = "state"
	initialKeyPath    = "/run/cryptsetup-keys.d/state.key"
	keyslot           = 0
)

// DiskEncryption manages the encrypted state mapper device.
type DiskEncryption struct {
	fs     afero.Fs
	device cryptdevice
}

// New creates a new Cryptsetup.
func New() *DiskEncryption {
	return &DiskEncryption{
		fs:     afero.NewOsFs(),
		device: cryptsetup.New(),
	}
}

// Open opens the cryptdevice.
func (c *DiskEncryption) Open() (free func(), err error) {
	return c.device.InitByName(stateMapperDevice)
}

// UUID gets the device's UUID.
// Only works after calling Open().
func (c *DiskEncryption) UUID() (string, error) {
	return c.device.GetUUID()
}

// UpdatePassphrase switches the initial random passphrase of the mapped crypt device to a permanent passphrase.
// Only works after calling Open().
func (c *DiskEncryption) UpdatePassphrase(passphrase string) error {
	initialPassphrase, err := c.getInitialPassphrase()
	if err != nil {
		return err
	}
	if err := c.device.KeyslotChangeByPassphrase(keyslot, keyslot, initialPassphrase, passphrase); err != nil {
		return err
	}

	// Set token as initialized.
	return c.device.SetConstellationStateDiskToken(cryptsetup.SetDiskInitialized)
}

// MarkDiskForReset marks the state disk as not initialized.
func (c *DiskEncryption) MarkDiskForReset() error {
	return c.device.SetConstellationStateDiskToken(cryptsetup.SetDiskNotInitialized)
}

// getInitialPassphrase retrieves the initial passphrase used on first boot.
func (c *DiskEncryption) getInitialPassphrase() (string, error) {
	passphrase, err := afero.ReadFile(c.fs, initialKeyPath)
	if err != nil {
		return "", fmt.Errorf("reading first boot encryption passphrase from disk: %w", err)
	}
	return string(passphrase), nil
}

type cryptdevice interface {
	InitByName(name string) (func(), error)
	GetUUID() (string, error)
	KeyslotChangeByPassphrase(currentKeyslot int, newKeyslot int, currentPassphrase string, newPassphrase string) error
	SetConstellationStateDiskToken(bool) error
}
