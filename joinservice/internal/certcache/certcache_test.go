/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package certcache

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/joinservice/internal/certcache/testdata"
	"github.com/google/go-sev-guest/abi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestCreateCertChainCache(t *testing.T) {
	notFoundErr := k8serrors.NewNotFound(schema.GroupResource{}, "test")

	testCases := map[string]struct {
		kubeClient *stubKubeClient
		kdsClient  *stubKdsClient
		wantAsk    bool
		wantArk    bool
		wantErr    bool
	}{
		"available in configmap": {
			kubeClient: &stubKubeClient{
				askResponse: string(testdata.Ask),
				arkResponse: string(testdata.Ark),
			},
			kdsClient: &stubKdsClient{},
			wantAsk:   true,
			wantArk:   true,
		},
		"query from kds": {
			kubeClient: &stubKubeClient{
				getConfigMapDataErr: notFoundErr,
			},
			kdsClient: &stubKdsClient{
				askResponse: []byte(testdata.Ask),
				arkResponse: []byte(testdata.Ark),
			},
			wantAsk: true,
			wantArk: true,
		},
		"only ask available in configmap": {
			kubeClient: &stubKubeClient{
				askResponse: string(testdata.Ask),
			},
			kdsClient: &stubKdsClient{
				arkResponse: []byte(testdata.Ark),
			},
			wantAsk: true,
			wantArk: true,
		},
		"only ark available in configmap": {
			kubeClient: &stubKubeClient{
				arkResponse: string(testdata.Ark),
			},
			kdsClient: &stubKdsClient{
				askResponse: []byte(testdata.Ask),
			},
			wantAsk: true,
			wantArk: true,
		},
		"get config map data err": {
			kubeClient: &stubKubeClient{
				getConfigMapDataErr: assert.AnError,
			},
			wantErr: true,
		},
		"update configmap err": {
			kubeClient: &stubKubeClient{
				askResponse:        string(testdata.Ask),
				updateConfigMapErr: assert.AnError,
			},
			kdsClient: &stubKdsClient{
				arkResponse: []byte(testdata.Ark),
			},
			wantErr: true,
		},
		"kds cert chain err": {
			kubeClient: &stubKubeClient{
				getConfigMapDataErr: notFoundErr,
			},
			kdsClient: &stubKdsClient{
				certChainErr: assert.AnError,
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			ctx := context.Background()

			c := &Client{
				attVariant: variant.Dummy{},
				log:        logger.NewTest(t),
				kubeClient: tc.kubeClient,
				kdsClient:  tc.kdsClient,
			}

			ask, ark, err := c.createCertChainCache(ctx, abi.NoneReportSigner)
			if tc.wantErr {
				assert.Error(err)
			} else {
				require.NoError(err)
				if tc.wantArk {
					assert.NotNil(ark)
				}
				if tc.wantAsk {
					assert.NotNil(ask)
				}
			}
		})
	}
}

type stubKdsClient struct {
	askResponse  []byte
	arkResponse  []byte
	certChainErr error
}

func (c *stubKdsClient) CertChain(abi.ReportSigner) (ask, ark *x509.Certificate, err error) {
	if c.askResponse != nil {
		ask = mustParsePEM(c.askResponse)
	}
	if c.arkResponse != nil {
		ark = mustParsePEM(c.arkResponse)
	}
	return ask, ark, c.certChainErr
}

func mustParsePEM(pemBytes []byte) *x509.Certificate {
	pemBlock, _ := pem.Decode(pemBytes)
	cert, err := x509.ParseCertificate(pemBlock.Bytes)
	if err != nil {
		panic(err)
	}
	return cert
}

func TestGetCertChainCache(t *testing.T) {
	testCases := map[string]struct {
		kubeClient *stubKubeClient
		wantErr    bool
	}{
		"success": {
			kubeClient: &stubKubeClient{
				askResponse: string(testdata.Ask),
				arkResponse: string(testdata.Ark),
			},
		},
		"empty ask": {
			kubeClient: &stubKubeClient{
				askResponse: "",
				arkResponse: string(testdata.Ark),
			},
		},
		"empty ark": {
			kubeClient: &stubKubeClient{
				askResponse: string(testdata.Ask),
				arkResponse: "",
			},
		},
		"error getting config map data": {
			kubeClient: &stubKubeClient{
				getConfigMapDataErr: assert.AnError,
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			ctx := context.Background()

			c := NewClient(logger.NewTest(t), tc.kubeClient, variant.Dummy{})

			_, _, err := c.getCertChainCache(ctx)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

type stubKubeClient struct {
	askResponse         string
	arkResponse         string
	createConfigMapErr  error
	updateConfigMapErr  error
	getConfigMapDataErr error
}

func (s *stubKubeClient) CreateConfigMap(context.Context, string, map[string]string) error {
	return s.createConfigMapErr
}

func (s *stubKubeClient) GetConfigMapData(ctx context.Context, name string, key string) (string, error) {
	if key == constants.CertCacheAskKey {
		return s.askResponse, s.getConfigMapDataErr
	}
	if key == constants.CertCacheArkKey {
		return s.arkResponse, s.getConfigMapDataErr
	}
	return "", s.getConfigMapDataErr
}

func (s *stubKubeClient) UpdateConfigMap(ctx context.Context, name, key, value string) error {
	return s.updateConfigMapErr
}
