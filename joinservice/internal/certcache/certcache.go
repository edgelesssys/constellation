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
// retrieved from the KDS and returns ASK and ARK. If the configmap already exists and both ASK and ARK are present,
// nothing is done and the existing ASK and ARK are returned. If the configmap already exists but either ASK or ARK
// are missing, the missing certificate is retrieved from the KDS and the configmap is updated with the missing value.
func (c *Client) createCertChainCache(ctx context.Context, signingType abi.ReportSigner) (ask, ark *x509.Certificate, err error) {
	c.log.Debugf("Creating certificate chain cache")
	var shouldCreateConfigMap bool

	cacheAsk, cacheArk, err := c.getCertChainCache(ctx)
	if k8serrors.IsNotFound(err) {
		c.log.Debugf("Certificate chain cache does not exist")
		shouldCreateConfigMap = true
	} else if err != nil {
		return nil, nil, fmt.Errorf("failed to get certificate chain cache: %w", err)
	}
	if cacheAsk != nil && cacheArk != nil {
		c.log.Debugf("ASK and ARK present in cache, returning cached values")
		return cacheAsk, cacheArk, nil
	}
	if cacheAsk != nil {
		ask = cacheAsk
	}
	if cacheArk != nil {
		ark = cacheArk
	}

	c.log.Debugf("Retrieving certificate chain from KDS")
	kdsAsk, kdsArk, err := c.kdsClient.CertChain(signingType)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to retrieve certificate chain from KDS: %w", err)
	}
	if kdsAsk != nil {
		ask = kdsAsk
	}
	if kdsArk != nil {
		ark = kdsArk
	}

	askWriter := &strings.Builder{}
	if err := pem.Encode(askWriter, &pem.Block{Type: "CERTIFICATE", Bytes: ask.Raw}); err != nil {
		return nil, nil, fmt.Errorf("failed to encode ask: %w", err)
	}
	arkWriter := &strings.Builder{}
	if err := pem.Encode(arkWriter, &pem.Block{Type: "CERTIFICATE", Bytes: ark.Raw}); err != nil {
		return nil, nil, fmt.Errorf("failed to encode ark: %w", err)
	}

	if shouldCreateConfigMap {
		// ConfigMap does not exist, create it.
		c.log.Debugf("Creating certificate chain cache configmap")
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
	} else {
		// ConfigMap already exists but either ASK or ARK are missing. Update the according value.
		if cacheAsk == nil {
			if err := c.kubeClient.UpdateConfigMap(ctx, constants.SevSnpCertCacheConfigMapName,
				constants.CertCacheAskKey, askWriter.String()); err != nil {
				return nil, nil, fmt.Errorf("failed to update ASK in certificate chain cache configmap: %w", err)
			}
		}
		if cacheArk == nil {
			if err := c.kubeClient.UpdateConfigMap(ctx, constants.SevSnpCertCacheConfigMapName,
				constants.CertCacheArkKey, arkWriter.String()); err != nil {
				return nil, nil, fmt.Errorf("failed to update ARK in certificate chain cache configmap: %w", err)
			}
		}
	}

	return ask, ark, nil
}

// getCertChainCache returns the cached ASK and ARK certificate, if available. If either of the keys
// is not present in the configmap, no error is returned.
func (c *Client) getCertChainCache(ctx context.Context) (ask, ark *x509.Certificate, err error) {
	c.log.Debugf("Retrieving certificate chain from cache")
	askRaw, err := c.kubeClient.GetConfigMapData(ctx, constants.SevSnpCertCacheConfigMapName, constants.CertCacheAskKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get ASK: %w", err)
	}
	if askRaw != "" {
		c.log.Debugf("ASK cache hit")
		askBlock, _ := pem.Decode([]byte(askRaw))
		ask, err = x509.ParseCertificate(askBlock.Bytes)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse ASK: %w", err)
		}
	}

	arkRaw, err := c.kubeClient.GetConfigMapData(ctx, constants.SevSnpCertCacheConfigMapName, constants.CertCacheArkKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get ARK: %w", err)
	}
	if arkRaw != "" {
		c.log.Debugf("ARK cache hit")
		arkBlock, _ := pem.Decode([]byte(arkRaw))
		ark, err = x509.ParseCertificate(arkBlock.Bytes)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse ARK: %w", err)
		}
	}

	return ask, ark, nil
}

type kubeClient interface {
	CreateConfigMap(ctx context.Context, name string, data map[string]string) error
	GetConfigMapData(ctx context.Context, name, key string) (string, error)
	UpdateConfigMap(ctx context.Context, name, key, value string) error
}

type kdsClient interface {
	CertChain(signingType abi.ReportSigner) (ask, ark *x509.Certificate, err error)
}
