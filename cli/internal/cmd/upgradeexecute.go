/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/edgelesssys/constellation/v2/cli/internal/cloudcmd"
	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/image"
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

	// TODO: remove before merging to main. For testing only.
	cmd.Flags().Bool("helm", false, "DEV ONLY: execute helm upgrade")
	cmd.Flags().Duration("timeout", 3*time.Minute, "DEV ONLY: change upgrade timeout. Might be useful for slow connections or big clusters.")

	return cmd
}

func runUpgradeExecute(cmd *cobra.Command, args []string) error {
	log, err := newCLILogger(cmd)
	if err != nil {
		return fmt.Errorf("creating logger: %w", err)
	}
	defer log.Sync()

	fileHandler := file.NewHandler(afero.NewOsFs())
	imageFetcher := image.New()
	upgrader, err := cloudcmd.NewUpgrader(cmd.OutOrStdout(), log)
	if err != nil {
		return err
	}

	return upgradeExecute(cmd, imageFetcher, upgrader, fileHandler)
}

func upgradeExecute(cmd *cobra.Command, imageFetcher imageFetcher, upgrader cloudUpgrader, fileHandler file.Handler) error {
	configPath, err := cmd.Flags().GetString("config")
	if err != nil {
		return err
	}
	conf, err := config.New(fileHandler, configPath)
	if err != nil {
		return displayConfigValidationErrors(cmd.ErrOrStderr(), err)
	}

	helm, err := cmd.Flags().GetBool("helm")
	if err != nil {
		return err
	}
	if helm {
		timeout, err := cmd.Flags().GetDuration("timeout")
		if err != nil {
			return err
		}
		if err := upgrader.UpgradeHelmServices(cmd.Context(), conf, timeout); err != nil {
			return fmt.Errorf("upgrading helm: %w", err)
		}
		return nil
	}

	// TODO: validate upgrade config? Should be basic things like checking image is not an empty string
	// More sophisticated validation, like making sure we don't downgrade the cluster, should be done by `constellation upgrade plan`

	// this config modification is temporary until we can remove the upgrade section from the config
	conf.Image = conf.Upgrade.Image
	imageReference, err := imageFetcher.FetchReference(cmd.Context(), conf)
	if err != nil {
		return err
	}

	return upgrader.Upgrade(cmd.Context(), imageReference, conf.Upgrade.Image, conf.Upgrade.Measurements)
}

type cloudUpgrader interface {
	Upgrade(ctx context.Context, imageReference, imageVersion string, measurements measurements.M) error
	UpgradeHelmServices(ctx context.Context, config *config.Config, timeout time.Duration) error
}

type imageFetcher interface {
	FetchReference(ctx context.Context, config *config.Config) (string, error)
}
