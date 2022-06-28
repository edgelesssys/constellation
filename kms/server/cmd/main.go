package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net"

	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/edgelesssys/constellation/internal/file"
	"github.com/edgelesssys/constellation/internal/logger"
	"github.com/edgelesssys/constellation/kms/server/kmsapi"
	"github.com/edgelesssys/constellation/kms/server/kmsapi/kmsproto"
	"github.com/edgelesssys/constellation/kms/server/setup"
	"github.com/spf13/afero"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"google.golang.org/grpc"
)

func main() {
	port := flag.String("port", "9000", "Port gRPC server listens on")
	masterSecretPath := flag.String("master-secret", "/constellation/constellation-mastersecret.base64", "Path to the Constellation master secret")

	flag.Parse()

	log := logger.New(logger.JSONLog, zapcore.InfoLevel)

	log.With(zap.String("version", constants.VersionInfo)).Infof("Constellation Key Management Service")

	masterKey, err := readMainSecret(*masterSecretPath)
	if err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to read master secret")
	}

	conKMS, err := setup.SetUpKMS(context.Background(), setup.NoStoreURI, setup.ClusterKMSURI)
	if err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to setup KMS")
	}

	if err := conKMS.CreateKEK(context.Background(), "Constellation", masterKey); err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to create KMS KEK from MasterKey")
	}

	lis, err := net.Listen("tcp", net.JoinHostPort("", *port))
	if err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to listen")
	}

	srv := kmsapi.New(log.Named("server"), conKMS)

	log.Named("gRPC").WithIncreasedLevel(zapcore.WarnLevel).ReplaceGRPCLogger()
	// TODO: Launch server with aTLS to allow attestation for clients.
	grpcServer := grpc.NewServer(log.Named("gRPC").GetServerUnaryInterceptor())

	kmsproto.RegisterAPIServer(grpcServer, srv)

	log.Infof("Starting key management service on %s", lis.Addr().String())
	if err := grpcServer.Serve(lis); err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to serve")
	}
}

// readMainSecret reads the base64 encoded main secret file from specified path and returns the secret as bytes.
func readMainSecret(fileName string) ([]byte, error) {
	if fileName == "" {
		return nil, errors.New("no filename to master secret provided")
	}

	fileHandler := file.NewHandler(afero.NewOsFs())

	secretBytes, err := fileHandler.Read(fileName)
	if err != nil {
		return nil, err
	}
	if len(secretBytes) < constants.MasterSecretLengthMin {
		return nil, fmt.Errorf("provided master secret is smaller than the required minimum of %d bytes", constants.MasterSecretLengthMin)
	}

	return secretBytes, nil
}
