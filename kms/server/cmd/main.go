package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net"

	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/edgelesssys/constellation/internal/file"
	"github.com/edgelesssys/constellation/internal/grpc/grpc_klog"
	"github.com/edgelesssys/constellation/kms/server/kmsapi"
	"github.com/edgelesssys/constellation/kms/server/kmsapi/kmsproto"
	"github.com/edgelesssys/constellation/kms/server/setup"
	"github.com/spf13/afero"
	"go.uber.org/zap"
	"k8s.io/klog/v2"

	"google.golang.org/grpc"
)

func main() {
	port := flag.String("port", "9000", "Port gRPC server listens on")
	masterSecretPath := flag.String("master-secret", "/constellation/constellation-mastersecret.base64", "Path to the Constellation master secret")

	klog.InitFlags(nil)
	flag.Parse()
	defer klog.Flush()

	klog.V(2).Infof("\nConstellation Key Management Service\nVersion: %s", constants.VersionInfo)

	masterKey, err := readMainSecret(*masterSecretPath)
	if err != nil {
		klog.Exitf("Failed to read master secret: %v", err)
	}

	conKMS, err := setup.SetUpKMS(context.Background(), setup.NoStoreURI, setup.ClusterKMSURI)
	if err != nil {
		klog.Exitf("Failed to setup KMS: %v", err)
	}

	if err := conKMS.CreateKEK(context.Background(), "Constellation", masterKey); err != nil {
		klog.Exitf("Failed to create KMS KEK from MasterKey: %v", err)
	}

	lis, err := net.Listen("tcp", net.JoinHostPort("", *port))
	if err != nil {
		klog.Exitf("Failed to listen: %v", err)
	}

	srv := kmsapi.New(&zap.Logger{}, conKMS)

	// TODO: Launch server with aTLS to allow attestation for clients.
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_klog.LogGRPC(2)),
	)

	kmsproto.RegisterAPIServer(grpcServer, srv)

	klog.V(2).Infof("Starting key management service on %s", lis.Addr().String())
	if err := grpcServer.Serve(lis); err != nil {
		klog.Exitf("Failed to serve: %s", err)
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
