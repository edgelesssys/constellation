/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package kubectl

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
)

// restClientGetter implements k8s.io/cli-runtime/pkg/resource.RESTClientGetter.
type restClientGetter struct {
	clientconfig clientcmd.ClientConfig
}

// newRESTClientGetter creates a new restClientGetter using a kubeconfig.
func newRESTClientGetter(kubeconfig []byte) (*restClientGetter, error) {
	clientconfig, err := clientcmd.NewClientConfigFromBytes(kubeconfig)
	if err != nil {
		return nil, err
	}

	rawconfig, err := clientconfig.RawConfig()
	if err != nil {
		return nil, err
	}

	clientconfig = clientcmd.NewDefaultClientConfig(rawconfig, &clientcmd.ConfigOverrides{})

	return &restClientGetter{clientconfig}, nil
}

// ToRESTConfig returns k8s REST client config.
func (r *restClientGetter) ToRESTConfig() (*rest.Config, error) {
	return r.clientconfig.ClientConfig()
}

// ToDiscoveryClient creates new k8s discovery client from restClientGetter.
func (r *restClientGetter) ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	restconfig, err := r.clientconfig.ClientConfig()
	if err != nil {
		return nil, err
	}
	dc, err := discovery.NewDiscoveryClientForConfig(restconfig)
	if err != nil {
		return nil, err
	}
	return memory.NewMemCacheClient(dc), nil
}

// ToRESTMapper creates new k8s RESTMapper from restClientGetter.
func (r *restClientGetter) ToRESTMapper() (meta.RESTMapper, error) {
	dc, err := r.ToDiscoveryClient()
	if err != nil {
		return nil, err
	}
	return restmapper.NewDeferredDiscoveryRESTMapper(dc), nil
}

// ToRawKubeConfigLoader returns the inner k8s ClientConfig.
func (r *restClientGetter) ToRawKubeConfigLoader() clientcmd.ClientConfig {
	return r.clientconfig
}
