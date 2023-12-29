/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// cli-k8s-compatibility generates JSON output for a CLI version and its supported Kubernetes versions.
package main

import (
  "context"
  "flag"
  "log/slog"
  "os"

	"github.com/edgelesssys/constellation/v2/internal/api/versionsapi"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/versions"
	"go.uber.org/zap/zapcore"
)

var (
  refFlag     = flag.String("ref", "", "the reference name of the image")
  streamFlag  = flag.String("stream", "", "the stream name of the image")
  versionFlag = flag.String("version", "", "the version of the image")
)

func main() {
  log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
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
		log.Fatalf("creating s3 client: %w", err)
	}
	defer func() {
		if err := cclose(ctx); err != nil {
			log.Fatalf("invalidating cache: %w", err)
		}
	}()

  if err := c.UpdateCLIInfo(ctx, cliInfo); err != nil {
    log.Error("updating cli info: %w", err)
    os.Exit(1)
  }
}
