package helm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseCRDs(t *testing.T) {
	testCases := map[string]struct {
		data    string
		wantErr bool
	}{
		"success": {
			data:    "apiVersion: apiextensions.k8s.io/v1\nkind: CustomResourceDefinition\nmetadata:\n  name: nodeimages.update.edgeless.systems\nspec:\n  group: update.edgeless.systems\n  names:\n    kind: NodeImage\n",
			wantErr: false,
		},
		"wrong kind": {
			data:    "apiVersion: v1\nkind: Secret\ntype: Opaque\nmetadata:\n  name: supersecret\n  namespace: testNamespace\ndata:\n  data: YWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWE=\n",
			wantErr: true,
		},
		"decoding error": {
			data:    "asdf",
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			_, err := parseCRD([]byte(tc.data))
			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
		})
	}
}
