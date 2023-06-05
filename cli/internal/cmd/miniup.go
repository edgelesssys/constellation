/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/edgelesssys/constellation/v2/cli/internal/cloudcmd"
	"github.com/edgelesssys/constellation/v2/cli/internal/libvirt"
	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/internal/api/attestationconfigapi"
	"github.com/edgelesssys/constellation/v2/internal/atls"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/grpc/dialer"
	"github.com/edgelesssys/constellation/v2/internal/license"
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

	// override global flag so we don't have a default value for the config
	cmd.Flags().String("config", "", "path to the configuration file to use for the cluster")
	cmd.Flags().Bool("merge-kubeconfig", true, "merge Constellation kubeconfig file with default kubeconfig file in $HOME/.kube/config")

	return cmd
}

type miniUpCmd struct {
	log           debugLog
	configFetcher attestationconfigapi.Fetcher
}

func runUp(cmd *cobra.Command, _ []string) error {
	log, err := newCLILogger(cmd)
	if err != nil {
		return fmt.Errorf("creating logger: %w", err)
	}
	defer log.Sync()
	spinner, err := newSpinnerOrStderr(cmd)
	if err != nil {
		return err
	}
	defer spinner.Stop()
	creator := cloudcmd.NewCreator(spinner)

	m := &miniUpCmd{log: log, configFetcher: attestationconfigapi.NewFetcher()}
	return m.up(cmd, creator, spinner)
}

func (m *miniUpCmd) up(cmd *cobra.Command, creator cloudCreator, spinner spinnerInterf) error {
	if err := m.checkSystemRequirements(cmd.ErrOrStderr()); err != nil {
		return fmt.Errorf("system requirements not met: %w", err)
	}

	flags, err := m.parseUpFlags(cmd)
	if err != nil {
		return fmt.Errorf("parsing flags: %w", err)
	}

	fileHandler := file.NewHandler(afero.NewOsFs())

	// create config if not passed as flag and set default values
	config, err := m.prepareConfig(cmd, fileHandler, flags)
	if err != nil {
		return fmt.Errorf("preparing config: %w", err)
	}

	// create cluster
	spinner.Start("Creating cluster in QEMU ", false)
	err = m.createMiniCluster(cmd.Context(), fileHandler, creator, config, flags.tfLogLevel)
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
	if err := m.initializeMiniCluster(cmd, fileHandler, spinner); err != nil {
		return fmt.Errorf("initializing cluster: %w", err)
	}
	m.log.Debugf("Initialized cluster")
	return nil
}

// prepareConfig reads a given config, or creates a new minimal QEMU config.
func (m *miniUpCmd) prepareConfig(cmd *cobra.Command, fileHandler file.Handler, flags upFlags) (*config.Config, error) {
	// check for existing config
	if flags.configPath != "" {
		conf, err := config.New(fileHandler, flags.configPath, m.configFetcher, flags.force)
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
	m.log.Debugf("Configuration path is %q", flags.configPath)
	if err := cmd.Flags().Set("config", constants.ConfigFilename); err != nil {
		return nil, err
	}
	_, err := fileHandler.Stat(constants.ConfigFilename)
	if err == nil {
		// config already exists, prompt user to overwrite
		cmd.PrintErrln("A config file already exists in the current workspace. Use --config to use an existing config file.")
		ok, err := askToConfirm(cmd, "Do you want to overwrite it?")
		if err != nil {
			return nil, err
		}

		if !ok {
			return nil, errors.New("not overwriting existing config")
		}
	}

	config := config.Default()
	config.Name = constants.MiniConstellationUID
	config.RemoveProviderAndAttestationExcept(cloudprovider.QEMU)
	config.StateDiskSizeGB = 8

	// only release images (e.g. v2.7.0) use the production NVRAM
	if !config.IsReleaseImage() {
		config.Provider.QEMU.NVRAM = "testing"
	}

	m.log.Debugf("Prepared configuration")

	return config, fileHandler.WriteYAML(constants.ConfigFilename, config, file.OptOverwrite)
}

// createMiniCluster creates a new cluster using the given config.
func (m *miniUpCmd) createMiniCluster(ctx context.Context, fileHandler file.Handler, creator cloudCreator, config *config.Config, tfLogLevel terraform.LogLevel) error {
	m.log.Debugf("Creating mini cluster")
	opts := cloudcmd.CreateOptions{
		Provider:          cloudprovider.QEMU,
		Config:            config,
		InsType:           "",
		ControlPlaneCount: 1,
		WorkerCount:       1,
		TFLogLevel:        tfLogLevel,
	}
	idFile, err := creator.Create(ctx, opts)
	if err != nil {
		return err
	}

	idFile.UID = constants.MiniConstellationUID // use UID "mini" to identify MiniConstellation clusters.
	m.log.Debugf("Cluster id file contains %v", idFile)
	return fileHandler.WriteJSON(constants.ClusterIDsFileName, idFile, file.OptNone)
}

// initializeMiniCluster initializes a QEMU cluster.
func (m *miniUpCmd) initializeMiniCluster(cmd *cobra.Command, fileHandler file.Handler, spinner spinnerInterf) (retErr error) {
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
	newDialer := func(validator atls.Validator) *dialer.Dialer {
		return dialer.New(nil, validator, &net.Dialer{})
	}
	m.log.Debugf("Created new dialer")
	cmd.Flags().String("master-secret", "", "")
	cmd.Flags().String("endpoint", "", "")
	cmd.Flags().Bool("conformance", false, "")
	log, err := newCLILogger(cmd)
	if err != nil {
		return fmt.Errorf("creating logger: %w", err)
	}
	m.log.Debugf("Created new logger")
	defer log.Sync()
	i := &initCmd{log: log, merger: &kubeconfigMerger{log: log}, spinner: spinner}
	if err := i.initialize(cmd, newDialer, fileHandler, license.NewClient(), m.configFetcher); err != nil {
		return err
	}
	m.log.Debugf("Initialized mini cluster")
	return nil
}

type upFlags struct {
	configPath string
	force      bool
	tfLogLevel terraform.LogLevel
}

func (m *miniUpCmd) parseUpFlags(cmd *cobra.Command) (upFlags, error) {
	m.log.Debugf("Preparing configuration")
	configPath, err := cmd.Flags().GetString("config")
	if err != nil {
		return upFlags{}, fmt.Errorf("parsing config string: %w", err)
	}
	m.log.Debugf("Configuration path is %q", configPath)
	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return upFlags{}, fmt.Errorf("parsing force bool: %w", err)
	}
	m.log.Debugf("force flag is %q", configPath)

	logLevelString, err := cmd.Flags().GetString("tf-log")
	if err != nil {
		return upFlags{}, fmt.Errorf("parsing tf-log string: %w", err)
	}
	logLevel, err := terraform.ParseLogLevel(logLevelString)
	if err != nil {
		return upFlags{}, fmt.Errorf("parsing Terraform log level %s: %w", logLevelString, err)
	}
	m.log.Debugf("Terraform logs will be written into %s at level %s", constants.TerraformLogFile, logLevel.String())

	return upFlags{
		configPath: configPath,
		force:      force,
		tfLogLevel: logLevel,
	}, nil
}
