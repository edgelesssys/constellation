package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"testing"

	"github.com/edgelesssys/constellation/debugd/coordinator"
	"github.com/edgelesssys/constellation/debugd/debugd/deploy"
	pb "github.com/edgelesssys/constellation/debugd/service"
	"github.com/edgelesssys/constellation/debugd/service/testdialer"
	"github.com/edgelesssys/constellation/debugd/ssh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestUploadAuthorizedKeys(t *testing.T) {
	endpoint := "192.0.2.1:4000"

	testCases := map[string]struct {
		ssh                    stubSSHDeployer
		serviceManager         stubServiceManager
		request                *pb.UploadAuthorizedKeysRequest
		expectErr              bool
		expectedResponseStatus pb.UploadAuthorizedKeysStatus
		expectedKeys           []ssh.SSHKey
	}{
		"upload authorized keys works": {
			request: &pb.UploadAuthorizedKeysRequest{
				Keys: []*pb.AuthorizedKey{
					{
						Username: "testuser",
						KeyValue: "teskey",
					},
				},
			},
			expectedResponseStatus: pb.UploadAuthorizedKeysStatus_UPLOAD_AUTHORIZED_KEYS_SUCCESS,
			expectedKeys: []ssh.SSHKey{
				{
					Username: "testuser",
					KeyValue: "teskey",
				},
			},
		},
		"deploy fails": {
			request: &pb.UploadAuthorizedKeysRequest{
				Keys: []*pb.AuthorizedKey{
					{
						Username: "testuser",
						KeyValue: "teskey",
					},
				},
			},
			ssh:                    stubSSHDeployer{deployErr: errors.New("ssh key deployment error")},
			expectedResponseStatus: pb.UploadAuthorizedKeysStatus_UPLOAD_AUTHORIZED_KEYS_FAILURE,
			expectedKeys: []ssh.SSHKey{
				{
					Username: "testuser",
					KeyValue: "teskey",
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			serv := debugdServer{
				ssh:            &tc.ssh,
				serviceManager: &tc.serviceManager,
				streamer:       &fakeStreamer{},
			}

			grpcServ, conn, err := setupServerWithConn(endpoint, &serv)
			require.NoError(err)
			defer conn.Close()
			client := pb.NewDebugdClient(conn)
			resp, err := client.UploadAuthorizedKeys(context.Background(), tc.request)

			grpcServ.GracefulStop()

			if tc.expectErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.expectedResponseStatus, resp.Status)
			assert.ElementsMatch(tc.ssh.sshKeys, tc.expectedKeys)
		})
	}
}

func TestUploadCoordinator(t *testing.T) {
	endpoint := "192.0.2.1:4000"

	testCases := map[string]struct {
		ssh                    stubSSHDeployer
		serviceManager         stubServiceManager
		streamer               fakeStreamer
		uploadChunks           [][]byte
		expectErr              bool
		expectedResponseStatus pb.UploadCoordinatorStatus
		expectFile             bool
		expectedChunks         [][]byte
	}{
		"upload works": {
			uploadChunks: [][]byte{
				[]byte("test"),
			},
			expectFile: true,
			expectedChunks: [][]byte{
				[]byte("test"),
			},
			expectedResponseStatus: pb.UploadCoordinatorStatus_UPLOAD_COORDINATOR_SUCCESS,
		},
		"recv fails": {
			streamer: fakeStreamer{
				writeStreamErr: errors.New("recv error"),
			},
			expectedResponseStatus: pb.UploadCoordinatorStatus_UPLOAD_COORDINATOR_UPLOAD_FAILED,
			expectFile:             true,
		},
		"starting coordinator fails": {
			uploadChunks: [][]byte{
				[]byte("test"),
			},
			serviceManager: stubServiceManager{
				systemdActionErr: errors.New("starting coordinator error"),
			},
			expectFile: true,
			expectedChunks: [][]byte{
				[]byte("test"),
			},
			expectedResponseStatus: pb.UploadCoordinatorStatus_UPLOAD_COORDINATOR_START_FAILED,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			serv := debugdServer{
				ssh:            &tc.ssh,
				serviceManager: &tc.serviceManager,
				streamer:       &tc.streamer,
			}

			grpcServ, conn, err := setupServerWithConn(endpoint, &serv)
			require.NoError(err)
			defer conn.Close()
			client := pb.NewDebugdClient(conn)
			stream, err := client.UploadCoordinator(context.Background())
			require.NoError(err)
			require.NoError(fakeWrite(stream, tc.uploadChunks))
			resp, err := stream.CloseAndRecv()

			grpcServ.GracefulStop()

			if tc.expectErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.expectedResponseStatus, resp.Status)
			if tc.expectFile {
				assert.Equal(tc.expectedChunks, tc.streamer.writeStreamChunks)
				assert.Equal("/opt/coordinator", tc.streamer.writeStreamFilename)
			} else {
				assert.Empty(tc.streamer.writeStreamChunks)
				assert.Empty(tc.streamer.writeStreamFilename)
			}
		})
	}
}

func TestDownloadCoordinator(t *testing.T) {
	endpoint := "192.0.2.1:4000"
	testCases := map[string]struct {
		ssh            stubSSHDeployer
		serviceManager stubServiceManager
		request        *pb.DownloadCoordinatorRequest
		streamer       fakeStreamer
		expectErr      bool
		expectedChunks [][]byte
	}{
		"download works": {
			request: &pb.DownloadCoordinatorRequest{},
			streamer: fakeStreamer{
				readStreamChunks: [][]byte{
					[]byte("test"),
				},
			},
			expectErr: false,
			expectedChunks: [][]byte{
				[]byte("test"),
			},
		},
		"download fails": {
			request: &pb.DownloadCoordinatorRequest{},
			streamer: fakeStreamer{
				readStreamErr: errors.New("read coordinator fails"),
			},
			expectErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			serv := debugdServer{
				ssh:            &tc.ssh,
				serviceManager: &tc.serviceManager,
				streamer:       &tc.streamer,
			}

			grpcServ, conn, err := setupServerWithConn(endpoint, &serv)
			require.NoError(err)
			defer conn.Close()
			client := pb.NewDebugdClient(conn)
			stream, err := client.DownloadCoordinator(context.Background(), tc.request)
			require.NoError(err)
			chunks, err := fakeRead(stream)
			grpcServ.GracefulStop()

			if tc.expectErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.expectedChunks, chunks)
			assert.Equal("/opt/coordinator", tc.streamer.readStreamFilename)
		})
	}
}

func TestUploadSystemServiceUnits(t *testing.T) {
	endpoint := "192.0.2.1:4000"
	testCases := map[string]struct {
		ssh                    stubSSHDeployer
		serviceManager         stubServiceManager
		request                *pb.UploadSystemdServiceUnitsRequest
		expectErr              bool
		expectedResponseStatus pb.UploadSystemdServiceUnitsStatus
		expectedUnitFiles      []deploy.SystemdUnit
	}{
		"upload systemd service units": {
			request: &pb.UploadSystemdServiceUnitsRequest{
				Units: []*pb.ServiceUnit{
					{
						Name:     "test.service",
						Contents: "testcontents",
					},
				},
			},
			expectedResponseStatus: pb.UploadSystemdServiceUnitsStatus_UPLOAD_SYSTEMD_SERVICE_UNITS_SUCCESS,
			expectedUnitFiles: []deploy.SystemdUnit{
				{
					Name:     "test.service",
					Contents: "testcontents",
				},
			},
		},
		"writing fails": {
			request: &pb.UploadSystemdServiceUnitsRequest{
				Units: []*pb.ServiceUnit{
					{
						Name:     "test.service",
						Contents: "testcontents",
					},
				},
			},
			serviceManager: stubServiceManager{
				writeSystemdUnitFileErr: errors.New("write error"),
			},
			expectedResponseStatus: pb.UploadSystemdServiceUnitsStatus_UPLOAD_SYSTEMD_SERVICE_UNITS_FAILURE,
			expectedUnitFiles: []deploy.SystemdUnit{
				{
					Name:     "test.service",
					Contents: "testcontents",
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			serv := debugdServer{
				ssh:            &tc.ssh,
				serviceManager: &tc.serviceManager,
				streamer:       &fakeStreamer{},
			}
			grpcServ, conn, err := setupServerWithConn(endpoint, &serv)
			require.NoError(err)
			defer conn.Close()
			client := pb.NewDebugdClient(conn)
			resp, err := client.UploadSystemServiceUnits(context.Background(), tc.request)

			grpcServ.GracefulStop()

			if tc.expectErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			require.NotNil(resp.Status)
			assert.Equal(tc.expectedResponseStatus, resp.Status)
			assert.ElementsMatch(tc.expectedUnitFiles, tc.serviceManager.unitFiles)
		})
	}
}

type stubSSHDeployer struct {
	sshKeys []ssh.SSHKey

	deployErr error
}

func (s *stubSSHDeployer) DeploySSHAuthorizedKey(ctx context.Context, sshKey ssh.SSHKey) error {
	s.sshKeys = append(s.sshKeys, sshKey)

	return s.deployErr
}

type stubServiceManager struct {
	requests                []deploy.ServiceManagerRequest
	unitFiles               []deploy.SystemdUnit
	systemdActionErr        error
	writeSystemdUnitFileErr error
}

func (s *stubServiceManager) SystemdAction(ctx context.Context, request deploy.ServiceManagerRequest) error {
	s.requests = append(s.requests, request)
	return s.systemdActionErr
}

func (s *stubServiceManager) WriteSystemdUnitFile(ctx context.Context, unit deploy.SystemdUnit) error {
	s.unitFiles = append(s.unitFiles, unit)
	return s.writeSystemdUnitFileErr
}

type dialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

func dial(ctx context.Context, dialer dialer, target string) (*grpc.ClientConn, error) {
	return grpc.DialContext(ctx, target,
		grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			return dialer.DialContext(ctx, "tcp", addr)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
}

type fakeStreamer struct {
	writeStreamChunks   [][]byte
	writeStreamFilename string
	writeStreamErr      error
	readStreamChunks    [][]byte
	readStreamFilename  string
	readStreamErr       error
}

func (f *fakeStreamer) WriteStream(filename string, stream coordinator.ReadChunkStream, showProgress bool) error {
	f.writeStreamFilename = filename
	for {
		chunk, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return f.writeStreamErr
			}
			return fmt.Errorf("reading stream failed: %w", err)
		}
		f.writeStreamChunks = append(f.writeStreamChunks, chunk.Content)
	}
}

func (f *fakeStreamer) ReadStream(filename string, stream coordinator.WriteChunkStream, chunksize uint, showProgress bool) error {
	f.readStreamFilename = filename
	for _, chunk := range f.readStreamChunks {
		if err := stream.Send(&pb.Chunk{Content: chunk}); err != nil {
			panic(err)
		}
	}
	return f.readStreamErr
}

func setupServerWithConn(endpoint string, serv *debugdServer) (*grpc.Server, *grpc.ClientConn, error) {
	dialer := testdialer.NewBufconnDialer()
	grpcServ := grpc.NewServer()
	pb.RegisterDebugdServer(grpcServ, serv)
	lis := dialer.GetListener(endpoint)
	go grpcServ.Serve(lis)

	conn, err := dial(context.Background(), dialer, endpoint)
	if err != nil {
		return nil, nil, err
	}

	return grpcServ, conn, nil
}

func fakeWrite(stream coordinator.WriteChunkStream, chunks [][]byte) error {
	for _, chunk := range chunks {
		err := stream.Send(&pb.Chunk{
			Content: chunk,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func fakeRead(stream coordinator.ReadChunkStream) ([][]byte, error) {
	var chunks [][]byte
	for {
		chunk, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				return chunks, nil
			}
			return nil, err
		}
		chunks = append(chunks, chunk.Content)
	}
}
