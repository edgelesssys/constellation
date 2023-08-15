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
	"time"

	"github.com/edgelesssys/constellation/v2/cli/internal/cloudcmd"
	"github.com/edgelesssys/constellation/v2/cli/internal/clusterid"
	"github.com/edgelesssys/constellation/v2/cli/internal/cmd/pathprefix"
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
	"github.com/edgelesssys/constellation/v2/internal/kms/uri"
	"github.com/edgelesssys/constellation/v2/internal/versions"
	"github.com/rogpeppe/go-internal/diff"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
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
	cmd.Flags().Bool("conformance", false, "enable conformance mode")
	cmd.Flags().Bool("skip-helm-wait", false, "install helm charts without waiting for deployments to be ready")
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

	upgrader, err := kubernetes.NewUpgrader(
		cmd.Context(), cmd.OutOrStdout(),
		constants.UpgradeDir, constants.AdminConfFilename,
		fileHandler, log, kubernetes.UpgradeCmdKindApply,
	)
	if err != nil {
		return err
	}

	configFetcher := attestationconfigapi.NewFetcher()
	tfClient, err := terraform.New(cmd.Context(), constants.TerraformWorkingDir)
	if err != nil {
		return fmt.Errorf("setting up terraform client: %w", err)
	}

	applyCmd := upgradeApplyCmd{upgrader: upgrader, log: log, configFetcher: configFetcher, clusterShower: tfClient, fileHandler: fileHandler}
	return applyCmd.upgradeApply(cmd)
}

type upgradeApplyCmd struct {
	upgrader      cloudUpgrader
	configFetcher attestationconfigapi.Fetcher
	clusterShower clusterShower
	fileHandler   file.Handler
	log           debugLog
}

func (u *upgradeApplyCmd) upgradeApply(cmd *cobra.Command) error {
	flags, err := parseUpgradeApplyFlags(cmd)
	if err != nil {
		return fmt.Errorf("parsing flags: %w", err)
	}

	conf, err := config.New(u.fileHandler, constants.ConfigFilename, u.configFetcher, flags.force)
	var configValidationErr *config.ValidationError
	if errors.As(err, &configValidationErr) {
		cmd.PrintErrln(configValidationErr.LongMessage())
	}
	if err != nil {
		return err
	}
	if upgradeRequiresIAMMigration(conf.GetProvider()) {
		cmd.Println("WARNING: This upgrade requires an IAM migration. Please make sure you have applied the IAM migration using `iam upgrade apply` before continuing.")
		if !flags.yes {
			yes, err := askToConfirm(cmd, "Did you upgrade the IAM resources?")
			if err != nil {
				return fmt.Errorf("asking for confirmation: %w", err)
			}
			if !yes {
				cmd.Println("Skipping upgrade.")
				return nil
			}
		}
	}
	validK8sVersion, err := validK8sVersion(cmd, conf.KubernetesVersion, flags.yes)
	if err != nil {
		return err
	}

	var idFile clusterid.File
	if err := u.fileHandler.ReadJSON(constants.ClusterIDsFilename, &idFile); err != nil {
		return fmt.Errorf("reading cluster ID file: %w", err)
	}
	if idFile.MeasurementSalt == nil {
		// TODO(elchead): remove after 2.10, since 2.9 does not yet save it in the idfile
		measurementSalt, err := u.upgrader.GetMeasurementSalt(cmd.Context())
		if err != nil {
			return fmt.Errorf("getting join-config: %w", err)
		}
		idFile.MeasurementSalt = measurementSalt
		if err := u.fileHandler.WriteJSON(constants.ClusterIDsFilename, idFile, file.OptOverwrite); err != nil {
			return fmt.Errorf("writing cluster ID file: %w", err)
		}
	}
	conf.UpdateMAAURL(idFile.AttestationURL)

	if err := u.confirmIfUpgradeAttestConfigHasDiff(cmd, conf.GetAttestationConfig(), flags); err != nil {
		return fmt.Errorf("upgrading measurements: %w", err)
	}
	// not moving existing Terraform migrator because of planned apply refactor
	tfOutput, err := u.migrateTerraform(cmd, conf, flags)
	if err != nil {
		return fmt.Errorf("performing Terraform migrations: %w", err)
	}
	// reload idFile after terraform migration
	// it might have been updated by the migration
	if err := u.fileHandler.ReadJSON(constants.ClusterIDsFilename, &idFile); err != nil {
		return fmt.Errorf("reading updated cluster ID file: %w", err)
	}

	// extend the clusterConfig cert SANs with any of the supported endpoints:
	// - (legacy) public IP
	// - fallback endpoint
	// - custom (user-provided) endpoint
	sans := append([]string{idFile.IP, conf.CustomEndpoint}, idFile.APIServerCertSANs...)
	if err := u.upgrader.ExtendClusterConfigCertSANs(cmd.Context(), sans); err != nil {
		return fmt.Errorf("extending cert SANs: %w", err)
	}

	if conf.GetProvider() == cloudprovider.Azure || conf.GetProvider() == cloudprovider.GCP || conf.GetProvider() == cloudprovider.AWS {
		var upgradeErr *compatibility.InvalidUpgradeError
		err = u.handleServiceUpgrade(cmd, conf, idFile, tfOutput, validK8sVersion, flags)
		switch {
		case errors.As(err, &upgradeErr):
			cmd.PrintErrln(err)
		case err == nil:
			cmd.Println("Successfully upgraded Constellation services.")
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

func diffAttestationCfg(currentAttestationCfg config.AttestationCfg, newAttestationCfg config.AttestationCfg) (string, error) {
	// cannot compare structs directly with go-cmp because of unexported fields in the attestation config
	currentYml, err := yaml.Marshal(currentAttestationCfg)
	if err != nil {
		return "", fmt.Errorf("marshalling remote attestation config: %w", err)
	}
	newYml, err := yaml.Marshal(newAttestationCfg)
	if err != nil {
		return "", fmt.Errorf("marshalling local attestation config: %w", err)
	}
	diff := string(diff.Diff("current", currentYml, "new", newYml))
	return diff, nil
}

// migrateTerraform checks if the Constellation version the cluster is being upgraded to requires a migration
// of cloud resources with Terraform. If so, the migration is performed.
func (u *upgradeApplyCmd) migrateTerraform(
	cmd *cobra.Command, conf *config.Config, flags upgradeApplyFlags,
) (res terraform.ApplyOutput, err error) {
	u.log.Debugf("Planning Terraform migrations")

	if err := u.upgrader.CheckTerraformMigrations(constants.UpgradeDir); err != nil {
		return res, fmt.Errorf("checking workspace: %w", err)
	}

	vars, err := cloudcmd.TerraformUpgradeVars(conf)
	if err != nil {
		return res, fmt.Errorf("parsing upgrade variables: %w", err)
	}
	u.log.Debugf("Using Terraform variables:\n%v", vars)

	opts := upgrade.TerraformUpgradeOptions{
		LogLevel:         flags.terraformLogLevel,
		CSP:              conf.GetProvider(),
		Vars:             vars,
		TFWorkspace:      constants.TerraformWorkingDir,
		UpgradeWorkspace: constants.UpgradeDir,
	}

	// Check if there are any Terraform migrations to apply

	// Add manual migrations here if required
	//
	// var manualMigrations []terraform.StateMigration
	// for _, migration := range manualMigrations {
	// 	  u.log.Debugf("Adding manual Terraform migration: %s", migration.DisplayName)
	// 	  u.upgrader.AddManualStateMigration(migration)
	// }

	hasDiff, err := u.upgrader.PlanTerraformMigrations(cmd.Context(), opts)
	if err != nil {
		return res, fmt.Errorf("planning terraform migrations: %w", err)
	}

	if hasDiff {
		// If there are any Terraform migrations to apply, ask for confirmation
		fmt.Fprintln(cmd.OutOrStdout(), "The upgrade requires a migration of Constellation cloud resources by applying an updated Terraform template. Please manually review the suggested changes below.")
		if !flags.yes {
			ok, err := askToConfirm(cmd, "Do you want to apply the Terraform migrations?")
			if err != nil {
				return res, fmt.Errorf("asking for confirmation: %w", err)
			}
			if !ok {
				cmd.Println("Aborting upgrade.")
				if err := u.upgrader.CleanUpTerraformMigrations(constants.UpgradeDir); err != nil {
					return res, fmt.Errorf("cleaning up workspace: %w", err)
				}
				return res, fmt.Errorf("aborted by user")
			}
		}
		u.log.Debugf("Applying Terraform migrations")
		tfOutput, err := u.upgrader.ApplyTerraformMigrations(cmd.Context(), opts)
		if err != nil {
			return tfOutput, fmt.Errorf("applying terraform migrations: %w", err)
		}

		// Patch MAA policy if we applied an Azure upgrade.
		newIDFile := newIDFile(opts, tfOutput)
		if err := mergeClusterIDFile(constants.ClusterIDsFilename, newIDFile, u.fileHandler); err != nil {
			return tfOutput, fmt.Errorf("merging cluster ID files: %w", err)
		}

		cmd.Printf("Terraform migrations applied successfully and output written to: %s\n"+
			"A backup of the pre-upgrade state has been written to: %s\n",
			flags.pf.PrefixPath(constants.ClusterIDsFilename), flags.pf.PrefixPath(filepath.Join(opts.UpgradeWorkspace, u.upgrader.GetUpgradeID(), constants.TerraformUpgradeBackupDir)))
	} else {
		u.log.Debugf("No Terraform diff detected")
	}
	u.log.Debugf("No Terraform diff detected")
	tfOutput, err := u.clusterShower.ShowCluster(cmd.Context(), conf.GetProvider())
	if err != nil {
		return tfOutput, fmt.Errorf("getting Terraform output: %w", err)
	}
	return tfOutput, nil
}

func newIDFile(opts upgrade.TerraformUpgradeOptions, tfOutput terraform.ApplyOutput) clusterid.File {
	newIDFile := clusterid.File{
		CloudProvider:     opts.CSP,
		InitSecret:        []byte(tfOutput.Secret),
		IP:                tfOutput.IP,
		APIServerCertSANs: tfOutput.APIServerCertSANs,
		UID:               tfOutput.UID,
	}
	if tfOutput.Azure != nil {
		newIDFile.AttestationURL = tfOutput.Azure.AttestationURL
	}
	return newIDFile
}

// validK8sVersion checks if the Kubernetes patch version is supported and asks for confirmation if not.
func validK8sVersion(cmd *cobra.Command, version string, yes bool) (validVersion versions.ValidK8sVersion, err error) {
	validVersion, err = versions.NewValidK8sVersion(version, true)
	if versions.IsPreviewK8sVersion(validVersion) {
		cmd.PrintErrf("Warning: Constellation with Kubernetes %v is still in preview. Use only for evaluation purposes.\n", validVersion)
	}
	valid := err == nil

	if !valid && !yes {
		confirmed, err := askToConfirm(cmd, fmt.Sprintf("WARNING: The Kubernetes patch version %s is not supported. If you continue, Kubernetes upgrades will be skipped. Do you want to continue anyway?", version))
		if err != nil {
			return validVersion, fmt.Errorf("asking for confirmation: %w", err)
		}
		if !confirmed {
			return validVersion, fmt.Errorf("aborted by user")
		}
	}

	return validVersion, nil
}

// confirmIfUpgradeAttestConfigHasDiff checks if the locally configured measurements are different from the cluster's measurements.
// If so the function will ask the user to confirm (if --yes is not set).
func (u *upgradeApplyCmd) confirmIfUpgradeAttestConfigHasDiff(cmd *cobra.Command, newConfig config.AttestationCfg, flags upgradeApplyFlags) error {
	clusterAttestationConfig, err := u.upgrader.GetClusterAttestationConfig(cmd.Context(), newConfig.GetVariant())
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
	cmd.Println("The configured attestation config is different from the attestation config in the cluster.")
	diffStr, err := diffAttestationCfg(clusterAttestationConfig, newConfig)
	if err != nil {
		return fmt.Errorf("diffing attestation configs: %w", err)
	}
	cmd.Println("The following changes will be applied to the attestation config:")
	cmd.Println(diffStr)
	if !flags.yes {
		ok, err := askToConfirm(cmd, "Are you sure you want to change your cluster's attestation config?")
		if err != nil {
			return fmt.Errorf("asking for confirmation: %w", err)
		}
		if !ok {
			return errors.New("aborting upgrade since attestation config is different")
		}
	}
	// TODO(elchead): move this outside this function to remove the side effect.
	if err := u.upgrader.BackupConfigMap(cmd.Context(), constants.JoinConfigMap); err != nil {
		return fmt.Errorf("backing up join-config: %w", err)
	}
	if err := u.upgrader.UpdateAttestationConfig(cmd.Context(), newConfig); err != nil {
		return fmt.Errorf("updating attestation config: %w", err)
	}
	return nil
}

func (u *upgradeApplyCmd) handleServiceUpgrade(cmd *cobra.Command, conf *config.Config, idFile clusterid.File, tfOutput terraform.ApplyOutput, validK8sVersion versions.ValidK8sVersion, flags upgradeApplyFlags) error {
	var secret uri.MasterSecret
	if err := u.fileHandler.ReadJSON(flags.pf.PrefixPath(constants.MasterSecretFilename), &secret); err != nil {
		return fmt.Errorf("reading master secret: %w", err)
	}
	serviceAccURI, err := cloudcmd.GetMarshaledServiceAccountURI(conf.GetProvider(), conf, flags.pf, u.log, u.fileHandler)
	if err != nil {
		return fmt.Errorf("getting service account URI: %w", err)
	}
	err = u.upgrader.UpgradeHelmServices(cmd.Context(), conf, idFile, flags.upgradeTimeout, helm.DenyDestructive, flags.force, flags.conformance, flags.helmWaitMode, secret, serviceAccURI, validK8sVersion, tfOutput)
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
		err = u.upgrader.UpgradeHelmServices(cmd.Context(), conf, idFile, flags.upgradeTimeout, helm.AllowDestructive, flags.force, flags.conformance, flags.helmWaitMode, secret, serviceAccURI, validK8sVersion, tfOutput)
	}

	return err
}

func parseUpgradeApplyFlags(cmd *cobra.Command) (upgradeApplyFlags, error) {
	workDir, err := cmd.Flags().GetString("workspace")
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

	conformance, err := cmd.Flags().GetBool("conformance")
	if err != nil {
		return upgradeApplyFlags{}, fmt.Errorf("parsing conformance flag: %w", err)
	}
	skipHelmWait, err := cmd.Flags().GetBool("skip-helm-wait")
	if err != nil {
		return upgradeApplyFlags{}, fmt.Errorf("parsing skip-helm-wait flag: %w", err)
	}
	helmWaitMode := helm.WaitModeAtomic
	if skipHelmWait {
		helmWaitMode = helm.WaitModeNone
	}
	return upgradeApplyFlags{
		pf:                pathprefix.New(workDir),
		yes:               yes,
		upgradeTimeout:    timeout,
		force:             force,
		terraformLogLevel: logLevel,
		conformance:       conformance,
		helmWaitMode:      helmWaitMode,
	}, nil
}

func mergeClusterIDFile(clusterIDPath string, newIDFile clusterid.File, fileHandler file.Handler) error {
	idFile := &clusterid.File{}
	if err := fileHandler.ReadJSON(clusterIDPath, idFile); err != nil {
		return fmt.Errorf("reading %s: %w", clusterIDPath, err)
	}

	if err := fileHandler.WriteJSON(clusterIDPath, idFile.Merge(newIDFile), file.OptOverwrite); err != nil {
		return fmt.Errorf("writing %s: %w", clusterIDPath, err)
	}

	return nil
}

type upgradeApplyFlags struct {
	pf                pathprefix.PathPrefixer
	yes               bool
	upgradeTimeout    time.Duration
	force             bool
	terraformLogLevel terraform.LogLevel
	conformance       bool
	helmWaitMode      helm.WaitMode
}

type cloudUpgrader interface {
	UpgradeNodeVersion(ctx context.Context, conf *config.Config, force bool) error
	UpgradeHelmServices(ctx context.Context, config *config.Config, idFile clusterid.File, timeout time.Duration, allowDestructive bool, force bool, conformance bool, helmWaitMode helm.WaitMode, masterSecret uri.MasterSecret, serviceAccURI string, validK8sVersion versions.ValidK8sVersion, tfOutput terraform.ApplyOutput) error
	ExtendClusterConfigCertSANs(ctx context.Context, alternativeNames []string) error
	GetClusterAttestationConfig(ctx context.Context, variant variant.Variant) (config.AttestationCfg, error)
	UpdateAttestationConfig(ctx context.Context, newAttestConfig config.AttestationCfg) error
	GetMeasurementSalt(ctx context.Context) ([]byte, error)
	PlanTerraformMigrations(ctx context.Context, opts upgrade.TerraformUpgradeOptions) (bool, error)
	ApplyTerraformMigrations(ctx context.Context, opts upgrade.TerraformUpgradeOptions) (terraform.ApplyOutput, error)
	CheckTerraformMigrations(upgradeWorkspace string) error
	CleanUpTerraformMigrations(upgradeWorkspace string) error
	GetUpgradeID() string
	BackupConfigMap(ctx context.Context, name string) error
}
