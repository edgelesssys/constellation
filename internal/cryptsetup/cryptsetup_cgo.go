//go:build linux && cgo

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package cryptsetup contains CGO bindings for cryptsetup.
package cryptsetup

// #include <libcryptsetup.h>
import "C"

const (
	// ReadWriteQueueBypass is a flag to disable the write and read workqueues for a crypt device.
	ReadWriteQueueBypass = C.CRYPT_ACTIVATE_NO_WRITE_WORKQUEUE | C.CRYPT_ACTIVATE_NO_READ_WORKQUEUE
)
