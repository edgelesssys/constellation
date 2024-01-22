/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

type debugLog interface {
	Debug(format string, args ...any)
}

func newCLILogger(cmd *cobra.Command) (debugLog, error) {
	logLvl := slog.LevelInfo
	debugLog, err := cmd.Flags().GetBool("debug")
	if err != nil {
		return nil, err
	}
	if debugLog {
		logLvl = slog.LevelDebug
	}

	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: logLvl})), nil
}
