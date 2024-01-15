/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package filetransfer implements the exchange of files between cdgb <-> debugd
// and between debugd <-> debugd pairs.
package filetransfer

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/edgelesssys/constellation/v2/debugd/internal/debugd"
	"github.com/edgelesssys/constellation/v2/debugd/internal/filetransfer/streamer"
	pb "github.com/edgelesssys/constellation/v2/debugd/service"
)

// RecvFilesStream is a stream that receives FileTransferMessages.
type RecvFilesStream interface {
	Recv() (*pb.FileTransferMessage, error)
}

// SendFilesStream is a stream that sends FileTransferMessages.
type SendFilesStream interface {
	Send(*pb.FileTransferMessage) error
}

// FileTransferer manages sending and receiving of files.
type FileTransferer struct {
	fileMux         sync.RWMutex
	log             *slog.Logger
	receiveStarted  bool
	receiveFinished atomic.Bool
	files           []FileStat
	streamer        streamReadWriter
	showProgress    bool
}

// New creates a new FileTransferer.
func New(log *slog.Logger, streamer streamReadWriter, showProgress bool) *FileTransferer {
	return &FileTransferer{
		log:          log,
		streamer:     streamer,
		showProgress: showProgress,
	}
}

// SendFiles sends files to the given stream.
// If the FileTransferer has not received any files to send, an error is returned.
func (s *FileTransferer) SendFiles(stream SendFilesStream) error {
	if !s.receiveFinished.Load() {
		return errors.New("cannot send files before receiving them")
	}

	s.fileMux.RLock()
	defer s.fileMux.RUnlock()

	for _, file := range s.files {
		if err := s.handleFileSend(stream, file); err != nil {
			return err
		}
	}
	return nil
}

// RecvFiles receives files from the given stream.
func (s *FileTransferer) RecvFiles(stream RecvFilesStream) (err error) {
	s.fileMux.Lock()
	defer s.fileMux.Unlock()
	if err := s.startRecv(); err != nil {
		return err
	}
	defer func() {
		if err != nil {
			s.abortRecv()
		} else {
			s.finishRecv()
		}
	}()
	var done bool
	for !done && err == nil {
		done, err = s.handleFileRecv(stream)
	}
	return err
}

// GetFiles returns the a copy of the list of files that have been received.
func (s *FileTransferer) GetFiles() []FileStat {
	s.fileMux.RLock()
	defer s.fileMux.RUnlock()
	res := make([]FileStat, len(s.files))
	copy(res, s.files)
	return res
}

// SetFiles sets the list of files that can be sent.
// This function is used for a sender which has not received any files through
// this FileTransferer i.e. the CLI.
func (s *FileTransferer) SetFiles(files []FileStat) {
	s.fileMux.Lock()
	defer s.fileMux.Unlock()
	res := make([]FileStat, len(files))
	copy(res, files)
	s.files = res
	s.receiveFinished.Store(true)
}

func (s *FileTransferer) handleFileSend(stream SendFilesStream, file FileStat) error {
	header := &pb.FileTransferMessage_Header{
		Header: &pb.FileTransferHeader{
			TargetPath: file.TargetPath,
			Mode:       uint32(file.Mode),
		},
	}
	if file.OverrideServiceUnit != "" {
		header.Header.OverrideServiceUnit = &file.OverrideServiceUnit
	}
	if err := stream.Send(&pb.FileTransferMessage{Kind: header}); err != nil {
		return err
	}
	sendChunkStream := &sendChunkStream{stream: stream}
	return s.streamer.ReadStream(file.SourcePath, sendChunkStream, debugd.Chunksize, s.showProgress)
}

// handleFileRecv handles the file receive of a single file.
// It returns true if the stream is finished (all of the file consumed) and false otherwise.
func (s *FileTransferer) handleFileRecv(stream RecvFilesStream) (bool, error) {
	// first message must be a header message
	msg, err := stream.Recv()
	switch {
	case err == nil:
		// nop
	case errors.Is(err, io.EOF):
		return true, nil // stream is finished
	default:
		return false, err
	}
	header := msg.GetHeader()
	if header == nil {
		return false, errors.New("first message must be a header message")
	}
	s.log.Info(fmt.Sprintf("Starting file receive of %q", header.TargetPath))
	s.addFile(FileStat{
		SourcePath: header.TargetPath,
		TargetPath: header.TargetPath,
		Mode:       fs.FileMode(header.Mode),
		OverrideServiceUnit: func() string {
			if header.OverrideServiceUnit != nil {
				return *header.OverrideServiceUnit
			}
			return ""
		}(),
	})
	recvChunkStream := &recvChunkStream{stream: stream}
	if err := s.streamer.WriteStream(header.TargetPath, recvChunkStream, s.showProgress); err != nil {
		s.log.With(slog.Any("error", err)).Error(fmt.Sprintf("Receive of file %q failed", header.TargetPath))
		return false, err
	}
	s.log.Info(fmt.Sprintf("Finished file receive of %q", header.TargetPath))
	return false, nil
}

// startRecv marks the file receive as started. It returns an error if receiving has already started.
func (s *FileTransferer) startRecv() error {
	switch {
	case s.receiveFinished.Load():
		return ErrReceiveFinished
	case s.receiveStarted:
		return ErrReceiveRunning
	}
	s.receiveStarted = true
	return nil
}

// abortRecv marks the file receive as failed.
// This allows for a retry of the file receive.
func (s *FileTransferer) abortRecv() {
	s.receiveStarted = false
	s.files = nil
}

// finishRecv marks the file receive as completed.
// This allows other debugd instances to request files from this server.
func (s *FileTransferer) finishRecv() {
	s.receiveStarted = false
	s.receiveFinished.Store(true)
}

// addFile adds a file to the list of received files.
func (s *FileTransferer) addFile(file FileStat) {
	s.files = append(s.files, file)
}

// FileStat contains the metadata of a file that can be up/downloaded.
type FileStat struct {
	SourcePath          string
	TargetPath          string
	Mode                fs.FileMode
	OverrideServiceUnit string // optional name of the service unit to override
}

var (
	// ErrReceiveRunning is returned if a file receive is already running.
	ErrReceiveRunning = errors.New("receive already running")
	// ErrReceiveFinished is returned if a file receive has already finished.
	ErrReceiveFinished = errors.New("receive finished")
)

const (
	// ShowProgress indicates that progress should be shown.
	ShowProgress = true
	// DontShowProgress indicates that progress should not be shown.
	DontShowProgress = false
)

type streamReadWriter interface {
	WriteStream(filename string, stream streamer.ReadChunkStream, showProgress bool) error
	ReadStream(filename string, stream streamer.WriteChunkStream, chunksize uint, showProgress bool) error
}
