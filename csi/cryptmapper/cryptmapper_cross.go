//go:build !linux || !cgo

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cryptmapper

import (
	"errors"

	ccryptsetup "github.com/edgelesssys/constellation/v2/internal/cryptsetup"
)

const (
	resizeFlags = 0x800 | ccryptsetup.ReadWriteQueueBypass
)

func getDiskFormat(_ string) (string, error) {
	return "", errors.New("getDiskFormat requires building with CGO enabled")
}
