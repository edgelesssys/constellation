/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// cli-k8s-compatibility generates JSON output for a CLI version and its supported Kubernetes versions.
package main

import (
	"context"
	"flag"

	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/versions"
	"github.com/edgelesssys/constellation/v2/internal/versionsapi"
	"github.com/edgelesssys/constellation/v2/internal/versionsapi/client"
	"go.uber.org/zap/zapcore"
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

	cliInfo := versionsapi.CLIInfo{
		Ref:        *refFlag,
		Stream:     *streamFlag,
		Version:    *versionFlag,
		Kubernetes: []string{},
	}

	for _, v := range versions.VersionConfigs {
		cliInfo.Kubernetes = append(cliInfo.Kubernetes, v.ClusterVersion)
	}

	c, err := client.NewClient(context.Background(), "eu-central-1", "cdn-constellation-backend", "", false, logger.New(logger.PlainLog, zapcore.DebugLevel))
	if err != nil {
		panic(err)
	}
	
	if err := c.UpdateCLIInfo(context.Background(), cliInfo); err != nil {
		panic(err)
	}
}
