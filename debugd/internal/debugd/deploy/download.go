/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package deploy

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/edgelesssys/constellation/v2/debugd/internal/bootstrapper"
	"github.com/edgelesssys/constellation/v2/debugd/internal/debugd"
	pb "github.com/edgelesssys/constellation/v2/debugd/service"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/deploy/ssh"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Download downloads a bootstrapper from a given debugd instance.
type Download struct {
	log                *logger.Logger
	dialer             NetDialer
	writer             streamToFileWriter
	serviceManager     serviceManager
	attemptedDownloads map[string]time.Time
}

// New creates a new Download.
func New(log *logger.Logger, dialer NetDialer, serviceManager serviceManager, writer streamToFileWriter) *Download {
	return &Download{
		log:                log,
		dialer:             dialer,
		writer:             writer,
		serviceManager:     serviceManager,
		attemptedDownloads: map[string]time.Time{},
	}
}

// DownloadDeployment will open a new grpc connection to another instance, attempting to download a bootstrapper from that instance.
func (d *Download) DownloadDeployment(ctx context.Context, ip string) ([]ssh.UserKey, error) {
	log := d.log.With(zap.String("ip", ip))
	serverAddr := net.JoinHostPort(ip, strconv.Itoa(constants.DebugdPort))

	// only retry download from same endpoint after backoff
	if lastAttempt, ok := d.attemptedDownloads[serverAddr]; ok && time.Since(lastAttempt) < debugd.BootstrapperDownloadRetryBackoff {
		return nil, fmt.Errorf("download failed too recently: %v / %v", time.Since(lastAttempt), debugd.BootstrapperDownloadRetryBackoff)
	}

	log.Infof("Connecting to server")
	d.attemptedDownloads[serverAddr] = time.Now()
	conn, err := d.dial(ctx, serverAddr)
	if err != nil {
		return nil, fmt.Errorf("connecting to other instance via gRPC: %w", err)
	}
	defer conn.Close()
	client := pb.NewDebugdClient(conn)

	log.Infof("Trying to download bootstrapper")
	stream, err := client.DownloadBootstrapper(ctx, &pb.DownloadBootstrapperRequest{})
	if err != nil {
		return nil, fmt.Errorf("starting bootstrapper download from other instance: %w", err)
	}
	if err := d.writer.WriteStream(debugd.BootstrapperDeployFilename, stream, true); err != nil {
		return nil, fmt.Errorf("streaming bootstrapper from other instance: %w", err)
	}
	log.Infof("Successfully downloaded bootstrapper")

	log.Infof("Trying to download ssh keys")
	resp, err := client.DownloadAuthorizedKeys(ctx, &pb.DownloadAuthorizedKeysRequest{})
	if err != nil {
		return nil, fmt.Errorf("downloading authorized keys: %w", err)
	}

	var keys []ssh.UserKey
	for _, key := range resp.Keys {
		keys = append(keys, ssh.UserKey{Username: key.Username, PublicKey: key.KeyValue})
	}

	// after the upload succeeds, try to restart the bootstrapper
	restartAction := ServiceManagerRequest{
		Unit:   debugd.BootstrapperSystemdUnitName,
		Action: Restart,
	}
	if err := d.serviceManager.SystemdAction(ctx, restartAction); err != nil {
		return nil, fmt.Errorf("restarting bootstrapper: %w", err)
	}

	return keys, nil
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
