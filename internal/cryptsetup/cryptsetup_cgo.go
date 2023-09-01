//go:build linux && cgo

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/
package cryptsetup

import (
	"errors"

	cryptsetup "github.com/malt3/purego-cryptsetup"
)

const (
	// ReadWriteQueueBypass is a flag to disable the write and read workqueues for a crypt device.
	ReadWriteQueueBypass          = cryptActivateNoReadWorkqueue | cryptActivateNoWriteWorkqueue
	cryptActivateNoReadWorkqueue  = 0x1000000
	cryptActivateNoWriteWorkqueue = 0x2000000
	wipeFlags                     = cryptsetup.CRYPT_ACTIVATE_PRIVATE | cryptsetup.CRYPT_ACTIVATE_NO_JOURNAL
	wipePattern                   = cryptsetup.CRYPT_WIPE_ZERO
)

var errInvalidType = errors.New("device is not a *cryptsetup.Device")

func format(device cryptDevice, integrity bool) error {
	switch d := device.(type) {
	case cgoFormatter:
		luks2Params := cryptsetup.LUKS2{
			SectorSize: 4096,
			PBKDFType: &cryptsetup.PbkdfType{
				// Use low memory recommendation from https://datatracker.ietf.org/doc/html/rfc9106#section-7
				Type:            "argon2id",
				TimeMs:          2000,
				Iterations:      3,
				ParallelThreads: 4,
				MaxMemoryKb:     65536, // ~64MiB
			},
		}
		genericParams := cryptsetup.GenericParams{
			Cipher:        "aes",
			CipherMode:    "xts-plain64",
			VolumeKeySize: 64, // 32*2 bytes for aes-xts-plain64 encryption
		}

		if integrity {
			luks2Params.Integrity = "hmac(sha256)"
			genericParams.VolumeKeySize += 32 // 32 bytes for hmac(sha256) integrity
		}

		return d.Format(luks2Params, genericParams)
	default:
		return errInvalidType
	}
}

func initByDevicePath(devicePath string) (cryptDevice, error) {
	return cryptsetup.Init(devicePath)
}

func initByName(name string) (cryptDevice, error) {
	return cryptsetup.InitByName(name)
}

func loadLUKS2(device cryptDevice) error {
	switch d := device.(type) {
	case cgoLoader:
		return d.Load(cryptsetup.LUKS2{})
	default:
		return errInvalidType
	}
}

type cgoFormatter interface {
	Format(deviceType cryptsetup.DeviceType, genericParams cryptsetup.GenericParams) error
}

type cgoLoader interface {
	Load(deviceType cryptsetup.DeviceType) error
}
