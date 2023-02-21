/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// cli-k8s-compatibility generates JSON output for a CLI version and its supported Kubernetes versions.
package main

import (
	"encoding/json"
	"flag"
	"fmt"

	"github.com/edgelesssys/constellation/v2/internal/versions"
	"github.com/edgelesssys/constellation/v2/internal/versionsapi"
)

var (
	refFlag     = flag.String("ref", "", "the reference name of the image")
	streamFlag  = flag.String("stream", "", "the stream name of the image")
	versionFlag = flag.String("version", "", "the version of the image")
)

func main() {
	flag.Parse()
	if *refFlag == "" {
		panic("ref must be set")
	}
	if *streamFlag == "" {
		panic("stream must be set")
	}
	if *versionFlag == "" {
		panic("version must be set")
	}

	compatibilityFile := versionsapi.CLIInfo{
		Ref:        *refFlag,
		Stream:     *streamFlag,
		Version:    *versionFlag,
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
