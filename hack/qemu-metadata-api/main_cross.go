//go:build !cgo

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: BUSL-1.1
*/

package main

func main() {
	panic("CGO disabled but started qemu-metadata-api")
}
