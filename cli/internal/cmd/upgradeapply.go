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

	"github.com/edgelesssys/constellation/v2/cli/internal/helm"
	"github.com/edgelesssys/constellation/v2/cli/internal/kubernetes"
	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/compatibility"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
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
		"WARNING: might delete your resources in case you are using cert-manager in your cluster. Please read the docs.\n"+
		"WARNING: might unintentionally overwrite measurements in the running cluster.")
	cmd.Flags().Duration("timeout", 3*time.Minute, "change helm upgrade timeout\n"+
		"Might be useful for slow connections or big clusters.")
	if err := cmd.Flags().MarkHidden("timeout"); err != nil {
		panic(err)
	}

	return cmd
}

func runUpgradeApply(cmd *cobra.Command, _ []string) error {
	log, err := newCLILogger(cmd)
	if err != nil {
		return fmt.Errorf("creating logger: %w", err)
	}
	defer log.Sync()

	fileHandler := file.NewHandler(afero.NewOsFs())
	upgrader, err := kubernetes.NewUpgrader(cmd.OutOrStdout(), log)
	if err != nil {
		return err
	}

	applyCmd := upgradeApplyCmd{upgrader: upgrader, log: log}
	return applyCmd.upgradeApply(cmd, fileHandler)
}

type upgradeApplyCmd struct {
	upgrader cloudUpgrader
	log      debugLog
}

func (u *upgradeApplyCmd) upgradeApply(cmd *cobra.Command, fileHandler file.Handler) error {
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

	if conf.GetProvider() == cloudprovider.Azure || conf.GetProvider() == cloudprovider.GCP {
		err = u.handleServiceUpgrade(cmd, conf, flags)
		upgradeErr := &compatibility.InvalidUpgradeError{}
		switch {
		case errors.As(err, &upgradeErr):
			cmd.PrintErrln(err)
		case err != nil:
			return fmt.Errorf("upgrading services: %w", err)
		}

		err = u.upgrader.UpgradeNodeVersion(cmd.Context(), conf)
		switch {
		case errors.Is(err, kubernetes.ErrInProgress):
			cmd.PrintErrln("Skipping image and Kubernetes upgrades. Another upgrade is in progress.")
		case errors.As(err, &upgradeErr):
			cmd.PrintErrln(err)
		case err != nil:
			return fmt.Errorf("upgrading NodeVersion: %w", err)
		}
	} else {
		cmd.PrintErrln("WARNING: Skipping service and image upgrades, which are currently only supported for Azure and GCP.")
	}

	// If an image upgrade was just executed there won't be a diff. The function will return nil in that case.
	if err := u.upgradeMeasurementsIfDiff(cmd, conf.GetMeasurements(), flags); err != nil {
		return fmt.Errorf("upgrading measurements: %w", err)
	}

	return nil
}

// upgradeMeasurementsIfDiff checks if the locally configured measurements are different from the cluster's measurements.
// If so the function will ask the user to confirm (if --yes is not set) and upgrade the measurements only.
func (u *upgradeApplyCmd) upgradeMeasurementsIfDiff(cmd *cobra.Command, newMeasurements measurements.M, flags upgradeApplyFlags) error {
	clusterMeasurements, _, err := u.upgrader.GetClusterMeasurements(cmd.Context())
	if err != nil {
		return fmt.Errorf("getting cluster measurements: %w", err)
	}
	if clusterMeasurements.EqualTo(newMeasurements) {
		return nil
	}

	if !flags.yes {
		ok, err := askToConfirm(cmd, "You are about to change your cluster's measurements. Are you sure you want to continue?")
		if err != nil {
			return fmt.Errorf("asking for confirmation: %w", err)
		}
		if !ok {
			cmd.Println("Aborting upgrade.")
			return nil
		}
	}
	if err := u.upgrader.UpdateMeasurements(cmd.Context(), newMeasurements); err != nil {
		return fmt.Errorf("updating measurements: %w", err)
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

	return err
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
	UpgradeNodeVersion(ctx context.Context, conf *config.Config) error
	UpgradeHelmServices(ctx context.Context, config *config.Config, timeout time.Duration, allowDestructive bool) error
	UpdateMeasurements(ctx context.Context, newMeasurements measurements.M) error
	GetClusterMeasurements(ctx context.Context) (measurements.M, *corev1.ConfigMap, error)
}
