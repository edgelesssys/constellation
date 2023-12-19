//go:build enterprise

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package license

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
)

// action performed by Constellation.
type action string

const (
	apiHost     = "license.confidential.cloud"
	licensePath = "api/v1/license"

	// initAction denotes the initialization of a Constellation cluster.
	initAction action = "init"
	// applyAction denotes an update of a Constellation cluster.
	// It is used after a cluster has already been initialized once.
	applyAction action = "apply"
)

// Checker checks the Constellation license.
type Checker struct {
	httpClient *http.Client
}

// NewChecker creates a new Checker.
func NewChecker() *Checker {
	return &Checker{
		httpClient: http.DefaultClient,
	}
}

// CheckLicense checks the Constellation license.
func (c *Checker) CheckLicense(ctx context.Context, csp cloudprovider.Provider, initRequest bool, licenseID string) (int, error) {
	checkRequest := quotaCheckRequest{
		Provider: csp.String(),
		License:  licenseID,
		Action:   applyAction,
	}
	if initRequest {
		checkRequest.Action = initAction
	}

	reqBody, err := json.Marshal(checkRequest)
	if err != nil {
		return 0, fmt.Errorf("unable to marshal input: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, licenseURL().String(), bytes.NewBuffer(reqBody))
	if err != nil {
		return 0, fmt.Errorf("unable to create request: %w", err)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("unable to do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("http error %d", resp.StatusCode)
	}

	responseContentType := resp.Header.Get("Content-Type")
	if responseContentType != "application/json" {
		return 0, fmt.Errorf("expected server JSON response but got '%s'", responseContentType)
	}

	var parsedResponse quotaCheckResponse
	err = json.NewDecoder(resp.Body).Decode(&parsedResponse)
	if err != nil {
		return 0, fmt.Errorf("unable to parse response: %w", err)
	}

	return parsedResponse.Quota, nil
}

// quotaCheckRequest is JSON request to license server to check quota for a given license and action.
type quotaCheckRequest struct {
	Action   action `json:"action"`
	Provider string `json:"provider"`
	License  string `json:"license"`
}

// quotaCheckResponse is JSON response by license server.
type quotaCheckResponse struct {
	Quota int `json:"quota"`
}

func licenseURL() *url.URL {
	return &url.URL{
		Scheme: "https",
		Host:   apiHost,
		Path:   licensePath,
	}
}
