/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package aws

import "github.com/edgelesssys/constellation/internal/oid"

type Issuer struct {
	oid.AWS
}

func (i *Issuer) Issue(userData []byte, nonce []byte) ([]byte, error) {
	panic("aws issuer not implemented")
}
