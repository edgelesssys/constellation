/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/spf13/cobra"
)

// newMeasurementsMergeCmd creates a new merge command.
func newMeasurementsMergeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "merge [flags] <measurements.json>...",
		Short: "Merge OS image measurements",
		Long:  "Merge OS image measurements.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  runMergeMeasurements,
	}

	cmd.SetOut(os.Stdout)
	cmd.Flags().String("out", "", "Optional path to write the merge result to. If not set, the result is written to stdout.")
	cmd.Flags().Bool("verbose", false, "Enable verbose output")

	return cmd
}

func runMergeMeasurements(cmd *cobra.Command, args []string) error {
	workdir := os.Getenv("BUILD_WORKING_DIRECTORY")
	if len(workdir) > 0 {
		must(os.Chdir(workdir))
	}

	flags, err := parseMergeMeasurementsFlags(cmd)
	if err != nil {
		return err
	}

	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: flags.logLevel}))
	log.Debug("Parsed flags: %+v", flags)

	mergedMeasurements, err := readMeasurementsArgs(args)
	if err != nil {
		return fmt.Errorf("merging measurements: reading input files: %w", err)
	}

	out := cmd.OutOrStdout()
	if len(flags.out) > 0 {
		outF, err := os.Create(flags.out)
		if err != nil {
			return fmt.Errorf("merging measurements: opening output file: %w", err)
		}
		defer outF.Close()
		out = outF
	}

	if err := json.NewEncoder(out).Encode(mergedMeasurements); err != nil {
		return fmt.Errorf("merging measurements: writing output file: %w", err)
	}
	log.Info("Merged image measurements")
	return nil
}

func readMeasurementsArgs(paths []string) (measurements.ImageMeasurementsV2, error) {
	measuremnts := make([]measurements.ImageMeasurementsV2, len(paths))
	for i, path := range paths {
		f, err := os.Open(path)
		if err != nil {
			return measurements.ImageMeasurementsV2{}, err
		}
		defer f.Close()
		if err := json.NewDecoder(f).Decode(&measuremnts[i]); err != nil {
			return measurements.ImageMeasurementsV2{}, err
		}
	}
	return measurements.MergeImageMeasurementsV2(measuremnts...)
}
