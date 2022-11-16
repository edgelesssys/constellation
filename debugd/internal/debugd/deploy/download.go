/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package deploy

import (
	"context"
	"fmt"
	"io"
	"net"
	"strconv"

	"github.com/edgelesssys/constellation/v2/debugd/internal/bootstrapper"
	"github.com/edgelesssys/constellation/v2/debugd/internal/debugd"
	pb "github.com/edgelesssys/constellation/v2/debugd/service"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Download downloads a bootstrapper from a given debugd instance.
type Download struct {
	log            *logger.Logger
	dialer         NetDialer
	writer         streamToFileWriter
	serviceManager serviceManager
	info           infoSetter
}

// New creates a new Download.
func New(log *logger.Logger, dialer NetDialer, serviceManager serviceManager,
	writer streamToFileWriter, info infoSetter,
) *Download {
	return &Download{
		log:            log,
		dialer:         dialer,
		writer:         writer,
		info:           info,
		serviceManager: serviceManager,
	}
}

// DownloadInfo will try to download the info from another instance.
func (d *Download) DownloadInfo(ctx context.Context, ip string) error {
	log := d.log.With(zap.String("ip", ip))
	serverAddr := net.JoinHostPort(ip, strconv.Itoa(constants.DebugdPort))

	client, closer, err := d.newClient(ctx, serverAddr, log)
	if err != nil {
		return err
	}
	defer closer.Close()

	log.Infof("Trying to download info")
	resp, err := client.GetInfo(ctx, &pb.GetInfoRequest{})
	if err != nil {
		return fmt.Errorf("getting info from other instance: %w", err)
	}
	log.Infof("Successfully downloaded info")

	return d.info.SetProto(resp.Info)
}

// DownloadDeployment will open a new grpc connection to another instance, attempting to download a bootstrapper from that instance.
func (d *Download) DownloadDeployment(ctx context.Context, ip string) error {
	log := d.log.With(zap.String("ip", ip))
	serverAddr := net.JoinHostPort(ip, strconv.Itoa(constants.DebugdPort))

	client, closer, err := d.newClient(ctx, serverAddr, log)
	if err != nil {
		return err
	}
	defer closer.Close()

	log.Infof("Trying to download bootstrapper")
	stream, err := client.DownloadBootstrapper(ctx, &pb.DownloadBootstrapperRequest{})
	if err != nil {
		return fmt.Errorf("starting bootstrapper download from other instance: %w", err)
	}
	if err := d.writer.WriteStream(debugd.BootstrapperDeployFilename, stream, true); err != nil {
		return fmt.Errorf("streaming bootstrapper from other instance: %w", err)
	}
	log.Infof("Successfully downloaded bootstrapper")

	// after the upload succeeds, try to restart the bootstrapper
	restartAction := ServiceManagerRequest{
		Unit:   debugd.BootstrapperSystemdUnitName,
		Action: Restart,
	}
	if err := d.serviceManager.SystemdAction(ctx, restartAction); err != nil {
		return fmt.Errorf("restarting bootstrapper: %w", err)
	}

	return nil
}

func (d *Download) newClient(ctx context.Context, serverAddr string, log *logger.Logger) (pb.DebugdClient, io.Closer, error) {
	log.Infof("Connecting to server")
	conn, err := d.dial(ctx, serverAddr)
	if err != nil {
		return nil, nil, fmt.Errorf("connecting to other instance via gRPC: %w", err)
	}
	return pb.NewDebugdClient(conn), conn, nil
}

func (d *Download) dial(ctx context.Context, target string) (*grpc.ClientConn, error) {
	return grpc.DialContext(ctx, target,
		d.grpcWithDialer(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
}

func (d *Download) grpcWithDialer() grpc.DialOption {
	return grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
		return d.dialer.DialContext(ctx, "tcp", addr)
	})
}

type infoSetter interface {
	SetProto(infos []*pb.Info) error
}

type serviceManager interface {
	SystemdAction(ctx context.Context, request ServiceManagerRequest) error
}

type streamToFileWriter interface {
	WriteStream(filename string, stream bootstrapper.ReadChunkStream, showProgress bool) error
}

// NetDialer can open a net.Conn.
type NetDialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}
