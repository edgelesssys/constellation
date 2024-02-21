/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package main

import (
	"fmt"
	"strings"

	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/spf13/cobra"
)

func isCloudProvider(arg int) cobra.PositionalArgs {
	return func(_ *cobra.Command, args []string) error {
		if provider := cloudprovider.FromString(args[arg]); provider == cloudprovider.Unknown {
			return fmt.Errorf("argument %s isn't a valid cloud provider", args[arg])
		}
		return nil
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
	unknown       objectKind = "unknown-kind"
	snpReport     objectKind = "snp-report"
	guestFirmware objectKind = "guest-firmware"
)

func kindFromString(s string) objectKind {
	lower := strings.ToLower(s)
	switch objectKind(lower) {
	case snpReport, guestFirmware:
		return objectKind(lower)
	default:
		return unknown
	}
}
