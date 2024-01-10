/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
Package logger provides logging functionality for Constellation services.
It is a thin wrapper around the zap package, providing a consistent interface for logging.
Use this package to implement logging for your Constellation services.

# Usage

1. Create a logger using New().

2. Defer the Sync() method to ensure that all log entries are flushed.

3. Use the Debugf(), Infof(), Warnf(), Errorf(), and Fatalf() methods depending on the level of logging you need.

4. Use the Named() method to create a named child logger.

5. Use the With() method to create a child logger with structured context.
This can also be used to add context to a single log message:

	logger.With(zap.String("key", "value")).Infof("log message")

# Log Levels

Use [Logger.Debugf] to log low level and detailed information that is useful for debugging.

Use [Logger.Infof] to log general information. This method is correct for most logging purposes.

Use [Logger.Warnf] to log information that may indicate unwanted behavior, but is not an error.

Use [Logger.Errorf] to log information about any errors that occurred.

Use [Logger.Fatalf] to log information about any errors that occurred and then exit the program.
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
    r := slog.Record{}

		switch lvl {
		case logging.LevelDebug:
      r = slog.NewRecord(time.Now(), slog.LevelDebug, fmt.Sprintf(msg, fields...), pcs[0])
		case logging.LevelInfo:
      r = slog.NewRecord(time.Now(), slog.LevelInfo, fmt.Sprintf(msg, fields...), pcs[0])
		case logging.LevelWarn:
      r = slog.NewRecord(time.Now(), slog.LevelWarn, fmt.Sprintf(msg, fields...), pcs[0])
		case logging.LevelError:
      r = slog.NewRecord(time.Now(), slog.LevelError, fmt.Sprintf(msg, fields...), pcs[0])
		default:
			panic(fmt.Sprintf("unknown level %v", lvl))
		}
    _ = l.Handler().Handle(context.Background(), r)
	})
}

// LevelHandler copied from the official LevelHandler example in the slog package documentation.

// LevelHandler wraps a Handler with an Enabled method
// that returns false for levels below a minimum.
type LevelHandler struct {
	level   slog.Leveler
	handler slog.Handler
}

// NewLevelHandler returns a LevelHandler with the given level.
// All methods except Enabled delegate to h.
func NewLevelHandler(level slog.Leveler, h slog.Handler) *LevelHandler {
	// Optimization: avoid chains of LevelHandlers.
	if lh, ok := h.(*LevelHandler); ok {
		h = lh.Handler()
	}
	return &LevelHandler{level, h}
}

// Enabled implements Handler.Enabled by reporting whether
// level is at least as large as h's level.
func (h *LevelHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level.Level()
}

// Handle implements Handler.Handle.
func (h *LevelHandler) Handle(ctx context.Context, r slog.Record) error {
	return h.handler.Handle(ctx, r)
}

// WithAttrs implements Handler.WithAttrs.
func (h *LevelHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return NewLevelHandler(h.level, h.handler.WithAttrs(attrs))
}

// WithGroup implements Handler.WithGroup.
func (h *LevelHandler) WithGroup(name string) slog.Handler {
	return NewLevelHandler(h.level, h.handler.WithGroup(name))
}

// Handler returns the Handler wrapped by h.
func (h *LevelHandler) Handler() slog.Handler {
	return h.handler
}

// TestWriter is a writer to a testing.T used in tests for logging with slog.
type TestWriter struct {
  T *testing.T
}

func (t TestWriter) Write(p []byte) (int, error) {
  t.T.Log(p)
  return len(p), nil
}
