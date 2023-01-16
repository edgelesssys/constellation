/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package gcp

import (
	"context"
	"errors"
	"testing"

	"cloud.google.com/go/kms/apiv1/kmspb"
	kmsInterface "github.com/edgelesssys/constellation/v2/internal/kms/kms"
	"github.com/edgelesssys/constellation/v2/internal/kms/kms/util"
	"github.com/edgelesssys/constellation/v2/internal/kms/storage"
	"github.com/googleapis/gax-go/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m,
		// https://github.com/census-instrumentation/opencensus-go/issues/1262
		goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start"),
	)
}

var testKeyRSA = `-----BEGIN PUBLIC KEY-----
MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAu+OepfHCTiTi27nkTGke
dn+AIkiM1AIWWDwqfqG85aNulcj60mGQGXIYV8LoEVkyKOhYBIUmJUaVczB4ltqq
ZhR7l46RQw2vnv+XiUmfK555d4ZDInyjTusO69hE6tkuYKdXLlG1HzcrhJ254LE2
wXtE1Yf9DygOsWet+S32gmpfH2whUY1mRTdwW4zoY4c3qtmmWImhVVNr6qR8Z95X
Y49EteCoNIomQNEZH7EnMlBsh34L7doOsckh1aTvQcrJorQSrBkWKbdV6kvuBKZp
fLK0DZiOh9BwZCZANtOqgH3V+AuNk338iON8eKCFRjoiQ40YGM6xKH3E6PHVnuKt
uIO0MPvE0qdV8Lvs+nCCrvwP5sJKZuciM40ioEO1pV1y3491xIxYhx3OfN4gg2h8
cgdKob/R8qwxqTrfceO36FBFb1vXCUApsm5oy6WxmUtIUgoYhK+6JYpVWDyOJYwP
iMJhdJA65n2ZliN8NxEhsaFoMgw76BOiD0wkt/CKPmNbOm5MGS3/fiZCt6A6u3cn
Ubhn4tvjy/q5XzVqZtBeoseW2TyyrsAN53LBkSqag5tG/264CQDigQ6Y/OADOE2x
n08MyrFHIL/wFMscOvJo7c2Eo4EW1yXkEkAy5tF5PZgnfRObakj4gdqPeq18FNzc
Y+t5OxL3kL15VzY1Ob0d5cMCAwEAAQ==
-----END PUBLIC KEY-----`

type stubGCPClient struct {
	createErr                           error
	createCryptoKeyCalled               bool
	createCryptoKeyErr                  error
	createImportJobErr                  error
	decryptResponse                     []byte
	decryptErr                          error
	encryptErr                          error
	getKeyRingErr                       error
	importCryptoKeyVersionErr           error
	updateCryptoKeyPrimaryVersionCalled bool
	updateCryptoKeyPrimaryVersionErr    error
	getImportJobErr                     error
	getImportJobResponse                *kmspb.ImportJob
}

func newStubGCPClientFactory(stub *stubGCPClient) func(ctx context.Context, opts ...option.ClientOption) (clientAPI, error) {
	return func(ctx context.Context, opts ...option.ClientOption) (clientAPI, error) {
		return stub, stub.createErr
	}
}

func (s *stubGCPClient) Close() error {
	return nil
}

func (s *stubGCPClient) CreateCryptoKey(ctx context.Context, req *kmspb.CreateCryptoKeyRequest, opts ...gax.CallOption) (*kmspb.CryptoKey, error) {
	s.createCryptoKeyCalled = true
	return &kmspb.CryptoKey{}, s.createCryptoKeyErr
}

func (s *stubGCPClient) CreateImportJob(ctx context.Context, req *kmspb.CreateImportJobRequest, opts ...gax.CallOption) (*kmspb.ImportJob, error) {
	return &kmspb.ImportJob{}, s.createImportJobErr
}

func (s *stubGCPClient) Decrypt(ctx context.Context, req *kmspb.DecryptRequest, opts ...gax.CallOption) (*kmspb.DecryptResponse, error) {
	return &kmspb.DecryptResponse{Plaintext: s.decryptResponse}, s.decryptErr
}

func (s *stubGCPClient) Encrypt(ctx context.Context, req *kmspb.EncryptRequest, opts ...gax.CallOption) (*kmspb.EncryptResponse, error) {
	return &kmspb.EncryptResponse{}, s.encryptErr
}

func (s *stubGCPClient) GetKeyRing(ctx context.Context, req *kmspb.GetKeyRingRequest, opts ...gax.CallOption) (*kmspb.KeyRing, error) {
	return &kmspb.KeyRing{}, s.getKeyRingErr
}

func (s *stubGCPClient) ImportCryptoKeyVersion(ctx context.Context, req *kmspb.ImportCryptoKeyVersionRequest, opts ...gax.CallOption) (*kmspb.CryptoKeyVersion, error) {
	return &kmspb.CryptoKeyVersion{}, s.importCryptoKeyVersionErr
}

func (s *stubGCPClient) UpdateCryptoKeyPrimaryVersion(ctx context.Context, req *kmspb.UpdateCryptoKeyPrimaryVersionRequest, opts ...gax.CallOption) (*kmspb.CryptoKey, error) {
	s.updateCryptoKeyPrimaryVersionCalled = true
	return &kmspb.CryptoKey{}, s.updateCryptoKeyPrimaryVersionErr
}

func (s *stubGCPClient) GetImportJob(ctx context.Context, req *kmspb.GetImportJobRequest, opts ...gax.CallOption) (*kmspb.ImportJob, error) {
	return s.getImportJobResponse, s.getImportJobErr
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

func TestCreateKEK(t *testing.T) {
	someErr := errors.New("error")
	importKey := []byte("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")

	testCases := map[string]struct {
		client    *stubGCPClient
		importKey []byte
		wantErr   bool
	}{
		"create new kek successful": {
			client: &stubGCPClient{},
		},
		"import kek successful": {
			client: &stubGCPClient{
				getImportJobResponse: &kmspb.ImportJob{
					PublicKey: &kmspb.ImportJob_WrappingPublicKey{
						Pem: testKeyRSA,
					},
					State: kmspb.ImportJob_ACTIVE,
				},
			},
			importKey: importKey,
		},
		"CreateCryptoKey fails": {
			client:  &stubGCPClient{createCryptoKeyErr: someErr},
			wantErr: true,
		},
		"CreatCryptoKey fails on import": {
			client: &stubGCPClient{
				createCryptoKeyErr: someErr,
				getImportJobResponse: &kmspb.ImportJob{
					PublicKey: &kmspb.ImportJob_WrappingPublicKey{
						Pem: testKeyRSA,
					},
					State: kmspb.ImportJob_ACTIVE,
				},
			},
			importKey: importKey,
			wantErr:   true,
		},
		"CreateImportJob fails": {
			client:    &stubGCPClient{createImportJobErr: someErr},
			importKey: importKey,
			wantErr:   true,
		},
		"ImportCryptoKeyVersion fails": {
			client: &stubGCPClient{
				getImportJobResponse: &kmspb.ImportJob{
					PublicKey: &kmspb.ImportJob_WrappingPublicKey{
						Pem: testKeyRSA,
					},
					State: kmspb.ImportJob_ACTIVE,
				},
				importCryptoKeyVersionErr: someErr,
			},
			importKey: importKey,
			wantErr:   true,
		},
		"UpdateCryptoKeyPrimaryVersion fails": {
			client: &stubGCPClient{
				getImportJobResponse: &kmspb.ImportJob{
					PublicKey: &kmspb.ImportJob_WrappingPublicKey{
						Pem: testKeyRSA,
					},
					State: kmspb.ImportJob_ACTIVE,
				},
				updateCryptoKeyPrimaryVersionErr: someErr,
			},
			importKey: importKey,
			wantErr:   true,
		},
		"GetImportJob fails during waitBackoff": {
			client:    &stubGCPClient{getImportJobErr: someErr},
			importKey: importKey,
			wantErr:   true,
		},
		"GetImportJob returns no key": {
			client: &stubGCPClient{
				getImportJobResponse: &kmspb.ImportJob{
					State: kmspb.ImportJob_ACTIVE,
				},
			},
			importKey: importKey,
			wantErr:   true,
		},
		"waitBackoff times out": {
			client: &stubGCPClient{
				getImportJobResponse: &kmspb.ImportJob{
					PublicKey: &kmspb.ImportJob_WrappingPublicKey{
						Pem: testKeyRSA,
					},
					State: kmspb.ImportJob_PENDING_GENERATION,
				},
			},
			importKey: importKey,
			wantErr:   true,
		},
		"creating client fails": {
			client:  &stubGCPClient{createErr: someErr},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			client := &KMSClient{
				projectID:        "test-project",
				locationID:       "global",
				keyRingID:        "test-ring",
				newClient:        newStubGCPClientFactory(tc.client),
				protectionLevel:  kmspb.ProtectionLevel_SOFTWARE,
				waitBackoffLimit: 1,
			}

			err := client.CreateKEK(context.Background(), "test-key", tc.importKey)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				if len(tc.importKey) != 0 {
					assert.True(tc.client.updateCryptoKeyPrimaryVersionCalled)
				} else {
					assert.True(tc.client.createCryptoKeyCalled)
				}
			}
		})
	}
}

func TestGetDEK(t *testing.T) {
	someErr := errors.New("error")
	testKey := []byte("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")

	testCases := map[string]struct {
		client  *stubGCPClient
		storage kmsInterface.Storage
		wantErr bool
	}{
		"GetDEK successful for new key": {
			client:  &stubGCPClient{},
			storage: &stubStorage{getErr: storage.ErrDEKUnset},
		},
		"GetDEK successful for existing key": {
			client:  &stubGCPClient{decryptResponse: testKey},
			storage: &stubStorage{key: testKey},
		},
		"Get from storage fails": {
			client:  &stubGCPClient{},
			storage: &stubStorage{getErr: someErr},
			wantErr: true,
		},
		"Encrypt fails": {
			client:  &stubGCPClient{encryptErr: someErr},
			storage: &stubStorage{getErr: storage.ErrDEKUnset},
			wantErr: true,
		},
		"Encrypt fails with notfound error": {
			client:  &stubGCPClient{encryptErr: status.Error(codes.NotFound, "error")},
			storage: &stubStorage{getErr: storage.ErrDEKUnset},
			wantErr: true,
		},
		"Put to storage fails": {
			client: &stubGCPClient{},
			storage: &stubStorage{
				getErr: storage.ErrDEKUnset,
				putErr: someErr,
			},
			wantErr: true,
		},
		"Decrypt fails": {
			client:  &stubGCPClient{decryptErr: someErr},
			storage: &stubStorage{key: testKey},
			wantErr: true,
		},
		"Decrypt fails with notfound error": {
			client:  &stubGCPClient{decryptErr: status.Error(codes.NotFound, "error")},
			storage: &stubStorage{key: testKey},
			wantErr: true,
		},
		"creating client fails": {
			client:  &stubGCPClient{createErr: someErr},
			storage: &stubStorage{getErr: storage.ErrDEKUnset},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			client := &KMSClient{
				projectID:        "test-project",
				locationID:       "global",
				keyRingID:        "test-ring",
				newClient:        newStubGCPClientFactory(tc.client),
				protectionLevel:  kmspb.ProtectionLevel_SOFTWARE,
				waitBackoffLimit: 1,
				storage:          tc.storage,
			}

			dek, err := client.GetDEK(context.Background(), "volume-01", 32)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Len(dek, 32)
			}
		})
	}
}

func TestConnection(t *testing.T) {
	someErr := errors.New("error")
	testCases := map[string]struct {
		client  *stubGCPClient
		wantErr bool
	}{
		"success": {
			client: &stubGCPClient{},
		},
		"newClient fails": {
			client:  &stubGCPClient{createErr: someErr},
			wantErr: true,
		},
		"GetKeyRing fails": {
			client:  &stubGCPClient{getKeyRingErr: someErr},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			client := &KMSClient{
				projectID:        "test-project",
				locationID:       "global",
				keyRingID:        "test-ring",
				newClient:        newStubGCPClientFactory(tc.client),
				protectionLevel:  kmspb.ProtectionLevel_SOFTWARE,
				waitBackoffLimit: 1,
			}

			err := client.testConnection(context.Background())
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

func TestWrapCryptoKey(t *testing.T) {
	assert := assert.New(t)

	rsaPub, err := util.ParsePEMtoPublicKeyRSA([]byte(testKeyRSA))
	assert.NoError(err)

	res, err := wrapCryptoKey([]byte("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"), rsaPub)
	assert.NoError(err)
	assert.Equal(552, len(res))

	_, err = wrapCryptoKey([]byte{0x1}, rsaPub)
	assert.Error(err)
}
