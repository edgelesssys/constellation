/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package gcp

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	// DefaultProjectID for Constellation images.
	DefaultProjectID = "constellation-images"
	// DefaultImageFamily for Constellation images.
	DefaultImageFamily = "constellation"
)

// Options for GCP image API client.
type Options struct {
	ProjectID   string
	ImageFamily string
	Filter      func(image string) (version string, err error)
}

// DefaultOptions creates an Options object with good defaults.
func DefaultOptions() Options {
	return Options{
		ProjectID:   DefaultProjectID,
		ImageFamily: DefaultImageFamily,
		Filter:      isGcpReleaseImage,
	}
}

func isGcpReleaseImage(image string) (imageVersion string, err error) {
	isReleaseRegEx := regexp.MustCompile(`^projects\/constellation-images\/global\/images\/constellation-v[\d]+-[\d]+-[\d]+$`)
	if !isReleaseRegEx.MatchString(image) {
		return "", fmt.Errorf("image does not look like release image")
	}
	findVersionRegEx := regexp.MustCompile(`v[\d]+-[\d]+-[\d]+$`)
	version := findVersionRegEx.FindString(image)
	semVer := strings.ReplaceAll(version, "-", ".")
	return semVer, nil
}
