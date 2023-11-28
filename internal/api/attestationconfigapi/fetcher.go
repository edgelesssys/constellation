/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package attestationconfigapi

import (
	"context"
	"errors"
	"fmt"

	apifetcher "github.com/edgelesssys/constellation/v2/internal/api/fetcher"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/sigstore"
)

const cosignPublicKey = constants.CosignPublicKeyReleases

// ErrNoVersionsFound is returned if no versions are found.
var ErrNoVersionsFound = errors.New("no versions found")

// Fetcher fetches config API resources without authentication.
type Fetcher interface {
	FetchSEVSNPVersion(ctx context.Context, version SEVSNPVersionAPI) (SEVSNPVersionAPI, error)
	FetchSEVSNPVersionList(ctx context.Context, list SEVSNPVersionList) (SEVSNPVersionList, error)
	FetchSEVSNPVersionLatest(ctx context.Context, attestation variant.Variant) (SEVSNPVersionAPI, error)
}

// fetcher fetches AttestationCfg API resources without authentication.
type fetcher struct {
	apifetcher.HTTPClient
	cdnURL   string
	verifier sigstore.Verifier
}

// NewFetcher returns a new apifetcher.
func NewFetcher() Fetcher {
	return NewFetcherWithClient(apifetcher.NewHTTPClient(), constants.CDNRepositoryURL)
}

// NewFetcherWithCustomCDNAndCosignKey returns a new fetcher with custom CDN URL.
func NewFetcherWithCustomCDNAndCosignKey(cdnURL, cosignKey string) Fetcher {
	verifier, err := sigstore.NewCosignVerifier([]byte(cosignKey))
	if err != nil {
		// This relies on an embedded public key. If this key can not be validated, there is no way to recover from this.
		panic(fmt.Errorf("creating cosign verifier: %w", err))
	}
	return newFetcherWithClientAndVerifier(apifetcher.NewHTTPClient(), verifier, cdnURL)
}

// NewFetcherWithClient returns a new fetcher with custom http client.
func NewFetcherWithClient(client apifetcher.HTTPClient, cdnURL string) Fetcher {
	verifier, err := sigstore.NewCosignVerifier([]byte(cosignPublicKey))
	if err != nil {
		// This relies on an embedded public key. If this key can not be validated, there is no way to recover from this.
		panic(fmt.Errorf("creating cosign verifier: %w", err))
	}
	return newFetcherWithClientAndVerifier(client, verifier, cdnURL)
}

func newFetcherWithClientAndVerifier(client apifetcher.HTTPClient, cosignVerifier sigstore.Verifier, url string) Fetcher {
	return &fetcher{HTTPClient: client, verifier: cosignVerifier, cdnURL: url}
}

// FetchSEVSNPVersionList fetches the version list information from the config API.
func (f *fetcher) FetchSEVSNPVersionList(ctx context.Context, list SEVSNPVersionList) (SEVSNPVersionList, error) {
	// TODO(derpsteb): Replace with FetchAndVerify once we move to v2 of the config API.
	fetchedList, err := apifetcher.Fetch(ctx, f.HTTPClient, f.cdnURL, list)
	if err != nil {
		return list, fmt.Errorf("fetching version list: %w", err)
	}

	// Need to set this explicitly as the variant is not part of the marshalled JSON.
	fetchedList.variant = list.variant

	return fetchedList, nil
}

// FetchSEVSNPVersion fetches the version information from the config API.
func (f *fetcher) FetchSEVSNPVersion(ctx context.Context, version SEVSNPVersionAPI) (SEVSNPVersionAPI, error) {
	fetchedVersion, err := apifetcher.FetchAndVerify(ctx, f.HTTPClient, f.cdnURL, version, f.verifier)
	if err != nil {
		return fetchedVersion, fmt.Errorf("fetching version %s: %w", version.Version, err)
	}

	// Need to set this explicitly as the variant is not part of the marshalled JSON.
	fetchedVersion.Variant = version.Variant

	return fetchedVersion, nil
}

// FetchSEVSNPVersionLatest returns the latest versions of the given type.
func (f *fetcher) FetchSEVSNPVersionLatest(ctx context.Context, attesation variant.Variant) (res SEVSNPVersionAPI, err error) {
	list, err := f.FetchSEVSNPVersionList(ctx, SEVSNPVersionList{variant: attesation})
	if err != nil {
		return res, ErrNoVersionsFound
	}

	getVersionRequest := SEVSNPVersionAPI{
		Version: list.List()[0], // latest version is first in list
		Variant: attesation,
	}
	res, err = f.FetchSEVSNPVersion(ctx, getVersionRequest)
	if err != nil {
		return res, err
	}
	return
}
