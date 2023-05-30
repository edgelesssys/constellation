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
	"io"
	"net/http"
	"net/url"
)

// fetcher fetches versions API resources without authentication.
type fetcher struct {
	httpc HTTPClient
}

// NewHTTPClient returns a new http client.
func NewHTTPClient() HTTPClient {
	return &http.Client{Transport: &http.Transport{DisableKeepAlives: true}} // DisableKeepAlives fixes concurrency issue see https://stackoverflow.com/a/75816347
}

func newFetcherWith(client HTTPClient) *fetcher {
	return &fetcher{
		httpc: client,
	}
}

func newFetcher() *fetcher {
	return newFetcherWith(NewHTTPClient()) // DisableKeepAlives fixes concurrency issue see https://stackoverflow.com/a/75816347
}

type apiObject interface {
	ValidateRequest() error
	Validate() error
	URL() (string, error)
}

// FetchFromURL fetches the content from the given URL and returns the content as a byte slice.
func FetchFromURL(ctx context.Context, client HTTPClient, sourceURL *url.URL) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, sourceURL.String(), http.NoBody)
	if err != nil {
		return []byte{}, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return []byte{}, fmt.Errorf("http status code: %d", resp.StatusCode)
	}
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}
	return content, nil
}

func fetch[T apiObject](ctx context.Context, c HTTPClient, obj T) (T, error) {
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

// HTTPClient is an interface for http clients.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}
