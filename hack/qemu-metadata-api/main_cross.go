//go:build !cgo

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package main

func main() {
	panic("CGO disabled but started qemu-metadata-api")
}
