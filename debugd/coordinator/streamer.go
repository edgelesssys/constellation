package coordinator

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	pb "github.com/edgelesssys/constellation/debugd/service"
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

// NewFileStreamer creates a new FileStreamer.
func NewFileStreamer(fs afero.Fs) *FileStreamer {
	return &FileStreamer{
		fs:  fs,
		mux: sync.RWMutex{},
	}
}

// WriteStream opens a file to write to and streams chunks from a gRPC stream into the file.
func (f *FileStreamer) WriteStream(filename string, stream ReadChunkStream, showProgress bool) error {
	// try to read from stream once before acquiring write lock
	chunk, err := stream.Recv()
	if err != nil {
		return fmt.Errorf("reading stream failed: %w", err)
	}

	f.mux.Lock()
	defer f.mux.Unlock()
	file, err := f.fs.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o755)
	if err != nil {
		return fmt.Errorf("open %v for writing failed: %w", filename, err)
	}
	defer file.Close()

	var bar *progressbar.ProgressBar
	if showProgress {
		bar = progressbar.NewOptions64(
			-1,
			progressbar.OptionSetDescription("receiving coordinator"),
			progressbar.OptionShowBytes(true),
			progressbar.OptionClearOnFinish(),
		)
		defer bar.Close()
	}

	for {
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			_ = file.Close()
			_ = f.fs.Remove(filename)
			return fmt.Errorf("reading stream failed: %w", err)
		}
		if _, err := file.Write(chunk.Content); err != nil {
			_ = file.Close()
			_ = f.fs.Remove(filename)
			return fmt.Errorf("writing chunk to disk failed: %w", err)
		}
		if showProgress {
			_ = bar.Add(len(chunk.Content))
		}
		chunk, err = stream.Recv()
	}

	return nil
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
		return errors.New("file is opened for writing cannot be read at this time")
	}
	file, err := f.fs.OpenFile(filename, os.O_RDONLY, 0o755)
	if err != nil {
		return fmt.Errorf("open %v for reading failed: %w", filename, err)
	}
	defer file.Close()

	var bar *progressbar.ProgressBar
	if showProgress {
		stat, err := file.Stat()
		if err != nil {
			return fmt.Errorf("performing stat on %v to get the file size failed: %w", filename, err)
		}
		bar = progressbar.NewOptions64(
			stat.Size(),
			progressbar.OptionSetDescription("uploading coordinator"),
			progressbar.OptionShowBytes(true),
			progressbar.OptionClearOnFinish(),
		)
		defer bar.Close()
	}

	buf := make([]byte, chunksize)
	for {
		n, err := file.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return fmt.Errorf("reading file chunk failed: %w", err)
		}

		if err = stream.Send(&pb.Chunk{Content: buf[:n]}); err != nil {
			return fmt.Errorf("sending chunk failed: %w", err)
		}
		if showProgress {
			_ = bar.Add(n)
		}
	}
}
