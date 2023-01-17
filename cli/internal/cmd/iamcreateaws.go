/*
Copyright (c) Edgeless Systems GmbH
SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"fmt"
	"strings"

	"github.com/edgelesssys/constellation/v2/cli/internal/cloudcmd"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

// newIAMCreateAWSCmd returns a new cobra.Command for the iam create aws command.
func newIAMCreateAWSCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "aws",
		Short: "Create IAM configuration on AWS for your Constellation cluster",
		Long:  "Create IAM configuration on AWS for your Constellation cluster.",
		Args:  cobra.ExactArgs(0),
		RunE:  runIAMCreateAWS,
	}

	cmd.Flags().String("prefix", "", "name prefix for all resources")
	must(cobra.MarkFlagRequired(cmd.Flags(), "prefix"))
	cmd.Flags().String("zone", "", "AWS availability zone the resources will be created in (e.g. us-east-2a)\n"+
		"Find available zones here: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-regions-availability-zones.html#concepts-availability-zones. Note that we do not support every zone / region. You can find a list of all supported regions in our docs")
	must(cobra.MarkFlagRequired(cmd.Flags(), "zone"))
	cmd.Flags().Bool("yes", false, "create the IAM configuration without further confirmation")

	return cmd
}

func runIAMCreateAWS(cmd *cobra.Command, args []string) error {
	spinner := newSpinner(cmd.ErrOrStderr())
	defer spinner.Stop()
	fileHandler := file.NewHandler(afero.NewOsFs())
	creator := cloudcmd.NewIAMCreator(spinner)

	return iamCreateAWS(cmd, spinner, creator, fileHandler)
}

func iamCreateAWS(cmd *cobra.Command, spinner spinnerInterf, creator iamCreator, fileHandler file.Handler) error {
	// Get input variables.
	awsFlags, err := parseAWSFlags(cmd)
	if err != nil {
		return err
	}

	// Confirmation.
	if !awsFlags.yesFlag {
		cmd.Printf("The following IAM configuration will be created:\n\n")
		cmd.Printf("Region:\t\t%s\n", awsFlags.region)
		cmd.Printf("Name Prefix:\t%s\n\n", awsFlags.prefix)
		if awsFlags.generateConfig {
			cmd.Printf("The configuration file %s will be automatically generated and populated with the IAM values.\n", awsFlags.configPath)
		}
		ok, err := askToConfirm(cmd, "Do you want to create the configuration?")
		if err != nil {
			return err
		}
		if !ok {
			cmd.Println("The creation of the configuration was aborted.")
			return nil
		}
	}

	// Creation.
	spinner.Start("Creating", false)

	conf := createConfig(cloudprovider.AWS)

	iamFile, err := creator.Create(cmd.Context(), cloudprovider.AWS, &cloudcmd.IAMConfig{
		AWS: cloudcmd.AWSIAMConfig{
			Region: awsFlags.region,
			Prefix: awsFlags.prefix,
		},
	})

	spinner.Stop()
	if err != nil {
		return err
	}
	cmd.Println() // Print empty line to separate after spinner ended.

	if awsFlags.generateConfig {
		conf.Provider.AWS.Region = awsFlags.region
		conf.Provider.AWS.Zone = awsFlags.zone
		conf.Provider.AWS.IAMProfileControlPlane = iamFile.AWSOutput.ControlPlaneInstanceProfile
		conf.Provider.AWS.IAMProfileWorkerNodes = iamFile.AWSOutput.WorkerNodeInstanceProfile

		if err := fileHandler.WriteYAML(awsFlags.configPath, conf, file.OptMkdirAll); err != nil {
			return err
		}
		cmd.Printf("Your IAM configuration was created and filled into %s successfully.\n", awsFlags.configPath)
		return nil
	}

	cmd.Printf("region:\t\t\t%s\n", awsFlags.region)
	cmd.Printf("zone:\t\t\t%s\n", awsFlags.zone)
	cmd.Printf("iamProfileControlPlane:\t%s\n", iamFile.AWSOutput.ControlPlaneInstanceProfile)
	cmd.Printf("iamProfileWorkerNodes:\t%s\n\n", iamFile.AWSOutput.WorkerNodeInstanceProfile)
	cmd.Println("Your IAM configuration was created successfully. Please fill the above values into your configuration file.")

	return nil
}

// parseAWSFlags parses and validates the flags of the iam create aws command.
func parseAWSFlags(cmd *cobra.Command) (awsFlags, error) {
	var region string

	prefix, err := cmd.Flags().GetString("prefix")
	if err != nil {
		return awsFlags{}, fmt.Errorf("parsing prefix string: %w", err)
	}
	zone, err := cmd.Flags().GetString("zone")
	if err != nil {
		return awsFlags{}, fmt.Errorf("parsing zone string: %w", err)
	}
	if strings.HasPrefix(zone, "eu-central-1") {
		region = "eu-central-1"
	} else if strings.HasPrefix(zone, "us-east-2") {
		region = "us-east-2"
	} else if strings.HasPrefix(zone, "ap-south-1") {
		region = "ap-south-1"
	} else {
		return awsFlags{}, fmt.Errorf("invalid AWS region, to find a correct region please refer to our docs and https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-regions-availability-zones.html#concepts-availability-zones")
	}
	configPath, err := cmd.Flags().GetString("config")
	if err != nil {
		return awsFlags{}, fmt.Errorf("parsing config string: %w", err)
	}
	generateConfig, err := cmd.Flags().GetBool("generate-config")
	if err != nil {
		return awsFlags{}, fmt.Errorf("parsing generate-config bool: %w", err)
	}
	yesFlag, err := cmd.Flags().GetBool("yes")
	if err != nil {
		return awsFlags{}, fmt.Errorf("parsing yes bool: %w", err)
	}
	return awsFlags{
		zone:           zone,
		prefix:         prefix,
		region:         region,
		generateConfig: generateConfig,
		configPath:     configPath,
		yesFlag:        yesFlag,
	}, nil
}

// awsFlags contains the parsed flags of the iam create aws command.
type awsFlags struct {
	prefix         string
	region         string
	zone           string
	generateConfig bool
	configPath     string
	yesFlag        bool
}
