package dialer

import (
	"context"
	"testing"

	"github.com/edgelesssys/constellation/coordinator/core"
	"github.com/edgelesssys/constellation/internal/grpc/atlscredentials"
	"github.com/edgelesssys/constellation/internal/grpc/testdialer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/grpc_testing"
)

func TestDial(t *testing.T) {
	testCases := map[string]struct {
		tls     bool
		dialFn  func(dialer *Dialer, ctx context.Context, target string) (*grpc.ClientConn, error)
		wantErr bool
	}{
		"Dial with tls on server works": {
			tls: true,
			dialFn: func(dialer *Dialer, ctx context.Context, target string) (*grpc.ClientConn, error) {
				return dialer.Dial(ctx, target)
			},
		},
		"Dial without tls on server fails": {
			dialFn: func(dialer *Dialer, ctx context.Context, target string) (*grpc.ClientConn, error) {
				return dialer.Dial(ctx, target)
			},
			wantErr: true,
		},
		"DialNoVerify with tls on server works": {
			tls: true,
			dialFn: func(dialer *Dialer, ctx context.Context, target string) (*grpc.ClientConn, error) {
				return dialer.DialNoVerify(ctx, target)
			},
		},
		"DialNoVerify without tls on server fails": {
			dialFn: func(dialer *Dialer, ctx context.Context, target string) (*grpc.ClientConn, error) {
				return dialer.DialNoVerify(ctx, target)
			},
			wantErr: true,
		},
		"DialInsecure without tls on server works": {
			dialFn: func(dialer *Dialer, ctx context.Context, target string) (*grpc.ClientConn, error) {
				return dialer.DialInsecure(ctx, target)
			},
		},
		"DialInsecure with tls on server fails": {
			tls: true,
			dialFn: func(dialer *Dialer, ctx context.Context, target string) (*grpc.ClientConn, error) {
				return dialer.DialInsecure(ctx, target)
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			netDialer := testdialer.NewBufconnDialer()
			dialer := New(nil, &core.MockValidator{}, netDialer)
			server := newServer(tc.tls)
			api := &testAPI{}
			grpc_testing.RegisterTestServiceServer(server, api)
			go server.Serve(netDialer.GetListener("192.0.2.1:1234"))
			defer server.Stop()
			conn, err := tc.dialFn(dialer, context.Background(), "192.0.2.1:1234")
			require.NoError(err)
			defer conn.Close()

			client := grpc_testing.NewTestServiceClient(conn)
			_, err = client.EmptyCall(context.Background(), &grpc_testing.Empty{})

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
		})
	}
}

func newServer(tls bool) *grpc.Server {
	if tls {
		creds := atlscredentials.New(&core.MockIssuer{}, nil)
		return grpc.NewServer(grpc.Creds(creds))
	}
	return grpc.NewServer()
}

type testAPI struct {
	grpc_testing.UnimplementedTestServiceServer
}

func (s *testAPI) EmptyCall(ctx context.Context, in *grpc_testing.Empty) (*grpc_testing.Empty, error) {
	return &grpc_testing.Empty{}, nil
}
