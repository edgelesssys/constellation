/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/edgelesssys/constellation/v2/internal/api/attestationconfig"
	"github.com/edgelesssys/constellation/v2/internal/api/fetcher"
	versionsapi "github.com/edgelesssys/constellation/v2/internal/api/versions"
	"github.com/edgelesssys/constellation/v2/internal/sigstore"
)

// AttestationConfigAPIFetcher fetches config API resources without authentication.
type AttestationConfigAPIFetcher interface {
	FetchAzureSEVSNPVersion(ctx context.Context, azureVersion attestationconfig.AzureSEVSNPVersionGet, version versionsapi.Version) (attestationconfig.AzureSEVSNPVersionGet, error)
	FetchAzureSEVSNPVersionList(ctx context.Context, attestation attestationconfig.AzureSEVSNPVersionList) (attestationconfig.AzureSEVSNPVersionList, error)
	FetchAzureSEVSNPVersionLatest(ctx context.Context, version versionsapi.Version) (attestationconfig.AzureSEVSNPVersionGet, error)
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
func (f *Fetcher) FetchAzureSEVSNPVersion(ctx context.Context, attestation attestationconfig.AzureSEVSNPVersionGet, version versionsapi.Version) (attestationconfig.AzureSEVSNPVersionGet, error) {
	cosignPublicKey, err := sigstore.CosignPublicKeyForVersion(version)
	if err != nil {
		return attestationconfig.AzureSEVSNPVersionGet{}, fmt.Errorf("get public key for config: %w", err)
	}
	urlString, err := attestation.URL()
	if err != nil {
		return attestationconfig.AzureSEVSNPVersionGet{}, err
	}
	fetchedVersion, err := fetcher.Fetch(ctx, f.HTTPClient, attestation)
	if err != nil {
		return fetchedVersion, fmt.Errorf("fetch version %s: %w", fetchedVersion.Version, err)
	}
	versionBytes, err := json.Marshal(fetchedVersion)
	if err != nil {
		return fetchedVersion, fmt.Errorf("marshal version for verify %s: %w", fetchedVersion.Version, err)
	}

	signature, err := fetchBytesFromRawURL(ctx, fmt.Sprintf("%s.sig", urlString), f.HTTPClient)
	if err != nil {
		return fetchedVersion, fmt.Errorf("fetch version %s signature: %w", fetchedVersion.Version, err)
	}

	err = sigstore.CosignVerifier{}.VerifySignature(versionBytes, signature, cosignPublicKey)
	if err != nil {
		return fetchedVersion, fmt.Errorf("verify version %s signature: %w", fetchedVersion.Version, err)
	}
	return fetchedVersion, nil
}

// FetchAzureSEVSNPVersionLatest returns the latest versions of the given type.
func (f *Fetcher) FetchAzureSEVSNPVersionLatest(ctx context.Context, version versionsapi.Version) (res attestationconfig.AzureSEVSNPVersionGet, err error) {
	var list attestationconfig.AzureSEVSNPVersionList
	list, err = f.FetchAzureSEVSNPVersionList(ctx, list)
	if err != nil {
		return res, fmt.Errorf("fetching versions list: %w", err)
	}
	get := attestationconfig.AzureSEVSNPVersionGet{Version: list[0]} // get latest version (as sorted reversely alphanumerically)
	get, err = f.FetchAzureSEVSNPVersion(ctx, get, version)
	if err != nil {
		return res, fmt.Errorf("failed fetching version: %w", err)
	}
	return get, nil
}

func fetchBytesFromRawURL(ctx context.Context, urlString string, client fetcher.HTTPClient) ([]byte, error) {
	url, err := url.Parse(urlString)
	if err != nil {
		return nil, fmt.Errorf("parse version url %s: %w", urlString, err)
	}
	return getFromURL(ctx, client, url)
}

// getFromURL fetches the content from the given URL and returns the content as a byte slice.
func getFromURL(ctx context.Context, client fetcher.HTTPClient, sourceURL *url.URL) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, sourceURL.String(), http.NoBody)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http status code: %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}
