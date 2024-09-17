/*
Copyright (c) Edgeless Systems GmbH
SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/edgelesssys/constellation/v2/cli/internal/cloudcmd"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// newIAMCreateAzureCmd returns a new cobra.Command for the iam create azure command.
func newIAMCreateAzureCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "azure",
		Short: "Create IAM configuration on Microsoft Azure for your Constellation cluster",
		Long:  "Create IAM configuration on Microsoft Azure for your Constellation cluster.",
		Args:  cobra.ExactArgs(0),
		RunE:  runIAMCreateAzure,
	}

	cmd.Flags().String("subscriptionID", "", "subscription ID of the Azure account. Required if the 'ARM_SUBSCRIPTION_ID' environment variable is not set")
	cmd.Flags().String("resourceGroup", "", "name prefix of the two resource groups your cluster / IAM resources will be created in (required)")
	must(cobra.MarkFlagRequired(cmd.Flags(), "resourceGroup"))
	cmd.Flags().String("region", "", "region the resources will be created in, e.g., westus (required)")
	must(cobra.MarkFlagRequired(cmd.Flags(), "region"))
	cmd.Flags().String("servicePrincipal", "", "name of the service principal that will be created (required)")
	must(cobra.MarkFlagRequired(cmd.Flags(), "servicePrincipal"))
	return cmd
}

func runIAMCreateAzure(cmd *cobra.Command, _ []string) error {
	creator := &azureIAMCreator{}
	if err := creator.flags.parse(cmd.Flags()); err != nil {
		return err
	}
	return runIAMCreate(cmd, creator, cloudprovider.Azure)
}

// azureIAMCreateFlags contains the parsed flags of the iam create azure command.
type azureIAMCreateFlags struct {
	subscriptionID   string
	region           string
	resourceGroup    string
	servicePrincipal string
}

func (f *azureIAMCreateFlags) parse(flags *pflag.FlagSet) error {
	var err error
	f.subscriptionID, err = flags.GetString("subscriptionID")
	if err != nil {
		return fmt.Errorf("getting 'subscriptionID' flag: %w", err)
	}
	if f.subscriptionID == "" && os.Getenv("ARM_SUBSCRIPTION_ID") == "" {
		return errors.New("either flag 'subscriptionID' or environment variable 'ARM_SUBSCRIPTION_ID' must be set")
	}

	f.region, err = flags.GetString("region")
	if err != nil {
		return fmt.Errorf("getting 'region' flag: %w", err)
	}
	f.resourceGroup, err = flags.GetString("resourceGroup")
	if err != nil {
		return fmt.Errorf("getting 'resourceGroup' flag: %w", err)
	}
	f.servicePrincipal, err = flags.GetString("servicePrincipal")
	if err != nil {
		return fmt.Errorf("getting 'servicePrincipal' flag: %w", err)
	}
	return nil
}

// azureIAMCreator implements the providerIAMCreator interface for Azure.
type azureIAMCreator struct {
	flags azureIAMCreateFlags
}

func (c *azureIAMCreator) getIAMConfigOptions() *cloudcmd.IAMConfigOptions {
	return &cloudcmd.IAMConfigOptions{
		Azure: cloudcmd.AzureIAMConfig{
			SubscriptionID:   c.flags.subscriptionID,
			Location:         c.flags.region,
			ResourceGroup:    c.flags.resourceGroup,
			ServicePrincipal: c.flags.servicePrincipal,
		},
	}
}

func (c *azureIAMCreator) printConfirmValues(cmd *cobra.Command) {
	cmd.Printf("Subscription ID:\t%s\n", c.flags.subscriptionID)
	cmd.Printf("Region:\t\t\t%s\n", c.flags.region)
	cmd.Printf("Resource Group:\t\t%s\n", c.flags.resourceGroup)
	cmd.Printf("Service Principal:\t%s\n\n", c.flags.servicePrincipal)
}

func (c *azureIAMCreator) printOutputValues(cmd *cobra.Command, iamFile cloudcmd.IAMOutput) {
	cmd.Printf("subscription:\t\t%s\n", iamFile.AzureOutput.SubscriptionID)
	cmd.Printf("tenant:\t\t\t%s\n", iamFile.AzureOutput.TenantID)
	cmd.Printf("location:\t\t%s\n", c.flags.region)
	cmd.Printf("resourceGroup:\t\t%s\n", c.flags.resourceGroup)
	cmd.Printf("userAssignedIdentity:\t%s\n", iamFile.AzureOutput.UAMIID)
}

func (c *azureIAMCreator) writeOutputValuesToConfig(conf *config.Config, iamFile cloudcmd.IAMOutput) {
	conf.Provider.Azure.SubscriptionID = iamFile.AzureOutput.SubscriptionID
	conf.Provider.Azure.TenantID = iamFile.AzureOutput.TenantID
	conf.Provider.Azure.Location = c.flags.region
	conf.Provider.Azure.ResourceGroup = c.flags.resourceGroup
	conf.Provider.Azure.UserAssignedIdentity = iamFile.AzureOutput.UAMIID
}

func (c *azureIAMCreator) parseAndWriteIDFile(_ cloudcmd.IAMOutput, _ file.Handler) error {
	return nil
}

func (c *azureIAMCreator) validateConfigWithFlagCompatibility(conf config.Config) error {
	return validateConfigWithFlagCompatibility(cloudprovider.Azure, conf, c.flags.region)
}
