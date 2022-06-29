package server

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/activation/activationproto"
	attestationtypes "github.com/edgelesssys/constellation/internal/attestation/types"
	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/edgelesssys/constellation/internal/file"
	"github.com/edgelesssys/constellation/internal/logger"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	kubeadmv1 "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta3"
)

func TestActivateNode(t *testing.T) {
	someErr := errors.New("error")
	testKey := []byte{0x1, 0x2, 0x3}
	testCert := []byte{0x4, 0x5, 0x6}
	testID := attestationtypes.ID{
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
		ca      stubCA
		id      []byte
		wantErr bool
	}{
		"success": {
			kubeadm: stubTokenGetter{
				token: testJoinToken,
			},
			kms: stubKeyGetter{
				dataKey: testKey,
			},
			ca: stubCA{
				cert: testCert,
				key:  testKey,
			},
			id: mustMarshalID(testID),
		},
		"GetDataKey fails": {
			kubeadm: stubTokenGetter{
				token: testJoinToken,
			},
			kms: stubKeyGetter{
				getDataKeyErr: someErr,
			},
			ca: stubCA{
				cert: testCert,
				key:  testKey,
			},
			id:      mustMarshalID(testID),
			wantErr: true,
		},
		"loading IDs fails": {
			kubeadm: stubTokenGetter{
				token: testJoinToken,
			},
			kms: stubKeyGetter{
				dataKey: testKey,
			},
			ca: stubCA{
				cert: testCert,
				key:  testKey,
			},
			id:      []byte{0x1, 0x2, 0x3},
			wantErr: true,
		},
		"no ID file": {
			kubeadm: stubTokenGetter{
				token: testJoinToken,
			},
			kms: stubKeyGetter{
				dataKey: testKey,
			},
			ca: stubCA{
				cert: testCert,
				key:  testKey,
			},
			wantErr: true,
		},
		"GetJoinToken fails": {
			kubeadm: stubTokenGetter{
				getJoinTokenErr: someErr,
			},
			kms: stubKeyGetter{
				dataKey: testKey,
			},
			ca: stubCA{
				cert: testCert,
				key:  testKey,
			},
			id:      mustMarshalID(testID),
			wantErr: true,
		},
		"GetCertificate fails": {
			kubeadm: stubTokenGetter{
				token: testJoinToken,
			},
			kms: stubKeyGetter{
				dataKey: testKey,
			},
			ca: stubCA{
				getCertErr: someErr,
			},
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
				require.NoError(file.Write(filepath.Join(constants.ServiceBasePath, constants.IDFilename), tc.id, 0o644))
			}
			api := New(
				logger.NewTest(t),
				file,
				tc.ca,
				tc.kubeadm,
				tc.kms,
			)

			resp, err := api.activateNode(context.Background(), "uuid", "test")
			if tc.wantErr {
				assert.Error(err)
				return
			}

			var expectedIDs attestationtypes.ID
			require.NoError(json.Unmarshal(tc.id, &expectedIDs))

			require.NoError(err)
			assert.Equal(tc.kms.dataKey, resp.stateDiskKey)
			assert.Equal(expectedIDs.Cluster, resp.id.Cluster)
			assert.Equal(expectedIDs.Owner, resp.id.Owner)
			assert.Equal(tc.kubeadm.token.APIServerEndpoint, resp.kubeArgs.APIServerEndpoint)
			assert.Equal(tc.kubeadm.token.CACertHashes[0], resp.kubeArgs.CACertHashes[0])
			assert.Equal(tc.kubeadm.token.Token, resp.kubeArgs.Token)
			assert.Equal(tc.ca.cert, resp.kubeletCert)
			assert.Equal(tc.ca.key, resp.kubeletKey)
		})
	}
}

func TestActivateWorkerNode(t *testing.T) {
	testKey := []byte{0x1, 0x2, 0x3}
	testCert := []byte{0x4, 0x5, 0x6}
	testID := attestationtypes.ID{
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
		ca      stubCA
		id      []byte
		wantErr bool
	}{
		"success": {
			kubeadm: stubTokenGetter{
				token: testJoinToken,
			},
			kms: stubKeyGetter{
				dataKey: testKey,
			},
			ca: stubCA{
				cert: testCert,
				key:  testKey,
			},
			id: mustMarshalID(testID),
		},
		"activateNode fails": {
			kubeadm: stubTokenGetter{
				token: testJoinToken,
			},
			kms: stubKeyGetter{
				getDataKeyErr: errors.New("error"),
			},
			ca: stubCA{
				cert: testCert,
				key:  testKey,
			},
			id:      mustMarshalID(testID),
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			file := file.NewHandler(afero.NewMemMapFs())
			require.NoError(file.Write(filepath.Join(constants.ServiceBasePath, constants.IDFilename), tc.id, 0o644))

			api := New(
				logger.NewTest(t),
				file,
				tc.ca,
				tc.kubeadm,
				tc.kms,
			)

			resp, err := api.ActivateWorkerNode(context.Background(), &activationproto.ActivateWorkerNodeRequest{DiskUuid: "uuid", NodeName: "test"})
			if tc.wantErr {
				assert.Error(err)
				return
			}

			var expectedIDs attestationtypes.ID
			require.NoError(json.Unmarshal(tc.id, &expectedIDs))

			require.NoError(err)
			assert.Equal(tc.kms.dataKey, resp.StateDiskKey)
			assert.Equal(expectedIDs.Cluster, resp.ClusterId)
			assert.Equal(expectedIDs.Owner, resp.OwnerId)
			assert.Equal(tc.kubeadm.token.APIServerEndpoint, resp.ApiServerEndpoint)
			assert.Equal(tc.kubeadm.token.CACertHashes[0], resp.DiscoveryTokenCaCertHash)
			assert.Equal(tc.kubeadm.token.Token, resp.Token)
			assert.Equal(tc.ca.cert, resp.KubeletCert)
			assert.Equal(tc.ca.key, resp.KubeletKey)
		})
	}
}

func TestActivateControlPlaneNode(t *testing.T) {
	someErr := errors.New("error")
	testKey := []byte{0x1, 0x2, 0x3}
	testCert := []byte{0x4, 0x5, 0x6}
	testID := attestationtypes.ID{
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
		ca      stubCA
		id      []byte
		wantErr bool
	}{
		"success": {
			kubeadm: stubTokenGetter{
				token:          testJoinToken,
				certificateKey: "test",
			},
			kms: stubKeyGetter{
				dataKey: testKey,
			},
			ca: stubCA{
				cert: testCert,
				key:  testKey,
			},
			id: mustMarshalID(testID),
		},
		"activateNode fails": {
			kubeadm: stubTokenGetter{
				token:          testJoinToken,
				certificateKey: "test",
			},
			kms: stubKeyGetter{
				getDataKeyErr: someErr,
			},
			ca: stubCA{
				cert: testCert,
				key:  testKey,
			},
			id:      mustMarshalID(testID),
			wantErr: true,
		},
		"GetControlPlaneCertificateKey fails": {
			kubeadm: stubTokenGetter{
				token:             testJoinToken,
				certificateKeyErr: someErr,
			},
			kms: stubKeyGetter{
				dataKey: testKey,
			},
			ca: stubCA{
				cert: testCert,
				key:  testKey,
			},
			id:      mustMarshalID(testID),
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			file := file.NewHandler(afero.NewMemMapFs())
			require.NoError(file.Write(filepath.Join(constants.ServiceBasePath, constants.IDFilename), tc.id, 0o644))

			api := New(
				logger.NewTest(t),
				file,
				tc.ca,
				tc.kubeadm,
				tc.kms,
			)

			resp, err := api.ActivateControlPlaneNode(context.Background(), &activationproto.ActivateControlPlaneNodeRequest{DiskUuid: "uuid", NodeName: "test"})
			if tc.wantErr {
				assert.Error(err)
				return
			}

			var expectedIDs attestationtypes.ID
			require.NoError(json.Unmarshal(tc.id, &expectedIDs))

			require.NoError(err)
			assert.Equal(tc.kms.dataKey, resp.StateDiskKey)
			assert.Equal(expectedIDs.Cluster, resp.ClusterId)
			assert.Equal(expectedIDs.Owner, resp.OwnerId)
			assert.Equal(tc.kubeadm.token.APIServerEndpoint, resp.ApiServerEndpoint)
			assert.Equal(tc.kubeadm.token.CACertHashes[0], resp.DiscoveryTokenCaCertHash)
			assert.Equal(tc.kubeadm.token.Token, resp.Token)
			assert.Equal(tc.ca.cert, resp.KubeletCert)
			assert.Equal(tc.ca.key, resp.KubeletKey)
			assert.Equal(tc.kubeadm.certificateKey, resp.CertificateKey)
		})
	}
}

func mustMarshalID(id attestationtypes.ID) []byte {
	b, err := json.Marshal(id)
	if err != nil {
		panic(err)
	}
	return b
}

type stubTokenGetter struct {
	token             *kubeadmv1.BootstrapTokenDiscovery
	getJoinTokenErr   error
	certificateKey    string
	certificateKeyErr error
}

func (f stubTokenGetter) GetJoinToken(time.Duration) (*kubeadmv1.BootstrapTokenDiscovery, error) {
	return f.token, f.getJoinTokenErr
}

func (f stubTokenGetter) GetControlPlaneCertificateKey() (string, error) {
	return f.certificateKey, f.certificateKeyErr
}

type stubKeyGetter struct {
	dataKey       []byte
	getDataKeyErr error
}

func (f stubKeyGetter) GetDataKey(context.Context, string, int) ([]byte, error) {
	return f.dataKey, f.getDataKeyErr
}

type stubCA struct {
	cert       []byte
	key        []byte
	getCertErr error
}

func (f stubCA) GetCertificate(string) ([]byte, []byte, error) {
	return f.cert, f.key, f.getCertErr
}
