/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package fetcher

import (
	"context"

	"github.com/edgelesssys/constellation/v2/internal/api/fetcher"
	"github.com/edgelesssys/constellation/v2/internal/api/versions"
)

// Fetcher fetches version API resources without authentication.
type Fetcher struct {
	fetcher.HTTPClient
}

// New returns a new Fetcher.
func New() *Fetcher {
	return &Fetcher{fetcher.NewHTTPClient()}
}

// FetchVersionList fetches the given version list from the versions API.
func (f *Fetcher) FetchVersionList(ctx context.Context, list versions.List) (versions.List, error) {
	return fetcher.Fetch(ctx, f.HTTPClient, list)
}

// FetchVersionLatest fetches the latest version from the versions API.
func (f *Fetcher) FetchVersionLatest(ctx context.Context, latest versions.Latest) (versions.Latest, error) {
	return fetcher.Fetch(ctx, f.HTTPClient, latest)
}

// FetchImageInfo fetches the given image info from the versions API.
func (f *Fetcher) FetchImageInfo(ctx context.Context, imageInfo versions.ImageInfo) (versions.ImageInfo, error) {
	return fetcher.Fetch(ctx, f.HTTPClient, imageInfo)
}

// FetchCLIInfo fetches the given cli info from the versions API.
func (f *Fetcher) FetchCLIInfo(ctx context.Context, cliInfo versions.CLIInfo) (versions.CLIInfo, error) {
	return fetcher.Fetch(ctx, f.HTTPClient, cliInfo)
}
