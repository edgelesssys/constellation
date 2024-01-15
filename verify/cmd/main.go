/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package main

import (
	"flag"
	"log/slog"
	"net"
	"os"
	"strconv"

	"github.com/edgelesssys/constellation/v2/internal/attestation/choose"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/verify/server"
)

func main() {
	attestationVariant := flag.String("attestation-variant", "", "attestation variant to use for aTLS connections")
	verbosity := flag.Int("v", 0, logger.CmdLineVerbosityDescription)

	flag.Parse()
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logger.VerbosityFromInt(*verbosity)}))

	log.With(slog.String("version", constants.BinaryVersion().String()), slog.String("attestationVariant", *attestationVariant)).
		Info("Constellation Verification Service")

	variant, err := variant.FromString(*attestationVariant)
	if err != nil {
		log.With(slog.Any("error", err)).Error("Failed to parse attestation variant")
		os.Exit(1)
	}
	issuer, err := choose.Issuer(variant, log.WithGroup("issuer"))
	if err != nil {
		log.With(slog.Any("error", err)).Error("Failed to create issuer")
		os.Exit(1)
	}

	server := server.New(log.WithGroup("server"), issuer)
	httpListener, err := net.Listen("tcp", net.JoinHostPort("", strconv.Itoa(constants.VerifyServicePortHTTP)))
	if err != nil {
		log.With(slog.Any("error", err), slog.Int("port", constants.VerifyServicePortHTTP)).
			Error("Failed to listen")
		os.Exit(1)
	}
	grpcListener, err := net.Listen("tcp", net.JoinHostPort("", strconv.Itoa(constants.VerifyServicePortGRPC)))
	if err != nil {
		log.With(slog.Any("error", err), slog.Int("port", constants.VerifyServicePortGRPC)).
			Error("Failed to listen")
		os.Exit(1)
	}

	if err := server.Run(httpListener, grpcListener); err != nil {
		log.With(slog.Any("error", err)).Error("Failed to run server")
		os.Exit(1)
	}
}
