/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: BUSL-1.1
*/

package fixtures

import _ "embed"

// UKI returns the UKI EFI binary.
func UKI() []byte {
	return ukiEFI[:]
}

//go:embed uki.efi
var ukiEFI []byte
