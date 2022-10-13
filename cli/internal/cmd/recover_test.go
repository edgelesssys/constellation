/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"bytes"
	"context"
	"errors"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/v2/cli/internal/cloudcmd"
	"github.com/edgelesssys/constellation/v2/cli/internal/clusterid"
	"github.com/edgelesssys/constellation/v2/disk-mapper/recoverproto"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/crypto/testvector"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/grpc/atlscredentials"
	"github.com/edgelesssys/constellation/v2/internal/grpc/dialer"
	"github.com/edgelesssys/constellation/v2/internal/grpc/testdialer"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestRecoverCmdArgumentValidation(t *testing.T) {
	testCases := map[string]struct {
		args    []string
		wantErr bool
	}{
		"no args":            {[]string{}, false},
		"too many arguments": {[]string{"abc"}, true},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			cmd := NewRecoverCmd()
			err := cmd.ValidateArgs(tc.args)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

func TestRecover(t *testing.T) {
	someErr := errors.New("error")
	unavailableErr := status.Error(codes.Unavailable, "unavailable")
	lbErr := status.Error(codes.Unavailable, `connection error: desc = "transport: authentication handshake failed: read tcp`)

	testCases := map[string]struct {
		doer            *stubDoer
		masterSecret    testvector.HKDF
		endpoint        string
		configFlag      string
		successfulCalls int
		wantErr         bool
	}{
		"works": {
			doer:            &stubDoer{returns: []error{nil}},
			endpoint:        "192.0.2.90",
			masterSecret:    testvector.HKDFZero,
			successfulCalls: 1,
		},
		"missing config": {
			doer:         &stubDoer{returns: []error{nil}},
			endpoint:     "192.0.2.89",
			masterSecret: testvector.HKDFZero,
			configFlag:   "nonexistent-config",
			wantErr:      true,
		},
		"success multiple nodes": {
			doer:            &stubDoer{returns: []error{nil, nil}},
			endpoint:        "192.0.2.90",
			masterSecret:    testvector.HKDFZero,
			successfulCalls: 2,
		},
		"no nodes to recover does not error": {
			doer:            &stubDoer{returns: []error{unavailableErr}},
			endpoint:        "192.0.2.90",
			masterSecret:    testvector.HKDFZero,
			successfulCalls: 0,
		},
		"error on first node": {
			doer:            &stubDoer{returns: []error{someErr, nil}},
			endpoint:        "192.0.2.90",
			masterSecret:    testvector.HKDFZero,
			successfulCalls: 0,
			wantErr:         true,
		},
		"unavailable error is retried once": {
			doer:            &stubDoer{returns: []error{unavailableErr, nil}},
			endpoint:        "192.0.2.90",
			masterSecret:    testvector.HKDFZero,
			successfulCalls: 1,
		},
		"unavailable error is not retried twice": {
			doer:            &stubDoer{returns: []error{unavailableErr, unavailableErr, nil}},
			endpoint:        "192.0.2.90",
			masterSecret:    testvector.HKDFZero,
			successfulCalls: 0,
		},
		"unavailable error is not retried twice after success": {
			doer:            &stubDoer{returns: []error{nil, unavailableErr, unavailableErr, nil}},
			endpoint:        "192.0.2.90",
			masterSecret:    testvector.HKDFZero,
			successfulCalls: 1,
		},
		"transient LB errors are retried": {
			doer:            &stubDoer{returns: []error{lbErr, lbErr, lbErr, nil}},
			endpoint:        "192.0.2.90",
			masterSecret:    testvector.HKDFZero,
			successfulCalls: 1,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			cmd := NewRecoverCmd()
			cmd.SetContext(context.Background())
			cmd.Flags().String("config", constants.ConfigFilename, "") // register persistent flag manually
			out := &bytes.Buffer{}
			cmd.SetOut(out)
			cmd.SetErr(out)
			require.NoError(cmd.Flags().Set("endpoint", tc.endpoint))

			if tc.configFlag != "" {
				require.NoError(cmd.Flags().Set("config", tc.configFlag))
			}

			fs := afero.NewMemMapFs()
			fileHandler := file.NewHandler(fs)

			config := defaultConfigWithExpectedMeasurements(t, config.Default(), cloudprovider.GCP)
			require.NoError(fileHandler.WriteYAML(constants.ConfigFilename, config))

			require.NoError(fileHandler.WriteJSON(
				"constellation-mastersecret.json",
				masterSecret{Key: tc.masterSecret.Secret, Salt: tc.masterSecret.Salt},
				file.OptNone,
			))

			newDialer := func(*cloudcmd.Validator) *dialer.Dialer { return nil }

			err := recover(cmd, fileHandler, time.Millisecond, tc.doer, newDialer)
			if tc.wantErr {
				assert.Error(err)
				if tc.successfulCalls > 0 {
					assert.Contains(out.String(), strconv.Itoa(tc.successfulCalls))
				}
				return
			}

			assert.NoError(err)
			if tc.successfulCalls > 0 {
				assert.Contains(out.String(), "Pushed recovery key.")
				assert.Contains(out.String(), strconv.Itoa(tc.successfulCalls))
			} else {
				assert.Contains(out.String(), "No control-plane nodes in need of recovery found.")
			}
		})
	}
}

func TestParseRecoverFlags(t *testing.T) {
	testCases := map[string]struct {
		args        []string
		wantFlags   recoverFlags
		writeIDFile bool
		wantErr     bool
	}{
		"no flags": {
			wantFlags: recoverFlags{
				endpoint:   "192.0.2.42:9999",
				secretPath: "constellation-mastersecret.json",
			},
			writeIDFile: true,
		},
		"no flags, no ID file": {
			wantFlags: recoverFlags{
				endpoint:   "192.0.2.42:9999",
				secretPath: "constellation-mastersecret.json",
			},
			wantErr: true,
		},
		"invalid endpoint": {
			args:    []string{"-e", "192.0.2.42:2:2"},
			wantErr: true,
		},
		"all args set": {
			args: []string{"-e", "192.0.2.42:2", "--config", "config-path", "--master-secret", "/path/super-secret.json"},
			wantFlags: recoverFlags{
				endpoint:   "192.0.2.42:2",
				secretPath: "/path/super-secret.json",
				configPath: "config-path",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			cmd := NewRecoverCmd()
			cmd.Flags().String("config", "", "") // register persistent flag manually
			require.NoError(cmd.ParseFlags(tc.args))

			fileHandler := file.NewHandler(afero.NewMemMapFs())
			if tc.writeIDFile {
				require.NoError(fileHandler.WriteJSON(constants.ClusterIDsFileName, &clusterid.File{IP: "192.0.2.42"}))
			}

			flags, err := parseRecoverFlags(cmd, fileHandler)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			assert.Equal(tc.wantFlags, flags)
		})
	}
}

func TestDoRecovery(t *testing.T) {
	someErr := errors.New("error")
	testCases := map[string]struct {
		recoveryServer *stubRecoveryServer
		wantErr        bool
	}{
		"success": {
			recoveryServer: &stubRecoveryServer{
				actions: [][]func(stream recoverproto.API_RecoverServer) error{{
					func(stream recoverproto.API_RecoverServer) error {
						_, err := stream.Recv()
						return err
					},
					func(stream recoverproto.API_RecoverServer) error {
						return stream.Send(&recoverproto.RecoverResponse{
							DiskUuid: "00000000-0000-0000-0000-000000000000",
						})
					},
					func(stream recoverproto.API_RecoverServer) error {
						_, err := stream.Recv()
						return err
					},
				}},
			},
		},
		"error on first recv": {
			recoveryServer: &stubRecoveryServer{
				actions: [][]func(stream recoverproto.API_RecoverServer) error{
					{
						func(stream recoverproto.API_RecoverServer) error {
							return someErr
						},
					},
				},
			},
			wantErr: true,
		},
		"error on send": {
			recoveryServer: &stubRecoveryServer{
				actions: [][]func(stream recoverproto.API_RecoverServer) error{
					{
						func(stream recoverproto.API_RecoverServer) error {
							_, err := stream.Recv()
							return err
						},
						func(stream recoverproto.API_RecoverServer) error {
							return someErr
						},
					},
				},
			},
			wantErr: true,
		},
		"error on second recv": {
			recoveryServer: &stubRecoveryServer{
				actions: [][]func(stream recoverproto.API_RecoverServer) error{
					{
						func(stream recoverproto.API_RecoverServer) error {
							_, err := stream.Recv()
							return err
						},
						func(stream recoverproto.API_RecoverServer) error {
							return stream.Send(&recoverproto.RecoverResponse{
								DiskUuid: "00000000-0000-0000-0000-000000000000",
							})
						},
						func(stream recoverproto.API_RecoverServer) error {
							return someErr
						},
					},
				},
			},
			wantErr: true,
		},
		"final message is an error": {
			recoveryServer: &stubRecoveryServer{
				actions: [][]func(stream recoverproto.API_RecoverServer) error{{
					func(stream recoverproto.API_RecoverServer) error {
						_, err := stream.Recv()
						return err
					},
					func(stream recoverproto.API_RecoverServer) error {
						return stream.Send(&recoverproto.RecoverResponse{
							DiskUuid: "00000000-0000-0000-0000-000000000000",
						})
					},
					func(stream recoverproto.API_RecoverServer) error {
						_, err := stream.Recv()
						return err
					},
					func(stream recoverproto.API_RecoverServer) error {
						return someErr
					},
				}},
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			netDialer := testdialer.NewBufconnDialer()
			serverCreds := atlscredentials.New(nil, nil)
			recoverServer := grpc.NewServer(grpc.Creds(serverCreds))
			recoverproto.RegisterAPIServer(recoverServer, tc.recoveryServer)
			addr := net.JoinHostPort("192.0.42.42", strconv.Itoa(constants.RecoveryPort))
			listener := netDialer.GetListener(addr)
			go recoverServer.Serve(listener)
			defer recoverServer.GracefulStop()

			recoverDoer := &recoverDoer{
				dialer:            dialer.New(nil, nil, netDialer),
				endpoint:          addr,
				measurementSecret: []byte("measurement-secret"),
				getDiskKey: func(string) ([]byte, error) {
					return []byte("disk-key"), nil
				},
			}

			err := recoverDoer.Do(context.Background())
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

func TestDeriveStateDiskKey(t *testing.T) {
	testCases := map[string]struct {
		masterSecret testvector.HKDF
	}{
		"all zero": {
			masterSecret: testvector.HKDFZero,
		},
		"all 0xff": {
			masterSecret: testvector.HKDF0xFF,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			getKeyFunc := getStateDiskKeyFunc(tc.masterSecret.Secret, tc.masterSecret.Salt)
			stateDiskKey, err := getKeyFunc(tc.masterSecret.Info)

			assert.NoError(err)
			assert.Equal(tc.masterSecret.Output, stateDiskKey)
		})
	}
}

type stubRecoveryServer struct {
	actions [][]func(recoverproto.API_RecoverServer) error
	calls   int
	recoverproto.UnimplementedAPIServer
}

func (s *stubRecoveryServer) Recover(stream recoverproto.API_RecoverServer) error {
	if s.calls >= len(s.actions) {
		return status.Error(codes.Unavailable, "server is unavailable")
	}
	s.calls++

	for _, action := range s.actions[s.calls-1] {
		if err := action(stream); err != nil {
			return err
		}
	}
	return nil
}

type stubDoer struct {
	returns []error
}

func (d *stubDoer) Do(context.Context) error {
	err := d.returns[0]
	if len(d.returns) > 1 {
		d.returns = d.returns[1:]
	} else {
		d.returns = []error{status.Error(codes.Unavailable, "unavailable")}
	}
	return err
}

func (d *stubDoer) setDialer(grpcDialer, string) {}

func (d *stubDoer) setSecrets(func(string) ([]byte, error), []byte) {}
