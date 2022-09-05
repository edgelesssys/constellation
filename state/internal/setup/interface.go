/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package setup

import (
	"io/fs"
	"os"
	"syscall"
)

// Mounter is an interface for mount and unmount operations.
type Mounter interface {
	Mount(source string, target string, fstype string, flags uintptr, data string) error
	Unmount(target string, flags int) error
	MkdirAll(path string, perm fs.FileMode) error
}

// DeviceMapper is an interface for device mapping operations.
type DeviceMapper interface {
	DiskUUID() string
	FormatDisk(passphrase string) error
	MapDisk(target string, passphrase string) error
	UnmapDisk(target string) error
}

// KeyWaiter is an interface to request and wait for disk decryption keys.
type KeyWaiter interface {
	WaitForDecryptionKey(uuid, addr string) (diskKey, measurementSecret []byte, err error)
	ResetKey()
}

// ConfigurationGenerator is an interface for generating systemd-cryptsetup@.service unit files.
type ConfigurationGenerator interface {
	Generate(volumeName, encryptedDevice, keyFile, options string) error
}

// DiskMounter uses the syscall package to mount disks.
type DiskMounter struct{}

// Mount performs a mount syscall.
func (m DiskMounter) Mount(source string, target string, fstype string, flags uintptr, data string) error {
	return syscall.Mount(source, target, fstype, flags, data)
}

// Unmount performs an unmount syscall.
func (m DiskMounter) Unmount(target string, flags int) error {
	return syscall.Unmount(target, flags)
}

// MkdirAll uses os.MkdirAll to create the directory.
func (m DiskMounter) MkdirAll(path string, perm fs.FileMode) error {
	return os.MkdirAll(path, perm)
}
