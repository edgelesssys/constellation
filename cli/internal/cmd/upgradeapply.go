/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/edgelesssys/constellation/v2/cli/internal/clusterid"
	"github.com/edgelesssys/constellation/v2/cli/internal/helm"
	"github.com/edgelesssys/constellation/v2/cli/internal/kubernetes"
	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/cli/internal/upgrade"
	"github.com/edgelesssys/constellation/v2/internal/api/attestationconfigapi"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/compatibility"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/imagefetcher"
	"github.com/edgelesssys/constellation/v2/internal/role"
	"github.com/edgelesssys/constellation/v2/internal/versions"
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
	cmd.Flags().Duration("timeout", 5*time.Minute, "change helm upgrade timeout\n"+
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
	upgradeID := kubernetes.NewUpgradeID()
	upgrader, err := kubernetes.NewUpgrader(cmd.Context(), cmd.OutOrStdout(), fileHandler, log, kubernetes.UpgradeCmdKindApply, upgradeID)
	if err != nil {
		return err
	}

	imagefetcher := imagefetcher.New()
	configFetcher := attestationconfigapi.NewFetcher()

	applyCmd := upgradeApplyCmd{upgrader: upgrader, log: log, imageFetcher: imagefetcher, configFetcher: configFetcher, migrationExecutor: &migrationCmdExecutor{log}}
	iamMigrateCmd, err := upgrade.NewIAMMigrateCmd(cmd.Context(), upgradeID.String(), cloudprovider.AWS, terraform.LogLevelDebug)
	if err != nil {
		return fmt.Errorf("setting up IAM migration command: %w", err)
	}
	applyCmd.migrationCmds = []upgrade.MigrationCmd{iamMigrateCmd}
	return applyCmd.upgradeApply(cmd, fileHandler)
}

type upgradeApplyCmd struct {
	upgrader          cloudUpgrader
	imageFetcher      imageFetcher
	configFetcher     attestationconfigapi.Fetcher
	log               debugLog
	migrationExecutor migrationExecutor
	migrationCmds     []upgrade.MigrationCmd
}

func (u *upgradeApplyCmd) upgradeApply(cmd *cobra.Command, fileHandler file.Handler) error {
	flags, err := parseUpgradeApplyFlags(cmd)
	if err != nil {
		return fmt.Errorf("parsing flags: %w", err)
	}
	conf, err := config.New(fileHandler, flags.configPath, u.configFetcher, flags.force)
	var configValidationErr *config.ValidationError
	if errors.As(err, &configValidationErr) {
		cmd.PrintErrln(configValidationErr.LongMessage())
	}
	if err != nil {
		return err
	}

	if err := handleInvalidK8sPatchVersion(cmd, conf.KubernetesVersion, flags.yes); err != nil {
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
	for _, migrationCmd := range u.migrationCmds {
		if err := u.migrationExecutor.executeMigration(cmd, fileHandler, migrationCmd, flags); err != nil {
			return fmt.Errorf("executing %s migration: %w", migrationCmd.String(), err)
		}
	}
	// not moving existing Terraform migrator because of planned apply refactor
	if err := u.migrateTerraform(cmd, u.imageFetcher, conf, flags); err != nil {
		return fmt.Errorf("performing Terraform migrations: %w", err)
	}

	if conf.GetProvider() == cloudprovider.Azure || conf.GetProvider() == cloudprovider.GCP || conf.GetProvider() == cloudprovider.AWS {
		var upgradeErr *compatibility.InvalidUpgradeError
		err = u.handleServiceUpgrade(cmd, conf, idFile, flags)
		switch {
		case errors.As(err, &upgradeErr):
			cmd.PrintErrln(err)
		case err != nil:
			return fmt.Errorf("upgrading services: %w", err)
		}

		err = u.upgrader.UpgradeNodeVersion(cmd.Context(), conf, flags.force)
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

// migrateTerraform checks if the Constellation version the cluster is being upgraded to requires a migration
// of cloud resources with Terraform. If so, the migration is performed.
func (u *upgradeApplyCmd) migrateTerraform(cmd *cobra.Command, fetcher imageFetcher, conf *config.Config, flags upgradeApplyFlags) error {
	u.log.Debugf("Planning Terraform migrations")

	if err := u.upgrader.CheckTerraformMigrations(); err != nil {
		return fmt.Errorf("checking workspace: %w", err)
	}

	// TODO(AB#3248): Remove this migration after we can assume that all existing clusters have been migrated.
	var awsZone string
	if conf.GetProvider() == cloudprovider.AWS {
		awsZone = conf.Provider.AWS.Zone
	}
	manualMigrations := terraformMigrationAWSNodeGroups(conf.GetProvider(), awsZone)
	for _, migration := range manualMigrations {
		u.log.Debugf("Adding manual Terraform migration: %s", migration.DisplayName)
		u.upgrader.AddManualStateMigration(migration)
	}

	vars, err := parseTerraformUpgradeVars(cmd, conf, fetcher)
	if err != nil {
		return fmt.Errorf("parsing upgrade variables: %w", err)
	}
	u.log.Debugf("Using Terraform variables:\n%v", vars)

	opts := upgrade.TerraformUpgradeOptions{
		LogLevel: flags.terraformLogLevel,
		CSP:      conf.GetProvider(),
		Vars:     vars,
	}

	// Check if there are any Terraform migrations to apply
	hasDiff, err := u.upgrader.PlanTerraformMigrations(cmd.Context(), opts)
	if err != nil {
		return fmt.Errorf("planning terraform migrations: %w", err)
	}

	if hasDiff {
		// If there are any Terraform migrations to apply, ask for confirmation
		fmt.Fprintln(cmd.OutOrStdout(), "The upgrade requires a migration of Constellation cloud resources by applying an updated Terraform template. Please manually review the suggested changes below.")
		if !flags.yes {
			ok, err := askToConfirm(cmd, "Do you want to apply the Terraform migrations?")
			if err != nil {
				return fmt.Errorf("asking for confirmation: %w", err)
			}
			if !ok {
				cmd.Println("Aborting upgrade.")
				if err := u.upgrader.CleanUpTerraformMigrations(); err != nil {
					return fmt.Errorf("cleaning up workspace: %w", err)
				}
				return fmt.Errorf("aborted by user")
			}
		}
		u.log.Debugf("Applying Terraform migrations")
		err := u.upgrader.ApplyTerraformMigrations(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("applying terraform migrations: %w", err)
		}
		cmd.Printf("Terraform migrations applied successfully and output written to: %s\n"+
			"A backup of the pre-upgrade state has been written to: %s\n",
			constants.ClusterIDsFileName, filepath.Join(constants.UpgradeDir, constants.TerraformUpgradeBackupDir))
	} else {
		u.log.Debugf("No Terraform diff detected")
	}

	return nil
}

// parseTerraformUpgradeVars parses the variables required to execute the Terraform script with.
func parseTerraformUpgradeVars(cmd *cobra.Command, conf *config.Config, fetcher imageFetcher) (terraform.Variables, error) {
	// Fetch variables to execute Terraform script with
	provider := conf.GetProvider()
	attestationVariant := conf.GetAttestationConfig().GetVariant()
	region := conf.GetRegion()
	imageRef, err := fetcher.FetchReference(cmd.Context(), provider, attestationVariant, conf.Image, region)
	if err != nil {
		return nil, fmt.Errorf("fetching image reference: %w", err)
	}

	switch conf.GetProvider() {
	case cloudprovider.AWS:
		vars := &terraform.AWSClusterVariables{
			Name: conf.Name,
			NodeGroups: map[string]terraform.AWSNodeGroup{
				"control_plane_default": {
					Role:            role.ControlPlane.TFString(),
					StateDiskSizeGB: conf.StateDiskSizeGB,
					Zone:            conf.Provider.AWS.Zone,
					InstanceType:    conf.Provider.AWS.InstanceType,
					DiskType:        conf.Provider.AWS.StateDiskType,
				},
				"worker_default": {
					Role:            role.Worker.TFString(),
					StateDiskSizeGB: conf.StateDiskSizeGB,
					Zone:            conf.Provider.AWS.Zone,
					InstanceType:    conf.Provider.AWS.InstanceType,
					DiskType:        conf.Provider.AWS.StateDiskType,
				},
			},
			Region:                 conf.Provider.AWS.Region,
			Zone:                   conf.Provider.AWS.Zone,
			AMIImageID:             imageRef,
			IAMProfileControlPlane: conf.Provider.AWS.IAMProfileControlPlane,
			IAMProfileWorkerNodes:  conf.Provider.AWS.IAMProfileWorkerNodes,
			Debug:                  conf.IsDebugCluster(),
		}
		return vars, nil
	case cloudprovider.Azure:
		// Azure Terraform provider is very strict about it's casing
		imageRef = strings.Replace(imageRef, "CommunityGalleries", "communityGalleries", 1)
		imageRef = strings.Replace(imageRef, "Images", "images", 1)
		imageRef = strings.Replace(imageRef, "Versions", "versions", 1)

		vars := &terraform.AzureClusterVariables{
			Name: conf.Name,
			NodeGroups: map[string]terraform.AzureNodeGroup{
				"control_plane_default": {
					Role:         "control-plane",
					InstanceType: conf.Provider.Azure.InstanceType,
					DiskSizeGB:   conf.StateDiskSizeGB,
					DiskType:     conf.Provider.Azure.StateDiskType,
				},
				"worker_default": {
					Role:         "worker",
					InstanceType: conf.Provider.Azure.InstanceType,
					DiskSizeGB:   conf.StateDiskSizeGB,
					DiskType:     conf.Provider.Azure.StateDiskType,
				},
			},
			Location:             conf.Provider.Azure.Location,
			ImageID:              imageRef,
			CreateMAA:            toPtr(conf.GetAttestationConfig().GetVariant().Equal(variant.AzureSEVSNP{})),
			Debug:                toPtr(conf.IsDebugCluster()),
			ConfidentialVM:       toPtr(conf.GetAttestationConfig().GetVariant().Equal(variant.AzureSEVSNP{})),
			SecureBoot:           conf.Provider.Azure.SecureBoot,
			UserAssignedIdentity: conf.Provider.Azure.UserAssignedIdentity,
			ResourceGroup:        conf.Provider.Azure.ResourceGroup,
		}
		return vars, nil
	case cloudprovider.GCP:
		vars := &terraform.GCPClusterVariables{
			Name: conf.Name,
			NodeGroups: map[string]terraform.GCPNodeGroup{
				"control_plane_default": {
					Role:            role.ControlPlane.TFString(),
					StateDiskSizeGB: conf.StateDiskSizeGB,
					Zone:            conf.Provider.GCP.Zone,
					InstanceType:    conf.Provider.GCP.InstanceType,
					DiskType:        conf.Provider.GCP.StateDiskType,
				},
				"worker_default": {
					Role:            role.Worker.TFString(),
					StateDiskSizeGB: conf.StateDiskSizeGB,
					Zone:            conf.Provider.GCP.Zone,
					InstanceType:    conf.Provider.GCP.InstanceType,
					DiskType:        conf.Provider.GCP.StateDiskType,
				},
			},
			Project: conf.Provider.GCP.Project,
			Region:  conf.Provider.GCP.Region,
			Zone:    conf.Provider.GCP.Zone,
			ImageID: imageRef,
			Debug:   conf.IsDebugCluster(),
		}
		return vars, nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", conf.GetProvider())
	}
}

// handleInvalidK8sPatchVersion checks if the Kubernetes patch version is supported and asks for confirmation if not.
func handleInvalidK8sPatchVersion(cmd *cobra.Command, version string, yes bool) error {
	_, err := versions.NewValidK8sVersion(version, true)
	valid := err == nil

	if !valid && !yes {
		confirmed, err := askToConfirm(cmd, fmt.Sprintf("WARNING: The Kubernetes patch version %s is not supported. If you continue, Kubernetes upgrades will be skipped. Do you want to continue anyway?", version))
		if err != nil {
			return fmt.Errorf("asking for confirmation: %w", err)
		}
		if !confirmed {
			return fmt.Errorf("aborted by user")
		}
	}

	return nil
}

type imageFetcher interface {
	FetchReference(ctx context.Context,
		provider cloudprovider.Provider, attestationVariant variant.Variant,
		image, region string,
	) (string, error)
}

// upgradeAttestConfigIfDiff checks if the locally configured measurements are different from the cluster's measurements.
// If so the function will ask the user to confirm (if --yes is not set) and upgrade the measurements only.
func (u *upgradeApplyCmd) upgradeAttestConfigIfDiff(cmd *cobra.Command, newConfig config.AttestationCfg, flags upgradeApplyFlags) error {
	clusterAttestationConfig, _, err := u.upgrader.GetClusterAttestationConfig(cmd.Context(), newConfig.GetVariant())
	if err != nil {
		return fmt.Errorf("getting cluster attestation config: %w", err)
	}
	// If the current config is equal, or there is an error when comparing the configs, we skip the upgrade.
	equal, err := newConfig.EqualTo(clusterAttestationConfig)
	if err != nil {
		return fmt.Errorf("comparing attestation configs: %w", err)
	}
	if equal {
		return nil
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

func (u *upgradeApplyCmd) handleServiceUpgrade(cmd *cobra.Command, conf *config.Config, idFile clusterid.File, flags upgradeApplyFlags) error {
	err := u.upgrader.UpgradeHelmServices(cmd.Context(), conf, idFile, flags.upgradeTimeout, helm.DenyDestructive, flags.force)
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
		err = u.upgrader.UpgradeHelmServices(cmd.Context(), conf, idFile, flags.upgradeTimeout, helm.AllowDestructive, flags.force)
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
	UpgradeNodeVersion(ctx context.Context, conf *config.Config, force bool) error
	UpgradeHelmServices(ctx context.Context, config *config.Config, idFile clusterid.File, timeout time.Duration, allowDestructive bool, force bool) error
	UpdateAttestationConfig(ctx context.Context, newConfig config.AttestationCfg) error
	GetClusterAttestationConfig(ctx context.Context, variant variant.Variant) (config.AttestationCfg, *corev1.ConfigMap, error)
	PlanTerraformMigrations(ctx context.Context, opts upgrade.TerraformUpgradeOptions) (bool, error)
	ApplyTerraformMigrations(ctx context.Context, opts upgrade.TerraformUpgradeOptions) error
	CheckTerraformMigrations() error
	CleanUpTerraformMigrations() error
	AddManualStateMigration(migration terraform.StateMigration)
}

type migrationExecutor interface {
	executeMigration(cmd *cobra.Command, file file.Handler, migrateCmd upgrade.MigrationCmd, flags upgradeApplyFlags) error
}

type migrationCmdExecutor struct {
	log debugLog
}

// planMigration checks for Terraform migrations and asks for confirmation if there are any. The user input is returned as confirmedDiff.
// adapted from migrateTerraform().
func (u *migrationCmdExecutor) planMigration(cmd *cobra.Command, file file.Handler, migrateCmd upgrade.MigrationCmd) (hasDiff bool, err error) {
	u.log.Debugf("Planning %s", migrateCmd.String())
	err = migrateCmd.CheckTerraformMigrations(file)
	if err != nil {
		return false, fmt.Errorf("checking workspace: %w", err)
	}
	hasDiff, err = migrateCmd.Plan(cmd.Context(), file, cmd.OutOrStdout())
	if err != nil {
		return hasDiff, fmt.Errorf("planning terraform migrations: %w", err)
	}
	return hasDiff, nil
}

func (u *migrationCmdExecutor) executeMigration(cmd *cobra.Command, file file.Handler, migrateCmd upgrade.MigrationCmd, flags upgradeApplyFlags) error {
	hasDiff, err := u.planMigration(cmd, file, migrateCmd)
	if err != nil {
		return fmt.Errorf("planning terraform migrations: %w", err)
	}
	if hasDiff {
		// If there are any Terraform migrations to apply, ask for confirmation
		fmt.Fprintf(cmd.OutOrStdout(), "The %s upgrade requires a migration of Constellation cloud resources by applying an updated Terraform template. Please manually review the suggested changes below.\n", migrateCmd.String())
		if !flags.yes {
			ok, err := askToConfirm(cmd, fmt.Sprintf("Do you want to apply the %s?", migrateCmd.String()))
			if err != nil {
				return fmt.Errorf("asking for confirmation: %w", err)
			}
			if !ok {
				cmd.Println("Aborting upgrade.")
				if err := upgrade.CleanUpTerraformMigrations(migrateCmd.UpgradeID(), file); err != nil {
					return fmt.Errorf("cleaning up workspace: %w", err)
				}
				return fmt.Errorf("aborted by user")
			}
		}
		u.log.Debugf("Applying Terraform %s migrations", migrateCmd.String())
		err := migrateCmd.Apply(cmd.Context(), file)
		if err != nil {
			return fmt.Errorf("applying terraform migrations: %w", err)
		}
	} else {
		u.log.Debugf("No Terraform diff detected")
	}
	return nil
}

func toPtr[T any](v T) *T {
	return &v
}
