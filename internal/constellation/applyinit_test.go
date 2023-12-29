/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package constellation

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/v2/bootstrapper/initproto"
	"github.com/edgelesssys/constellation/v2/internal/atls"
	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/constellation/state"
	"github.com/edgelesssys/constellation/v2/internal/grpc/atlscredentials"
	"github.com/edgelesssys/constellation/v2/internal/grpc/dialer"
	"github.com/edgelesssys/constellation/v2/internal/grpc/testdialer"
	"github.com/edgelesssys/constellation/v2/internal/kms/uri"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"k8s.io/client-go/tools/clientcmd"
	k8sclientapi "k8s.io/client-go/tools/clientcmd/api"
)

func TestInit(t *testing.T) {
	respKubeconfig := k8sclientapi.Config{
		Clusters: map[string]*k8sclientapi.Cluster{
			"cluster": {
				Server: "https://192.0.2.1:6443",
			},
		},
	}
	respKubeconfigBytes, err := clientcmd.Write(respKubeconfig)
	require.NoError(t, err)

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
							Kubeconfig: respKubeconfigBytes,
							OwnerId:    []byte{},
							ClusterId:  []byte{},
						},
					},
				}),
			state:              newState(clusterEndpoint),
			initServerEndpoint: clusterEndpoint,
		},
		"kubeconfig without clusters": {
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
			wantErr:            true,
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
							Kubeconfig: respKubeconfigBytes,
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
				log:     slog.New(slog.NewTextHandler(logger.TestWriter{T: t}, nil)),
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

func TestAttestation(t *testing.T) {
	assert := assert.New(t)

	initServerAPI := &stubInitServer{res: []*initproto.InitResponse{
		{
			Kind: &initproto.InitResponse_InitSuccess{
				InitSuccess: &initproto.InitSuccessResponse{
					Kubeconfig: []byte("kubeconfig"),
					OwnerId:    []byte("ownerID"),
					ClusterId:  []byte("clusterID"),
				},
			},
		},
	}}

	netDialer := testdialer.NewBufconnDialer()

	issuer := &testIssuer{
		Getter: variant.QEMUVTPM{},
		pcrs: map[uint32][]byte{
			0: bytes.Repeat([]byte{0xFF}, 32),
			1: bytes.Repeat([]byte{0xFF}, 32),
			2: bytes.Repeat([]byte{0xFF}, 32),
			3: bytes.Repeat([]byte{0xFF}, 32),
		},
	}
	serverCreds := atlscredentials.New(issuer, nil)
	initServer := grpc.NewServer(grpc.Creds(serverCreds))
	initproto.RegisterAPIServer(initServer, initServerAPI)
	port := strconv.Itoa(constants.BootstrapperPort)
	listener := netDialer.GetListener(net.JoinHostPort("192.0.2.4", port))
	go initServer.Serve(listener)
	defer initServer.GracefulStop()

	validator := &testValidator{
		Getter: variant.QEMUVTPM{},
		pcrs: measurements.M{
			0:  measurements.WithAllBytes(0x00, measurements.Enforce, measurements.PCRMeasurementLength),
			1:  measurements.WithAllBytes(0x11, measurements.Enforce, measurements.PCRMeasurementLength),
			2:  measurements.WithAllBytes(0x22, measurements.Enforce, measurements.PCRMeasurementLength),
			3:  measurements.WithAllBytes(0x33, measurements.Enforce, measurements.PCRMeasurementLength),
			4:  measurements.WithAllBytes(0x44, measurements.Enforce, measurements.PCRMeasurementLength),
			9:  measurements.WithAllBytes(0x99, measurements.Enforce, measurements.PCRMeasurementLength),
			12: measurements.WithAllBytes(0xcc, measurements.Enforce, measurements.PCRMeasurementLength),
		},
	}
	state := &state.State{Version: state.Version1, Infrastructure: state.Infrastructure{ClusterEndpoint: "192.0.2.4"}}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()

	initer := &Applier{
		log: slog.New(slog.NewTextHandler(logger.TestWriter{T: t}, nil)),
		newDialer: func(v atls.Validator) *dialer.Dialer {
			return dialer.New(nil, v, netDialer)
		},
		spinner: &nopSpinner{},
	}

	_, err := initer.Init(ctx, validator, state, io.Discard, InitPayload{
		MasterSecret:    uri.MasterSecret{},
		MeasurementSalt: []byte{},
		K8sVersion:      "v1.26.5",
		ConformanceMode: false,
	})
	assert.Error(err)
	// make sure the error is actually a TLS handshake error
	assert.Contains(err.Error(), "transport: authentication handshake failed")
	if validationErr, ok := err.(*config.ValidationError); ok {
		t.Log(validationErr.LongMessage())
	}
}

type testValidator struct {
	variant.Getter
	pcrs measurements.M
}

func (v *testValidator) Validate(_ context.Context, attDoc []byte, _ []byte) ([]byte, error) {
	var attestation struct {
		UserData []byte
		PCRs     map[uint32][]byte
	}
	if err := json.Unmarshal(attDoc, &attestation); err != nil {
		return nil, err
	}

	for k, pcr := range v.pcrs {
		if !bytes.Equal(attestation.PCRs[k], pcr.Expected[:]) {
			return nil, errors.New("invalid PCR value")
		}
	}
	return attestation.UserData, nil
}

type testIssuer struct {
	variant.Getter
	pcrs map[uint32][]byte
}

func (i *testIssuer) Issue(_ context.Context, userData []byte, _ []byte) ([]byte, error) {
	return json.Marshal(
		struct {
			UserData []byte
			PCRs     map[uint32][]byte
		}{
			UserData: userData,
			PCRs:     i.pcrs,
		},
	)
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
