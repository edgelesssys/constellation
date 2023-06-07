/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package versionsapi

import (
	"context"

	"github.com/edgelesssys/constellation/v2/internal/api/fetcher"
)

// Fetcher fetches version API resources without authentication.
type Fetcher struct {
	fetcher.HTTPClient
}

// NewFetcher returns a new Fetcher.
func NewFetcher() *Fetcher {
	return &Fetcher{fetcher.NewHTTPClient()}
}

// FetchVersionList fetches the given version list from the versions API.
func (f *Fetcher) FetchVersionList(ctx context.Context, list List) (List, error) {
	return fetcher.Fetch(ctx, f.HTTPClient, list)
}

// FetchVersionLatest fetches the latest version from the versions API.
func (f *Fetcher) FetchVersionLatest(ctx context.Context, latest Latest) (Latest, error) {
	return fetcher.Fetch(ctx, f.HTTPClient, latest)
}

// FetchImageInfo fetches the given image info from the versions API.
func (f *Fetcher) FetchImageInfo(ctx context.Context, imageInfo ImageInfo) (ImageInfo, error) {
	return fetcher.Fetch(ctx, f.HTTPClient, imageInfo)
}

// FetchCLIInfo fetches the given cli info from the versions API.
func (f *Fetcher) FetchCLIInfo(ctx context.Context, cliInfo CLIInfo) (CLIInfo, error) {
	return fetcher.Fetch(ctx, f.HTTPClient, cliInfo)
}
