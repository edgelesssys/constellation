/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package mpimage

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/semver"
)

// MarketplaceImage represents a CSP-agnostic marketplace image.
type MarketplaceImage interface {
	URI() string
}

// NewFromURI returns a new MarketplaceImage for the given image URI.
func NewFromURI(uri string) (MarketplaceImage, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	if u.Scheme != constants.MarketplaceImageURIScheme {
		return nil, fmt.Errorf("invalid scheme: %s", u.Scheme)
	}

	switch u.Host {
	case cloudprovider.Azure.String():
		ver, err := semver.New(u.Query().Get(constants.AzureMarketplaceImageVersionKey))
		if err != nil {
			return nil, fmt.Errorf("invalid image version: %w", err)
		}
		return NewAzureMarketplaceImage(ver), nil
	default:
		return nil, fmt.Errorf("invalid host: %s", u.Host)
	}
}

// AzureMarketplaceImage represents an Azure marketplace image.
type AzureMarketplaceImage struct {
	Publisher string
	Offer     string
	SKU       string
	Version   string
}

// NewAzureMarketplaceImage returns a new Constellation marketplace image for the given version.
func NewAzureMarketplaceImage(version semver.Semver) AzureMarketplaceImage {
	return AzureMarketplaceImage{
		Publisher: constants.AzureMarketplaceImagePublisher,
		Offer:     constants.AzureMarketplaceImageOffer,
		SKU:       constants.AzureMarketplaceImagePlan,
		Version:   strings.TrimPrefix(version.String(), "v"), // Azure requires X.Y.Z format
	}
}

// URI returns the URI for the image.
func (i AzureMarketplaceImage) URI() string {
	u := &url.URL{
		Scheme: constants.MarketplaceImageURIScheme,
		Host:   cloudprovider.Azure.String(),
	}

	q := u.Query()
	q.Set(constants.AzureMarketplaceImagePublisherKey, i.Publisher)
	q.Set(constants.AzureMarketplaceImageOfferKey, i.Offer)
	q.Set(constants.AzureMarketplaceImageSkuKey, i.SKU)
	q.Set(constants.AzureMarketplaceImageVersionKey, i.Version)
	u.RawQuery = q.Encode()

	return u.String()
}
