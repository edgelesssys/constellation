/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import "github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"

type helmLoader interface {
	Load(csp cloudprovider.Provider, conformanceMode bool, masterSecret []byte, salt []byte, enforcedPCRs []uint32, enforceIDKeyDigest bool) ([]byte, error)
}

type stubHelmLoader struct {
	loadErr error
}

func (d *stubHelmLoader) Load(csp cloudprovider.Provider, conformanceMode bool, masterSecret []byte, salt []byte, enforcedPCRs []uint32, enforceIDKeyDigest bool) ([]byte, error) {
	return nil, d.loadErr
}
