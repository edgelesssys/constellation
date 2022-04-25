package kubernetes

import (
	"errors"
	"testing"

	"github.com/edgelesssys/constellation/coordinator/kubernetes/k8sapi"
	"github.com/edgelesssys/constellation/coordinator/kubernetes/k8sapi/resources"
	"github.com/edgelesssys/constellation/coordinator/role"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
	kubeadm "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta3"
	"sigs.k8s.io/yaml"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m,
		// https://github.com/kubernetes/klog/issues/282, https://github.com/kubernetes/klog/issues/188
		goleak.IgnoreTopFunction("k8s.io/klog/v2.(*loggingT).flushDaemon"),
	)
}

type stubClusterUtil struct {
	joinClusterRequest               *kubeadm.BootstrapTokenDiscovery
	initClusterErr                   error
	setupPodNetworkErr               error
	setupAutoscalingError            error
	setupCloudControllerManagerError error
	setupCloudNodeManagerError       error
	joinClusterErr                   error
	restartKubeletErr                error

	initConfigs [][]byte
	joinConfigs [][]byte
}

func (s *stubClusterUtil) InitCluster(initConfig []byte) (*kubeadm.BootstrapTokenDiscovery, error) {
	s.initConfigs = append(s.initConfigs, initConfig)
	return s.joinClusterRequest, s.initClusterErr
}

func (s *stubClusterUtil) SetupPodNetwork(kubectl k8sapi.Client, podNetworkConfiguration resources.Marshaler) error {
	return s.setupPodNetworkErr
}

func (s *stubClusterUtil) SetupAutoscaling(kubectl k8sapi.Client, clusterAutoscalerConfiguration resources.Marshaler, secrets resources.Marshaler) error {
	return s.setupAutoscalingError
}

func (s *stubClusterUtil) SetupCloudControllerManager(kubectl k8sapi.Client, cloudControllerManagerConfiguration resources.Marshaler, configMaps resources.Marshaler, secrets resources.Marshaler) error {
	return s.setupCloudControllerManagerError
}

func (s *stubClusterUtil) SetupCloudNodeManager(kubectl k8sapi.Client, cloudNodeManagerConfiguration resources.Marshaler) error {
	return s.setupCloudNodeManagerError
}

func (s *stubClusterUtil) JoinCluster(joinConfig []byte) error {
	s.joinConfigs = append(s.joinConfigs, joinConfig)
	return s.joinClusterErr
}

func (s *stubClusterUtil) RestartKubelet() error {
	return s.restartKubeletErr
}

func (s *stubClusterUtil) GetControlPlaneJoinCertificateKey() (string, error) {
	return "", nil
}

type stubConfigProvider struct {
	InitConfig k8sapi.KubeadmInitYAML
	JoinConfig k8sapi.KubeadmJoinYAML
}

func (s *stubConfigProvider) InitConfiguration() k8sapi.KubeadmInitYAML {
	return s.InitConfig
}

func (s *stubConfigProvider) JoinConfiguration() k8sapi.KubeadmJoinYAML {
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

func TestInitCluster(t *testing.T) {
	someErr := errors.New("failed")
	coordinatorVPNIP := "192.0.2.0"
	coordinatorProviderID := "somecloudprovider://instance-id"
	instanceName := "instance-id"
	supportsClusterAutoscaler := false
	cloudprovider := "some-cloudprovider"
	cloudControllerManagerImage := "some-image:latest"
	cloudControllerManagerPath := "/some_path"
	autoscalingNodeGroups := []string{"0,10,autoscaling_group_0"}

	testCases := map[string]struct {
		clusterUtil      stubClusterUtil
		kubeCTL          stubKubeCTL
		kubeconfigReader stubKubeconfigReader
		initConfig       k8sapi.KubeadmInitYAML
		joinConfig       k8sapi.KubeadmJoinYAML
		expectErr        bool
	}{
		"kubeadm init works": {
			clusterUtil: stubClusterUtil{
				joinClusterRequest: &kubeadm.BootstrapTokenDiscovery{
					APIServerEndpoint: "192.0.2.0",
					Token:             "kube-fake-token",
					CACertHashes:      []string{"sha256:a60ebe9b0879090edd83b40a4df4bebb20506bac1e51d518ff8f4505a721930f"},
				},
			},
			kubeconfigReader: stubKubeconfigReader{
				Kubeconfig: []byte("someKubeconfig"),
			},
			expectErr: false,
		},
		"kubeadm init errors": {
			clusterUtil: stubClusterUtil{
				joinClusterRequest: nil,
				initClusterErr:     someErr,
			},
			kubeconfigReader: stubKubeconfigReader{
				Kubeconfig: []byte("someKubeconfig"),
			},
			expectErr: true,
		},
		"pod network setup errors": {
			clusterUtil: stubClusterUtil{
				joinClusterRequest: nil,
				setupPodNetworkErr: someErr,
			},
			kubeconfigReader: stubKubeconfigReader{
				Kubeconfig: []byte("someKubeconfig"),
			},
			expectErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			kube := KubeWrapper{
				clusterUtil:      &tc.clusterUtil,
				configProvider:   &stubConfigProvider{InitConfig: k8sapi.KubeadmInitYAML{}},
				client:           &tc.kubeCTL,
				kubeconfigReader: &tc.kubeconfigReader,
			}
			joinCommand, err := kube.InitCluster(
				InitClusterInput{
					APIServerAdvertiseIP:        coordinatorVPNIP,
					NodeName:                    instanceName,
					ProviderID:                  coordinatorProviderID,
					SupportClusterAutoscaler:    supportsClusterAutoscaler,
					AutoscalingCloudprovider:    cloudprovider,
					AutoscalingNodeGroups:       autoscalingNodeGroups,
					CloudControllerManagerName:  cloudprovider,
					CloudControllerManagerImage: cloudControllerManagerImage,
					CloudControllerManagerPath:  cloudControllerManagerPath,
				},
			)

			if tc.expectErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.clusterUtil.joinClusterRequest, joinCommand)

			var kubeadmConfig k8sapi.KubeadmInitYAML
			require.NoError(resources.UnmarshalK8SResources(tc.clusterUtil.initConfigs[0], &kubeadmConfig))
			assert.Equal(kubeadmConfig.InitConfiguration.LocalAPIEndpoint.AdvertiseAddress, "192.0.2.0")
			assert.Equal(kubeadmConfig.ClusterConfiguration.Networking.PodSubnet, "10.244.0.0/16")
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
	coordinatorProviderID := "somecloudprovider://instance-id"
	instanceName := "instance-id"
	client := fakeK8SClient{}

	testCases := map[string]struct {
		clusterUtil stubClusterUtil
		expectErr   bool
	}{
		"kubeadm join works": {
			clusterUtil: stubClusterUtil{},
			expectErr:   false,
		},
		"kubeadm join errors": {
			clusterUtil: stubClusterUtil{joinClusterErr: someErr},
			expectErr:   true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			kube := New(&tc.clusterUtil, &stubConfigProvider{}, &client)
			err := kube.JoinCluster(joinCommand, instanceName, nodeVPNIP, nodeVPNIP, coordinatorProviderID, "", role.Node)
			if tc.expectErr {
				assert.Error(err)
				return
			}
			require.NoError(err)

			var joinConfig kubeadm.JoinConfiguration
			require.NoError(yaml.Unmarshal(tc.clusterUtil.joinConfigs[0], &joinConfig))

			assert.Equal("192.0.2.0:6443", joinConfig.Discovery.BootstrapToken.APIServerEndpoint)
			assert.Equal("kube-fake-token", joinConfig.Discovery.BootstrapToken.Token)
			assert.Equal([]string{"sha256:a60ebe9b0879090edd83b40a4df4bebb20506bac1e51d518ff8f4505a721930f"}, joinConfig.Discovery.BootstrapToken.CACertHashes)
			assert.Equal(map[string]string{"node-ip": "192.0.2.0"}, joinConfig.NodeRegistration.KubeletExtraArgs)
		})
	}
}

func TestGetKubeconfig(t *testing.T) {
	testCases := map[string]struct {
		Kubewrapper KubeWrapper
		expectErr   bool
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
