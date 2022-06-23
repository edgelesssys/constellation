/*
Package logger provides logging functionality for Constellation services.
It is a thin wrapper around the zap package, providing a consistent interface for logging.
Use this package to implement logging for your Constellation services.

Usage

1. Create a logger using New().

2. Defer the Sync() method to ensure that all log entries are flushed.

3. Use the Debugf(), Infof(), Warnf(), Errorf(), and Fatalf() methods depending on the level of logging you need.

4. Use the Named() method to create a named child logger.

5. Use the With() method to create a child logger with structured context.
This can also be used to add context to a single log message:

	logger.With(zap.String("key", "value")).Infof("log message")

Log Levels

Use Debugf() to log low level and detailed information that is useful for debugging.

Use Infof() to log general information. This method is correct for most logging purposes.

Use Warnf() to log information that may indicate unwanted behavior, but is not an error.

Use Errorf() to log information about any errors that occurred.

Use Fatalf() to log information about any errors that occurred and then exit the program.
*/
package logger

import (
	"os"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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
	return &Logger{logger: l.GetZapLogger().WithOptions(zap.IncreaseLevel(level)).Sugar()}
}

// With returns a logger with structured context.
func (l *Logger) With(fields ...any) *Logger {
	return &Logger{logger: l.logger.With(fields...)}
}

// Named returns a named logger.
func (l *Logger) Named(name string) *Logger {
	return &Logger{logger: l.logger.Named(name)}
}

// GetZapLogger returns the underlying zap logger.
func (l *Logger) GetZapLogger() *zap.Logger {
	return l.logger.Desugar()
}

// GetServerUnaryInterceptor returns a gRPC server option for intercepting unary gRPC logs.
func (l *Logger) GetServerUnaryInterceptor() grpc.ServerOption {
	return grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
		grpc_ctxtags.UnaryServerInterceptor(),
		grpc_zap.UnaryServerInterceptor(l.GetZapLogger()),
	))
}

// GetServerStreamInterceptor returns a gRPC server option for intercepting streaming gRPC logs.
func (l *Logger) GetServerStreamInterceptor() grpc.ServerOption {
	return grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
		grpc_ctxtags.StreamServerInterceptor(),
		grpc_zap.StreamServerInterceptor(l.GetZapLogger()),
	))
}

// GetClientUnaryInterceptor returns a gRPC client option for intercepting unary gRPC logs.
func (l *Logger) GetClientUnaryInterceptor() grpc.DialOption {
	return grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(
		grpc_zap.UnaryClientInterceptor(l.GetZapLogger()),
	))
}

// GetClientStreamInterceptor returns a gRPC client option for intercepting stream gRPC logs.
func (l *Logger) GetClientStreamInterceptor() grpc.DialOption {
	return grpc.WithStreamInterceptor(grpc_middleware.ChainStreamClient(
		grpc_zap.StreamClientInterceptor(l.GetZapLogger()),
	))
}
