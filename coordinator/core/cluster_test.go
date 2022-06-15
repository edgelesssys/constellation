package core

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/coordinator/pubapi/pubproto"
	"github.com/edgelesssys/constellation/coordinator/role"
	"github.com/edgelesssys/constellation/internal/attestation/simulator"
	attestationtypes "github.com/edgelesssys/constellation/internal/attestation/types"
	"github.com/edgelesssys/constellation/internal/deploy/user"
	"github.com/edgelesssys/constellation/internal/file"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	kubeadm "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta3"
)

func TestInitCluster(t *testing.T) {
	someErr := errors.New("someErr")
	kubeconfigContent := []byte("kubeconfig")

	testMS := []byte{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8}
	testSSHUsers := make([]*pubproto.SSHUserKey, 0)
	testSSHUser := &pubproto.SSHUserKey{
		Username:  "testUser",
		PublicKey: "ssh-rsa testKey",
	}
	testSSHUsers = append(testSSHUsers, testSSHUser)

	testCases := map[string]struct {
		cluster               Cluster
		vpn                   VPN
		metadata              ProviderMetadata
		masterSecret          []byte
		autoscalingNodeGroups []string
		sshUsers              []*pubproto.SSHUserKey
		wantErr               bool
	}{
		"InitCluster works": {
			cluster: &clusterStub{
				kubeconfig: kubeconfigContent,
			},
			vpn:                   &stubVPN{interfaceIP: "192.0.2.1"},
			metadata:              &stubMetadata{supportedRes: true},
			autoscalingNodeGroups: []string{"someNodeGroup"},
		},
		"InitCluster works even if signal role fails": {
			cluster: &clusterStub{
				kubeconfig: kubeconfigContent,
			},
			vpn:                   &stubVPN{interfaceIP: "192.0.2.1"},
			metadata:              &stubMetadata{supportedRes: true, signalRoleErr: someErr},
			autoscalingNodeGroups: []string{"someNodeGroup"},
		},
		"InitCluster works with SSH and KMS": {
			cluster: &clusterStub{
				kubeconfig: kubeconfigContent,
			},
			vpn:                   &stubVPN{interfaceIP: "192.0.2.1"},
			metadata:              &stubMetadata{supportedRes: true},
			autoscalingNodeGroups: []string{"someNodeGroup"},
			masterSecret:          testMS,
			sshUsers:              testSSHUsers,
		},
		"cannot get VPN IP": {
			cluster: &clusterStub{
				kubeconfig: kubeconfigContent,
			},
			vpn:                   &stubVPN{getInterfaceIPErr: someErr},
			autoscalingNodeGroups: []string{"someNodeGroup"},
			wantErr:               true,
		},
		"cannot init kubernetes": {
			cluster: &clusterStub{
				initErr: someErr,
			},
			vpn:                   &stubVPN{interfaceIP: "192.0.2.1"},
			metadata:              &stubMetadata{supportedRes: true},
			autoscalingNodeGroups: []string{"someNodeGroup"},
			wantErr:               true,
		},
		"cannot get kubeconfig": {
			cluster: &clusterStub{
				getKubeconfigErr: someErr,
			},
			vpn:                   &stubVPN{interfaceIP: "192.0.2.1"},
			metadata:              &stubMetadata{supportedRes: true},
			autoscalingNodeGroups: []string{"someNodeGroup"},
			wantErr:               true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			zapLogger, err := zap.NewDevelopment()
			require.NoError(err)
			fs := afero.NewMemMapFs()
			core, err := NewCore(tc.vpn, tc.cluster, tc.metadata, nil, zapLogger, simulator.OpenSimulatedTPM, nil, file.NewHandler(fs), user.NewLinuxUserManagerFake(fs))
			require.NoError(err)

			id := attestationtypes.ID{Owner: []byte{0x1}, Cluster: []byte{0x2}}
			kubeconfig, err := core.InitCluster(context.Background(), tc.autoscalingNodeGroups, "cloud-service-account-uri", id, tc.masterSecret, tc.sshUsers)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(kubeconfigContent, kubeconfig)
		})
	}
}

func TestJoinCluster(t *testing.T) {
	someErr := errors.New("someErr")

	testCases := map[string]struct {
		cluster  Cluster
		metadata ProviderMetadata
		vpn      VPN
		wantErr  bool
	}{
		"JoinCluster works": {
			vpn: &stubVPN{
				interfaceIP: "192.0.2.0",
			},
			cluster:  &clusterStub{},
			metadata: &stubMetadata{supportedRes: true},
		},
		"JoinCluster works even if signal role fails": {
			vpn: &stubVPN{
				interfaceIP: "192.0.2.0",
			},
			cluster:  &clusterStub{},
			metadata: &stubMetadata{supportedRes: true, signalRoleErr: someErr},
		},
		"cannot get VPN IP": {
			vpn:      &stubVPN{getInterfaceIPErr: someErr},
			cluster:  &clusterStub{},
			metadata: &stubMetadata{supportedRes: true},
			wantErr:  true,
		},
		"joining kuberentes fails": {
			vpn: &stubVPN{
				interfaceIP: "192.0.2.0",
			},
			cluster:  &clusterStub{joinErr: someErr},
			metadata: &stubMetadata{supportedRes: true},
			wantErr:  true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			zapLogger, err := zap.NewDevelopment()
			require.NoError(err)
			fs := afero.NewMemMapFs()
			core, err := NewCore(tc.vpn, tc.cluster, tc.metadata, nil, zapLogger, simulator.OpenSimulatedTPM, nil, file.NewHandler(fs), user.NewLinuxUserManagerFake(fs))
			require.NoError(err)

			joinReq := &kubeadm.BootstrapTokenDiscovery{
				APIServerEndpoint: "192.0.2.0:6443",
				Token:             "someToken",
				CACertHashes:      []string{"someHash"},
			}
			err = core.JoinCluster(context.Background(), joinReq, "", role.Node)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
		})
	}
}

type clusterStub struct {
	initErr              error
	joinErr              error
	kubeconfig           []byte
	getKubeconfigErr     error
	getJoinTokenResponse *kubeadm.BootstrapTokenDiscovery
	getJoinTokenErr      error
	startKubeletErr      error

	inAutoscalingNodeGroups  []string
	inCloudServiceAccountURI string
	inVpnIP                  string
}

func (c *clusterStub) InitCluster(
	ctx context.Context, autoscalingNodeGroups []string, cloudServiceAccountURI string, vpnIP string, id attestationtypes.ID, masterSecret []byte, sshUsers map[string]string,
) error {
	c.inAutoscalingNodeGroups = autoscalingNodeGroups
	c.inCloudServiceAccountURI = cloudServiceAccountURI
	c.inVpnIP = vpnIP

	return c.initErr
}

func (c *clusterStub) JoinCluster(ctx context.Context, args *kubeadm.BootstrapTokenDiscovery, nodeVPNIP string, certKey string, peerRole role.Role) error {
	return c.joinErr
}

func (c *clusterStub) GetKubeconfig() ([]byte, error) {
	return c.kubeconfig, c.getKubeconfigErr
}

func (c *clusterStub) GetKubeadmCertificateKey(context.Context) (string, error) {
	return "dummy", nil
}

func (c *clusterStub) GetJoinToken(ctx context.Context, ttl time.Duration) (*kubeadm.BootstrapTokenDiscovery, error) {
	return c.getJoinTokenResponse, c.getJoinTokenErr
}

func (c *clusterStub) StartKubelet() error {
	return c.startKubeletErr
}
