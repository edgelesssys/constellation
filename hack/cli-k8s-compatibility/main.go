/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// cli-k8s-compatibility generates JSON output for a CLI version and its supported Kubernetes versions.
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/edgelesssys/constellation/v2/internal/api/versionsapi"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/versions"
)

var (
	refFlag     = flag.String("ref", "", "the reference name of the image")
	streamFlag  = flag.String("stream", "", "the stream name of the image")
	versionFlag = flag.String("version", "", "the version of the image")
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	ctx := context.Background()

	flag.Parse()
	if *refFlag == "" {
		log.Error("ref must be set")
		os.Exit(1)
	}
	if *streamFlag == "" {
		log.Error("stream must be set")
		os.Exit(1)
	}
	if *versionFlag == "" {
		log.Error("version must be set")
		os.Exit(1)
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

	c, cclose, err := versionsapi.NewClient(ctx, "eu-central-1", "cdn-constellation-backend", constants.CDNDefaultDistributionID, false, log)
	if err != nil {
		log.Error(fmt.Sprintf("creating s3 client: %s", err))
		os.Exit(1)
	}
	defer func() {
		if err := cclose(ctx); err != nil {
			log.Error(fmt.Sprintf("invalidating cache: %s", err))
			os.Exit(1)
		}
	}()

	if err := c.UpdateCLIInfo(ctx, cliInfo); err != nil {
		log.Error(fmt.Sprintf("updating cli info: %s", err))
		os.Exit(1)
	}
}
