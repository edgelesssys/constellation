/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package main

import (
	"flag"
	"log/slog"
	"net"
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
	log := logger.New(logger.JSONLog, logger.VerbosityFromInt(*verbosity))

	log.With(slog.String("version", constants.BinaryVersion().String()), slog.String("attestationVariant", *attestationVariant)).
		Infof("Constellation Verification Service")

	variant, err := variant.FromString(*attestationVariant)
	if err != nil {
		log.With(slog.Any("error", err)).Fatalf("Failed to parse attestation variant")
	}
	issuer, err := choose.Issuer(variant, log.Grouped("issuer"))
	if err != nil {
		log.With(slog.Any("error", err)).Fatalf("Failed to create issuer")
	}

	server := server.New(log.Grouped("server"), issuer)
	httpListener, err := net.Listen("tcp", net.JoinHostPort("", strconv.Itoa(constants.VerifyServicePortHTTP)))
	if err != nil {
		log.With(slog.Any("error", err), slog.Int("port", constants.VerifyServicePortHTTP)).
			Fatalf("Failed to listen")
	}
	grpcListener, err := net.Listen("tcp", net.JoinHostPort("", strconv.Itoa(constants.VerifyServicePortGRPC)))
	if err != nil {
		log.With(slog.Any("error", err), slog.Int("port", constants.VerifyServicePortGRPC)).
			Fatalf("Failed to listen")
	}

	if err := server.Run(httpListener, grpcListener); err != nil {
		log.With(slog.Any("error", err)).Fatalf("Failed to run server")
	}
}
