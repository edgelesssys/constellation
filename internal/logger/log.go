/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
Package logger provides helper functions that can be used in combination with slog to increase functionality or make
working with slog easier.

1. Logging in unit tests

To log in unit tests you can create a new slog logger that uses logger.testWriter as its writer. This can be constructed
by creating a logger like this: `logger.NewTest(t)`.

2. Creating a new logger with an increased log level based on another logger

You can create a new logger with a new log level by creating a new slog.Logger with the LevelHandler in this package
and passing the handler of the other logger. As an example, if you have a slog.Logger named `log` you can create a
new logger with an increased log level (here slog.LevelWarn) like this:

	slog.New(logger.NewLevelHandler(slog.LevelWarn, log.Handler()))
*/
package logger

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"testing"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"google.golang.org/grpc"
)

// ReplaceGRPCLogger replaces grpc's internal logger with the given logger.
func ReplaceGRPCLogger(l *slog.Logger) {
	replaceGRPCLogger(l)
}

// GetServerUnaryInterceptor returns a gRPC server option for intercepting unary gRPC logs.
func GetServerUnaryInterceptor(l *slog.Logger) grpc.ServerOption {
	return grpc.UnaryInterceptor(
		logging.UnaryServerInterceptor(middlewareLogger(l)),
	)
}

// GetServerStreamInterceptor returns a gRPC server option for intercepting streaming gRPC logs.
func GetServerStreamInterceptor(l *slog.Logger) grpc.ServerOption {
	return grpc.StreamInterceptor(
		logging.StreamServerInterceptor(middlewareLogger(l)),
	)
}

// GetClientUnaryInterceptor returns a gRPC client option for intercepting unary gRPC logs.
func GetClientUnaryInterceptor(l *slog.Logger) grpc.DialOption {
	return grpc.WithUnaryInterceptor(
		logging.UnaryClientInterceptor(middlewareLogger(l)),
	)
}

// GetClientStreamInterceptor returns a gRPC client option for intercepting stream gRPC logs.
func GetClientStreamInterceptor(l *slog.Logger) grpc.DialOption {
	return grpc.WithStreamInterceptor(
		logging.StreamClientInterceptor(middlewareLogger(l)),
	)
}

func middlewareLogger(l *slog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		f := make([]slog.Attr, 0, len(fields)/2)

		for i := 0; i < len(fields); i += 2 {
			key := fields[i]
			value := fields[i+1]

			switch v := value.(type) {
			case string:
				f = append(f, slog.String(key.(string), v))
			case int:
				f = append(f, slog.Int(key.(string), v))
			case bool:
				f = append(f, slog.Bool(key.(string), v))
			default:
				f = append(f, slog.Any(key.(string), v))
			}
		}

		var pcs [1]uintptr
		runtime.Callers(2, pcs[:]) // skip [Callers, LoggerFunc]
		level := slog.LevelDebug

		switch lvl {
		case logging.LevelDebug:
			break
		case logging.LevelInfo:
			level = slog.LevelInfo
		case logging.LevelWarn:
			level = slog.LevelWarn
		case logging.LevelError:
			level = slog.LevelError
		default:
			panic(fmt.Sprintf("unknown level %v", lvl))
		}

		r := slog.NewRecord(time.Now(), level, fmt.Sprintf(msg, fields...), pcs[0])
		_ = l.Handler().Handle(context.Background(), r)
	})
}

func NewTest(t *testing.T) *slog.Logger {
	return slog.New(slog.NewTextHandler(testWriter{t: t}, nil))
}

// TestWriter is a writer to a testing.T used in tests for logging with slog.
type testWriter struct {
	t *testing.T
}

func (t testWriter) Write(p []byte) (int, error) {
	t.t.Helper()
	t.t.Log(string(p))
	return len(p), nil
}
