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
	"strings"

	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/shortname"
	"github.com/spf13/afero"
)

// imageName is a struct that describes a Constellation OS imageName name.
type imageName struct {
	Ref     string
	Stream  string
	Version string
}

func newImageName(name string) (*imageName, error) {
	ref, stream, version, err := shortname.ToParts(name)
	if err != nil {
		return nil, err
	}
	return &imageName{
		Ref:     ref,
		Stream:  stream,
		Version: version,
	}, nil
}

func (i *imageName) infoPath() string {
	return path.Join(constants.CDNAPIPrefix, "ref", i.Ref, "stream", i.Stream, "image", i.Version, "info.json")
}

func (i *imageName) shortname() string {
	return shortname.FromParts(i.Ref, i.Stream, i.Version)
}

// filename is the override file name for the image info file.
func (i *imageName) filename() string {
	name := i.shortname()
	// replace all non-alphanumeric characters with an underscore
	name = strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '.' {
			return r
		}
		return '_'
	}, name)
	return name + ".json"
}

// imageInfo is a lookup table for image references.
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
//	},
//	"version": "1.0.0",
//	"debug": false
//	}
type imageInfo struct {
	AWS     map[string]string `json:"aws,omitempty"`
	Azure   map[string]string `json:"azure,omitempty"`
	GCP     map[string]string `json:"gcp,omitempty"`
	QEMU    map[string]string `json:"qemu,omitempty"`
	Debug   bool              `json:"debug,omitempty"`
	Version string            `json:"version,omitempty"`
}

// getReference returns the image reference for a given CSP and image variant.
func (l *imageInfo) getReference(csp, variant string) (string, error) {
	if l == nil {
		return "", fmt.Errorf("image info is nil")
	}

	var cspList map[string]string
	switch cloudprovider.FromString(csp) {
	case cloudprovider.AWS:
		cspList = l.AWS
	case cloudprovider.Azure:
		cspList = l.Azure
	case cloudprovider.GCP:
		cspList = l.GCP
	case cloudprovider.QEMU:
		cspList = l.QEMU
	default:
		return "", fmt.Errorf("image not available for CSP %q", csp)
	}

	if cspList == nil {
		return "", fmt.Errorf("image not available for CSP %q", csp)
	}

	ref, ok := cspList[variant]
	if !ok {
		return "", fmt.Errorf("image not available for variant %q", variant)
	}

	return ref, nil
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
	image, err := newImageName(config.Image)
	if err != nil {
		return "", err
	}
	return f.fetch(ctx, provider, image, variant)
}

// fetch fetches the image reference for a given image name, uid, CSP and image variant.
func (f *Fetcher) fetch(ctx context.Context, csp cloudprovider.Provider, img *imageName, variant string) (string, error) {
	raw, err := getFromFile(f.fs, img)
	if err != nil && os.IsNotExist(err) {
		raw, err = getFromURL(ctx, f.httpc, img)
	}
	if err != nil {
		return "", fmt.Errorf("fetching image reference: %w", err)
	}
	var info imageInfo
	if err := json.Unmarshal(raw, &info); err != nil {
		return "", fmt.Errorf("decoding image reference: %w", err)
	}
	return info.getReference(strings.ToLower(csp.String()), variant)
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

func getFromFile(fs *afero.Afero, img *imageName) ([]byte, error) {
	return fs.ReadFile(img.filename())
}

// getFromURL fetches the image lookup table from a URL.
func getFromURL(ctx context.Context, client httpc, img *imageName) ([]byte, error) {
	url, err := url.Parse(constants.CDNRepositoryURL)
	if err != nil {
		return nil, fmt.Errorf("parsing image version repository URL: %w", err)
	}
	url.Path = img.infoPath()
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
			return nil, fmt.Errorf("image %q does not exist", img.shortname())
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
