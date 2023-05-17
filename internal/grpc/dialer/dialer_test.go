/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package dialer

import (
	"context"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/atls"
	"github.com/edgelesssys/constellation/v2/internal/grpc/atlscredentials"
	"github.com/edgelesssys/constellation/v2/internal/grpc/testdialer"
	"github.com/edgelesssys/constellation/v2/internal/variant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
	"google.golang.org/grpc"
	"google.golang.org/grpc/interop/grpc_testing"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

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
			dialer := New(nil, atls.NewFakeValidator(variant.Dummy{}), netDialer)
			server := newServer(variant.Dummy{}, tc.tls)
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

func newServer(oid variant.Getter, tls bool) *grpc.Server {
	if tls {
		creds := atlscredentials.New(atls.NewFakeIssuer(oid), nil)
		return grpc.NewServer(grpc.Creds(creds))
	}
	return grpc.NewServer()
}

type testAPI struct {
	grpc_testing.UnimplementedTestServiceServer
}

func (s *testAPI) EmptyCall(_ context.Context, _ *grpc_testing.Empty) (*grpc_testing.Empty, error) {
	return &grpc_testing.Empty{}, nil
}
