/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package helm

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/deploy/helm"
	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/chart/loader"
)

func TestLoad(t *testing.T) {
	assert := assert.New(t)

	chartLoader := ChartLoader{}
	release, err := chartLoader.Load(cloudprovider.GCP, true, []byte("secret"), []byte("salt"), nil, false)
	assert.NoError(err)

	var helmReleases helm.Releases
	err = json.Unmarshal(release, &helmReleases)
	assert.NoError(err)
	reader := bytes.NewReader(helmReleases.ConstellationServices.Chart)
	chart, err := loader.LoadArchive(reader)
	assert.NoError(err)
	assert.NotNil(chart.Dependencies())
}
