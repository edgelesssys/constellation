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
	"github.com/edgelesssys/constellation/v2/internal/imagefetcher"
	"github.com/edgelesssys/constellation/v2/internal/versions"
	"github.com/rogpeppe/go-internal/diff"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
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

	upgrader, err := kubernetes.NewUpgrader(
		cmd.Context(), cmd.OutOrStdout(),
		constants.UpgradeDir, constants.AdminConfFilename,
		fileHandler, log, kubernetes.UpgradeCmdKindApply,
	)
	if err != nil {
		return err
	}

	imagefetcher := imagefetcher.New()
	configFetcher := attestationconfigapi.NewFetcher()
	applyCmd := upgradeApplyCmd{upgrader: upgrader, log: log, imageFetcher: imagefetcher, configFetcher: configFetcher}
	return applyCmd.upgradeApply(cmd, fileHandler, stableClientFactoryImpl)
}

type stableClientFactory func(kubeconfigPath string) (getConfigMapper, error)

// needed because StableClient returns the bigger kubernetes.StableInterface.
func stableClientFactoryImpl(kubeconfigPath string) (getConfigMapper, error) {
	return kubernetes.NewStableClient(kubeconfigPath)
}

type getConfigMapper interface {
	GetCurrentConfigMap(ctx context.Context, name string) (*corev1.ConfigMap, error)
}

type upgradeApplyCmd struct {
	upgrader      cloudUpgrader
	imageFetcher  imageFetcher
	configFetcher attestationconfigapi.Fetcher
	log           debugLog
}

func (u *upgradeApplyCmd) upgradeApply(cmd *cobra.Command, fileHandler file.Handler, stableClientFactory stableClientFactory) error {
	flags, err := parseUpgradeApplyFlags(cmd)
	if err != nil {
		return fmt.Errorf("parsing flags: %w", err)
	}

	conf, err := config.New(fileHandler, constants.ConfigFilename, u.configFetcher, flags.force)
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
	if err := handleInvalidK8sPatchVersion(cmd, conf.KubernetesVersion, flags.yes); err != nil {
		return err
	}

	var idFile clusterid.File
	if err := fileHandler.ReadJSON(constants.ClusterIDsFilename, &idFile); err != nil {
		return fmt.Errorf("reading cluster ID file: %w", err)
	}
	conf.UpdateMAAURL(idFile.AttestationURL)

	// If an image upgrade was just executed there won't be a diff. The function will return nil in that case.
	stableClient, err := stableClientFactory(constants.AdminConfFilename)
	if err != nil {
		return fmt.Errorf("creating stable client: %w", err)
	}
	if err := u.upgradeAttestConfigIfDiff(cmd, stableClient, conf.GetAttestationConfig(), flags); err != nil {
		return fmt.Errorf("upgrading measurements: %w", err)
	}
	// not moving existing Terraform migrator because of planned apply refactor
	if err := u.migrateTerraform(cmd, u.imageFetcher, conf, fileHandler, flags); err != nil {
		return fmt.Errorf("performing Terraform migrations: %w", err)
	}
	// reload idFile after terraform migration
	// it might have been updated by the migration
	if err := fileHandler.ReadJSON(constants.ClusterIDsFilename, &idFile); err != nil {
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

func getImage(ctx context.Context, conf *config.Config, fetcher imageFetcher) (string, error) {
	// Fetch variables to execute Terraform script with
	provider := conf.GetProvider()
	attestationVariant := conf.GetAttestationConfig().GetVariant()
	region := conf.GetRegion()
	return fetcher.FetchReference(ctx, provider, attestationVariant, conf.Image, region)
}

// migrateTerraform checks if the Constellation version the cluster is being upgraded to requires a migration
// of cloud resources with Terraform. If so, the migration is performed.
func (u *upgradeApplyCmd) migrateTerraform(
	cmd *cobra.Command, fetcher imageFetcher, conf *config.Config, fileHandler file.Handler, flags upgradeApplyFlags,
) error {
	u.log.Debugf("Planning Terraform migrations")

	if err := u.upgrader.CheckTerraformMigrations(constants.UpgradeDir); err != nil {
		return fmt.Errorf("checking workspace: %w", err)
	}

	imageRef, err := getImage(cmd.Context(), conf, fetcher)
	if err != nil {
		return fmt.Errorf("fetching image reference: %w", err)
	}

	vars, err := cloudcmd.TerraformUpgradeVars(conf, imageRef)
	if err != nil {
		return fmt.Errorf("parsing upgrade variables: %w", err)
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
				if err := u.upgrader.CleanUpTerraformMigrations(constants.UpgradeDir); err != nil {
					return fmt.Errorf("cleaning up workspace: %w", err)
				}
				return fmt.Errorf("aborted by user")
			}
		}

		u.log.Debugf("Applying Terraform migrations")
		newIDFile, err := u.upgrader.ApplyTerraformMigrations(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("applying terraform migrations: %w", err)
		}
		if err := mergeClusterIDFile(constants.ClusterIDsFilename, newIDFile, fileHandler); err != nil {
			return fmt.Errorf("merging cluster ID files: %w", err)
		}

		cmd.Printf("Terraform migrations applied successfully and output written to: %s\n"+
			"A backup of the pre-upgrade state has been written to: %s\n",
			flags.pf.PrefixPath(constants.ClusterIDsFilename), flags.pf.PrefixPath(filepath.Join(opts.UpgradeWorkspace, u.upgrader.GetUpgradeID(), constants.TerraformUpgradeBackupDir)))
	} else {
		u.log.Debugf("No Terraform diff detected")
	}

	return nil
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
func (u *upgradeApplyCmd) upgradeAttestConfigIfDiff(cmd *cobra.Command, stableClient getConfigMapper, newConfig config.AttestationCfg, flags upgradeApplyFlags) error {
	clusterAttestationConfig, err := getAttestationConfig(cmd.Context(), stableClient, newConfig.GetVariant())
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
		cmd.Println("The configured attestation config is different from the attestation config in the cluster.")
		diffStr, err := diffAttestationCfg(clusterAttestationConfig, newConfig)
		if err != nil {
			return fmt.Errorf("diffing attestation configs: %w", err)
		}
		cmd.Println("The following changes will be applied to the attestation config:")
		cmd.Println(diffStr)
		ok, err := askToConfirm(cmd, "Are you sure you want to change your cluster's attestation config?")
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

	return upgradeApplyFlags{
		pf:                pathprefix.New(workDir),
		yes:               yes,
		upgradeTimeout:    timeout,
		force:             force,
		terraformLogLevel: logLevel,
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
}

type cloudUpgrader interface {
	UpgradeNodeVersion(ctx context.Context, conf *config.Config, force bool) error
	UpgradeHelmServices(ctx context.Context, config *config.Config, idFile clusterid.File, timeout time.Duration, allowDestructive bool, force bool) error
	UpdateAttestationConfig(ctx context.Context, newConfig config.AttestationCfg) error
	ExtendClusterConfigCertSANs(ctx context.Context, alternativeNames []string) error
	GetClusterAttestationConfig(ctx context.Context, variant variant.Variant) (config.AttestationCfg, *corev1.ConfigMap, error)
	PlanTerraformMigrations(ctx context.Context, opts upgrade.TerraformUpgradeOptions) (bool, error)
	ApplyTerraformMigrations(ctx context.Context, opts upgrade.TerraformUpgradeOptions) (clusterid.File, error)
	CheckTerraformMigrations(upgradeWorkspace string) error
	CleanUpTerraformMigrations(upgradeWorkspace string) error
	GetUpgradeID() string
}
