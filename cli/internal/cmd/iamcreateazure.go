/*
Copyright (c) Edgeless Systems GmbH
SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"fmt"

	"github.com/edgelesssys/constellation/v2/cli/internal/cloudcmd"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/spf13/cobra"
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

	cmd.Flags().String("resourceGroup", "", "Name of the resource group your IAM resources will be created in.")
	must(cobra.MarkFlagRequired(cmd.Flags(), "resourceGroup"))
	cmd.Flags().String("region", "", "Region the resources will be created in. (e.g. westus)")
	must(cobra.MarkFlagRequired(cmd.Flags(), "region"))
	cmd.Flags().String("servicePrincipal", "", "Name of the service principal that will be created.")
	must(cobra.MarkFlagRequired(cmd.Flags(), "servicePrincipal"))
	cmd.Flags().Bool("yes", false, "Create the IAM configuration without further confirmation")

	return cmd
}

func runIAMCreateAzure(cmd *cobra.Command, args []string) error {
	spinner := newSpinner(cmd.ErrOrStderr())
	defer spinner.Stop()
	creator := cloudcmd.NewIAMCreator(spinner)

	return iamCreateAzure(cmd, spinner, creator)
}

func iamCreateAzure(cmd *cobra.Command, spinner spinnerInterf, creator iamCreator) error {
	// Get input variables.
	azureFlags, err := parseAzureFlags(cmd)
	if err != nil {
		return err
	}

	// Confirmation.
	if !azureFlags.yesFlag {
		cmd.Printf("The following IAM configuration will be created:\n")
		cmd.Printf("Region:\t%s\n", azureFlags.region)
		cmd.Printf("Resource Group:\t%s\n", azureFlags.resourceGroup)
		cmd.Printf("Service Principal:\t%s\n", azureFlags.servicePrincipal)
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
	iamFile, err := creator.Create(cmd.Context(), cloudprovider.Azure, &cloudcmd.IAMConfig{
		Azure: cloudcmd.AzureIAMConfig{
			Region:           azureFlags.region,
			ServicePrincipal: azureFlags.servicePrincipal,
			ResourceGroup:    azureFlags.resourceGroup,
		},
	})
	spinner.Stop()
	if err != nil {
		return err
	}

	cmd.Printf("subscription:\t%s\n", iamFile.AzureOutput.SubscriptionID)
	cmd.Printf("tenant:\t%s\n", iamFile.AzureOutput.TenantID)
	cmd.Printf("location:\t%s\n", azureFlags.region)
	cmd.Printf("resourceGroup:\t%s\n", azureFlags.resourceGroup)
	cmd.Printf("userAssignedIdentity:\t%s\n", iamFile.AzureOutput.UAMIID)
	cmd.Printf("appClientID:\t%s\n", iamFile.AzureOutput.ApplicationID)
	cmd.Printf("appClientSecretValue:\t%s\n", iamFile.AzureOutput.ApplicationClientSecretValue)
	cmd.Println("Your IAM configuration was created successfully. Please fill the above values into your configuration file.")

	return nil
}

// parseAzureFlags parses and validates the flags of the iam create azure command.
func parseAzureFlags(cmd *cobra.Command) (azureFlags, error) {
	region, err := cmd.Flags().GetString("region")
	if err != nil {
		return azureFlags{}, fmt.Errorf("parsing region string: %w", err)
	}
	resourceGroup, err := cmd.Flags().GetString("resourceGroup")
	if err != nil {
		return azureFlags{}, fmt.Errorf("parsing resourceGroup string: %w", err)
	}
	servicePrincipal, err := cmd.Flags().GetString("servicePrincipal")
	if err != nil {
		return azureFlags{}, fmt.Errorf("parsing servicePrincipal string: %w", err)
	}
	yesFlag, err := cmd.Flags().GetBool("yes")
	if err != nil {
		return azureFlags{}, fmt.Errorf("parsing yes bool: %w", err)
	}
	return azureFlags{
		servicePrincipal: servicePrincipal,
		resourceGroup:    resourceGroup,
		region:           region,
		yesFlag:          yesFlag,
	}, nil
}

// azureFlags contains the parsed flags of the iam create azure command.
type azureFlags struct {
	region           string
	resourceGroup    string
	servicePrincipal string
	yesFlag          bool
}
