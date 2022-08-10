package recoveryserver

import (
	"context"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/disk-mapper/recoverproto"
	"github.com/edgelesssys/constellation/internal/atls"
	"github.com/edgelesssys/constellation/internal/grpc/dialer"
	"github.com/edgelesssys/constellation/internal/grpc/testdialer"
	"github.com/edgelesssys/constellation/internal/logger"
	"github.com/edgelesssys/constellation/internal/oid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestServe(t *testing.T) {
	assert := assert.New(t)
	log := logger.NewTest(t)
	uuid := "uuid"
	server := New(atls.NewFakeIssuer(oid.Dummy{}), log)
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

	server = New(atls.NewFakeIssuer(oid.Dummy{}), log)
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
		initialMsg message
		keyMsg     message
		wantErr    bool
	}{
		"success": {
			initialMsg: message{
				recoverMsg: &recoverproto.RecoverMessage{
					Request: &recoverproto.RecoverMessage_MeasurementSecret{
						MeasurementSecret: []byte("measurementSecret"),
					},
				},
			},
			keyMsg: message{
				recoverMsg: &recoverproto.RecoverMessage{
					Request: &recoverproto.RecoverMessage_StateDiskKey{
						StateDiskKey: []byte("diskKey"),
					},
				},
			},
		},
		"first message is not a measurement secret": {
			initialMsg: message{
				recoverMsg: &recoverproto.RecoverMessage{
					Request: &recoverproto.RecoverMessage_StateDiskKey{
						StateDiskKey: []byte("diskKey"),
					},
				},
				wantErr: true,
			},
			keyMsg: message{
				recoverMsg: &recoverproto.RecoverMessage{
					Request: &recoverproto.RecoverMessage_StateDiskKey{
						StateDiskKey: []byte("diskKey"),
					},
				},
			},
		},
		"second message is not a state disk key": {
			initialMsg: message{
				recoverMsg: &recoverproto.RecoverMessage{
					Request: &recoverproto.RecoverMessage_MeasurementSecret{
						MeasurementSecret: []byte("measurementSecret"),
					},
				},
			},
			keyMsg: message{
				recoverMsg: &recoverproto.RecoverMessage{
					Request: &recoverproto.RecoverMessage_MeasurementSecret{
						MeasurementSecret: []byte("measurementSecret"),
					},
				},
				wantErr: true,
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			ctx := context.Background()
			serverUUID := "uuid"
			server := New(atls.NewFakeIssuer(oid.Dummy{}), logger.NewTest(t))
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
			client, err := recoverproto.NewAPIClient(conn).Recover(ctx)
			require.NoError(err)

			// Send initial message
			err = client.Send(tc.initialMsg.recoverMsg)
			require.NoError(err)

			// Receive uuid
			uuid, err := client.Recv()
			if tc.initialMsg.wantErr {
				assert.Error(err)
				return
			}
			assert.Equal(serverUUID, uuid.DiskUuid)

			// Send key message
			err = client.Send(tc.keyMsg.recoverMsg)
			require.NoError(err)

			_, err = client.Recv()
			if tc.keyMsg.wantErr {
				assert.Error(err)
				return
			}
			assert.ErrorIs(io.EOF, err)

			wg.Wait()
			assert.NoError(serveErr)
			assert.Equal(tc.initialMsg.recoverMsg.GetMeasurementSecret(), measurementSecret)
			assert.Equal(tc.keyMsg.recoverMsg.GetStateDiskKey(), diskKey)
		})
	}
}

type message struct {
	recoverMsg *recoverproto.RecoverMessage
	wantErr    bool
}
