/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package main

import (
	"flag"
	"net"
	"strconv"

	"github.com/edgelesssys/constellation/v2/internal/attestation/aws"
	"github.com/edgelesssys/constellation/v2/internal/attestation/azure"
	"github.com/edgelesssys/constellation/v2/internal/attestation/gcp"
	"github.com/edgelesssys/constellation/v2/internal/attestation/qemu"
	"github.com/edgelesssys/constellation/v2/internal/attestation/tdx"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/verify/server"
	"go.uber.org/zap"
)

func main() {
	provider := flag.String("cloud-provider", "", "cloud service provider this binary is running on")
	verbosity := flag.Int("v", 0, logger.CmdLineVerbosityDescription)

	flag.Parse()
	log := logger.New(logger.JSONLog, logger.VerbosityFromInt(*verbosity))

	log.With(zap.String("version", constants.VersionInfo), zap.String("cloudProvider", *provider)).
		Infof("Constellation Verification Service")

	var issuer server.AttestationIssuer
	switch cloudprovider.FromString(*provider) {
	case cloudprovider.AWS:
		issuer = aws.NewIssuer()
	case cloudprovider.GCP:
		issuer = gcp.NewIssuer()
	case cloudprovider.Azure:
		issuer = azure.NewIssuer()
	case cloudprovider.QEMU:
		if tdx.Available() {
			issuer = tdx.NewIssuer(log)
		} else {
			issuer = qemu.NewIssuer()
		}
	default:
		log.With(zap.String("cloudProvider", *provider)).Fatalf("Unknown cloud provider")
	}

	server := server.New(log.Named("server"), issuer)
	httpListener, err := net.Listen("tcp", net.JoinHostPort("", strconv.Itoa(constants.VerifyServicePortHTTP)))
	if err != nil {
		log.With(zap.Error(err), zap.Int("port", constants.VerifyServicePortHTTP)).
			Fatalf("Failed to listen")
	}
	grpcListener, err := net.Listen("tcp", net.JoinHostPort("", strconv.Itoa(constants.VerifyServicePortGRPC)))
	if err != nil {
		log.With(zap.Error(err), zap.Int("port", constants.VerifyServicePortGRPC)).
			Fatalf("Failed to listen")
	}

	if err := server.Run(httpListener, grpcListener); err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to run server")
	}
}
