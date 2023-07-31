/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
Package fetcher implements a client for the Constellation Resource API.

The fetcher is used to get information from the versions API without having to
authenticate with AWS, where the API is currently hosted. This package should be
used in user-facing application code most of the time, like the CLI.

Each sub-API included in the Constellation Resource API should define it's resources by implementing types that implement apiObject.
The new package can then call this package's Fetch function to get the resource from the API.
To modify resources, pkg internal/api/client should be used in a similar fashion.
*/
package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/edgelesssys/constellation/v2/internal/sigstore"
)

// NewHTTPClient returns a new http client.
func NewHTTPClient() HTTPClient {
	return &http.Client{Transport: &http.Transport{DisableKeepAlives: true}} // DisableKeepAlives fixes concurrency issue see https://stackoverflow.com/a/75816347
}

// Fetch fetches the given apiObject from the public Constellation CDN.
// Fetch does not require authentication.
func Fetch[T apiObject](ctx context.Context, c HTTPClient, obj T) (T, error) {
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

// FetchAndVerify fetches the given apiObject, checks if it can fetch an accompanying signature and verifies if the signature matches the found object.
// The public key used to verify the signature is embedded in the verifier argument.
// FetchAndVerify uses a generic to return a new object of type T.
// Otherwise the caller would have to cast the interface type to a concrete object, which could fail.
func FetchAndVerify[T apiObject](ctx context.Context, c HTTPClient, obj T, cosignVerifier sigstore.Verifier) (T, error) {
	fetchedObj, err := Fetch(ctx, c, obj)
	if err != nil {
		return fetchedObj, fmt.Errorf("fetching object: %w", err)
	}
	marshalledObj, err := json.Marshal(fetchedObj)
	if err != nil {
		return fetchedObj, fmt.Errorf("marshalling object: %w", err)
	}

	signature, err := Fetch(ctx, c, signature{Signed: fetchedObj})
	if err != nil {
		return fetchedObj, fmt.Errorf("fetching signature: %w", err)
	}

	err = cosignVerifier.VerifySignature(marshalledObj, signature.Signature)
	if err != nil {
		return fetchedObj, fmt.Errorf("verifying signature: %w", err)
	}
	return fetchedObj, nil
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

type apiObject interface {
	ValidateRequest() error
	Validate() error
	URL() (string, error)
}

// signature wraps another APIObject and adds a signature to it.
type signature struct {
	// Signed is the object that is signed.
	Signed apiObject `json:"-"`
	// Signature is the signature of `Signed`.
	Signature []byte `json:"signature"`
}

// URL returns the URL for the request to the config api.
func (s signature) URL() (string, error) {
	url, err := s.Signed.URL()
	if err != nil {
		return "", err
	}
	return url + ".sig", nil
}

// ValidateRequest validates the request.
func (s signature) ValidateRequest() error {
	return s.Signed.ValidateRequest()
}

// Validate is a No-Op at the moment.
func (s signature) Validate() error {
	return s.Signed.Validate()
}
