/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// CmdLineVerbosityDescription explains numeric log levels.
const CmdLineVerbosityDescription = "log verbosity in zap logging levels. Use -1 for debug information, 0 for info, 1 for warn, 2 for error"

// VerbosityFromInt converts a verbosity level from an integer to a zapcore.Level.
func VerbosityFromInt(verbosity int) zapcore.Level {
	switch {
	case verbosity <= -1:
		return zap.DebugLevel
	case verbosity == 0:
		return zap.InfoLevel
	case verbosity == 1:
		return zap.WarnLevel
	case verbosity >= 2:
		return zap.ErrorLevel
	default:
		return zap.InfoLevel
	}
}
