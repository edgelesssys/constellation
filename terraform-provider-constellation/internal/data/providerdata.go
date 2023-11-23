/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/
package data

import (
	"context"

	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
)

// ProviderData is the data that get's passed down from the provider
// configuration to the resources and data sources.
type ProviderData struct {
	ImageFetcher ImageFetcher
}

// ImageFetcher gets an image reference from the versionsapi.
type ImageFetcher interface {
	FetchReference(ctx context.Context,
		provider cloudprovider.Provider, attestationVariant variant.Variant,
		image, region string,
	) (string, error)
}
