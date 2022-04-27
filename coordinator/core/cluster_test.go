package core

import (
	"errors"
	"regexp"
	"testing"

	"github.com/edgelesssys/constellation/cli/file"
	"github.com/edgelesssys/constellation/coordinator/attestation/simulator"
	"github.com/edgelesssys/constellation/coordinator/kubernetes"
	"github.com/edgelesssys/constellation/coordinator/kubernetes/k8sapi/resources"
	"github.com/edgelesssys/constellation/coordinator/role"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	k8s "k8s.io/api/core/v1"
	kubeadm "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta3"
)

func TestInitCluster(t *testing.T) {
	someErr := errors.New("someErr")

	testCases := map[string]struct {
		cluster                clusterStub
		metadata               stubMetadata
		cloudControllerManager stubCloudControllerManager
		cloudNodeManager       stubCloudNodeManager
		clusterAutoscaler      stubClusterAutoscaler
		autoscalingNodeGroups  []string
		wantErr                bool
		wantInitClusterInput   kubernetes.InitClusterInput
	}{
		"InitCluster works": {
			cluster: clusterStub{
				kubeconfig: []byte("kubeconfig"),
			},
			autoscalingNodeGroups: []string{"someNodeGroup"},
			wantInitClusterInput: kubernetes.InitClusterInput{
				APIServerAdvertiseIP:           "10.118.0.1",
				NodeIP:                         "10.118.0.1",
				NodeName:                       "10.118.0.1",
				SupportsCloudControllerManager: false,
				SupportClusterAutoscaler:       false,
				AutoscalingNodeGroups:          []string{"someNodeGroup"},
			},
			wantErr: false,
		},
		"Instance metadata is retrieved": {
			cluster: clusterStub{
				kubeconfig: []byte("kubeconfig"),
			},
			metadata: stubMetadata{
				selfRes: Instance{
					Name:       "some-name",
					ProviderID: "fake://providerid",
				},
				supportedRes: true,
			},
			wantInitClusterInput: kubernetes.InitClusterInput{
				APIServerAdvertiseIP:           "10.118.0.1",
				NodeIP:                         "10.118.0.1",
				NodeName:                       "some-name",
				ProviderID:                     "fake://providerid",
				SupportsCloudControllerManager: false,
				SupportClusterAutoscaler:       false,
			},
			wantErr: false,
		},
		"metadata of self retrieval error is checked": {
			cluster: clusterStub{
				kubeconfig: []byte("kubeconfig"),
			},
			metadata: stubMetadata{
				supportedRes: true,
				selfErr:      errors.New("metadata retrieval error"),
			},
			wantErr: true,
		},
		"Autoscaler is prepared when supported": {
			cluster: clusterStub{
				kubeconfig: []byte("kubeconfig"),
			},
			clusterAutoscaler: stubClusterAutoscaler{
				nameRes:      "some-name",
				supportedRes: true,
			},
			autoscalingNodeGroups: []string{"someNodeGroup"},
			wantInitClusterInput: kubernetes.InitClusterInput{
				APIServerAdvertiseIP:           "10.118.0.1",
				NodeIP:                         "10.118.0.1",
				NodeName:                       "10.118.0.1",
				SupportsCloudControllerManager: false,
				SupportClusterAutoscaler:       true,
				AutoscalingCloudprovider:       "some-name",
				AutoscalingNodeGroups:          []string{"someNodeGroup"},
			},
			wantErr: false,
		},
		"Node is prepared for CCM if supported": {
			cluster: clusterStub{
				kubeconfig: []byte("kubeconfig"),
			},
			cloudControllerManager: stubCloudControllerManager{
				supportedRes: true,
				nameRes:      "some-name",
				imageRes:     "someImage",
				pathRes:      "/some/path",
			},
			wantInitClusterInput: kubernetes.InitClusterInput{
				APIServerAdvertiseIP:           "10.118.0.1",
				NodeIP:                         "10.118.0.1",
				NodeName:                       "10.118.0.1",
				SupportsCloudControllerManager: true,
				SupportClusterAutoscaler:       false,
				CloudControllerManagerName:     "some-name",
				CloudControllerManagerImage:    "someImage",
				CloudControllerManagerPath:     "/some/path",
			},
			wantErr: false,
		},
		"Node preparation for CCM can fail": {
			cluster: clusterStub{
				kubeconfig: []byte("kubeconfig"),
			},
			metadata: stubMetadata{
				supportedRes: true,
			},
			cloudControllerManager: stubCloudControllerManager{
				supportedRes:       true,
				nameRes:            "some-name",
				imageRes:           "someImage",
				pathRes:            "/some/path",
				prepareInstanceRes: errors.New("preparing node for CCM failed"),
			},
			wantErr: true,
		},
		"updating role fails without error": {
			cluster: clusterStub{
				kubeconfig: []byte("kubeconfig"),
			},
			metadata: stubMetadata{
				signalRoleErr: errors.New("updating role fails"),
				supportedRes:  true,
			},
			wantErr: false,
			wantInitClusterInput: kubernetes.InitClusterInput{
				APIServerAdvertiseIP: "10.118.0.1",
				NodeIP:               "10.118.0.1",
			},
		},
		"getting kubeconfig fail detected": {
			cluster: clusterStub{
				getKubeconfigErr: errors.New("getting kubeconfig fails"),
			},
			wantErr: true,
		},
		"InitCluster fail detected": {
			cluster: clusterStub{
				initErr: someErr,
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			zapLogger, err := zap.NewDevelopment()
			require.NoError(err)
			core, err := NewCore(&stubVPN{}, &tc.cluster, &tc.metadata, &tc.cloudControllerManager, &tc.cloudNodeManager, &tc.clusterAutoscaler, nil, zapLogger, simulator.OpenSimulatedTPM, nil, file.NewHandler(afero.NewMemMapFs()))
			require.NoError(err)

			kubeconfig, err := core.InitCluster(tc.autoscalingNodeGroups, "cloud-service-account-uri")

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			require.Len(tc.cluster.initInputs, 1)
			assert.Equal(tc.wantInitClusterInput, tc.cluster.initInputs[0])
			assert.Equal(tc.cluster.kubeconfig, kubeconfig)
		})
	}
}

func TestJoinCluster(t *testing.T) {
	someErr := errors.New("someErr")

	testCases := map[string]struct {
		cluster                clusterStub
		metadata               stubMetadata
		cloudControllerManager stubCloudControllerManager
		cloudNodeManager       stubCloudNodeManager
		clusterAutoscaler      stubClusterAutoscaler
		vpn                    stubVPN
		wantErr                bool
		wantJoinClusterArgs    joinClusterArgs
	}{
		"JoinCluster works": {
			vpn: stubVPN{
				interfaceIP: "192.0.2.0",
			},
			wantJoinClusterArgs: joinClusterArgs{
				args: &kubeadm.BootstrapTokenDiscovery{
					APIServerEndpoint: "192.0.2.0:6443",
					Token:             "someToken",
					CACertHashes:      []string{"someHash"},
				},
				nodeName: "192.0.2.0",
				nodeIP:   "192.0.2.0",
			},
		},
		"JoinCluster fail detected": {
			cluster: clusterStub{
				joinErr: someErr,
			},
			wantErr: true,
		},
		"retrieving vpn ip failure detected": {
			vpn: stubVPN{
				getInterfaceIPErr: errors.New("retrieving interface ip error"),
			},
			wantErr: true,
		},
		"Instance metadata is retrieved": {
			metadata: stubMetadata{
				selfRes: Instance{
					Name:       "some-name",
					ProviderID: "fake://providerid",
				},
				supportedRes: true,
			},
			wantJoinClusterArgs: joinClusterArgs{
				args: &kubeadm.BootstrapTokenDiscovery{
					APIServerEndpoint: "192.0.2.0:6443",
					Token:             "someToken",
					CACertHashes:      []string{"someHash"},
				},
				nodeName:   "some-name",
				providerID: "fake://providerid",
			},
			wantErr: false,
		},
		"Instance metadata retrieval can fail": {
			metadata: stubMetadata{
				supportedRes: true,
				selfErr:      errors.New("metadata retrieval error"),
			},
			wantErr: true,
		},
		"CCM preparation failure is detected": {
			metadata: stubMetadata{
				supportedRes: true,
			},
			cloudControllerManager: stubCloudControllerManager{
				supportedRes:       true,
				prepareInstanceRes: errors.New("ccm prepare fails"),
			},
			wantErr: true,
		},
		"updating role fails without error": {
			metadata: stubMetadata{
				signalRoleErr: errors.New("updating role fails"),
				supportedRes:  true,
			},
			wantErr: false,
			wantJoinClusterArgs: joinClusterArgs{
				args: &kubeadm.BootstrapTokenDiscovery{
					APIServerEndpoint: "192.0.2.0:6443",
					Token:             "someToken",
					CACertHashes:      []string{"someHash"},
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			zapLogger, err := zap.NewDevelopment()
			require.NoError(err)
			core, err := NewCore(&tc.vpn, &tc.cluster, &tc.metadata, &tc.cloudControllerManager, &tc.cloudNodeManager, &tc.clusterAutoscaler, nil, zapLogger, simulator.OpenSimulatedTPM, nil, file.NewHandler(afero.NewMemMapFs()))
			require.NoError(err)

			joinReq := &kubeadm.BootstrapTokenDiscovery{
				APIServerEndpoint: "192.0.2.0:6443",
				Token:             "someToken",
				CACertHashes:      []string{"someHash"},
			}
			err = core.JoinCluster(joinReq, "", role.Node)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			require.Len(tc.cluster.joinClusterArgs, 1)
			assert.Equal(tc.wantJoinClusterArgs, tc.cluster.joinClusterArgs[0])
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

type clusterStub struct {
	initJoinArgs     kubeadm.BootstrapTokenDiscovery
	initErr          error
	joinErr          error
	kubeconfig       []byte
	getKubeconfigErr error

	initInputs      []kubernetes.InitClusterInput
	joinClusterArgs []joinClusterArgs
}

func (c *clusterStub) InitCluster(in kubernetes.InitClusterInput) (*kubeadm.BootstrapTokenDiscovery, error) {
	c.initInputs = append(c.initInputs, in)

	return &c.initJoinArgs, c.initErr
}

func (c *clusterStub) JoinCluster(args *kubeadm.BootstrapTokenDiscovery, nodeName, nodeIP, nodeVPNIP, providerID, certKey string, _ bool, _ role.Role) error {
	c.joinClusterArgs = append(c.joinClusterArgs, joinClusterArgs{
		args:       args,
		nodeName:   nodeName,
		nodeIP:     nodeIP,
		providerID: providerID,
	})

	return c.joinErr
}

func (c *clusterStub) GetKubeconfig() ([]byte, error) {
	return c.kubeconfig, c.getKubeconfigErr
}

func (c *clusterStub) GetKubeadmCertificateKey() (string, error) {
	return "dummy", nil
}

type prepareInstanceRequest struct {
	instance Instance
	vpnIP    string
}

type stubCloudControllerManager struct {
	imageRes           string
	pathRes            string
	nameRes            string
	prepareInstanceRes error
	extraArgsRes       []string
	configMapsRes      resources.ConfigMaps
	configMapsErr      error
	secretsRes         resources.Secrets
	secretsErr         error
	volumesRes         []k8s.Volume
	volumeMountRes     []k8s.VolumeMount
	envRes             []k8s.EnvVar
	supportedRes       bool

	prepareInstanceRequests []prepareInstanceRequest
}

func (s *stubCloudControllerManager) Image() string {
	return s.imageRes
}

func (s *stubCloudControllerManager) Path() string {
	return s.pathRes
}

func (s *stubCloudControllerManager) Name() string {
	return s.nameRes
}

func (s *stubCloudControllerManager) PrepareInstance(instance Instance, vpnIP string) error {
	s.prepareInstanceRequests = append(s.prepareInstanceRequests, prepareInstanceRequest{
		instance: instance,
		vpnIP:    vpnIP,
	})
	return s.prepareInstanceRes
}

func (s *stubCloudControllerManager) ExtraArgs() []string {
	return s.extraArgsRes
}

func (s *stubCloudControllerManager) ConfigMaps(instance Instance) (resources.ConfigMaps, error) {
	return s.configMapsRes, s.configMapsErr
}

func (s *stubCloudControllerManager) Secrets(instance Instance, cloudServiceAccountURI string) (resources.Secrets, error) {
	return s.secretsRes, s.secretsErr
}

func (s *stubCloudControllerManager) Volumes() []k8s.Volume {
	return s.volumesRes
}

func (s *stubCloudControllerManager) VolumeMounts() []k8s.VolumeMount {
	return s.volumeMountRes
}

func (s *stubCloudControllerManager) Env() []k8s.EnvVar {
	return s.envRes
}

func (s *stubCloudControllerManager) Supported() bool {
	return s.supportedRes
}

type stubCloudNodeManager struct {
	imageRes     string
	pathRes      string
	extraArgsRes []string
	supportedRes bool
}

func (s *stubCloudNodeManager) Image() string {
	return s.imageRes
}

func (s *stubCloudNodeManager) Path() string {
	return s.pathRes
}

func (s *stubCloudNodeManager) ExtraArgs() []string {
	return s.extraArgsRes
}

func (s *stubCloudNodeManager) Supported() bool {
	return s.supportedRes
}

type stubClusterAutoscaler struct {
	nameRes        string
	supportedRes   bool
	secretsRes     resources.Secrets
	secretsErr     error
	volumesRes     []k8s.Volume
	volumeMountRes []k8s.VolumeMount
	envRes         []k8s.EnvVar
}

func (s *stubClusterAutoscaler) Name() string {
	return s.nameRes
}

func (s *stubClusterAutoscaler) Secrets(instance Instance, cloudServiceAccountURI string) (resources.Secrets, error) {
	return s.secretsRes, s.secretsErr
}

func (s *stubClusterAutoscaler) Volumes() []k8s.Volume {
	return s.volumesRes
}

func (s *stubClusterAutoscaler) VolumeMounts() []k8s.VolumeMount {
	return s.volumeMountRes
}

func (s *stubClusterAutoscaler) Env() []k8s.EnvVar {
	return s.envRes
}

func (s *stubClusterAutoscaler) Supported() bool {
	return s.supportedRes
}

type joinClusterArgs struct {
	args       *kubeadm.BootstrapTokenDiscovery
	nodeName   string
	nodeIP     string
	providerID string
}
