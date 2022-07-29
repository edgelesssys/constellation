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
			uri:     fmt.Sprintf("%s?salt=%s", ClusterKMSURI, base64.URLEncoding.EncodeToString([]byte("salt"))),
			wantErr: false,
		},
		"aws kms": {
			uri:     fmt.Sprintf(AWSKMSURI, ""),
			wantErr: true,
		},
		"azure kms": {
			uri:     fmt.Sprintf(AzureKMSURI, "", ""),
			wantErr: true,
		},
		"azure hsm": {
			uri:     fmt.Sprintf(AzureHSMURI, ""),
			wantErr: true,
		},
		"gcp kms": {
			uri:     fmt.Sprintf(GCPKMSURI, "", "", "", ""),
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

	kms, err := SetUpKMS(context.Background(), "storage://unknown", "kms://unknown")
	assert.Error(err)
	assert.Nil(kms)

	kms, err = SetUpKMS(context.Background(), "storage://no-store", "kms://cluster-kms?salt="+base64.URLEncoding.EncodeToString([]byte("salt")))
	assert.NoError(err)
	assert.NotNil(kms)
}

func TestGetAWSKMSConfig(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	policy := "{keyPolicy: keyPolicy}"
	escapedPolicy := url.QueryEscape(policy)
	uri, err := url.Parse(fmt.Sprintf(AWSKMSURI, escapedPolicy))
	require.NoError(err)
	policyProducer, err := getAWSKMSConfig(uri)
	require.NoError(err)
	keyPolicy, err := policyProducer.CreateKeyPolicy("")
	require.NoError(err)
	assert.Equal(policy, keyPolicy)
}

func TestGetAzureBlobConfig(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	connStr := "DefaultEndpointsProtocol=https;AccountName=test;AccountKey=Q29uc3RlbGxhdGlvbg==;EndpointSuffix=core.windows.net"
	escapedConnStr := url.QueryEscape(connStr)
	container := "test"
	uri, err := url.Parse(fmt.Sprintf(AzureBlobURI, container, escapedConnStr))
	require.NoError(err)
	rContainer, rConnStr, err := getAzureBlobConfig(uri)
	require.NoError(err)
	assert.Equal(container, rContainer)
	assert.Equal(connStr, rConnStr)
}

func TestGetGCPKMSConfig(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	project := "test-project"
	location := "global"
	keyRing := "test-ring"
	protectionLvl := "2"
	uri, err := url.Parse(fmt.Sprintf(GCPKMSURI, project, location, keyRing, protectionLvl))
	require.NoError(err)
	rProject, rLocation, rKeyRing, rProtectionLvl, err := getGCPKMSConfig(uri)
	require.NoError(err)
	assert.Equal(project, rProject)
	assert.Equal(location, rLocation)
	assert.Equal(keyRing, rKeyRing)
	assert.Equal(2, rProtectionLvl)

	uri, err = url.Parse(fmt.Sprintf(GCPKMSURI, project, location, keyRing, "invalid"))
	require.NoError(err)
	_, _, _, _, err = getGCPKMSConfig(uri)
	assert.Error(err)
}

func TestGetClusterKMSConfig(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	expectedSalt := []byte{
		0x0, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf,
		0x0, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf,
	}

	uri, err := url.Parse(ClusterKMSURI + "?salt=" + base64.URLEncoding.EncodeToString(expectedSalt))
	require.NoError(err)

	salt, err := getClusterKMSConfig(uri)
	assert.NoError(err)
	assert.Equal(expectedSalt, salt)
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
