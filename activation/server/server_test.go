package server

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	proto "github.com/edgelesssys/constellation/activation/activationproto"
	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/edgelesssys/constellation/internal/file"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	kubeadmv1 "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta3"
)

func TestActivateNode(t *testing.T) {
	someErr := errors.New("error")
	testKey := []byte{0x1, 0x2, 0x3}
	testID := id{
		Owner:   []byte{0x4, 0x5, 0x6},
		Cluster: []byte{0x7, 0x8, 0x9},
	}
	testJoinToken := &kubeadmv1.BootstrapTokenDiscovery{
		APIServerEndpoint: "192.0.2.1",
		CACertHashes:      []string{"hash"},
		Token:             "token",
	}

	testCases := map[string]struct {
		kubeadm stubTokenGetter
		kms     stubKeyGetter
		id      []byte
		wantErr bool
	}{
		"success": {
			kubeadm: stubTokenGetter{token: testJoinToken},
			kms:     stubKeyGetter{dataKey: testKey},
			id:      mustMarshalID(testID),
		},
		"GetDataKey fails": {
			kubeadm: stubTokenGetter{token: testJoinToken},
			kms:     stubKeyGetter{getDataKeyErr: someErr},
			id:      mustMarshalID(testID),
			wantErr: true,
		},
		"loading IDs fails": {
			kubeadm: stubTokenGetter{token: testJoinToken},
			kms:     stubKeyGetter{dataKey: testKey},
			id:      []byte{0x1, 0x2, 0x3},
			wantErr: true,
		},
		"no ID file": {
			kubeadm: stubTokenGetter{token: testJoinToken},
			kms:     stubKeyGetter{dataKey: testKey},
			wantErr: true,
		},
		"GetJoinToken fails": {
			kubeadm: stubTokenGetter{getJoinTokenErr: someErr},
			kms:     stubKeyGetter{dataKey: testKey},
			id:      mustMarshalID(testID),
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			file := file.NewHandler(afero.NewMemMapFs())
			if len(tc.id) > 0 {
				require.NoError(file.Write(constants.ActivationIDFilename, tc.id, 0o644))
			}
			api := New(file, tc.kubeadm, tc.kms)

			resp, err := api.ActivateNode(context.Background(), &proto.ActivateNodeRequest{DiskUuid: "uuid"})
			if tc.wantErr {
				assert.Error(err)
				return
			}

			var expectedIDs id
			require.NoError(json.Unmarshal(tc.id, &expectedIDs))

			require.NoError(err)
			assert.Equal(tc.kms.dataKey, resp.StateDiskKey)
			assert.Equal(expectedIDs.Cluster, resp.ClusterId)
			assert.Equal(expectedIDs.Owner, resp.OwnerId)
			assert.Equal(tc.kubeadm.token.APIServerEndpoint, resp.ApiServerEndpoint)
			assert.Equal(tc.kubeadm.token.CACertHashes[0], resp.DiscoveryTokenCaCertHash)
			assert.Equal(tc.kubeadm.token.Token, resp.Token)
		})
	}
}

func mustMarshalID(id id) []byte {
	b, err := json.Marshal(id)
	if err != nil {
		panic(err)
	}
	return b
}

type stubTokenGetter struct {
	token           *kubeadmv1.BootstrapTokenDiscovery
	getJoinTokenErr error
}

func (f stubTokenGetter) GetJoinToken(time.Duration) (*kubeadmv1.BootstrapTokenDiscovery, error) {
	return f.token, f.getJoinTokenErr
}

type stubKeyGetter struct {
	dataKey       []byte
	getDataKeyErr error
}

func (f stubKeyGetter) GetDataKey(context.Context, string, int) ([]byte, error) {
	return f.dataKey, f.getDataKeyErr
}
