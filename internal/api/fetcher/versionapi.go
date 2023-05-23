/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package fetcher

import (
	"context"

	"github.com/edgelesssys/constellation/v2/internal/api/versionsapi"
)

// VersionAPIFetcher fetches version API resources without authentication.
type VersionAPIFetcher struct {
	*fetcher
}

// NewVersionAPIFetcher returns a new Fetcher.
func NewVersionAPIFetcher() *VersionAPIFetcher {
	return &VersionAPIFetcher{newFetcher()}
}

// FetchVersionList fetches the given version list from the versions API.
func (f *VersionAPIFetcher) FetchVersionList(ctx context.Context, list versionsapi.List) (versionsapi.List, error) {
	return fetch(ctx, f.httpc, list)
}

// FetchVersionLatest fetches the latest version from the versions API.
func (f *VersionAPIFetcher) FetchVersionLatest(ctx context.Context, latest versionsapi.Latest) (versionsapi.Latest, error) {
	return fetch(ctx, f.httpc, latest)
}

// FetchImageInfo fetches the given image info from the versions API.
func (f *VersionAPIFetcher) FetchImageInfo(ctx context.Context, imageInfo versionsapi.ImageInfo) (versionsapi.ImageInfo, error) {
	return fetch(ctx, f.httpc, imageInfo)
}

// FetchCLIInfo fetches the given cli info from the versions API.
func (f *VersionAPIFetcher) FetchCLIInfo(ctx context.Context, cliInfo versionsapi.CLIInfo) (versionsapi.CLIInfo, error) {
	return fetch(ctx, f.httpc, cliInfo)
}
