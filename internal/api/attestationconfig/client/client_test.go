/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/
package client

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/api/attestationconfig"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	assert.Contains(ops, updateCmd{
		apiObject: attestationconfig.AzureSEVSNPVersionAPI{
			Version:            dateStr,
			AzureSEVSNPVersion: version,
		},
	})
	assert.Contains(ops, updateCmd{
		apiObject: attestationconfig.AzureSEVSNPVersionSignature{
			Version:   dateStr,
			Signature: []byte("signature"),
		},
	})
	assert.Contains(ops, updateCmd{
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
		path: "constellation/v1/attestation/azure-sev-snp/2021-01-01.json",
	})
	assert.Contains(ops, deleteCmd{
		path: "constellation/v1/attestation/azure-sev-snp/2021-01-01.json.sig",
	})

	removedVersions := attestationconfig.AzureSEVSNPVersionList([]string{"2023-01-01.json", "2019-01-01.json"})
	dt, err := json.Marshal(removedVersions)
	require.NoError(t, err)
	assert.Contains(ops, putCmd{
		path: "constellation/v1/attestation/azure-sev-snp/list",
		data: dt,
	})
}

type fakeSigner struct{}

func (fakeSigner) Sign(_ []byte) ([]byte, error) {
	return []byte("signature"), nil
}
