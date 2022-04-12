package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/edgelesssys/constellation/internal/file"
	"github.com/edgelesssys/constellation/kms/server/kmsapi"
	"github.com/edgelesssys/constellation/kms/server/kmsapi/kmsproto"
	"github.com/edgelesssys/constellation/kms/server/setup"
	"github.com/spf13/afero"
	"go.uber.org/zap"

	"google.golang.org/grpc"
)

func main() {
	port := flag.String("p", "9000", "Port gRPC server listens on")
	masterSecretPath := flag.String("master-secret", "/constellation/constellation-mastersecret.base64", "Path to the Constellation master secret")
	flag.Parse()

	masterKey, err := readMainSecret(*masterSecretPath)
	if err != nil {
		log.Fatalf("Failed to read master secret: %v", err)
	}

	conKMS, err := setup.SetUpKMS(context.Background(), setup.NoStoreURI, setup.ClusterKMSURI)
	if err != nil {
		log.Fatalf("Failed to setup KMS: %v", err)
	}

	if err := conKMS.CreateKEK(context.Background(), "Constellation", masterKey); err != nil {
		log.Fatalf("Failed to create KMS KEK from MasterKey: %v", err)
	}

	lis, err := net.Listen("tcp", net.JoinHostPort("0.0.0.0", *port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	srv := kmsapi.New(&zap.Logger{}, conKMS)

	// TODO: Launch server with aTLS to allow attestation for clients.
	grpcServer := grpc.NewServer()

	kmsproto.RegisterAPIServer(grpcServer, srv)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %s", err)
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
