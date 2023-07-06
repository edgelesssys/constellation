//go:build linux && cgo

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cryptmapper

import (
	"fmt"

	ccryptsetup "github.com/edgelesssys/constellation/v2/internal/cryptsetup"
	cryptsetup "github.com/martinjungblut/go-cryptsetup"
	mount "k8s.io/mount-utils"
	utilexec "k8s.io/utils/exec"
)

const (
	resizeFlags = cryptsetup.CRYPT_ACTIVATE_KEYRING_KEY | ccryptsetup.ReadWriteQueueBypass
)

func init() {
	cryptsetup.SetDebugLevel(cryptsetup.CRYPT_LOG_NORMAL)
	cryptsetup.SetLogCallback(func(_ int, message string) { fmt.Printf("libcryptsetup: %s\n", message) })
}

func getDiskFormat(disk string) (string, error) {
	mountUtil := &mount.SafeFormatAndMount{Exec: utilexec.New()}
	return mountUtil.GetDiskFormat(disk)
}
