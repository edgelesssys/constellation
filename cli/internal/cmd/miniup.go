/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/edgelesssys/constellation/v2/cli/internal/cloudcmd"
	"github.com/edgelesssys/constellation/v2/cli/internal/cmd/pathprefix"
	"github.com/edgelesssys/constellation/v2/cli/internal/featureset"
	"github.com/edgelesssys/constellation/v2/cli/internal/helm"
	"github.com/edgelesssys/constellation/v2/cli/internal/kubecmd"
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
	err = m.createMiniCluster(cmd.Context(), fileHandler, creator, config, flags)
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
	_, err := fileHandler.Stat(constants.ConfigFilename)
	if err == nil {
		// config already exists, prompt user if they want to use this file
		cmd.PrintErrln("A config file already exists in the configured workspace.")
		ok, err := askToConfirm(cmd, "Do you want to create the Constellation using that config?")
		if err != nil {
			return nil, err
		}
		if ok {
			return m.prepareExistingConfig(cmd, fileHandler, flags)
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

	return config, fileHandler.WriteYAML(constants.ConfigFilename, config, file.OptOverwrite)
}

func (m *miniUpCmd) prepareExistingConfig(cmd *cobra.Command, fileHandler file.Handler, flags upFlags) (*config.Config, error) {
	conf, err := config.New(fileHandler, constants.ConfigFilename, m.configFetcher, flags.force)
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
func (m *miniUpCmd) createMiniCluster(ctx context.Context, fileHandler file.Handler, creator cloudCreator, config *config.Config, flags upFlags) error {
	m.log.Debugf("Creating mini cluster")
	opts := cloudcmd.CreateOptions{
		Provider:    cloudprovider.QEMU,
		Config:      config,
		TFWorkspace: constants.TerraformWorkingDir,
		TFLogLevel:  flags.tfLogLevel,
	}
	idFile, err := creator.Create(ctx, opts)
	if err != nil {
		return err
	}

	idFile.UID = constants.MiniConstellationUID // use UID "mini" to identify MiniConstellation clusters.
	m.log.Debugf("Cluster id file contains %v", idFile)
	return fileHandler.WriteJSON(constants.ClusterIDsFilename, idFile, file.OptNone)
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
	cmd.Flags().String("endpoint", "", "")
	cmd.Flags().Bool("conformance", false, "")
	cmd.Flags().Bool("skip-helm-wait", false, "install helm charts without waiting for deployments to be ready")
	log, err := newCLILogger(cmd)
	if err != nil {
		return fmt.Errorf("creating logger: %w", err)
	}
	m.log.Debugf("Created new logger")
	defer log.Sync()

	newAttestationApplier := func(w io.Writer, kubeConfig string, log debugLog) (attestationConfigApplier, error) {
		return kubecmd.New(w, kubeConfig, fileHandler, log)
	}
	newHelmClient := func(kubeConfigPath string, log debugLog) (helmApplier, error) {
		return helm.NewClient(kubeConfigPath, log)
	} // need to defer helm client instantiation until kubeconfig is available
	i := newInitCmd(fileHandler, spinner, &kubeconfigMerger{log: log}, log)
	if err := i.initialize(cmd, newDialer, license.NewClient(), m.configFetcher,
		newAttestationApplier, newHelmClient); err != nil {
		return err
	}
	m.log.Debugf("Initialized mini cluster")
	return nil
}

type upFlags struct {
	force      bool
	tfLogLevel terraform.LogLevel
}

func (m *miniUpCmd) parseUpFlags(cmd *cobra.Command) (upFlags, error) {
	m.log.Debugf("Preparing configuration")
	workDir, err := cmd.Flags().GetString("workspace")
	if err != nil {
		return upFlags{}, fmt.Errorf("parsing config string: %w", err)
	}
	m.log.Debugf("Workspace set to %q", workDir)
	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return upFlags{}, fmt.Errorf("parsing force bool: %w", err)
	}
	m.log.Debugf("force flag is %q", force)

	logLevelString, err := cmd.Flags().GetString("tf-log")
	if err != nil {
		return upFlags{}, fmt.Errorf("parsing tf-log string: %w", err)
	}
	logLevel, err := terraform.ParseLogLevel(logLevelString)
	if err != nil {
		return upFlags{}, fmt.Errorf("parsing Terraform log level %s: %w", logLevelString, err)
	}
	m.log.Debugf("Terraform logs will be written into %s at level %s", pathprefix.New(workDir).PrefixPrintablePath(constants.TerraformLogFile), logLevel.String())

	return upFlags{
		force:      force,
		tfLogLevel: logLevel,
	}, nil
}
