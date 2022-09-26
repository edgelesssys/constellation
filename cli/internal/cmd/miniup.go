/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"context"
	"errors"
	"net"

	"github.com/edgelesssys/constellation/v2/cli/internal/cloudcmd"
	"github.com/edgelesssys/constellation/v2/cli/internal/helm"
	"github.com/edgelesssys/constellation/v2/cli/internal/libvirt"
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
		Short: "Create and initialize a new mini Constellation cluster",
		Long: "Create and initialize a new mini Constellation cluster.\n" +
			"A mini cluster consists of a single control-plane and worker node, hosted using QEMU/KVM.\n",
		Args: cobra.ExactArgs(0),
		PreRun: func(cmd *cobra.Command, args []string) {
			// mark image flag as required, if no config is specified
			configPath, err := cmd.Flags().GetString("config")
			if err != nil {
				panic(err)
			}
			if configPath == "" {
				must(cmd.MarkFlagRequired("image"))
			}
		},
		RunE: runUp,
	}

	cmd.Flags().StringP("image", "i", "", "path to the image to use for the cluster")
	// override global flag so we don't have a default value for the config
	cmd.Flags().String("config", "", "path to the config file to use for the cluster")

	return cmd
}

func runUp(cmd *cobra.Command, args []string) error {
	fileHandler := file.NewHandler(afero.NewOsFs())

	// create config if not passed as flag and set default values
	config, err := prepareConfig(cmd, fileHandler)
	if err != nil {
		return err
	}

	// create cluster
	cmd.Println("Creating cluster in QEMU ...")
	if err := createMiniCluster(cmd.Context(), fileHandler, cloudcmd.NewCreator(cmd.OutOrStdout()), config); err != nil {
		return err
	}
	cmd.Println("Cluster successfully created")
	connectURI := config.Provider.QEMU.LibvirtURI
	if connectURI == "" {
		connectURI = libvirt.LibvirtTCPConnectURI
	}
	cmd.Printf("Connect to the VMs using: \"virsh -c %s\"\n", connectURI)

	// initialize cluster
	return initializeMiniCluster(cmd, fileHandler)
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
			return nil, errors.New("invalid provider for mini constellation cluster")
		}
		return config, nil
	}

	imagePath, err := cmd.Flags().GetString("image")
	if err != nil {
		return nil, err
	}

	config := config.Default()
	config.RemoveProviderExcept(cloudprovider.QEMU)
	config.StateDiskSizeGB = 10

	config.Provider.QEMU.Image = imagePath

	return config, fileHandler.WriteYAML(constants.ConfigFilename, config)
}

// createMiniCluster creates a new cluster using the given config.
func createMiniCluster(ctx context.Context, fileHandler file.Handler, creator cloudCreator, config *config.Config) error {
	state, err := creator.Create(ctx, cloudprovider.QEMU, config, "mini", "", 1, 1)
	if err != nil {
		return err
	}
	if err := fileHandler.WriteJSON(constants.StateFilename, state); err != nil {
		return err
	}

	return writeIPtoIDFile(fileHandler, state)
}

// initializeMiniCluster initializes a QEMU cluster.
func initializeMiniCluster(cmd *cobra.Command, fileHandler file.Handler) error {
	newDialer := func(validator *cloudcmd.Validator) *dialer.Dialer {
		return dialer.New(nil, validator.V(cmd), &net.Dialer{})
	}
	helmLoader := &helm.ChartLoader{}

	cmd.Flags().String("master-secret", "", "")
	cmd.Flags().String("endpoint", "", "")
	cmd.Flags().Bool("conformance", false, "")

	if err := initialize(cmd, newDialer, fileHandler, helmLoader, license.NewClient()); err != nil {
		return err
	}
	return nil
}
