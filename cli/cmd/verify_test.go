package cmd

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"testing"

	"github.com/edgelesssys/constellation/cli/status"
	"github.com/edgelesssys/constellation/coordinator/atls"
	"github.com/edgelesssys/constellation/coordinator/attestation/gcp"
	"github.com/edgelesssys/constellation/coordinator/attestation/vtpm"
	"github.com/edgelesssys/constellation/coordinator/pubapi/pubproto"
	"github.com/edgelesssys/constellation/coordinator/state"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	rpcStatus "google.golang.org/grpc/status"
)

func TestVerify(t *testing.T) {
	testCases := map[string]struct {
		connErr  error
		checkErr error
		state    state.State
		wantErr  bool
	}{
		"connection error": {
			connErr:  errors.New("connection error"),
			checkErr: nil,
			state:    0,
			wantErr:  true,
		},
		"check error": {
			connErr:  nil,
			checkErr: errors.New("check error"),
			state:    0,
			wantErr:  true,
		},
		"check error, rpc status": {
			connErr:  nil,
			checkErr: rpcStatus.Error(codes.Unavailable, "check error"),
			state:    0,
			wantErr:  true,
		},
		"verify on worker node": {
			connErr:  nil,
			checkErr: nil,
			state:    state.IsNode,
			wantErr:  false,
		},
		"verify on master node": {
			connErr:  nil,
			checkErr: nil,
			state:    state.ActivatingNodes,
			wantErr:  false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			ctx := context.Background()
			var out bytes.Buffer

			verifier := verifier{
				newConn: stubNewConnFunc(tc.connErr),
				newClient: stubNewClientFunc(&stubPeerStatusClient{
					state:    tc.state,
					checkErr: tc.checkErr,
				}),
			}

			pcrs := map[uint32][]byte{
				0: []byte("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"),
				1: []byte("BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB"),
			}
			err := verify(ctx, &out, "", []atls.Validator{gcp.NewValidator(pcrs)}, verifier)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Contains(out.String(), "OK")
			}
		})
	}
}

func stubNewConnFunc(errStub error) func(ctx context.Context, target string, validators []atls.Validator) (status.ClientConn, error) {
	return func(ctx context.Context, target string, validators []atls.Validator) (status.ClientConn, error) {
		return &stubClientConn{}, errStub
	}
}

type stubClientConn struct{}

func (c *stubClientConn) Invoke(ctx context.Context, method string, args, reply interface{}, opts ...grpc.CallOption) error {
	return nil
}

func (c *stubClientConn) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	return nil, nil
}

func (c *stubClientConn) Close() error {
	return nil
}

func stubNewClientFunc(stubClient pubproto.APIClient) func(cc grpc.ClientConnInterface) pubproto.APIClient {
	return func(cc grpc.ClientConnInterface) pubproto.APIClient {
		return stubClient
	}
}

type stubPeerStatusClient struct {
	state    state.State
	checkErr error
	pubproto.APIClient
}

func (c *stubPeerStatusClient) GetState(ctx context.Context, in *pubproto.GetStateRequest, opts ...grpc.CallOption) (*pubproto.GetStateResponse, error) {
	resp := &pubproto.GetStateResponse{State: uint32(c.state)}
	return resp, c.checkErr
}

func TestPrepareValidator(t *testing.T) {
	testCases := map[string]struct {
		ownerID   string
		clusterID string
		wantErr   bool
	}{
		"no input": {
			ownerID:   "",
			clusterID: "",
			wantErr:   true,
		},
		"unencoded secret ID": {
			ownerID:   "owner-id",
			clusterID: base64.StdEncoding.EncodeToString([]byte("unique-id")),
			wantErr:   true,
		},
		"unencoded cluster ID": {
			ownerID:   base64.StdEncoding.EncodeToString([]byte("owner-id")),
			clusterID: "unique-id",
			wantErr:   true,
		},
		"correct input": {
			ownerID:   base64.StdEncoding.EncodeToString([]byte("owner-id")),
			clusterID: base64.StdEncoding.EncodeToString([]byte("unique-id")),
			wantErr:   false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			cmd := newVerifyCmd()
			cmd.Flags().String("owner-id", "", "")
			cmd.Flags().String("unique-id", "", "")
			require.NoError(cmd.Flags().Set("owner-id", tc.ownerID))
			require.NoError(cmd.Flags().Set("unique-id", tc.clusterID))
			var out bytes.Buffer
			cmd.SetOut(&out)
			var errOut bytes.Buffer
			cmd.SetErr(&errOut)

			pcrs := map[uint32][]byte{
				0: []byte("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"),
				1: []byte("BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB"),
			}

			err := prepareValidator(cmd, pcrs)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				if tc.clusterID != "" {
					assert.Len(pcrs[uint32(vtpm.PCRIndexClusterID)], 32)
				} else {
					assert.Nil(pcrs[uint32(vtpm.PCRIndexClusterID)])
				}
				if tc.ownerID != "" {
					assert.Len(pcrs[uint32(vtpm.PCRIndexOwnerID)], 32)
				} else {
					assert.Nil(pcrs[uint32(vtpm.PCRIndexOwnerID)])
				}
			}
		})
	}
}

func TestAddOrSkipPcr(t *testing.T) {
	emptyMap := map[uint32][]byte{}
	defaultMap := map[uint32][]byte{
		0: []byte("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"),
		1: []byte("BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB"),
	}

	testCases := map[string]struct {
		pcrMap      map[uint32][]byte
		pcrIndex    uint32
		encoded     string
		wantEntries int
		wantErr     bool
	}{
		"empty input, empty map": {
			pcrMap:      emptyMap,
			pcrIndex:    10,
			encoded:     "",
			wantEntries: 0,
			wantErr:     false,
		},
		"empty input, default map": {
			pcrMap:      defaultMap,
			pcrIndex:    10,
			encoded:     "",
			wantEntries: len(defaultMap),
			wantErr:     false,
		},
		"correct input, empty map": {
			pcrMap:      emptyMap,
			pcrIndex:    10,
			encoded:     base64.StdEncoding.EncodeToString([]byte("Constellation")),
			wantEntries: 1,
			wantErr:     false,
		},
		"correct input, default map": {
			pcrMap:      defaultMap,
			pcrIndex:    10,
			encoded:     base64.StdEncoding.EncodeToString([]byte("Constellation")),
			wantEntries: len(defaultMap) + 1,
			wantErr:     false,
		},
		"unencoded input, empty map": {
			pcrMap:      emptyMap,
			pcrIndex:    10,
			encoded:     "Constellation",
			wantEntries: 0,
			wantErr:     true,
		},
		"unencoded input, default map": {
			pcrMap:      defaultMap,
			pcrIndex:    10,
			encoded:     "Constellation",
			wantEntries: len(defaultMap),
			wantErr:     true,
		},
		"empty input at occupied index": {
			pcrMap:      defaultMap,
			pcrIndex:    0,
			encoded:     "",
			wantEntries: len(defaultMap) - 1,
			wantErr:     false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			res := make(map[uint32][]byte)
			for k, v := range tc.pcrMap {
				res[k] = v
			}

			err := addOrSkipPCR(res, tc.pcrIndex, tc.encoded)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
			assert.Len(res, tc.wantEntries)
			for _, v := range res {
				assert.Len(v, 32)
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
			toComplete:  "192.0.2.1",
			wantResult:  []string{},
			wantShellCD: cobra.ShellCompDirectiveNoFileComp,
		},
		"second arg": {
			args:        []string{"192.0.2.1"},
			toComplete:  "443",
			wantResult:  []string{},
			wantShellCD: cobra.ShellCompDirectiveNoFileComp,
		},
		"third arg": {
			args:        []string{"192.0.2.1", "443"},
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
