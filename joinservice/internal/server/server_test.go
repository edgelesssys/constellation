/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package server

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/attestation"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/versions/components"
	"github.com/edgelesssys/constellation/v2/joinservice/joinproto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
	kubeadmv1 "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta3"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestIssueJoinTicket(t *testing.T) {
	someErr := errors.New("error")
	testKey := []byte{0x1, 0x2, 0x3}
	testCert := []byte{0x4, 0x5, 0x6}
	measurementSecret := []byte{0x7, 0x8, 0x9}
	uuid := "uuid"

	testJoinToken := &kubeadmv1.BootstrapTokenDiscovery{
		APIServerEndpoint: "192.0.2.1",
		CACertHashes:      []string{"hash"},
		Token:             "token",
	}

	clusterComponents := components.Components{
		{
			URL:         "URL",
			Hash:        "hash",
			InstallPath: "install-path",
			Extract:     true,
		},
	}

	testCases := map[string]struct {
		isControlPlane                 bool
		kubeadm                        stubTokenGetter
		kms                            stubKeyGetter
		ca                             stubCA
		kubeClient                     stubKubeClient
		missingComponentsReferenceFile bool
		wantErr                        bool
	}{
		"worker node": {
			kubeadm: stubTokenGetter{token: testJoinToken},
			kms: stubKeyGetter{dataKeys: map[string][]byte{
				uuid:                                 testKey,
				attestation.MeasurementSecretContext: measurementSecret,
			}},
			ca:         stubCA{cert: testCert, nodeName: "node"},
			kubeClient: stubKubeClient{getComponentsVal: clusterComponents, getK8sComponentsRefFromNodeVersionCRDVal: "k8s-components-ref"},
		},
		"kubeclient fails": {
			kubeadm: stubTokenGetter{token: testJoinToken},
			kms: stubKeyGetter{dataKeys: map[string][]byte{
				uuid:                                 testKey,
				attestation.MeasurementSecretContext: measurementSecret,
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
			}},
			ca:         stubCA{cert: testCert, nodeName: "node"},
			kubeClient: stubKubeClient{getComponentsVal: clusterComponents, getK8sComponentsRefFromNodeVersionCRDVal: "k8s-components-ref"},
			wantErr:    true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			salt := []byte{0xA, 0xB, 0xC}

			api := Server{
				measurementSalt: salt,
				ca:              tc.ca,
				joinTokenGetter: tc.kubeadm,
				dataKeyGetter:   tc.kms,
				kubeClient:      &tc.kubeClient,
				log:             logger.NewTest(t),
			}

			req := &joinproto.IssueJoinTicketRequest{
				DiskUuid:       "uuid",
				IsControlPlane: tc.isControlPlane,
			}
			resp, err := api.IssueJoinTicket(context.Background(), req)
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
			assert.Equal(tc.kubeClient.getComponentsVal.ToJoinProto(), resp.KubernetesComponents)
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
			}

			req := &joinproto.IssueRejoinTicketRequest{
				DiskUuid: uuid,
			}
			resp, err := api.IssueRejoinTicket(context.Background(), req)
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
	getComponentsVal components.Components
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
