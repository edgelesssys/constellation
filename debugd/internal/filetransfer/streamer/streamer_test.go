/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package streamer

import (
	"errors"
	"io"
	"testing"

	pb "github.com/edgelesssys/constellation/v2/debugd/service"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m, goleak.IgnoreAnyFunction("github.com/bazelbuild/rules_go/go/tools/bzltestutil.RegisterTimeoutHandler.func1"))
}

func TestWriteStream(t *testing.T) {
	filename := "testfile"

	testCases := map[string]struct {
		readChunkStream fakeReadChunkStream
		fs              afero.Fs
		showProgress    bool
		wantFile        []byte
		wantErr         bool
	}{
		"stream works": {
			readChunkStream: fakeReadChunkStream{
				chunks: [][]byte{
					[]byte("test"),
				},
			},
			fs:       afero.NewMemMapFs(),
			wantFile: []byte("test"),
			wantErr:  false,
		},
		"chunking works": {
			readChunkStream: fakeReadChunkStream{
				chunks: [][]byte{
					[]byte("te"),
					[]byte("st"),
				},
			},
			fs:       afero.NewMemMapFs(),
			wantFile: []byte("test"),
			wantErr:  false,
		},
		"showProgress works": {
			readChunkStream: fakeReadChunkStream{
				chunks: [][]byte{
					[]byte("test"),
				},
			},
			fs:           afero.NewMemMapFs(),
			showProgress: true,
			wantFile:     []byte("test"),
			wantErr:      false,
		},
		"Open fails": {
			fs:      afero.NewReadOnlyFs(afero.NewMemMapFs()),
			wantErr: true,
		},
		"recv fails": {
			readChunkStream: fakeReadChunkStream{
				recvErr: errors.New("someErr"),
			},
			fs:      afero.NewMemMapFs(),
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			writer := New(tc.fs)
			err := writer.WriteStream(filename, &tc.readChunkStream, tc.showProgress)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			fileContents, err := afero.ReadFile(tc.fs, filename)
			require.NoError(err)
			assert.Equal(tc.wantFile, fileContents)
		})
	}
}

func TestReadStream(t *testing.T) {
	correctFilename := "testfile"
	eof := []byte{}

	testCases := map[string]struct {
		writeChunkStream stubWriteChunkStream
		filename         string
		chunksize        uint
		showProgress     bool
		wantChunks       [][]byte
		wantErr          bool
	}{
		"stream works": {
			writeChunkStream: stubWriteChunkStream{},
			filename:         correctFilename,
			chunksize:        4,
			wantChunks: [][]byte{
				[]byte("test"),
				eof,
			},
			wantErr: false,
		},
		"chunking works": {
			writeChunkStream: stubWriteChunkStream{},
			filename:         correctFilename,
			chunksize:        2,
			wantChunks: [][]byte{
				[]byte("te"),
				[]byte("st"),
				eof,
			},
			wantErr: false,
		},
		"chunksize of 0 detected": {
			writeChunkStream: stubWriteChunkStream{},
			filename:         correctFilename,
			chunksize:        0,
			wantErr:          true,
		},
		"showProgress works": {
			writeChunkStream: stubWriteChunkStream{},
			filename:         correctFilename,
			chunksize:        4,
			showProgress:     true,
			wantChunks: [][]byte{
				[]byte("test"),
				eof,
			},
			wantErr: false,
		},
		"Open fails": {
			filename:  "incorrect-filename",
			chunksize: 4,
			wantErr:   true,
		},
		"send fails": {
			writeChunkStream: stubWriteChunkStream{
				sendErr: errors.New("someErr"),
			},
			filename:  correctFilename,
			chunksize: 4,
			wantErr:   true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			fs := afero.NewMemMapFs()
			assert.NoError(afero.WriteFile(fs, correctFilename, []byte("test"), 0o755))
			reader := New(fs)
			err := reader.ReadStream(tc.filename, &tc.writeChunkStream, tc.chunksize, tc.showProgress)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantChunks, tc.writeChunkStream.chunks)
		})
	}
}

type fakeReadChunkStream struct {
	chunks  [][]byte
	pos     int
	recvErr error
}

func (s *fakeReadChunkStream) Recv() (*pb.Chunk, error) {
	if s.recvErr != nil {
		return nil, s.recvErr
	}

	isLastChunk := s.pos == len(s.chunks)-1

	if s.pos < len(s.chunks) {
		result := &pb.Chunk{Content: s.chunks[s.pos], Last: isLastChunk}
		s.pos++
		return result, nil
	}

	return nil, io.EOF
}

type stubWriteChunkStream struct {
	chunks  [][]byte
	sendErr error
}

func (s *stubWriteChunkStream) Send(chunk *pb.Chunk) error {
	cpy := make([]byte, len(chunk.Content))
	copy(cpy, chunk.Content)
	s.chunks = append(s.chunks, cpy)
	return s.sendErr
}
