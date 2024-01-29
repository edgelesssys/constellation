/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package logger

import (
	"log/slog"
)

// CmdLineVerbosityDescription explains numeric log levels.
const CmdLineVerbosityDescription = "log verbosity in slog logging levels. Use -4 for debug information, 0 for info, 4 for warn, 8 for error"

// VerbosityFromInt converts a verbosity level from an integer to a slog.Level.
func VerbosityFromInt(verbosity int) slog.Level {
	switch {
	case verbosity <= -4:
		return slog.LevelDebug
	case verbosity == 0:
		return slog.LevelInfo
	case verbosity == 4:
		return slog.LevelWarn
	case verbosity >= 8:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
