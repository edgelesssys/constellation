/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package logger

import (
	"context"
	"log/slog"
	"os"

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
	if !l.logger.Enabled(context.Background(), LevelInfo) {
		return
	}
	originalFunction := getOriginalCaller()
	createAndLogRecord(l.logger, LevelInfo, originalFunction, args...)
}

func (l *grpcLogger) Infoln(args ...interface{}) {
	if !l.logger.Enabled(context.Background(), LevelInfo) {
		return
	}
	originalFunction := getOriginalCaller()
	createAndLogRecord(l.logger, LevelInfo, originalFunction, args...)
}

func (l *grpcLogger) Infof(format string, args ...interface{}) {
	if !l.logger.Enabled(context.Background(), LevelInfo) {
		return
	}
	originalFunction := getOriginalCaller()
	createAndLogRecordf(l.logger, LevelInfo, originalFunction, format, args...)
}

func (l *grpcLogger) Warning(args ...interface{}) {
	if !l.logger.Enabled(context.Background(), LevelWarn) {
		return
	}
	originalFunction := getOriginalCaller()
	createAndLogRecord(l.logger, LevelWarn, originalFunction, args...)
}

func (l *grpcLogger) Warningln(args ...interface{}) {
	if !l.logger.Enabled(context.Background(), LevelWarn) {
		return
	}
	originalFunction := getOriginalCaller()
	createAndLogRecord(l.logger, LevelWarn, originalFunction, args...)
}

func (l *grpcLogger) Warningf(format string, args ...interface{}) {
	if !l.logger.Enabled(context.Background(), LevelWarn) {
		return
	}
	originalFunction := getOriginalCaller()
	createAndLogRecordf(l.logger, LevelWarn, originalFunction, format, args...)
}

func (l *grpcLogger) Error(args ...interface{}) {
	if !l.logger.Enabled(context.Background(), LevelError) {
		return
	}
	originalFunction := getOriginalCaller()
	createAndLogRecord(l.logger, LevelError, originalFunction, args...)
}

func (l *grpcLogger) Errorln(args ...interface{}) {
	if !l.logger.Enabled(context.Background(), LevelError) {
		return
	}
	originalFunction := getOriginalCaller()
	createAndLogRecord(l.logger, LevelError, originalFunction, args...)
}

func (l *grpcLogger) Errorf(format string, args ...interface{}) {
	if !l.logger.Enabled(context.Background(), LevelError) {
		return
	}
	originalFunction := getOriginalCaller()
	createAndLogRecordf(l.logger, LevelError, originalFunction, format, args...)
}

func (l *grpcLogger) Fatal(args ...interface{}) {
	if !l.logger.Enabled(context.Background(), LevelFatal) {
		return
	}
	originalFunction := getOriginalCaller()
	createAndLogRecord(l.logger, LevelFatal, originalFunction, args...)
	os.Exit(1)
}

func (l *grpcLogger) Fatalln(args ...interface{}) {
	if !l.logger.Enabled(context.Background(), LevelFatal) {
		return
	}
	originalFunction := getOriginalCaller()
	createAndLogRecord(l.logger, LevelFatal, originalFunction, args...)
	os.Exit(1)
}

func (l *grpcLogger) Fatalf(format string, args ...interface{}) {
	if !l.logger.Enabled(context.Background(), LevelFatal) {
		return
	}
	originalFunction := getOriginalCaller()
	createAndLogRecordf(l.logger, LevelFatal, originalFunction, format, args...)
	os.Exit(1)
}

func (l *grpcLogger) V(level int) bool {
	// Check whether the verbosity of the current log ('level') is within the specified threshold ('l.verbosity').
	// As in https://github.com/grpc/grpc-go/blob/41e044e1c82fcf6a5801d6cbd7ecf952505eecb1/grpclog/loggerv2.go#L199-L201.
	return level <= l.verbosity
}
