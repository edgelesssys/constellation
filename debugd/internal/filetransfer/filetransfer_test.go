/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package filetransfer

import (
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/edgelesssys/constellation/v2/debugd/internal/filetransfer/streamer"
	pb "github.com/edgelesssys/constellation/v2/debugd/service"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m, goleak.IgnoreAnyFunction("github.com/bazelbuild/rules_go/go/tools/bzltestutil.RegisterTimeoutHandler.func1"))
}

func TestSendFiles(t *testing.T) {
	testCases := map[string]struct {
		files           *[]FileStat
		receiveFinished bool
		sendErr         error
		readStreamErr   error
		wantHeaders     []*pb.FileTransferMessage
		wantErr         bool
	}{
		"can send files": {
			files: &[]FileStat{
				{
					TargetPath:          "testfileA",
					Mode:                0o644,
					OverrideServiceUnit: "somesvcA",
				},
				{
					TargetPath:          "testfileB",
					Mode:                0o644,
					OverrideServiceUnit: "somesvcB",
				},
			},
			receiveFinished: true,
			wantHeaders: []*pb.FileTransferMessage{
				{
					Kind: &pb.FileTransferMessage_Header{
						Header: &pb.FileTransferHeader{
							TargetPath:          "testfileA",
							Mode:                0o644,
							OverrideServiceUnit: func() *string { s := "somesvcA"; return &s }(),
						},
					},
				},
				{
					Kind: &pb.FileTransferMessage_Header{
						Header: &pb.FileTransferHeader{
							TargetPath:          "testfileB",
							Mode:                0o644,
							OverrideServiceUnit: func() *string { s := "somesvcB"; return &s }(),
						},
					},
				},
			},
		},
		"not finished receiving": {
			files: &[]FileStat{
				{
					TargetPath:          "testfileA",
					Mode:                0o644,
					OverrideServiceUnit: "somesvcA",
				},
				{
					TargetPath:          "testfileB",
					Mode:                0o644,
					OverrideServiceUnit: "somesvcB",
				},
			},
			receiveFinished: false,
			wantErr:         true,
		},
		"send fails": {
			files: &[]FileStat{
				{
					TargetPath:          "testfileA",
					Mode:                0o644,
					OverrideServiceUnit: "somesvcA",
				},
			},
			receiveFinished: true,
			sendErr:         errors.New("send failed"),
			wantErr:         true,
		},
		"read stream fails": {
			files: &[]FileStat{
				{
					TargetPath:          "testfileA",
					Mode:                0o644,
					OverrideServiceUnit: "somesvcA",
				},
			},
			receiveFinished: true,
			readStreamErr:   errors.New("read stream failed"),
			wantErr:         true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			streamer := &stubStreamReadWriter{readStreamErr: tc.readStreamErr}
			stream := &stubSendFilesStream{sendErr: tc.sendErr}
			transfer := &FileTransferer{
				log:          slog.New(slog.NewTextHandler(logger.TestWriter{T: t}, nil)),
				streamer:     streamer,
				showProgress: false,
			}
			if tc.files != nil {
				transfer.files = *tc.files
			}
			transfer.receiveFinished.Store(tc.receiveFinished)

			err := transfer.SendFiles(stream)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantHeaders, stream.msgs)
		})
	}
}

func TestRecvFiles(t *testing.T) {
	testCases := map[string]struct {
		msgs                []*pb.FileTransferMessage
		recvAlreadyStarted  bool
		recvAlreadyFinished bool
		recvErr             error
		writeStreamErr      error
		wantFiles           []FileStat
		wantErr             bool
	}{
		"can recv files": {
			msgs: []*pb.FileTransferMessage{
				{
					Kind: &pb.FileTransferMessage_Header{
						Header: &pb.FileTransferHeader{
							TargetPath:          "testfileA",
							Mode:                0o644,
							OverrideServiceUnit: func() *string { s := "somesvcA"; return &s }(),
						},
					},
				},
				// Chunk messages left out since they would be consumed by the streamReadWriter
				{
					Kind: &pb.FileTransferMessage_Header{
						Header: &pb.FileTransferHeader{
							TargetPath: "testfileB",
							Mode:       0o644,
						},
					},
				},
				// Chunk messages left out since they would be consumed by the streamReadWriter
			},
			wantFiles: []FileStat{
				{
					SourcePath:          "testfileA",
					TargetPath:          "testfileA",
					Mode:                0o644,
					OverrideServiceUnit: "somesvcA",
				},
				{
					SourcePath: "testfileB",
					TargetPath: "testfileB",
					Mode:       0o644,
				},
			},
		},
		"no messages": {},
		"recv fails": {
			recvErr: errors.New("recv failed"),
			wantErr: true,
		},
		"first recv does not yield file header": {
			msgs: []*pb.FileTransferMessage{
				{
					Kind: &pb.FileTransferMessage_Chunk{},
				},
			},
			wantErr: true,
		},
		"write stream fails": {
			msgs: []*pb.FileTransferMessage{
				{
					Kind: &pb.FileTransferMessage_Header{
						Header: &pb.FileTransferHeader{
							TargetPath:          "testfileA",
							Mode:                0o644,
							OverrideServiceUnit: func() *string { s := "somesvcA"; return &s }(),
						},
					},
				},
				// Chunk messages left out since they would be consumed by the streamReadWriter
			},
			writeStreamErr: errors.New("write stream failed"),
			wantErr:        true,
		},
		"recv has already started": {
			msgs: []*pb.FileTransferMessage{
				{
					Kind: &pb.FileTransferMessage_Header{
						Header: &pb.FileTransferHeader{
							TargetPath:          "testfileA",
							Mode:                0o644,
							OverrideServiceUnit: func() *string { s := "somesvcA"; return &s }(),
						},
					},
				},
				// Chunk messages left out since they would be consumed by the streamReadWriter
			},
			recvAlreadyStarted: true,
			wantErr:            true,
		},
		"recv has already finished": {
			msgs: []*pb.FileTransferMessage{
				{
					Kind: &pb.FileTransferMessage_Header{
						Header: &pb.FileTransferHeader{
							TargetPath:          "testfileA",
							Mode:                0o644,
							OverrideServiceUnit: func() *string { s := "somesvcA"; return &s }(),
						},
					},
				},
				// Chunk messages left out since they would be consumed by the streamReadWriter
			},
			recvAlreadyFinished: true,
			wantErr:             true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			streamer := &stubStreamReadWriter{writeStreamErr: tc.writeStreamErr}
			stream := &fakeRecvFilesStream{msgs: tc.msgs, recvErr: tc.recvErr}
			transfer := New(slog.New(slog.NewTextHandler(logger.TestWriter{T: t}, nil)), streamer, false)
			if tc.recvAlreadyStarted {
				transfer.receiveStarted = true
			}
			if tc.recvAlreadyFinished {
				transfer.receiveFinished.Store(true)
			}
			err := transfer.RecvFiles(stream)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantFiles, transfer.files)
		})
	}
}

func TestGetSetFiles(t *testing.T) {
	testCases := map[string]struct {
		setFiles  *[]FileStat
		wantFiles []FileStat
		wantErr   bool
	}{
		"no files": {
			wantFiles: []FileStat{},
		},
		"files": {
			setFiles: &[]FileStat{
				{
					SourcePath:          "testfileA",
					TargetPath:          "testfileA",
					Mode:                0o644,
					OverrideServiceUnit: "somesvcA",
				},
			},
			wantFiles: []FileStat{
				{
					SourcePath:          "testfileA",
					TargetPath:          "testfileA",
					Mode:                0o644,
					OverrideServiceUnit: "somesvcA",
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			streamer := &dummyStreamReadWriter{}
			transfer := New(slog.New(slog.NewTextHandler(logger.TestWriter{T: t}, nil)), streamer, false)
			if tc.setFiles != nil {
				transfer.SetFiles(*tc.setFiles)
			}
			gotFiles := transfer.GetFiles()
			assert.Equal(tc.wantFiles, gotFiles)
			assert.Equal(tc.setFiles != nil, transfer.receiveFinished.Load())
		})
	}
}

func TestConcurrency(t *testing.T) {
	ft := New(slog.New(slog.NewTextHandler(logger.TestWriter{T: t}, nil)), &stubStreamReadWriter{}, false)

	sendFiles := func() {
		_ = ft.SendFiles(&stubSendFilesStream{})
	}

	recvFiles := func() {
		_ = ft.RecvFiles(&stubRecvFilesStream{})
	}

	getFiles := func() {
		_ = ft.GetFiles()
	}

	setFiles := func() {
		ft.SetFiles([]FileStat{{SourcePath: "file", TargetPath: "file", Mode: 0o644}})
	}

	go sendFiles()
	go sendFiles()
	go sendFiles()
	go sendFiles()
	go recvFiles()
	go recvFiles()
	go recvFiles()
	go recvFiles()
	go getFiles()
	go getFiles()
	go getFiles()
	go getFiles()
	go setFiles()
	go setFiles()
	go setFiles()
	go setFiles()
}

type stubStreamReadWriter struct {
	readStreamErr  error
	writeStreamErr error
}

func (s *stubStreamReadWriter) ReadStream(_ string, _ streamer.WriteChunkStream, _ uint, _ bool) error {
	return s.readStreamErr
}

func (s *stubStreamReadWriter) WriteStream(_ string, _ streamer.ReadChunkStream, _ bool) error {
	return s.writeStreamErr
}

type fakeRecvFilesStream struct {
	msgs    []*pb.FileTransferMessage
	pos     int
	recvErr error
}

func (s *fakeRecvFilesStream) Recv() (*pb.FileTransferMessage, error) {
	if s.recvErr != nil {
		return nil, s.recvErr
	}

	if s.pos < len(s.msgs) {
		s.pos++
		return s.msgs[s.pos-1], nil
	}

	return nil, io.EOF
}

type dummyStreamReadWriter struct{}

func (s *dummyStreamReadWriter) ReadStream(_ string, _ streamer.WriteChunkStream, _ uint, _ bool) error {
	panic("dummy")
}

func (s *dummyStreamReadWriter) WriteStream(_ string, _ streamer.ReadChunkStream, _ bool) error {
	panic("dummy")
}
