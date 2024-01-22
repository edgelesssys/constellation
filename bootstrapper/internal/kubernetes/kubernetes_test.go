/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package kubernetes

import (
	"context"
	"errors"
	"net"
	"strconv"
	"testing"
  "log/slog"

	"github.com/edgelesssys/constellation/v2/bootstrapper/internal/kubernetes/k8sapi"
	"github.com/edgelesssys/constellation/v2/bootstrapper/internal/kubernetes/kubewaiter"
	"github.com/edgelesssys/constellation/v2/internal/cloud/metadata"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/kubernetes"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/role"
	"github.com/edgelesssys/constellation/v2/internal/versions"
	"github.com/edgelesssys/constellation/v2/internal/versions/components"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
	corev1 "k8s.io/api/core/v1"
	kubeadm "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta3"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m, goleak.IgnoreAnyFunction("github.com/bazelbuild/rules_go/go/tools/bzltestutil.RegisterTimeoutHandler.func1"))
}

func TestInitCluster(t *testing.T) {
	nodeName := "node-name"
	providerID := "provider-id"
	privateIP := "192.0.2.1"
	loadbalancerIP := "192.0.2.3"
	aliasIPRange := "192.0.2.0/24"

	testCases := map[string]struct {
		clusterUtil      stubClusterUtil
		kubectl          stubKubectl
		kubeAPIWaiter    stubKubeAPIWaiter
		providerMetadata ProviderMetadata
		wantConfig       k8sapi.KubeadmInitYAML
		wantErr          bool
		k8sVersion       versions.ValidK8sVersion
	}{
		"kubeadm init works with metadata and loadbalancer": {
			clusterUtil:   stubClusterUtil{kubeconfig: []byte("someKubeconfig")},
			kubeAPIWaiter: stubKubeAPIWaiter{},
			providerMetadata: &stubProviderMetadata{
				selfResp: metadata.InstanceMetadata{
					Name:          nodeName,
					ProviderID:    providerID,
					VPCIP:         privateIP,
					AliasIPRanges: []string{aliasIPRange},
				},
				getLoadBalancerHostResp: loadbalancerIP,
				getLoadBalancerPortResp: strconv.Itoa(constants.KubernetesPort),
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
					ClusterName:          "kubernetes",
					ControlPlaneEndpoint: loadbalancerIP,
					APIServer: kubeadm.APIServer{
						CertSANs: []string{privateIP},
					},
				},
			},
			wantErr:    false,
			k8sVersion: versions.Default,
		},
		"kubeadm init fails when annotating itself": {
			clusterUtil:   stubClusterUtil{kubeconfig: []byte("someKubeconfig")},
			kubeAPIWaiter: stubKubeAPIWaiter{},
			providerMetadata: &stubProviderMetadata{
				selfResp: metadata.InstanceMetadata{
					Name:          nodeName,
					ProviderID:    providerID,
					VPCIP:         privateIP,
					AliasIPRanges: []string{aliasIPRange},
				},
				getLoadBalancerHostResp: loadbalancerIP,
				getLoadBalancerPortResp: strconv.Itoa(constants.KubernetesPort),
			},
			kubectl:    stubKubectl{annotateNodeErr: assert.AnError},
			wantErr:    true,
			k8sVersion: versions.Default,
		},
		"kubeadm init fails when retrieving metadata self": {
			clusterUtil:   stubClusterUtil{kubeconfig: []byte("someKubeconfig")},
			kubeAPIWaiter: stubKubeAPIWaiter{},
			providerMetadata: &stubProviderMetadata{
				selfErr: assert.AnError,
			},
			wantErr:    true,
			k8sVersion: versions.Default,
		},
		"kubeadm init fails when retrieving metadata loadbalancer ip": {
			clusterUtil: stubClusterUtil{kubeconfig: []byte("someKubeconfig")},
			providerMetadata: &stubProviderMetadata{
				getLoadBalancerEndpointErr: assert.AnError,
			},
			wantErr:    true,
			k8sVersion: versions.Default,
		},
		"kubeadm init fails when applying the init config": {
			clusterUtil: stubClusterUtil{
				initClusterErr: assert.AnError,
				kubeconfig:     []byte("someKubeconfig"),
			},
			kubeAPIWaiter:    stubKubeAPIWaiter{},
			providerMetadata: &stubProviderMetadata{},
			wantErr:          true,
			k8sVersion:       versions.Default,
		},
		"kubeadm init fails when deploying cilium": {
			clusterUtil:      stubClusterUtil{kubeconfig: []byte("someKubeconfig")},
			providerMetadata: &stubProviderMetadata{},
			wantErr:          true,
			k8sVersion:       versions.Default,
		},
		"kubeadm init fails when setting up constellation-services chart": {
			clusterUtil:      stubClusterUtil{kubeconfig: []byte("someKubeconfig")},
			kubeAPIWaiter:    stubKubeAPIWaiter{},
			providerMetadata: &stubProviderMetadata{},
			wantErr:          true,
			k8sVersion:       versions.Default,
		},
		"kubeadm init fails when reading kubeconfig": {
			clusterUtil:      stubClusterUtil{kubeconfig: []byte("someKubeconfig")},
			kubeAPIWaiter:    stubKubeAPIWaiter{},
			providerMetadata: &stubProviderMetadata{},
			wantErr:          true,
			k8sVersion:       versions.Default,
		},
		"kubeadm init fails when setting up verification service": {
			clusterUtil:      stubClusterUtil{kubeconfig: []byte("someKubeconfig")},
			kubeAPIWaiter:    stubKubeAPIWaiter{},
			providerMetadata: &stubProviderMetadata{},
			wantErr:          true,
			k8sVersion:       versions.Default,
		},
		"kubeadm init fails when waiting for kubeAPI server": {
			clusterUtil:      stubClusterUtil{kubeconfig: []byte("someKubeconfig")},
			kubeAPIWaiter:    stubKubeAPIWaiter{waitErr: assert.AnError},
			providerMetadata: &stubProviderMetadata{},
			k8sVersion:       versions.Default,
			wantErr:          true,
		},
		"unsupported k8sVersion fails cluster creation": {
			clusterUtil:      stubClusterUtil{kubeconfig: []byte("someKubeconfig")},
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
				cloudProvider:    "aws", // provide a valid cloud provider for cilium installation
				clusterUtil:      &tc.clusterUtil,
				providerMetadata: tc.providerMetadata,
				kubeAPIWaiter:    &tc.kubeAPIWaiter,
				configProvider:   &stubConfigProvider{initConfig: k8sapi.KubeadmInitYAML{}},
				client:           &tc.kubectl,
				getIPAddr:        func() (string, error) { return privateIP, nil },
			}

			_, err := kube.InitCluster(
				context.Background(), string(tc.k8sVersion), "kubernetes",
				false, nil, nil, "", logger.NewTest(t),
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
	joinCommand := &kubeadm.BootstrapTokenDiscovery{
		APIServerEndpoint: "192.0.2.0:" + strconv.Itoa(constants.KubernetesPort),
		Token:             "kube-fake-token",
		CACertHashes:      []string{"sha256:a60ebe9b0879090edd83b40a4df4bebb20506bac1e51d518ff8f4505a721930f"},
	}

	privateIP := "192.0.2.1"

	k8sComponents := components.Components{
		{
			Url:         "URL",
			Hash:        "Hash",
			InstallPath: "InstallPath",
			Extract:     true,
		},
	}

	testCases := map[string]struct {
		clusterUtil      stubClusterUtil
		providerMetadata ProviderMetadata
		wantConfig       kubeadm.JoinConfiguration
		role             role.Role
		k8sComponents    components.Components
		wantErr          bool
	}{
		"kubeadm join worker works with metadata and remote Kubernetes Components": {
			clusterUtil: stubClusterUtil{},
			providerMetadata: &stubProviderMetadata{
				selfResp: metadata.InstanceMetadata{
					ProviderID: "provider-id",
					Name:       "metadata-name",
					VPCIP:      "192.0.2.1",
				},
			},
			k8sComponents: k8sComponents,
			role:          role.Worker,
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
		"kubeadm join worker works with metadata and local Kubernetes components": {
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
		"kubeadm join worker fails when installing remote Kubernetes components": {
			clusterUtil: stubClusterUtil{installComponentsErr: errors.New("error")},
			providerMetadata: &stubProviderMetadata{
				selfResp: metadata.InstanceMetadata{
					ProviderID: "provider-id",
					Name:       "metadata-name",
					VPCIP:      "192.0.2.1",
				},
			},
			k8sComponents: k8sComponents,
			role:          role.Worker,
			wantErr:       true,
		},
		"kubeadm join worker fails when retrieving self metadata": {
			clusterUtil: stubClusterUtil{},
			providerMetadata: &stubProviderMetadata{
				selfErr: assert.AnError,
			},
			role:    role.Worker,
			wantErr: true,
		},
		"kubeadm join worker fails when applying the join config": {
			clusterUtil:      stubClusterUtil{joinClusterErr: assert.AnError},
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

			err := kube.JoinCluster(context.Background(), joinCommand, tc.role, tc.k8sComponents, logger.NewTest(t))
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
	testCases := map[string]struct {
		input    string
		expected string
		wantErr  bool
	}{
		"no change": {
			input:    "test",
			expected: "test",
		},
		"uppercase": {
			input:    "TEST",
			expected: "test",
		},
		"underscore": {
			input:    "test_node",
			expected: "test-node",
		},
		"empty": {
			input:    "",
			expected: "",
			wantErr:  true,
		},
		"error": {
			input:    "test_node_",
			expected: "",
			wantErr:  true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			actual, err := k8sCompliantHostname(tc.input)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			assert.Equal(tc.expected, actual)
		})
	}
}

type stubClusterUtil struct {
	installComponentsErr  error
	initClusterErr        error
	setupAutoscalingError error
	setupGCPGuestAgentErr error
	setupOLMErr           error
	setupNMOErr           error
	setupNodeOperatorErr  error
	joinClusterErr        error
	startKubeletErr       error

	kubeconfig []byte

	initConfigs [][]byte
	joinConfigs [][]byte
}

func (s *stubClusterUtil) InstallComponents(_ context.Context, _ components.Components) error {
	return s.installComponentsErr
}

func (s *stubClusterUtil) InitCluster(_ context.Context, initConfig []byte, _, _ string, _ []net.IP, _ bool, _ *slog.Logger) ([]byte, error) {
	s.initConfigs = append(s.initConfigs, initConfig)
	return s.kubeconfig, s.initClusterErr
}

func (s *stubClusterUtil) SetupAutoscaling(_ k8sapi.Client, _ kubernetes.Marshaler, _ kubernetes.Marshaler) error {
	return s.setupAutoscalingError
}

func (s *stubClusterUtil) SetupGCPGuestAgent(_ k8sapi.Client, _ kubernetes.Marshaler) error {
	return s.setupGCPGuestAgentErr
}

func (s *stubClusterUtil) SetupOperatorLifecycleManager(_ context.Context, _ k8sapi.Client, _, _ kubernetes.Marshaler, _ []string) error {
	return s.setupOLMErr
}

func (s *stubClusterUtil) SetupNodeMaintenanceOperator(_ k8sapi.Client, _ kubernetes.Marshaler) error {
	return s.setupNMOErr
}

func (s *stubClusterUtil) SetupNodeOperator(_ context.Context, _ k8sapi.Client, _ kubernetes.Marshaler) error {
	return s.setupNodeOperatorErr
}

func (s *stubClusterUtil) JoinCluster(_ context.Context, joinConfig []byte, _ *slog.Logger) error {
	s.joinConfigs = append(s.joinConfigs, joinConfig)
	return s.joinClusterErr
}

func (s *stubClusterUtil) StartKubelet() error {
	return s.startKubeletErr
}

type stubConfigProvider struct {
	initConfig k8sapi.KubeadmInitYAML
	joinConfig k8sapi.KubeadmJoinYAML
}

func (s *stubConfigProvider) InitConfiguration(_ bool, _ string) k8sapi.KubeadmInitYAML {
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
	createConfigMapErr               error
	addTNodeSelectorsToDeploymentErr error
	waitForCRDsErr                   error
	listAllNamespacesErr             error
	annotateNodeErr                  error
	enforceCoreDNSSpreadErr          error

	listAllNamespacesResp *corev1.NamespaceList
}

func (s *stubKubectl) Initialize(_ []byte) error {
	return nil
}

func (s *stubKubectl) CreateConfigMap(_ context.Context, _ *corev1.ConfigMap) error {
	return s.createConfigMapErr
}

func (s *stubKubectl) AddNodeSelectorsToDeployment(_ context.Context, _ map[string]string, _, _ string) error {
	return s.addTNodeSelectorsToDeploymentErr
}

func (s *stubKubectl) AnnotateNode(_ context.Context, _, _, _ string) error {
	return s.annotateNodeErr
}

func (s *stubKubectl) PatchFirstNodePodCIDR(_ context.Context, _ string) error {
	return nil
}

func (s *stubKubectl) WaitForCRDs(_ context.Context, _ []string) error {
	return s.waitForCRDsErr
}

func (s *stubKubectl) ListAllNamespaces(_ context.Context) (*corev1.NamespaceList, error) {
	return s.listAllNamespacesResp, s.listAllNamespacesErr
}

func (s *stubKubectl) EnforceCoreDNSSpread(_ context.Context) error {
	return s.enforceCoreDNSSpreadErr
}

type stubKubeAPIWaiter struct {
	waitErr error
}

func (s *stubKubeAPIWaiter) Wait(_ context.Context, _ kubewaiter.KubernetesClient) error {
	return s.waitErr
}
