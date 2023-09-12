/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package certcache implements an in-cluster SEV-SNP certificate cache.
package certcache

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"

	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/joinservice/internal/certcache/amdkds"
	"github.com/google/go-sev-guest/abi"
	"github.com/google/go-sev-guest/verify/trust"
)

// CertCacheClient is a client for interacting with the certificate chain cache.
type CertCacheClient struct {
	log *logger.Logger
	kdsClient
	kubeClient kubeClient
}

// NewCertCacheClient creates a new CertCacheClient.
func NewCertCacheClient(log *logger.Logger, kubeClient kubeClient) *CertCacheClient {
	kdsClient := amdkds.NewKDSClient(trust.DefaultHTTPSGetter())

	return &CertCacheClient{
		log:        log,
		kubeClient: kubeClient,
		kdsClient:  kdsClient,
	}
}

// GetCertChainCache returns the cached ASK and ARK certificate.
func (c *CertCacheClient) GetCertChainCache(ctx context.Context) (ask, ark *x509.Certificate, err error) {
	c.log.Debugf("Retrieving certificate chain from cache")
	askRaw, err := c.kubeClient.GetConfigMapData(ctx, constants.SevSnpCertCacheConfigMapName, constants.CertCacheAskKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get ask: %w", err)
	}
	askBlock, _ := pem.Decode([]byte(askRaw))
	if askBlock == nil {
		return nil, nil, fmt.Errorf("failed to decode ask")
	}
	ask, err = x509.ParseCertificate(askBlock.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse ask: %w", err)
	}

	arkRaw, err := c.kubeClient.GetConfigMapData(ctx, constants.SevSnpCertCacheConfigMapName, constants.CertCacheArkKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get ark: %w", err)
	}
	arkBlock, _ := pem.Decode([]byte(arkRaw))
	if arkBlock == nil {
		return nil, nil, fmt.Errorf("failed to decode ark")
	}
	ark, err = x509.ParseCertificate(arkBlock.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse ark: %w", err)
	}

	c.log.Debugf("Retrieved certificate chain from cache")
	return ask, ark, nil
}

// CreateCertChainCache creates the certificate chain cache configmap with the provided ask and ark.
// If the configmap already exists, nothing is done.
func (c *CertCacheClient) CreateCertChainCache(ctx context.Context, signingType abi.ReportSigner) error {
	c.log.Debugf("Creating certificate chain cache")
	exists, err := c.kubeClient.ConfigMapExists(ctx, constants.SevSnpCertCacheConfigMapName)
	if err != nil {
		return fmt.Errorf("failed to check if configmap exists: %w", err)
	}
	if exists {
		c.log.Debugf("Certificate chain cache already exists")
		return nil
	}

	c.log.Debugf("Retrieving certificate chain from KDS")
	ask, ark, err := c.kdsClient.CertChain(signingType)
	if err != nil {
		return fmt.Errorf("failed to retrieve certificate chain from KDS: %w", err)
	}

	askWriter := &strings.Builder{}
	if err := pem.Encode(askWriter, &pem.Block{Type: "CERTIFICATE", Bytes: ask.Raw}); err != nil {
		return fmt.Errorf("failed to encode ask: %w", err)
	}
	arkWriter := &strings.Builder{}
	if err := pem.Encode(arkWriter, &pem.Block{Type: "CERTIFICATE", Bytes: ark.Raw}); err != nil {
		return fmt.Errorf("failed to encode ark: %w", err)
	}
	return c.kubeClient.CreateConfigMap(ctx, constants.SevSnpCertCacheConfigMapName, map[string]string{
		constants.CertCacheAskKey: askWriter.String(),
		constants.CertCacheArkKey: arkWriter.String(),
	})
}

type kubeClient interface {
	CreateConfigMap(ctx context.Context, name string, data map[string]string) error
	GetConfigMapData(ctx context.Context, name, key string) (string, error)
	ConfigMapExists(ctx context.Context, name string) (bool, error)
}

type kdsClient interface {
	CertChain(signingType abi.ReportSigner) (ask, ark *x509.Certificate, err error)
}
