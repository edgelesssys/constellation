//go:build !linux || !cgo

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: BUSL-1.1
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

func headerRestore(_ cryptDevice, _ string) error {
	return errCGONotSupported
}

func headerBackup(_ cryptDevice, _ string) error {
	return errCGONotSupported
}

func initByDevicePath(_ string) (cryptDevice, cryptDevice, string, string, error) {
	return nil, nil, "", "", errCGONotSupported
}

func initByName(_ string) (cryptDevice, cryptDevice, string, string, error) {
	return nil, nil, "", "", errCGONotSupported
}

func loadLUKS2(_ cryptDevice) error {
	return errCGONotSupported
}

func detachLoopbackDevice(_ string) error {
	return errCGONotSupported
}
