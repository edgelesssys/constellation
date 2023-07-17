//go:build !linux || !cgo

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/
package cryptsetup

import (
	"errors"
)

const (
	// ReadWriteQueueBypass is a flag to disable the write and read workqueues for a crypt device.
	ReadWriteQueueBypass          = cryptActivateNoReadWorkqueue | cryptActivateNoWriteWorkqueue
	cryptActivateNoReadWorkqueue  = 0x1000000
	cryptActivateNoWriteWorkqueue = 0x2000000
	wipeFlags                     = 0x10 | 0x1000
	wipePattern                   = 0
)

var errCGONotSupported = errors.New("using cryptsetup requires building with CGO")

func format(_ cryptDevice, _ bool) error {
	return errCGONotSupported
}

func initByDevicePath(_ string) (cryptDevice, error) {
	return nil, errCGONotSupported
}

func initByName(_ string) (cryptDevice, error) {
	return nil, errCGONotSupported
}

func loadLUKS2(_ cryptDevice) error {
	return errCGONotSupported
}
