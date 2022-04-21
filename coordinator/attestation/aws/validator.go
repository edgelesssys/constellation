package aws

import (
	"github.com/edgelesssys/constellation/coordinator/oid"
)

type Validator struct {
	oid.AWS
}

func (a *Validator) Validate(attDoc []byte, nonce []byte) ([]byte, error) {
	panic("aws validator not implemented")
}
