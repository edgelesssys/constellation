/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/spf13/cobra"
	"go.uber.org/zap/zapcore"
)

type DebugLog interface {
	Debugf(format string, args ...any)
	Sync()
}

func NewCLILogger(cmd *cobra.Command) (DebugLog, error) {
	logLvl := zapcore.InfoLevel
	debugLog, err := cmd.Flags().GetBool("debug")
	if err != nil {
		return nil, err
	}
	if debugLog {
		logLvl = zapcore.DebugLevel
	}

	return logger.New(logger.PlainLog, logLvl), nil
}
