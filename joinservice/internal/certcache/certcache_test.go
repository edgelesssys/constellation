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
				AskResponse: string(testdata.Ask),
				ArkResponse: string(testdata.Ark),
			},
			kdsClient: &stubKdsClient{},
			wantAsk:   true,
			wantArk:   true,
		},
		"query from KDS": {
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
	pemBlock, _ := pem.Decode(c.askResponse)
	ask, err = x509.ParseCertificate(pemBlock.Bytes)
	if err != nil {
		return nil, nil, err
	}

	pemBlock, _ = pem.Decode(c.arkResponse)
	ark, err = x509.ParseCertificate(pemBlock.Bytes)
	if err != nil {
		return nil, nil, err
	}

	return ask, ark, c.certChainErr
}

func TestGetCertChainCache(t *testing.T) {
	testCases := map[string]struct {
		kubeClient *stubKubeClient
		wantErr    bool
	}{
		"success": {
			kubeClient: &stubKubeClient{
				AskResponse: string(testdata.Ask),
				ArkResponse: string(testdata.Ark),
			},
		},
		"empty ask": {
			kubeClient: &stubKubeClient{
				AskResponse: "",
				ArkResponse: string(testdata.Ark),
			},
			wantErr: true,
		},
		"empty ark": {
			kubeClient: &stubKubeClient{
				AskResponse: string(testdata.Ask),
				ArkResponse: "",
			},
			wantErr: true,
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
	AskResponse         string
	ArkResponse         string
	createConfigMapErr  error
	getConfigMapDataErr error
}

func (s *stubKubeClient) CreateConfigMap(context.Context, string, map[string]string) error {
	return s.createConfigMapErr
}

func (s *stubKubeClient) GetConfigMapData(ctx context.Context, name string, key string) (string, error) {
	if key == constants.CertCacheAskKey {
		return s.AskResponse, s.getConfigMapDataErr
	}
	if key == constants.CertCacheArkKey {
		return s.ArkResponse, s.getConfigMapDataErr
	}
	return "", s.getConfigMapDataErr
}
