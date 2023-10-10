/*
Copyright (c) Edgeless Systems GmbH
SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"errors"
	"fmt"

	"github.com/edgelesssys/constellation/v2/cli/internal/cloudcmd"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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

	cmd.Flags().String("prefix", "", "name prefix for all resources (required)")
	must(cobra.MarkFlagRequired(cmd.Flags(), "prefix"))
	cmd.Flags().String("zone", "", "AWS availability zone the resources will be created in, e.g., us-east-2a (required)\n"+
		"See the Constellation docs for a list of currently supported regions.")
	must(cobra.MarkFlagRequired(cmd.Flags(), "zone"))
	return cmd
}

func runIAMCreateAWS(cmd *cobra.Command, _ []string) error {
	creator := &awsIAMCreator{}
	if err := creator.flags.parse(cmd.Flags()); err != nil {
		return err
	}
	return runIAMCreate(cmd, creator, cloudprovider.AWS)
}

// awsIAMCreateFlags contains the parsed flags of the iam create aws command.
type awsIAMCreateFlags struct {
	prefix string
	region string
	zone   string
}

func (f *awsIAMCreateFlags) parse(flags *pflag.FlagSet) error {
	var err error
	f.prefix, err = flags.GetString("prefix")
	if err != nil {
		return fmt.Errorf("getting 'prefix' flag: %w", err)
	}
	if len(f.prefix) > 36 {
		return errors.New("prefix must be 36 characters or less")
	}
	f.zone, err = flags.GetString("zone")
	if err != nil {
		return fmt.Errorf("getting 'zone' flag: %w", err)
	}
	if !config.ValidateAWSZone(f.zone) {
		return errors.New("invalid AWS zone. To find a valid zone, please refer to our docs and https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-regions-availability-zones.html#concepts-availability-zones")
	}
	// Infer region from zone.
	f.region = f.zone[:len(f.zone)-1]
	if !config.ValidateAWSRegion(f.region) {
		return fmt.Errorf("invalid AWS region: %s", f.region)
	}

	return nil
}

// awsIAMCreator implements the providerIAMCreator interface for AWS.
type awsIAMCreator struct {
	flags awsIAMCreateFlags
}

func (c *awsIAMCreator) getIAMConfigOptions() *cloudcmd.IAMConfigOptions {
	return &cloudcmd.IAMConfigOptions{
		AWS: cloudcmd.AWSIAMConfig{
			Region: c.flags.region,
			Prefix: c.flags.prefix,
		},
	}
}

func (c *awsIAMCreator) printConfirmValues(cmd *cobra.Command) {
	cmd.Printf("Region:\t\t%s\n", c.flags.region)
	cmd.Printf("Name Prefix:\t%s\n\n", c.flags.prefix)
}

func (c *awsIAMCreator) printOutputValues(cmd *cobra.Command, iamFile cloudcmd.IAMOutput) {
	cmd.Printf("region:\t\t\t%s\n", c.flags.region)
	cmd.Printf("zone:\t\t\t%s\n", c.flags.zone)
	cmd.Printf("iamProfileControlPlane:\t%s\n", iamFile.AWSOutput.ControlPlaneInstanceProfile)
	cmd.Printf("iamProfileWorkerNodes:\t%s\n\n", iamFile.AWSOutput.WorkerNodeInstanceProfile)
}

func (c *awsIAMCreator) writeOutputValuesToConfig(conf *config.Config, iamFile cloudcmd.IAMOutput) {
	conf.Provider.AWS.Region = c.flags.region
	conf.Provider.AWS.Zone = c.flags.zone
	conf.Provider.AWS.IAMProfileControlPlane = iamFile.AWSOutput.ControlPlaneInstanceProfile
	conf.Provider.AWS.IAMProfileWorkerNodes = iamFile.AWSOutput.WorkerNodeInstanceProfile
	for groupName, group := range conf.NodeGroups {
		group.Zone = c.flags.zone
		conf.NodeGroups[groupName] = group
	}
}

func (c *awsIAMCreator) parseAndWriteIDFile(_ cloudcmd.IAMOutput, _ file.Handler) error {
	return nil
}

func (c *awsIAMCreator) validateConfigWithFlagCompatibility(conf config.Config) error {
	return validateConfigWithFlagCompatibility(cloudprovider.AWS, conf, c.flags.zone)
}
