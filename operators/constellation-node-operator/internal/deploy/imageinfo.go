/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: BUSL-1.1
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
func (i *ImageInfo) ImageVersion() (string, error) {
	var version string
	var err error
	for _, path := range osReleasePaths {
		version, err = i.getOSReleaseImageVersion(path)
		if err == nil {
			break
		}
	}
	if err != nil {
		return "", err
	}
	if version == "" {
		return "", fmt.Errorf("IMAGE_VERSION not found in %s", strings.Join(osReleasePaths, ", "))
	}
	return version, nil
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

const versionKey = "IMAGE_VERSION"

var (
	osReleaseLine     = regexp.MustCompile(`^(?P<name>[a-zA-Z0-9_]+)=("(?P<v1>.*)"|'(?P<v2>.*)'|(?P<v3>[^\n"']+))$`)
	osReleaseUnescape = regexp.MustCompile(`\\([\\\$\"\'` + "`" + `])`)
	osReleasePaths    = []string{
		"/host/etc/os-release",
		"/host/usr/lib/os-release",
	}
)
