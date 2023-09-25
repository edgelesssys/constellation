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

	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/joinservice/internal/certcache/amdkds"
	"github.com/google/go-sev-guest/abi"
	"github.com/google/go-sev-guest/verify/trust"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
)

// Client is a client for interacting with the certificate chain cache.
type Client struct {
	log        *logger.Logger
	attVariant variant.Variant
	kdsClient
	kubeClient kubeClient
}

// NewClient creates a new CertCacheClient.
func NewClient(log *logger.Logger, kubeClient kubeClient, attVariant variant.Variant) *Client {
	kdsClient := amdkds.NewKDSClient(trust.DefaultHTTPSGetter())

	return &Client{
		attVariant: attVariant,
		log:        log,
		kubeClient: kubeClient,
		kdsClient:  kdsClient,
	}
}

// CreateCertChainCache creates a certificate chain cache for the given attestation variant
// and returns the cached certificates, if applicable.
// If the certificate chain cache already exists, nothing is done.
func (c *Client) CreateCertChainCache(ctx context.Context) (*CachedCerts, error) {
	switch c.attVariant {
	case variant.AzureSEVSNP{}:
		c.log.Debugf("Creating azure SEV-SNP certificate chain cache")
		ask, ark, err := c.createCertChainCache(ctx, abi.VcekReportSigner)
		if err != nil {
			return nil, fmt.Errorf("failed to create azure SEV-SNP certificate chain cache: %w", err)
		}
		return &CachedCerts{
			ask: ask,
			ark: ark,
		}, nil
	// TODO(derpsteb): Add AWS
	default:
		c.log.Debugf("No certificate chain caching possible for attestation variant %s", c.attVariant)
		return nil, nil
	}
}

// CachedCerts contains the cached certificates.
type CachedCerts struct {
	ask *x509.Certificate
	ark *x509.Certificate
}

// SevSnpCerts returns the cached SEV-SNP ASK and ARK certificates.
func (c *CachedCerts) SevSnpCerts() (ask, ark *x509.Certificate) {
	return c.ask, c.ark
}

// createCertChainCache creates a certificate chain cache configmap with the ASK and ARK
// retrieved from the KDS and returns ASK and ARK.
// If the configmap already exists, nothing is done and the existing ASK and ARK are returned.
func (c *Client) createCertChainCache(ctx context.Context, signingType abi.ReportSigner) (ask, ark *x509.Certificate, err error) {
	c.log.Debugf("Creating certificate chain cache")
	ask, ark, err = c.getCertChainCache(ctx)
	if k8serrors.IsNotFound(err) {
		c.log.Debugf("Certificate chain cache does not exist")
	} else if err != nil {
		return nil, nil, fmt.Errorf("failed to get certificate chain cache: %w", err)
	}
	if ask != nil && ark != nil {
		c.log.Debugf("Certificate chain cache already exists")
		return ask, ark, nil
	}

	c.log.Debugf("Retrieving certificate chain from KDS")
	ask, ark, err = c.kdsClient.CertChain(signingType)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to retrieve certificate chain from KDS: %w", err)
	}

	askWriter := &strings.Builder{}
	if err := pem.Encode(askWriter, &pem.Block{Type: "CERTIFICATE", Bytes: ask.Raw}); err != nil {
		return nil, nil, fmt.Errorf("failed to encode ask: %w", err)
	}
	arkWriter := &strings.Builder{}
	if err := pem.Encode(arkWriter, &pem.Block{Type: "CERTIFICATE", Bytes: ark.Raw}); err != nil {
		return nil, nil, fmt.Errorf("failed to encode ark: %w", err)
	}

	c.log.Debugf("Creating certificate chain cache configmap")
	// TODO(msanft): Make this function update the config instead of trying to create it
	// if either the ASK or ARK is missing.
	if err := c.kubeClient.CreateConfigMap(ctx, constants.SevSnpCertCacheConfigMapName, map[string]string{
		constants.CertCacheAskKey: askWriter.String(),
		constants.CertCacheArkKey: arkWriter.String(),
	}); err != nil {
		// If the ConfigMap already exists, another JoinService instance created the certificate cache while this operation was running.
		// Calling this function again should now retrieve the cached certificates.
		if k8serrors.IsAlreadyExists(err) {
			c.log.Debugf("Certificate chain cache configmap already exists, retrieving cached certificates")
			return c.getCertChainCache(ctx)
		}
		return nil, nil, fmt.Errorf("failed to create certificate chain cache configmap: %w", err)
	}

	return ask, ark, nil
}

// getCertChainCache returns the cached ASK and ARK certificate.
func (c *Client) getCertChainCache(ctx context.Context) (ask, ark *x509.Certificate, err error) {
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

type kubeClient interface {
	CreateConfigMap(ctx context.Context, name string, data map[string]string) error
	GetConfigMapData(ctx context.Context, name, key string) (string, error)
}

type kdsClient interface {
	CertChain(signingType abi.ReportSigner) (ask, ark *x509.Certificate, err error)
}
