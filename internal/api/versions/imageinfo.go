/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package versions

import (
	"errors"
	"fmt"
	"net/url"
	"path"
	"sort"

	"github.com/edgelesssys/constellation/v2/internal/constants"
	"golang.org/x/mod/semver"
)

// ImageInfo contains information about the OS images for a specific version.
type ImageInfo struct {
	// Ref is the reference name of the image.
	Ref string `json:"ref,omitempty"`
	// Stream is the stream name of the image.
	Stream string `json:"stream,omitempty"`
	// Version is the version of the image.
	Version string `json:"version,omitempty"`
	// List contains the image variants for this version.
	List []ImageInfoEntry `json:"list,omitempty"`
}

// ImageInfoEntry contains information about a single image variant.
type ImageInfoEntry struct {
	CSP                string `json:"csp"`
	AttestationVariant string `json:"attestationVariant"`
	Reference          string `json:"reference"`
	Region             string `json:"region,omitempty"`
}

// JSONPath returns the S3 JSON path for this object.
func (i ImageInfo) JSONPath() string {
	return path.Join(
		constants.CDNAPIPrefixV2,
		"ref", i.Ref,
		"stream", i.Stream,
		i.Version,
		"image",
		"info.json",
	)
}

// URL returns the URL to the JSON file for this object.
func (i ImageInfo) URL() (string, error) {
	url, err := url.Parse(constants.CDNRepositoryURL)
	if err != nil {
		return "", fmt.Errorf("parsing CDN URL: %w", err)
	}
	url.Path = i.JSONPath()
	return url.String(), nil
}

// ValidateRequest validates the request parameters of the list.
// The provider specific maps must be empty.
func (i ImageInfo) ValidateRequest() error {
	var retErr error
	if err := ValidateRef(i.Ref); err != nil {
		retErr = errors.Join(retErr, err)
	}
	if err := ValidateStream(i.Ref, i.Stream); err != nil {
		retErr = errors.Join(retErr, err)
	}
	if !semver.IsValid(i.Version) {
		retErr = errors.Join(retErr, fmt.Errorf("version %q is not a valid semver", i.Version))
	}
	if len(i.List) != 0 {
		retErr = errors.Join(retErr, errors.New("list must be empty for request"))
	}

	return retErr
}

// Validate checks if the image info is valid.
func (i ImageInfo) Validate() error {
	var retErr error
	if err := ValidateRef(i.Ref); err != nil {
		retErr = errors.Join(retErr, err)
	}
	if err := ValidateStream(i.Ref, i.Stream); err != nil {
		retErr = errors.Join(retErr, err)
	}
	if !semver.IsValid(i.Version) {
		retErr = errors.Join(retErr, fmt.Errorf("version %q is not a valid semver", i.Version))
	}
	if len(i.List) == 0 {
		retErr = errors.Join(retErr, errors.New("one or more image variants must be specified in the list"))
	}

	return retErr
}

// MergeImageInfos combines the image info entries from multiple sources into a single
// image info object.
func MergeImageInfos(infos ...ImageInfo) (ImageInfo, error) {
	if len(infos) == 0 {
		return ImageInfo{}, errors.New("no image info objects specified")
	}
	if len(infos) == 1 {
		return infos[0], nil
	}
	out := ImageInfo{
		Ref:     infos[0].Ref,
		Stream:  infos[0].Stream,
		Version: infos[0].Version,
		List:    []ImageInfoEntry{},
	}
	for _, info := range infos {
		if info.Ref != out.Ref {
			return ImageInfo{}, errors.New("image info objects have different refs")
		}
		if info.Stream != out.Stream {
			return ImageInfo{}, errors.New("image info objects have different streams")
		}
		if info.Version != out.Version {
			return ImageInfo{}, errors.New("image info objects have different versions")
		}
		out.List = append(out.List, info.List...)
	}
	sort.SliceStable(out.List, func(i, j int) bool {
		if out.List[i].CSP != out.List[j].CSP {
			return out.List[i].CSP < out.List[j].CSP
		}
		if out.List[i].AttestationVariant != out.List[j].AttestationVariant {
			return out.List[i].AttestationVariant < out.List[j].AttestationVariant
		}
		if out.List[i].Region != out.List[j].Region {
			return out.List[i].Region < out.List[j].Region
		}
		return out.List[i].Reference < out.List[j].Reference
	})
	return out, nil
}
