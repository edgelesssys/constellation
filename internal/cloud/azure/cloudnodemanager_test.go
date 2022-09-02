/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package azure

import (
	"testing"

	"github.com/edgelesssys/constellation/internal/versions"
	"github.com/stretchr/testify/assert"
)

func TestTrivialCNMFunctions(t *testing.T) {
	assert := assert.New(t)
	cloud := CloudNodeManager{}

	assert.NotEmpty(cloud.Image(versions.Default))
	assert.NotEmpty(cloud.Path())
	assert.NotEmpty(cloud.ExtraArgs())
	assert.True(cloud.Supported())
}
