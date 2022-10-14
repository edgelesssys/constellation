/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package aws

import (
	"fmt"

	"github.com/edgelesssys/constellation/v2/internal/attestation/vtpm"
	"github.com/edgelesssys/constellation/v2/internal/oid"
)

type Validator struct {
	oid.AWS
	*vtpm.Validator
}

func NewValidator(pcrs map[uint32][]byte, enforcedPCRs []uint32, log vtpm.WarnLogger) *Validator {
	return &Validator{
		Validator: vtpm.NewValidator(
			pcrs,
			enforcedPCRs,
			nil,
			nil,
			nil,
			log,
		),
	}
}

func (a *Validator) Validate(attDoc []byte, nonce []byte) ([]byte, error) {
	panic("aws validator not implemented")
}

func getTrustedKey() {
	fmt.Println("you have to trust aws on this")
}

type awsInstanceInfo struct {
	Region     string `json:"region"`
	AccountId  string `json:"accountId"`
	InstanceId string `json:"instanceId"`
}
