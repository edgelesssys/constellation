/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// NewMeasurementsCmd creates a new measurements command. Measurements needs another
// verb, and does nothing on its own.
func NewMeasurementsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "measurements",
		Short: "Handle OS image measurements",
		Long:  "Handle OS image measurements.",
		Args:  cobra.ExactArgs(0),
	}

	cmd.SetOut(os.Stdout)

	cmd.AddCommand(newMeasurementsUploadCmd())
	cmd.AddCommand(newMeasurementsMergeCmd())
	cmd.AddCommand(newMeasurementsEnvelopeCmd())

	return cmd
}
