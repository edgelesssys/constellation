package azure

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/keyvault/azkeys"
	"github.com/Azure/azure-sdk-for-go/sdk/keyvault/azkeys/crypto"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/edgelesssys/constellation/kms/internal/storage"
	"github.com/edgelesssys/constellation/kms/kms"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubHSMClient struct {
	keyCreated      bool
	createOCTKeyErr error
	importKeyErr    error
	getKeyErr       error
	keyVersion      string
}

func (s *stubHSMClient) CreateOCTKey(ctx context.Context, name string, options *azkeys.CreateOCTKeyOptions) (azkeys.CreateOCTKeyResponse, error) {
	s.keyCreated = true
	return azkeys.CreateOCTKeyResponse{}, s.createOCTKeyErr
}

func (s *stubHSMClient) ImportKey(ctx context.Context, keyName string, key azkeys.JSONWebKey, options *azkeys.ImportKeyOptions) (azkeys.ImportKeyResponse, error) {
	s.keyCreated = true
	return azkeys.ImportKeyResponse{}, s.importKeyErr
}

func (s *stubHSMClient) GetKey(ctx context.Context, keyName string, options *azkeys.GetKeyOptions) (azkeys.GetKeyResponse, error) {
	return azkeys.GetKeyResponse{
		KeyBundle: azkeys.KeyBundle{
			Key: &azkeys.JSONWebKey{
				ID: to.StringPtr(s.keyVersion),
			},
		},
	}, s.getKeyErr
}

type stubCryptoClient struct {
	createErr       error
	unwrapKeyErr    error
	unwrapKeyResult []byte
	wrapKeyErr      error
}

func newStubCryptoClientFactory(stub *stubCryptoClient) func(keyURL string, credential azcore.TokenCredential, options *crypto.ClientOptions) (cryptoClientAPI, error) {
	return func(keyURL string, credential azcore.TokenCredential, options *crypto.ClientOptions) (cryptoClientAPI, error) {
		return stub, stub.createErr
	}
}

func (s *stubCryptoClient) UnwrapKey(ctx context.Context, alg crypto.KeyWrapAlgorithm, encryptedKey []byte, options *crypto.UnwrapKeyOptions) (crypto.UnwrapKeyResponse, error) {
	return crypto.UnwrapKeyResponse{
		KeyOperationResult: crypto.KeyOperationResult{
			Result: s.unwrapKeyResult,
		},
	}, s.unwrapKeyErr
}

func (s *stubCryptoClient) WrapKey(ctx context.Context, alg crypto.KeyWrapAlgorithm, key []byte, options *crypto.WrapKeyOptions) (crypto.WrapKeyResponse, error) {
	return crypto.WrapKeyResponse{}, s.wrapKeyErr
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
	keyVersion := "https://test.managedhsm.azure.net/keys/test-key/test-key-version"

	testCases := map[string]struct {
		client       hsmClientAPI
		storage      kms.Storage
		cryptoClient *stubCryptoClient
		wantErr      bool
	}{
		"successful": {
			client:       &stubHSMClient{keyVersion: keyVersion},
			cryptoClient: &stubCryptoClient{},
			storage:      storage.NewMemMapStorage(),
		},
		"Get from storage fails": {
			client:       &stubHSMClient{keyVersion: keyVersion},
			cryptoClient: &stubCryptoClient{},
			storage:      &stubStorage{getErr: someErr},
			wantErr:      true,
		},
		"Put to storage fails": {
			client:       &stubHSMClient{keyVersion: keyVersion},
			cryptoClient: &stubCryptoClient{},
			storage: &stubStorage{
				getErr: storage.ErrDEKUnset,
				putErr: someErr,
			},
			wantErr: true,
		},
		"GetKey fails": {
			client:       &stubHSMClient{getKeyErr: someErr},
			cryptoClient: &stubCryptoClient{},
			storage:      storage.NewMemMapStorage(),
			wantErr:      true,
		},
		"WrapKey fails": {
			client:       &stubHSMClient{keyVersion: keyVersion},
			cryptoClient: &stubCryptoClient{wrapKeyErr: someErr},
			storage:      storage.NewMemMapStorage(),
			wantErr:      true,
		},
		"creating crypto client fails": {
			client:       &stubHSMClient{keyVersion: keyVersion},
			cryptoClient: &stubCryptoClient{createErr: someErr},
			storage:      storage.NewMemMapStorage(),
			wantErr:      true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			client := HSMClient{
				client:          tc.client,
				newCryptoClient: newStubCryptoClientFactory(tc.cryptoClient),
				storage:         tc.storage,
				opts:            &crypto.ClientOptions{},
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
		client       hsmClientAPI
		cryptoClient *stubCryptoClient
		wantErr      bool
	}{
		"successful": {
			client:       &stubHSMClient{keyVersion: keyVersion},
			cryptoClient: &stubCryptoClient{unwrapKeyResult: testKey},
		},
		"GetKey fails": {
			client: &stubHSMClient{
				keyVersion: keyVersion,
				getKeyErr:  someErr,
			},
			cryptoClient: &stubCryptoClient{},
			wantErr:      true,
		},
		"UnwrapKey fails": {
			client:       &stubHSMClient{keyVersion: keyVersion},
			cryptoClient: &stubCryptoClient{unwrapKeyErr: someErr},
			wantErr:      true,
		},
		"creating crypto client fails": {
			client:       &stubHSMClient{keyVersion: keyVersion},
			cryptoClient: &stubCryptoClient{createErr: someErr},
			wantErr:      true,
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
				client:          tc.client,
				newCryptoClient: newStubCryptoClientFactory(tc.cryptoClient),
				storage:         storage,
				opts:            &crypto.ClientOptions{},
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

func TestGetKeyVersion(t *testing.T) {
	testVersion := "test-key-version"
	testCases := map[string]struct {
		client  *stubHSMClient
		wantErr bool
	}{
		"valid key version": {
			client: &stubHSMClient{keyVersion: fmt.Sprintf("https://test.managedhsm.azure.net/keys/test-key/%s", testVersion)},
		},
		"GetKey fails": {
			client:  &stubHSMClient{getKeyErr: errors.New("error")},
			wantErr: true,
		},
		"key ID is not an URL": {
			client:  &stubHSMClient{keyVersion: string([]byte{0x0, 0x1, 0x2})},
			wantErr: true,
		},
		"invalid key ID URL": {
			client:  &stubHSMClient{keyVersion: "https://test.managedhsm.azure.net/keys/test-key/test-key-version/another-version/and-another-one"},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			client := HSMClient{client: tc.client}

			keyVersion, err := client.getKeyVersion(context.Background(), "test")
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(testVersion, keyVersion)
			}
		})
	}
}
