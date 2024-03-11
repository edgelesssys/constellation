/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/
package clouds

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"

	"github.com/edgelesssys/constellation/v2/internal/file"
)

// ReadCloudsYAML reads a clouds.yaml file and returns its contents.
func ReadCloudsYAML(fileHandler file.Handler, path string) (Clouds, error) {
	// Order of operations as performed by the OpenStack CLI:

	// Define a search path for clouds.yaml:
	// 1. If OS_CLIENT_CONFIG_FILE is set, use it as search path
	// 2. Otherwise, use the following paths:
	//    - current directory
	//    - `openstack` directory under standard user config directory (e.g. ~/.config/openstack)
	//    - /etc/openstack (Unix only)

	var searchPaths []string
	if path != "" {
		expanded, err := homedir.Expand(path)
		if err == nil {
			searchPaths = append(searchPaths, expanded)
		} else {
			searchPaths = append(searchPaths, path)
		}
	} else if osClientConfigFile := os.Getenv("OS_CLIENT_CONFIG_FILE"); osClientConfigFile != "" {
		searchPaths = append(searchPaths, filepath.Join(osClientConfigFile, "clouds.yaml"))
	} else {
		searchPaths = append(searchPaths, "clouds.yaml")
		confDir, err := os.UserConfigDir()
		if err != nil {
			return Clouds{}, fmt.Errorf("getting user config directory: %w", err)
		}
		searchPaths = append(searchPaths, filepath.Join(confDir, "openstack", "clouds.yaml"))
		if os.PathSeparator == '/' {
			searchPaths = append(searchPaths, "/etc/openstack/clouds.yaml")
		}
	}

	var cloudsYAML Clouds
	for _, path := range searchPaths {
		if err := fileHandler.ReadYAML(path, &cloudsYAML); err == nil {
			return cloudsYAML, nil
		}
	}

	return Clouds{}, fmt.Errorf("clouds.yaml not found in search paths: %v", searchPaths)
}
