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

	"github.com/edgelesssys/constellation/v2/cli/internal/clusterid"
	"github.com/edgelesssys/constellation/v2/cli/internal/helm"
	"github.com/edgelesssys/constellation/v2/cli/internal/kubernetes"
	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/compatibility"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/variant"
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
	upgrader, err := kubernetes.NewUpgrader(cmd.Context(), cmd.OutOrStdout(), log)
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

	var idFile clusterid.File
	if err := fileHandler.ReadJSON(constants.ClusterIDsFileName, &idFile); err != nil {
		return fmt.Errorf("reading cluster ID file: %w", err)
	}
	conf.UpdateMAAURL(idFile.AttestationURL)

	// If an image upgrade was just executed there won't be a diff. The function will return nil in that case.
	if err := u.upgradeAttestConfigIfDiff(cmd, conf.GetAttestationConfig(), flags); err != nil {
		return fmt.Errorf("upgrading measurements: %w", err)
	}


	if conf.GetProvider() == cloudprovider.Azure {
		u.log.Debugf("Planning Terraform migrations")
		hasDiff, err := u.upgrader.PlanTerraformMigrations(cmd.Context(), flags.terraformLogLevel, cloudprovider.Azure)
		if err != nil {
			return fmt.Errorf("planning Terraform migrations: %w", err)
		}
		if hasDiff {
			if !flags.yes {
				ok, err := askToConfirm(cmd, "Do you want to apply the Terraform migrations?")
				if err != nil {
					return fmt.Errorf("asking for confirmation: %w", err)
				}
				if !ok {
					cmd.Println("Aborting upgrade.")
					return nil
				}
			}
			u.log.Debugf("Applying Terraform migrations")
			// TODO: Apply Terraform migrations
		} else {
			u.log.Debugf("No Terraform diff detected")
		}
	}

	if conf.GetProvider() == cloudprovider.Azure || conf.GetProvider() == cloudprovider.GCP || conf.GetProvider() == cloudprovider.AWS {
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
		cmd.PrintErrln("WARNING: Skipping service and image upgrades, which are currently only supported for AWS, Azure, and GCP.")
	}

	return nil
}

// upgradeAttestConfigIfDiff checks if the locally configured measurements are different from the cluster's measurements.
// If so the function will ask the user to confirm (if --yes is not set) and upgrade the measurements only.
func (u *upgradeApplyCmd) upgradeAttestConfigIfDiff(cmd *cobra.Command, newConfig config.AttestationCfg, flags upgradeApplyFlags) error {
	clusterAttestationConfig, _, err := u.upgrader.GetClusterAttestationConfig(cmd.Context(), newConfig.GetVariant())
	// Config migration from v2.7 to v2.8 requires us to skip comparing configs if the cluster is still using the legacy config.
	// TODO: v2.9 Remove error type check and always run comparison.
	if err != nil && !errors.Is(err, kubernetes.ErrLegacyJoinConfig) {
		return fmt.Errorf("getting cluster measurements: %w", err)
	}
	if err == nil {
		// If the current config is equal, or there is an error when comparing the configs, we skip the upgrade.
		if equal, err := newConfig.EqualTo(clusterAttestationConfig); err != nil || equal {
			return err
		}
	}

	if !flags.yes {
		ok, err := askToConfirm(cmd, "You are about to change your cluster's attestation config. Are you sure you want to continue?")
		if err != nil {
			return fmt.Errorf("asking for confirmation: %w", err)
		}
		if !ok {
			cmd.Println("Skipping upgrade.")
			return nil
		}
	}
	if err := u.upgrader.UpdateAttestationConfig(cmd.Context(), newConfig); err != nil {
		return fmt.Errorf("updating attestation config: %w", err)
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
				cmd.Println("Skipping upgrade.")
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

	logLevelString, err := cmd.Flags().GetString("tf-log")
	if err != nil {
		return upgradeApplyFlags{}, fmt.Errorf("parsing tf-log string: %w", err)
	}
	logLevel, err := terraform.ParseLogLevel(logLevelString)
	if err != nil {
		return upgradeApplyFlags{}, fmt.Errorf("parsing Terraform log level %s: %w", logLevelString, err)
	}

	return upgradeApplyFlags{
		configPath:        configPath,
		yes:               yes,
		upgradeTimeout:    timeout,
		force:             force,
		terraformLogLevel: logLevel,
	}, nil
}

type upgradeApplyFlags struct {
	configPath        string
	yes               bool
	upgradeTimeout    time.Duration
	force             bool
	terraformLogLevel terraform.LogLevel
}

type cloudUpgrader interface {
	UpgradeNodeVersion(ctx context.Context, conf *config.Config) error
	UpgradeHelmServices(ctx context.Context, config *config.Config, timeout time.Duration, allowDestructive bool) error
	UpdateAttestationConfig(ctx context.Context, newConfig config.AttestationCfg) error
	GetClusterAttestationConfig(ctx context.Context, variant variant.Variant) (config.AttestationCfg, *corev1.ConfigMap, error)
	UpdateMeasurements(ctx context.Context, newMeasurements measurements.M) error
	GetClusterMeasurements(ctx context.Context) (measurements.M, *corev1.ConfigMap, error)
	PlanTerraformMigrations(ctx context.Context, logLevel terraform.LogLevel, csp cloudprovider.Provider) (bool, error)
}
