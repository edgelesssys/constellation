/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package recoveryserver

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/v2/disk-mapper/recoverproto"
	"github.com/edgelesssys/constellation/v2/internal/atls"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/grpc/dialer"
	"github.com/edgelesssys/constellation/v2/internal/grpc/testdialer"
	"github.com/edgelesssys/constellation/v2/internal/kms/kms"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m,
		// https://github.com/census-instrumentation/opencensus-go/issues/1262
		goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start"),
		goleak.IgnoreAnyFunction("github.com/bazelbuild/rules_go/go/tools/bzltestutil.RegisterTimeoutHandler.func1"),
	)
}

func TestServe(t *testing.T) {
	assert := assert.New(t)
	log := slog.New(slog.NewTextHandler(logger.TestWriter{T: t}, nil))
	uuid := "uuid"
	server := New(atls.NewFakeIssuer(variant.Dummy{}), newStubKMS(nil, nil), log)
	dialer := testdialer.NewBufconnDialer()
	listener := dialer.GetListener("192.0.2.1:1234")
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	// Serve method returns when context is canceled
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, _, err := server.Serve(ctx, listener, uuid)
		assert.ErrorIs(err, context.Canceled)
	}()
	time.Sleep(100 * time.Millisecond)
	cancel()
	wg.Wait()

	server = New(atls.NewFakeIssuer(variant.Dummy{}), newStubKMS(nil, nil), log)
	dialer = testdialer.NewBufconnDialer()
	listener = dialer.GetListener("192.0.2.1:1234")

	// Serve method returns without error when the server is shut down
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, _, err := server.Serve(context.Background(), listener, uuid)
		assert.NoError(err)
	}()
	time.Sleep(100 * time.Millisecond)
	server.grpcServer.GracefulStop()
	wg.Wait()

	// Serve method returns an error when serving is unsuccessful
	_, _, err := server.Serve(context.Background(), listener, uuid)
	assert.Error(err)
}

func TestRecover(t *testing.T) {
	testCases := map[string]struct {
		kmsURI     string
		storageURI string
		factory    kmsFactory
		wantErr    bool
	}{
		"success": {
			// base64 encoded: key=masterkey&salt=somesalt
			kmsURI:     "kms://cluster-kms?key=bWFzdGVya2V5&salt=c29tZXNhbHQ=",
			storageURI: "storage://no-store",
			factory:    newStubKMS(nil, nil),
		},
		"kms init fails": {
			factory: newStubKMS(errors.New("setup failed"), nil),
			wantErr: true,
		},
		"GetDEK fails": {
			kmsURI:     "kms://cluster-kms?key=bWFzdGVya2V5&salt=c29tZXNhbHQ=",
			storageURI: "storage://no-store",
			factory:    newStubKMS(nil, errors.New("GetDEK failed")),
			wantErr:    true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			ctx := context.Background()
			serverUUID := "uuid"
			server := New(atls.NewFakeIssuer(variant.Dummy{}), tc.factory, slog.New(slog.NewTextHandler(logger.TestWriter{T: t}, nil)))
			netDialer := testdialer.NewBufconnDialer()
			listener := netDialer.GetListener("192.0.2.1:1234")

			var diskKey, measurementSecret []byte
			var serveErr error
			var wg sync.WaitGroup
			defer wg.Wait()

			serveCtx, cancel := context.WithCancel(ctx)
			defer cancel()
			wg.Add(1)
			go func() {
				defer wg.Done()
				diskKey, measurementSecret, serveErr = server.Serve(serveCtx, listener, serverUUID)
			}()

			conn, err := dialer.New(nil, nil, netDialer).Dial(ctx, "192.0.2.1:1234")
			require.NoError(err)
			defer conn.Close()

			req := recoverproto.RecoverMessage{
				KmsUri:     tc.kmsURI,
				StorageUri: tc.storageURI,
			}
			_, err = recoverproto.NewAPIClient(conn).Recover(ctx, &req)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			wg.Wait()
			require.NoError(serveErr)
			assert.NoError(err)
			assert.NotNil(measurementSecret)
			assert.NotNil(diskKey)
		})
	}
}

func newStubKMS(setupErr, getDEKErr error) kmsFactory {
	return func(ctx context.Context, storageURI string, kmsURI string) (kms.CloudKMS, error) {
		if setupErr != nil {
			return nil, setupErr
		}
		return &stubKMS{getDEKErr: getDEKErr}, nil
	}
}

type stubKMS struct {
	kms.CloudKMS
	getDEKErr error
}

func (s *stubKMS) GetDEK(_ context.Context, _ string, _ int) ([]byte, error) {
	if s.getDEKErr != nil {
		return nil, s.getDEKErr
	}
	return []byte("someDEK"), nil
}
