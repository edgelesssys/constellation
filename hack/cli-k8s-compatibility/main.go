/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// cli-k8s-compatibility generates JSON output for a CLI version and its supported Kubernetes versions.
package main

import (
	"encoding/json"
	"fmt"

	"github.com/edgelesssys/constellation/v2/internal/versions"
)

// cliCompatibilityConfig describes the compatibility of a CLI version with Kubernetes versions.
// It is specified in rfc/cli-api.md.
type cliCompatibilityConfig struct {
	Version    string   `json:"version"`
	Ref        string   `json:"ref"`
	Stream     string   `json:"stream"`
	Kubernetes []string `json:"kubernetes"`
}

func main() {
	compatibilityFile := cliCompatibilityConfig{
		Kubernetes: []string{},
	}

	for _, v := range versions.VersionConfigs {
		compatibilityFile.Kubernetes = append(compatibilityFile.Kubernetes, v.ClusterVersion)
	}

	output, err := json.Marshal(compatibilityFile)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(output))
}
