package cmd

import (
	"errors"
	"fmt"

	"github.com/edgelesssys/constellation/internal/cloud/cloudprovider"
	"github.com/spf13/cobra"
)

// warnAWS warns that AWS isn't supported.
func warnAWS(providerPos int) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if cloudprovider.FromString(args[providerPos]) == cloudprovider.AWS {
			return errors.New("AWS isn't supported by this version of Constellation")
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
