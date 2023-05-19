/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
Package fetcher implements a client for the versions API.

The fetcher is used to get information from the versions API without having to
authenticate with AWS, where the API is currently hosted. This package should be
used in user-facing application code most of the time, like the CLI.
*/
package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/edgelesssys/constellation/v2/internal/versionsapi"
)

// Fetcher fetches versions API resources without authentication.
type Fetcher struct {
	httpc httpc
}

// NewFetcher returns a new Fetcher.
func NewFetcher() *Fetcher {
	return &Fetcher{
		httpc: &http.Client{Transport: &http.Transport{DisableKeepAlives: true}}, // DisableKeepAlives fixes concurrency issue see https://stackoverflow.com/a/75816347
	}
}

// FetchVersionList fetches the given version list from the versions API.
func (f *Fetcher) FetchVersionList(ctx context.Context, list versionsapi.List) (versionsapi.List, error) {
	return fetch(ctx, f.httpc, list)
}

// FetchVersionLatest fetches the latest version from the versions API.
func (f *Fetcher) FetchVersionLatest(ctx context.Context, latest versionsapi.Latest) (versionsapi.Latest, error) {
	return fetch(ctx, f.httpc, latest)
}

// FetchImageInfo fetches the given image info from the versions API.
func (f *Fetcher) FetchImageInfo(ctx context.Context, imageInfo versionsapi.ImageInfo) (versionsapi.ImageInfo, error) {
	return fetch(ctx, f.httpc, imageInfo)
}

// FetchCLIInfo fetches the given cli info from the versions API.
func (f *Fetcher) FetchCLIInfo(ctx context.Context, cliInfo versionsapi.CLIInfo) (versionsapi.CLIInfo, error) {
	return fetch(ctx, f.httpc, cliInfo)
}

// FetchAttestationList fetches the version list information from the config API.
func (f *Fetcher) FetchAttestationList(ctx context.Context, attestation versionsapi.AzureSEVSNPVersionList) (versionsapi.AzureSEVSNPVersionList, error) {
	return fetch(ctx, f.httpc, attestation)
}

// FetchAttestationVersion fetches the version information from the config API.
func (f *Fetcher) FetchAttestationVersion(ctx context.Context, attestation versionsapi.AzureSEVSNPVersionGet) (versionsapi.AzureSEVSNPVersionGet, error) {
	return fetch(ctx, f.httpc, attestation)
}

type apiObject interface {
	ValidateRequest() error
	Validate() error
	URL() (string, error)
}

func fetch[T apiObject](ctx context.Context, c httpc, obj T) (T, error) {
	if err := obj.ValidateRequest(); err != nil {
		return *new(T), fmt.Errorf("validating request for %T: %w", obj, err)
	}

	url, err := obj.URL()
	if err != nil {
		return *new(T), fmt.Errorf("getting URL for %T: %w", obj, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return *new(T), fmt.Errorf("creating request for %T: %w", obj, err)
	}

	resp, err := c.Do(req)
	if err != nil {
		return *new(T), fmt.Errorf("sending request for %T: %w", obj, err)
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusNotFound:
		return *new(T), &NotFoundError{fmt.Errorf("requesting resource at %s returned status code 404", url)}
	default:
		return *new(T), fmt.Errorf("unexpected status code %d while requesting resource", resp.StatusCode)
	}

	var newObj T
	if err := json.NewDecoder(resp.Body).Decode(&newObj); err != nil {
		return *new(T), fmt.Errorf("decoding %T: %w", obj, err)
	}

	if newObj.Validate() != nil {
		return *new(T), fmt.Errorf("received invalid %T: %w", newObj, newObj.Validate())
	}

	return newObj, nil
}

// NotFoundError is an error that is returned when a resource is not found.
type NotFoundError struct {
	err error
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("the requested resource was not found: %s", e.err.Error())
}

func (e *NotFoundError) Unwrap() error {
	return e.err
}

type httpc interface {
	Do(req *http.Request) (*http.Response, error)
}
