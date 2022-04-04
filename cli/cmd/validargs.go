package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/edgelesssys/constellation/cli/azure"
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

// isValidAWSCoordinatorCount checks if argument at position arg is an integer exactly 1.
func isValidAWSCoordinatorCount(arg int) cobra.PositionalArgs {
	return cobra.MatchAll(isIntArg(arg), func(cmd *cobra.Command, args []string) error {
		if v, _ := strconv.Atoi(args[arg]); v != 1 {
			return fmt.Errorf("argument %d is %d, that is not a valid coordinator count for AWS, currently the only supported coordinator count is 1", arg, v)
		}
		return nil
	})
}

// isIntGreaterZeroArg checks if argument at position arg is a positive non zero integer.
func isIntGreaterZeroArg(arg int) cobra.PositionalArgs {
	return cobra.MatchAll(isIntGreaterArg(arg, 0))
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
