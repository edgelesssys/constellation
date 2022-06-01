package aws

import (
	"github.com/edgelesssys/constellation/internal/oid"
)

type Validator struct {
	oid.AWS
}

func (a *Validator) Validate(attDoc []byte, nonce []byte) ([]byte, error) {
	panic("aws validator not implemented")
}
