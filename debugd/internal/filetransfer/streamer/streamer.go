/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package streamer

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	pb "github.com/edgelesssys/constellation/v2/debugd/service"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/afero"
)

// FileStreamer handles reading and writing of a file using a stream of chunks.
type FileStreamer struct {
	fs  afero.Fs
	mux sync.RWMutex
}

// ReadChunkStream is abstraction over a gRPC stream that allows us to receive chunks via gRPC.
type ReadChunkStream interface {
	Recv() (*pb.Chunk, error)
}

// WriteChunkStream is abstraction over a gRPC stream that allows us to send chunks via gRPC.
type WriteChunkStream interface {
	Send(chunk *pb.Chunk) error
}

// New creates a new FileStreamer.
func New(fs afero.Fs) *FileStreamer {
	return &FileStreamer{
		fs:  fs,
		mux: sync.RWMutex{},
	}
}

// WriteStream opens a file to write to and streams chunks from a gRPC stream into the file.
func (f *FileStreamer) WriteStream(filename string, stream ReadChunkStream, showProgress bool) error {
	f.mux.Lock()
	defer f.mux.Unlock()
	file, err := f.fs.OpenFile(filename, os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		return fmt.Errorf("open %v for writing: %w", filename, err)
	}
	defer file.Close()
	stat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("performing stat on %v to get the file size: %w", filename, err)
	}

	var bar *progressbar.ProgressBar
	if showProgress {
		bar = newProgressBar(stat.Size())
		defer bar.Close()
	}

	return writeInner(file, stream, bar)
}

// ReadStream opens a file to read from and streams its contents chunkwise over gRPC.
func (f *FileStreamer) ReadStream(filename string, stream WriteChunkStream, chunksize uint, showProgress bool) error {
	if chunksize == 0 {
		return errors.New("invalid chunksize")
	}
	// fail if file is currently RW locked
	if f.mux.TryRLock() {
		defer f.mux.RUnlock()
	} else {
		return errors.New("a file is opened for writing so cannot read at this time")
	}
	file, err := f.fs.OpenFile(filename, os.O_RDONLY, 0o755)
	if err != nil {
		return fmt.Errorf("open %v for reading: %w", filename, err)
	}
	defer file.Close()
	stat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("performing stat on %v to get the file size: %w", filename, err)
	}

	var bar *progressbar.ProgressBar
	if showProgress {
		bar = newProgressBar(stat.Size())
		defer bar.Close()
	}

	return readInner(file, stream, chunksize, bar)
}

// readInner reads from a an io.Reader and sends chunks over a gRPC stream.
func readInner(fp io.Reader, stream WriteChunkStream, chunksize uint, bar *progressbar.ProgressBar) error {
	buf := make([]byte, chunksize)
	for {
		n, readErr := fp.Read(buf)
		isLast := errors.Is(readErr, io.EOF)
		if readErr != nil && !isLast {
			return fmt.Errorf("reading file chunk: %w", readErr)
		}
		if err := stream.Send(&pb.Chunk{Content: buf[:n], Last: isLast}); err != nil {
			return fmt.Errorf("sending chunk: %w", err)
		}
		if bar != nil {
			_ = bar.Add(n)
		}
		if isLast {
			return nil
		}
	}
}

// writeInner writes chunks from a gRPC stream to an io.Writer.
func writeInner(fp io.Writer, stream ReadChunkStream, bar *progressbar.ProgressBar) error {
	for {
		chunk, recvErr := stream.Recv()
		if recvErr != nil {
			return fmt.Errorf("reading stream: %w", recvErr)
		}
		if _, err := fp.Write(chunk.Content); err != nil {
			return fmt.Errorf("writing chunk to disk: %w", err)
		}
		if bar != nil {
			_ = bar.Add(len(chunk.Content))
		}
		if chunk.Last {
			return nil
		}
	}
}

// newProgressBar creates a new progress bar.
func newProgressBar(size int64) *progressbar.ProgressBar {
	return progressbar.NewOptions64(
		size,
		progressbar.OptionSetDescription("transferring file"),
		progressbar.OptionShowBytes(true),
		progressbar.OptionClearOnFinish(),
	)
}
