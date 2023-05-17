//go:build !linux || !cgo

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
Package diskencryption handles interaction with a node's state disk.

This package is not thread safe, since libcryptsetup is not thread safe.
There should only be one instance using this package per process.
*/
package diskencryption

import "errors"

// Cryptsetup manages the encrypted state mapper device.
type Cryptsetup struct{}

// New creates a new Cryptsetup.
// This function panics if CGO is disabled.
func New() *Cryptsetup {
	return &Cryptsetup{}
}

// Open opens the cryptdevice.
// This function does nothing if CGO is disabled.
func (c *Cryptsetup) Open() error {
	return errors.New("using cryptsetup requires building with CGO")
}

// Close closes the cryptdevice.
// This function errors if CGO is disabled.
func (c *Cryptsetup) Close() error {
	return errors.New("using cryptsetup requires building with CGO")
}

// UUID gets the device's UUID.
// This function errors if CGO is disabled.
func (c *Cryptsetup) UUID() (string, error) {
	return "", errors.New("using cryptsetup requires building with CGO")
}

// UpdatePassphrase switches the initial random passphrase of the mapped crypt device to a permanent passphrase.
// This function errors if CGO is disabled.
func (c *Cryptsetup) UpdatePassphrase(_ string) error {
	return errors.New("using cryptsetup requires building with CGO")
}
