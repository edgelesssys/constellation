/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
Package logger provides logging functionality for Constellation services.
It is a thin wrapper around the slog package, providing a consistent interface for logging.
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
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"go.uber.org/zap"
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

const (
	LevelDebug = slog.LevelDebug
	LevelInfo  = slog.LevelInfo
	LevelWarn  = slog.LevelWarn
	LevelError = slog.LevelError
	LevelFatal = 12
)

func replaceAttrFunction(groups []string, a slog.Attr) slog.Attr {
	// change the time format to rfc 3339
	if a.Key == slog.TimeKey {
		logTime := a.Value.Any().(time.Time)
		a.Value = slog.StringValue(logTime.Format(time.RFC3339))
	}

	// include fatal log level
	if a.Key == slog.LevelKey {
		level := a.Value.Any().(slog.Level)
		if level <= LevelError {
			return a
		}

		a.Value = slog.StringValue("FATAL")
	}

	return a
}

type testWriter struct {
	t *testing.T
}

func (w *testWriter) Write(p []byte) (n int, err error) {
	n = len(p)

	w.t.Logf("%s", p)

	return n, nil
}

// Logger is a wrapper for slog logger.
// The purpose is to provide a simple interface for logging with sensible defaults.
type Logger struct {
	logger *slog.Logger
}

// New creates a new Logger.
// Set name to an empty string to create an unnamed logger.
func New(logType LogType, logLevel slog.Level) *Logger {
	handlerOptions := &slog.HandlerOptions{
		// add the file and line number
		AddSource:   true,
		Level:       logLevel,
		ReplaceAttr: replaceAttrFunction,
	}

	var logger *slog.Logger
	if logType == PlainLog {
		logger = slog.New(slog.NewTextHandler(os.Stderr, handlerOptions))
	} else {
		logger = slog.New(slog.NewJSONHandler(os.Stderr, handlerOptions))
	}

	return &Logger{logger}
}

// NewTest creates a logger for unit / integration tests.
func NewTest(t *testing.T) *Logger {
	return &Logger{
		logger: slog.New(slog.NewTextHandler(&testWriter{t}, &slog.HandlerOptions{
			AddSource:   true,
			Level:       LevelDebug,
			ReplaceAttr: replaceAttrFunction,
		})),
	}
}

// Debugf logs a message at Debug level.
// Debug logs are typically voluminous, and contain detailed information on the flow of execution.
func (l *Logger) Debugf(format string, args ...any) {
	l.logger.Debug(format, args...)
}

// Infof logs a message at Info level.
// This is the default logging priority and should be used for all normal messages.
func (l *Logger) Infof(format string, args ...any) {
	l.logger.Info(format, args...)
}

// Warnf logs a message at Warn level.
// Warn logs are more important than Info, but they don't need human review or necessarily indicate an error.
func (l *Logger) Warnf(format string, args ...any) {
	l.logger.Warn(format, args...)
}

// Errorf logs a message at Error level.
// Error logs are high priority and indicate something has gone wrong.
func (l *Logger) Errorf(format string, args ...any) {
	l.logger.Error(format, args...)
}

// Fatalf logs the message and then calls os.Exit(1).
// Use this to exit your program when a fatal error occurs.
func (l *Logger) Fatalf(format string, args ...any) {
	l.logger.Log(context.Background(), LevelFatal, format, args...)
	os.Exit(1)
}

// WithIncreasedLevel returns a logger with increased logging level.
func (l *Logger) WithIncreasedLevel(level slog.Level) *Logger {
	var logger Logger
	if _, ok := l.logger.Handler().(*slog.TextHandler); ok {
		logger = *New(PlainLog, level)
	} else {
		logger = *New(JSONLog, level)
	}
	return &logger
}

// With returns a logger with structured context.
func (l *Logger) With(fields ...any) *Logger {
	return &Logger{logger: l.logger.With(fields...)}
}

// Named returns a named logger.
// TODO: rename function to Grouped()
func (l *Logger) Named(name string) *Logger {
	return &Logger{logger: l.logger.WithGroup(name)}
}

// ReplaceGRPCLogger replaces grpc's internal logger with the given logger.
func (l *Logger) ReplaceGRPCLogger() {
	replaceGRPCLogger(l.logger)
}

// GetServerUnaryInterceptor returns a gRPC server option for intercepting unary gRPC logs.
func (l *Logger) GetServerUnaryInterceptor() grpc.ServerOption {
	return grpc.UnaryInterceptor(
		logging.UnaryServerInterceptor(l.middlewareLogger()),
	)
}

// GetServerStreamInterceptor returns a gRPC server option for intercepting streaming gRPC logs.
func (l *Logger) GetServerStreamInterceptor() grpc.ServerOption {
	return grpc.StreamInterceptor(
		logging.StreamServerInterceptor(l.middlewareLogger()),
	)
}

// GetClientUnaryInterceptor returns a gRPC client option for intercepting unary gRPC logs.
func (l *Logger) GetClientUnaryInterceptor() grpc.DialOption {
	return grpc.WithUnaryInterceptor(
		logging.UnaryClientInterceptor(l.middlewareLogger()),
	)
}

// GetClientStreamInterceptor returns a gRPC client option for intercepting stream gRPC logs.
func (l *Logger) GetClientStreamInterceptor() grpc.DialOption {
	return grpc.WithStreamInterceptor(
		logging.StreamClientInterceptor(l.middlewareLogger()),
	)
}

func (l *Logger) middlewareLogger() logging.Logger {
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

		logger := l.logger

		switch lvl {
		case logging.LevelDebug:
			logger.Debug(msg)
		case logging.LevelInfo:
			logger.Info(msg)
		case logging.LevelWarn:
			logger.Warn(msg)
		case logging.LevelError:
			logger.Error(msg)
		default:
			panic(fmt.Sprintf("unknown level %v", lvl))
		}
	})
}
