/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package kubeadm

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
	"k8s.io/client-go/kubernetes/fake"
	kubeconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m, goleak.IgnoreAnyFunction("github.com/bazelbuild/rules_go/go/tools/bzltestutil.RegisterTimeoutHandler.func1"))
}

func TestGetJoinToken(t *testing.T) {
	validConf := `apiVersion: v1
kind: Config
clusters:
- cluster:
    certificate-authority-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUMvakNDQWVhZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFWTVJNd0VRWURWUVFERXdwcmRXSmwKY201bGRHVnpNQjRYRFRJeU1EVXpNREE0TWpJd01Gb1hEVE15TURVeU56QTRNakl3TUZvd0ZURVRNQkVHQTFVRQpBeE1LYTNWaVpYSnVaWFJsY3pDQ0FTSXdEUVlKS29aSWh2Y05BUUVCQlFBRGdnRVBBRENDQVFvQ2dnRUJBTmV5CnNubVJQbDYxaXZGWWRIUjFJUjdyRS9PNjNSOVhpVERwM1V4T2tMQzdMaW94bFA0SmRINzdHMUJ4Y2NCSjVISDIKZHBUTklzcjNxMEZ3ckdtK1JVYzdoRjBmZjgwdUtyUVVMN3UrYWlIRU5HSExVSFVnc3V4Tmd1bUxRdnlrRTUzNQp4dWRVSWpVV0g5M3NuRU5GempuWkRZM09SWVdNQ253OVlxMk5CZDdBRktKY1o3WDc3U1I3eStNK3czdGkvQlZpCmNtR1BvRW1WTTV3V0VReFQwYlpxNjcxTXltcmhEenFwbEZ2dkpranFIdVp6dUFhZ0pXWW9nejNsYjZLbCtmdmgKTjBjbFBDMjJyUUJJY01JWDVHdG40bzJ5U2JvQnBoRWNEWkx6TjIyU0tZZ2ViSGQwOU9lcktWdGw5bDl6cmQvVApBWm5jOTNQVCtvWTFsSmdldUE4Q0F3RUFBYU5aTUZjd0RnWURWUjBQQVFIL0JBUURBZ0trTUE4R0ExVWRFd0VCCi93UUZNQU1CQWY4d0hRWURWUjBPQkJZRUZOVmNPNUZZY2NUTVN1SHpJWFZMYlppUnZRVVZNQlVHQTFVZEVRUU8KTUF5Q0NtdDFZbVZ5Ym1WMFpYTXdEUVlKS29aSWh2Y05BUUVMQlFBRGdnRUJBTDBsRERnbEsvY1JCNHVoQXJBRwpRSDhOeGtCSnhGNXYrVWMyZVFGa3dRTlB5SkU3QTRMV2p1eEVLN25meWVrTk91c2N2Wm1DQzVJNFhVZHAzb0ptCnZzSVlsN2YvMEFaMUt3d1RvQSt3cFF2QVB1NHlhM251MkZkMC9DVkViazNUZTV1MzRmQkxvL0YzK0Q2dFZLb2gKbVpGYmdoVjdMZms5SlQ4UzZjbGxyYjZkT3dCdGViUDBMQWZJd0hWaDBZNEsyY0thc3ZtU2xtMktpRXdURlBrbgpTSkNWWnI1aUJ3eGFadk1mYlpEaDk1bGZCbEtCVkdMNm5CcWs2TEpKM0VVd0tocTFGZEoyT0lSTkF0em14Z0R3CnNkOWd0SE4rK0pUcnhDa0ZBUTdwVWptdXBjZmpDOWhRRk1HOTRzTzk5elhZd2svTEdhV3FlS0pBYlRiNVdoRWcKYU5ZPQotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg==
    server: https://127.0.0.1:16443
  name: kubernetes
contexts:
- context:
    cluster: kubernetes
    user: kubernetes-admin
  name: kubernetes-admin@kubernetes
current-context: kubernetes-admin@kubernetes`

	missingCA := `apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://127.0.0.1:16443
  name: kubernetes
contexts:
- context:
    cluster: kubernetes
    user: kubernetes-admin
  name: kubernetes-admin@kubernetes
current-context: kubernetes-admin@kubernetes`

	testCases := map[string]struct {
		adminConf string
		wantErr   bool
	}{
		"success": {
			adminConf: validConf,
		},
		"no certificate-authority-data": {
			adminConf: missingCA,
			wantErr:   true,
		},
		"no cluster config": {
			adminConf: `apiVersion: v1
kind: Config`,
			wantErr: true,
		},
		"invalid config": {
			adminConf: "not a config",
			wantErr:   true,
		},
		"config does not exist": {
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			client := &Kubeadm{
				log:    logger.NewTest(t),
				file:   file.NewHandler(afero.NewMemMapFs()),
				client: fake.NewSimpleClientset(),
			}
			if tc.adminConf != "" {
				require.NoError(client.file.Write(constants.ControlPlaneAdminConfFilename, []byte(tc.adminConf), file.OptNone))
			}

			res, err := client.GetJoinToken(time.Minute)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.NotNil(res)
			}
		})
	}
}

func TestGetControlPlaneCertificatesAndKeys(t *testing.T) {
	someData := []byte{0x1, 0x2, 0x3}

	testCases := map[string]struct {
		preExistingFiles map[string][]byte
		wantErr          bool
	}{
		"success": {
			preExistingFiles: map[string][]byte{
				kubeconstants.CAKeyName:                    someData,
				kubeconstants.ServiceAccountPrivateKeyName: someData,
				kubeconstants.FrontProxyCAKeyName:          someData,
				kubeconstants.EtcdCAKeyName:                someData,
				kubeconstants.CACertName:                   someData,
				kubeconstants.ServiceAccountPublicKeyName:  someData,
				kubeconstants.FrontProxyCACertName:         someData,
				kubeconstants.EtcdCACertName:               someData,
			},
		},
		"missing key": {
			preExistingFiles: map[string][]byte{
				kubeconstants.CACertName:                  someData,
				kubeconstants.ServiceAccountPublicKeyName: someData,
				kubeconstants.FrontProxyCACertName:        someData,
				kubeconstants.EtcdCACertName:              someData,
			},
			wantErr: true,
		},
		"missing cert": {
			preExistingFiles: map[string][]byte{
				kubeconstants.CAKeyName:                    someData,
				kubeconstants.ServiceAccountPrivateKeyName: someData,
				kubeconstants.FrontProxyCAKeyName:          someData,
				kubeconstants.EtcdCAKeyName:                someData,
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			client := &Kubeadm{
				log:    logger.NewTest(t),
				file:   file.NewHandler(afero.NewMemMapFs()),
				client: fake.NewSimpleClientset(),
			}
			for filename, content := range tc.preExistingFiles {
				require.NoError(client.file.Write(
					filepath.Join(kubeconstants.KubernetesDir, kubeconstants.DefaultCertificateDir, filename),
					content,
					file.OptNone,
				))
			}

			files, err := client.GetControlPlaneCertificatesAndKeys()
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.NotNil(files)
			}
		})
	}
}
