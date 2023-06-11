//go:build linux && cgo

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package diskencryption

import (
	"errors"
	"fmt"
	"sync"

	"github.com/martinjungblut/go-cryptsetup"
	"github.com/spf13/afero"
)

const (
	stateMapperDevice = "state"
	initialKeyPath    = "/run/cryptsetup-keys.d/state.key"
	keyslot           = 0
)

var (
	// packageLock is needed to block concurrent use of package functions, since libcryptsetup is not thread safe.
	// See: https://gitlab.com/cryptsetup/cryptsetup/-/issues/710
	// 		https://stackoverflow.com/questions/30553386/cryptsetup-backend-safe-with-multithreading
	packageLock          = sync.Mutex{}
	errDeviceNotOpen     = errors.New("cryptdevice not open")
	errDeviceAlreadyOpen = errors.New("cryptdevice already open")
)

// Cryptsetup manages the encrypted state mapper device.
type Cryptsetup struct {
	fs afero.Fs
	// device     cryptdevice
	initByName initByName
}

type OpenCryptsetup struct {
	*Cryptsetup
	device cryptdevice
}

// New creates a new Cryptsetup.
func New() *Cryptsetup {
	return &Cryptsetup{
		fs: afero.NewOsFs(),
		initByName: func(name string) (cryptdevice, error) {
			return cryptsetup.InitByName(name)
		},
	}
}

// Open opens the cryptdevice.
func (c *Cryptsetup) Open() (*OpenCryptsetup, error) {
	packageLock.Lock()
	defer packageLock.Unlock()
	var err error
	device, err := c.initByName(stateMapperDevice)
	if err != nil {
		return nil, fmt.Errorf("initializing crypt device for mapped device %q: %w", stateMapperDevice, err)
	}
	return &OpenCryptsetup{c, device}, nil
}

// Close closes the cryptdevice.
func (c *OpenCryptsetup) Close() error {
	packageLock.Lock()
	defer packageLock.Unlock()
	//if c.device == nil {
	//	return errDeviceNotOpen
	//}
	c.device.Free()
	c.device = nil // How to prevent close from being called twice? Return closeFn in constructor which suggests defer closeFn() pattern?

	return nil
}

func (c *OpenCryptsetup) UUID() (string, error) {
	packageLock.Lock()
	defer packageLock.Unlock()
	if c.device == nil {
		return "", errDeviceNotOpen
	}
	uuid := c.device.GetUUID()
	if uuid == "" {
		return "", fmt.Errorf("unable to get UUID for mapped device %q", stateMapperDevice)
	}
	return uuid, nil
}

// UpdatePassphrase switches the initial random passphrase of the mapped crypt device to a permanent passphrase.
// Only works after calling Open().
func (c *OpenCryptsetup) UpdatePassphrase(passphrase string) error {
	packageLock.Lock()
	defer packageLock.Unlock()
	if c.device == nil {
		return errDeviceNotOpen
	}
	initialPassphrase, err := c.getInitialPassphrase()
	if err != nil {
		return err
	}
	if err := c.device.KeyslotChangeByPassphrase(keyslot, keyslot, initialPassphrase, passphrase); err != nil {
		return fmt.Errorf("changing passphrase for mapped device %q: %w", stateMapperDevice, err)
	}
	return nil
}

// getInitialPassphrase retrieves the initial passphrase used on first boot.
func (c *Cryptsetup) getInitialPassphrase() (string, error) {
	passphrase, err := afero.ReadFile(c.fs, initialKeyPath)
	if err != nil {
		return "", fmt.Errorf("reading first boot encryption passphrase from disk: %w", err)
	}
	return string(passphrase), nil
}

type cryptdevice interface {
	GetUUID() string
	KeyslotChangeByPassphrase(currentKeyslot int, newKeyslot int, currentPassphrase string, newPassphrase string) error
	Free() bool
}

type initByName func(name string) (cryptdevice, error)
