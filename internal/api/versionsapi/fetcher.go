/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package versionsapi

import (
	"context"

	"github.com/edgelesssys/constellation/v2/internal/api/fetcher"
	"github.com/edgelesssys/constellation/v2/internal/constants"
)

// Fetcher fetches version API resources without authentication.
type Fetcher struct {
	fetcher.HTTPClient
	cdnURL string
}

// NewFetcher returns a new Fetcher.
func NewFetcher() *Fetcher {
	return &Fetcher{fetcher.NewHTTPClient(), constants.CDNRepositoryURL}
}

// FetchVersionList fetches the given version list from the versions API.
func (f *Fetcher) FetchVersionList(ctx context.Context, list List) (List, error) {
	return fetcher.Fetch(ctx, f.HTTPClient, f.cdnURL, list)
}

// FetchVersionLatest fetches the latest version from the versions API.
func (f *Fetcher) FetchVersionLatest(ctx context.Context, latest Latest) (Latest, error) {
	return fetcher.Fetch(ctx, f.HTTPClient, f.cdnURL, latest)
}

// FetchImageInfo fetches the given image info from the versions API.
func (f *Fetcher) FetchImageInfo(ctx context.Context, imageInfo ImageInfo) (ImageInfo, error) {
	return fetcher.Fetch(ctx, f.HTTPClient, f.cdnURL, imageInfo)
}

// FetchCLIInfo fetches the given cli info from the versions API.
func (f *Fetcher) FetchCLIInfo(ctx context.Context, cliInfo CLIInfo) (CLIInfo, error) {
	return fetcher.Fetch(ctx, f.HTTPClient, f.cdnURL, cliInfo)
}
