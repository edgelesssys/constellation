/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package kubectl

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

const testingKubeconfig = `
apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: ""
    server: https://192.0.2.0:6443
  name: kubernetes
contexts:
- context:
    cluster: kubernetes
    user: kubernetes-admin
  name: kubernetes-admin@kubernetes
current-context: kubernetes-admin@kubernetes
kind: Config
preferences: {}
users:
- name: kubernetes-admin
  user:
    client-certificate-data: ""
    client-key-data: ""
`

type stubClientConfig struct {
	rawConfigConfig     clientcmdapi.Config
	rawConfigErr        error
	clientConfigConfig  *restclient.Config
	clientConfigErr     error
	namespaceString     string
	namespaceOverridden bool
	namespaceErr        error
	configAccessResult  clientcmd.ConfigAccess
}

func (s *stubClientConfig) RawConfig() (clientcmdapi.Config, error) {
	return s.rawConfigConfig, s.rawConfigErr
}

func (s *stubClientConfig) ClientConfig() (*restclient.Config, error) {
	return s.clientConfigConfig, s.clientConfigErr
}

func (s *stubClientConfig) Namespace() (string, bool, error) {
	return s.namespaceString, s.namespaceOverridden, s.namespaceErr
}

func (s *stubClientConfig) ConfigAccess() clientcmd.ConfigAccess {
	return s.configAccessResult
}

func TestNewRESTClientGetter(t *testing.T) {
	require := require.New(t)
	result, err := newRESTClientGetter([]byte(testingKubeconfig))
	require.NoError(err)
	require.NotNil(result)
}

func TestToRESTConfig(t *testing.T) {
	require := require.New(t)
	getter := restClientGetter{
		clientconfig: &stubClientConfig{
			clientConfigConfig: &restclient.Config{},
		},
	}
	result, err := getter.ToRESTConfig()
	require.NoError(err)
	require.NotNil(result)
}

func TestToDiscoveryClient(t *testing.T) {
	require := require.New(t)
	getter := restClientGetter{
		clientconfig: &stubClientConfig{
			clientConfigConfig: &restclient.Config{},
		},
	}
	result, err := getter.ToDiscoveryClient()
	require.NoError(err)
	require.NotNil(result)
}

func TestToDiscoveryClientFail(t *testing.T) {
	require := require.New(t)
	getter := restClientGetter{
		clientconfig: &stubClientConfig{
			clientConfigErr: errors.New("someErr"),
		},
	}
	_, err := getter.ToDiscoveryClient()
	require.Error(err)
}

func TestToRESTMapper(t *testing.T) {
	require := require.New(t)
	getter := restClientGetter{
		clientconfig: &stubClientConfig{
			clientConfigConfig: &restclient.Config{},
		},
	}
	result, err := getter.ToRESTMapper()
	require.NoError(err)
	require.NotNil(result)
}

func TestToRESTMapperFail(t *testing.T) {
	require := require.New(t)
	getter := restClientGetter{
		clientconfig: &stubClientConfig{
			clientConfigErr: errors.New("someErr"),
		},
	}
	_, err := getter.ToRESTMapper()
	require.Error(err)
}

func TestToRawKubeConfigLoader(t *testing.T) {
	clientConfig := stubClientConfig{
		clientConfigConfig: &restclient.Config{},
	}
	require := require.New(t)
	getter := restClientGetter{
		clientconfig: &clientConfig,
	}
	result := getter.ToRawKubeConfigLoader()
	require.Equal(&clientConfig, result)
}
