package resources

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	k8s "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestMarshalK8SResources(t *testing.T) {
	testCases := map[string]struct {
		resources    interface{}
		expectErr    bool
		expectedYAML string
	}{
		"ConfigMap as only field can be marshaled": {
			resources: &struct {
				ConfigMap k8s.ConfigMap
			}{
				ConfigMap: k8s.ConfigMap{
					TypeMeta: v1.TypeMeta{
						Kind:       "ConfigMap",
						APIVersion: "v1",
					},
					Data: map[string]string{
						"key": "value",
					},
				},
			},
			expectedYAML: `apiVersion: v1
data:
  key: value
kind: ConfigMap
metadata:
  creationTimestamp: null
`,
		},
		"Multiple fields are correctly encoded": {
			resources: &struct {
				ConfigMap k8s.ConfigMap
				Secret    k8s.Secret
			}{
				ConfigMap: k8s.ConfigMap{
					TypeMeta: v1.TypeMeta{
						Kind:       "ConfigMap",
						APIVersion: "v1",
					},
					Data: map[string]string{
						"key": "value",
					},
				},
				Secret: k8s.Secret{
					TypeMeta: v1.TypeMeta{
						Kind:       "Secret",
						APIVersion: "v1",
					},
					Data: map[string][]byte{
						"key": []byte("value"),
					},
				},
			},
			expectedYAML: `apiVersion: v1
data:
  key: value
kind: ConfigMap
metadata:
  creationTimestamp: null
---
apiVersion: v1
data:
  key: dmFsdWU=
kind: Secret
metadata:
  creationTimestamp: null
`,
		},
		"Non-pointer is detected": {
			resources: "non-pointer",
			expectErr: true,
		},
		"Nil resource pointer is detected": {
			resources: nil,
			expectErr: true,
		},
		"Non-pointer field is ignored": {
			resources: &struct{ String string }{String: "somestring"},
		},
		"nil field is ignored": {
			resources: &struct {
				ConfigMap *k8s.ConfigMap
			}{
				ConfigMap: nil,
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			yaml, err := MarshalK8SResources(tc.resources)

			if tc.expectErr {
				assert.Error(err)
				return
			}
			require.NoError(err)

			assert.Equal(tc.expectedYAML, string(yaml))
		})
	}
}

func TestUnmarshalK8SResources(t *testing.T) {
	testCases := map[string]struct {
		data        string
		into        interface{}
		expectedObj interface{}
		expectErr   bool
	}{
		"ConfigMap as only field can be unmarshaled": {
			data: `apiVersion: v1
data:
  key: value
kind: ConfigMap
metadata:
  creationTimestamp: null
`,
			into: &struct {
				ConfigMap k8s.ConfigMap
			}{},
			expectedObj: &struct {
				ConfigMap k8s.ConfigMap
			}{
				ConfigMap: k8s.ConfigMap{
					TypeMeta: v1.TypeMeta{
						Kind:       "ConfigMap",
						APIVersion: "v1",
					},
					Data: map[string]string{
						"key": "value",
					},
				},
			},
		},
		"Multiple fields are correctly unmarshaled": {
			data: `apiVersion: v1
data:
  key: value
kind: ConfigMap
metadata:
  creationTimestamp: null
---
apiVersion: v1
data:
  key: dmFsdWU=
kind: Secret
metadata:
  creationTimestamp: null
`,
			into: &struct {
				ConfigMap k8s.ConfigMap
				Secret    k8s.Secret
			}{},
			expectedObj: &struct {
				ConfigMap k8s.ConfigMap
				Secret    k8s.Secret
			}{
				ConfigMap: k8s.ConfigMap{
					TypeMeta: v1.TypeMeta{
						Kind:       "ConfigMap",
						APIVersion: "v1",
					},
					Data: map[string]string{
						"key": "value",
					},
				},
				Secret: k8s.Secret{
					TypeMeta: v1.TypeMeta{
						Kind:       "Secret",
						APIVersion: "v1",
					},
					Data: map[string][]byte{
						"key": []byte("value"),
					},
				},
			},
		},
		"Mismatching amount of fields is detected": {
			data: `apiVersion: v1
data:
  key: value
kind: ConfigMap
metadata:
  creationTimestamp: null
---
apiVersion: v1
data:
  key: dmFsdWU=
kind: Secret
metadata:
  creationTimestamp: null
`,
			into: &struct {
				ConfigMap k8s.ConfigMap
			}{},
			expectErr: true,
		},
		"Non-struct pointer is detected": {
			into:      proto.String("test"),
			expectErr: true,
		},
		"Nil into is detected": {
			into:      nil,
			expectErr: true,
		},
		"Invalid yaml is detected": {
			data: `duplicateKey: value
		duplicateKey: value`,
			into: &struct {
				ConfigMap k8s.ConfigMap
			}{},
			expectErr: true,
		},
		"Struct field cannot interface with runtime.Object": {
			data: `apiVersion: v1
data:
  key: value
kind: ConfigMap
metadata:
  creationTimestamp: null
`,
			into: &struct {
				String string
			}{},
			expectErr: true,
		},
		"Struct field mismatch": {
			data: `apiVersion: v1
data:
  key: value
kind: ConfigMap
metadata:
  creationTimestamp: null
`,
			into: &struct {
				Secret k8s.Secret
			}{},
			expectErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			err := UnmarshalK8SResources([]byte(tc.data), tc.into)

			if tc.expectErr {
				assert.Error(err)
				return
			}
			require.NoError(err)

			assert.Equal(tc.expectedObj, tc.into)
		})
	}
}

func TestMarshalK8SResourcesList(t *testing.T) {
	testCases := map[string]struct {
		resources    []runtime.Object
		expectErr    bool
		expectedYAML string
	}{
		"ConfigMap as only element be marshaled": {
			resources: []runtime.Object{
				&k8s.ConfigMap{
					TypeMeta: v1.TypeMeta{
						Kind:       "ConfigMap",
						APIVersion: "v1",
					},
					Data: map[string]string{
						"key": "value",
					},
				},
			},
			expectedYAML: `apiVersion: v1
data:
  key: value
kind: ConfigMap
metadata:
  creationTimestamp: null
`,
		},
		"Multiple fields are correctly encoded": {
			resources: []runtime.Object{
				&k8s.ConfigMap{
					TypeMeta: v1.TypeMeta{
						Kind:       "ConfigMap",
						APIVersion: "v1",
					},
					Data: map[string]string{
						"key": "value",
					},
				},
				&k8s.Secret{
					TypeMeta: v1.TypeMeta{
						Kind:       "Secret",
						APIVersion: "v1",
					},
					Data: map[string][]byte{
						"key": []byte("value"),
					},
				},
			},
			expectedYAML: `apiVersion: v1
data:
  key: value
kind: ConfigMap
metadata:
  creationTimestamp: null
---
apiVersion: v1
data:
  key: dmFsdWU=
kind: Secret
metadata:
  creationTimestamp: null
`,
		},
		"Nil resource pointer is encodes": {
			resources:    []runtime.Object{nil},
			expectedYAML: "null\n",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			yaml, err := MarshalK8SResourcesList(tc.resources)

			if tc.expectErr {
				assert.Error(err)
				return
			}
			require.NoError(err)

			assert.Equal(tc.expectedYAML, string(yaml))
		})
	}
}
