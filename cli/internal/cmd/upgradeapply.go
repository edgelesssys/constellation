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
	"github.com/edgelesssys/constellation/v2/internal/versions"
	"github.com/edgelesssys/constellation/v2/internal/versions/components"
	"github.com/edgelesssys/constellation/v2/internal/versionsapi"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func newUpgradeApplyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply",
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

	applyCmd := upgradeApplyCmd{upgrader: upgrader, log: log}
	return applyCmd.upgradeApply(cmd, imageFetcher, fileHandler)
}

type upgradeApplyCmd struct {
	upgrader cloudUpgrader
	log      debugLog
}

func (u *upgradeApplyCmd) upgradeApply(cmd *cobra.Command, imageFetcher imageFetcher, fileHandler file.Handler) error {
	flags, err := parseUpgradeApplyFlags(cmd)
	if err != nil {
		return fmt.Errorf("parsing flags: %w", err)
	}
	conf, err := config.New(fileHandler, flags.configPath, flags.force)
	var configValidationErr *config.ValidationError
	if errors.As(err, &configValidationErr) {
		cmd.PrintErrln(configValidationErr.LongMessage())
	}
	if err != nil {
		return err
	}

	if err := u.handleServiceUpgrade(cmd, conf, flags); err != nil {
		return fmt.Errorf("service upgrade: %w", err)
	}

	invalidUpgradeErr := &cloudcmd.InvalidUpgradeError{}
	err = u.handleK8sUpgrade(cmd.Context(), conf)
	skipCtr := 0
	switch {
	case errors.Is(err, cloudcmd.ErrInProgress):
		skipCtr = skipCtr + 1
		cmd.PrintErrln("Skipping Kubernetes components upgrades. Another Kubernetes components upgrade is in progress")
	case errors.As(err, &invalidUpgradeErr):
		skipCtr = skipCtr + 1
		cmd.PrintErrf("Skipping Kubernetes components upgrades: %s\n", err)
	case err != nil:
		return fmt.Errorf("upgrading Kubernetes components: %w", err)
	}

	err = u.handleImageUpgrade(cmd.Context(), conf, imageFetcher)
	switch {
	case errors.Is(err, cloudcmd.ErrInProgress):
		skipCtr = skipCtr + 1
		cmd.PrintErrln("Skipping image upgrades. Another image upgrade is in progress")
	case errors.As(err, &invalidUpgradeErr):
		skipCtr = skipCtr + 1
		cmd.PrintErrf("Skipping image upgrades: %s\n", err)
	case err != nil:
		return fmt.Errorf("upgrading image: %w", err)
	}

	if skipCtr < 2 {
		fmt.Printf("Nodes will restart automatically\n")
	}

	return nil
}

func (u *upgradeApplyCmd) handleServiceUpgrade(cmd *cobra.Command, conf *config.Config, flags upgradeApplyFlags) error {
	err := u.upgrader.UpgradeHelmServices(cmd.Context(), conf, flags.upgradeTimeout, helm.DenyDestructive)
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
		err = u.upgrader.UpgradeHelmServices(cmd.Context(), conf, flags.upgradeTimeout, helm.AllowDestructive)
	}
	if err != nil {
		return fmt.Errorf("upgrading helm: %w", err)
	}

	return nil
}

func (u *upgradeApplyCmd) handleImageUpgrade(ctx context.Context, conf *config.Config, imageFetcher imageFetcher) error {
	imageReference, err := imageFetcher.FetchReference(ctx, conf)
	if err != nil {
		return fmt.Errorf("fetching image reference: %w", err)
	}

	imageVersion, err := versionsapi.NewVersionFromShortPath(conf.Image, versionsapi.VersionKindImage)
	if err != nil {
		return fmt.Errorf("parsing version from image short path: %w", err)
	}

	err = u.upgrader.UpgradeImage(ctx, imageReference, imageVersion.Version, conf.GetMeasurements())
	if err != nil {
		return fmt.Errorf("upgrading image: %w", err)
	}

	return nil
}

func (u *upgradeApplyCmd) handleK8sUpgrade(ctx context.Context, conf *config.Config) error {
	currentVersion, err := versions.NewValidK8sVersion(conf.KubernetesVersion)
	if err != nil {
		return fmt.Errorf("getting Kubernetes version: %w", err)
	}
	versionConfig := versions.VersionConfigs[currentVersion]

	err = u.upgrader.UpgradeK8s(ctx, versionConfig.ClusterVersion, versionConfig.KubernetesComponents)
	if err != nil {
		return fmt.Errorf("upgrading Kubernetes: %w", err)
	}

	return nil
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
	UpgradeK8s(ctx context.Context, clusterVersion string, components components.Components) error
}

type imageFetcher interface {
	FetchReference(ctx context.Context, config *config.Config) (string, error)
}
