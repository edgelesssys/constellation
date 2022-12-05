package helm

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tj/assert"
)

func TestParseCRDs(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	data, err := ioutil.ReadFile("charts/edgeless/operators/charts/constellation-operator/crds/nodeimage-crd.yaml")
	require.NoError(err)

	crd, err := parseCRD(data)
	assert.NoError(err)
	fmt.Println(crd)
}
