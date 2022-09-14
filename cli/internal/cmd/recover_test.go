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

	"github.com/edgelesssys/constellation/cli/internal/cloudcmd"
	"github.com/edgelesssys/constellation/disk-mapper/recoverproto"
	"github.com/edgelesssys/constellation/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/internal/config"
	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/edgelesssys/constellation/internal/crypto/testvector"
	"github.com/edgelesssys/constellation/internal/file"
	"github.com/edgelesssys/constellation/internal/grpc/atlscredentials"
	"github.com/edgelesssys/constellation/internal/grpc/dialer"
	"github.com/edgelesssys/constellation/internal/grpc/testdialer"
	"github.com/edgelesssys/constellation/internal/state"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
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
	validState := state.ConstellationState{CloudProvider: "GCP"}
	invalidCSPState := state.ConstellationState{CloudProvider: "invalid"}
	successActions := []func(stream recoverproto.API_RecoverServer) error{
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
	}

	testCases := map[string]struct {
		existingState    state.ConstellationState
		recoverServerAPI *stubRecoveryServer
		masterSecret     testvector.HKDF
		endpointFlag     string
		masterSecretFlag string
		configFlag       string
		stateless        bool
		wantErr          bool
	}{
		"works": {
			existingState:    validState,
			recoverServerAPI: &stubRecoveryServer{actions: successActions},
			endpointFlag:     "192.0.2.1",
			masterSecret:     testvector.HKDFZero,
		},
		"missing flags": {
			recoverServerAPI: &stubRecoveryServer{actions: successActions},
			wantErr:          true,
		},
		"missing config": {
			recoverServerAPI: &stubRecoveryServer{actions: successActions},
			endpointFlag:     "192.0.2.1",
			masterSecret:     testvector.HKDFZero,
			configFlag:       "nonexistent-config",
			wantErr:          true,
		},
		"missing state": {
			existingState:    validState,
			recoverServerAPI: &stubRecoveryServer{actions: successActions},
			endpointFlag:     "192.0.2.1",
			masterSecret:     testvector.HKDFZero,
			stateless:        true,
			wantErr:          true,
		},
		"invalid cloud provider": {
			existingState:    invalidCSPState,
			recoverServerAPI: &stubRecoveryServer{actions: successActions},
			endpointFlag:     "192.0.2.1",
			masterSecret:     testvector.HKDFZero,
			wantErr:          true,
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
			cmd.SetErr(&bytes.Buffer{})
			if tc.endpointFlag != "" {
				require.NoError(cmd.Flags().Set("endpoint", tc.endpointFlag))
			}
			if tc.masterSecretFlag != "" {
				require.NoError(cmd.Flags().Set("master-secret", tc.masterSecretFlag))
			}
			if tc.configFlag != "" {
				require.NoError(cmd.Flags().Set("config", tc.configFlag))
			}

			fs := afero.NewMemMapFs()
			fileHandler := file.NewHandler(fs)

			config := defaultConfigWithExpectedMeasurements(t, config.Default(), cloudprovider.FromString(tc.existingState.CloudProvider))
			require.NoError(fileHandler.WriteYAML(constants.ConfigFilename, config))

			require.NoError(fileHandler.WriteJSON(
				"constellation-mastersecret.json",
				masterSecret{Key: tc.masterSecret.Secret, Salt: tc.masterSecret.Salt},
				file.OptNone,
			))

			if !tc.stateless {
				require.NoError(fileHandler.WriteJSON(
					constants.StateFilename,
					tc.existingState,
					file.OptNone,
				))
			}

			netDialer := testdialer.NewBufconnDialer()
			newDialer := func(*cloudcmd.Validator) *dialer.Dialer {
				return dialer.New(nil, nil, netDialer)
			}
			serverCreds := atlscredentials.New(nil, nil)
			recoverServer := grpc.NewServer(grpc.Creds(serverCreds))
			recoverproto.RegisterAPIServer(recoverServer, tc.recoverServerAPI)
			listener := netDialer.GetListener(net.JoinHostPort("192.0.2.1", strconv.Itoa(constants.RecoveryPort)))
			go recoverServer.Serve(listener)
			defer recoverServer.GracefulStop()

			err := recover(cmd, fileHandler, newDialer)

			if tc.wantErr {
				assert.Error(err)
				return
			}

			assert.NoError(err)
			assert.Contains(out.String(), "Pushed recovery key.")
		})
	}
}

func TestParseRecoverFlags(t *testing.T) {
	testCases := map[string]struct {
		args      []string
		wantFlags recoverFlags
		wantErr   bool
	}{
		"no flags": {
			wantErr: true,
		},
		"invalid ip": {
			args:    []string{"-e", "192.0.2.1:2:2"},
			wantErr: true,
		},
		"minimal args set": {
			args: []string{"-e", "192.0.2.1:2"},
			wantFlags: recoverFlags{
				endpoint:   "192.0.2.1:2",
				secretPath: "constellation-mastersecret.json",
			},
		},
		"all args set": {
			args: []string{"-e", "192.0.2.1:2", "--config", "config-path", "--master-secret", "/path/super-secret.json"},
			wantFlags: recoverFlags{
				endpoint:   "192.0.2.1:2",
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
			flags, err := parseRecoverFlags(cmd)

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
			recoveryServer: &stubRecoveryServer{actions: []func(stream recoverproto.API_RecoverServer) error{
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
		"error on first recv": {
			recoveryServer: &stubRecoveryServer{actions: []func(stream recoverproto.API_RecoverServer) error{
				func(stream recoverproto.API_RecoverServer) error {
					return someErr
				},
			}},
			wantErr: true,
		},
		"error on send": {
			recoveryServer: &stubRecoveryServer{actions: []func(stream recoverproto.API_RecoverServer) error{
				func(stream recoverproto.API_RecoverServer) error {
					_, err := stream.Recv()
					return err
				},
				func(stream recoverproto.API_RecoverServer) error {
					return someErr
				},
			}},
			wantErr: true,
		},
		"error on second recv": {
			recoveryServer: &stubRecoveryServer{actions: []func(stream recoverproto.API_RecoverServer) error{
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
			}},
			wantErr: true,
		},
		"final message is an error": {
			recoveryServer: &stubRecoveryServer{actions: []func(stream recoverproto.API_RecoverServer) error{
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
			addr := net.JoinHostPort("192.0.2.1", strconv.Itoa(constants.RecoveryPort))
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
	actions []func(recoverproto.API_RecoverServer) error
	recoverproto.UnimplementedAPIServer
}

func (s *stubRecoveryServer) Recover(stream recoverproto.API_RecoverServer) error {
	for _, action := range s.actions {
		if err := action(stream); err != nil {
			return err
		}
	}
	return nil
}
