/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/
package client

import (
	"encoding/json"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/api/attestationconfig"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TODO replace with unit testable version
//func TestUploadAzureSEVSNPVersions(t *testing.T) {
//	ctx := context.Background()
//	client, clientClose, err := New(ctx, cfg, []byte(*cosignPwd), privateKey)
//	require.NoError(t, err)
//	defer func() { _ = clientClose(ctx) }()
//	d := time.Date(2021, 1, 1, 1, 1, 1, 1, time.UTC)
//	require.NoError(t, client.UploadAzureSEVSNP(ctx, versionValues, d))
//}

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
