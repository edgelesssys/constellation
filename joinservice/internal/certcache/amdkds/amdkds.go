/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// The AMDKDS package implements interaction with the AMD KDS (Key Distribution Service).
package amdkds

import (
	"crypto/x509"
	"fmt"

	"github.com/google/go-sev-guest/abi"
	"github.com/google/go-sev-guest/verify/trust"
)

// a KDSClient is a client for interacting with the AMD KDS.
type KDSClient struct {
	getter trust.HTTPSGetter
}

// NewKDSClient creates a new KDS Client.
func NewKDSClient(getter trust.HTTPSGetter) *KDSClient {
	return &KDSClient{
		getter: getter,
	}
}

// CertChain queries the AMD KDS for the certificate chain for given signing type (VCEK / VLEK).
func (c *KDSClient) CertChain(signingType abi.ReportSigner) (ask, ark *x509.Certificate, err error) {
	askark, err := trust.GetProductChain("Milan", signingType, c.getter)
	if err != nil {
		return nil, nil, fmt.Errorf("retrieving certificate chain: %v", err)
	}

	return askark.Ask, askark.Ark, nil
}
