/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package certcache

import (
	"context"
	"crypto/x509"
	"log/slog"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/crypto"
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
		kubeClient  *stubKubeClient
		kdsClient   *stubKdsClient
		expectedArk *x509.Certificate
		expectedAsk *x509.Certificate
		wantErr     bool
	}{
		"available in configmap": {
			kubeClient: &stubKubeClient{
				askResponse: string(testdata.Ask),
				arkResponse: string(testdata.Ark),
			},
			kdsClient:   &stubKdsClient{},
			expectedArk: mustParsePEM(testdata.Ark),
			expectedAsk: mustParsePEM(testdata.Ask),
		},
		"query from kds": {
			kubeClient: &stubKubeClient{
				getConfigMapDataErr: notFoundErr,
			},
			kdsClient: &stubKdsClient{
				askResponse: testdata.Ask,
				arkResponse: testdata.Ark,
			},
			expectedArk: mustParsePEM(testdata.Ark),
			expectedAsk: mustParsePEM(testdata.Ask),
		},
		"only replace uncached cert": {
			kubeClient: &stubKubeClient{
				askResponse: string(testdata.Ark), // on purpose
			},
			kdsClient: &stubKdsClient{
				arkResponse: testdata.Ark,
				askResponse: testdata.Ask,
			},
			expectedArk: mustParsePEM(testdata.Ark),
			expectedAsk: mustParsePEM(testdata.Ark), // on purpose
		},
		"only ask available in configmap": {
			kubeClient: &stubKubeClient{
				askResponse: string(testdata.Ask),
			},
			kdsClient: &stubKdsClient{
				arkResponse: testdata.Ark,
			},
			expectedArk: mustParsePEM(testdata.Ark),
			expectedAsk: mustParsePEM(testdata.Ask),
		},
		"only ark available in configmap": {
			kubeClient: &stubKubeClient{
				arkResponse: string(testdata.Ark),
			},
			kdsClient: &stubKdsClient{
				askResponse: testdata.Ask,
			},
			expectedArk: mustParsePEM(testdata.Ark),
			expectedAsk: mustParsePEM(testdata.Ask),
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
				arkResponse: testdata.Ark,
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
				log:        slog.New(slog.NewTextHandler(logger.TestWriter{T: t}, nil)),
				kubeClient: tc.kubeClient,
				kdsClient:  tc.kdsClient,
			}

			ask, ark, err := c.createCertChainCache(ctx, abi.NoneReportSigner)
			if tc.wantErr {
				assert.Error(err)
			} else {
				require.NoError(err)
				assert.Equal(tc.expectedArk, ark)
				assert.Equal(tc.expectedAsk, ask)
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
	cert, err := crypto.PemToX509Cert(pemBytes)
	if err != nil {
		panic(err)
	}
	return cert
}

func TestGetCertChainCache(t *testing.T) {
	testCases := map[string]struct {
		kubeClient  *stubKubeClient
		expectedAsk *x509.Certificate
		expectedArk *x509.Certificate
		wantErr     bool
	}{
		"success": {
			kubeClient: &stubKubeClient{
				askResponse: string(testdata.Ask),
				arkResponse: string(testdata.Ark),
			},
			expectedAsk: mustParsePEM(testdata.Ask),
			expectedArk: mustParsePEM(testdata.Ark),
		},
		"empty ask": {
			kubeClient: &stubKubeClient{
				askResponse: "",
				arkResponse: string(testdata.Ark),
			},
			expectedAsk: nil,
			expectedArk: mustParsePEM(testdata.Ark),
		},
		"empty ark": {
			kubeClient: &stubKubeClient{
				askResponse: string(testdata.Ask),
				arkResponse: "",
			},
			expectedAsk: mustParsePEM(testdata.Ask),
			expectedArk: nil,
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

			c := NewClient(slog.New(slog.NewTextHandler(logger.TestWriter{T: t}, nil)), tc.kubeClient, variant.Dummy{})

			ask, ark, err := c.getCertChainCache(ctx)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.expectedArk, ark)
				assert.Equal(tc.expectedAsk, ask)
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

func (s *stubKubeClient) GetConfigMapData(_ context.Context, _ string, key string) (string, error) {
	if key == constants.CertCacheAskKey {
		return s.askResponse, s.getConfigMapDataErr
	}
	if key == constants.CertCacheArkKey {
		return s.arkResponse, s.getConfigMapDataErr
	}
	return "", s.getConfigMapDataErr
}

func (s *stubKubeClient) UpdateConfigMap(context.Context, string, string, string) error {
	return s.updateConfigMapErr
}
