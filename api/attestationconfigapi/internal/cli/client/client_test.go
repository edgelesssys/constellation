/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/
package client

import (
	"testing"

	"github.com/edgelesssys/constellation/v2/api/attestationconfigapi"
	"github.com/stretchr/testify/assert"
)

func TestDeleteAzureSEVSNPVersions(t *testing.T) {
	sut := Client{
		bucketID: "bucket",
	}
	versions := attestationconfigapi.List{List: []string{"2023-01-01.json", "2021-01-01.json", "2019-01-01.json"}}

	ops, err := sut.deleteVersion(versions, "2021-01-01")

	assert := assert.New(t)
	assert.NoError(err)
	assert.Contains(ops, deleteCmd{
		apiObject: attestationconfigapi.Entry{
			Version: "2021-01-01.json",
		},
	})

	assert.Contains(ops, putCmd{
		apiObject: attestationconfigapi.List{List: []string{"2023-01-01.json", "2019-01-01.json"}},
	})
}
