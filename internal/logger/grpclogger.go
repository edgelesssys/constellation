/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package logger

import (
	"fmt"
	"log/slog"
	"runtime"

	"google.golang.org/grpc/grpclog"
)

func replaceGRPCLogger(log *slog.Logger) {
	// get correct reference in callstack
	// to log function that called the
	// wrapper
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:])

	gl := &grpcLogger{
		logger:    log.With(slog.String("system", "grpc"), slog.Bool("grpc_log", true), pcs[0]),
		verbosity: 0,
	}
	grpclog.SetLoggerV2(gl)
}

type grpcLogger struct {
	logger    *slog.Logger
	verbosity int
}

func (l *grpcLogger) Info(args ...interface{}) {
	l.logger.Info(fmt.Sprint(args...))
}

func (l *grpcLogger) Infoln(args ...interface{}) {
	l.logger.Info(fmt.Sprint(args...))
}

func (l *grpcLogger) Infof(format string, args ...interface{}) {
	l.logger.Info(fmt.Sprintf(format, args...))
}

func (l *grpcLogger) Warning(args ...interface{}) {
	l.logger.Warn(fmt.Sprint(args...))
}

func (l *grpcLogger) Warningln(args ...interface{}) {
	l.logger.Warn(fmt.Sprint(args...))
}

func (l *grpcLogger) Warningf(format string, args ...interface{}) {
	l.logger.Warn(fmt.Sprintf(format, args...))
}

func (l *grpcLogger) Error(args ...interface{}) {
	l.logger.Error(fmt.Sprint(args...))
}

func (l *grpcLogger) Errorln(args ...interface{}) {
	l.logger.Error(fmt.Sprint(args...))
}

func (l *grpcLogger) Errorf(format string, args ...interface{}) {
	l.logger.Error(fmt.Sprintf(format, args...))
}

func (l *grpcLogger) Fatal(args ...interface{}) {
	l.Fatal(fmt.Sprint(args...))
}

func (l *grpcLogger) Fatalln(args ...interface{}) {
	l.Fatal(fmt.Sprint(args...))
}

func (l *grpcLogger) Fatalf(format string, args ...interface{}) {
	l.Fatal(fmt.Sprintf(format, args...))
}

func (l *grpcLogger) V(level int) bool {
	// Check whether the verbosity of the current log ('level') is within the specified threshold ('l.verbosity').
	// As in https://github.com/grpc/grpc-go/blob/41e044e1c82fcf6a5801d6cbd7ecf952505eecb1/grpclog/loggerv2.go#L199-L201.
	return level <= l.verbosity
}
