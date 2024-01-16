/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package internal

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	cloudkms "cloud.google.com/go/kms/apiv1"
	"github.com/edgelesssys/constellation/v2/internal/kms/storage"
	wrapping "github.com/hashicorp/go-kms-wrapping/v2"
	"github.com/hashicorp/go-kms-wrapping/wrappers/gcpckms/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m,
		// https://github.com/census-instrumentation/opencensus-go/issues/1262
		goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start"),
		goleak.IgnoreAnyFunction("github.com/bazelbuild/rules_go/go/tools/bzltestutil.RegisterTimeoutHandler.func1"),
	)
}

type stubWrapper struct {
	decryptErr      error
	decryptResponse []byte
	encryptErr      error
	encryptResponse *wrapping.BlobInfo
}

func (s *stubWrapper) Decrypt(context.Context, *wrapping.BlobInfo, ...wrapping.Option) ([]byte, error) {
	return s.decryptResponse, s.decryptErr
}

func (s *stubWrapper) Encrypt(context.Context, []byte, ...wrapping.Option) (*wrapping.BlobInfo, error) {
	return s.encryptResponse, s.encryptErr
}

func (s *stubWrapper) Client() *cloudkms.KeyManagementClient {
	return &cloudkms.KeyManagementClient{}
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

func TestGetDEK(t *testing.T) {
	someErr := errors.New("failed")
	testKey := []byte("00112233445566778899aabbccddeeff")
	savedTestKey, err := json.Marshal(&wrapping.BlobInfo{
		Ciphertext: []byte("encrypted-dek"),
		Iv:         []byte("iv"),
		KeyInfo: &wrapping.KeyInfo{
			Mechanism:  gcpckms.GcpCkmsEnvelopeAesGcmEncrypt,
			KeyId:      "kek-id",
			WrappedKey: []byte("dek-encryption-key"),
		},
	})
	require.NoError(t, err)

	testCases := map[string]struct {
		wrapper *stubWrapper
		storage *stubStorage
		wantErr bool
	}{
		"GetDEK successful for new key": {
			wrapper: &stubWrapper{},
			storage: &stubStorage{getErr: storage.ErrDEKUnset},
		},
		"GetDEK successful for existing key": {
			wrapper: &stubWrapper{decryptResponse: testKey},
			storage: &stubStorage{key: savedTestKey},
		},
		"Get from storage fails": {
			wrapper: &stubWrapper{},
			storage: &stubStorage{getErr: someErr},
			wantErr: true,
		},
		"Encrypt fails": {
			wrapper: &stubWrapper{encryptErr: someErr},
			storage: &stubStorage{getErr: storage.ErrDEKUnset},
			wantErr: true,
		},
		"Encrypt fails with notfound error": {
			wrapper: &stubWrapper{encryptErr: status.Error(codes.NotFound, "error")},
			storage: &stubStorage{getErr: storage.ErrDEKUnset},
			wantErr: true,
		},
		"Put to storage fails": {
			wrapper: &stubWrapper{},
			storage: &stubStorage{
				getErr: storage.ErrDEKUnset,
				putErr: someErr,
			},
			wantErr: true,
		},
		"Decrypt fails": {
			wrapper: &stubWrapper{decryptErr: someErr},
			storage: &stubStorage{key: savedTestKey},
			wantErr: true,
		},
		"Decrypt fails with notfound error": {
			wrapper: &stubWrapper{decryptErr: status.Error(codes.NotFound, "error")},
			storage: &stubStorage{key: savedTestKey},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			client := &KMSClient{
				Wrapper: tc.wrapper,
				Storage: tc.storage,
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
