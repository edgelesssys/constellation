/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package azure

import (
	"context"
	"errors"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/keyvault/azkeys"
	"github.com/edgelesssys/constellation/v2/keyservice/internal/storage"
	"github.com/edgelesssys/constellation/v2/keyservice/kms"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubHSMClient struct {
	keyCreated      bool
	createOCTKeyErr error
	importKeyErr    error
	getKeyErr       error
	keyID           string
	unwrapKeyErr    error
	unwrapKeyResult []byte
	wrapKeyErr      error
}

func (s *stubHSMClient) CreateKey(ctx context.Context, name string, parameters azkeys.CreateKeyParameters, options *azkeys.CreateKeyOptions) (azkeys.CreateKeyResponse, error) {
	s.keyCreated = true
	return azkeys.CreateKeyResponse{}, s.createOCTKeyErr
}

func (s *stubHSMClient) ImportKey(ctx context.Context, name string, parameters azkeys.ImportKeyParameters, options *azkeys.ImportKeyOptions) (azkeys.ImportKeyResponse, error) {
	s.keyCreated = true
	return azkeys.ImportKeyResponse{}, s.importKeyErr
}

func (s *stubHSMClient) GetKey(ctx context.Context, name string, version string, options *azkeys.GetKeyOptions) (azkeys.GetKeyResponse, error) {
	return azkeys.GetKeyResponse{
		KeyBundle: azkeys.KeyBundle{
			Key: &azkeys.JSONWebKey{
				KID: to.Ptr(azkeys.ID(s.keyID)),
			},
		},
	}, s.getKeyErr
}

func (s *stubHSMClient) UnwrapKey(ctx context.Context, name string, version string, parameters azkeys.KeyOperationsParameters, options *azkeys.UnwrapKeyOptions) (azkeys.UnwrapKeyResponse, error) {
	return azkeys.UnwrapKeyResponse{
		KeyOperationResult: azkeys.KeyOperationResult{
			Result: s.unwrapKeyResult,
		},
	}, s.unwrapKeyErr
}

func (s *stubHSMClient) WrapKey(ctx context.Context, name string, version string, parameters azkeys.KeyOperationsParameters, options *azkeys.WrapKeyOptions) (azkeys.WrapKeyResponse, error) {
	return azkeys.WrapKeyResponse{}, s.wrapKeyErr
}

type stubStorage struct {
	key    []byte
	getErr error
	putErr error
}

func (s *stubStorage) Get(context.Context, string) ([]byte, error) {
	return s.key, s.getErr
}

func (s *stubStorage) Put(context.Context, string, []byte) error {
	return s.putErr
}

func TestHSMCreateKEK(t *testing.T) {
	someErr := errors.New("error")
	importKey := []byte("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")

	testCases := map[string]struct {
		client    *stubHSMClient
		importKey []byte
		wantErr   bool
	}{
		"create new kek successful": {
			client: &stubHSMClient{},
		},
		"CreateOCTKey fails": {
			client:  &stubHSMClient{createOCTKeyErr: someErr},
			wantErr: true,
		},
		"import key successful": {
			client:    &stubHSMClient{},
			importKey: importKey,
		},
		"ImportKey fails": {
			client:    &stubHSMClient{importKeyErr: someErr},
			importKey: importKey,
			wantErr:   true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			client := HSMClient{
				client:  tc.client,
				storage: storage.NewMemMapStorage(),
			}

			err := client.CreateKEK(context.Background(), "test-key", tc.importKey)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.True(tc.client.keyCreated)
			}
		})
	}
}

func TestHSMGetNewDEK(t *testing.T) {
	someErr := errors.New("error")
	keyID := "https://test.managedhsm.azure.net/keys/test-key/test-key-version"

	testCases := map[string]struct {
		client  hsmClientAPI
		storage kms.Storage
		wantErr bool
	}{
		"successful": {
			client:  &stubHSMClient{keyID: keyID},
			storage: storage.NewMemMapStorage(),
		},
		"Get from storage fails": {
			client:  &stubHSMClient{keyID: keyID},
			storage: &stubStorage{getErr: someErr},
			wantErr: true,
		},
		"Put to storage fails": {
			client: &stubHSMClient{keyID: keyID},
			storage: &stubStorage{
				getErr: storage.ErrDEKUnset,
				putErr: someErr,
			},
			wantErr: true,
		},
		"WrapKey fails": {
			client:  &stubHSMClient{keyID: keyID, wrapKeyErr: someErr},
			storage: storage.NewMemMapStorage(),
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			client := HSMClient{
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

func TestHSMGetExistingDEK(t *testing.T) {
	someErr := errors.New("error")
	keyVersion := "https://test.managedhsm.azure.net/keys/test-key/test-key-version"
	testKey := []byte("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")

	testCases := map[string]struct {
		client  hsmClientAPI
		wantErr bool
	}{
		"successful": {
			client: &stubHSMClient{keyID: keyVersion, unwrapKeyResult: testKey},
		},
		"UnwrapKey fails": {
			client:  &stubHSMClient{keyID: keyVersion, unwrapKeyErr: someErr},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			keyID := "volume-01"
			storage := storage.NewMemMapStorage()
			require.NoError(storage.Put(context.Background(), keyID, testKey))

			client := HSMClient{
				client:  tc.client,
				storage: storage,
			}

			dek, err := client.GetDEK(context.Background(), "test-key", keyID, len(testKey))
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.Len(dek, len(testKey))
				assert.NoError(err)
			}
		})
	}
}
