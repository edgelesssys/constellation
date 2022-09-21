/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package aws

import (
	"github.com/edgelesssys/constellation/v2/internal/oid"
)

type Validator struct {
	oid.AWS
}

func (a *Validator) Validate(attDoc []byte, nonce []byte) ([]byte, error) {
	panic("aws validator not implemented")
}
