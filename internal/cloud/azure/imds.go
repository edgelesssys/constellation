package azure

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

// subset of azure imds API: https://docs.microsoft.com/en-us/azure/virtual-machines/windows/instance-metadata-service?tabs=linux
// this is not yet available through the azure sdk (see https://github.com/Azure/azure-rest-api-specs/issues/4408)

const (
	imdsURL        = "http://169.254.169.254/metadata/instance"
	imdsAPIVersion = "2021-02-01"
)

type imdsClient struct {
	client *http.Client
}

// Retrieve retrieves instance metadata from the azure imds API.
func (c *imdsClient) Retrieve(ctx context.Context) (metadataResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", imdsURL, http.NoBody)
	if err != nil {
		return metadataResponse{}, err
	}
	req.Header.Add("Metadata", "True")
	query := req.URL.Query()
	query.Add("format", "json")
	query.Add("api-version", imdsAPIVersion)
	req.URL.RawQuery = query.Encode()
	resp, err := c.client.Do(req)
	if err != nil {
		return metadataResponse{}, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return metadataResponse{}, err
	}
	var res metadataResponse
	if err := json.Unmarshal(body, &res); err != nil {
		return metadataResponse{}, err
	}
	return res, nil
}

// metadataResponse contains metadataResponse with only the required values.
type metadataResponse struct {
	Compute struct {
		ResourceID string `json:"resourceId,omitempty"`
	} `json:"compute,omitempty"`
}
