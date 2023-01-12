/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package license provides functions to check a user's Constellation license.
package license

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

const (
	// CommunityLicense is used by everyone who has not bought an enterprise license.
	CommunityLicense = "00000000-0000-0000-0000-000000000000"
	apiHost          = "license.confidential.cloud"
	licensePath      = "api/v1/license"
)

type (
	// Action performed by Constellation.
	Action string
)

const (
	// Init action denotes the initialization of a Constellation cluster.
	Init Action = "init"
	// test action is only to be used in testing.
	test Action = "test"
)

// Client interacts with the ES license server.
type Client struct {
	httpClient *http.Client
}

// NewClient creates a new client to interact with ES license server.
func NewClient() *Client {
	return &Client{
		httpClient: http.DefaultClient,
	}
}

// QuotaCheckRequest is JSON request to license server to check quota for a given license and action.
type QuotaCheckRequest struct {
	Action   Action `json:"action"`
	Provider string `json:"provider"`
	License  string `json:"license"`
}

// QuotaCheckResponse is JSON response by license server.
type QuotaCheckResponse struct {
	Quota int `json:"quota"`
}

// QuotaCheck for a given license and action, passed via CheckQuotaRequest.
func (c *Client) QuotaCheck(ctx context.Context, checkRequest QuotaCheckRequest) (QuotaCheckResponse, error) {
	reqBody, err := json.Marshal(checkRequest)
	if err != nil {
		return QuotaCheckResponse{}, fmt.Errorf("unable to marshal input: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, licenseURL().String(), bytes.NewBuffer(reqBody))
	if err != nil {
		return QuotaCheckResponse{}, fmt.Errorf("unable to create request: %w", err)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return QuotaCheckResponse{}, fmt.Errorf("unable to do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return QuotaCheckResponse{}, fmt.Errorf("http error %d", resp.StatusCode)
	}

	responseContentType := resp.Header.Get("Content-Type")
	if responseContentType != "application/json" {
		return QuotaCheckResponse{}, fmt.Errorf("expected server JSON response but got '%s'", responseContentType)
	}

	var parsedResponse QuotaCheckResponse
	err = json.NewDecoder(resp.Body).Decode(&parsedResponse)
	if err != nil {
		return QuotaCheckResponse{}, fmt.Errorf("unable to parse response: %w", err)
	}

	return parsedResponse, nil
}

func licenseURL() *url.URL {
	return &url.URL{
		Scheme: "https",
		Host:   apiHost,
		Path:   licensePath,
	}
}

// QuotaChecker checks the vCPU quota for a given license.
type QuotaChecker interface {
	QuotaCheck(ctx context.Context, checkRequest QuotaCheckRequest) (QuotaCheckResponse, error)
}
