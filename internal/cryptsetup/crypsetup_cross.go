//go:build !linux || !cgo

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/
package cryptsetup

const (
	// ReadWriteQueueBypass is a flag to disable the write and read workqueues for a crypt device.
	ReadWriteQueueBypass          = cryptActivateNoReadWorkqueue | cryptActivateNoWriteWorkqueue
	cryptActivateNoReadWorkqueue  = 0x1000000
	cryptActivateNoWriteWorkqueue = 0x2000000
)
