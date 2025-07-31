/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: BUSL-1.1
*/

package server

import (
	"context"
	"crypto/ed25519"
	"errors"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/attestation"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/versions/components"
	"github.com/edgelesssys/constellation/v2/joinservice/joinproto"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
	"golang.org/x/crypto/ssh"
	kubeadmv1 "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta3"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m, goleak.IgnoreAnyFunction("github.com/bazelbuild/rules_go/go/tools/bzltestutil.RegisterTimeoutHandler.func1"))
}

func TestIssueJoinTicket(t *testing.T) {
	someErr := errors.New("error")
	testKey := []byte{0x1, 0x2, 0x3}
	testCaKey := make([]byte, ed25519.SeedSize)
	testCert := []byte{0x4, 0x5, 0x6}
	measurementSecret := []byte{0x7, 0x8, 0x9}
	uuid := "uuid"

	pubkey, _, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)
	hostSSHPubKey, err := ssh.NewPublicKey(pubkey)
	require.NoError(t, err)

	testJoinToken := &kubeadmv1.BootstrapTokenDiscovery{
		APIServerEndpoint: "192.0.2.1",
		CACertHashes:      []string{"hash"},
		Token:             "token",
	}

	clusterComponents := components.Components{
		{
			Url:         "URL",
			Hash:        "hash",
			InstallPath: "install-path",
			Extract:     true,
		},
	}

	testCases := map[string]struct {
		isControlPlane                  bool
		kubeadm                         stubTokenGetter
		kms                             stubKeyGetter
		ca                              stubCA
		kubeClient                      stubKubeClient
		missingComponentsReferenceFile  bool
		missingAdditionalPrincipalsFile bool
		missingSSHHostKey               bool
		wantErr                         bool
	}{
		"worker node": {
			kubeadm: stubTokenGetter{token: testJoinToken},
			kms: stubKeyGetter{dataKeys: map[string][]byte{
				uuid:                                 testKey,
				attestation.MeasurementSecretContext: measurementSecret,
				constants.SSHCAKeySuffix:             testCaKey,
			}},
			ca:         stubCA{cert: testCert, nodeName: "node"},
			kubeClient: stubKubeClient{getComponentsVal: clusterComponents, getK8sComponentsRefFromNodeVersionCRDVal: "k8s-components-ref"},
		},
		"kubeclient fails": {
			kubeadm: stubTokenGetter{token: testJoinToken},
			kms: stubKeyGetter{dataKeys: map[string][]byte{
				uuid:                                 testKey,
				attestation.MeasurementSecretContext: measurementSecret,
				constants.SSHCAKeySuffix:             testCaKey,
			}},
			ca:         stubCA{cert: testCert, nodeName: "node"},
			kubeClient: stubKubeClient{getComponentsErr: someErr},
			wantErr:    true,
		},
		"Getting Node Name from CSR fails": {
			kubeadm: stubTokenGetter{token: testJoinToken},
			kms: stubKeyGetter{dataKeys: map[string][]byte{
				uuid:                                 testKey,
				attestation.MeasurementSecretContext: measurementSecret,
				constants.SSHCAKeySuffix:             testCaKey,
			}},
			ca:         stubCA{cert: testCert, nodeName: "node", getNameErr: someErr},
			kubeClient: stubKubeClient{getComponentsVal: clusterComponents, getK8sComponentsRefFromNodeVersionCRDVal: "k8s-components-ref"},
			wantErr:    true,
		},
		"Cannot add node to JoiningNode CRD": {
			kubeadm: stubTokenGetter{token: testJoinToken},
			kms: stubKeyGetter{dataKeys: map[string][]byte{
				uuid:                                 testKey,
				attestation.MeasurementSecretContext: measurementSecret,
				constants.SSHCAKeySuffix:             testCaKey,
			}},
			ca:         stubCA{cert: testCert, nodeName: "node"},
			kubeClient: stubKubeClient{getComponentsVal: clusterComponents, addNodeToJoiningNodesErr: someErr, getK8sComponentsRefFromNodeVersionCRDVal: "k8s-components-ref"},
			wantErr:    true,
		},
		"GetDataKey fails": {
			kubeadm:    stubTokenGetter{token: testJoinToken},
			kms:        stubKeyGetter{dataKeys: make(map[string][]byte), getDataKeyErr: someErr},
			ca:         stubCA{cert: testCert, nodeName: "node"},
			kubeClient: stubKubeClient{getComponentsVal: clusterComponents, getK8sComponentsRefFromNodeVersionCRDVal: "k8s-components-ref"},
			wantErr:    true,
		},
		"GetJoinToken fails": {
			kubeadm: stubTokenGetter{getJoinTokenErr: someErr},
			kms: stubKeyGetter{dataKeys: map[string][]byte{
				uuid:                                 testKey,
				attestation.MeasurementSecretContext: measurementSecret,
				constants.SSHCAKeySuffix:             testCaKey,
			}},
			ca:         stubCA{cert: testCert, nodeName: "node"},
			kubeClient: stubKubeClient{getComponentsVal: clusterComponents, getK8sComponentsRefFromNodeVersionCRDVal: "k8s-components-ref"},
			wantErr:    true,
		},
		"GetCertificate fails": {
			kubeadm: stubTokenGetter{token: testJoinToken},
			kms: stubKeyGetter{dataKeys: map[string][]byte{
				uuid:                                 testKey,
				attestation.MeasurementSecretContext: measurementSecret,
				constants.SSHCAKeySuffix:             testCaKey,
			}},
			ca:         stubCA{getCertErr: someErr, nodeName: "node"},
			kubeClient: stubKubeClient{getComponentsVal: clusterComponents, getK8sComponentsRefFromNodeVersionCRDVal: "k8s-components-ref"},
			wantErr:    true,
		},
		"control plane": {
			isControlPlane: true,
			kubeadm: stubTokenGetter{
				token: testJoinToken,
				files: map[string][]byte{"test": {0x1, 0x2, 0x3}},
			},
			kms: stubKeyGetter{dataKeys: map[string][]byte{
				uuid:                                 testKey,
				attestation.MeasurementSecretContext: measurementSecret,
				constants.SSHCAKeySuffix:             testCaKey,
			}},
			ca:         stubCA{cert: testCert, nodeName: "node"},
			kubeClient: stubKubeClient{getComponentsVal: clusterComponents, getK8sComponentsRefFromNodeVersionCRDVal: "k8s-components-ref"},
		},
		"GetControlPlaneCertificateKey fails": {
			isControlPlane: true,
			kubeadm:        stubTokenGetter{token: testJoinToken, certificateKeyErr: someErr},
			kms: stubKeyGetter{dataKeys: map[string][]byte{
				uuid:                                 testKey,
				attestation.MeasurementSecretContext: measurementSecret,
				constants.SSHCAKeySuffix:             testCaKey,
			}},
			ca:         stubCA{cert: testCert, nodeName: "node"},
			kubeClient: stubKubeClient{getComponentsVal: clusterComponents, getK8sComponentsRefFromNodeVersionCRDVal: "k8s-components-ref"},
			wantErr:    true,
		},
		"CA data key to short": {
			kubeadm: stubTokenGetter{token: testJoinToken},
			kms: stubKeyGetter{dataKeys: map[string][]byte{
				uuid:                                 testKey,
				attestation.MeasurementSecretContext: measurementSecret,
				constants.SSHCAKeySuffix:             testKey,
			}},
			ca:         stubCA{cert: testCert, nodeName: "node"},
			kubeClient: stubKubeClient{getComponentsVal: clusterComponents, getK8sComponentsRefFromNodeVersionCRDVal: "k8s-components-ref"},
			wantErr:    true,
		},
		"CA data key doesn't exist": {
			kubeadm: stubTokenGetter{token: testJoinToken},
			kms: stubKeyGetter{dataKeys: map[string][]byte{
				uuid:                                 testKey,
				attestation.MeasurementSecretContext: measurementSecret,
			}},
			ca:         stubCA{cert: testCert, nodeName: "node"},
			kubeClient: stubKubeClient{getComponentsVal: clusterComponents, getK8sComponentsRefFromNodeVersionCRDVal: "k8s-components-ref"},
			wantErr:    true,
		},
		"Additional principals file is missing": {
			kubeadm: stubTokenGetter{token: testJoinToken},
			kms: stubKeyGetter{dataKeys: map[string][]byte{
				uuid:                                 testKey,
				attestation.MeasurementSecretContext: measurementSecret,
				constants.SSHCAKeySuffix:             testCaKey,
			}},
			ca:                              stubCA{cert: testCert, nodeName: "node"},
			kubeClient:                      stubKubeClient{getComponentsVal: clusterComponents, getK8sComponentsRefFromNodeVersionCRDVal: "k8s-components-ref"},
			missingAdditionalPrincipalsFile: true,
		},
		"Host pubkey is missing": {
			kubeadm: stubTokenGetter{token: testJoinToken},
			kms: stubKeyGetter{dataKeys: map[string][]byte{
				uuid:                                 testKey,
				attestation.MeasurementSecretContext: measurementSecret,
				constants.SSHCAKeySuffix:             testCaKey,
			}},
			ca:                stubCA{cert: testCert, nodeName: "node"},
			kubeClient:        stubKubeClient{getComponentsVal: clusterComponents, getK8sComponentsRefFromNodeVersionCRDVal: "k8s-components-ref"},
			missingSSHHostKey: true,
			wantErr:           true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			salt := []byte{0xA, 0xB, 0xC}

			fh := file.NewHandler(afero.NewMemMapFs())
			if !tc.missingAdditionalPrincipalsFile {
				require.NoError(fh.Write("/var/kubeadm-config/ClusterConfiguration", []byte(clusterConfig), file.OptMkdirAll))
			}

			api := Server{
				measurementSalt: salt,
				ca:              tc.ca,
				joinTokenGetter: tc.kubeadm,
				dataKeyGetter:   tc.kms,
				kubeClient:      &tc.kubeClient,
				log:             logger.NewTest(t),
				fileHandler:     fh,
			}

			var keyToSend []byte
			if tc.missingSSHHostKey {
				keyToSend = nil
			} else {
				keyToSend = hostSSHPubKey.Marshal()
			}

			req := &joinproto.IssueJoinTicketRequest{
				DiskUuid:       "uuid",
				IsControlPlane: tc.isControlPlane,
				HostPublicKey:  keyToSend,
			}
			resp, err := api.IssueJoinTicket(t.Context(), req)
			if tc.wantErr {
				assert.Error(err)
				return
			}

			require.NoError(err)
			assert.Equal(tc.kms.dataKeys[uuid], resp.StateDiskKey)
			assert.Equal(salt, resp.MeasurementSalt)
			assert.Equal(tc.kms.dataKeys[attestation.MeasurementSecretContext], resp.MeasurementSecret)
			assert.Equal(tc.kubeadm.token.APIServerEndpoint, resp.ApiServerEndpoint)
			assert.Equal(tc.kubeadm.token.CACertHashes[0], resp.DiscoveryTokenCaCertHash)
			assert.Equal(tc.kubeadm.token.Token, resp.Token)
			assert.Equal(tc.ca.cert, resp.KubeletCert)
			assert.Equal(tc.kubeClient.getComponentsVal, resp.KubernetesComponents)
			assert.Equal(tc.ca.nodeName, tc.kubeClient.joiningNodeName)
			assert.Equal(tc.kubeClient.getK8sComponentsRefFromNodeVersionCRDVal, tc.kubeClient.componentsRef)

			if tc.isControlPlane {
				assert.Len(resp.ControlPlaneFiles, len(tc.kubeadm.files))
			}
		})
	}
}

func TestIssueRejoinTicker(t *testing.T) {
	uuid := "uuid"

	testCases := map[string]struct {
		keyGetter stubKeyGetter
		wantErr   bool
	}{
		"success": {
			keyGetter: stubKeyGetter{
				dataKeys: map[string][]byte{
					uuid:                                 {0x1, 0x2, 0x3},
					attestation.MeasurementSecretContext: {0x4, 0x5, 0x6},
				},
			},
		},
		"failure": {
			keyGetter: stubKeyGetter{
				dataKeys:      make(map[string][]byte),
				getDataKeyErr: errors.New("error"),
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			api := Server{
				ca:              stubCA{},
				joinTokenGetter: stubTokenGetter{},
				dataKeyGetter:   tc.keyGetter,
				log:             logger.NewTest(t),
				fileHandler:     file.NewHandler(afero.NewMemMapFs()),
			}

			req := &joinproto.IssueRejoinTicketRequest{
				DiskUuid: uuid,
			}
			resp, err := api.IssueRejoinTicket(t.Context(), req)
			if tc.wantErr {
				assert.Error(err)
				return
			}

			require.NoError(err)
			assert.Equal(tc.keyGetter.dataKeys[attestation.MeasurementSecretContext], resp.MeasurementSecret)
			assert.Equal(tc.keyGetter.dataKeys[uuid], resp.StateDiskKey)
		})
	}
}

type stubTokenGetter struct {
	token             *kubeadmv1.BootstrapTokenDiscovery
	getJoinTokenErr   error
	files             map[string][]byte
	certificateKeyErr error
}

func (f stubTokenGetter) GetJoinToken(time.Duration) (*kubeadmv1.BootstrapTokenDiscovery, error) {
	return f.token, f.getJoinTokenErr
}

func (f stubTokenGetter) GetControlPlaneCertificatesAndKeys() (map[string][]byte, error) {
	return f.files, f.certificateKeyErr
}

type stubKeyGetter struct {
	dataKeys      map[string][]byte
	getDataKeyErr error
}

func (f stubKeyGetter) GetDataKey(_ context.Context, name string, _ int) ([]byte, error) {
	return f.dataKeys[name], f.getDataKeyErr
}

type stubCA struct {
	cert       []byte
	getCertErr error
	nodeName   string
	getNameErr error
}

func (f stubCA) GetCertificate(_ []byte) ([]byte, error) {
	return f.cert, f.getCertErr
}

func (f stubCA) GetNodeNameFromCSR(_ []byte) (string, error) {
	return f.nodeName, f.getNameErr
}

type stubKubeClient struct {
	getComponentsVal []*components.Component
	getComponentsErr error

	getK8sComponentsRefFromNodeVersionCRDErr error
	getK8sComponentsRefFromNodeVersionCRDVal string

	addNodeToJoiningNodesErr error
	joiningNodeName          string
	componentsRef            string
}

func (s *stubKubeClient) GetK8sComponentsRefFromNodeVersionCRD(_ context.Context, _ string) (string, error) {
	return s.getK8sComponentsRefFromNodeVersionCRDVal, s.getK8sComponentsRefFromNodeVersionCRDErr
}

func (s *stubKubeClient) GetComponents(_ context.Context, _ string) (components.Components, error) {
	return s.getComponentsVal, s.getComponentsErr
}

func (s *stubKubeClient) AddNodeToJoiningNodes(_ context.Context, nodeName string, componentsRef string, _ bool) error {
	s.joiningNodeName = nodeName
	s.componentsRef = componentsRef
	return s.addNodeToJoiningNodesErr
}

const clusterConfig = `
apiServer:
  certSANs:
  - "*"
  extraArgs:
  - name: audit-log-maxage
    value: "30"
  - name: audit-log-maxbackup
    value: "10"
  - name: audit-log-maxsize
    value: "100"
  - name: audit-log-path
    value: /var/log/kubernetes/audit/audit.log
  - name: audit-policy-file
    value: /etc/kubernetes/audit-policy.yaml
  - name: kubelet-certificate-authority
    value: /etc/kubernetes/pki/ca.crt
  - name: profiling
    value: "false"
  - name: tls-cipher-suites
    value: TLS_AES_128_GCM_SHA256,TLS_AES_256_GCM_SHA384,TLS_CHACHA20_POLY1305_SHA256,TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA,TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,TLS_RSA_WITH_3DES_EDE_CBC_SHA,TLS_RSA_WITH_AES_128_CBC_SHA,TLS_RSA_WITH_AES_128_GCM_SHA256,TLS_RSA_WITH_AES_256_CBC_SHA,TLS_RSA_WITH_AES_256_GCM_SHA384
  extraVolumes:
  - hostPath: /var/log/kubernetes/audit/
    mountPath: /var/log/kubernetes/audit/
    name: audit-log
    pathType: DirectoryOrCreate
  - hostPath: /etc/kubernetes/audit-policy.yaml
    mountPath: /etc/kubernetes/audit-policy.yaml
    name: audit
    pathType: File
    readOnly: true
apiVersion: kubeadm.k8s.io/v1beta4
caCertificateValidityPeriod: 87600h0m0s
certificateValidityPeriod: 8760h0m0s
certificatesDir: /etc/kubernetes/pki
clusterName: mr-cilium-7d6460ea
controlPlaneEndpoint: 34.8.0.20:6443
controllerManager:
  extraArgs:
  - name: cloud-provider
    value: external
  - name: configure-cloud-routes
    value: "false"
  - name: flex-volume-plugin-dir
    value: /opt/libexec/kubernetes/kubelet-plugins/volume/exec/
  - name: profiling
    value: "false"
  - name: terminated-pod-gc-threshold
    value: "1000"
dns: {}
encryptionAlgorithm: RSA-2048
etcd:
  local:
    dataDir: /var/lib/etcd
imageRepository: registry.k8s.io
kind: ClusterConfiguration
kubernetesVersion: v1.30.14
networking:
  dnsDomain: cluster.local
  serviceSubnet: 10.96.0.0/12
proxy: {}
scheduler:
  extraArgs:
  - name: profiling
    value: "false"
`
