/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package deploy

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"strconv"

	"github.com/edgelesssys/constellation/v2/debugd/internal/filetransfer"
	pb "github.com/edgelesssys/constellation/v2/debugd/service"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Download downloads a bootstrapper from a given debugd instance.
type Download struct {
	log            *slog.Logger
	dialer         NetDialer
	transfer       fileTransferer
	serviceManager serviceManager
	info           infoSetter
}

// New creates a new Download.
func New(log *slog.Logger, dialer NetDialer, serviceManager serviceManager,
	transfer fileTransferer, info infoSetter,
) *Download {
	return &Download{
		log:            log,
		dialer:         dialer,
		transfer:       transfer,
		info:           info,
		serviceManager: serviceManager,
	}
}

// DownloadInfo will try to download the info from another instance.
func (d *Download) DownloadInfo(ctx context.Context, ip string) error {
	if d.info.Received() {
		return nil
	}

	log := d.log.With(slog.String("ip", ip))
	serverAddr := net.JoinHostPort(ip, strconv.Itoa(constants.DebugdPort))

	client, closer, err := d.newClient(ctx, serverAddr, log)
	if err != nil {
		return err
	}
	defer closer.Close()

	log.Info("Trying to download info")
	resp, err := client.GetInfo(ctx, &pb.GetInfoRequest{})
	if err != nil {
		return fmt.Errorf("getting info from other instance: %w", err)
	}
	log.Info("Successfully downloaded info")

	return d.info.SetProto(resp.Info)
}

// DownloadDeployment will open a new grpc connection to another instance, attempting to download files from that instance.
func (d *Download) DownloadDeployment(ctx context.Context, ip string) error {
	log := d.log.With(slog.String("ip", ip))
	serverAddr := net.JoinHostPort(ip, strconv.Itoa(constants.DebugdPort))

	client, closer, err := d.newClient(ctx, serverAddr, log)
	if err != nil {
		return err
	}
	defer closer.Close()

	log.Info("Trying to download files")
	stream, err := client.DownloadFiles(ctx, &pb.DownloadFilesRequest{})
	if err != nil {
		return fmt.Errorf("starting file download from other instance: %w", err)
	}

	err = d.transfer.RecvFiles(stream)
	switch {
	case err == nil:
		d.log.Info("Downloading files succeeded")
	case errors.Is(err, filetransfer.ErrReceiveRunning):
		d.log.Warn("Download already in progress")
		return err
	case errors.Is(err, filetransfer.ErrReceiveFinished):
		d.log.Warn("Download already finished")
		return nil
	default:
		d.log.With(slog.Any("error", err)).Error("Downloading files failed")
		return err
	}

	files := d.transfer.GetFiles()
	for _, file := range files {
		if file.OverrideServiceUnit == "" {
			continue
		}
		if err := d.serviceManager.OverrideServiceUnitExecStart(
			ctx, file.OverrideServiceUnit, file.TargetPath,
		); err != nil {
			// continue on error to allow other units to be overridden
			d.log.With(slog.Any("error", err)).Error(fmt.Sprintf("Failed to override service unit %s", file.OverrideServiceUnit))
		}
	}

	return nil
}

func (d *Download) newClient(ctx context.Context, serverAddr string, log *slog.Logger) (pb.DebugdClient, io.Closer, error) {
	log.Info("Connecting to server")
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
	Received() bool
}

type serviceManager interface {
	SystemdAction(ctx context.Context, request ServiceManagerRequest) error
	OverrideServiceUnitExecStart(ctx context.Context, unitName string, execStart string) error
}

type fileTransferer interface {
	RecvFiles(stream filetransfer.RecvFilesStream) error
	GetFiles() []filetransfer.FileStat
}

// NetDialer can open a net.Conn.
type NetDialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}
