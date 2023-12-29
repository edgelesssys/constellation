/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package main

import (
	"flag"
	"log/slog"
	"os"

	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/upgrade-agent/internal/server"
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

	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logger.VerbosityFromInt(*verbosity)})).WithGroup("bootstrapper")

	if *gRPCDebug {
		logger.ReplaceGRPCLogger(log.WithGroup("gRPC"))
	} else {
    // TODO(miampf): Find a good way to change log level dynamically
		log.WithGroup("gRPC").WithIncreasedLevel(zap.WarnLevel).ReplaceGRPCLogger()
	}

	handler := file.NewHandler(afero.NewOsFs())
	server, err := server.New(log, handler)
	if err != nil {
		log.With(slog.Any("error", err)).Error("Failed to create update server")
    os.Exit(1)
	}

	err = server.Run(protocol, constants.UpgradeAgentSocketPath)
	if err != nil {
		log.With(slog.Any("error", err)).Error("Failed to start update server")
    os.Exit(1)
	}
}
