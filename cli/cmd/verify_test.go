package cmd

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"testing"

	"github.com/edgelesssys/constellation/cli/cloudprovider"
	"github.com/edgelesssys/constellation/cli/file"
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
		"no args":              {[]string{}, true},
		"valid azure":          {[]string{"azure", "192.0.2.1", "1234"}, false},
		"valid gcp":            {[]string{"gcp", "192.0.2.1", "1234"}, false},
		"invalid provider":     {[]string{"invalid", "192.0.2.1", "1234"}, true},
		"invalid ip":           {[]string{"gcp", "invalid", "1234"}, true},
		"invalid port":         {[]string{"gcp", "192.0.2.1", "invalid"}, true},
		"invalid port 2":       {[]string{"gcp", "192.0.2.1", "65536"}, true},
		"not enough arguments": {[]string{"gcp", "192.0.2.1"}, true},
		"too many arguments":   {[]string{"gcp", "192.0.2.1", "1234", "5678"}, true},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			cmd := newVerifyCmd()
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
		setupFs       func(*require.Assertions) afero.Fs
		provider      cloudprovider.Provider
		protoClient   protoClient
		devConfigFlag string
		ownerIDFlag   string
		clusterIDFlag string
		wantErr       bool
	}{
		"gcp": {
			setupFs:     func(require *require.Assertions) afero.Fs { return afero.NewMemMapFs() },
			provider:    cloudprovider.GCP,
			ownerIDFlag: zeroBase64,
			protoClient: &stubProtoClient{},
		},
		"azure": {
			setupFs:     func(require *require.Assertions) afero.Fs { return afero.NewMemMapFs() },
			provider:    cloudprovider.Azure,
			ownerIDFlag: zeroBase64,
			protoClient: &stubProtoClient{},
		},
		"neither owner id nor cluster id set": {
			setupFs:  func(require *require.Assertions) afero.Fs { return afero.NewMemMapFs() },
			provider: cloudprovider.GCP,
			wantErr:  true,
		},
		"dev config file not existing": {
			setupFs:       func(require *require.Assertions) afero.Fs { return afero.NewMemMapFs() },
			provider:      cloudprovider.GCP,
			ownerIDFlag:   zeroBase64,
			devConfigFlag: "./file",
			wantErr:       true,
		},
		"error protoClient Connect": {
			setupFs:     func(require *require.Assertions) afero.Fs { return afero.NewMemMapFs() },
			provider:    cloudprovider.Azure,
			ownerIDFlag: zeroBase64,
			protoClient: &stubProtoClient{connectErr: someErr},
			wantErr:     true,
		},
		"error protoClient GetState": {
			setupFs:     func(require *require.Assertions) afero.Fs { return afero.NewMemMapFs() },
			provider:    cloudprovider.Azure,
			ownerIDFlag: zeroBase64,
			protoClient: &stubProtoClient{getStateErr: rpcStatus.Error(codes.Internal, "failed")},
			wantErr:     true,
		},
		"error protoClient GetState not rpc": {
			setupFs:     func(require *require.Assertions) afero.Fs { return afero.NewMemMapFs() },
			provider:    cloudprovider.Azure,
			ownerIDFlag: zeroBase64,
			protoClient: &stubProtoClient{getStateErr: someErr},
			wantErr:     true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			cmd := newVerifyCmd()
			cmd.Flags().String("dev-config", "", "") // register persisten flag manually
			out := bytes.NewBufferString("")
			cmd.SetOut(out)
			cmd.SetErr(bytes.NewBufferString(""))
			if tc.devConfigFlag != "" {
				require.NoError(cmd.Flags().Set("dev-config", tc.devConfigFlag))
			}
			if tc.ownerIDFlag != "" {
				require.NoError(cmd.Flags().Set("owner-id", tc.ownerIDFlag))
			}
			if tc.clusterIDFlag != "" {
				require.NoError(cmd.Flags().Set("cluster-id", tc.clusterIDFlag))
			}
			fileHandler := file.NewHandler(tc.setupFs(require))

			ctx := context.Background()
			err := verify(ctx, cmd, tc.provider, "192.0.2.1", "1234", fileHandler, tc.protoClient)

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
		"second arg": {
			args:        []string{"gcp"},
			toComplete:  "192.0.2.1",
			wantResult:  []string{},
			wantShellCD: cobra.ShellCompDirectiveNoFileComp,
		},
		"third arg": {
			args:        []string{"gcp", "192.0.2.1"},
			toComplete:  "443",
			wantResult:  []string{},
			wantShellCD: cobra.ShellCompDirectiveNoFileComp,
		},
		"additional arg": {
			args:        []string{"gcp", "192.0.2.1", "443"},
			toComplete:  "./file",
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
