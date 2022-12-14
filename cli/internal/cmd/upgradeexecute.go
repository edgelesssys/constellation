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
	"github.com/edgelesssys/constellation/v2/cli/internal/helm"
	"github.com/edgelesssys/constellation/v2/cli/internal/image"
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

	cmd.Flags().Bool("helm", false, "Execute helm upgrade. This feature is still in development an may change without anounncement. Upgrades all helm charts deployed during constellation-init.")
	cmd.Flags().BoolP("yes", "y", false, "Run upgrades without further confirmation. WARNING: might delete your resources in case you are using cert-manager in your cluster. Please read the docs.")
	cmd.Flags().Duration("timeout", 3*time.Minute, "Change helm upgrade timeout. This feature is still in development an may change without anounncement. Might be useful for slow connections or big clusters.")
	if err := cmd.Flags().MarkHidden("helm"); err != nil {
		panic(err)
	}
	if err := cmd.Flags().MarkHidden("timeout"); err != nil {
		panic(err)
	}

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
	flags, err := parseUpgradeExecuteFlags(cmd)
	if err != nil {
		return fmt.Errorf("parsing flags: %w", err)
	}
	conf, err := config.New(fileHandler, flags.configPath)
	if err != nil {
		return displayConfigValidationErrors(cmd.ErrOrStderr(), err)
	}

	if flags.helmFlag {
		err = upgrader.UpgradeHelmServices(cmd.Context(), conf, flags.upgradeTimeout, helm.DenyDestructive)
		if errors.Is(err, helm.ErrConfirmationMissing) {
			if !flags.yes {
				cmd.PrintErrln("WARNING: Upgrading cert-manager will destroy all custom resources you have manually created that are based on the current version of cert-manager.")
				ok, askErr := askToConfirm(cmd, "Do you want to upgrade cert-manager anyway?")
				if askErr != nil {
					return fmt.Errorf("asking for confirmation: %w", err)
				}
				if !ok {
					cmd.Println("Aborting upgrade.")
					return nil
				}
			}
			err = upgrader.UpgradeHelmServices(cmd.Context(), conf, flags.upgradeTimeout, helm.AllowDestructive)
		}
		if err != nil {
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

func parseUpgradeExecuteFlags(cmd *cobra.Command) (upgradeExecuteFlags, error) {
	configPath, err := cmd.Flags().GetString("config")
	if err != nil {
		return upgradeExecuteFlags{}, err
	}

	helmFlag, err := cmd.Flags().GetBool("helm")
	if err != nil {
		return upgradeExecuteFlags{}, err
	}

	yes, err := cmd.Flags().GetBool("yes")
	if err != nil {
		return upgradeExecuteFlags{}, err
	}

	timeout, err := cmd.Flags().GetDuration("timeout")
	if err != nil {
		return upgradeExecuteFlags{}, err
	}
	return upgradeExecuteFlags{configPath: configPath, helmFlag: helmFlag, yes: yes, upgradeTimeout: timeout}, nil
}

type upgradeExecuteFlags struct {
	configPath     string
	helmFlag       bool
	yes            bool
	upgradeTimeout time.Duration
}

type cloudUpgrader interface {
	Upgrade(ctx context.Context, imageReference, imageVersion string, measurements measurements.M) error
	UpgradeHelmServices(ctx context.Context, config *config.Config, timeout time.Duration, allowDestructive bool) error
}

type imageFetcher interface {
	FetchReference(ctx context.Context, config *config.Config) (string, error)
}
