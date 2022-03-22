package client

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
	RawConfigConfig     clientcmdapi.Config
	RawConfigErr        error
	ClientConfigConfig  *restclient.Config
	ClientConfigErr     error
	NamespaceString     string
	NamespaceOverridden bool
	NamespaceErr        error
	ConfigAccessResult  clientcmd.ConfigAccess
}

func (s *stubClientConfig) RawConfig() (clientcmdapi.Config, error) {
	return s.RawConfigConfig, s.RawConfigErr
}

func (s *stubClientConfig) ClientConfig() (*restclient.Config, error) {
	return s.ClientConfigConfig, s.ClientConfigErr
}

func (s *stubClientConfig) Namespace() (string, bool, error) {
	return s.NamespaceString, s.NamespaceOverridden, s.NamespaceErr
}

func (s *stubClientConfig) ConfigAccess() clientcmd.ConfigAccess {
	return s.ConfigAccessResult
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
			ClientConfigConfig: &restclient.Config{},
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
			ClientConfigConfig: &restclient.Config{},
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
			ClientConfigErr: errors.New("someErr"),
		},
	}
	_, err := getter.ToDiscoveryClient()
	require.Error(err)
}

func TestToRESTMapper(t *testing.T) {
	require := require.New(t)
	getter := restClientGetter{
		clientconfig: &stubClientConfig{
			ClientConfigConfig: &restclient.Config{},
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
			ClientConfigErr: errors.New("someErr"),
		},
	}
	_, err := getter.ToRESTMapper()
	require.Error(err)
}

func TestToRawKubeConfigLoader(t *testing.T) {
	clientConfig := stubClientConfig{
		ClientConfigConfig: &restclient.Config{},
	}
	require := require.New(t)
	getter := restClientGetter{
		clientconfig: &clientConfig,
	}
	result := getter.ToRawKubeConfigLoader()
	require.Equal(&clientConfig, result)
}
