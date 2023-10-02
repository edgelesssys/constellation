/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"github.com/edgelesssys/constellation/v2/cli/internal/cloudcmd"
	"github.com/edgelesssys/constellation/v2/cli/internal/clusterid"
	"github.com/edgelesssys/constellation/v2/cli/internal/cmd/pathprefix"
	"github.com/edgelesssys/constellation/v2/cli/internal/helm"
	"github.com/edgelesssys/constellation/v2/cli/internal/kubecmd"
	"github.com/edgelesssys/constellation/v2/cli/internal/state"
	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/internal/api/attestationconfigapi"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/compatibility"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/kms/uri"
	"github.com/edgelesssys/constellation/v2/internal/semver"
	"github.com/edgelesssys/constellation/v2/internal/versions"
	"github.com/rogpeppe/go-internal/diff"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

const (
	// skipInfrastructurePhase skips the terraform apply of the upgrade process.
	skipInfrastructurePhase skipPhase = "infrastructure"
	// skipHelmPhase skips the helm upgrade of the upgrade process.
	skipHelmPhase skipPhase = "helm"
	// skipImagePhase skips the image upgrade of the upgrade process.
	skipImagePhase skipPhase = "image"
	// skipK8sPhase skips the k8s upgrade of the upgrade process.
	skipK8sPhase skipPhase = "k8s"
)

// skipPhase is a phase of the upgrade process that can be skipped.
type skipPhase string

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
	cmd.Flags().StringSlice("skip-phases", nil, "comma-separated list of upgrade phases to skip\n"+
		"one or multiple of { infrastructure | helm | image | k8s }")
	if err := cmd.Flags().MarkHidden("timeout"); err != nil {
		panic(err)
	}

	return cmd
}

func runUpgradeApply(cmd *cobra.Command, _ []string) error {
	flags, err := parseUpgradeApplyFlags(cmd)
	if err != nil {
		return fmt.Errorf("parsing flags: %w", err)
	}

	log, err := newCLILogger(cmd)
	if err != nil {
		return fmt.Errorf("creating logger: %w", err)
	}
	defer log.Sync()

	fileHandler := file.NewHandler(afero.NewOsFs())
	upgradeID := generateUpgradeID(upgradeCmdKindApply)

	kubeUpgrader, err := kubecmd.New(cmd.OutOrStdout(), constants.AdminConfFilename, fileHandler, log)
	if err != nil {
		return err
	}

	configFetcher := attestationconfigapi.NewFetcher()

	// Set up terraform upgrader
	upgradeDir := filepath.Join(constants.UpgradeDir, upgradeID)
	clusterUpgrader, err := cloudcmd.NewClusterUpgrader(
		cmd.Context(),
		constants.TerraformWorkingDir,
		upgradeDir,
		flags.terraformLogLevel,
		fileHandler,
	)
	if err != nil {
		return fmt.Errorf("setting up cluster upgrader: %w", err)
	}

	// Set up terraform client to show existing cluster resources and information required for Helm upgrades
	tfShower, err := terraform.New(cmd.Context(), constants.TerraformWorkingDir)
	if err != nil {
		return fmt.Errorf("setting up terraform client: %w", err)
	}
	helmClient, err := helm.NewClient(constants.AdminConfFilename, log)
	if err != nil {
		return fmt.Errorf("creating Helm client: %w", err)
	}

	applyCmd := upgradeApplyCmd{
		kubeUpgrader:    kubeUpgrader,
		helmApplier:     helmClient,
		clusterUpgrader: clusterUpgrader,
		configFetcher:   configFetcher,
		clusterShower:   tfShower,
		fileHandler:     fileHandler,
		log:             log,
	}
	return applyCmd.upgradeApply(cmd, upgradeDir, flags)
}

type upgradeApplyCmd struct {
	helmApplier     helmApplier
	kubeUpgrader    kubernetesUpgrader
	clusterUpgrader clusterUpgrader
	configFetcher   attestationconfigapi.Fetcher
	clusterShower   infrastructureShower
	fileHandler     file.Handler
	log             debugLog
}

type infrastructureShower interface {
	ShowInfrastructure(ctx context.Context, provider cloudprovider.Provider) (state.Infrastructure, error)
}

func (u *upgradeApplyCmd) upgradeApply(cmd *cobra.Command, upgradeDir string, flags upgradeApplyFlags) error {
	conf, err := config.New(u.fileHandler, constants.ConfigFilename, u.configFetcher, flags.force)
	var configValidationErr *config.ValidationError
	if errors.As(err, &configValidationErr) {
		cmd.PrintErrln(configValidationErr.LongMessage())
	}
	if err != nil {
		return err
	}
	if cloudcmd.UpgradeRequiresIAMMigration(conf.GetProvider()) {
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
	conf.KubernetesVersion, err = validK8sVersion(cmd, string(conf.KubernetesVersion), flags.yes)
	if err != nil {
		return err
	}

	stateFile, err := state.ReadFromFile(u.fileHandler, constants.StateFilename)
	// TODO(msanft): Remove reading from idFile once v2.12.0 is released and read from state file directly.
	// For now, this is only here to ensure upgradability from an id-file to a state file version.
	if errors.Is(err, fs.ErrNotExist) {
		u.log.Debugf("%s does not exist in current directory, falling back to reading from %s",
			constants.StateFilename, constants.ClusterIDsFilename)
		var idFile clusterid.File
		if err := u.fileHandler.ReadJSON(constants.ClusterIDsFilename, &idFile); err != nil {
			return fmt.Errorf("reading cluster ID file: %w", err)
		}
		// Convert id-file to state file
		stateFile = state.NewFromIDFile(idFile)
		if stateFile.Infrastructure.Azure != nil {
			conf.UpdateMAAURL(stateFile.Infrastructure.Azure.AttestationURL)
		}
	} else if err != nil {
		return fmt.Errorf("reading state file: %w", err)
	}

	// Apply migrations necessary for the upgrade
	if err := migrateFrom2_10(cmd.Context(), u.kubeUpgrader); err != nil {
		return fmt.Errorf("applying migration for upgrading from v2.10: %w", err)
	}
	if err := migrateFrom2_11(cmd.Context(), u.kubeUpgrader); err != nil {
		return fmt.Errorf("applying migration for upgrading from v2.11: %w", err)
	}

	if err := u.confirmAndUpgradeAttestationConfig(cmd, conf.GetAttestationConfig(), stateFile.ClusterValues.MeasurementSalt, flags); err != nil {
		return fmt.Errorf("upgrading measurements: %w", err)
	}

	// If infrastructure phase is skipped, we expect the new infrastructure
	// to be in the Terraform configuration already. Otherwise, perform
	// the Terraform migrations.
	var postMigrationInfraState state.Infrastructure
	if flags.skipPhases.contains(skipInfrastructurePhase) {
		postMigrationInfraState, err = u.clusterShower.ShowInfrastructure(cmd.Context(), conf.GetProvider())
		if err != nil {
			return fmt.Errorf("getting Terraform state: %w", err)
		}
	} else {
		postMigrationInfraState, err = u.migrateTerraform(cmd, conf, upgradeDir, flags)
		if err != nil {
			return fmt.Errorf("performing Terraform migrations: %w", err)
		}
	}

	// Merge the pre-upgrade state with the post-migration infrastructure values
	if _, err := stateFile.Merge(
		// temporary state with post-migration infrastructure values
		state.New().SetInfrastructure(postMigrationInfraState),
	); err != nil {
		return fmt.Errorf("merging pre-upgrade state with post-migration infrastructure values: %w", err)
	}

	// Write the post-migration state to disk
	if err := stateFile.WriteToFile(u.fileHandler, constants.StateFilename); err != nil {
		return fmt.Errorf("writing state file: %w", err)
	}

	// extend the clusterConfig cert SANs with any of the supported endpoints:
	// - (legacy) public IP
	// - fallback endpoint
	// - custom (user-provided) endpoint
	// At this point, state file and id-file should have been merged, so we can use the state file.
	sans := append([]string{stateFile.Infrastructure.ClusterEndpoint, conf.CustomEndpoint}, stateFile.Infrastructure.APIServerCertSANs...)
	if err := u.kubeUpgrader.ExtendClusterConfigCertSANs(cmd.Context(), sans); err != nil {
		return fmt.Errorf("extending cert SANs: %w", err)
	}

	if conf.GetProvider() != cloudprovider.Azure && conf.GetProvider() != cloudprovider.GCP && conf.GetProvider() != cloudprovider.AWS {
		cmd.PrintErrln("WARNING: Skipping service and image upgrades, which are currently only supported for AWS, Azure, and GCP.")
		return nil
	}

	var upgradeErr *compatibility.InvalidUpgradeError
	if !flags.skipPhases.contains(skipHelmPhase) {
		err = u.handleServiceUpgrade(cmd, conf, stateFile, upgradeDir, flags)
		switch {
		case errors.As(err, &upgradeErr):
			cmd.PrintErrln(err)
		case err == nil:
			cmd.Println("Successfully upgraded Constellation services.")
		case err != nil:
			return fmt.Errorf("upgrading services: %w", err)
		}
	}
	skipImageUpgrade := flags.skipPhases.contains(skipImagePhase)
	skipK8sUpgrade := flags.skipPhases.contains(skipK8sPhase)
	if !(skipImageUpgrade && skipK8sUpgrade) {
		err = u.kubeUpgrader.UpgradeNodeVersion(cmd.Context(), conf, flags.force, skipImageUpgrade, skipK8sUpgrade)
		switch {
		case errors.Is(err, kubecmd.ErrInProgress):
			cmd.PrintErrln("Skipping image and Kubernetes upgrades. Another upgrade is in progress.")
		case errors.As(err, &upgradeErr):
			cmd.PrintErrln(err)
		case err != nil:
			return fmt.Errorf("upgrading NodeVersion: %w", err)
		}
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
// of cloud resources with Terraform. If so, the migration is performed and the post-migration infrastructure state is returned.
func (u *upgradeApplyCmd) migrateTerraform(
	cmd *cobra.Command, conf *config.Config, upgradeDir string, flags upgradeApplyFlags,
) (state.Infrastructure, error) {
	u.log.Debugf("Planning Terraform migrations")

	vars, err := cloudcmd.TerraformUpgradeVars(conf)
	if err != nil {
		return state.Infrastructure{}, fmt.Errorf("parsing upgrade variables: %w", err)
	}
	u.log.Debugf("Using Terraform variables:\n%v", vars)

	// Check if there are any Terraform migrations to apply

	// Add manual migrations here if required
	//
	// var manualMigrations []terraform.StateMigration
	// for _, migration := range manualMigrations {
	// 	  u.log.Debugf("Adding manual Terraform migration: %s", migration.DisplayName)
	// 	  u.upgrader.AddManualStateMigration(migration)
	// }

	hasDiff, err := u.clusterUpgrader.PlanClusterUpgrade(cmd.Context(), cmd.OutOrStdout(), vars, conf.GetProvider())
	if err != nil {
		return state.Infrastructure{}, fmt.Errorf("planning terraform migrations: %w", err)
	}

	if hasDiff {
		// If there are any Terraform migrations to apply, ask for confirmation
		fmt.Fprintln(cmd.OutOrStdout(), "The upgrade requires a migration of Constellation cloud resources by applying an updated Terraform template. Please manually review the suggested changes below.")
		if !flags.yes {
			ok, err := askToConfirm(cmd, "Do you want to apply the Terraform migrations?")
			if err != nil {
				return state.Infrastructure{}, fmt.Errorf("asking for confirmation: %w", err)
			}
			if !ok {
				cmd.Println("Aborting upgrade.")
				// User doesn't expect to see any changes in his workspace after aborting an "upgrade apply",
				// therefore, roll back to the backed up state.
				if err := u.clusterUpgrader.RestoreClusterWorkspace(); err != nil {
					return state.Infrastructure{}, fmt.Errorf(
						"restoring Terraform workspace: %w, restore the Terraform workspace manually from %s ",
						err,
						filepath.Join(upgradeDir, constants.TerraformUpgradeBackupDir),
					)
				}
				return state.Infrastructure{}, fmt.Errorf("cluster upgrade aborted by user")
			}
		}
		u.log.Debugf("Applying Terraform migrations")

		infraState, err := u.clusterUpgrader.ApplyClusterUpgrade(cmd.Context(), conf.GetProvider())
		if err != nil {
			return state.Infrastructure{}, fmt.Errorf("applying terraform migrations: %w", err)
		}

		cmd.Printf("Infrastructure migrations applied successfully and output written to: %s\n"+
			"A backup of the pre-upgrade state has been written to: %s\n",
			flags.pf.PrefixPrintablePath(constants.StateFilename),
			flags.pf.PrefixPrintablePath(filepath.Join(upgradeDir, constants.TerraformUpgradeBackupDir)),
		)
		return infraState, nil
	}

	u.log.Debugf("No Terraform diff detected")
	return state.Infrastructure{}, nil
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

// confirmAndUpgradeAttestationConfig checks if the locally configured measurements are different from the cluster's measurements.
// If so the function will ask the user to confirm (if --yes is not set) and upgrade the cluster's config.
func (u *upgradeApplyCmd) confirmAndUpgradeAttestationConfig(
	cmd *cobra.Command, newConfig config.AttestationCfg, measurementSalt []byte, flags upgradeApplyFlags,
) error {
	clusterAttestationConfig, err := u.kubeUpgrader.GetClusterAttestationConfig(cmd.Context(), newConfig.GetVariant())
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

	if err := u.kubeUpgrader.ApplyJoinConfig(cmd.Context(), newConfig, measurementSalt); err != nil {
		return fmt.Errorf("updating attestation config: %w", err)
	}
	cmd.Println("Successfully updated the cluster's attestation config")
	return nil
}

func (u *upgradeApplyCmd) handleServiceUpgrade(
	cmd *cobra.Command, conf *config.Config, stateFile *state.State,
	upgradeDir string, flags upgradeApplyFlags,
) error {
	var secret uri.MasterSecret
	if err := u.fileHandler.ReadJSON(constants.MasterSecretFilename, &secret); err != nil {
		return fmt.Errorf("reading master secret: %w", err)
	}
	serviceAccURI, err := cloudcmd.GetMarshaledServiceAccountURI(conf.GetProvider(), conf, flags.pf, u.log, u.fileHandler)
	if err != nil {
		return fmt.Errorf("getting service account URI: %w", err)
	}
	options := helm.Options{
		Force:        flags.force,
		Conformance:  flags.conformance,
		HelmWaitMode: flags.helmWaitMode,
	}

	prepareApply := func(allowDestructive bool) (helm.Applier, bool, error) {
		options.AllowDestructive = allowDestructive
		executor, includesUpgrades, err := u.helmApplier.PrepareApply(conf, stateFile, options, serviceAccURI, secret)
		var upgradeErr *compatibility.InvalidUpgradeError
		switch {
		case errors.As(err, &upgradeErr):
			cmd.PrintErrln(err)
		case err != nil:
			return nil, false, fmt.Errorf("getting chart executor: %w", err)
		}
		return executor, includesUpgrades, nil
	}

	executor, includesUpgrades, err := prepareApply(helm.DenyDestructive)
	if err != nil {
		if !errors.Is(err, helm.ErrConfirmationMissing) {
			return fmt.Errorf("upgrading charts with deny destructive mode: %w", err)
		}
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
		executor, includesUpgrades, err = prepareApply(helm.AllowDestructive)
		if err != nil {
			return fmt.Errorf("upgrading charts with allow destructive mode: %w", err)
		}
	}

	// Save the Helm charts for the upgrade to disk
	chartDir := filepath.Join(upgradeDir, "helm-charts")
	if err := executor.SaveCharts(chartDir, u.fileHandler); err != nil {
		return fmt.Errorf("saving Helm charts to disk: %w", err)
	}
	u.log.Debugf("Helm charts saved to %s", chartDir)

	if includesUpgrades {
		u.log.Debugf("Creating backup of CRDs and CRs")
		crds, err := u.kubeUpgrader.BackupCRDs(cmd.Context(), upgradeDir)
		if err != nil {
			return fmt.Errorf("creating CRD backup: %w", err)
		}
		if err := u.kubeUpgrader.BackupCRs(cmd.Context(), crds, upgradeDir); err != nil {
			return fmt.Errorf("creating CR backup: %w", err)
		}
	}
	if err := executor.Apply(cmd.Context()); err != nil {
		return fmt.Errorf("applying Helm charts: %w", err)
	}

	return nil
}

// migrateFrom2_10 applies migrations necessary for upgrading from v2.10 to v2.11
// TODO(v2.11): Remove this function after v2.11 is released.
func migrateFrom2_10(ctx context.Context, kubeUpgrader kubernetesUpgrader) error {
	// Sanity check to make sure we only run migrations on upgrades with CLI version 2.10 < v < 2.12
	if !constants.BinaryVersion().MajorMinorEqual(semver.NewFromInt(2, 11, 0, "")) {
		return nil
	}

	if err := kubeUpgrader.RemoveAttestationConfigHelmManagement(ctx); err != nil {
		return fmt.Errorf("removing helm management from attestation config: %w", err)
	}
	return nil
}

// migrateFrom2_11 applies migrations necessary for upgrading from v2.11 to v2.12
// TODO(v2.12): Remove this function after v2.12 is released.
func migrateFrom2_11(ctx context.Context, kubeUpgrader kubernetesUpgrader) error {
	// Sanity check to make sure we only run migrations on upgrades with CLI version 2.11 < v < 2.13
	if !constants.BinaryVersion().MajorMinorEqual(semver.NewFromInt(2, 12, 0, "")) {
		return nil
	}

	if err := kubeUpgrader.RemoveHelmKeepAnnotation(ctx); err != nil {
		return fmt.Errorf("removing helm keep annotation: %w", err)
	}
	return nil
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

	rawSkipPhases, err := cmd.Flags().GetStringSlice("skip-phases")
	if err != nil {
		return upgradeApplyFlags{}, fmt.Errorf("parsing skip-phases flag: %w", err)
	}
	var skipPhases []skipPhase
	for _, phase := range rawSkipPhases {
		switch skipPhase(phase) {
		case skipInfrastructurePhase, skipHelmPhase, skipImagePhase, skipK8sPhase:
			skipPhases = append(skipPhases, skipPhase(phase))
		default:
			return upgradeApplyFlags{}, fmt.Errorf("invalid phase %s", phase)
		}
	}

	return upgradeApplyFlags{
		pf:                pathprefix.New(workDir),
		yes:               yes,
		upgradeTimeout:    timeout,
		force:             force,
		terraformLogLevel: logLevel,
		conformance:       conformance,
		helmWaitMode:      helmWaitMode,
		skipPhases:        skipPhases,
	}, nil
}

type upgradeApplyFlags struct {
	pf                pathprefix.PathPrefixer
	yes               bool
	upgradeTimeout    time.Duration
	force             bool
	terraformLogLevel terraform.LogLevel
	conformance       bool
	helmWaitMode      helm.WaitMode
	skipPhases        skipPhases
}

// skipPhases is a list of phases that can be skipped during the upgrade process.
type skipPhases []skipPhase

// contains returns true if the list of phases contains the given phase.
func (s skipPhases) contains(phase skipPhase) bool {
	for _, p := range s {
		if strings.EqualFold(string(p), string(phase)) {
			return true
		}
	}
	return false
}

type kubernetesUpgrader interface {
	UpgradeNodeVersion(ctx context.Context, conf *config.Config, force, skipImage, skipK8s bool) error
	ExtendClusterConfigCertSANs(ctx context.Context, alternativeNames []string) error
	GetClusterAttestationConfig(ctx context.Context, variant variant.Variant) (config.AttestationCfg, error)
	ApplyJoinConfig(ctx context.Context, newAttestConfig config.AttestationCfg, measurementSalt []byte) error
	BackupCRs(ctx context.Context, crds []apiextensionsv1.CustomResourceDefinition, upgradeDir string) error
	BackupCRDs(ctx context.Context, upgradeDir string) ([]apiextensionsv1.CustomResourceDefinition, error)
	// TODO(v2.11): Remove this function after v2.11 is released.
	RemoveAttestationConfigHelmManagement(ctx context.Context) error
	// TODO(v2.12): Remove this function after v2.12 is released.
	RemoveHelmKeepAnnotation(ctx context.Context) error
}

type clusterUpgrader interface {
	PlanClusterUpgrade(ctx context.Context, outWriter io.Writer, vars terraform.Variables, csp cloudprovider.Provider) (bool, error)
	ApplyClusterUpgrade(ctx context.Context, csp cloudprovider.Provider) (state.Infrastructure, error)
	RestoreClusterWorkspace() error
}
