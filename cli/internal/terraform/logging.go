/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package terraform

import (
	"fmt"
	"strings"
)

const (
	// LogLevelNone represents a log level that does not produce any output.
	LogLevelNone LogLevel = iota
	// LogLevelError enables log output at ERROR level.
	LogLevelError
	// LogLevelWarn enables log output at WARN level.
	LogLevelWarn
	// LogLevelInfo enables log output at INFO level.
	LogLevelInfo
	// LogLevelDebug enables log output at DEBUG level.
	LogLevelDebug
	// LogLevelTrace enables log output at TRACE level.
	LogLevelTrace
	// LogLevelJSON enables log output at TRACE level in JSON format.
	LogLevelJSON
)

// LogLevel is a Terraform log level.
// As per https://developer.hashicorp.com/terraform/internals/debugging
type LogLevel int

// ParseLogLevel parses a log level string into a Terraform log level.
func ParseLogLevel(level string) (LogLevel, error) {
	switch strings.ToUpper(level) {
	case "NONE":
		return LogLevelNone, nil
	case "ERROR":
		return LogLevelError, nil
	case "WARN":
		return LogLevelWarn, nil
	case "INFO":
		return LogLevelInfo, nil
	case "DEBUG":
		return LogLevelDebug, nil
	case "TRACE":
		return LogLevelTrace, nil
	case "JSON":
		return LogLevelJSON, nil
	default:
		return LogLevelNone, fmt.Errorf("invalid log level %s", level)
	}
}

// String returns the string representation of a Terraform log level.
func (l LogLevel) String() string {
	switch l {
	case LogLevelError:
		return "ERROR"
	case LogLevelWarn:
		return "WARN"
	case LogLevelInfo:
		return "INFO"
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelTrace:
		return "TRACE"
	case LogLevelJSON:
		return "JSON"
	default:
		return ""
	}
}
