/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package filetransfer

import (
	"errors"
	"testing"

	pb "github.com/edgelesssys/constellation/v2/debugd/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecv(t *testing.T) {
	testCases := map[string]struct {
		stream    stubRecvFilesStream
		wantChunk *pb.Chunk
		wantErr   bool
	}{
		"chunk is received": {
			stream: stubRecvFilesStream{
				msg: &pb.FileTransferMessage{
					Kind: &pb.FileTransferMessage_Chunk{
						Chunk: &pb.Chunk{
							Content: []byte("test"),
						},
					},
				},
			},
			wantChunk: &pb.Chunk{
				Content: []byte("test"),
			},
		},
		"wrong type": {
			stream: stubRecvFilesStream{
				msg: &pb.FileTransferMessage{
					Kind: &pb.FileTransferMessage_Header{},
				},
			},
			wantErr: true,
		},
		"empty msg": {
			stream:  stubRecvFilesStream{},
			wantErr: true,
		},
		"recv fails": {
			stream: stubRecvFilesStream{
				recvErr: errors.New("someErr"),
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			stream := recvChunkStream{stream: &tc.stream}
			chunk, err := stream.Recv()

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantChunk, chunk)
		})
	}
}

func TestSend(t *testing.T) {
	testCases := map[string]struct {
		stream   stubSendFilesStream
		chunk    *pb.Chunk
		wantMsgs []*pb.FileTransferMessage
		wantErr  bool
	}{
		"chunk is wrapped correctly": {
			chunk: &pb.Chunk{
				Content: []byte("test"),
			},
			wantMsgs: []*pb.FileTransferMessage{
				{
					Kind: &pb.FileTransferMessage_Chunk{
						Chunk: &pb.Chunk{
							Content: []byte("test"),
						},
					},
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			stream := sendChunkStream{stream: &tc.stream}
			err := stream.Send(tc.chunk)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.EqualValues(tc.wantMsgs, tc.stream.msgs)
		})
	}
}

type stubRecvFilesStream struct {
	msg     *pb.FileTransferMessage
	recvErr error
}

func (s *stubRecvFilesStream) Recv() (*pb.FileTransferMessage, error) {
	return s.msg, s.recvErr
}

type stubSendFilesStream struct {
	msgs    []*pb.FileTransferMessage
	sendErr error
}

func (s *stubSendFilesStream) Send(msg *pb.FileTransferMessage) error {
	s.msgs = append(s.msgs, msg)
	return s.sendErr
}
