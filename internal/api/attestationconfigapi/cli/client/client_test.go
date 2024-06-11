/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/
package client

import (
	"testing"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/api/attestationconfigapi"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/stretchr/testify/assert"
)

func TestUploadAzureSEVSNP(t *testing.T) {
	sut := Client{
		bucketID: "bucket",
		signer:   fakeSigner{},
	}
	version := attestationconfigapi.SEVSNPVersion{}
	date := time.Date(2023, 1, 1, 1, 1, 1, 1, time.UTC)
	ops := sut.constructUploadCmd(variant.AzureSEVSNP{}, version, attestationconfigapi.SEVSNPVersionList{List: []string{"2021-01-01-01-01.json", "2019-01-01-01-01.json"}, Variant: variant.AzureSEVSNP{}}, date)
	dateStr := "2023-01-01-01-01.json"
	assert := assert.New(t)
	assert.Contains(ops, putCmd{
		apiObject: attestationconfigapi.SEVSNPVersionAPI{
			Variant:       variant.AzureSEVSNP{},
			Version:       dateStr,
			SEVSNPVersion: version,
		},
		signer: fakeSigner{},
	})
	assert.Contains(ops, putCmd{
		apiObject: attestationconfigapi.SEVSNPVersionList{Variant: variant.AzureSEVSNP{}, List: []string{"2023-01-01-01-01.json", "2021-01-01-01-01.json", "2019-01-01-01-01.json"}},
		signer:    fakeSigner{},
	})
}

func TestDeleteAzureSEVSNPVersions(t *testing.T) {
	sut := Client{
		bucketID: "bucket",
	}
	versions := attestationconfigapi.SEVSNPVersionList{List: []string{"2023-01-01.json", "2021-01-01.json", "2019-01-01.json"}}

	ops, err := sut.deleteSEVSNPVersion(versions, "2021-01-01")

	assert := assert.New(t)
	assert.NoError(err)
	assert.Contains(ops, deleteCmd{
		apiObject: attestationconfigapi.SEVSNPVersionAPI{
			Version: "2021-01-01.json",
		},
	})

	assert.Contains(ops, putCmd{
		apiObject: attestationconfigapi.SEVSNPVersionList{List: []string{"2023-01-01.json", "2019-01-01.json"}},
	})
}

type fakeSigner struct{}

func (fakeSigner) Sign(_ []byte) ([]byte, error) {
	return []byte("signature"), nil
}
