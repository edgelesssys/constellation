/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/spf13/cobra"
)

func arg0isAttestationVariant() cobra.PositionalArgs {
	return func(_ *cobra.Command, args []string) error {
		attestationVariant, err := variant.FromString(args[0])
		if err != nil {
			return errors.New("argument 0 isn't a valid attestation variant")
		}
		switch attestationVariant {
		case variant.AWSSEVSNP{}, variant.AzureSEVSNP{}, variant.AzureTDX{}, variant.GCPSEVSNP{}:
			return nil
		default:
			return errors.New("argument 0 isn't a supported attestation variant")
		}
	}
}

func isValidKind(arg int) cobra.PositionalArgs {
	return func(_ *cobra.Command, args []string) error {
		if kind := kindFromString(args[arg]); kind == unknown {
			return fmt.Errorf("argument %s isn't a valid kind", args[arg])
		}
		return nil
	}
}

// objectKind encodes the available actions.
type objectKind string

const (
	// unknown is the default objectKind and does nothing.
	unknown           objectKind = "unknown-kind"
	attestationReport objectKind = "attestation-report"
	guestFirmware     objectKind = "guest-firmware"
)

func kindFromString(s string) objectKind {
	lower := strings.ToLower(s)
	switch objectKind(lower) {
	case attestationReport, guestFirmware:
		return objectKind(lower)
	default:
		return unknown
	}
}
