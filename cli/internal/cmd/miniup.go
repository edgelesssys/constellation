/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/edgelesssys/constellation/v2/cli/internal/libvirt"
	"github.com/edgelesssys/constellation/v2/internal/api/attestationconfigapi"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/constellation/featureset"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func newMiniUpCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "up",
		Short: "Create and initialize a new MiniConstellation cluster",
		Long: "Create and initialize a new MiniConstellation cluster.\n\n" +
			"A mini cluster consists of a single control-plane and worker node, hosted using QEMU/KVM.",
		Args: cobra.ExactArgs(0),
		RunE: runUp,
	}

	cmd.Flags().Bool("merge-kubeconfig", true, "merge Constellation kubeconfig file with default kubeconfig file in $HOME/.kube/config")

	return cmd
}

type miniUpCmd struct {
	log           debugLog
	configFetcher attestationconfigapi.Fetcher
	fileHandler   file.Handler
	flags         rootFlags
}

func runUp(cmd *cobra.Command, _ []string) error {
	log, err := newCLILogger(cmd)
	if err != nil {
		return fmt.Errorf("creating logger: %w", err)
	}

	m := &miniUpCmd{
		log:           log,
		configFetcher: attestationconfigapi.NewFetcher(),
		fileHandler:   file.NewHandler(afero.NewOsFs()),
	}
	if err := m.flags.parse(cmd.Flags()); err != nil {
		return err
	}

	return m.up(cmd)
}

func (m *miniUpCmd) up(cmd *cobra.Command) (retErr error) {
	if err := m.checkSystemRequirements(cmd.ErrOrStderr()); err != nil {
		return fmt.Errorf("system requirements not met: %w", err)
	}

	if clean, err := m.fileHandler.IsEmpty(constants.TerraformWorkingDir); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("checking if terraform working directory is empty: %w", err)
	} else if err == nil && !clean {
		return fmt.Errorf(
			"directory %q already exists and is not empty, run 'constellation mini down' before creating a new cluster",
			m.flags.pathPrefixer.PrefixPrintablePath(constants.TerraformWorkingDir),
		)
	}

	// create config if not present in directory and set default values
	config, err := m.prepareConfig(cmd)
	if err != nil {
		return fmt.Errorf("preparing config: %w", err)
	}

	// clean up cluster resources if setup fails
	defer func() {
		if retErr != nil {
			cmd.PrintErrf("An error occurred: %s\n", retErr)
			cmd.PrintErrln("Attempting to roll back.")
			err = runDown(cmd, []string{})
			if err != nil {
				cmd.PrintErrf("Rollback failed: %s\n", err)
			} else {
				cmd.PrintErrf("Rollback succeeded.\n\n")
			}
		}
	}()

	// set flags not defined by "mini up"
	cmd.Flags().StringSlice("skip-phases", []string{}, "")
	cmd.Flags().Bool("yes", true, "")
	cmd.Flags().Bool("skip-helm-wait", false, "")
	cmd.Flags().Bool("conformance", false, "")
	cmd.Flags().Duration("helm-timeout", time.Hour, "")

	// create and initialize the cluster
	if err := runApply(cmd, nil); err != nil {
		return fmt.Errorf("creating cluster: %w", err)
	}

	connectURI := config.Provider.QEMU.LibvirtURI
	if connectURI == "" {
		connectURI = libvirt.LibvirtTCPConnectURI
	}
	cmd.Println("Connect to the VMs by executing:")
	cmd.Printf("\tvirsh -c %s\n\n", connectURI)

	return nil
}

// prepareConfig reads a given config, or creates a new minimal QEMU config.
func (m *miniUpCmd) prepareConfig(cmd *cobra.Command) (*config.Config, error) {
	_, err := m.fileHandler.Stat(constants.ConfigFilename)
	if err == nil {
		// config already exists, prompt user if they want to use this file
		cmd.PrintErrln("A config file already exists in the configured workspace.")
		ok, err := askToConfirm(cmd, "Do you want to create the Constellation using that config?")
		if err != nil {
			return nil, err
		}
		if ok {
			return m.prepareExistingConfig(cmd)
		}

		// user declined to reuse config file, prompt if they want to overwrite it
		ok, err = askToConfirm(cmd, "Do you want to overwrite it and create a new config?")
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, errors.New("not overwriting existing config")
		}
	}

	if !featureset.CanUseEmbeddedMeasurmentsAndImage {
		cmd.PrintErrln("Generating a valid default config is not supported in the OSS build of the Constellation CLI. Consult the documentation for instructions on where to download the enterprise version.")
		return nil, errors.New("cannot create a mini cluster without a config file in the OSS build")
	}
	config, err := config.MiniDefault()
	if err != nil {
		return nil, fmt.Errorf("mini default config is invalid: %v", err)
	}
	m.log.Debug("Prepared configuration")

	return config, m.fileHandler.WriteYAML(constants.ConfigFilename, config, file.OptOverwrite)
}

func (m *miniUpCmd) prepareExistingConfig(cmd *cobra.Command) (*config.Config, error) {
	conf, err := config.New(m.fileHandler, constants.ConfigFilename, m.configFetcher, m.flags.force)
	var configValidationErr *config.ValidationError
	if errors.As(err, &configValidationErr) {
		cmd.PrintErrln(configValidationErr.LongMessage())
	}
	if err != nil {
		return nil, err
	}
	if conf.GetProvider() != cloudprovider.QEMU {
		return nil, errors.New("invalid provider for MiniConstellation cluster")
	}
	return conf, nil
}
