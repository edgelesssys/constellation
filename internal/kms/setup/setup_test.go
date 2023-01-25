/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package setup

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m,
		// https://github.com/census-instrumentation/opencensus-go/issues/1262
		goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start"),
	)
}

func TestGetStore(t *testing.T) {
	testCases := map[string]struct {
		uri     string
		wantErr bool
	}{
		"no store": {
			uri:     NoStoreURI,
			wantErr: false,
		},
		"aws s3": {
			uri:     fmt.Sprintf(AWSS3URI, ""),
			wantErr: true,
		},
		"azure blob": {
			uri:     fmt.Sprintf(AzureBlobURI, "", ""),
			wantErr: true,
		},
		"gcp storage": {
			uri:     fmt.Sprintf(GCPStorageURI, "", ""),
			wantErr: true,
		},
		"unknown store": {
			uri:     "storage://unknown",
			wantErr: true,
		},
		"invalid scheme": {
			uri:     ClusterKMSURI,
			wantErr: true,
		},
		"not a url": {
			uri:     ":/123",
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			_, err := getStore(context.Background(), tc.uri)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

func TestGetKMS(t *testing.T) {
	testCases := map[string]struct {
		uri     string
		wantErr bool
	}{
		"cluster kms": {
			uri:     fmt.Sprintf(ClusterKMSURI, base64.URLEncoding.EncodeToString([]byte("key")), base64.URLEncoding.EncodeToString([]byte("salt"))),
			wantErr: false,
		},
		"aws kms": {
			uri:     fmt.Sprintf(AWSKMSURI, "", ""),
			wantErr: true,
		},
		"azure kms": {
			uri:     fmt.Sprintf(AzureKMSURI, "", "", "", "", "", ""),
			wantErr: true,
		},
		"gcp kms": {
			uri:     fmt.Sprintf(GCPKMSURI, "", "", "", "", ""),
			wantErr: true,
		},
		"unknown kms": {
			uri:     "kms://unknown",
			wantErr: true,
		},
		"invalid scheme": {
			uri:     NoStoreURI,
			wantErr: true,
		},
		"not a url": {
			uri:     ":/123",
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			kms, err := getKMS(context.Background(), tc.uri, nil)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.NotNil(kms)
			}
		})
	}
}

func TestSetUpKMS(t *testing.T) {
	assert := assert.New(t)

	kms, err := KMS(context.Background(), "storage://unknown", "kms://unknown")
	assert.Error(err)
	assert.Nil(kms)

	masterSecret := MasterSecret{Key: []byte("key"), Salt: []byte("salt")}
	kms, err = KMS(context.Background(), "storage://no-store", masterSecret.EncodeToURI())
	assert.NoError(err)
	assert.NotNil(kms)
}

func TestGetConfig(t *testing.T) {
	const testURI = "test://config?name=test-name&data=test-data&value=test-value"

	testCases := map[string]struct {
		uri     string
		keys    []string
		wantErr bool
	}{
		"success": {
			uri:     testURI,
			keys:    []string{"name", "data", "value"},
			wantErr: false,
		},
		"less keys than capture groups": {
			uri:     testURI,
			keys:    []string{"name", "data"},
			wantErr: false,
		},
		"invalid regex": {
			uri:     testURI,
			keys:    []string{"name", "data", "test-value"},
			wantErr: true,
		},
		"missing value": {
			uri:     "test://config?name=test-name&data=test-data&value",
			keys:    []string{"name", "data", "value"},
			wantErr: true,
		},
		"more keys than expected": {
			uri:     testURI,
			keys:    []string{"name", "data", "value", "anotherValue"},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			uri, err := url.Parse(tc.uri)
			require.NoError(err)

			res, err := getConfig(uri.Query(), tc.keys)
			if tc.wantErr {
				assert.Error(err)
				assert.Len(res, len(tc.keys))
			} else {
				assert.NoError(err)
				require.Len(res, len(tc.keys))
				for i := range tc.keys {
					assert.NotEmpty(res[i])
				}
			}
		})
	}
}
