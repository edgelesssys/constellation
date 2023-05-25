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
)

// fetcher fetches versions API resources without authentication.
type fetcher struct {
	httpc HttpClienter
}

// NewHTTPClient returns a new http client.
func NewHTTPClient() HttpClienter {
	return &http.Client{Transport: &http.Transport{DisableKeepAlives: true}} // DisableKeepAlives fixes concurrency issue see https://stackoverflow.com/a/75816347
}

func newFetcherWith(client HttpClienter) *fetcher {
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

func fetch[T apiObject](ctx context.Context, c HttpClienter, obj T) (T, error) {
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

// HttpClienter is an interface for http clients.
type HttpClienter interface {
	Do(req *http.Request) (*http.Response, error)
}
