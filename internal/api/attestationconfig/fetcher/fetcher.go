/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package fetcher

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/edgelesssys/constellation/v2/internal/api/attestationconfig"
	"github.com/edgelesssys/constellation/v2/internal/api/fetcher"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/sigstore"
)

const cosignPublicKey = constants.CosignPublicKeyReleases

// AttestationConfigAPIFetcher fetches config API resources without authentication.
type AttestationConfigAPIFetcher interface {
	FetchAzureSEVSNPVersion(ctx context.Context, azureVersion attestationconfig.AzureSEVSNPVersionAPI) (attestationconfig.AzureSEVSNPVersionAPI, error)
	FetchAzureSEVSNPVersionList(ctx context.Context, attestation attestationconfig.AzureSEVSNPVersionList) (attestationconfig.AzureSEVSNPVersionList, error)
	FetchAzureSEVSNPVersionLatest(ctx context.Context) (attestationconfig.AzureSEVSNPVersionAPI, error)
}

// Fetcher fetches AttestationCfg API resources without authentication.
type Fetcher struct {
	fetcher.HTTPClient
}

// New returns a new Fetcher.
func New() *Fetcher {
	return NewWithClient(fetcher.NewHTTPClient())
}

// NewWithClient returns a new Fetcher with custom http client.
func NewWithClient(client fetcher.HTTPClient) *Fetcher {
	return &Fetcher{client}
}

// FetchAzureSEVSNPVersionList fetches the version list information from the config API.
func (f *Fetcher) FetchAzureSEVSNPVersionList(ctx context.Context, attestation attestationconfig.AzureSEVSNPVersionList) (attestationconfig.AzureSEVSNPVersionList, error) {
	return fetcher.Fetch(ctx, f.HTTPClient, attestation)
}

// FetchAzureSEVSNPVersion fetches the version information from the config API.
func (f *Fetcher) FetchAzureSEVSNPVersion(ctx context.Context, azureVersion attestationconfig.AzureSEVSNPVersionAPI) (attestationconfig.AzureSEVSNPVersionAPI, error) {
	fetchedVersion, err := fetcher.Fetch(ctx, f.HTTPClient, azureVersion)
	if err != nil {
		return fetchedVersion, fmt.Errorf("fetch version %s: %w", fetchedVersion.Version, err)
	}
	versionBytes, err := json.Marshal(fetchedVersion)
	if err != nil {
		return fetchedVersion, fmt.Errorf("marshal version for verify %s: %w", fetchedVersion.Version, err)
	}

	signature, err := fetcher.Fetch(ctx, f.HTTPClient, attestationconfig.AzureSEVSNPVersionSignature{
		Version: azureVersion.Version,
	})
	if err != nil {
		return fetchedVersion, fmt.Errorf("fetch version %s signature: %w", fetchedVersion.Version, err)
	}

	err = sigstore.CosignVerifier{}.VerifySignature(versionBytes, signature.Signature, []byte(cosignPublicKey))
	if err != nil {
		return fetchedVersion, fmt.Errorf("verify version %s signature: %w", fetchedVersion.Version, err)
	}
	return fetchedVersion, nil
}

// FetchAzureSEVSNPVersionLatest returns the latest versions of the given type.
func (f *Fetcher) FetchAzureSEVSNPVersionLatest(ctx context.Context) (res attestationconfig.AzureSEVSNPVersionAPI, err error) {
	var list attestationconfig.AzureSEVSNPVersionList
	list, err = f.FetchAzureSEVSNPVersionList(ctx, list)
	if err != nil {
		return res, fmt.Errorf("fetching versions list: %w", err)
	}
	get := attestationconfig.AzureSEVSNPVersionAPI{Version: list[0]} // get latest version (as sorted reversely alphanumerically)
	get, err = f.FetchAzureSEVSNPVersion(ctx, get)
	if err != nil {
		return res, fmt.Errorf("fetching version: %w", err)
	}
	return get, nil
}
