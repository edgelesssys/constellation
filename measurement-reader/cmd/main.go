/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/measurement-reader/internal/sorted"
	"github.com/edgelesssys/constellation/v2/measurement-reader/internal/tdx"
	"github.com/edgelesssys/constellation/v2/measurement-reader/internal/tpm"
)

func main() {
	log := logger.NewJSONLogger(slog.LevelInfo)
	variantString := os.Getenv(constants.AttestationVariant)
	attestationVariant, err := variant.FromString(variantString)
	if err != nil {
		log.With(slog.Any("error", err)).Error("Failed to parse attestation variant")
		os.Exit(1)
	}

	var m []sorted.Measurement
	switch attestationVariant {
	case variant.AWSNitroTPM{}, variant.AWSSEVSNP{}, variant.AzureSEVSNP{}, variant.AzureTrustedLaunch{}, variant.GCPSEVES{}, variant.GCPSEVSNP{}, variant.QEMUVTPM{}:
		m, err = tpm.Measurements()
		if err != nil {
			log.With(slog.Any("error", err)).Error("Failed to read TPM measurements")
			os.Exit(1)
		}
	case variant.QEMUTDX{}:
		m, err = tdx.Measurements()
		if err != nil {
			log.With(slog.Any("error", err)).Error("Failed to read Intel TDX measurements")
			os.Exit(1)
		}
	default:
		log.With(slog.String("attestationVariant", variantString)).Error("Unsupported attestation variant")
		os.Exit(1)
	}

	fmt.Println("Measurements:")
	for _, measurement := range m {
		// -7 should ensure consistent padding across all current prefixes: PCR[xx], MRTD, RTMR[x].
		// If the prefix gets longer somewhen in the future, this might need adjustment for consistent padding.
		fmt.Printf("\t%-7s : 0x%0X\n", measurement.Index, measurement.Value)
	}
}
