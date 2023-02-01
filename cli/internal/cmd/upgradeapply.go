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

func newUpgradeApplyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "execute",
		Short: "Apply an upgrade to a Constellation cluster",
		Long:  "Apply an upgrade to a Constellation cluster by applying the chosen configuration.",
		Args:  cobra.NoArgs,
		RunE:  runUpgradeApply,
	}

	cmd.Flags().BoolP("yes", "y", false, "run upgrades without further confirmation\n"+
		"WARNING: might delete your resources in case you are using cert-manager in your cluster. Please read the docs.")
	cmd.Flags().Duration("timeout", 3*time.Minute, "change helm upgrade timeout\n"+
		"Might be useful for slow connections or big clusters.")
	if err := cmd.Flags().MarkHidden("timeout"); err != nil {
		panic(err)
	}

	return cmd
}

func runUpgradeApply(cmd *cobra.Command, args []string) error {
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

	return upgradeApply(cmd, imageFetcher, upgrader, fileHandler)
}

func upgradeApply(cmd *cobra.Command, imageFetcher imageFetcher, upgrader cloudUpgrader, fileHandler file.Handler) error {
	flags, err := parseUpgradeApplyFlags(cmd)
	if err != nil {
		return fmt.Errorf("parsing flags: %w", err)
	}
	conf, err := config.New(fileHandler, flags.configPath, flags.force)
	if err != nil {
		return config.DisplayValidationErrors(cmd.ErrOrStderr(), err)
	}

	if err := handleServiceUpgrade(cmd, upgrader, conf, flags); err != nil {
		return err
	}

	// TODO: validate upgrade config? Should be basic things like checking image is not an empty string
	// More sophisticated validation, like making sure we don't downgrade the cluster, should be done by `constellation upgrade plan`

	return handleImageUpgrade(cmd.Context(), conf, imageFetcher, upgrader)
}

func handleServiceUpgrade(cmd *cobra.Command, upgrader cloudUpgrader, conf *config.Config, flags upgradeApplyFlags) error {
	err := upgrader.UpgradeHelmServices(cmd.Context(), conf, flags.upgradeTimeout, helm.DenyDestructive)
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

func handleImageUpgrade(ctx context.Context, conf *config.Config, imageFetcher imageFetcher, upgrader cloudUpgrader) error {
	// this config modification is temporary until we can remove the upgrade section from the config
	conf.Image = conf.Upgrade.Image
	imageReference, err := imageFetcher.FetchReference(ctx, conf)
	if err != nil {
		return err
	}

	return upgrader.UpgradeImage(ctx, imageReference, conf.Upgrade.Image, conf.Upgrade.Measurements)
}

func parseUpgradeApplyFlags(cmd *cobra.Command) (upgradeApplyFlags, error) {
	configPath, err := cmd.Flags().GetString("config")
	if err != nil {
		return upgradeApplyFlags{}, err
	}

	yes, err := cmd.Flags().GetBool("yes")
	if err != nil {
		return upgradeApplyFlags{}, err
	}

	timeout, err := cmd.Flags().GetDuration("timeout")
	if err != nil {
		return upgradeApplyFlags{}, err
	}

	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return upgradeApplyFlags{}, fmt.Errorf("parsing force argument: %w", err)
	}

	return upgradeApplyFlags{configPath: configPath, yes: yes, upgradeTimeout: timeout, force: force}, nil
}

type upgradeApplyFlags struct {
	configPath     string
	yes            bool
	upgradeTimeout time.Duration
	force          bool
}

type cloudUpgrader interface {
	UpgradeImage(ctx context.Context, imageReference, imageVersion string, measurements measurements.M) error
	UpgradeHelmServices(ctx context.Context, config *config.Config, timeout time.Duration, allowDestructive bool) error
}

type imageFetcher interface {
	FetchReference(ctx context.Context, config *config.Config) (string, error)
}
