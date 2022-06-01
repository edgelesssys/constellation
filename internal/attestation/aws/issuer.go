package aws

import "github.com/edgelesssys/constellation/internal/oid"

type Issuer struct {
	oid.AWS
}

func (i *Issuer) Issue(userData []byte, nonce []byte) ([]byte, error) {
	panic("aws issuer not implemented")
}
