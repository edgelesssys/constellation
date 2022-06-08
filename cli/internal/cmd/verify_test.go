package cmd

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"testing"

	"github.com/edgelesssys/constellation/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/internal/file"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	rpcStatus "google.golang.org/grpc/status"
)

func TestVerifyCmdArgumentValidation(t *testing.T) {
	testCases := map[string]struct {
		args    []string
		wantErr bool
	}{
		"no args":          {[]string{}, true},
		"valid azure":      {[]string{"azure"}, false},
		"valid gcp":        {[]string{"gcp"}, false},
		"invalid provider": {[]string{"invalid", "192.0.2.1", "1234"}, true},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			cmd := NewVerifyCmd()
			err := cmd.ValidateArgs(tc.args)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

func TestVerify(t *testing.T) {
	zeroBase64 := base64.StdEncoding.EncodeToString([]byte("00000000000000000000000000000000"))
	someErr := errors.New("failed")

	testCases := map[string]struct {
		setupFs          func(*require.Assertions) afero.Fs
		provider         cloudprovider.Provider
		protoClient      protoClient
		nodeEndpointFlag string
		configFlag       string
		ownerIDFlag      string
		clusterIDFlag    string
		wantErr          bool
	}{
		"gcp": {
			setupFs:          func(require *require.Assertions) afero.Fs { return afero.NewMemMapFs() },
			provider:         cloudprovider.GCP,
			nodeEndpointFlag: "192.0.2.1:1234",
			ownerIDFlag:      zeroBase64,
			protoClient:      &stubProtoClient{},
		},
		"azure": {
			setupFs:          func(require *require.Assertions) afero.Fs { return afero.NewMemMapFs() },
			provider:         cloudprovider.Azure,
			nodeEndpointFlag: "192.0.2.1:1234",
			ownerIDFlag:      zeroBase64,
			protoClient:      &stubProtoClient{},
		},
		"default port": {
			setupFs:          func(require *require.Assertions) afero.Fs { return afero.NewMemMapFs() },
			provider:         cloudprovider.GCP,
			nodeEndpointFlag: "192.0.2.1",
			ownerIDFlag:      zeroBase64,
			protoClient:      &stubProtoClient{},
		},
		"invalid endpoint": {
			setupFs:          func(require *require.Assertions) afero.Fs { return afero.NewMemMapFs() },
			provider:         cloudprovider.GCP,
			nodeEndpointFlag: ":::::",
			ownerIDFlag:      zeroBase64,
			protoClient:      &stubProtoClient{},
			wantErr:          true,
		},
		"neither owner id nor cluster id set": {
			setupFs:          func(require *require.Assertions) afero.Fs { return afero.NewMemMapFs() },
			provider:         cloudprovider.GCP,
			nodeEndpointFlag: "192.0.2.1:1234",
			wantErr:          true,
		},
		"config file not existing": {
			setupFs:          func(require *require.Assertions) afero.Fs { return afero.NewMemMapFs() },
			provider:         cloudprovider.GCP,
			ownerIDFlag:      zeroBase64,
			nodeEndpointFlag: "192.0.2.1:1234",
			configFlag:       "./file",
			wantErr:          true,
		},
		"error protoClient Connect": {
			setupFs:          func(require *require.Assertions) afero.Fs { return afero.NewMemMapFs() },
			provider:         cloudprovider.Azure,
			nodeEndpointFlag: "192.0.2.1:1234",
			ownerIDFlag:      zeroBase64,
			protoClient:      &stubProtoClient{connectErr: someErr},
			wantErr:          true,
		},
		"error protoClient GetState": {
			setupFs:          func(require *require.Assertions) afero.Fs { return afero.NewMemMapFs() },
			provider:         cloudprovider.Azure,
			nodeEndpointFlag: "192.0.2.1:1234",
			ownerIDFlag:      zeroBase64,
			protoClient:      &stubProtoClient{getStateErr: rpcStatus.Error(codes.Internal, "failed")},
			wantErr:          true,
		},
		"error protoClient GetState not rpc": {
			setupFs:          func(require *require.Assertions) afero.Fs { return afero.NewMemMapFs() },
			provider:         cloudprovider.Azure,
			nodeEndpointFlag: "192.0.2.1:1234",
			ownerIDFlag:      zeroBase64,
			protoClient:      &stubProtoClient{getStateErr: someErr},
			wantErr:          true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			cmd := NewVerifyCmd()
			cmd.Flags().String("config", "", "") // register persisten flag manually
			out := &bytes.Buffer{}
			cmd.SetOut(out)
			cmd.SetErr(&bytes.Buffer{})
			if tc.configFlag != "" {
				require.NoError(cmd.Flags().Set("config", tc.configFlag))
			}
			if tc.ownerIDFlag != "" {
				require.NoError(cmd.Flags().Set("owner-id", tc.ownerIDFlag))
			}
			if tc.clusterIDFlag != "" {
				require.NoError(cmd.Flags().Set("cluster-id", tc.clusterIDFlag))
			}
			if tc.nodeEndpointFlag != "" {
				require.NoError(cmd.Flags().Set("node-endpoint", tc.nodeEndpointFlag))
			}
			fileHandler := file.NewHandler(tc.setupFs(require))

			ctx := context.Background()
			err := verify(ctx, cmd, tc.provider, fileHandler, tc.protoClient)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Contains(out.String(), "OK")
			}
		})
	}
}

func TestVerifyCompletion(t *testing.T) {
	testCases := map[string]struct {
		args        []string
		toComplete  string
		wantResult  []string
		wantShellCD cobra.ShellCompDirective
	}{
		"first arg": {
			args:        []string{},
			toComplete:  "az",
			wantResult:  []string{"gcp", "azure"},
			wantShellCD: cobra.ShellCompDirectiveNoFileComp,
		},
		"additional arg": {
			args:        []string{"gcp", "foo"},
			wantResult:  []string{},
			wantShellCD: cobra.ShellCompDirectiveError,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			cmd := &cobra.Command{}
			result, shellCD := verifyCompletion(cmd, tc.args, tc.toComplete)
			assert.Equal(tc.wantResult, result)
			assert.Equal(tc.wantShellCD, shellCD)
		})
	}
}
