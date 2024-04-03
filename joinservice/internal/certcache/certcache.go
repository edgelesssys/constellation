/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package certcache implements an in-cluster SEV-SNP certificate cache.
package certcache

import (
	"context"
	"crypto/x509"
	"fmt"
	"log/slog"

	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/crypto"
	"github.com/edgelesssys/constellation/v2/joinservice/internal/certcache/amdkds"
	"github.com/google/go-sev-guest/abi"
	"github.com/google/go-sev-guest/verify/trust"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
)

// Client is a client for interacting with the certificate chain cache.
type Client struct {
	log        *slog.Logger
	attVariant variant.Variant
	kdsClient
	kubeClient kubeClient
}

// NewClient creates a new CertCacheClient.
func NewClient(log *slog.Logger, kubeClient kubeClient, attVariant variant.Variant) *Client {
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
	var reportSigner abi.ReportSigner
	switch c.attVariant {
	case variant.AzureSEVSNP{}:
		reportSigner = abi.VcekReportSigner
	case variant.AWSSEVSNP{}:
		reportSigner = abi.VlekReportSigner
	default:
		c.log.Debug(fmt.Sprintf("No certificate chain caching possible for attestation variant %q", c.attVariant))
		return nil, nil
	}

	c.log.Debug(fmt.Sprintf("Creating %q certificate chain cache", c.attVariant))
	ask, ark, err := c.createCertChainCache(ctx, reportSigner)
	if err != nil {
		return nil, fmt.Errorf("creating %s certificate chain cache: %w", c.attVariant, err)
	}
	return &CachedCerts{
		ask: ask,
		ark: ark,
	}, nil
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
	c.log.Debug("Creating certificate chain cache")
	var shouldCreateConfigMap bool

	// Check if ASK or ARK is already cached.
	// If both are cached, return them.
	// If only one is cached, retrieve the other one from the
	// KDS and update the configmap with the missing value.
	cacheAsk, cacheArk, err := c.getCertChainCache(ctx)
	if k8serrors.IsNotFound(err) {
		c.log.Debug("Certificate chain cache does not exist")
		shouldCreateConfigMap = true
	} else if err != nil {
		return nil, nil, fmt.Errorf("getting certificate chain cache: %w", err)
	}
	if cacheAsk != nil && cacheArk != nil {
		c.log.Debug("ASK and ARK present in cache, returning cached values")
		return cacheAsk, cacheArk, nil
	}
	if cacheAsk != nil {
		ask = cacheAsk
	}
	if cacheArk != nil {
		ark = cacheArk
	}

	// If only one certificate is cached, retrieve the other one from the KDS.
	c.log.Debug("Retrieving certificate chain from KDS")
	kdsAsk, kdsArk, err := c.kdsClient.CertChain(signingType)
	if err != nil {
		return nil, nil, fmt.Errorf("retrieving certificate chain from KDS: %w", err)
	}
	if ask == nil && kdsAsk != nil {
		ask = kdsAsk
	}
	if ark == nil && kdsArk != nil {
		ark = kdsArk
	}

	askPem, err := crypto.X509CertToPem(ask)
	if err != nil {
		return nil, nil, fmt.Errorf("encoding ASK: %w", err)
	}
	arkPem, err := crypto.X509CertToPem(ark)
	if err != nil {
		return nil, nil, fmt.Errorf("encoding ARK: %w", err)
	}

	if shouldCreateConfigMap {
		// ConfigMap does not exist, create it.
		c.log.Debug("Creating certificate chain cache configmap")
		if err := c.kubeClient.CreateConfigMap(ctx, constants.SevSnpCertCacheConfigMapName, map[string]string{
			constants.CertCacheAskKey: string(askPem),
			constants.CertCacheArkKey: string(arkPem),
		}); err != nil {
			// If the ConfigMap already exists, another JoinService instance created the certificate cache while this operation was running.
			// Calling this function again should now retrieve the cached certificates.
			if k8serrors.IsAlreadyExists(err) {
				c.log.Debug("Certificate chain cache configmap already exists, retrieving cached certificates")
				return c.getCertChainCache(ctx)
			}
			return nil, nil, fmt.Errorf("creating certificate chain cache configmap: %w", err)
		}
	} else {
		// ConfigMap already exists but either ASK or ARK are missing. Update the according value.
		if cacheAsk == nil {
			if err := c.kubeClient.UpdateConfigMap(ctx, constants.SevSnpCertCacheConfigMapName,
				constants.CertCacheAskKey, string(askPem)); err != nil {
				return nil, nil, fmt.Errorf("updating ASK in certificate chain cache configmap: %w", err)
			}
		}
		if cacheArk == nil {
			if err := c.kubeClient.UpdateConfigMap(ctx, constants.SevSnpCertCacheConfigMapName,
				constants.CertCacheArkKey, string(arkPem)); err != nil {
				return nil, nil, fmt.Errorf("updating ARK in certificate chain cache configmap: %w", err)
			}
		}
	}

	return ask, ark, nil
}

// getCertChainCache returns the cached ASK and ARK certificate, if available. If either of the keys
// is not present in the configmap, no error is returned.
func (c *Client) getCertChainCache(ctx context.Context) (ask, ark *x509.Certificate, err error) {
	c.log.Debug("Retrieving certificate chain from cache")
	askRaw, err := c.kubeClient.GetConfigMapData(ctx, constants.SevSnpCertCacheConfigMapName, constants.CertCacheAskKey)
	if err != nil {
		return nil, nil, fmt.Errorf("getting ASK from configmap: %w", err)
	}
	if askRaw != "" {
		c.log.Debug("ASK cache hit")
		ask, err = crypto.PemToX509Cert([]byte(askRaw))
		if err != nil {
			return nil, nil, fmt.Errorf("parsing ASK: %w", err)
		}
	}

	arkRaw, err := c.kubeClient.GetConfigMapData(ctx, constants.SevSnpCertCacheConfigMapName, constants.CertCacheArkKey)
	if err != nil {
		return nil, nil, fmt.Errorf("getting ARK from configmap: %w", err)
	}
	if arkRaw != "" {
		c.log.Debug("ARK cache hit")
		ark, err = crypto.PemToX509Cert([]byte(arkRaw))
		if err != nil {
			return nil, nil, fmt.Errorf("parsing ARK: %w", err)
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
