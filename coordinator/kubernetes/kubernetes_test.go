package kubernetes

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/coordinator/cloudprovider/cloudtypes"
	"github.com/edgelesssys/constellation/coordinator/kubernetes/k8sapi"
	"github.com/edgelesssys/constellation/coordinator/kubernetes/k8sapi/resources"
	"github.com/edgelesssys/constellation/coordinator/role"
	attestationtypes "github.com/edgelesssys/constellation/internal/attestation/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
	kubeadm "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta3"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestInitCluster(t *testing.T) {
	someErr := errors.New("failed")
	coordinatorVPNIP := "192.0.2.0"
	serviceAccountUri := "some-service-account-uri"
	masterSecret := []byte("some-master-secret")
	autoscalingNodeGroups := []string{"0,10,autoscaling_group_0"}

	nodeName := "node-name"
	providerID := "provider-id"
	privateIP := "192.0.2.1"
	publicIP := "192.0.2.2"
	loadbalancerIP := "192.0.2.3"
	aliasIPRange := "192.0.2.0/24"
	k8sVersion := "1.23.8"

	testCases := map[string]struct {
		clusterUtil            stubClusterUtil
		kubeCTL                stubKubeCTL
		providerMetadata       ProviderMetadata
		CloudControllerManager CloudControllerManager
		CloudNodeManager       CloudNodeManager
		ClusterAutoscaler      ClusterAutoscaler
		kubeconfigReader       configReader
		wantConfig             k8sapi.KubeadmInitYAML
		wantErr                bool
	}{
		"kubeadm init works without metadata": {
			clusterUtil: stubClusterUtil{},
			kubeconfigReader: &stubKubeconfigReader{
				Kubeconfig: []byte("someKubeconfig"),
			},
			providerMetadata:       &stubProviderMetadata{SupportedResp: false},
			CloudControllerManager: &stubCloudControllerManager{},
			CloudNodeManager:       &stubCloudNodeManager{SupportedResp: false},
			ClusterAutoscaler:      &stubClusterAutoscaler{},
			wantConfig: k8sapi.KubeadmInitYAML{
				InitConfiguration: kubeadm.InitConfiguration{
					NodeRegistration: kubeadm.NodeRegistrationOptions{
						KubeletExtraArgs: map[string]string{
							"node-ip":     "",
							"provider-id": "",
						},
						Name: coordinatorVPNIP,
					},
				},
				ClusterConfiguration: kubeadm.ClusterConfiguration{},
			},
		},
		"kubeadm init works with metadata and loadbalancer": {
			clusterUtil: stubClusterUtil{},
			kubeconfigReader: &stubKubeconfigReader{
				Kubeconfig: []byte("someKubeconfig"),
			},
			providerMetadata: &stubProviderMetadata{
				SupportedResp: true,
				SelfResp: cloudtypes.Instance{
					Name:          nodeName,
					ProviderID:    providerID,
					PrivateIPs:    []string{privateIP},
					PublicIPs:     []string{publicIP},
					AliasIPRanges: []string{aliasIPRange},
				},
				GetLoadBalancerIPResp:    loadbalancerIP,
				SupportsLoadBalancerResp: true,
			},
			CloudControllerManager: &stubCloudControllerManager{},
			CloudNodeManager:       &stubCloudNodeManager{SupportedResp: false},
			ClusterAutoscaler:      &stubClusterAutoscaler{},
			wantConfig: k8sapi.KubeadmInitYAML{
				InitConfiguration: kubeadm.InitConfiguration{
					NodeRegistration: kubeadm.NodeRegistrationOptions{
						KubeletExtraArgs: map[string]string{
							"node-ip":     privateIP,
							"provider-id": providerID,
						},
						Name: nodeName,
					},
				},
				ClusterConfiguration: kubeadm.ClusterConfiguration{
					ControlPlaneEndpoint: loadbalancerIP,
					APIServer: kubeadm.APIServer{
						CertSANs: []string{publicIP, privateIP},
					},
				},
			},
			wantErr: false,
		},
		"kubeadm init fails when retrieving metadata self": {
			clusterUtil: stubClusterUtil{},
			kubeconfigReader: &stubKubeconfigReader{
				Kubeconfig: []byte("someKubeconfig"),
			},
			providerMetadata: &stubProviderMetadata{
				SelfErr:       someErr,
				SupportedResp: true,
			},
			CloudControllerManager: &stubCloudControllerManager{},
			CloudNodeManager:       &stubCloudNodeManager{},
			ClusterAutoscaler:      &stubClusterAutoscaler{},
			wantErr:                true,
		},
		"kubeadm init fails when retrieving metadata subnetwork cidr": {
			clusterUtil: stubClusterUtil{},
			kubeconfigReader: &stubKubeconfigReader{
				Kubeconfig: []byte("someKubeconfig"),
			},
			providerMetadata: &stubProviderMetadata{
				GetSubnetworkCIDRErr: someErr,
				SupportedResp:        true,
			},
			CloudControllerManager: &stubCloudControllerManager{},
			CloudNodeManager:       &stubCloudNodeManager{},
			ClusterAutoscaler:      &stubClusterAutoscaler{},
			wantErr:                true,
		},
		"kubeadm init fails when retrieving metadata loadbalancer ip": {
			clusterUtil: stubClusterUtil{},
			kubeconfigReader: &stubKubeconfigReader{
				Kubeconfig: []byte("someKubeconfig"),
			},
			providerMetadata: &stubProviderMetadata{
				GetLoadBalancerIPErr:     someErr,
				SupportsLoadBalancerResp: true,
				SupportedResp:            true,
			},
			CloudControllerManager: &stubCloudControllerManager{},
			CloudNodeManager:       &stubCloudNodeManager{},
			ClusterAutoscaler:      &stubClusterAutoscaler{},
			wantErr:                true,
		},
		"kubeadm init fails when applying the init config": {
			clusterUtil: stubClusterUtil{initClusterErr: someErr},
			kubeconfigReader: &stubKubeconfigReader{
				Kubeconfig: []byte("someKubeconfig"),
			},
			providerMetadata:       &stubProviderMetadata{},
			CloudControllerManager: &stubCloudControllerManager{},
			CloudNodeManager:       &stubCloudNodeManager{},
			ClusterAutoscaler:      &stubClusterAutoscaler{},
			wantErr:                true,
		},
		"kubeadm init fails when setting up the pod network": {
			clusterUtil: stubClusterUtil{setupPodNetworkErr: someErr},
			kubeconfigReader: &stubKubeconfigReader{
				Kubeconfig: []byte("someKubeconfig"),
			},
			providerMetadata:       &stubProviderMetadata{},
			CloudControllerManager: &stubCloudControllerManager{},
			CloudNodeManager:       &stubCloudNodeManager{},
			ClusterAutoscaler:      &stubClusterAutoscaler{},
			wantErr:                true,
		},
		"kubeadm init fails when setting up the activation service": {
			clusterUtil: stubClusterUtil{setupActivationServiceError: someErr},
			kubeconfigReader: &stubKubeconfigReader{
				Kubeconfig: []byte("someKubeconfig"),
			},
			providerMetadata:       &stubProviderMetadata{},
			CloudControllerManager: &stubCloudControllerManager{SupportedResp: true},
			CloudNodeManager:       &stubCloudNodeManager{},
			ClusterAutoscaler:      &stubClusterAutoscaler{},
			wantErr:                true,
		},
		"kubeadm init fails when setting the cloud contoller manager": {
			clusterUtil: stubClusterUtil{setupCloudControllerManagerError: someErr},
			kubeconfigReader: &stubKubeconfigReader{
				Kubeconfig: []byte("someKubeconfig"),
			},
			providerMetadata:       &stubProviderMetadata{},
			CloudControllerManager: &stubCloudControllerManager{SupportedResp: true},
			CloudNodeManager:       &stubCloudNodeManager{},
			ClusterAutoscaler:      &stubClusterAutoscaler{},
			wantErr:                true,
		},
		"kubeadm init fails when setting the cloud node manager": {
			clusterUtil: stubClusterUtil{setupCloudNodeManagerError: someErr},
			kubeconfigReader: &stubKubeconfigReader{
				Kubeconfig: []byte("someKubeconfig"),
			},
			providerMetadata:       &stubProviderMetadata{},
			CloudControllerManager: &stubCloudControllerManager{},
			CloudNodeManager:       &stubCloudNodeManager{SupportedResp: true},
			ClusterAutoscaler:      &stubClusterAutoscaler{},
			wantErr:                true,
		},
		"kubeadm init fails when setting the cluster autoscaler": {
			clusterUtil: stubClusterUtil{setupAutoscalingError: someErr},
			kubeconfigReader: &stubKubeconfigReader{
				Kubeconfig: []byte("someKubeconfig"),
			},
			providerMetadata:       &stubProviderMetadata{},
			CloudControllerManager: &stubCloudControllerManager{},
			CloudNodeManager:       &stubCloudNodeManager{},
			ClusterAutoscaler:      &stubClusterAutoscaler{SupportedResp: true},
			wantErr:                true,
		},
		"kubeadm init fails when reading kubeconfig": {
			clusterUtil: stubClusterUtil{},
			kubeconfigReader: &stubKubeconfigReader{
				ReadErr: someErr,
			},
			providerMetadata:       &stubProviderMetadata{},
			CloudControllerManager: &stubCloudControllerManager{},
			CloudNodeManager:       &stubCloudNodeManager{},
			ClusterAutoscaler:      &stubClusterAutoscaler{},
			wantErr:                true,
		},
		"kubeadm init fails when setting up the kms": {
			clusterUtil: stubClusterUtil{setupKMSError: someErr},
			kubeconfigReader: &stubKubeconfigReader{
				Kubeconfig: []byte("someKubeconfig"),
			},
			providerMetadata:       &stubProviderMetadata{SupportedResp: false},
			CloudControllerManager: &stubCloudControllerManager{},
			CloudNodeManager:       &stubCloudNodeManager{SupportedResp: false},
			ClusterAutoscaler:      &stubClusterAutoscaler{},
			wantErr:                true,
		},
		"kubeadm init fails when setting up verification service": {
			clusterUtil: stubClusterUtil{setupVerificationServiceErr: someErr},
			kubeconfigReader: &stubKubeconfigReader{
				Kubeconfig: []byte("someKubeconfig"),
			},
			providerMetadata:       &stubProviderMetadata{SupportedResp: false},
			CloudControllerManager: &stubCloudControllerManager{},
			CloudNodeManager:       &stubCloudNodeManager{SupportedResp: false},
			ClusterAutoscaler:      &stubClusterAutoscaler{},
			wantErr:                true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			kube := KubeWrapper{
				clusterUtil:            &tc.clusterUtil,
				providerMetadata:       tc.providerMetadata,
				cloudControllerManager: tc.CloudControllerManager,
				cloudNodeManager:       tc.CloudNodeManager,
				clusterAutoscaler:      tc.ClusterAutoscaler,
				configProvider:         &stubConfigProvider{InitConfig: k8sapi.KubeadmInitYAML{}},
				client:                 &tc.kubeCTL,
				kubeconfigReader:       tc.kubeconfigReader,
			}
			err := kube.InitCluster(context.Background(), autoscalingNodeGroups, serviceAccountUri, k8sVersion, attestationtypes.ID{}, KMSConfig{MasterSecret: masterSecret}, nil)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)

			var kubeadmConfig k8sapi.KubeadmInitYAML
			require.NoError(resources.UnmarshalK8SResources(tc.clusterUtil.initConfigs[0], &kubeadmConfig))
			require.Equal(tc.wantConfig.ClusterConfiguration, kubeadmConfig.ClusterConfiguration)
			require.Equal(tc.wantConfig.InitConfiguration, kubeadmConfig.InitConfiguration)
		})
	}
}

func TestJoinCluster(t *testing.T) {
	someErr := errors.New("failed")
	joinCommand := &kubeadm.BootstrapTokenDiscovery{
		APIServerEndpoint: "192.0.2.0:6443",
		Token:             "kube-fake-token",
		CACertHashes:      []string{"sha256:a60ebe9b0879090edd83b40a4df4bebb20506bac1e51d518ff8f4505a721930f"},
	}

	nodeVPNIP := "192.0.2.0"
	certKey := "cert-key"

	testCases := map[string]struct {
		clusterUtil            stubClusterUtil
		providerMetadata       ProviderMetadata
		CloudControllerManager CloudControllerManager
		wantConfig             kubeadm.JoinConfiguration
		role                   role.Role
		wantErr                bool
	}{
		"kubeadm join worker works without metadata": {
			clusterUtil:            stubClusterUtil{},
			providerMetadata:       &stubProviderMetadata{},
			CloudControllerManager: &stubCloudControllerManager{},
			role:                   role.Node,
			wantConfig: kubeadm.JoinConfiguration{
				Discovery: kubeadm.Discovery{
					BootstrapToken: joinCommand,
				},
				NodeRegistration: kubeadm.NodeRegistrationOptions{
					Name:             nodeVPNIP,
					KubeletExtraArgs: map[string]string{"node-ip": "192.0.2.0"},
				},
			},
		},
		"kubeadm join worker works with metadata": {
			clusterUtil: stubClusterUtil{},
			providerMetadata: &stubProviderMetadata{
				SupportedResp: true,
				SelfResp: cloudtypes.Instance{
					ProviderID: "provider-id",
					Name:       "metadata-name",
					PrivateIPs: []string{"192.0.2.1"},
				},
			},
			CloudControllerManager: &stubCloudControllerManager{},
			role:                   role.Node,
			wantConfig: kubeadm.JoinConfiguration{
				Discovery: kubeadm.Discovery{
					BootstrapToken: joinCommand,
				},
				NodeRegistration: kubeadm.NodeRegistrationOptions{
					Name:             "metadata-name",
					KubeletExtraArgs: map[string]string{"node-ip": "192.0.2.1"},
				},
			},
		},
		"kubeadm join worker works with metadata and cloud controller manager": {
			clusterUtil: stubClusterUtil{},
			providerMetadata: &stubProviderMetadata{
				SupportedResp: true,
				SelfResp: cloudtypes.Instance{
					ProviderID: "provider-id",
					Name:       "metadata-name",
					PrivateIPs: []string{"192.0.2.1"},
				},
			},
			CloudControllerManager: &stubCloudControllerManager{
				SupportedResp: true,
			},
			role: role.Node,
			wantConfig: kubeadm.JoinConfiguration{
				Discovery: kubeadm.Discovery{
					BootstrapToken: joinCommand,
				},
				NodeRegistration: kubeadm.NodeRegistrationOptions{
					Name:             "metadata-name",
					KubeletExtraArgs: map[string]string{"node-ip": "192.0.2.1"},
				},
			},
		},
		"kubeadm join control-plane node works with metadata": {
			clusterUtil: stubClusterUtil{},
			providerMetadata: &stubProviderMetadata{
				SupportedResp: true,
				SelfResp: cloudtypes.Instance{
					ProviderID: "provider-id",
					Name:       "metadata-name",
					PrivateIPs: []string{"192.0.2.1"},
				},
			},
			CloudControllerManager: &stubCloudControllerManager{},
			role:                   role.Coordinator,
			wantConfig: kubeadm.JoinConfiguration{
				Discovery: kubeadm.Discovery{
					BootstrapToken: joinCommand,
				},
				NodeRegistration: kubeadm.NodeRegistrationOptions{
					Name:             "metadata-name",
					KubeletExtraArgs: map[string]string{"node-ip": "192.0.2.1"},
				},
				ControlPlane: &kubeadm.JoinControlPlane{
					LocalAPIEndpoint: kubeadm.APIEndpoint{
						AdvertiseAddress: "192.0.2.1",
						BindPort:         6443,
					},
					CertificateKey: certKey,
				},
			},
		},
		"kubeadm join worker fails when retrieving self metadata": {
			clusterUtil: stubClusterUtil{},
			providerMetadata: &stubProviderMetadata{
				SupportedResp: true,
				SelfErr:       someErr,
			},
			CloudControllerManager: &stubCloudControllerManager{},
			role:                   role.Node,
			wantErr:                true,
		},
		"kubeadm join worker fails when applying the join config": {
			clusterUtil:            stubClusterUtil{joinClusterErr: someErr},
			providerMetadata:       &stubProviderMetadata{},
			CloudControllerManager: &stubCloudControllerManager{},
			role:                   role.Node,
			wantErr:                true,
		},
		"kubeadm join worker works fails when setting the metadata for the cloud controller manager": {
			clusterUtil: stubClusterUtil{},
			providerMetadata: &stubProviderMetadata{
				SupportedResp: true,
				SelfResp: cloudtypes.Instance{
					ProviderID: "provider-id",
					Name:       "metadata-name",
					PrivateIPs: []string{"192.0.2.1"},
				},
				SetVPNIPErr: someErr,
			},
			CloudControllerManager: &stubCloudControllerManager{
				SupportedResp: true,
			},
			role:    role.Node,
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			kube := KubeWrapper{
				clusterUtil:            &tc.clusterUtil,
				providerMetadata:       tc.providerMetadata,
				cloudControllerManager: tc.CloudControllerManager,
				configProvider:         &stubConfigProvider{},
			}

			err := kube.JoinCluster(context.Background(), joinCommand, nodeVPNIP, certKey, tc.role)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)

			var joinYaml k8sapi.KubeadmJoinYAML
			joinYaml, err = joinYaml.Unmarshal(tc.clusterUtil.joinConfigs[0])
			require.NoError(err)

			assert.Equal(tc.wantConfig, joinYaml.JoinConfiguration)
		})
	}
}

func TestGetKubeconfig(t *testing.T) {
	testCases := map[string]struct {
		Kubewrapper KubeWrapper
		wantErr     bool
	}{
		"check single replacement": {
			Kubewrapper: KubeWrapper{kubeconfigReader: &stubKubeconfigReader{
				Kubeconfig: []byte("127.0.0.1:16443"),
			}},
		},
		"check multiple replacement": {
			Kubewrapper: KubeWrapper{kubeconfigReader: &stubKubeconfigReader{
				Kubeconfig: []byte("127.0.0.1:16443...127.0.0.1:16443"),
			}},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)
			data, err := tc.Kubewrapper.GetKubeconfig()
			require.NoError(err)
			assert.NotContains(string(data), "127.0.0.1:16443")
			assert.Contains(string(data), "10.118.0.1:6443")
		})
	}
}

func TestK8sCompliantHostname(t *testing.T) {
	compliantHostname := regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`)
	testCases := map[string]struct {
		hostname     string
		wantHostname string
	}{
		"azure scale set names work": {
			hostname:     "constellation-scale-set-coordinators-name_0",
			wantHostname: "constellation-scale-set-coordinators-name-0",
		},
		"compliant hostname is not modified": {
			hostname:     "abcd-123",
			wantHostname: "abcd-123",
		},
		"uppercase hostnames are lowercased": {
			hostname:     "ABCD",
			wantHostname: "abcd",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			hostname := k8sCompliantHostname(tc.hostname)

			assert.Equal(tc.wantHostname, hostname)
			assert.Regexp(compliantHostname, hostname)
		})
	}
}

type stubClusterUtil struct {
	installComponentsErr             error
	initClusterErr                   error
	setupPodNetworkErr               error
	setupAutoscalingError            error
	setupActivationServiceError      error
	setupCloudControllerManagerError error
	setupCloudNodeManagerError       error
	setupKMSError                    error
	setupAccessManagerError          error
	setupVerificationServiceErr      error
	joinClusterErr                   error
	startKubeletErr                  error
	restartKubeletErr                error
	createJoinTokenResponse          *kubeadm.BootstrapTokenDiscovery
	createJoinTokenErr               error

	initConfigs [][]byte
	joinConfigs [][]byte
}

func (s *stubClusterUtil) InstallComponents(ctx context.Context, version string) error {
	return s.installComponentsErr
}

func (s *stubClusterUtil) InitCluster(ctx context.Context, initConfig []byte) error {
	s.initConfigs = append(s.initConfigs, initConfig)
	return s.initClusterErr
}

func (s *stubClusterUtil) SetupPodNetwork(context.Context, k8sapi.SetupPodNetworkInput) error {
	return s.setupPodNetworkErr
}

func (s *stubClusterUtil) SetupAutoscaling(kubectl k8sapi.Client, clusterAutoscalerConfiguration resources.Marshaler, secrets resources.Marshaler) error {
	return s.setupAutoscalingError
}

func (s *stubClusterUtil) SetupActivationService(kubectl k8sapi.Client, activationServiceConfiguration resources.Marshaler) error {
	return s.setupActivationServiceError
}

func (s *stubClusterUtil) SetupCloudControllerManager(kubectl k8sapi.Client, cloudControllerManagerConfiguration resources.Marshaler, configMaps resources.Marshaler, secrets resources.Marshaler) error {
	return s.setupCloudControllerManagerError
}

func (s *stubClusterUtil) SetupKMS(kubectl k8sapi.Client, kmsDeployment resources.Marshaler) error {
	return s.setupKMSError
}

func (s *stubClusterUtil) SetupAccessManager(kubectl k8sapi.Client, accessManagerConfiguration resources.Marshaler) error {
	return s.setupAccessManagerError
}

func (s *stubClusterUtil) SetupCloudNodeManager(kubectl k8sapi.Client, cloudNodeManagerConfiguration resources.Marshaler) error {
	return s.setupCloudNodeManagerError
}

func (s *stubClusterUtil) SetupVerificationService(kubectl k8sapi.Client, verificationServiceConfiguration resources.Marshaler) error {
	return s.setupVerificationServiceErr
}

func (s *stubClusterUtil) JoinCluster(ctx context.Context, joinConfig []byte) error {
	s.joinConfigs = append(s.joinConfigs, joinConfig)
	return s.joinClusterErr
}

func (s *stubClusterUtil) StartKubelet() error {
	return s.startKubeletErr
}

func (s *stubClusterUtil) RestartKubelet() error {
	return s.restartKubeletErr
}

func (s *stubClusterUtil) GetControlPlaneJoinCertificateKey(context.Context) (string, error) {
	return "", nil
}

func (s *stubClusterUtil) CreateJoinToken(ctx context.Context, ttl time.Duration) (*kubeadm.BootstrapTokenDiscovery, error) {
	return s.createJoinTokenResponse, s.createJoinTokenErr
}

func (s *stubClusterUtil) FixCilium(nodeName string) {
}

type stubConfigProvider struct {
	InitConfig k8sapi.KubeadmInitYAML
	JoinConfig k8sapi.KubeadmJoinYAML
}

func (s *stubConfigProvider) InitConfiguration(_ bool) k8sapi.KubeadmInitYAML {
	return s.InitConfig
}

func (s *stubConfigProvider) JoinConfiguration(_ bool) k8sapi.KubeadmJoinYAML {
	s.JoinConfig = k8sapi.KubeadmJoinYAML{
		JoinConfiguration: kubeadm.JoinConfiguration{
			Discovery: kubeadm.Discovery{
				BootstrapToken: &kubeadm.BootstrapTokenDiscovery{},
			},
		},
	}
	return s.JoinConfig
}

type stubKubeCTL struct {
	ApplyErr error

	resources   []resources.Marshaler
	kubeconfigs [][]byte
}

func (s *stubKubeCTL) Apply(resources resources.Marshaler, forceConflicts bool) error {
	s.resources = append(s.resources, resources)
	return s.ApplyErr
}

func (s *stubKubeCTL) SetKubeconfig(kubeconfig []byte) {
	s.kubeconfigs = append(s.kubeconfigs, kubeconfig)
}

type stubKubeconfigReader struct {
	Kubeconfig []byte
	ReadErr    error
}

func (s *stubKubeconfigReader) ReadKubeconfig() ([]byte, error) {
	return s.Kubeconfig, s.ReadErr
}
