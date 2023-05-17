//go:build !linux || !cgo

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cryptmapper

import (
	"context"
	"errors"
)

// deviceMapper is an interface for device mapper methods.
type deviceMapper interface{}

// CryptDevice is a wrapper for cryptsetup.Device.
type CryptDevice struct{}

// Init initializes a crypt device backed by 'devicePath'.
// This function errors if CGO is disabled.
func (c *CryptDevice) Init(_ string) error {
	return errors.New("using cryptmapper requires building with CGO")
}

// InitByName initializes a crypt device from provided active device 'name'.
// This function panics if CGO is disabled.
func (c *CryptDevice) InitByName(_ string) error {
	return errors.New("using cryptmapper requires building with CGO")
}

// Free releases crypt device context and used memory.
// This function does nothing if CGO is disabled.
func (c *CryptDevice) Free() bool {
	return false
}

// CryptMapper manages dm-crypt volumes.
type CryptMapper struct{}

// New initializes a new CryptMapper with the given kms client and key-encryption-key ID.
// This function panics if CGO is disabled.
func New(_ KeyCreator, _ deviceMapper) *CryptMapper {
	panic("CGO is disabled but requested CryptMapper instance")
}

// CloseCryptDevice closes the crypt device mapped for volumeID.
// This function errors if CGO is disabled.
func (c *CryptMapper) CloseCryptDevice(_ string) error {
	return errors.New("using cryptmapper requires building with CGO")
}

// OpenCryptDevice maps the volume at source to the crypt device identified by volumeID.
// This function errors if CGO is disabled.
func (c *CryptMapper) OpenCryptDevice(_ context.Context, _, _ string, _ bool) (string, error) {
	return "", errors.New("using cryptmapper requires building with CGO")
}

// ResizeCryptDevice resizes the underlying crypt device and returns the mapped device path.
// This function errors if CGO is disabled.
func (c *CryptMapper) ResizeCryptDevice(_ context.Context, _ string) (string, error) {
	return "", errors.New("using cryptmapper requires building with CGO")
}

// GetDevicePath returns the device path of a mapped crypt device.
// This function errors if CGO is disabled.
func (c *CryptMapper) GetDevicePath(_ string) (string, error) {
	return "", errors.New("using cryptmapper requires building with CGO")
}

// IsIntegrityFS checks if the fstype string contains an integrity suffix.
// This function does nothing if CGO is disabled.
func IsIntegrityFS(_ string) (string, bool) {
	return "", false
}
