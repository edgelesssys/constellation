/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/
package attestationconfigapi

import (
	"testing"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/stretchr/testify/assert"
)

func TestUploadAzureSEVSNP(t *testing.T) {
	sut := Client{
		bucketID: "bucket",
		signer:   fakeSigner{},
	}
	version := SEVSNPVersion{}
	date := time.Date(2023, 1, 1, 1, 1, 1, 1, time.UTC)
	ops := sut.constructUploadCmd(variant.AzureSEVSNP{}, version, SEVSNPVersionList{list: []string{"2021-01-01-01-01.json", "2019-01-01-01-01.json"}, variant: variant.AzureSEVSNP{}}, date)
	dateStr := "2023-01-01-01-01.json"
	assert := assert.New(t)
	assert.Contains(ops, putCmd{
		apiObject: SEVSNPVersionAPI{
			Variant:       variant.AzureSEVSNP{},
			Version:       dateStr,
			SEVSNPVersion: version,
		},
		signer: fakeSigner{},
	})
	assert.Contains(ops, putCmd{
		apiObject: SEVSNPVersionList{variant: variant.AzureSEVSNP{}, list: []string{"2023-01-01-01-01.json", "2021-01-01-01-01.json", "2019-01-01-01-01.json"}},
		signer:    fakeSigner{},
	})
}

func TestDeleteAzureSEVSNPVersions(t *testing.T) {
	sut := Client{
		bucketID: "bucket",
	}
	versions := SEVSNPVersionList{list: []string{"2023-01-01.json", "2021-01-01.json", "2019-01-01.json"}}

	ops, err := sut.deleteSEVSNPVersion(versions, "2021-01-01")

	assert := assert.New(t)
	assert.NoError(err)
	assert.Contains(ops, deleteCmd{
		apiObject: SEVSNPVersionAPI{
			Version: "2021-01-01.json",
		},
	})

	assert.Contains(ops, putCmd{
		apiObject: SEVSNPVersionList{list: []string{"2023-01-01.json", "2019-01-01.json"}},
	})
}

type fakeSigner struct{}

func (fakeSigner) Sign(_ []byte) ([]byte, error) {
	return []byte("signature"), nil
}
