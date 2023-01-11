/*
Copyright (c) Edgeless Systems GmbH
SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"fmt"

	"github.com/edgelesssys/constellation/v2/cli/internal/cloudcmd"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/spf13/afero"
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

	cmd.Flags().String("resourceGroup", "", "Name prefix of the two resource groups your cluster / IAM resources will be created in.")
	must(cobra.MarkFlagRequired(cmd.Flags(), "resourceGroup"))
	cmd.Flags().String("region", "", "Region the resources will be created in. (e.g. westus)")
	must(cobra.MarkFlagRequired(cmd.Flags(), "region"))
	cmd.Flags().String("servicePrincipal", "", "Name of the service principal that will be created.")
	must(cobra.MarkFlagRequired(cmd.Flags(), "servicePrincipal"))
	cmd.Flags().Bool("yes", false, "Create the IAM configuration without further confirmation.")

	return cmd
}

func runIAMCreateAzure(cmd *cobra.Command, args []string) error {
	spinner := newSpinner(cmd.ErrOrStderr())
	defer spinner.Stop()
	fileHandler := file.NewHandler(afero.NewOsFs())
	creator := cloudcmd.NewIAMCreator(spinner)

	return iamCreateAzure(cmd, spinner, creator, fileHandler)
}

func iamCreateAzure(cmd *cobra.Command, spinner spinnerInterf, creator iamCreator, fileHandler file.Handler) error {
	// Get input variables.
	azureFlags, err := parseAzureFlags(cmd)
	if err != nil {
		return err
	}

	// Confirmation.
	if !azureFlags.yesFlag {
		cmd.Printf("The following IAM configuration will be created:\n\n")
		cmd.Printf("Region:\t\t\t%s\n", azureFlags.region)
		cmd.Printf("Resource Group:\t\t%s\n", azureFlags.resourceGroup)
		cmd.Printf("Service Principal:\t%s\n\n", azureFlags.servicePrincipal)
		if azureFlags.generateConfig {
			cmd.Printf("The configuration file %s will be automatically generated and populated with the IAM values.\n", azureFlags.configPath)
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

	conf := createConfig(cloudprovider.Azure)

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
	cmd.Println() // Print empty line to separate after spinner ended.

	if azureFlags.generateConfig {
		conf.Provider.Azure.SubscriptionID = iamFile.AzureOutput.SubscriptionID
		conf.Provider.Azure.TenantID = iamFile.AzureOutput.TenantID
		conf.Provider.Azure.Location = azureFlags.region
		conf.Provider.Azure.ResourceGroup = azureFlags.resourceGroup
		conf.Provider.Azure.UserAssignedIdentity = iamFile.AzureOutput.UAMIID
		conf.Provider.Azure.AppClientID = iamFile.AzureOutput.ApplicationID
		conf.Provider.Azure.ClientSecretValue = iamFile.AzureOutput.ApplicationClientSecretValue
		if err := fileHandler.WriteYAML(azureFlags.configPath, conf, file.OptMkdirAll); err != nil {
			return err
		}
		cmd.Printf("Your IAM configuration was created and filled into %s successfully.\n", azureFlags.configPath)
		return nil
	}

	cmd.Printf("subscription:\t\t%s\n", iamFile.AzureOutput.SubscriptionID)
	cmd.Printf("tenant:\t\t\t%s\n", iamFile.AzureOutput.TenantID)
	cmd.Printf("location:\t\t%s\n", azureFlags.region)
	cmd.Printf("resourceGroup:\t\t%s\n", azureFlags.resourceGroup)
	cmd.Printf("userAssignedIdentity:\t%s\n", iamFile.AzureOutput.UAMIID)
	cmd.Printf("appClientID:\t\t%s\n", iamFile.AzureOutput.ApplicationID)
	cmd.Printf("appClientSecretValue:\t%s\n\n", iamFile.AzureOutput.ApplicationClientSecretValue)
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

	configPath, err := cmd.Flags().GetString("config")
	if err != nil {
		return azureFlags{}, fmt.Errorf("parsing config string: %w", err)
	}

	generateConfig, err := cmd.Flags().GetBool("generate-config")
	if err != nil {
		return azureFlags{}, fmt.Errorf("parsing generate-config bool: %w", err)
	}

	yesFlag, err := cmd.Flags().GetBool("yes")
	if err != nil {
		return azureFlags{}, fmt.Errorf("parsing yes bool: %w", err)
	}

	return azureFlags{
		servicePrincipal: servicePrincipal,
		resourceGroup:    resourceGroup,
		region:           region,
		generateConfig:   generateConfig,
		configPath:       configPath,
		yesFlag:          yesFlag,
	}, nil
}

// azureFlags contains the parsed flags of the iam create azure command.
type azureFlags struct {
	region           string
	resourceGroup    string
	servicePrincipal string

	generateConfig bool
	configPath     string
	yesFlag        bool
}
