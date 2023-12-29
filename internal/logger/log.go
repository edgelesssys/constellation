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
	"os"
	"testing"
  "runtime"
  "time"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc"
)

// LogType indicates the output encoding of the log.
type LogType int

const (
	// JSONLog encodes logs in JSON format.
	JSONLog LogType = iota
	// PlainLog encodes logs as human readable text.
	PlainLog
)

// Logger is a wrapper for zap logger.
// The purpose is to provide a simple interface for logging with sensible defaults.
type Logger struct {
	logger *zap.SugaredLogger
}

// New creates a new Logger.
// Set name to an empty string to create an unnamed logger.
func New(logType LogType, logLevel zapcore.Level) *Logger {
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.StacktraceKey = zapcore.OmitKey
	encoderCfg.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderCfg.EncodeTime = zapcore.RFC3339TimeEncoder

	var encoder zapcore.Encoder
	if logType == PlainLog {
		encoder = zapcore.NewConsoleEncoder(encoderCfg)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderCfg)
	}

	logCore := zapcore.NewCore(encoder, zapcore.Lock(os.Stderr), zap.NewAtomicLevelAt(logLevel))

	logger := zap.New(
		logCore,
		zap.AddCaller(),      // add the file and line number of the logging call
		zap.AddCallerSkip(1), // skip the first caller so that we don't only see this package as the caller
	)

	return &Logger{logger: logger.Sugar()}
}

// NewTest creates a logger for unit / integration tests.
func NewTest(t *testing.T) *Logger {
	return &Logger{
		logger: zaptest.NewLogger(t).Sugar().Named(fmt.Sprintf("%q", t.Name())),
	}
}

// Debugf logs a message at Debug level.
// Debug logs are typically voluminous, and contain detailed information on the flow of execution.
func (l *Logger) Debugf(format string, args ...any) {
	l.logger.Debugf(format, args...)
}

// Infof logs a message at Info level.
// This is the default logging priority and should be used for all normal messages.
func (l *Logger) Infof(format string, args ...any) {
	l.logger.Infof(format, args...)
}

// Warnf logs a message at Warn level.
// Warn logs are more important than Info, but they don't need human review or necessarily indicate an error.
func (l *Logger) Warnf(format string, args ...any) {
	l.logger.Warnf(format, args...)
}

// Errorf logs a message at Error level.
// Error logs are high priority and indicate something has gone wrong.
func (l *Logger) Errorf(format string, args ...any) {
	l.logger.Errorf(format, args...)
}

// Fatalf logs the message and then calls os.Exit(1).
// Use this to exit your program when a fatal error occurs.
func (l *Logger) Fatalf(format string, args ...any) {
	l.logger.Fatalf(format, args...)
}

// Sync flushes any buffered log entries.
// Applications should take care to call Sync before exiting.
func (l *Logger) Sync() {
	_ = l.logger.Sync()
}

// WithIncreasedLevel returns a logger with increased logging level.
func (l *Logger) WithIncreasedLevel(level zapcore.Level) *Logger {
	return &Logger{logger: l.getZapLogger().WithOptions(zap.IncreaseLevel(level)).Sugar()}
}

// With returns a logger with structured context.
func (l *Logger) With(fields ...any) *Logger {
	return &Logger{logger: l.logger.With(fields...)}
}

// Named returns a named logger.
func (l *Logger) Named(name string) *Logger {
	return &Logger{logger: l.logger.Named(name)}
}

// ReplaceGRPCLogger replaces grpc's internal logger with the given logger.
func (l *Logger) ReplaceGRPCLogger() {
	replaceGRPCLogger(l.getZapLogger())
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

// getZapLogger returns the underlying zap logger.
func (l *Logger) getZapLogger() *zap.Logger {
	return l.logger.Desugar()
}

func middlewareLogger(l *slog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		f := make([]zap.Field, 0, len(fields)/2)

		for i := 0; i < len(fields); i += 2 {
			key := fields[i]
			value := fields[i+1]

			switch v := value.(type) {
			case string:
				f = append(f, zap.String(key.(string), v))
			case int:
				f = append(f, zap.Int(key.(string), v))
			case bool:
				f = append(f, zap.Bool(key.(string), v))
			default:
				f = append(f, zap.Any(key.(string), v))
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
