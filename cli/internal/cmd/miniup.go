/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/edgelesssys/constellation/v2/cli/internal/cloudcmd"
	"github.com/edgelesssys/constellation/v2/cli/internal/featureset"
	"github.com/edgelesssys/constellation/v2/cli/internal/libvirt"
	"github.com/edgelesssys/constellation/v2/cli/internal/state"
	"github.com/edgelesssys/constellation/v2/internal/api/attestationconfigapi"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
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
	spinner, err := newSpinnerOrStderr(cmd)
	if err != nil {
		return err
	}
	defer spinner.Stop()
	creator := cloudcmd.NewCreator(spinner)

	m := &miniUpCmd{
		log:           log,
		configFetcher: attestationconfigapi.NewFetcher(),
		fileHandler:   file.NewHandler(afero.NewOsFs()),
	}
	if err := m.flags.parse(cmd.Flags()); err != nil {
		return err
	}
	return m.up(cmd, creator, spinner)
}

func (m *miniUpCmd) up(cmd *cobra.Command, creator cloudCreator, spinner spinnerInterf) error {
	if err := m.checkSystemRequirements(cmd.ErrOrStderr()); err != nil {
		return fmt.Errorf("system requirements not met: %w", err)
	}

	// create config if not passed as flag and set default values
	config, err := m.prepareConfig(cmd)
	if err != nil {
		return fmt.Errorf("preparing config: %w", err)
	}

	// create cluster
	spinner.Start("Creating cluster in QEMU ", false)
	err = m.createMiniCluster(cmd.Context(), creator, config)
	spinner.Stop()
	if err != nil {
		return fmt.Errorf("creating cluster: %w", err)
	}
	cmd.Println("Cluster successfully created.")
	connectURI := config.Provider.QEMU.LibvirtURI
	m.log.Debugf("Using connect URI %s", connectURI)
	if connectURI == "" {
		connectURI = libvirt.LibvirtTCPConnectURI
	}
	cmd.Println("Connect to the VMs by executing:")
	cmd.Printf("\tvirsh -c %s\n\n", connectURI)

	// initialize cluster
	if err := m.initializeMiniCluster(cmd); err != nil {
		return fmt.Errorf("initializing cluster: %w", err)
	}
	m.log.Debugf("Initialized cluster")
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
	m.log.Debugf("Prepared configuration")

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

// createMiniCluster creates a new cluster using the given config.
func (m *miniUpCmd) createMiniCluster(ctx context.Context, creator cloudCreator, config *config.Config) error {
	m.log.Debugf("Creating mini cluster")
	opts := cloudcmd.CreateOptions{
		Provider:    cloudprovider.QEMU,
		Config:      config,
		TFWorkspace: constants.TerraformWorkingDir,
		TFLogLevel:  m.flags.tfLogLevel,
	}
	infraState, err := creator.Create(ctx, opts)
	if err != nil {
		return err
	}

	infraState.UID = constants.MiniConstellationUID // use UID "mini" to identify MiniConstellation clusters.

	stateFile := state.New().
		SetInfrastructure(infraState)

	m.log.Debugf("Cluster state file contains %v", stateFile)
	return stateFile.WriteToFile(m.fileHandler, constants.StateFilename)
}

// initializeMiniCluster initializes a QEMU cluster.
func (m *miniUpCmd) initializeMiniCluster(cmd *cobra.Command) (retErr error) {
	m.log.Debugf("Initializing mini cluster")
	// clean up cluster resources if initialization fails
	defer func() {
		if retErr != nil {
			cmd.PrintErrf("An error occurred: %s\n", retErr)
			cmd.PrintErrln("Attempting to roll back.")
			_ = runDown(cmd, []string{})
			cmd.PrintErrf("Rollback succeeded.\n\n")
		}
	}()

	// Define flags for apply backend that are not set by mini up
	cmd.Flags().StringSlice(
		"skip-phases",
		[]string{string(skipInfrastructurePhase), string(skipK8sPhase), string(skipImagePhase)},
		"",
	)
	cmd.Flags().Bool("yes", false, "")
	cmd.Flags().Bool("skip-helm-wait", false, "")
	cmd.Flags().Bool("conformance", false, "")
	cmd.Flags().Duration("timeout", time.Hour, "")

	if err := runApply(cmd, nil); err != nil {
		return err
	}
	m.log.Debugf("Initialized mini cluster")
	return nil
}
