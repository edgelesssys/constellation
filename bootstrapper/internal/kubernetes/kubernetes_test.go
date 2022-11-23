/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package kubernetes

import (
	"context"
	"errors"
	"net"
	"regexp"
	"strconv"
	"testing"

	"github.com/edgelesssys/constellation/v2/bootstrapper/internal/kubernetes/k8sapi"
	kubewaiter "github.com/edgelesssys/constellation/v2/bootstrapper/internal/kubernetes/kubeWaiter"
	"github.com/edgelesssys/constellation/v2/internal/cloud/metadata"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/deploy/helm"
	"github.com/edgelesssys/constellation/v2/internal/kubernetes"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/role"
	"github.com/edgelesssys/constellation/v2/internal/versions"
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

	nodeName := "node-name"
	providerID := "provider-id"
	privateIP := "192.0.2.1"
	loadbalancerIP := "192.0.2.3"
	aliasIPRange := "192.0.2.0/24"

	testCases := map[string]struct {
		clusterUtil      stubClusterUtil
		helmClient       stubHelmClient
		kubectl          stubKubectl
		kubeAPIWaiter    stubKubeAPIWaiter
		providerMetadata ProviderMetadata
		kubeconfigReader configReader
		wantConfig       k8sapi.KubeadmInitYAML
		wantErr          bool
		k8sVersion       versions.ValidK8sVersion
	}{
		"kubeadm init works with metadata and loadbalancer": {
			clusterUtil: stubClusterUtil{},
			kubeconfigReader: &stubKubeconfigReader{
				kubeconfig: []byte("someKubeconfig"),
			},
			kubeAPIWaiter: stubKubeAPIWaiter{},
			providerMetadata: &stubProviderMetadata{
				selfResp: metadata.InstanceMetadata{
					Name:          nodeName,
					ProviderID:    providerID,
					VPCIP:         privateIP,
					AliasIPRanges: []string{aliasIPRange},
				},
				getLoadBalancerEndpointResp: loadbalancerIP,
			},
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
						CertSANs: []string{privateIP},
					},
				},
			},
			wantErr:    false,
			k8sVersion: versions.Default,
		},
		"kubeadm init fails when retrieving metadata self": {
			clusterUtil: stubClusterUtil{},
			kubeconfigReader: &stubKubeconfigReader{
				kubeconfig: []byte("someKubeconfig"),
			},
			kubeAPIWaiter: stubKubeAPIWaiter{},
			providerMetadata: &stubProviderMetadata{
				selfErr: someErr,
			},
			wantErr:    true,
			k8sVersion: versions.Default,
		},
		"kubeadm init fails when retrieving metadata loadbalancer ip": {
			clusterUtil: stubClusterUtil{},
			kubeconfigReader: &stubKubeconfigReader{
				kubeconfig: []byte("someKubeconfig"),
			},
			providerMetadata: &stubProviderMetadata{
				getLoadBalancerEndpointErr: someErr,
			},
			wantErr:    true,
			k8sVersion: versions.Default,
		},
		"kubeadm init fails when applying the init config": {
			clusterUtil: stubClusterUtil{initClusterErr: someErr},
			kubeconfigReader: &stubKubeconfigReader{
				kubeconfig: []byte("someKubeconfig"),
			},
			kubeAPIWaiter:    stubKubeAPIWaiter{},
			providerMetadata: &stubProviderMetadata{},
			wantErr:          true,
			k8sVersion:       versions.Default,
		},
		"kubeadm init fails when deploying cilium": {
			clusterUtil: stubClusterUtil{},
			helmClient:  stubHelmClient{ciliumError: someErr},
			kubeconfigReader: &stubKubeconfigReader{
				kubeconfig: []byte("someKubeconfig"),
			},
			providerMetadata: &stubProviderMetadata{},
			wantErr:          true,
			k8sVersion:       versions.Default,
		},
		"kubeadm init fails when setting up constellation-services chart": {
			clusterUtil: stubClusterUtil{},
			helmClient:  stubHelmClient{servicesError: someErr},
			kubeconfigReader: &stubKubeconfigReader{
				kubeconfig: []byte("someKubeconfig"),
			},
			kubeAPIWaiter:    stubKubeAPIWaiter{},
			providerMetadata: &stubProviderMetadata{},
			wantErr:          true,
			k8sVersion:       versions.Default,
		},
		"kubeadm init fails when setting the cloud node manager": {
			clusterUtil: stubClusterUtil{},
			helmClient:  stubHelmClient{servicesError: someErr},
			kubeconfigReader: &stubKubeconfigReader{
				kubeconfig: []byte("someKubeconfig"),
			},
			kubeAPIWaiter:    stubKubeAPIWaiter{},
			providerMetadata: &stubProviderMetadata{},
			wantErr:          true,
			k8sVersion:       versions.Default,
		},
		"kubeadm init fails when setting the cluster autoscaler": {
			clusterUtil: stubClusterUtil{},
			helmClient:  stubHelmClient{servicesError: someErr},
			kubeconfigReader: &stubKubeconfigReader{
				kubeconfig: []byte("someKubeconfig"),
			},
			kubeAPIWaiter:    stubKubeAPIWaiter{},
			providerMetadata: &stubProviderMetadata{},
			wantErr:          true,
			k8sVersion:       versions.Default,
		},
		"kubeadm init fails when reading kubeconfig": {
			clusterUtil: stubClusterUtil{},
			kubeconfigReader: &stubKubeconfigReader{
				readErr: someErr,
			},
			kubeAPIWaiter:    stubKubeAPIWaiter{},
			providerMetadata: &stubProviderMetadata{},
			wantErr:          true,
			k8sVersion:       versions.Default,
		},
		"kubeadm init fails when setting up konnectivity": {
			clusterUtil: stubClusterUtil{setupKonnectivityError: someErr},
			kubeconfigReader: &stubKubeconfigReader{
				kubeconfig: []byte("someKubeconfig"),
			},
			kubeAPIWaiter:    stubKubeAPIWaiter{},
			providerMetadata: &stubProviderMetadata{},
			wantErr:          true,
			k8sVersion:       versions.Default,
		},
		"kubeadm init fails when setting up verification service": {
			clusterUtil: stubClusterUtil{},
			helmClient:  stubHelmClient{servicesError: someErr},
			kubeconfigReader: &stubKubeconfigReader{
				kubeconfig: []byte("someKubeconfig"),
			},
			kubeAPIWaiter:    stubKubeAPIWaiter{},
			providerMetadata: &stubProviderMetadata{},
			wantErr:          true,
			k8sVersion:       versions.Default,
		},
		"kubeadm init fails when waiting for kubeAPI server": {
			clusterUtil: stubClusterUtil{},
			kubeconfigReader: &stubKubeconfigReader{
				kubeconfig: []byte("someKubeconfig"),
			},
			kubeAPIWaiter:    stubKubeAPIWaiter{waitErr: someErr},
			providerMetadata: &stubProviderMetadata{},
			k8sVersion:       versions.Default,
			wantErr:          true,
		},
		"unsupported k8sVersion fails cluster creation": {
			clusterUtil: stubClusterUtil{},
			kubeconfigReader: &stubKubeconfigReader{
				kubeconfig: []byte("someKubeconfig"),
			},
			kubeAPIWaiter:    stubKubeAPIWaiter{},
			providerMetadata: &stubProviderMetadata{},
			k8sVersion:       "1.19",
			wantErr:          true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			kube := KubeWrapper{
				clusterUtil:      &tc.clusterUtil,
				helmClient:       &tc.helmClient,
				providerMetadata: tc.providerMetadata,
				kubeAPIWaiter:    &tc.kubeAPIWaiter,
				configProvider:   &stubConfigProvider{initConfig: k8sapi.KubeadmInitYAML{}},
				client:           &tc.kubectl,
				kubeconfigReader: tc.kubeconfigReader,
				getIPAddr:        func() (string, error) { return privateIP, nil },
			}

			_, err := kube.InitCluster(
				context.Background(), serviceAccountURI, string(tc.k8sVersion),
				nil, nil, false, nil, true, []byte("{}"), false, nil, logger.NewTest(t),
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
	k8sVersion := versions.Default

	testCases := map[string]struct {
		clusterUtil      stubClusterUtil
		providerMetadata ProviderMetadata
		wantConfig       kubeadm.JoinConfiguration
		role             role.Role
		wantErr          bool
	}{
		"kubeadm join worker works with metadata": {
			clusterUtil: stubClusterUtil{},
			providerMetadata: &stubProviderMetadata{
				selfResp: metadata.InstanceMetadata{
					ProviderID: "provider-id",
					Name:       "metadata-name",
					VPCIP:      "192.0.2.1",
				},
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
		"kubeadm join worker works with metadata and cloud controller manager": {
			clusterUtil: stubClusterUtil{},
			providerMetadata: &stubProviderMetadata{
				selfResp: metadata.InstanceMetadata{
					ProviderID: "provider-id",
					Name:       "metadata-name",
					VPCIP:      "192.0.2.1",
				},
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
				selfResp: metadata.InstanceMetadata{
					ProviderID: "provider-id",
					Name:       "metadata-name",
					VPCIP:      "192.0.2.1",
				},
			},
			role: role.ControlPlane,
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
				selfErr: someErr,
			},
			role:    role.Worker,
			wantErr: true,
		},
		"kubeadm join worker fails when applying the join config": {
			clusterUtil:      stubClusterUtil{joinClusterErr: someErr},
			providerMetadata: &stubProviderMetadata{},
			role:             role.Worker,
			wantErr:          true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			kube := KubeWrapper{
				clusterUtil:      &tc.clusterUtil,
				providerMetadata: tc.providerMetadata,
				configProvider:   &stubConfigProvider{},
				getIPAddr:        func() (string, error) { return privateIP, nil },
			}

			err := kube.JoinCluster(context.Background(), joinCommand, tc.role, string(k8sVersion), logger.NewTest(t))
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)

			var joinYaml k8sapi.KubeadmJoinYAML
			require.NoError(kubernetes.UnmarshalK8SResources(tc.clusterUtil.joinConfigs[0], &joinYaml))

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
	installComponentsErr        error
	installComponentsFromCLIErr error
	initClusterErr              error
	setupAutoscalingError       error
	setupKonnectivityError      error
	setupGCPGuestAgentErr       error
	setupOLMErr                 error
	setupNMOErr                 error
	setupNodeOperatorErr        error
	joinClusterErr              error
	startKubeletErr             error

	initConfigs [][]byte
	joinConfigs [][]byte
}

func (s *stubClusterUtil) SetupKonnectivity(kubectl k8sapi.Client, konnectivityAgentsDaemonSet kubernetes.Marshaler) error {
	return s.setupKonnectivityError
}

func (s *stubClusterUtil) InstallComponents(ctx context.Context, version versions.ValidK8sVersion) error {
	return s.installComponentsErr
}

func (s *stubClusterUtil) InstallComponentsFromCLI(ctx context.Context, kubernetesComponents versions.ComponentVersions) error {
	return s.installComponentsFromCLIErr
}

// TODO: Upon changing this function, please refactor it to reduce the number of arguments to <= 5.
//
//revive:disable-next-line
func (s *stubClusterUtil) InitCluster(ctx context.Context, initConfig []byte, nodeName string, ips []net.IP, controlPlaneEndpoint string, conformanceMode bool, log *logger.Logger) error {
	s.initConfigs = append(s.initConfigs, initConfig)
	return s.initClusterErr
}

func (s *stubClusterUtil) SetupAutoscaling(kubectl k8sapi.Client, clusterAutoscalerConfiguration kubernetes.Marshaler, secrets kubernetes.Marshaler) error {
	return s.setupAutoscalingError
}

func (s *stubClusterUtil) SetupGCPGuestAgent(kubectl k8sapi.Client, gcpGuestAgentConfiguration kubernetes.Marshaler) error {
	return s.setupGCPGuestAgentErr
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

func (s *stubClusterUtil) JoinCluster(ctx context.Context, joinConfig []byte, peerRole role.Role, controlPlaneEndpoint string, log *logger.Logger) error {
	s.joinConfigs = append(s.joinConfigs, joinConfig)
	return s.joinClusterErr
}

func (s *stubClusterUtil) StartKubelet() error {
	return s.startKubeletErr
}

func (s *stubClusterUtil) FixCilium(log *logger.Logger) {
}

type stubConfigProvider struct {
	initConfig k8sapi.KubeadmInitYAML
	joinConfig k8sapi.KubeadmJoinYAML
}

func (s *stubConfigProvider) InitConfiguration(_ bool, _ versions.ValidK8sVersion) k8sapi.KubeadmInitYAML {
	return s.initConfig
}

func (s *stubConfigProvider) JoinConfiguration(_ bool) k8sapi.KubeadmJoinYAML {
	s.joinConfig = k8sapi.KubeadmJoinYAML{
		JoinConfiguration: kubeadm.JoinConfiguration{
			Discovery: kubeadm.Discovery{
				BootstrapToken: &kubeadm.BootstrapTokenDiscovery{},
			},
		},
	}
	return s.joinConfig
}

type stubKubectl struct {
	applyErr                         error
	createConfigMapErr               error
	addTolerationsToDeploymentErr    error
	addTNodeSelectorsToDeploymentErr error
	waitForCRDsErr                   error
	listAllNamespacesErr             error

	listAllNamespacesResp *corev1.NamespaceList
	resources             []kubernetes.Marshaler
	kubeconfigs           [][]byte
}

func (s *stubKubectl) Apply(resources kubernetes.Marshaler, forceConflicts bool) error {
	s.resources = append(s.resources, resources)
	return s.applyErr
}

func (s *stubKubectl) SetKubeconfig(kubeconfig []byte) {
	s.kubeconfigs = append(s.kubeconfigs, kubeconfig)
}

func (s *stubKubectl) CreateConfigMap(ctx context.Context, configMap corev1.ConfigMap) error {
	return s.createConfigMapErr
}

func (s *stubKubectl) AddTolerationsToDeployment(ctx context.Context, tolerations []corev1.Toleration, name string, namespace string) error {
	return s.addTolerationsToDeploymentErr
}

func (s *stubKubectl) AddNodeSelectorsToDeployment(ctx context.Context, selectors map[string]string, name string, namespace string) error {
	return s.addTNodeSelectorsToDeploymentErr
}

func (s *stubKubectl) WaitForCRDs(ctx context.Context, crds []string) error {
	return s.waitForCRDsErr
}

func (s *stubKubectl) ListAllNamespaces(ctx context.Context) (*corev1.NamespaceList, error) {
	return s.listAllNamespacesResp, s.listAllNamespacesErr
}

type stubKubeconfigReader struct {
	kubeconfig []byte
	readErr    error
}

func (s *stubKubeconfigReader) ReadKubeconfig() ([]byte, error) {
	return s.kubeconfig, s.readErr
}

type stubHelmClient struct {
	ciliumError      error
	certManagerError error
	operatorsError   error
	servicesError    error
}

func (s *stubHelmClient) InstallCilium(ctx context.Context, kubectl k8sapi.Client, release helm.Release, in k8sapi.SetupPodNetworkInput) error {
	return s.ciliumError
}

func (s *stubHelmClient) InstallCertManager(ctx context.Context, release helm.Release) error {
	return s.certManagerError
}

func (s *stubHelmClient) InstallOperators(ctx context.Context, release helm.Release, extraVals map[string]any) error {
	return s.operatorsError
}

func (s *stubHelmClient) InstallConstellationServices(ctx context.Context, release helm.Release, extraVals map[string]any) error {
	return s.servicesError
}

type stubKubeAPIWaiter struct {
	waitErr error
}

func (s *stubKubeAPIWaiter) Wait(_ context.Context, _ kubewaiter.KubernetesClient) error {
	return s.waitErr
}
