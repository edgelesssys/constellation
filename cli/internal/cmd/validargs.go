package cmd

import (
	"errors"
	"fmt"

	"github.com/edgelesssys/constellation/cli/internal/azure"
	"github.com/edgelesssys/constellation/cli/internal/gcp"
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

func validInstanceTypeForProvider(cmd *cobra.Command, insType string, provider cloudprovider.Provider) error {
	switch provider {
	case cloudprovider.GCP:
		for _, instanceType := range gcp.InstanceTypes {
			if insType == instanceType {
				return nil
			}
		}
		cmd.SetUsageTemplate("GCP instance types:\n" + formatInstanceTypes(gcp.InstanceTypes))
		cmd.SilenceUsage = false
		return fmt.Errorf("%s isn't a valid GCP instance type", insType)
	case cloudprovider.Azure:
		for _, instanceType := range azure.CVMInstanceTypes {
			if insType == instanceType {
				return nil
			}
		}
		for _, instanceType := range azure.TrustedLaunchInstanceTypes {
			if insType == instanceType {
				return nil
			}
		}
		cmd.SetUsageTemplate("Azure CVM instance types:\n" + formatInstanceTypes(azure.CVMInstanceTypes) +
			"\n\nAzure Trusted Launch instance types:\n" + formatInstanceTypes(azure.TrustedLaunchInstanceTypes))
		cmd.SilenceUsage = false
		return fmt.Errorf("%s isn't a valid Azure instance type", insType)
	default:
		return fmt.Errorf("%s isn't a valid cloud platform", provider)
	}
}
