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
	"net/http"
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
	"github.com/edgelesssys/constellation/v2/internal/versions"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"golang.org/x/sys/unix"
)

func newMiniUpCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "up",
		Short: "Create and initialize a new MiniConstellation cluster",
		Long: "Create and initialize a new MiniConstellation cluster.\n" +
			"A mini cluster consists of a single control-plane and worker node, hosted using QEMU/KVM.\n",
		Args: cobra.ExactArgs(0),
		RunE: runUp,
	}

	// override global flag so we don't have a default value for the config
	cmd.Flags().String("config", "", "path to the config file to use for the cluster")

	return cmd
}

func runUp(cmd *cobra.Command, args []string) error {
	spinner := newSpinner(cmd.OutOrStdout())
	defer spinner.Stop()
	creator := cloudcmd.NewCreator(spinner)

	return up(cmd, creator, spinner)
}

func up(cmd *cobra.Command, creator cloudCreator, spinner spinnerInterf) error {
	if err := checkSystemRequirements(cmd.OutOrStdout()); err != nil {
		return fmt.Errorf("system requirements not met: %w", err)
	}

	fileHandler := file.NewHandler(afero.NewOsFs())

	// create config if not passed as flag and set default values
	config, err := prepareConfig(cmd, fileHandler)
	if err != nil {
		return fmt.Errorf("preparing config: %w", err)
	}

	// create cluster
	spinner.Start("Creating cluster in QEMU ", false)
	err = createMiniCluster(cmd.Context(), fileHandler, creator, config)
	spinner.Stop()
	if err != nil {
		return fmt.Errorf("creating cluster: %w", err)
	}
	cmd.Println("Cluster successfully created.")
	connectURI := config.Provider.QEMU.LibvirtURI
	if connectURI == "" {
		connectURI = libvirt.LibvirtTCPConnectURI
	}
	cmd.Println("Connect to the VMs by executing:")
	cmd.Printf("\tvirsh -c %s\n\n", connectURI)

	// initialize cluster
	if err := initializeMiniCluster(cmd, fileHandler, spinner); err != nil {
		return fmt.Errorf("initializing cluster: %w", err)
	}
	return nil
}

// checkSystemRequirements checks if the system meets the requirements for running a MiniConstellation cluster.
// We do so by verifying that the host:
// - arch/os is linux/amd64.
// - has access to /dev/kvm.
// - has at least 4 CPU cores.
// - has at least 4GB of memory.
// - has at least 20GB of free disk space.
func checkSystemRequirements(out io.Writer) error {
	// check arch/os
	if runtime.GOARCH != "amd64" || runtime.GOOS != "linux" {
		return fmt.Errorf("creation of a QEMU based Constellation is not supported for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	// check if /dev/kvm exists
	if _, err := os.Stat("/dev/kvm"); err != nil {
		return fmt.Errorf("unable to access KVM device: %w", err)
	}

	// check CPU cores
	if runtime.NumCPU() < 4 {
		return fmt.Errorf("insufficient CPU cores: %d, at least 4 cores are required by MiniConstellation", runtime.NumCPU())
	}
	if runtime.NumCPU() < 6 {
		fmt.Fprintf(out, "WARNING: Only %d CPU cores available. This may cause performance issues.\n", runtime.NumCPU())
	}

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
	memGB := memKB / 1024 / 1024
	if memGB < 4 {
		return fmt.Errorf("insufficient memory: %dGB, at least 4GB of memory are required by MiniConstellation", memGB)
	}
	if memGB < 6 {
		fmt.Fprintln(out, "WARNING: Less than 6GB of memory available. This may cause performance issues.")
	}

	var stat unix.Statfs_t
	if err := unix.Statfs(".", &stat); err != nil {
		return err
	}
	freeSpaceGB := stat.Bavail * uint64(stat.Bsize) / 1024 / 1024 / 1024
	if freeSpaceGB < 20 {
		return fmt.Errorf("insufficient disk space: %dGB, at least 20GB of disk space are required by MiniConstellation", freeSpaceGB)
	}

	return nil
}

// prepareConfig reads a given config, or creates a new minimal QEMU config.
func prepareConfig(cmd *cobra.Command, fileHandler file.Handler) (*config.Config, error) {
	configPath, err := cmd.Flags().GetString("config")
	if err != nil {
		return nil, err
	}

	// check for existing config
	if configPath != "" {
		config, err := readConfig(cmd.OutOrStdout(), fileHandler, configPath)
		if err != nil {
			return nil, err
		}
		if config.GetProvider() != cloudprovider.QEMU {
			return nil, errors.New("invalid provider for MiniConstellation cluster")
		}
		return config, nil
	}
	if err := cmd.Flags().Set("config", constants.ConfigFilename); err != nil {
		return nil, err
	}
	_, err = fileHandler.Stat(constants.ConfigFilename)
	if err == nil {
		// config already exists, prompt user to overwrite
		cmd.Println("A config file already exists in the current workspace.")
		ok, err := askToConfirm(cmd, "Do you want to overwrite it?")
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, errors.New("not overwriting existing config")
		}
	}

	// download image to current directory if it doesn't exist
	const imagePath = "./constellation.qcow2"
	if _, err := os.Stat(imagePath); err == nil {
		cmd.Printf("Using existing image at %s\n\n", imagePath)
	} else if errors.Is(err, os.ErrNotExist) {
		cmd.Printf("Downloading image to %s\n", imagePath)
		if err := installImage(cmd.Context(), cmd.OutOrStdout(), versions.ConstellationQEMUImageURL, imagePath); err != nil {
			return nil, fmt.Errorf("downloading image to %s: %w", imagePath, err)
		}
	} else {
		return nil, fmt.Errorf("checking if image exists: %w", err)
	}

	config := config.Default()
	config.RemoveProviderExcept(cloudprovider.QEMU)
	config.StateDiskSizeGB = 8
	config.Provider.QEMU.Image = imagePath

	return config, fileHandler.WriteYAML(constants.ConfigFilename, config, file.OptOverwrite)
}

// createMiniCluster creates a new cluster using the given config.
func createMiniCluster(ctx context.Context, fileHandler file.Handler, creator cloudCreator, config *config.Config) error {
	idFile, err := creator.Create(ctx, cloudprovider.QEMU, config, "mini", "", 1, 1)
	if err != nil {
		return err
	}

	idFile.UID = "mini" // use UID "mini" to identify MiniConstellation clusters.

	return fileHandler.WriteJSON(constants.ClusterIDsFileName, idFile, file.OptNone)
}

// initializeMiniCluster initializes a QEMU cluster.
func initializeMiniCluster(cmd *cobra.Command, fileHandler file.Handler, spinner spinnerInterf) (retErr error) {
	// clean up cluster resources if initialization fails
	defer func() {
		if retErr != nil {
			cmd.Printf("An error occurred: %s\n", retErr)
			cmd.Println("Attempting to roll back.")
			_ = runDown(cmd, []string{})
			cmd.Printf("Rollback succeeded.\n\n")
		}
	}()
	newDialer := func(validator *cloudcmd.Validator) *dialer.Dialer {
		return dialer.New(nil, validator.V(cmd), &net.Dialer{})
	}

	cmd.Flags().String("master-secret", "", "")
	cmd.Flags().String("endpoint", "", "")
	cmd.Flags().Bool("conformance", false, "")

	if err := initialize(cmd, newDialer, fileHandler, license.NewClient(), spinner); err != nil {
		return err
	}
	return nil
}

// installImage downloads the image from sourceURL to the destination.
func installImage(ctx context.Context, out io.Writer, sourceURL, destination string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, sourceURL, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("downloading image: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("downloading image: %s", resp.Status)
	}

	f, err := os.OpenFile(destination, os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	bar := progressbar.NewOptions64(
		resp.ContentLength,
		progressbar.OptionSetWriter(out),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
		progressbar.OptionClearOnFinish(),
		progressbar.OptionOnCompletion(func() { fmt.Fprintf(out, "Done.\n\n") }),
	)
	defer bar.Close()

	_, err = io.Copy(io.MultiWriter(f, bar), resp.Body)
	if err != nil {
		return err
	}

	return nil
}
