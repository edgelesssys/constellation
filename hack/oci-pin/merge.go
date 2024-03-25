/*
Copyright (c) Edgeless Systems GmbH
SPDX-License-Identifier: AGPL-3.0-only
*/

package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/edgelesssys/constellation/v2/hack/oci-pin/internal/sums"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/spf13/cobra"
)

func newMergeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "merge",
		Short: "Merge multiple sha256sum files that pin OCI images.",
		RunE:  runMerge,
	}

	cmd.Flags().StringArray("input", nil, "Path to existing sha256sum file that should be merged.")
	cmd.Flags().String("output", "-", "Output file. If not set, the output is written to stdout.")
	must(cmd.MarkFlagRequired("input"))

	return cmd
}

func runMerge(cmd *cobra.Command, _ []string) error {
	flags, err := parseMergeFlags(cmd)
	if err != nil {
		return err
	}
	log := logger.NewTextLogger(flags.logLevel)
	log.Debug("Using flags", "inputs", flags.inputs, "output", flags.output, "logLevel", flags.logLevel)

	log.Debug(fmt.Sprintf("Merging sum file from %q into %q.", flags.inputs, flags.output))

	var out io.Writer
	if flags.output == "-" {
		out = cmd.OutOrStdout()
	} else {
		f, err := os.Create(flags.output)
		if err != nil {
			return fmt.Errorf("creating output file %q: %w", flags.output, err)
		}
		defer f.Close()
		out = f
	}

	unmergedRefs, err := parseInputs(flags.inputs)
	if err != nil {
		return fmt.Errorf("reading input files: %w", err)
	}

	if err := sums.Merge(unmergedRefs, out); err != nil {
		return fmt.Errorf("creating merged sum file: %w", err)
	}

	log.Debug(fmt.Sprintf("Sum file created at %q 🤖", flags.output))
	return nil
}

func parseInputs(inputs []string) ([][]sums.PinnedImageReference, error) {
	var unmergedRefs [][]sums.PinnedImageReference
	for _, input := range inputs {
		refs, err := parseInput(input)
		if err != nil {
			return nil, err
		}
		unmergedRefs = append(unmergedRefs, refs)
	}
	return unmergedRefs, nil
}

func parseInput(input string) ([]sums.PinnedImageReference, error) {
	in, err := os.Open(input)
	if err != nil {
		return nil, fmt.Errorf("opening sum file at %q: %w", input, err)
	}
	defer in.Close()
	refs, err := sums.Parse(in)
	if err != nil {
		return nil, fmt.Errorf("parsing sums %q: %w", input, err)
	}
	return refs, nil
}

type mergeFlags struct {
	inputs   []string
	output   string
	logLevel slog.Level
}

func parseMergeFlags(cmd *cobra.Command) (mergeFlags, error) {
	inputs, err := cmd.Flags().GetStringArray("input")
	if err != nil {
		return mergeFlags{}, err
	}
	output, err := cmd.Flags().GetString("output")
	if err != nil {
		return mergeFlags{}, err
	}
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return mergeFlags{}, err
	}
	logLevel := slog.LevelInfo
	if verbose {
		logLevel = slog.LevelDebug
	}

	return mergeFlags{
		inputs:   inputs,
		output:   output,
		logLevel: logLevel,
	}, nil
}
