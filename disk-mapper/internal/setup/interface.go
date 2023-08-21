/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package setup

import (
	"context"
	"io/fs"
	"os"

	"github.com/edgelesssys/constellation/v2/internal/cloud/metadata"
)

// Mounter is an interface for mount and unmount operations.
type Mounter interface {
	Mount(source string, target string, fstype string, flags uintptr, data string) error
	Unmount(target string, flags int) error
	MkdirAll(path string, perm fs.FileMode) error
}

// DeviceMapper is an interface for device mapping operations.
type DeviceMapper interface {
	DiskUUID() (string, error)
	FormatDisk(passphrase string) error
	MapDisk(target string, passphrase string) error
	UnmapDisk(target string) error
}

// ConfigurationGenerator is an interface for generating systemd-cryptsetup@.service unit files.
type ConfigurationGenerator interface {
	Generate(volumeName, encryptedDevice, keyFile, options string) error
}

// MetadataAPI is an interface for accessing cloud metadata.
type MetadataAPI interface {
	metadata.InstanceSelfer
	metadata.InstanceLister
	GetLoadBalancerEndpoint(ctx context.Context) (host, port string, err error)
}

// RecoveryDoer is an interface to perform key recovery operations.
// Calls to Do may be blocking, and if successful return a passphrase and measurementSecret.
type RecoveryDoer interface {
	Do(uuid, endpoint string) (passphrase, measurementSecret []byte, err error)
}

// DiskMounter uses the syscall package to mount disks.
type DiskMounter struct{}

// MkdirAll uses os.MkdirAll to create the directory.
func (m DiskMounter) MkdirAll(path string, perm fs.FileMode) error {
	return os.MkdirAll(path, perm)
}
