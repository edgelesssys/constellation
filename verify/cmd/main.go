package main

import (
	"flag"
	"net"
	"strconv"

	"github.com/edgelesssys/constellation/internal/attestation/azure"
	"github.com/edgelesssys/constellation/internal/attestation/gcp"
	"github.com/edgelesssys/constellation/internal/attestation/qemu"
	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/edgelesssys/constellation/internal/logger"
	"github.com/edgelesssys/constellation/verify/server"
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
	switch *provider {
	case "gcp":
		issuer = gcp.NewIssuer()
	case "azure":
		issuer = azure.NewIssuer()
	case "qemu":
		issuer = qemu.NewIssuer()
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
