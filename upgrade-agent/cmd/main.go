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
)

const (
	protocol = "unix"
)

func main() {
	verbosity := flag.Int("v", 0, logger.CmdLineVerbosityDescription)
	flag.Parse()
	log := logger.NewJSONLogger(logger.VerbosityFromInt(*verbosity)).WithGroup("upgrade-agent")
	logger.ReplaceGRPCLogger(logger.GRPCLogger(log))

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
