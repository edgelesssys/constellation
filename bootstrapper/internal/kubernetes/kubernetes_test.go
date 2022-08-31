package kubernetes

import (
	"context"
	"errors"
	"net"
	"regexp"
	"strconv"
	"testing"

	"github.com/edgelesssys/constellation/bootstrapper/internal/kubernetes/k8sapi"
	"github.com/edgelesssys/constellation/bootstrapper/internal/kubernetes/k8sapi/resources"
	"github.com/edgelesssys/constellation/internal/cloud/metadata"
	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/edgelesssys/constellation/internal/kubernetes"
	"github.com/edgelesssys/constellation/internal/logger"
	"github.com/edgelesssys/constellation/internal/role"
	"github.com/edgelesssys/constellation/internal/versions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
	corev1 "k8s.io/api/core/v1"
	kubeadm "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta3"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestInitCluster(t *testing.T) {
	someErr := errors.New("failed")
	serviceAccountURI := "some-service-account-uri"
	masterSecret := []byte("some-master-secret")
	autoscalingNodeGroups := []string{"0,10,autoscaling_group_0"}

	nodeName := "node-name"
	providerID := "provider-id"
	privateIP := "192.0.2.1"
	publicIP := "192.0.2.2"
	loadbalancerIP := "192.0.2.3"
	aliasIPRange := "192.0.2.0/24"

	testCases := map[string]struct {
		clusterUtil            stubClusterUtil
		kubectl                stubKubectl
		providerMetadata       ProviderMetadata
		CloudControllerManager CloudControllerManager
		CloudNodeManager       CloudNodeManager
		ClusterAutoscaler      ClusterAutoscaler
		kubeconfigReader       configReader
		wantConfig             k8sapi.KubeadmInitYAML
		wantErr                bool
		k8sVersion             versions.ValidK8sVersion
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
						Name: privateIP,
					},
				},
				ClusterConfiguration: kubeadm.ClusterConfiguration{},
			},
			k8sVersion: versions.Latest,
		},
		"kubeadm init works with metadata and loadbalancer": {
			clusterUtil: stubClusterUtil{},
			kubeconfigReader: &stubKubeconfigReader{
				Kubeconfig: []byte("someKubeconfig"),
			},
			providerMetadata: &stubProviderMetadata{
				SupportedResp: true,
				SelfResp: metadata.InstanceMetadata{
					Name:          nodeName,
					ProviderID:    providerID,
					VPCIP:         privateIP,
					PublicIP:      publicIP,
					AliasIPRanges: []string{aliasIPRange},
				},
				GetLoadBalancerEndpointResp: loadbalancerIP,
				SupportsLoadBalancerResp:    true,
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
			wantErr:    false,
			k8sVersion: versions.Latest,
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
			k8sVersion:             versions.Latest,
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
			k8sVersion:             versions.Latest,
		},
		"kubeadm init fails when retrieving metadata loadbalancer ip": {
			clusterUtil: stubClusterUtil{},
			kubeconfigReader: &stubKubeconfigReader{
				Kubeconfig: []byte("someKubeconfig"),
			},
			providerMetadata: &stubProviderMetadata{
				GetLoadBalancerEndpointErr: someErr,
				SupportsLoadBalancerResp:   true,
				SupportedResp:              true,
			},
			CloudControllerManager: &stubCloudControllerManager{},
			CloudNodeManager:       &stubCloudNodeManager{},
			ClusterAutoscaler:      &stubClusterAutoscaler{},
			wantErr:                true,
			k8sVersion:             versions.Latest,
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
			k8sVersion:             versions.Latest,
		},
		"kubeadm init fails when deploying helm charts": {
			clusterUtil: stubClusterUtil{setupHelmDeploymentsErr: someErr},
			kubeconfigReader: &stubKubeconfigReader{
				Kubeconfig: []byte("someKubeconfig"),
			},
			providerMetadata:       &stubProviderMetadata{},
			CloudControllerManager: &stubCloudControllerManager{},
			CloudNodeManager:       &stubCloudNodeManager{},
			ClusterAutoscaler:      &stubClusterAutoscaler{},
			wantErr:                true,
			k8sVersion:             versions.Latest,
		},
		"kubeadm init fails when setting up the join service": {
			clusterUtil: stubClusterUtil{setupJoinServiceError: someErr},
			kubeconfigReader: &stubKubeconfigReader{
				Kubeconfig: []byte("someKubeconfig"),
			},
			providerMetadata:       &stubProviderMetadata{},
			CloudControllerManager: &stubCloudControllerManager{SupportedResp: true},
			CloudNodeManager:       &stubCloudNodeManager{},
			ClusterAutoscaler:      &stubClusterAutoscaler{},
			wantErr:                true,
			k8sVersion:             versions.Latest,
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
			k8sVersion:             versions.Latest,
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
			k8sVersion:             versions.Latest,
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
			k8sVersion:             versions.Latest,
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
			k8sVersion:             versions.Latest,
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
			k8sVersion:             versions.Latest,
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
			k8sVersion:             versions.Latest,
		},
		"unsupported k8sVersion fails cluster creation": {
			clusterUtil: stubClusterUtil{},
			kubeconfigReader: &stubKubeconfigReader{
				Kubeconfig: []byte("someKubeconfig"),
			},
			providerMetadata:       &stubProviderMetadata{},
			CloudControllerManager: &stubCloudControllerManager{},
			CloudNodeManager:       &stubCloudNodeManager{},
			ClusterAutoscaler:      &stubClusterAutoscaler{},
			k8sVersion:             "1.19",
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
				client:                 &tc.kubectl,
				kubeconfigReader:       tc.kubeconfigReader,
				getIPAddr:              func() (string, error) { return privateIP, nil },
			}

			_, err := kube.InitCluster(
				context.Background(), autoscalingNodeGroups, serviceAccountURI, string(tc.k8sVersion),
				nil, nil, false, true, resources.KMSConfig{MasterSecret: masterSecret}, nil, nil, logger.NewTest(t),
			)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)

			var kubeadmConfig k8sapi.KubeadmInitYAML
			require.NoError(kubernetes.UnmarshalK8SResources(tc.clusterUtil.initConfigs[0], &kubeadmConfig))
			require.Equal(tc.wantConfig.ClusterConfiguration, kubeadmConfig.ClusterConfiguration)
			require.Equal(tc.wantConfig.InitConfiguration, kubeadmConfig.InitConfiguration)
		})
	}
}

func TestJoinCluster(t *testing.T) {
	someErr := errors.New("failed")
	joinCommand := &kubeadm.BootstrapTokenDiscovery{
		APIServerEndpoint: "192.0.2.0:" + strconv.Itoa(constants.KubernetesPort),
		Token:             "kube-fake-token",
		CACertHashes:      []string{"sha256:a60ebe9b0879090edd83b40a4df4bebb20506bac1e51d518ff8f4505a721930f"},
	}

	privateIP := "192.0.2.1"
	k8sVersion := versions.Latest

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
			role:                   role.Worker,
			wantConfig: kubeadm.JoinConfiguration{
				Discovery: kubeadm.Discovery{
					BootstrapToken: joinCommand,
				},
				NodeRegistration: kubeadm.NodeRegistrationOptions{
					Name:             privateIP,
					KubeletExtraArgs: map[string]string{"node-ip": privateIP},
				},
			},
		},
		"kubeadm join worker works with metadata": {
			clusterUtil: stubClusterUtil{},
			providerMetadata: &stubProviderMetadata{
				SupportedResp: true,
				SelfResp: metadata.InstanceMetadata{
					ProviderID: "provider-id",
					Name:       "metadata-name",
					VPCIP:      "192.0.2.1",
				},
			},
			CloudControllerManager: &stubCloudControllerManager{},
			role:                   role.Worker,
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
				SelfResp: metadata.InstanceMetadata{
					ProviderID: "provider-id",
					Name:       "metadata-name",
					VPCIP:      "192.0.2.1",
				},
			},
			CloudControllerManager: &stubCloudControllerManager{
				SupportedResp: true,
			},
			role: role.Worker,
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
				SelfResp: metadata.InstanceMetadata{
					ProviderID: "provider-id",
					Name:       "metadata-name",
					VPCIP:      "192.0.2.1",
				},
			},
			CloudControllerManager: &stubCloudControllerManager{},
			role:                   role.ControlPlane,
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
						BindPort:         constants.KubernetesPort,
					},
				},
				SkipPhases: []string{"control-plane-prepare/download-certs"},
			},
		},
		"kubeadm join worker fails when retrieving self metadata": {
			clusterUtil: stubClusterUtil{},
			providerMetadata: &stubProviderMetadata{
				SupportedResp: true,
				SelfErr:       someErr,
			},
			CloudControllerManager: &stubCloudControllerManager{},
			role:                   role.Worker,
			wantErr:                true,
		},
		"kubeadm join worker fails when applying the join config": {
			clusterUtil:            stubClusterUtil{joinClusterErr: someErr},
			providerMetadata:       &stubProviderMetadata{},
			CloudControllerManager: &stubCloudControllerManager{},
			role:                   role.Worker,
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
				configProvider:         &stubConfigProvider{},
				getIPAddr:              func() (string, error) { return privateIP, nil },
			}

			err := kube.JoinCluster(context.Background(), joinCommand, tc.role, string(k8sVersion), logger.NewTest(t))
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

func TestK8sCompliantHostname(t *testing.T) {
	compliantHostname := regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`)
	testCases := map[string]struct {
		hostname     string
		wantHostname string
	}{
		"azure scale set names work": {
			hostname:     "constellation-scale-set-bootstrappers-name_0",
			wantHostname: "constellation-scale-set-bootstrappers-name-0",
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
	setupHelmDeploymentsErr          error
	setupAutoscalingError            error
	setupJoinServiceError            error
	setupCloudControllerManagerError error
	setupCloudNodeManagerError       error
	setupKMSError                    error
	setupAccessManagerError          error
	setupVerificationServiceErr      error
	setupGCPGuestAgentErr            error
	setupOLMErr                      error
	setupNMOErr                      error
	setupNodeOperatorErr             error
	joinClusterErr                   error
	startKubeletErr                  error
	restartKubeletErr                error

	initConfigs [][]byte
	joinConfigs [][]byte
}

func (s *stubClusterUtil) InstallComponents(ctx context.Context, version versions.ValidK8sVersion) error {
	return s.installComponentsErr
}

func (s *stubClusterUtil) InitCluster(ctx context.Context, initConfig []byte, nodeName string, ips []net.IP, log *logger.Logger) error {
	s.initConfigs = append(s.initConfigs, initConfig)
	return s.initClusterErr
}

func (s *stubClusterUtil) SetupHelmDeployments(context.Context, k8sapi.Client, []byte, k8sapi.SetupPodNetworkInput, *logger.Logger) error {
	return s.setupHelmDeploymentsErr
}

func (s *stubClusterUtil) SetupAutoscaling(kubectl k8sapi.Client, clusterAutoscalerConfiguration kubernetes.Marshaler, secrets kubernetes.Marshaler) error {
	return s.setupAutoscalingError
}

func (s *stubClusterUtil) SetupJoinService(kubectl k8sapi.Client, joinServiceConfiguration kubernetes.Marshaler) error {
	return s.setupJoinServiceError
}

func (s *stubClusterUtil) SetupGCPGuestAgent(kubectl k8sapi.Client, gcpGuestAgentConfiguration kubernetes.Marshaler) error {
	return s.setupGCPGuestAgentErr
}

func (s *stubClusterUtil) SetupCloudControllerManager(kubectl k8sapi.Client, cloudControllerManagerConfiguration kubernetes.Marshaler, configMaps kubernetes.Marshaler, secrets kubernetes.Marshaler) error {
	return s.setupCloudControllerManagerError
}

func (s *stubClusterUtil) SetupKMS(kubectl k8sapi.Client, kmsDeployment kubernetes.Marshaler) error {
	return s.setupKMSError
}

func (s *stubClusterUtil) SetupAccessManager(kubectl k8sapi.Client, accessManagerConfiguration kubernetes.Marshaler) error {
	return s.setupAccessManagerError
}

func (s *stubClusterUtil) SetupCloudNodeManager(kubectl k8sapi.Client, cloudNodeManagerConfiguration kubernetes.Marshaler) error {
	return s.setupCloudNodeManagerError
}

func (s *stubClusterUtil) SetupVerificationService(kubectl k8sapi.Client, verificationServiceConfiguration kubernetes.Marshaler) error {
	return s.setupVerificationServiceErr
}

func (s *stubClusterUtil) SetupOperatorLifecycleManager(ctx context.Context, kubectl k8sapi.Client, olmCRDs, olmConfiguration kubernetes.Marshaler, crdNames []string) error {
	return s.setupOLMErr
}

func (s *stubClusterUtil) SetupNodeMaintenanceOperator(kubectl k8sapi.Client, nodeMaintenanceOperatorConfiguration kubernetes.Marshaler) error {
	return s.setupNMOErr
}

func (s *stubClusterUtil) SetupNodeOperator(ctx context.Context, kubectl k8sapi.Client, nodeOperatorConfiguration kubernetes.Marshaler) error {
	return s.setupNodeOperatorErr
}

func (s *stubClusterUtil) JoinCluster(ctx context.Context, joinConfig []byte, log *logger.Logger) error {
	s.joinConfigs = append(s.joinConfigs, joinConfig)
	return s.joinClusterErr
}

func (s *stubClusterUtil) StartKubelet() error {
	return s.startKubeletErr
}

func (s *stubClusterUtil) RestartKubelet() error {
	return s.restartKubeletErr
}

func (s *stubClusterUtil) FixCilium(nodeName string, log *logger.Logger) {
}

type stubConfigProvider struct {
	InitConfig k8sapi.KubeadmInitYAML
	JoinConfig k8sapi.KubeadmJoinYAML
}

func (s *stubConfigProvider) InitConfiguration(_ bool, _ versions.ValidK8sVersion) k8sapi.KubeadmInitYAML {
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

type stubKubectl struct {
	ApplyErr                         error
	createConfigMapErr               error
	AddTolerationsToDeploymentErr    error
	AddTNodeSelectorsToDeploymentErr error
	waitForCRDsErr                   error

	resources   []kubernetes.Marshaler
	kubeconfigs [][]byte
}

func (s *stubKubectl) Apply(resources kubernetes.Marshaler, forceConflicts bool) error {
	s.resources = append(s.resources, resources)
	return s.ApplyErr
}

func (s *stubKubectl) SetKubeconfig(kubeconfig []byte) {
	s.kubeconfigs = append(s.kubeconfigs, kubeconfig)
}

func (s *stubKubectl) CreateConfigMap(ctx context.Context, configMap corev1.ConfigMap) error {
	return s.createConfigMapErr
}

func (s *stubKubectl) AddTolerationsToDeployment(ctx context.Context, tolerations []corev1.Toleration, name string, namespace string) error {
	return s.AddTolerationsToDeploymentErr
}

func (s *stubKubectl) AddNodeSelectorsToDeployment(ctx context.Context, selectors map[string]string, name string, namespace string) error {
	return s.AddTNodeSelectorsToDeploymentErr
}

func (s *stubKubectl) WaitForCRDs(ctx context.Context, crds []string) error {
	return s.waitForCRDsErr
}

type stubKubeconfigReader struct {
	Kubeconfig []byte
	ReadErr    error
}

func (s *stubKubeconfigReader) ReadKubeconfig() ([]byte, error) {
	return s.Kubeconfig, s.ReadErr
}
