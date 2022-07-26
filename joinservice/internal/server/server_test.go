package server

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/internal/attestation"
	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/edgelesssys/constellation/internal/file"
	"github.com/edgelesssys/constellation/internal/logger"
	"github.com/edgelesssys/constellation/internal/versions"
	"github.com/edgelesssys/constellation/joinservice/joinproto"
	"github.com/spf13/afero"
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
	testK8sVersion := versions.Latest

	testCases := map[string]struct {
		isControlPlane bool
		kubeadm        stubTokenGetter
		kms            stubKeyGetter
		ca             stubCA
		wantErr        bool
	}{
		"worker node": {
			kubeadm: stubTokenGetter{token: testJoinToken},
			kms: stubKeyGetter{dataKeys: map[string][]byte{
				uuid:                                 testKey,
				attestation.MeasurementSecretContext: measurementSecret,
			}},
			ca: stubCA{cert: testCert},
		},
		"GetDataKey fails": {
			kubeadm: stubTokenGetter{token: testJoinToken},
			kms:     stubKeyGetter{dataKeys: make(map[string][]byte), getDataKeyErr: someErr},
			ca:      stubCA{cert: testCert},
			wantErr: true,
		},
		"GetJoinToken fails": {
			kubeadm: stubTokenGetter{getJoinTokenErr: someErr},
			kms: stubKeyGetter{dataKeys: map[string][]byte{
				uuid:                                 testKey,
				attestation.MeasurementSecretContext: measurementSecret,
			}},
			ca:      stubCA{cert: testCert},
			wantErr: true,
		},
		"GetCertificate fails": {
			kubeadm: stubTokenGetter{token: testJoinToken},
			kms: stubKeyGetter{dataKeys: map[string][]byte{
				uuid:                                 testKey,
				attestation.MeasurementSecretContext: measurementSecret,
			}},
			ca:      stubCA{getCertErr: someErr},
			wantErr: true,
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
			ca: stubCA{cert: testCert},
		},
		"GetControlPlaneCertificateKey fails": {
			isControlPlane: true,
			kubeadm:        stubTokenGetter{token: testJoinToken, certificateKeyErr: someErr},
			kms: stubKeyGetter{dataKeys: map[string][]byte{
				uuid:                                 testKey,
				attestation.MeasurementSecretContext: measurementSecret,
			}},
			ca:      stubCA{cert: testCert},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			handler := file.NewHandler(afero.NewMemMapFs())
			// IssueJoinTicket tries to read the k8s-version ConfigMap from a mounted file.
			require.NoError(handler.Write(filepath.Join(constants.ServiceBasePath, constants.K8sVersion), []byte(testK8sVersion), file.OptNone))
			salt := []byte{0xA, 0xB, 0xC}

			api := New(
				salt,
				handler,
				tc.ca,
				tc.kubeadm,
				tc.kms,
				logger.NewTest(t),
			)

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

			api := New(
				nil,
				file.Handler{},
				stubCA{},
				stubTokenGetter{},
				tc.keyGetter,
				logger.NewTest(t),
			)

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
}

func (f stubCA) GetCertificate(csr []byte) ([]byte, error) {
	return f.cert, f.getCertErr
}
