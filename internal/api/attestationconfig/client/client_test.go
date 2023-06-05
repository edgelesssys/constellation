/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/
package client

import (
	"testing"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/api/attestationconfig"
	"github.com/stretchr/testify/assert"
)

func TestUploadAzureSEVSNP(t *testing.T) {
	sut := Client{
		bucketID: "bucket",
		signer:   fakeSigner{},
	}
	version := attestationconfig.AzureSEVSNPVersion{}
	date := time.Date(2023, 1, 1, 1, 1, 1, 1, time.UTC)
	ops, err := sut.uploadAzureSEVSNP(version, []string{"2021-01-01-01-01.json", "2019-01-01-01-01.json"}, date)
	assert := assert.New(t)
	assert.NoError(err)
	dateStr := "2023-01-01-01-01.json"
	assert.Contains(ops, putCmd{
		apiObject: attestationconfig.AzureSEVSNPVersionAPI{
			Version:            dateStr,
			AzureSEVSNPVersion: version,
		},
	})
	assert.Contains(ops, putCmd{
		apiObject: attestationconfig.AzureSEVSNPVersionSignature{
			Version:   dateStr,
			Signature: []byte("signature"),
		},
	})
	assert.Contains(ops, putCmd{
		apiObject: attestationconfig.AzureSEVSNPVersionList([]string{"2023-01-01-01-01.json", "2021-01-01-01-01.json", "2019-01-01-01-01.json"}),
	})
}

func TestDeleteAzureSEVSNPVersions(t *testing.T) {
	sut := Client{
		bucketID: "bucket",
	}
	versions := attestationconfig.AzureSEVSNPVersionList([]string{"2023-01-01.json", "2021-01-01.json", "2019-01-01.json"})

	ops, err := sut.deleteAzureSEVSNPVersion(versions, "2021-01-01")

	assert := assert.New(t)
	assert.NoError(err)
	assert.Contains(ops, deleteCmd{
		apiObject: attestationconfig.AzureSEVSNPVersionAPI{
			Version: "2021-01-01.json",
		},
	})
	assert.Contains(ops, deleteCmd{
		apiObject: attestationconfig.AzureSEVSNPVersionSignature{
			Version: "2021-01-01.json",
		},
	})

	assert.Contains(ops, putCmd{
		apiObject: attestationconfig.AzureSEVSNPVersionList([]string{"2023-01-01.json", "2019-01-01.json"}),
	})
}

type fakeSigner struct{}

func (fakeSigner) Sign(_ []byte) ([]byte, error) {
	return []byte("signature"), nil
}
