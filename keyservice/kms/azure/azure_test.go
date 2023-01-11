/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package azure

import (
	"context"
	"encoding/base64"
	"errors"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/keyvault/azsecrets"
	"github.com/edgelesssys/constellation/v2/keyservice/internal/storage"
	"github.com/edgelesssys/constellation/v2/keyservice/kms"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m,
		// https://github.com/census-instrumentation/opencensus-go/issues/1262
		goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start"),
	)
}

type stubAzureClient struct {
	setSecretCalled bool
	setSecretErr    error
	getSecretErr    error
	secret          []byte
}

func (s *stubAzureClient) SetSecret(ctx context.Context, secretName string, parameters azsecrets.SetSecretParameters, options *azsecrets.SetSecretOptions) (azsecrets.SetSecretResponse, error) {
	s.setSecretCalled = true
	return azsecrets.SetSecretResponse{}, s.setSecretErr
}

func (s *stubAzureClient) GetSecret(ctx context.Context, secretName string, version string, options *azsecrets.GetSecretOptions) (azsecrets.GetSecretResponse, error) {
	return azsecrets.GetSecretResponse{
		SecretBundle: azsecrets.SecretBundle{
			Value: to.Ptr(base64.StdEncoding.EncodeToString(s.secret)),
		},
	}, s.getSecretErr
}

func TestKMSCreateKEK(t *testing.T) {
	someErr := errors.New("error")
	importKey := []byte("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")

	testCases := map[string]struct {
		client    *stubAzureClient
		importKey []byte
		wantErr   bool
	}{
		"create new kek successful": {
			client: &stubAzureClient{},
		},
		"import kek successful": {
			client:    &stubAzureClient{},
			importKey: importKey,
		},
		"SetSecret fails on new": {
			client:  &stubAzureClient{setSecretErr: someErr},
			wantErr: true,
		},
		"SetSecret fails on import": {
			client:    &stubAzureClient{setSecretErr: someErr},
			importKey: importKey,
			wantErr:   true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			client := &KMSClient{
				client: tc.client,
			}

			err := client.CreateKEK(context.Background(), "test-key", tc.importKey)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.True(tc.client.setSecretCalled)
			}
		})
	}
}

func TestKMSGetDEK(t *testing.T) {
	someErr := errors.New("error")
	wrapKey := []byte("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")

	testCases := map[string]struct {
		client  kmsClientAPI
		storage kms.Storage
		wantErr bool
	}{
		"successful for new key": {
			client:  &stubAzureClient{secret: wrapKey},
			storage: storage.NewMemMapStorage(),
		},
		"successful for existing key": {
			// test keys taken from `kms/util/crypto_test.go`
			client:  &stubAzureClient{secret: []byte{0xD6, 0x8A, 0xED, 0xF5, 0xDB, 0x89, 0x95, 0x66, 0xA9, 0xFF, 0xD9, 0x31, 0x27, 0x4E, 0x30, 0x2D, 0x21, 0xA9, 0x46, 0x21, 0x16, 0x6C, 0x16, 0x17, 0xD1, 0x96, 0x5D, 0xB2, 0xE9, 0x0E, 0x96, 0xD1}},
			storage: &stubStorage{key: []byte{0x14, 0x48, 0xC4, 0xEA, 0x4B, 0x4B, 0xCA, 0xE4, 0x5A, 0xD4, 0xCC, 0xE3, 0xF7, 0xDD, 0xD5, 0x78, 0xA5, 0xA9, 0xEF, 0x9A, 0x93, 0x36, 0x09, 0xD6, 0x23, 0x01, 0xF5, 0x5F, 0xE1, 0x20, 0xDD, 0xFC, 0xBC, 0xF3, 0xA9, 0x67, 0x8B, 0x89, 0x54, 0x96}},
		},
		"Get from storage fails": {
			client:  &stubAzureClient{},
			storage: &stubStorage{getErr: someErr},
			wantErr: true,
		},
		"Put to storage fails": {
			client: &stubAzureClient{secret: wrapKey},
			storage: &stubStorage{
				getErr: storage.ErrDEKUnset,
				putErr: someErr,
			},
			wantErr: true,
		},
		"GetSecret fails": {
			client:  &stubAzureClient{getSecretErr: someErr},
			storage: storage.NewMemMapStorage(),
			wantErr: true,
		},
		"GetSecret fails with unknown kek": {
			client:  &stubAzureClient{getSecretErr: errors.New("SecretNotFound")},
			storage: storage.NewMemMapStorage(),
			wantErr: true,
		},
		"key wrapping fails": {
			client:  &stubAzureClient{secret: []byte{0x1}},
			storage: storage.NewMemMapStorage(),
			wantErr: true,
		},
		"key unwrapping fails": {
			client:  &stubAzureClient{secret: wrapKey},
			storage: &stubStorage{key: []byte{0x1}},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			client := KMSClient{
				client:  tc.client,
				storage: tc.storage,
			}

			dek, err := client.GetDEK(context.Background(), "test-key", "volume-01", 32)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.Len(dek, 32)
				assert.NoError(err)
			}
		})
	}
}
