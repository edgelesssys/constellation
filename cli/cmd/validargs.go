package cmd

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/edgelesssys/constellation/cli/azure"
	"github.com/edgelesssys/constellation/cli/cloudprovider"
	"github.com/edgelesssys/constellation/cli/ec2"
	"github.com/edgelesssys/constellation/cli/gcp"
	"github.com/spf13/cobra"
)

// isIntArg checks if argument at position arg is an integer.
func isIntArg(arg int) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if _, err := strconv.Atoi(args[arg]); err != nil {
			return fmt.Errorf("argument %d must be an integer", arg)
		}
		return nil
	}
}

// isIntGreaterArg checks if argument at position arg is an integer and greater i.
func isIntGreaterArg(arg int, i int) cobra.PositionalArgs {
	return cobra.MatchAll(isIntArg(arg), func(cmd *cobra.Command, args []string) error {
		if v, _ := strconv.Atoi(args[arg]); v <= i {
			return fmt.Errorf("argument %d must be greater %d, but it's %d", arg, i, v)
		}
		return nil
	})
}

func isIntLessArg(arg int, i int) cobra.PositionalArgs {
	return cobra.MatchAll(isIntArg(arg), func(cmd *cobra.Command, args []string) error {
		if v, _ := strconv.Atoi(args[arg]); v >= i {
			return fmt.Errorf("argument %d must be less %d, but it's %d", arg, i, v)
		}
		return nil
	})
}

// warnAWS warns that AWS isn't supported.
func warnAWS(providerPos int) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if cloudprovider.FromString(args[providerPos]) == cloudprovider.AWS {
			return errors.New("AWS isn't supported")
		}
		return nil
	}
}

// isIntGreaterZeroArg checks if argument at position arg is a positive non zero integer.
func isIntGreaterZeroArg(arg int) cobra.PositionalArgs {
	return isIntGreaterArg(arg, 0)
}

func isPort(arg int) cobra.PositionalArgs {
	return cobra.MatchAll(
		isIntGreaterArg(arg, -1),
		isIntLessArg(arg, 65536),
	)
}

func isIP(arg int) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if ip := net.ParseIP(args[arg]); ip == nil {
			return fmt.Errorf("argument %s isn't a valid IP address", args[arg])
		}
		return nil
	}
}

// isEC2InstanceType checks if argument at position arg is a key in m.
// The argument will always be converted to lower case letters.
func isEC2InstanceType(arg int) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if _, ok := ec2.InstanceTypes[strings.ToLower(args[arg])]; !ok {
			return fmt.Errorf("'%s' isn't an AWS EC2 instance type", args[arg])
		}
		return nil
	}
}

func isGCPInstanceType(arg int) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		for _, instanceType := range gcp.InstanceTypes {
			if args[arg] == instanceType {
				return nil
			}
		}
		return fmt.Errorf("argument %s isn't a valid GCP instance type", args[arg])
	}
}

func isAzureInstanceType(arg int) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		for _, instanceType := range azure.InstanceTypes {
			if args[arg] == instanceType {
				return nil
			}
		}
		return fmt.Errorf("argument %s isn't a valid Azure instance type", args[arg])
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

// isInstanceTypeForProvider returns a argument validation function that checks if the argument
// at position typePos is a valid instance type for the cloud provider string at position
// providerPos.
func isInstanceTypeForProvider(typePos, providerPos int) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return fmt.Errorf("requires 2 arguments, but only %d are provided", len(args))
		}
		if len(args) <= typePos {
			return fmt.Errorf(
				"%d arguments provided, but index %d of typePos is out of bound",
				len(args), typePos,
			)
		}
		if len(args) <= providerPos {
			return fmt.Errorf(
				"%d arguments provided, but index %d of providerPos is out of bound",
				len(args), providerPos,
			)
		}

		switch cloudprovider.FromString(args[providerPos]) {
		case cloudprovider.GCP:
			return isGCPInstanceType(typePos)(cmd, args)
		case cloudprovider.Azure:
			return isAzureInstanceType(typePos)(cmd, args)
		default:
			return fmt.Errorf("argument %s isn't a valid cloud platform", args[providerPos])
		}
	}
}
