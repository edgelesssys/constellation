/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package logger

import (
	"log/slog"
)

// CmdLineVerbosityDescription explains numeric log levels.
const CmdLineVerbosityDescription = "log verbosity: Use -1 for debug information, 0 for info, 1 for warn, 2 for error"

// VerbosityFromInt converts a verbosity level from an integer to a slog.Level.
func VerbosityFromInt(verbosity int) slog.Level {
	switch {
	case verbosity <= -1:
		return slog.LevelDebug
	case verbosity == 0:
		return slog.LevelInfo
	case verbosity == 1:
		return slog.LevelWarn
	case verbosity >= 2:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
