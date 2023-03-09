/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package main

import (
	"fmt"
	"os"

	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/oid"
	"github.com/edgelesssys/constellation/v2/measurement-reader/internal/sorted"
	"github.com/edgelesssys/constellation/v2/measurement-reader/internal/tdx"
	"github.com/edgelesssys/constellation/v2/measurement-reader/internal/tpm"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	log := logger.New(logger.JSONLog, zapcore.InfoLevel)
	variant := os.Getenv(constants.AttestationVariant)
	attestationVariant, err := oid.FromString(variant)
	if err != nil {
		log.With(zap.Error(err)).Fatalf("Failed to parse attestation variant")
	}

	var m []sorted.Measurement
	switch attestationVariant {
	case oid.AWSNitroTPM{}, oid.AzureSEVSNP{}, oid.AzureTrustedLaunch{}, oid.GCPSEVES{}, oid.QEMUVTPM{}:
		m, err = tpm.Measurements()
		if err != nil {
			log.With(zap.Error(err)).Fatalf("Failed to read TPM measurements")
		}
	case oid.QEMUTDX{}:
		m, err = tdx.Measurements()
		if err != nil {
			log.With(zap.Error(err)).Fatalf("Failed to read Intel TDX measurements")
		}
	default:
		log.With(zap.String("attestationVariant", variant)).Fatalf("Unsupported attestation variant")
	}

	fmt.Println("Measurements:")
	for _, measurement := range m {
		// -7 should ensure consistent padding across all current prefixes: PCR[xx], MRTD, RTMR[x].
		// If the prefix gets longer somewhen in the future, this might need adjustment for consistent padding.
		fmt.Printf("\t%-7s : 0x%0X\n", measurement.Index, measurement.Value)
	}
}
