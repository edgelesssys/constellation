/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package deploy

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"

	"github.com/spf13/afero"
)

// ImageInfo retrieves OS image information.
type ImageInfo struct {
	fs *afero.Afero
}

// NewImageInfo creates a new imageInfo.
func NewImageInfo() *ImageInfo {
	return &ImageInfo{
		fs: &afero.Afero{Fs: afero.NewOsFs()},
	}
}

// ImageVersion tries to parse the image version from the host mounted os-release file.
// If the file is not present or does not contain the version, a fallback lookup is performed.
func (i *ImageInfo) ImageVersion(imageReference string) (string, error) {
	var version string
	var err error
	for _, path := range osReleasePaths {
		version, err = i.getOSReleaseImageVersion(path)
		if err == nil {
			break
		}
	}
	if version != "" {
		return version, nil
	}
	return imageVersionFromFallback(imageReference)
}

// getOSReleaseImageVersion reads the os-release file and returns the image version (if present).
func (i *ImageInfo) getOSReleaseImageVersion(path string) (string, error) {
	osRelease, err := i.fs.Open(path)
	if err != nil {
		return "", err
	}
	defer osRelease.Close()
	osReleaseMap, err := parseOSRelease(bufio.NewScanner(osRelease))
	if err != nil {
		return "", err
	}
	version, ok := osReleaseMap[versionKey]
	if !ok {
		return "", fmt.Errorf("IMAGE_VERSION not found in %s", path)
	}
	return version, nil
}

// parseOSRelease parses the os-release file and returns a map of key-value pairs.
// The os-release file is a simple key-value file.
// The format is specified in https://www.freedesktop.org/software/systemd/man/os-release.html.
func parseOSRelease(osRelease *bufio.Scanner) (map[string]string, error) {
	osReleaseMap := make(map[string]string)
	for osRelease.Scan() {
		line := osRelease.Text()
		matches := osReleaseLine.FindStringSubmatch(line)
		if len(matches) < 6 {
			continue
		}
		key := matches[1]
		var value string
		// group 3 is the value with double quotes
		// group 4 is the value with single quotes
		// group 5 is the value without quotes
		for i := 3; i < 6; i++ {
			if matches[i] != "" {
				value = matches[i]
				break
			}
		}
		// unescape the following characters: \\, \$, \", \', \`
		value = osReleaseUnescape.ReplaceAllString(value, "$1")
		osReleaseMap[key] = value
	}
	if err := osRelease.Err(); err != nil {
		return nil, err
	}
	return osReleaseMap, nil
}

// imageVersionFromFallback tries to guess the image version from the image reference.
// It is a fallback mechanism in case the os-release file is not present or does not contain the version.
// This was the case for older images (< v2.3.0).
func imageVersionFromFallback(imageReference string) (string, error) {
	version, ok := fallbackLookup[strings.ToLower(imageReference)]
	if !ok {
		return "", fmt.Errorf("image version not found in fallback lookup")
	}
	return version, nil
}

const versionKey = "IMAGE_VERSION"

var (
	osReleaseLine     = regexp.MustCompile(`^(?P<name>[a-zA-Z0-9_]+)=("(?P<v1>.*)"|'(?P<v2>.*)'|(?P<v3>[^\n"']+))$`)
	osReleaseUnescape = regexp.MustCompile(`\\([\\\$\"\'` + "`" + `])`)
	osReleasePaths    = []string{
		"/host/etc/os-release",
		"/host/usr/lib/os-release",
	}

	fallbackLookup = map[string]string{
		// AWS
		"ami-06b8cbf4837a0a57c": "v2.2.2",
		"ami-02e96dc04a9e438cd": "v2.2.2",
		"ami-028ead928a9034b2f": "v2.2.2",
		"ami-032ac10dd8d8266e3": "v2.2.1",
		"ami-032e0d57cc4395088": "v2.2.1",
		"ami-053c3e49e19b96bdd": "v2.2.1",
		"ami-0e27ebcefc38f648b": "v2.2.0",
		"ami-098cd37f66523b7c3": "v2.2.0",
		"ami-04a87d302e2509aad": "v2.2.0",

		// Azure
		"/communitygalleries/constellationcvm-b3782fa0-0df7-4f2f-963e-fc7fc42663df/images/constellation/versions/2.2.2":                                                                       "v2.2.2",
		"/subscriptions/0d202bbb-4fa7-4af8-8125-58c269a05435/resourcegroups/constellation-images/providers/microsoft.compute/galleries/constellation/images/constellation/versions/2.2.2":     "v2.2.2",
		"/subscriptions/0d202bbb-4fa7-4af8-8125-58c269a05435/resourcegroups/constellation-images/providers/microsoft.compute/galleries/constellation_cvm/images/constellation/versions/2.2.2": "v2.2.2",
		"/communitygalleries/constellationcvm-b3782fa0-0df7-4f2f-963e-fc7fc42663df/images/constellation/versions/2.2.1":                                                                       "v2.2.1",
		"/subscriptions/0d202bbb-4fa7-4af8-8125-58c269a05435/resourcegroups/constellation-images/providers/microsoft.compute/galleries/constellation/images/constellation/versions/2.2.1":     "v2.2.1",
		"/subscriptions/0d202bbb-4fa7-4af8-8125-58c269a05435/resourcegroups/constellation-images/providers/microsoft.compute/galleries/constellation_cvm/images/constellation/versions/2.2.1": "v2.2.1",
		"/communitygalleries/constellationcvm-b3782fa0-0df7-4f2f-963e-fc7fc42663df/images/constellation/versions/2.2.0":                                                                       "v2.2.0",
		"/subscriptions/0d202bbb-4fa7-4af8-8125-58c269a05435/resourcegroups/constellation-images/providers/microsoft.compute/galleries/constellation/images/constellation/versions/2.2.0":     "v2.2.0",
		"/subscriptions/0d202bbb-4fa7-4af8-8125-58c269a05435/resourcegroups/constellation-images/providers/microsoft.compute/galleries/constellation_cvm/images/constellation/versions/2.2.0": "v2.2.0",
		"/communitygalleries/constellationcvm-b3782fa0-0df7-4f2f-963e-fc7fc42663df/images/constellation/versions/2.1.0":                                                                       "v2.1.0",
		"/subscriptions/0d202bbb-4fa7-4af8-8125-58c269a05435/resourcegroups/constellation-images/providers/microsoft.compute/galleries/constellation/images/constellation/versions/2.1.0":     "v2.1.0",
		"/subscriptions/0d202bbb-4fa7-4af8-8125-58c269a05435/resourcegroups/constellation-images/providers/microsoft.compute/galleries/constellation_cvm/images/constellation/versions/2.1.0": "v2.1.0",
		"/communitygalleries/constellationcvm-b3782fa0-0df7-4f2f-963e-fc7fc42663df/images/constellation/versions/2.0.0":                                                                       "v2.0.0",
		"/subscriptions/0d202bbb-4fa7-4af8-8125-58c269a05435/resourcegroups/constellation-images/providers/microsoft.compute/galleries/constellation/images/constellation/versions/2.0.0":     "v2.0.0",
		"/subscriptions/0d202bbb-4fa7-4af8-8125-58c269a05435/resourcegroups/constellation-images/providers/microsoft.compute/galleries/constellation_cvm/images/constellation/versions/2.0.0": "v2.0.0",

		// GCP
		"projects/constellation-images/global/images/constellation-v2-2-2": "v2.2.2",
		"projects/constellation-images/global/images/constellation-v2-2-1": "v2.2.1",
		"projects/constellation-images/global/images/constellation-v2-2-0": "v2.2.0",
		"projects/constellation-images/global/images/constellation-v2-1-0": "v2.1.0",
		"projects/constellation-images/global/images/constellation-v2-0-0": "v2.0.0",
	}
)
