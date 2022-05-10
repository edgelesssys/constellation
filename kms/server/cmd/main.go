package main

import (
	"context"
	"flag"
	"log"
	"net"

	"github.com/edgelesssys/constellation/coordinator/util"
	"github.com/edgelesssys/constellation/kms/server/kmsapi"
	"github.com/edgelesssys/constellation/kms/server/kmsapi/kmsproto"
	"github.com/edgelesssys/constellation/kms/server/setup"
	"go.uber.org/zap"

	"google.golang.org/grpc"
)

func main() {
	port := flag.String("p", "9000", "Port gRPC server listens on")
	flag.Parse()

	// TODO: Get masterSecret from Constellation CLI / after activation from cluster.
	masterKey, err := util.GenerateRandomBytes(32)
	if err != nil {
		log.Fatalf("Failed to generate key: %v", err)
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
