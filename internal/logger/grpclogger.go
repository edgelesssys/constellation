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
		logger:    log,
		verbosity: 0,
	}
	grpclog.SetLoggerV2(gl)
}

func (l *grpcLogger) log(level slog.Level, args ...interface{}) {
	if l.logger.Enabled(context.Background(), level) {
		var pcs [1]uintptr
		runtime.Callers(3, pcs[:])
		r := slog.NewRecord(time.Now(), level, fmt.Sprint(args...), pcs[0])
		_ = l.logger.Handler().Handle(context.Background(), r)
	}
}

func (l *grpcLogger) logf(level slog.Level, format string, args ...interface{}) {
	if l.logger.Enabled(context.Background(), level) {
		var pcs [1]uintptr
		runtime.Callers(3, pcs[:])
		r := slog.NewRecord(time.Now(), level, fmt.Sprintf(format, args...), pcs[0])
		_ = l.logger.Handler().Handle(context.Background(), r)
	}
}

type grpcLogger struct {
	logger    *slog.Logger
	verbosity int
}

func (l *grpcLogger) Info(args ...interface{}) {
	l.log(slog.LevelInfo, args...)
}

func (l *grpcLogger) Infoln(args ...interface{}) {
	l.log(slog.LevelInfo, args...)
}

func (l *grpcLogger) Infof(format string, args ...interface{}) {
	l.logf(slog.LevelInfo, format, args...)
}

func (l *grpcLogger) Warning(args ...interface{}) {
	l.log(slog.LevelWarn, args...)
}

func (l *grpcLogger) Warningln(args ...interface{}) {
	l.log(slog.LevelWarn, args...)
}

func (l *grpcLogger) Warningf(format string, args ...interface{}) {
	l.logf(slog.LevelWarn, format, args...)
}

func (l *grpcLogger) Error(args ...interface{}) {
	l.log(slog.LevelError, args...)
}

func (l *grpcLogger) Errorln(args ...interface{}) {
	l.log(slog.LevelError, args...)
}

func (l *grpcLogger) Errorf(format string, args ...interface{}) {
	l.logf(slog.LevelError, format, args...)
}

func (l *grpcLogger) Fatal(args ...interface{}) {
	l.log(slog.LevelError, args...)
	os.Exit(1)
}

func (l *grpcLogger) Fatalln(args ...interface{}) {
	l.log(slog.LevelError, args...)
	os.Exit(1)
}

func (l *grpcLogger) Fatalf(format string, args ...interface{}) {
	l.logf(slog.LevelError, format, args...)
	os.Exit(1)
}

func (l *grpcLogger) V(level int) bool {
	// Check whether the verbosity of the current log ('level') is within the specified threshold ('l.verbosity').
	// As in https://github.com/grpc/grpc-go/blob/41e044e1c82fcf6a5801d6cbd7ecf952505eecb1/grpclog/loggerv2.go#L199-L201.
	return level <= l.verbosity
}
