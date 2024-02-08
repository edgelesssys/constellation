/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/spf13/cobra"
)

// newMeasurementsEnvelopeCmd creates a new envelope command.
func newMeasurementsEnvelopeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "envelope",
		Short: "Envelope OS image measurements",
		Long:  "Envelope OS image measurements for one variant to follow the measurements v2 format.",
		Args:  cobra.ExactArgs(0),
		RunE:  runEnvelopeMeasurements,
	}

	cmd.SetOut(os.Stdout)
	cmd.Flags().String("version", "", "Shortname of the os image version.")
	cmd.Flags().String("csp", "", "CSP of this image measurement.")
	cmd.Flags().String("attestation-variant", "", "Attestation variant of the image measurements.")
	cmd.Flags().String("in", "", "Path to read the raw measurements from.")
	cmd.Flags().String("out", "", "Optional path to write the enveloped result to. If not set, the result is written to stdout.")
	cmd.Flags().Bool("verbose", false, "Enable verbose output")

	must(cmd.MarkFlagRequired("version"))
	must(cmd.MarkFlagRequired("csp"))
	must(cmd.MarkFlagRequired("attestation-variant"))
	must(cmd.MarkFlagRequired("in"))

	return cmd
}

func runEnvelopeMeasurements(cmd *cobra.Command, _ []string) error {
	workdir := os.Getenv("BUILD_WORKING_DIRECTORY")
	if len(workdir) > 0 {
		must(os.Chdir(workdir))
	}

	flags, err := parseEnvelopeMeasurementsFlags(cmd)
	if err != nil {
		return err
	}

	log := logger.NewTextLogger(flags.logLevel)
	log.Debug(fmt.Sprintf("Parsed flags: %+v", flags))

	f, err := os.Open(flags.in)
	if err != nil {
		return fmt.Errorf("enveloping measurements: opening input file: %w", err)
	}
	defer f.Close()
	var measuremnt rawMeasurements
	if err := json.NewDecoder(f).Decode(&measuremnt); err != nil {
		return fmt.Errorf("enveloping measurements: reading input file: %w", err)
	}

	measuremnt.Measurements, err = measurements.ApplyOverrides(measuremnt.Measurements, flags.csp, flags.attestationVariant)
	if err != nil {
		return fmt.Errorf("enveloping measurements: overriding static measurements: %w", err)
	}

	enveloped := measurements.ImageMeasurementsV2{
		Ref:     flags.version.Ref(),
		Stream:  flags.version.Stream(),
		Version: flags.version.Version(),
		List: []measurements.ImageMeasurementsV2Entry{
			{
				CSP:                flags.csp,
				AttestationVariant: flags.attestationVariant,
				Measurements:       measuremnt.Measurements,
			},
		},
	}

	out := cmd.OutOrStdout()
	if len(flags.out) > 0 {
		outF, err := os.Create(flags.out)
		if err != nil {
			return fmt.Errorf("enveloping measurements: opening output file: %w", err)
		}
		defer outF.Close()
		out = outF
	}

	if err := json.NewEncoder(out).Encode(enveloped); err != nil {
		return fmt.Errorf("enveloping measurements: writing output file: %w", err)
	}
	log.Info("Enveloped image measurements")
	return nil
}

type rawMeasurements struct {
	Measurements measurements.M `json:"measurements"`
}
