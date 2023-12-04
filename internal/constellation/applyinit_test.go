/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package constellation

import (
	"bytes"
	"context"
	"io"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/v2/bootstrapper/initproto"
	"github.com/edgelesssys/constellation/v2/internal/atls"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/grpc/atlscredentials"
	"github.com/edgelesssys/constellation/v2/internal/grpc/dialer"
	"github.com/edgelesssys/constellation/v2/internal/grpc/testdialer"
	"github.com/edgelesssys/constellation/v2/internal/kms/uri"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestInit(t *testing.T) {
	clusterEndpoint := "192.0.2.1"
	newState := func(endpoint string) *state.State {
		return &state.State{
			Infrastructure: state.Infrastructure{
				ClusterEndpoint: endpoint,
			},
		}
	}
	newInitServer := func(initErr error, responses ...*initproto.InitResponse) *stubInitServer {
		return &stubInitServer{
			res:     responses,
			initErr: initErr,
		}
	}

	testCases := map[string]struct {
		server             initproto.APIServer
		state              *state.State
		initServerEndpoint string
		wantClusterLogs    []byte
		wantErr            bool
	}{
		"success": {
			server: newInitServer(nil,
				&initproto.InitResponse{
					Kind: &initproto.InitResponse_InitSuccess{
						InitSuccess: &initproto.InitSuccessResponse{
							Kubeconfig: []byte{},
							OwnerId:    []byte{},
							ClusterId:  []byte{},
						},
					},
				}),
			state:              newState(clusterEndpoint),
			initServerEndpoint: clusterEndpoint,
		},
		"no response": {
			server:             newInitServer(nil),
			state:              newState(clusterEndpoint),
			initServerEndpoint: clusterEndpoint,
			wantErr:            true,
		},
		"nil response": {
			server:             newInitServer(nil, &initproto.InitResponse{Kind: nil}),
			state:              newState(clusterEndpoint),
			initServerEndpoint: clusterEndpoint,
			wantErr:            true,
		},
		"failure response": {
			server: newInitServer(nil,
				&initproto.InitResponse{
					Kind: &initproto.InitResponse_InitFailure{
						InitFailure: &initproto.InitFailureResponse{
							Error: assert.AnError.Error(),
						},
					},
				}),
			state:              newState(clusterEndpoint),
			initServerEndpoint: clusterEndpoint,
			wantErr:            true,
		},
		"setup server error": {
			server:             newInitServer(assert.AnError),
			state:              newState(clusterEndpoint),
			initServerEndpoint: clusterEndpoint,
			wantErr:            true,
		},
		"expected log response, got failure": {
			server: newInitServer(nil,
				&initproto.InitResponse{
					Kind: &initproto.InitResponse_InitFailure{
						InitFailure: &initproto.InitFailureResponse{
							Error: assert.AnError.Error(),
						},
					},
				},
				&initproto.InitResponse{
					Kind: &initproto.InitResponse_InitFailure{
						InitFailure: &initproto.InitFailureResponse{
							Error: assert.AnError.Error(),
						},
					},
				},
			),
			state:              newState(clusterEndpoint),
			initServerEndpoint: clusterEndpoint,
			wantErr:            true,
		},
		"expected log response, got success": {
			server: newInitServer(nil,
				&initproto.InitResponse{
					Kind: &initproto.InitResponse_InitFailure{
						InitFailure: &initproto.InitFailureResponse{
							Error: assert.AnError.Error(),
						},
					},
				},
				&initproto.InitResponse{
					Kind: &initproto.InitResponse_InitSuccess{
						InitSuccess: &initproto.InitSuccessResponse{
							Kubeconfig: []byte{},
							OwnerId:    []byte{},
							ClusterId:  []byte{},
						},
					},
				},
			),
			state:              newState(clusterEndpoint),
			initServerEndpoint: clusterEndpoint,
			wantErr:            true,
		},
		"collect logs": {
			server: newInitServer(nil,
				&initproto.InitResponse{
					Kind: &initproto.InitResponse_InitFailure{
						InitFailure: &initproto.InitFailureResponse{
							Error: assert.AnError.Error(),
						},
					},
				},
				&initproto.InitResponse{
					Kind: &initproto.InitResponse_Log{
						Log: &initproto.LogResponseType{
							Log: []byte("some log"),
						},
					},
				},
			),
			wantClusterLogs:    []byte("some log"),
			state:              newState(clusterEndpoint),
			initServerEndpoint: clusterEndpoint,
			wantErr:            true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := require.New(t)

			netDialer := testdialer.NewBufconnDialer()
			stop := setupTestInitServer(netDialer, tc.server, tc.initServerEndpoint)
			defer stop()

			a := &Applier{
				log:     logger.NewTest(t),
				spinner: &nopSpinner{},
				newDialer: func(atls.Validator) *dialer.Dialer {
					return dialer.New(nil, nil, netDialer)
				},
			}

			clusterLogs := &bytes.Buffer{}
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*4)
			defer cancel()
			_, err := a.Init(ctx, nil, tc.state, clusterLogs, InitPayload{
				MasterSecret:    uri.MasterSecret{},
				MeasurementSalt: []byte{},
				K8sVersion:      "v1.26.5",
				ConformanceMode: false,
			})
			if tc.wantErr {
				assert.Error(err)
				assert.Equal(tc.wantClusterLogs, clusterLogs.Bytes())
			} else {
				assert.NoError(err)
			}
		})
	}
}

type nopSpinner struct {
	io.Writer
}

func (s *nopSpinner) Start(string, bool) {}
func (s *nopSpinner) Stop()              {}
func (s *nopSpinner) Write(p []byte) (n int, err error) {
	return s.Writer.Write(p)
}

func setupTestInitServer(dialer *testdialer.BufconnDialer, server initproto.APIServer, host string) func() {
	serverCreds := atlscredentials.New(nil, nil)
	initServer := grpc.NewServer(grpc.Creds(serverCreds))
	initproto.RegisterAPIServer(initServer, server)
	listener := dialer.GetListener(net.JoinHostPort(host, strconv.Itoa(constants.BootstrapperPort)))
	go initServer.Serve(listener)
	return initServer.GracefulStop
}

type stubInitServer struct {
	res     []*initproto.InitResponse
	initErr error

	initproto.UnimplementedAPIServer
}

func (s *stubInitServer) Init(_ *initproto.InitRequest, stream initproto.API_InitServer) error {
	for _, r := range s.res {
		_ = stream.Send(r)
	}
	return s.initErr
}
