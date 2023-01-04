/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package versionsapi

import (
	"errors"
	"fmt"
	"net/url"
	"path"

	"github.com/edgelesssys/constellation/v2/internal/constants"
	"go.uber.org/multierr"
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
	// AWS is a map of AWS regions to AMI IDs.
	AWS map[string]string `json:"aws,omitempty"`
	// Azure is a map of image types to Azure image IDs.
	Azure map[string]string `json:"azure,omitempty"`
	// GCP is a map of image types to GCP image IDs.
	GCP map[string]string `json:"gcp,omitempty"`
	// QEMU is a map of image types to QEMU image URLs.
	QEMU map[string]string `json:"qemu,omitempty"`
}

// JSONPath returns the S3 JSON path for this object.
func (i ImageInfo) JSONPath() string {
	return path.Join(
		constants.CDNAPIPrefix,
		"ref", i.Ref,
		"stream", i.Stream,
		"image", i.Version,
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
		retErr = multierr.Append(retErr, err)
	}
	if err := ValidateStream(i.Ref, i.Stream); err != nil {
		retErr = multierr.Append(retErr, err)
	}
	if !semver.IsValid(i.Version) {
		retErr = multierr.Append(retErr, fmt.Errorf("version %q is not a valid semver", i.Version))
	}
	if len(i.AWS) != 0 {
		retErr = multierr.Append(retErr, errors.New("AWS map must be empty for request"))
	}
	if len(i.Azure) != 0 {
		retErr = multierr.Append(retErr, errors.New("Azure map must be empty for request"))
	}
	if len(i.GCP) != 0 {
		retErr = multierr.Append(retErr, errors.New("GCP map must be empty for request"))
	}
	if len(i.QEMU) != 0 {
		retErr = multierr.Append(retErr, errors.New("QEMU map must be empty for request"))
	}

	return retErr
}

// Validate checks if the image info is valid.
func (i ImageInfo) Validate() error {
	var retErr error
	if err := ValidateRef(i.Ref); err != nil {
		retErr = multierr.Append(retErr, err)
	}
	if err := ValidateStream(i.Ref, i.Stream); err != nil {
		retErr = multierr.Append(retErr, err)
	}
	if !semver.IsValid(i.Version) {
		retErr = multierr.Append(retErr, fmt.Errorf("version %q is not a valid semver", i.Version))
	}
	if len(i.AWS) == 0 {
		retErr = multierr.Append(retErr, errors.New("AWS map must not be empty"))
	}
	if len(i.Azure) == 0 {
		retErr = multierr.Append(retErr, errors.New("Azure map must not be empty"))
	}
	if len(i.GCP) == 0 {
		retErr = multierr.Append(retErr, errors.New("GCP map must not be empty"))
	}
	if len(i.QEMU) == 0 {
		retErr = multierr.Append(retErr, errors.New("QEMU map must not be empty"))
	}

	return retErr
}
