package aws

import "github.com/edgelesssys/constellation/coordinator/oid"

type Issuer struct {
	oid.AWS
}

func NewIssuer() *Issuer {
	return &Issuer{}
}

func (i *Issuer) Issue(userData []byte, nonce []byte) ([]byte, error) {
	return NsmGetAttestationDoc(userData, nonce)
}
