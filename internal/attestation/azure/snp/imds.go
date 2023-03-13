/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package snp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/edgelesssys/constellation/v2/internal/cloud/azure"
)

// Modified version of bootstrapper/cloudprovider/azure/imds.go

const (
	imdsVcekURL = "http://169.254.169.254/metadata/THIM/amd/certification"
	tagMAAURL   = "constellation-maa-url"
)

type imdsClient struct {
	client     *http.Client
	imdsClient azure.IMDSClient
}

func newIMDSClient() *imdsClient {
	return &imdsClient{
		client:     &http.Client{Transport: &http.Transport{Proxy: nil}},
		imdsClient: azure.NewIMDSClient(),
	}
}

func (c *imdsClient) getMAAURL(ctx context.Context) (string, error) {
	tags, err := c.imdsClient.Tags(ctx)
	if err != nil {
		return "", fmt.Errorf("getting tags: %w", err)
	}

	return tags[tagMAAURL], nil
}

// Retrieve retrieves instance metadata from the azure imds API.
func (c *imdsClient) getVcek(ctx context.Context) (vcekResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imdsVcekURL, http.NoBody)
	if err != nil {
		return vcekResponse{}, err
	}
	req.Header.Add("Metadata", "True")
	resp, err := c.client.Do(req)
	if err != nil {
		return vcekResponse{}, err
	}
	defer resp.Body.Close()

	var res vcekResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return vcekResponse{}, err
	}

	return res, nil
}

type vcekResponse struct {
	VcekCert         string
	CertificateChain string
}
