/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"runtime"
	"strings"

	"github.com/edgelesssys/constellation/v2/cli/internal/cloudcmd"
	"github.com/edgelesssys/constellation/v2/cli/internal/libvirt"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/grpc/dialer"
	"github.com/edgelesssys/constellation/v2/internal/license"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"golang.org/x/sys/unix"
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

	return cmd
}

type miniUpCmd struct {
	log debugLog
}

func runUp(cmd *cobra.Command, args []string) error {
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

	m := &miniUpCmd{log: log}
	return m.up(cmd, creator, spinner)
}

func (m *miniUpCmd) up(cmd *cobra.Command, creator cloudCreator, spinner spinnerInterf) error {
	if err := m.checkSystemRequirements(cmd.ErrOrStderr()); err != nil {
		return fmt.Errorf("system requirements not met: %w", err)
	}

	fileHandler := file.NewHandler(afero.NewOsFs())

	// create config if not passed as flag and set default values
	config, err := m.prepareConfig(cmd, fileHandler)
	if err != nil {
		return fmt.Errorf("preparing config: %w", err)
	}
	m.log.Debugf("Prepared configuration")

	// create cluster
	spinner.Start("Creating cluster in QEMU ", false)
	err = m.createMiniCluster(cmd.Context(), fileHandler, creator, config)
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

// checkSystemRequirements checks if the system meets the requirements for running a MiniConstellation cluster.
// We do so by verifying that the host:
// - arch/os is linux/amd64.
// - has access to /dev/kvm.
// - has at least 4 CPU cores.
// - has at least 4GB of memory.
// - has at least 20GB of free disk space.
func (m *miniUpCmd) checkSystemRequirements(out io.Writer) error {
	// check arch/os
	if runtime.GOARCH != "amd64" || runtime.GOOS != "linux" {
		return fmt.Errorf("creation of a QEMU based Constellation is not supported for %s/%s, a linux/amd64 platform is required", runtime.GOOS, runtime.GOARCH)
	}

	m.log.Debugf("Checked arch and os")
	// check if /dev/kvm exists
	if _, err := os.Stat("/dev/kvm"); err != nil {
		return fmt.Errorf("unable to access KVM device: %w", err)
	}
	m.log.Debugf("Checked that /dev/kvm exists")
	// check CPU cores
	if runtime.NumCPU() < 4 {
		return fmt.Errorf("insufficient CPU cores: %d, at least 4 cores are required by MiniConstellation", runtime.NumCPU())
	}
	if runtime.NumCPU() < 6 {
		fmt.Fprintf(out, "WARNING: Only %d CPU cores available. This may cause performance issues.\n", runtime.NumCPU())
	}
	m.log.Debugf("Checked CPU cores - there are %d", runtime.NumCPU())

	// check memory
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return fmt.Errorf("determining available memory: failed to open /proc/meminfo: %w", err)
	}
	defer f.Close()
	var memKB int
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), "MemTotal:") {
			_, err = fmt.Sscanf(scanner.Text(), "MemTotal:%d", &memKB)
			if err != nil {
				return fmt.Errorf("determining available memory: failed to parse /proc/meminfo: %w", err)
			}
		}
	}
	m.log.Debugf("Scanned for available memory")
	memGB := memKB / 1024 / 1024
	if memGB < 4 {
		return fmt.Errorf("insufficient memory: %dGB, at least 4GB of memory are required by MiniConstellation", memGB)
	}
	if memGB < 6 {
		fmt.Fprintln(out, "WARNING: Less than 6GB of memory available. This may cause performance issues.")
	}
	m.log.Debugf("Checked available memory, you have %dGB available", memGB)

	var stat unix.Statfs_t
	if err := unix.Statfs(".", &stat); err != nil {
		return err
	}
	freeSpaceGB := stat.Bavail * uint64(stat.Bsize) / 1024 / 1024 / 1024
	if freeSpaceGB < 20 {
		return fmt.Errorf("insufficient disk space: %dGB, at least 20GB of disk space are required by MiniConstellation", freeSpaceGB)
	}
	m.log.Debugf("Checked for free space available, you have %dGB available", freeSpaceGB)

	return nil
}

// prepareConfig reads a given config, or creates a new minimal QEMU config.
func (m *miniUpCmd) prepareConfig(cmd *cobra.Command, fileHandler file.Handler) (*config.Config, error) {
	m.log.Debugf("Preparing configuration")
	configPath, err := cmd.Flags().GetString("config")
	if err != nil {
		return nil, err
	}
	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return nil, fmt.Errorf("parsing force argument: %w", err)
	}

	// check for existing config
	if configPath != "" {
		conf, err := config.New(fileHandler, configPath, force)
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
	m.log.Debugf("Configuration path is %q", configPath)
	if err := cmd.Flags().Set("config", constants.ConfigFilename); err != nil {
		return nil, err
	}
	_, err = fileHandler.Stat(constants.ConfigFilename)
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
	config.RemoveProviderExcept(cloudprovider.QEMU)
	config.StateDiskSizeGB = 8
	m.log.Debugf("Prepared configuration")

	return config, fileHandler.WriteYAML(constants.ConfigFilename, config, file.OptOverwrite)
}

// createMiniCluster creates a new cluster using the given config.
func (m *miniUpCmd) createMiniCluster(ctx context.Context, fileHandler file.Handler, creator cloudCreator, config *config.Config) error {
	m.log.Debugf("Creating mini cluster")
	idFile, err := creator.Create(ctx, cloudprovider.QEMU, config, "", 1, 1)
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
	newDialer := func(validator *cloudcmd.Validator) *dialer.Dialer {
		return dialer.New(nil, validator.V(cmd), &net.Dialer{})
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
	i := &initCmd{log: log}
	if err := i.initialize(cmd, newDialer, fileHandler, license.NewClient(), spinner); err != nil {
		return err
	}
	m.log.Debugf("Initialized mini cluster")
	return nil
}
