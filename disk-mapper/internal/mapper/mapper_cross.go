//go:build !linux || !cgo

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package mapper

import (
	"errors"

	"github.com/edgelesssys/constellation/v2/internal/logger"
)

// Mapper handles actions for formatting and mapping crypt devices.
type Mapper struct{}

// New creates a new crypt device for the device at path.
// This function errors if CGO is disabled.
func New(_ string, _ *logger.Logger) (*Mapper, error) {
	return nil, errors.New("using mapper requires building with CGO")
}

// Close closes and frees memory allocated for the crypt device.
// This function errors if CGO is disabled.
func (m *Mapper) Close() error {
	return errors.New("using mapper requires building with CGO")
}

// IsLUKSDevice returns true if the device is formatted as a LUKS device.
// This function does nothing if CGO is disabled.
func (m *Mapper) IsLUKSDevice() bool {
	return false
}

// DiskUUID gets the device's UUID.
// This function does nothing if CGO is disabled.
func (m *Mapper) DiskUUID() string {
	return ""
}

// FormatDisk formats the disk and adds passphrase in keyslot 0.
// This function errors if CGO is disabled.
func (m *Mapper) FormatDisk(_ string) error {
	return errors.New("using mapper requires building with CGO")
}

// MapDisk maps a crypt device to /dev/mapper/target using the provided passphrase.
// This function errors if CGO is disabled.
func (m *Mapper) MapDisk(_, _ string) error {
	return errors.New("using mapper requires building with CGO")
}

// UnmapDisk removes the mapping of target.
// This function errors if CGO is disabled.
func (m *Mapper) UnmapDisk(_ string) error {
	return errors.New("using mapper requires building with CGO")
}

// Wipe overwrites the device with zeros to initialize integrity checksums.
// This function errors if CGO is disabled.
func (m *Mapper) Wipe(_ int) error {
	return errors.New("using mapper requires building with CGO")
}
