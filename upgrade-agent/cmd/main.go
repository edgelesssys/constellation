/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package main

import (
	"flag"

	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	upgradeagent "github.com/edgelesssys/constellation/v2/upgrade-agent"
	"github.com/spf13/afero"
	"go.uber.org/zap"
)

const (
	protocol = "unix"
)

func main() {
	gRPCDebug := flag.Bool("debug", false, "Enable gRPC debug logging")
	verbosity := flag.Int("v", 0, logger.CmdLineVerbosityDescription)
	flag.Parse()

	log := logger.New(logger.JSONLog, logger.VerbosityFromInt(*verbosity)).Named("bootstrapper")
	defer log.Sync()

	if *gRPCDebug {
		log.Named("gRPC").ReplaceGRPCLogger()
	} else {
		log.Named("gRPC").WithIncreasedLevel(zap.WarnLevel).ReplaceGRPCLogger()
	}

	handler := file.NewHandler(afero.NewOsFs())
	server, err := upgradeagent.New(log, handler)
	if err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to create update server")
	}

	err = server.Run(protocol, constants.UpgradeAgentSocketPath)
	if err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to start update server")
	}
}
