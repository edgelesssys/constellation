/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package image

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/spf13/afero"
)

// imageLookupTable is a lookup table for image references.
//
// Example:
//
//	{
//	  "aws": {
//	    "us-west-2": "ami-0123456789abcdef0"
//	  },
//	  "azure": {
//	    "cvm": "cvm-image-id"
//	  },
//	  "gcp": {
//	    "sev-es": "projects/<project>/global/images/<image>"
//	  },
//	  "qemu": {
//	    "default": "https://cdn.example.com/image.raw"
//	  }
//	}
type imageLookupTable map[string]map[string]string

// getReference returns the image reference for a given CSP and image variant.
func (l *imageLookupTable) getReference(csp, variant string) (string, error) {
	if l == nil {
		return "", fmt.Errorf("image lookup table is nil")
	}
	if _, ok := (*l)[csp]; !ok {
		return "", fmt.Errorf("image not available for CSP %q", csp)
	}
	if _, ok := (*l)[csp][variant]; !ok {
		return "", fmt.Errorf("image not available for variant %q", variant)
	}
	return (*l)[csp][variant], nil
}

// Fetcher fetches image references using a lookup table.
type Fetcher struct {
	httpc httpc
	fs    *afero.Afero
}

// New returns a new image fetcher.
func New() *Fetcher {
	return &Fetcher{
		httpc: http.DefaultClient,
		fs:    &afero.Afero{Fs: afero.NewOsFs()},
	}
}

// FetchReference fetches the image reference for a given image version uid, CSP and image variant.
func (f *Fetcher) FetchReference(ctx context.Context, config *config.Config) (string, error) {
	provider := config.GetProvider()
	variant, err := variant(provider, config)
	if err != nil {
		return "", err
	}
	return f.fetch(ctx, provider, config.Image, variant)
}

// fetch fetches the image reference for a given image version uid, CSP and image variant.
func (f *Fetcher) fetch(ctx context.Context, csp cloudprovider.Provider, version, variant string) (string, error) {
	raw, err := getFromFile(f.fs, version)
	if err != nil && os.IsNotExist(err) {
		raw, err = getFromURL(ctx, f.httpc, version)
	}
	if err != nil {
		return "", fmt.Errorf("fetching image reference: %w", err)
	}
	lut := make(imageLookupTable)
	if err := json.Unmarshal(raw, &lut); err != nil {
		return "", fmt.Errorf("decoding image reference: %w", err)
	}
	return lut.getReference(strings.ToLower(csp.String()), variant)
}

// variant returns the image variant for a given CSP and configuration.
func variant(provider cloudprovider.Provider, config *config.Config) (string, error) {
	switch provider {
	case cloudprovider.AWS:
		return config.Provider.AWS.Region, nil
	case cloudprovider.Azure:
		if *config.Provider.Azure.ConfidentialVM {
			return "cvm", nil
		}
		return "trustedlaunch", nil

	case cloudprovider.GCP:
		return "sev-es", nil
	case cloudprovider.QEMU:
		return "default", nil
	default:
		return "", fmt.Errorf("unsupported provider: %s", provider)
	}
}

func getFromFile(fs *afero.Afero, version string) ([]byte, error) {
	version = filepath.Base(version)
	return fs.ReadFile(version + ".json")
}

// getFromURL fetches the image lookup table from a URL.
func getFromURL(ctx context.Context, client httpc, version string) ([]byte, error) {
	url, err := url.Parse(constants.CDNRepositoryURL)
	if err != nil {
		return nil, fmt.Errorf("parsing image version repository URL: %w", err)
	}
	versionFilename := path.Base(version) + ".json"
	url.Path = path.Join(constants.CDNImagePath, versionFilename)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), http.NoBody)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		switch resp.StatusCode {
		case http.StatusNotFound:
			return nil, fmt.Errorf("image %q does not exist", version)
		default:
			return nil, fmt.Errorf("unexpected status code %d", resp.StatusCode)
		}
	}
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return content, nil
}

type httpc interface {
	Do(req *http.Request) (*http.Response, error)
}
