/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: BUSL-1.1
*/

package filetransfer

import (
	"errors"

	pb "github.com/edgelesssys/constellation/v2/debugd/service"
)

// recvChunkStream is a wrapper around a RecvFilesStream that only returns chunks.
type recvChunkStream struct {
	stream RecvFilesStream
}

// Recv receives a FileTransferMessage and returns the chunk.
func (s *recvChunkStream) Recv() (*pb.Chunk, error) {
	msg, err := s.stream.Recv()
	if err != nil {
		return nil, err
	}
	chunk := msg.GetChunk()
	if chunk == nil {
		return nil, errors.New("expected chunk")
	}
	return chunk, nil
}

// sendChunkStream is a wrapper around a SendFilesStream that wraps chunks for every message.
type sendChunkStream struct {
	stream SendFilesStream
}

// Send wraps the given chunk in a FileTransferMessage and sends it.
func (s *sendChunkStream) Send(chunk *pb.Chunk) error {
	chunkMessage := &pb.FileTransferMessage_Chunk{
		Chunk: chunk,
	}
	message := &pb.FileTransferMessage{
		Kind: chunkMessage,
	}
	return s.stream.Send(message)
}
