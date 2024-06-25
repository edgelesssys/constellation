package upgrade

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/versions/components"
	"github.com/edgelesssys/constellation/v2/upgrade-agent/upgradeproto"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

// TestGRPCDialer is a regression test to ensure the upgrade client can connect to a UDS.
func TestGRPCDialer(t *testing.T) {
	require := require.New(t)

	dir := t.TempDir()
	sockAddr := filepath.Join(dir, "test.socket")

	upgradeAgent := &fakeUpgradeAgent{}
	grpcServer := grpc.NewServer()
	upgradeproto.RegisterUpdateServer(grpcServer, upgradeAgent)

	listener, err := net.Listen("unix", sockAddr)
	require.NoError(err)
	go grpcServer.Serve(listener)
	t.Cleanup(grpcServer.Stop)

	fileInfo, err := os.Stat(sockAddr)
	require.NoError(err)
	require.Equal(os.ModeSocket, fileInfo.Mode()&os.ModeType)

	upgradeClient := newClientWithAddress(sockAddr)
	require.NoError(upgradeClient.Upgrade(context.Background(), []*components.Component{}, "v1.29.6"))
}

type fakeUpgradeAgent struct {
	upgradeproto.UnimplementedUpdateServer
}

func (s *fakeUpgradeAgent) ExecuteUpdate(_ context.Context, _ *upgradeproto.ExecuteUpdateRequest) (*upgradeproto.ExecuteUpdateResponse, error) {
	return &upgradeproto.ExecuteUpdateResponse{}, nil
}
