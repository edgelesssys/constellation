package deploy

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/edgelesssys/constellation/debugd/coordinator"
	"github.com/edgelesssys/constellation/debugd/debugd"
	pb "github.com/edgelesssys/constellation/debugd/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Download downloads a coordinator from a given debugd instance.
type Download struct {
	dialer             Dialer
	writer             streamToFileWriter
	serviceManager     serviceManager
	attemptedDownloads map[string]time.Time
}

// New creates a new Download.
func New(dialer Dialer, serviceManager serviceManager, writer streamToFileWriter) *Download {
	return &Download{
		dialer:             dialer,
		writer:             writer,
		serviceManager:     serviceManager,
		attemptedDownloads: map[string]time.Time{},
	}
}

// DownloadCoordinator will open a new grpc connection to another instance, attempting to download a coordinator from that instance.
func (d *Download) DownloadCoordinator(ctx context.Context, ip string) error {
	serverAddr := net.JoinHostPort(ip, debugd.DebugdPort)
	// only retry download from same endpoint after backoff
	if lastAttempt, ok := d.attemptedDownloads[serverAddr]; ok && time.Since(lastAttempt) < debugd.CoordinatorDownloadRetryBackoff {
		return fmt.Errorf("download failed too recently: %v / %v", time.Since(lastAttempt), debugd.CoordinatorDownloadRetryBackoff)
	}
	log.Printf("Trying to download coordinator from %s\n", ip)
	d.attemptedDownloads[serverAddr] = time.Now()
	conn, err := d.dial(ctx, serverAddr)
	if err != nil {
		return fmt.Errorf("error connecting to other instance via gRPC: %w", err)
	}
	defer conn.Close()
	client := pb.NewDebugdClient(conn)

	stream, err := client.DownloadCoordinator(ctx, &pb.DownloadCoordinatorRequest{})
	if err != nil {
		return fmt.Errorf("starting coordinator download from other instance failed: %w", err)
	}
	if err := d.writer.WriteStream(debugd.CoordinatorDeployFilename, stream, true); err != nil {
		return fmt.Errorf("streaming coordinator from other instance failed: %w", err)
	}

	log.Printf("Successfully downloaded coordinator from %s\n", ip)

	// after the upload succeeds, try to restart the coordinator
	restartAction := ServiceManagerRequest{
		Unit:   debugd.CoordinatorSystemdUnitName,
		Action: Restart,
	}
	if err := d.serviceManager.SystemdAction(ctx, restartAction); err != nil {
		return fmt.Errorf("restarting coordinator failed: %w", err)
	}

	return nil
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
	WriteStream(filename string, stream coordinator.ReadChunkStream, showProgress bool) error
}

type Dialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}
