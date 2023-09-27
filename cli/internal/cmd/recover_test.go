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

	"github.com/edgelesssys/constellation/v2/cli/internal/clusterid"
	"github.com/edgelesssys/constellation/v2/disk-mapper/recoverproto"
	"github.com/edgelesssys/constellation/v2/internal/atls"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/crypto"
	"github.com/edgelesssys/constellation/v2/internal/crypto/testvector"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/grpc/atlscredentials"
	"github.com/edgelesssys/constellation/v2/internal/grpc/dialer"
	"github.com/edgelesssys/constellation/v2/internal/grpc/testdialer"
	"github.com/edgelesssys/constellation/v2/internal/kms/uri"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
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
	unavailableErr := grpcstatus.Error(codes.Unavailable, "unavailable")
	lbErr := grpcstatus.Error(codes.Unavailable, `connection error: desc = "transport: authentication handshake failed: read tcp`)

	testCases := map[string]struct {
		doer               *stubDoer
		masterSecret       testvector.HKDF
		endpoint           string
		successfulCalls    int
		skipConfigCreation bool
		wantErr            bool
	}{
		"works": {
			doer:            &stubDoer{returns: []error{nil}},
			endpoint:        "192.0.2.90",
			masterSecret:    testvector.HKDFZero,
			successfulCalls: 1,
		},
		"missing config": {
			doer:               &stubDoer{returns: []error{nil}},
			endpoint:           "192.0.2.89",
			masterSecret:       testvector.HKDFZero,
			skipConfigCreation: true,
			wantErr:            true,
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
			cmd.Flags().String("workspace", "", "") // register persistent flag manually
			cmd.Flags().Bool("force", true, "")     // register persistent flag manually
			out := &bytes.Buffer{}
			cmd.SetOut(out)
			cmd.SetErr(out)
			require.NoError(cmd.Flags().Set("endpoint", tc.endpoint))

			fs := afero.NewMemMapFs()
			fileHandler := file.NewHandler(fs)

			if !tc.skipConfigCreation {
				config := defaultConfigWithExpectedMeasurements(t, config.Default(), cloudprovider.GCP)
				require.NoError(fileHandler.WriteYAML(constants.ConfigFilename, config))
			}

			require.NoError(fileHandler.WriteJSON(
				"constellation-mastersecret.json",
				uri.MasterSecret{Key: tc.masterSecret.Secret, Salt: tc.masterSecret.Salt},
				file.OptNone,
			))

			newDialer := func(atls.Validator) *dialer.Dialer { return nil }
			r := &recoverCmd{log: logger.NewTest(t), configFetcher: stubAttestationFetcher{}}
			err := r.recover(cmd, fileHandler, time.Millisecond, tc.doer, newDialer)
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
				endpoint: "192.0.2.42:9999",
			},
			writeIDFile: true,
		},
		"no flags, no ID file": {
			wantFlags: recoverFlags{
				endpoint: "192.0.2.42:9999",
			},
			wantErr: true,
		},
		"invalid endpoint": {
			args:    []string{"-e", "192.0.2.42:2:2"},
			wantErr: true,
		},
		"all args set": {
			args: []string{"-e", "192.0.2.42:2", "--workspace", "./constellation-workspace"},
			wantFlags: recoverFlags{
				endpoint: "192.0.2.42:2",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			cmd := NewRecoverCmd()
			cmd.Flags().String("workspace", "", "") // register persistent flag manually
			cmd.Flags().Bool("force", false, "")    // register persistent flag manually
			require.NoError(cmd.ParseFlags(tc.args))

			fileHandler := file.NewHandler(afero.NewMemMapFs())
			if tc.writeIDFile {
				require.NoError(fileHandler.WriteJSON(constants.ClusterIDsFilename, &clusterid.File{IP: "192.0.2.42"}))
			}
			r := &recoverCmd{log: logger.NewTest(t)}
			flags, err := r.parseRecoverFlags(cmd, fileHandler)

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
	testCases := map[string]struct {
		recoveryServer *stubRecoveryServer
		wantErr        bool
	}{
		"success": {
			recoveryServer: &stubRecoveryServer{},
		},
		"server responds with error": {
			recoveryServer: &stubRecoveryServer{recoverError: errors.New("someErr")},
			wantErr:        true,
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

			r := &recoverCmd{log: logger.NewTest(t)}
			recoverDoer := &recoverDoer{
				dialer:   dialer.New(nil, nil, netDialer),
				endpoint: addr,
				log:      r.log,
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

func getStateDiskKeyFunc(masterKey, salt []byte) func(uuid string) ([]byte, error) {
	return func(uuid string) ([]byte, error) {
		return crypto.DeriveKey(masterKey, salt, []byte(crypto.DEKPrefix+uuid), crypto.StateDiskKeyLength)
	}
}

type stubRecoveryServer struct {
	recoverError error
	recoverproto.UnimplementedAPIServer
}

func (s *stubRecoveryServer) Recover(context.Context, *recoverproto.RecoverMessage) (*recoverproto.RecoverResponse, error) {
	if s.recoverError != nil {
		return nil, s.recoverError
	}
	return &recoverproto.RecoverResponse{}, nil
}

type stubDoer struct {
	returns []error
}

func (d *stubDoer) Do(context.Context) error {
	err := d.returns[0]
	if len(d.returns) > 1 {
		d.returns = d.returns[1:]
	} else {
		d.returns = []error{grpcstatus.Error(codes.Unavailable, "unavailable")}
	}
	return err
}

func (d *stubDoer) setDialer(grpcDialer, string) {}

func (d *stubDoer) setURIs(_, _ string) {}
