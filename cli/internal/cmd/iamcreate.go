/*
Copyright (c) Edgeless Systems GmbH
SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/edgelesssys/constellation/v2/cli/internal/cloudcmd"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	// GCP-specific validation regexes
	// Source: https://cloud.google.com/compute/docs/regions-zones
	zoneRegex   = regexp.MustCompile(`^\w+-\w+-[abc]$`)
	regionRegex = regexp.MustCompile(`^\w+-\w+[0-9]$`)
	// Source: https://cloud.google.com/resource-manager/reference/rest/v1/projects.
	gcpIDRegex = regexp.MustCompile(`^[a-z][-a-z0-9]{4,28}[a-z0-9]$`)
)

// newIAMCreateCmd returns a new cobra.Command for the iam create parent command. It needs another verb, and does nothing on its own.
func newIAMCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create IAM configuration on a cloud platform for your Constellation cluster",
		Long:  "Create IAM configuration on a cloud platform for your Constellation cluster.",
		Args:  cobra.ExactArgs(0),
	}

	cmd.PersistentFlags().BoolP("yes", "y", false, "create the IAM configuration without further confirmation")
	cmd.PersistentFlags().Bool("update-config", false, "update the config file with the specific IAM information")

	cmd.AddCommand(newIAMCreateAWSCmd())
	cmd.AddCommand(newIAMCreateAzureCmd())
	cmd.AddCommand(newIAMCreateGCPCmd())

	return cmd
}

type iamCreateFlags struct {
	rootFlags
	yes          bool
	updateConfig bool
}

func (f *iamCreateFlags) parse(flags *pflag.FlagSet) error {
	var err error
	if err = f.rootFlags.parse(flags); err != nil {
		return err
	}
	f.yes, err = flags.GetBool("yes")
	if err != nil {
		return fmt.Errorf("getting 'yes' flag: %w", err)
	}
	f.updateConfig, err = flags.GetBool("update-config")
	if err != nil {
		return fmt.Errorf("getting 'update-config' flag: %w", err)
	}
	return nil
}

func runIAMCreate(cmd *cobra.Command, providerCreator providerIAMCreator, provider cloudprovider.Provider) error {
	spinner, err := newSpinnerOrStderr(cmd)
	if err != nil {
		return fmt.Errorf("creating spinner: %w", err)
	}
	defer spinner.Stop()
	log, err := newCLILogger(cmd)
	if err != nil {
		return fmt.Errorf("creating logger: %w", err)
	}
	defer log.Sync()

	iamCreator := &iamCreator{
		cmd:             cmd,
		spinner:         spinner,
		log:             log,
		creator:         cloudcmd.NewIAMCreator(spinner),
		fileHandler:     file.NewHandler(afero.NewOsFs()),
		providerCreator: providerCreator,
		provider:        provider,
	}
	if err := iamCreator.flags.parse(cmd.Flags()); err != nil {
		return err
	}

	return iamCreator.create(cmd.Context())
}

// iamCreator is the iamCreator for the iam create command.
type iamCreator struct {
	cmd             *cobra.Command
	spinner         spinnerInterf
	creator         cloudIAMCreator
	fileHandler     file.Handler
	provider        cloudprovider.Provider
	providerCreator providerIAMCreator
	iamConfig       *cloudcmd.IAMConfigOptions
	log             debugLog
	flags           iamCreateFlags
}

// create IAM configuration on the iamCreator's cloud provider.
func (c *iamCreator) create(ctx context.Context) error {
	if err := c.checkWorkingDir(); err != nil {
		return err
	}

	if !c.flags.yes {
		c.cmd.Printf("The following IAM configuration will be created:\n\n")
		c.providerCreator.printConfirmValues(c.cmd)
		ok, err := askToConfirm(c.cmd, "Do you want to create the configuration?")
		if err != nil {
			return err
		}
		if !ok {
			c.cmd.Println("The creation of the configuration was aborted.")
			return nil
		}
	}

	var conf config.Config
	if c.flags.updateConfig {
		c.log.Debugf("Parsing config %s", c.flags.pathPrefixer.PrefixPrintablePath(constants.ConfigFilename))
		if err := c.fileHandler.ReadYAML(constants.ConfigFilename, &conf); err != nil {
			return fmt.Errorf("error reading the configuration file: %w", err)
		}
		if err := c.providerCreator.validateConfigWithFlagCompatibility(conf); err != nil {
			return err
		}
		c.cmd.Printf("The configuration file %q will be automatically updated with the IAM values and zone/region information.\n", c.flags.pathPrefixer.PrefixPrintablePath(constants.ConfigFilename))
	}

	iamConfig := c.providerCreator.getIAMConfigOptions()
	iamConfig.TFWorkspace = constants.TerraformIAMWorkingDir
	iamConfig.TFLogLevel = c.flags.tfLogLevel
	c.spinner.Start("Creating", false)
	iamFile, err := c.creator.Create(ctx, c.provider, iamConfig)
	c.spinner.Stop()
	if err != nil {
		return err
	}
	c.cmd.Println() // Print empty line to separate after spinner ended.
	c.log.Debugf("Successfully created the IAM cloud resources")

	err = c.providerCreator.parseAndWriteIDFile(iamFile, c.fileHandler)
	if err != nil {
		return err
	}

	if c.flags.updateConfig {
		c.log.Debugf("Writing IAM configuration to %s", c.flags.pathPrefixer.PrefixPrintablePath(constants.ConfigFilename))
		c.providerCreator.writeOutputValuesToConfig(&conf, iamFile)
		if err := c.fileHandler.WriteYAML(constants.ConfigFilename, conf, file.OptOverwrite); err != nil {
			return err
		}
		c.cmd.Printf("Your IAM configuration was created and filled into %s successfully.\n", c.flags.pathPrefixer.PrefixPrintablePath(constants.ConfigFilename))
		return nil
	}

	c.providerCreator.printOutputValues(c.cmd, iamFile)
	c.cmd.Println("Your IAM configuration was created successfully. Please fill the above values into your configuration file.")

	return nil
}

// checkWorkingDir checks if the current working directory already contains a Terraform dir.
func (c *iamCreator) checkWorkingDir() error {
	if _, err := c.fileHandler.Stat(constants.TerraformIAMWorkingDir); err == nil {
		return fmt.Errorf(
			"the current working directory already contains the Terraform workspace directory %q. Please run the command in a different directory or destroy the existing workspace",
			c.flags.pathPrefixer.PrefixPrintablePath(constants.TerraformIAMWorkingDir),
		)
	}
	return nil
}

// providerIAMCreator is an interface for the IAM actions of different cloud providers.
type providerIAMCreator interface {
	// printConfirmValues prints the values that will be created on the cloud provider and need to be confirmed by the user.
	printConfirmValues(cmd *cobra.Command)
	// printOutputValues prints the values that were created on the cloud provider.
	printOutputValues(cmd *cobra.Command, iamFile cloudcmd.IAMOutput)
	// writeOutputValuesToConfig writes the output values of the IAM creation to the constellation config file.
	writeOutputValuesToConfig(conf *config.Config, iamFile cloudcmd.IAMOutput)
	// getIAMConfigOptions sets up the IAM values required to create the IAM configuration.
	getIAMConfigOptions() *cloudcmd.IAMConfigOptions
	// parseAndWriteIDFile parses the GCP service account key and writes it to a keyfile. It is only implemented for GCP.
	parseAndWriteIDFile(iamFile cloudcmd.IAMOutput, fileHandler file.Handler) error

	validateConfigWithFlagCompatibility(config.Config) error
}

// parseIDFile parses the given base64 encoded JSON string of the GCP service account key and returns a map.
func parseIDFile(serviceAccountKeyBase64 string) (map[string]string, error) {
	dec, err := base64.StdEncoding.DecodeString(serviceAccountKeyBase64)
	if err != nil {
		return nil, err
	}

	out := make(map[string]string)
	if err = json.Unmarshal(dec, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// validateConfigWithFlagCompatibility checks if the config is compatible with the flags.
func validateConfigWithFlagCompatibility(iamProvider cloudprovider.Provider, cfg config.Config, zone string) error {
	if !cfg.HasProvider(iamProvider) {
		return fmt.Errorf("cloud provider from the the configuration file differs from the one provided via the command %q", iamProvider)
	}
	return checkIfCfgZoneAndFlagZoneDiffer(zone, cfg)
}

func checkIfCfgZoneAndFlagZoneDiffer(zone string, cfg config.Config) error {
	configZone := cfg.GetZone()
	if configZone != "" && zone != configZone {
		return fmt.Errorf("zone/region from the configuration file %q differs from the one provided via flags %q", configZone, zone)
	}
	return nil
}
