package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"testing"

	"github.com/edgelesssys/constellation/debugd/bootstrapper"
	"github.com/edgelesssys/constellation/debugd/debugd/deploy"
	pb "github.com/edgelesssys/constellation/debugd/service"
	"github.com/edgelesssys/constellation/internal/deploy/ssh"
	"github.com/edgelesssys/constellation/internal/grpc/testdialer"
	"github.com/edgelesssys/constellation/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestUploadAuthorizedKeys(t *testing.T) {
	endpoint := "192.0.2.1:4000"

	testCases := map[string]struct {
		ssh                stubSSHDeployer
		serviceManager     stubServiceManager
		request            *pb.UploadAuthorizedKeysRequest
		wantErr            bool
		wantResponseStatus pb.UploadAuthorizedKeysStatus
		wantKeys           []ssh.UserKey
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
			wantResponseStatus: pb.UploadAuthorizedKeysStatus_UPLOAD_AUTHORIZED_KEYS_SUCCESS,
			wantKeys: []ssh.UserKey{
				{
					Username:  "testuser",
					PublicKey: "teskey",
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
			ssh:                stubSSHDeployer{deployErr: errors.New("ssh key deployment error")},
			wantResponseStatus: pb.UploadAuthorizedKeysStatus_UPLOAD_AUTHORIZED_KEYS_FAILURE,
			wantKeys: []ssh.UserKey{
				{
					Username:  "testuser",
					PublicKey: "teskey",
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			serv := debugdServer{
				log:            logger.NewTest(t),
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

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantResponseStatus, resp.Status)
			assert.ElementsMatch(tc.ssh.sshKeys, tc.wantKeys)
		})
	}
}

func TestUploadBootstrapper(t *testing.T) {
	endpoint := "192.0.2.1:4000"

	testCases := map[string]struct {
		ssh                stubSSHDeployer
		serviceManager     stubServiceManager
		streamer           fakeStreamer
		uploadChunks       [][]byte
		wantErr            bool
		wantResponseStatus pb.UploadBootstrapperStatus
		wantFile           bool
		wantChunks         [][]byte
	}{
		"upload works": {
			uploadChunks: [][]byte{
				[]byte("test"),
			},
			wantFile: true,
			wantChunks: [][]byte{
				[]byte("test"),
			},
			wantResponseStatus: pb.UploadBootstrapperStatus_UPLOAD_BOOTSTRAPPER_SUCCESS,
		},
		"recv fails": {
			streamer: fakeStreamer{
				writeStreamErr: errors.New("recv error"),
			},
			wantResponseStatus: pb.UploadBootstrapperStatus_UPLOAD_BOOTSTRAPPER_UPLOAD_FAILED,
			wantErr:            true,
		},
		"starting bootstrapper fails": {
			uploadChunks: [][]byte{
				[]byte("test"),
			},
			serviceManager: stubServiceManager{
				systemdActionErr: errors.New("starting bootstrapper error"),
			},
			wantFile: true,
			wantChunks: [][]byte{
				[]byte("test"),
			},
			wantResponseStatus: pb.UploadBootstrapperStatus_UPLOAD_BOOTSTRAPPER_START_FAILED,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			serv := debugdServer{
				log:            logger.NewTest(t),
				ssh:            &tc.ssh,
				serviceManager: &tc.serviceManager,
				streamer:       &tc.streamer,
			}

			grpcServ, conn, err := setupServerWithConn(endpoint, &serv)
			require.NoError(err)
			defer conn.Close()
			client := pb.NewDebugdClient(conn)
			stream, err := client.UploadBootstrapper(context.Background())
			require.NoError(err)
			require.NoError(fakeWrite(stream, tc.uploadChunks))
			resp, err := stream.CloseAndRecv()

			grpcServ.GracefulStop()

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantResponseStatus, resp.Status)
			if tc.wantFile {
				assert.Equal(tc.wantChunks, tc.streamer.writeStreamChunks)
				assert.Equal("/opt/bootstrapper", tc.streamer.writeStreamFilename)
			} else {
				assert.Empty(tc.streamer.writeStreamChunks)
				assert.Empty(tc.streamer.writeStreamFilename)
			}
		})
	}
}

func TestDownloadBootstrapper(t *testing.T) {
	endpoint := "192.0.2.1:4000"
	testCases := map[string]struct {
		ssh            stubSSHDeployer
		serviceManager stubServiceManager
		request        *pb.DownloadBootstrapperRequest
		streamer       fakeStreamer
		wantErr        bool
		wantChunks     [][]byte
	}{
		"download works": {
			request: &pb.DownloadBootstrapperRequest{},
			streamer: fakeStreamer{
				readStreamChunks: [][]byte{
					[]byte("test"),
				},
			},
			wantErr: false,
			wantChunks: [][]byte{
				[]byte("test"),
			},
		},
		"download fails": {
			request: &pb.DownloadBootstrapperRequest{},
			streamer: fakeStreamer{
				readStreamErr: errors.New("read bootstrapper fails"),
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			serv := debugdServer{
				log:            logger.NewTest(t),
				ssh:            &tc.ssh,
				serviceManager: &tc.serviceManager,
				streamer:       &tc.streamer,
			}

			grpcServ, conn, err := setupServerWithConn(endpoint, &serv)
			require.NoError(err)
			defer conn.Close()
			client := pb.NewDebugdClient(conn)
			stream, err := client.DownloadBootstrapper(context.Background(), tc.request)
			require.NoError(err)
			chunks, err := fakeRead(stream)
			grpcServ.GracefulStop()

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantChunks, chunks)
			assert.Equal("/opt/bootstrapper", tc.streamer.readStreamFilename)
		})
	}
}

func TestUploadSystemServiceUnits(t *testing.T) {
	endpoint := "192.0.2.1:4000"
	testCases := map[string]struct {
		ssh                stubSSHDeployer
		serviceManager     stubServiceManager
		request            *pb.UploadSystemdServiceUnitsRequest
		wantErr            bool
		wantResponseStatus pb.UploadSystemdServiceUnitsStatus
		wantUnitFiles      []deploy.SystemdUnit
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
			wantResponseStatus: pb.UploadSystemdServiceUnitsStatus_UPLOAD_SYSTEMD_SERVICE_UNITS_SUCCESS,
			wantUnitFiles: []deploy.SystemdUnit{
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
			wantResponseStatus: pb.UploadSystemdServiceUnitsStatus_UPLOAD_SYSTEMD_SERVICE_UNITS_FAILURE,
			wantUnitFiles: []deploy.SystemdUnit{
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
				log:            logger.NewTest(t),
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

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			require.NotNil(resp.Status)
			assert.Equal(tc.wantResponseStatus, resp.Status)
			assert.ElementsMatch(tc.wantUnitFiles, tc.serviceManager.unitFiles)
		})
	}
}

type stubSSHDeployer struct {
	sshKeys []ssh.UserKey

	deployErr error
}

func (s *stubSSHDeployer) DeployAuthorizedKey(ctx context.Context, sshKey ssh.UserKey) error {
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

type netDialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

func dial(ctx context.Context, dialer netDialer, target string) (*grpc.ClientConn, error) {
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

func (f *fakeStreamer) WriteStream(filename string, stream bootstrapper.ReadChunkStream, showProgress bool) error {
	f.writeStreamFilename = filename
	for {
		chunk, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return f.writeStreamErr
			}
			return fmt.Errorf("reading stream: %w", err)
		}
		f.writeStreamChunks = append(f.writeStreamChunks, chunk.Content)
	}
}

func (f *fakeStreamer) ReadStream(filename string, stream bootstrapper.WriteChunkStream, chunksize uint, showProgress bool) error {
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

func fakeWrite(stream bootstrapper.WriteChunkStream, chunks [][]byte) error {
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

func fakeRead(stream bootstrapper.ReadChunkStream) ([][]byte, error) {
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
