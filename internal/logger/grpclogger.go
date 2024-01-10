/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"time"

	"google.golang.org/grpc/grpclog"
)

func replaceGRPCLogger(log *slog.Logger) {
	gl := &grpcLogger{
		logger:    log.With(slog.String("system", "grpc"), slog.Bool("grpc_log", true)), 
		verbosity: 0,
	}
	grpclog.SetLoggerV2(gl)
}

type grpcLogger struct {
	logger    *slog.Logger
	verbosity int
}

func (l *grpcLogger) Info(args ...interface{}) {
  var pcs [1]uintptr
  runtime.Callers(2, pcs[:])
  r := slog.NewRecord(time.Now(), slog.LevelInfo, fmt.Sprint(args...), pcs[0])
  _ = l.logger.Handler().Handle(context.Background(), r)
}

func (l *grpcLogger) Infoln(args ...interface{}) {
	var pcs [1]uintptr
  runtime.Callers(2, pcs[:])
  r := slog.NewRecord(time.Now(), slog.LevelInfo, fmt.Sprint(args...), pcs[0])
  _ = l.logger.Handler().Handle(context.Background(), r)
}

func (l *grpcLogger) Infof(format string, args ...interface{}) {
	var pcs [1]uintptr
  runtime.Callers(2, pcs[:])
  r := slog.NewRecord(time.Now(), slog.LevelInfo, fmt.Sprintf(format, args...), pcs[0])
  _ = l.logger.Handler().Handle(context.Background(), r)
}

func (l *grpcLogger) Warning(args ...interface{}) {
	var pcs [1]uintptr
  runtime.Callers(2, pcs[:])
  r := slog.NewRecord(time.Now(), slog.LevelWarn, fmt.Sprint(args...), pcs[0])
  _ = l.logger.Handler().Handle(context.Background(), r)
}

func (l *grpcLogger) Warningln(args ...interface{}) {
	var pcs [1]uintptr
  runtime.Callers(2, pcs[:])
  r := slog.NewRecord(time.Now(), slog.LevelWarn, fmt.Sprint(args...), pcs[0])
  _ = l.logger.Handler().Handle(context.Background(), r)
}

func (l *grpcLogger) Warningf(format string, args ...interface{}) {
	var pcs [1]uintptr
  runtime.Callers(2, pcs[:])
  r := slog.NewRecord(time.Now(), slog.LevelWarn, fmt.Sprintf(format, args...), pcs[0])
  _ = l.logger.Handler().Handle(context.Background(), r)
}

func (l *grpcLogger) Error(args ...interface{}) {
	var pcs [1]uintptr
  runtime.Callers(2, pcs[:])
  r := slog.NewRecord(time.Now(), slog.LevelError, fmt.Sprint(args...), pcs[0])
  _ = l.logger.Handler().Handle(context.Background(), r)
}

func (l *grpcLogger) Errorln(args ...interface{}) {
	var pcs [1]uintptr
  runtime.Callers(2, pcs[:])
  r := slog.NewRecord(time.Now(), slog.LevelError, fmt.Sprint(args...), pcs[0])
  _ = l.logger.Handler().Handle(context.Background(), r)
}

func (l *grpcLogger) Errorf(format string, args ...interface{}) {
	var pcs [1]uintptr
  runtime.Callers(2, pcs[:])
  r := slog.NewRecord(time.Now(), slog.LevelError, fmt.Sprintf(format, args...), pcs[0])
  _ = l.logger.Handler().Handle(context.Background(), r)
}

func (l *grpcLogger) Fatal(args ...interface{}) {
	var pcs [1]uintptr
  runtime.Callers(2, pcs[:])
  r := slog.NewRecord(time.Now(), slog.LevelError, fmt.Sprint(args...), pcs[0])
  _ = l.logger.Handler().Handle(context.Background(), r)
  os.Exit(1)
}

func (l *grpcLogger) Fatalln(args ...interface{}) {
	var pcs [1]uintptr
  runtime.Callers(2, pcs[:])
  r := slog.NewRecord(time.Now(), slog.LevelError, fmt.Sprint(args...), pcs[0])
  _ = l.logger.Handler().Handle(context.Background(), r)
  os.Exit(1)
}

func (l *grpcLogger) Fatalf(format string, args ...interface{}) {
	var pcs [1]uintptr
  runtime.Callers(2, pcs[:])
  r := slog.NewRecord(time.Now(), slog.LevelError, fmt.Sprintf(format, args...), pcs[0])
  _ = l.logger.Handler().Handle(context.Background(), r)
  os.Exit(1)
}

func (l *grpcLogger) V(level int) bool {
	// Check whether the verbosity of the current log ('level') is within the specified threshold ('l.verbosity').
	// As in https://github.com/grpc/grpc-go/blob/41e044e1c82fcf6a5801d6cbd7ecf952505eecb1/grpclog/loggerv2.go#L199-L201.
	return level <= l.verbosity
}
