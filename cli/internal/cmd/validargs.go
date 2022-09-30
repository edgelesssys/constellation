/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/spf13/cobra"
)

// warnAWS warns that AWS isn't supported.
func warnAWS(providerPos int) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if cloudprovider.FromString(args[providerPos]) == cloudprovider.AWS && os.Getenv("CONSTELLATION_AWS_DEV") != "1" {
			return errors.New("AWS is not supported yet")
		}
		return nil
	}
}

func isCloudProvider(arg int) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if provider := cloudprovider.FromString(args[arg]); provider == cloudprovider.Unknown {
			return fmt.Errorf("argument %s isn't a valid cloud provider", args[arg])
		}
		return nil
	}
}
