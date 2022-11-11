/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"context"

	"github.com/edgelesssys/constellation/v2/cli/internal/cloudcmd"
	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func newUpgradeExecuteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "execute",
		Short: "Execute an upgrade of a Constellation cluster",
		Long:  "Execute an upgrade of a Constellation cluster by applying the chosen configuration.",
		Args:  cobra.NoArgs,
		RunE:  runUpgradeExecute,
	}

	return cmd
}

func runUpgradeExecute(cmd *cobra.Command, args []string) error {
	fileHandler := file.NewHandler(afero.NewOsFs())
	upgrader, err := cloudcmd.NewUpgrader(cmd.OutOrStdout())
	if err != nil {
		return err
	}

	return upgradeExecute(cmd, upgrader, fileHandler)
}

func upgradeExecute(cmd *cobra.Command, upgrader cloudUpgrader, fileHandler file.Handler) error {
	configPath, err := cmd.Flags().GetString("config")
	if err != nil {
		return err
	}
	config, err := config.New(config.WithDefaultOptions(fileHandler, configPath)...)
	if err != nil {
		return displayConfigValidationErrors(cmd.ErrOrStderr(), err)
	}

	// TODO: validate upgrade config? Should be basic things like checking image is not an empty string
	// More sophisticated validation, like making sure we don't downgrade the cluster, should be done by `constellation upgrade plan`

	return upgrader.Upgrade(cmd.Context(), config.Upgrade.Image, config.Upgrade.Measurements)
}

type cloudUpgrader interface {
	Upgrade(ctx context.Context, image string, measurements measurements.Measurements) error
}
