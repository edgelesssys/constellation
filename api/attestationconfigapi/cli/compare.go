/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/
package main

import (
	"fmt"
	"os"
	"slices"

	"github.com/edgelesssys/constellation/v2/api/attestationconfigapi/cli/client"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/verify"
	"github.com/google/go-tdx-guest/proto/tdx"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func newCompareCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "compare VARIANT FILE [FILE...]",
		Short:   "Returns the minimum version of all given attestation reports.",
		Long:    "Compare a list of attestation reports and return the report with the minimum version.",
		Example: "cli compare azure-sev-snp report1.json report2.json",
		Args:    cobra.MatchAll(cobra.MinimumNArgs(2), arg0isAttestationVariant()),
		RunE:    runCompare,
	}

	return cmd
}

func runCompare(cmd *cobra.Command, args []string) error {
	cmd.SetOut(os.Stdout)

	variant, err := variant.FromString(args[0])
	if err != nil {
		return fmt.Errorf("parsing variant: %w", err)
	}

	return compare(cmd, variant, args[1:], file.NewHandler(afero.NewOsFs()))
}

func compare(cmd *cobra.Command, attestationVariant variant.Variant, files []string, fs file.Handler) (retErr error) {
	if !slices.Contains([]variant.Variant{variant.AWSSEVSNP{}, variant.AzureSEVSNP{}, variant.GCPSEVSNP{}, variant.AzureTDX{}}, attestationVariant) {
		return fmt.Errorf("variant %s not supported", attestationVariant)
	}

	lowestVersion, err := compareVersions(attestationVariant, files, fs)
	if err != nil {
		return fmt.Errorf("comparing versions: %w", err)
	}

	cmd.Println(lowestVersion)
	return nil
}

func compareVersions(attestationVariant variant.Variant, files []string, fs file.Handler) (string, error) {
	readReport := readSNPReport
	if attestationVariant.Equal(variant.AzureTDX{}) {
		readReport = readTDXReport
	}

	lowestVersion := files[0]
	lowestReport, err := readReport(files[0], fs)
	if err != nil {
		return "", fmt.Errorf("reading report: %w", err)
	}

	for _, file := range files[1:] {
		report, err := readReport(file, fs)
		if err != nil {
			return "", fmt.Errorf("reading report: %w", err)
		}

		if client.IsInputNewerThanOtherVersion(attestationVariant, lowestReport, report) {
			lowestVersion = file
			lowestReport = report
		}
	}

	return lowestVersion, nil
}

func readSNPReport(file string, fs file.Handler) (any, error) {
	var report verify.Report
	if err := fs.ReadJSON(file, &report); err != nil {
		return nil, fmt.Errorf("reading snp report: %w", err)
	}
	return convertTCBVersionToSNPVersion(report.SNPReport.LaunchTCB), nil
}

func readTDXReport(file string, fs file.Handler) (any, error) {
	var report *tdx.QuoteV4
	if err := fs.ReadJSON(file, &report); err != nil {
		return nil, fmt.Errorf("reading tdx report: %w", err)
	}
	return convertQuoteToTDXVersion(report), nil
}
