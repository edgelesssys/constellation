package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/edgelesssys/constellation/cli/file"
	"github.com/edgelesssys/constellation/coordinator/atls"
	"github.com/edgelesssys/constellation/coordinator/attestation/vtpm"
	"github.com/edgelesssys/constellation/coordinator/core"
	"github.com/edgelesssys/constellation/coordinator/pubapi"
	"github.com/edgelesssys/constellation/coordinator/pubapi/pubproto"
	"github.com/edgelesssys/constellation/coordinator/store"
	"github.com/edgelesssys/constellation/coordinator/vpnapi"
	"github.com/edgelesssys/constellation/coordinator/vpnapi/vpnproto"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var version = "0.0.0"

func run(validator core.QuoteValidator, issuer core.QuoteIssuer, vpn core.VPN, openTPM vtpm.TPMOpenFunc, getPublicIPAddr func() (string, error), dialer pubapi.Dialer, fileHandler file.Handler,
	kube core.Cluster, metadata core.ProviderMetadata, cloudControllerManager core.CloudControllerManager, cloudNodeManager core.CloudNodeManager, clusterAutoscaler core.ClusterAutoscaler, etcdEndpoint string, etcdTLS bool, bindIP, bindPort string, zapLoggerCore *zap.Logger,
) {
	defer zapLoggerCore.Sync()
	zapLoggerCore.Info("starting coordinator", zap.String("version", version))

	tlsConfig, err := atls.CreateAttestationServerTLSConfig(issuer)
	if err != nil {
		zapLoggerCore.Fatal("failed to create server TLS config", zap.Error(err))
	}

	etcdStoreFactory := &store.EtcdStoreFactory{
		Endpoint: etcdEndpoint,
		ForceTLS: etcdTLS,
		Logger:   zapLoggerCore.WithOptions(zap.IncreaseLevel(zap.WarnLevel)).Named("etcd"),
	}
	core, err := core.NewCore(vpn, kube, metadata, cloudControllerManager, cloudNodeManager, clusterAutoscaler, zapLoggerCore, openTPM, etcdStoreFactory, fileHandler)
	if err != nil {
		zapLoggerCore.Fatal("failed to create core", zap.Error(err))
	}
	// initialize state machine and wait for re-joining of the VPN (if applicable)
	nodeActivated, err := core.Initialize()
	if err != nil {
		zapLoggerCore.Fatal("failed to initialize core", zap.Error(err))
	}

	vapiServer := &vpnAPIServer{logger: zapLoggerCore.Named("vpnapi"), core: core}
	zapLoggerPubapi := zapLoggerCore.Named("pubapi")
	papi := pubapi.New(zapLoggerPubapi, core, dialer, vapiServer, validator, getPublicIPAddr)

	zapLoggergRPC := zapLoggerPubapi.Named("gRPC")

	grpcServer := grpc.NewServer(
		grpc.Creds(credentials.NewTLS(tlsConfig)),
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			grpc_ctxtags.StreamServerInterceptor(),
			grpc_zap.StreamServerInterceptor(zapLoggergRPC),
		)),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_zap.UnaryServerInterceptor(zapLoggergRPC),
		)),
	)
	pubproto.RegisterAPIServer(grpcServer, papi)

	lis, err := net.Listen("tcp", net.JoinHostPort(bindIP, bindPort))
	if err != nil {
		zapLoggergRPC.Fatal("failed to create listener", zap.Error(err))
	}
	zapLoggergRPC.Info("server listener created", zap.String("address", lis.Addr().String()))

	var wg sync.WaitGroup
	defer wg.Wait()
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := grpcServer.Serve(lis); err != nil {
			zapLoggergRPC.Fatal("failed to serve gRPC", zap.Error(err))
		}
	}()

	if !nodeActivated {
		zapLoggerStartupJoin := zapLoggerCore.Named("startup-join")
		if err := tryJoinClusterOnStartup(getPublicIPAddr, metadata, bindPort, zapLoggerStartupJoin); err != nil {
			zapLoggerStartupJoin.Info("joining existing cluster on startup failed. Waiting for connection.", zap.Error(err))
		}
	}
}

func tryJoinClusterOnStartup(getPublicIPAddr func() (string, error), metadata core.ProviderMetadata, bindPort string, logger *zap.Logger) error {
	nodePublicIP, err := getPublicIPAddr()
	if err != nil {
		return fmt.Errorf("failed to retrieve own public ip: %w", err)
	}
	nodeEndpoint := net.JoinHostPort(nodePublicIP, bindPort)
	if !metadata.Supported() {
		logger.Info("Metadata API not implemented for cloud provider")
		return errors.New("metadata API not implemented")
	}
	coordinatorEndpoints, err := core.CoordinatorEndpoints(context.TODO(), metadata)
	if err != nil {
		return fmt.Errorf("failed to retrieve coordinatorEndpoints from cloud provider api: %w", err)
	}
	logger.Info("Retrieved endpoints from cloud-provider API", zap.Strings("endpoints", coordinatorEndpoints))

	// We create an client unverified connection, since the node does not need to verify the Coordinator.
	// ActivateAdditionalNodes triggers the Coordinator to call ActivateAsNode. This rpc lets the Coordinator verify the node.
	tlsClientConfig, err := atls.CreateUnverifiedClientTLSConfig()
	if err != nil {
		return fmt.Errorf("failed to create client TLS config: %w", err)
	}

	// try to notify a coordinator to activate this node
	for _, coordinatorEndpoint := range coordinatorEndpoints {
		conn, err := grpc.Dial(coordinatorEndpoint, grpc.WithTransportCredentials(credentials.NewTLS(tlsClientConfig)))
		if err != nil {
			logger.Info("Dial failed:", zap.String("endpoint", coordinatorEndpoint), zap.Error(err))
			continue
		}
		defer conn.Close()
		client := pubproto.NewAPIClient(conn)
		logger.Info("Activating as node on startup")
		_, err = client.ActivateAdditionalNodes(context.Background(), &pubproto.ActivateAdditionalNodesRequest{NodePublicEndpoints: []string{nodeEndpoint}})
		return err
	}

	return errors.New("could not connect to any coordinator endpoint")
}

type vpnAPIServer struct {
	logger   *zap.Logger
	core     vpnapi.Core
	listener net.Listener
	server   *grpc.Server
}

func (v *vpnAPIServer) Listen(endpoint string) error {
	api := vpnapi.New(v.logger, v.core)
	grpcLogger := v.logger.Named("gRPC")
	v.server = grpc.NewServer(
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			grpc_ctxtags.StreamServerInterceptor(),
			grpc_zap.StreamServerInterceptor(grpcLogger),
		)),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_zap.UnaryServerInterceptor(grpcLogger),
		)),
	)
	vpnproto.RegisterAPIServer(v.server, api)

	lis, err := net.Listen("tcp", endpoint)
	if err != nil {
		return err
	}
	v.listener = lis
	return nil
}

func (v *vpnAPIServer) Serve() error {
	return v.server.Serve(v.listener)
}

func (v *vpnAPIServer) Close() {
	if v.server != nil {
		v.server.GracefulStop()
	}
}
