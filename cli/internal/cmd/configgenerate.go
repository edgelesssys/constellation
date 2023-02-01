/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"errors"
	"fmt"

	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/siderolabs/talos/pkg/machinery/config/encoder"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func newConfigGenerateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate {aws|azure|gcp|qemu}",
		Short: "Generate a default configuration file",
		Long:  "Generate a default configuration file for your selected cloud provider.",
		Args: cobra.MatchAll(
			cobra.ExactArgs(1),
			isCloudProvider(0),
		),
		ValidArgsFunction: generateCompletion,
		RunE:              runConfigGenerate,
	}
	cmd.Flags().StringP("file", "f", constants.ConfigFilename, "path to output file, or '-' for stdout")
	cmd.Flags().String("name", "constell", "name to use for cluster creation and configuration")

	return cmd
}

type generateFlags struct {
	file string
	name string
}

type configGenerateCmd struct {
	log debugLog
}

func runConfigGenerate(cmd *cobra.Command, args []string) error {
	log, err := newCLILogger(cmd)
	if err != nil {
		return fmt.Errorf("creating logger: %w", err)
	}
	defer log.Sync()
	fileHandler := file.NewHandler(afero.NewOsFs())
	provider := cloudprovider.FromString(args[0])
	cg := &configGenerateCmd{log: log}
	return cg.configGenerate(cmd, fileHandler, provider)
}

func (cg *configGenerateCmd) configGenerate(cmd *cobra.Command, fileHandler file.Handler, provider cloudprovider.Provider) error {
	flags, err := parseGenerateFlags(cmd)
	if err != nil {
		return err
	}

	if provider == cloudprovider.AWS && len(flags.name) > 10 {
		return errors.New("cluster name on AWS must not be longer than 10 characters")
	}

	cg.log.Debugf("Parsed flags as %v", flags)
	cg.log.Debugf("Using cloud provider %s", provider.String())
	conf := createConfig(provider, flags.name)
	if flags.file == "-" {
		content, err := encoder.NewEncoder(conf).Encode()
		if err != nil {
			return fmt.Errorf("encoding config content: %w", err)
		}

		cg.log.Debugf("Writing YAML data to stdout")
		_, err = cmd.OutOrStdout().Write(content)
		return err
	}

	cg.log.Debugf("Writing YAML data to configuration file")
	if err := fileHandler.WriteYAML(flags.file, conf, file.OptMkdirAll); err != nil {
		return err
	}

	cmd.Println("Config file written to", flags.file)
	cmd.Println("Please fill in your CSP-specific configuration before proceeding.")
	cmd.Println("For more information refer to the documentation:")
	cmd.Println("\thttps://docs.edgeless.systems/constellation/getting-started/first-steps")

	return nil
}

// createConfig creates a config file for the given provider.
func createConfig(provider cloudprovider.Provider, name string) *config.Config {
	conf := config.Default()
	conf.RemoveProviderExcept(provider)
	conf.Name = name

	// set a lower default for QEMU's state disk
	if provider == cloudprovider.QEMU {
		conf.StateDiskSizeGB = 10
	}

	return conf
}

func parseGenerateFlags(cmd *cobra.Command) (generateFlags, error) {
	file, err := cmd.Flags().GetString("file")
	if err != nil {
		return generateFlags{}, fmt.Errorf("parsing config generate flags: %w", err)
	}
	name, err := parseNameFlag(cmd)
	if err != nil {
		return generateFlags{}, fmt.Errorf("parsing config generate flags: %w", err)
	}
	return generateFlags{
		file: file,
		name: name,
	}, nil
}

func parseNameFlag(cmd *cobra.Command) (string, error) {
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return "", fmt.Errorf("parsing name argument: %w", err)
	}

	if len(name) > constants.ConstellationNameLength {
		return "", fmt.Errorf(
			"name for Constellation cluster too long, maximum length is %d, got %d: %s",
			constants.ConstellationNameLength, len(name), name,
		)
	}

	return name, nil
}

// createCompletion handles the completion of the create command. It is frequently called
// while the user types arguments of the command to suggest completion.
func generateCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	switch len(args) {
	case 0:
		return []string{"aws", "gcp", "azure", "qemu"}, cobra.ShellCompDirectiveNoFileComp
	default:
		return []string{}, cobra.ShellCompDirectiveError
	}
}
