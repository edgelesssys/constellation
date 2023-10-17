/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/
package main

import (
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/verify"
	"github.com/stretchr/testify/assert"
)

func TestAllEqual(t *testing.T) {
	// Test case 1: One input arg
	assert.True(t, allEqual(verify.TCBVersion{Bootloader: 1, Microcode: 2, SNP: 3, TEE: 4}), "Expected allEqual to return true for one input arg, but got false")

	// Test case 2: Three input args that are equal
	assert.True(t, allEqual(
		verify.TCBVersion{Bootloader: 1, Microcode: 2, SNP: 3, TEE: 4},
		verify.TCBVersion{Bootloader: 1, Microcode: 2, SNP: 3, TEE: 4},
		verify.TCBVersion{Bootloader: 1, Microcode: 2, SNP: 3, TEE: 4},
	), "Expected allEqual to return true for three equal input args, but got false")

	// Test case 3: Three input args where second and third element are different
	assert.False(t, allEqual(
		verify.TCBVersion{Bootloader: 2, Microcode: 2, SNP: 3, TEE: 4},
		verify.TCBVersion{Bootloader: 2, Microcode: 2, SNP: 3, TEE: 4},
		verify.TCBVersion{Bootloader: 2, Microcode: 3, SNP: 3, TEE: 4},
	), "Expected allEqual to return false for three input args with different second and third elements, but got true")
}
