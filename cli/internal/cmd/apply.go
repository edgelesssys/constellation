/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/edgelesssys/constellation/v2/cli/internal/cloudcmd"
	"github.com/edgelesssys/constellation/v2/cli/internal/helm"
	"github.com/edgelesssys/constellation/v2/cli/internal/kubecmd"
	"github.com/edgelesssys/constellation/v2/cli/internal/state"
	"github.com/edgelesssys/constellation/v2/internal/api/attestationconfigapi"
	"github.com/edgelesssys/constellation/v2/internal/atls"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/compatibility"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/grpc/dialer"
	"github.com/edgelesssys/constellation/v2/internal/license"
	"github.com/edgelesssys/constellation/v2/internal/versions"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
)

// phases that can be skipped during apply.
// New phases should also be added to [allPhases].
const (
	// skipInfrastructurePhase skips the Terraform apply of the apply process.
	skipInfrastructurePhase skipPhase = "infrastructure"
	// skipInitPhase skips the init RPC of the apply process.
	skipInitPhase skipPhase = "init"
	// skipAttestationConfigPhase skips the attestation config upgrade of the apply process.
	skipAttestationConfigPhase skipPhase = "attestationconfig"
	// skipCertSANsPhase skips the cert SANs upgrade of the apply process.
	skipCertSANsPhase skipPhase = "certsans"
	// skipHelmPhase skips the helm upgrade of the apply process.
	skipHelmPhase skipPhase = "helm"
	// skipImagePhase skips the image upgrade of the apply process.
	skipImagePhase skipPhase = "image"
	// skipK8sPhase skips the Kubernetes version upgrade of the apply process.
	skipK8sPhase skipPhase = "k8s"
)

// allPhases returns a list of all phases that can be skipped as strings.
func allPhases() []string {
	return []string{
		string(skipInfrastructurePhase),
		string(skipInitPhase),
		string(skipAttestationConfigPhase),
		string(skipCertSANsPhase),
		string(skipHelmPhase),
		string(skipImagePhase),
		string(skipK8sPhase),
	}
}

// formatSkipPhases returns a formatted string of all phases that can be skipped.
func formatSkipPhases() string {
	return fmt.Sprintf("{ %s }", strings.Join(allPhases(), " | "))
}

// skipPhase is a phase of the upgrade process that can be skipped.
type skipPhase string

// skipPhases is a list of phases that can be skipped during the upgrade process.
type skipPhases map[skipPhase]struct{}

// contains returns true if the list of phases contains the given phase.
func (s skipPhases) contains(phase skipPhase) bool {
	_, ok := s[skipPhase(strings.ToLower(string(phase)))]
	return ok
}

// add a phase to the list of phases.
func (s *skipPhases) add(phases ...skipPhase) {
	if *s == nil {
		*s = make(skipPhases)
	}
	for _, phase := range phases {
		(*s)[skipPhase(strings.ToLower(string(phase)))] = struct{}{}
	}
}

// NewApplyCmd creates the apply command.
func NewApplyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply a configuration to a Constellation cluster",
		Long:  "Apply a configuration to a Constellation cluster to initialize or upgrade the cluster.",
		Args:  cobra.NoArgs,
		RunE:  runApply,
	}

	cmd.Flags().Bool("conformance", false, "enable conformance mode")
	cmd.Flags().Bool("skip-helm-wait", false, "install helm charts without waiting for deployments to be ready")
	cmd.Flags().Bool("merge-kubeconfig", false, "merge Constellation kubeconfig file with default kubeconfig file in $HOME/.kube/config")
	cmd.Flags().BoolP("yes", "y", false, "run command without further confirmation\n"+
		"WARNING: the command might delete or update existing resources without additional checks. Please read the docs.\n")
	cmd.Flags().Duration("timeout", 5*time.Minute, "change helm upgrade timeout\n"+
		"Might be useful for slow connections or big clusters.")
	cmd.Flags().StringSlice("skip-phases", nil, "comma-separated list of upgrade phases to skip\n"+
		fmt.Sprintf("one or multiple of %s", formatSkipPhases()))

	must(cmd.Flags().MarkHidden("timeout"))

	return cmd
}

// applyFlags defines the flags for the apply command.
type applyFlags struct {
	rootFlags
	yes            bool
	conformance    bool
	mergeConfigs   bool
	upgradeTimeout time.Duration
	helmWaitMode   helm.WaitMode
	skipPhases     skipPhases
}

// parse the apply command flags.
func (f *applyFlags) parse(flags *pflag.FlagSet) error {
	if err := f.rootFlags.parse(flags); err != nil {
		return err
	}

	rawSkipPhases, err := flags.GetStringSlice("skip-phases")
	if err != nil {
		return fmt.Errorf("getting 'skip-phases' flag: %w", err)
	}
	var skipPhases skipPhases
	for _, phase := range rawSkipPhases {
		phase = strings.ToLower(phase)
		if slices.Contains(allPhases(), phase) {
			skipPhases.add(skipPhase(phase))
		} else {
			return fmt.Errorf("invalid phase %s", phase)
		}
	}
	f.skipPhases = skipPhases

	f.yes, err = flags.GetBool("yes")
	if err != nil {
		return fmt.Errorf("getting 'yes' flag: %w", err)
	}

	f.upgradeTimeout, err = flags.GetDuration("timeout")
	if err != nil {
		return fmt.Errorf("getting 'timeout' flag: %w", err)
	}

	f.conformance, err = flags.GetBool("conformance")
	if err != nil {
		return fmt.Errorf("getting 'conformance' flag: %w", err)
	}

	skipHelmWait, err := flags.GetBool("skip-helm-wait")
	if err != nil {
		return fmt.Errorf("getting 'skip-helm-wait' flag: %w", err)
	}
	f.helmWaitMode = helm.WaitModeAtomic
	if skipHelmWait {
		f.helmWaitMode = helm.WaitModeNone
	}

	f.mergeConfigs, err = flags.GetBool("merge-kubeconfig")
	if err != nil {
		return fmt.Errorf("getting 'merge-kubeconfig' flag: %w", err)
	}
	return nil
}

// runApply sets up the apply command and runs it.
func runApply(cmd *cobra.Command, _ []string) error {
	log, err := newCLILogger(cmd)
	if err != nil {
		return fmt.Errorf("creating logger: %w", err)
	}
	defer log.Sync()
	spinner, err := newSpinnerOrStderr(cmd)
	if err != nil {
		return err
	}
	defer spinner.Stop()

	flags := applyFlags{}
	if err := flags.parse(cmd.Flags()); err != nil {
		return err
	}

	fileHandler := file.NewHandler(afero.NewOsFs())

	newDialer := func(validator atls.Validator) *dialer.Dialer {
		return dialer.New(nil, validator, &net.Dialer{})
	}
	newKubeUpgrader := func(w io.Writer, kubeConfigPath string, log debugLog) (kubernetesUpgrader, error) {
		return kubecmd.New(w, kubeConfigPath, fileHandler, log)
	}
	newHelmClient := func(kubeConfigPath string, log debugLog) (helmApplier, error) {
		return helm.NewClient(kubeConfigPath, log)
	}

	upgradeID := generateUpgradeID(upgradeCmdKindApply)
	upgradeDir := filepath.Join(constants.UpgradeDir, upgradeID)

	newInfraApplier := func(ctx context.Context) (cloudApplier, func(), error) {
		return cloudcmd.NewApplier(
			ctx,
			spinner,
			constants.TerraformWorkingDir,
			upgradeDir,
			flags.tfLogLevel,
			fileHandler,
		)
	}

	apply := &applyCmd{
		fileHandler:     fileHandler,
		flags:           flags,
		log:             log,
		spinner:         spinner,
		merger:          &kubeconfigMerger{log: log},
		quotaChecker:    license.NewClient(),
		newHelmClient:   newHelmClient,
		newDialer:       newDialer,
		newKubeUpgrader: newKubeUpgrader,
		newInfraApplier: newInfraApplier,
	}

	ctx, cancel := context.WithTimeout(cmd.Context(), time.Hour)
	defer cancel()
	cmd.SetContext(ctx)

	return apply.apply(cmd, attestationconfigapi.NewFetcher(), upgradeDir)
}

type applyCmd struct {
	fileHandler file.Handler
	flags       applyFlags

	log     debugLog
	spinner spinnerInterf

	merger       configMerger
	quotaChecker license.QuotaChecker

	newHelmClient   func(kubeConfigPath string, log debugLog) (helmApplier, error)
	newDialer       func(validator atls.Validator) *dialer.Dialer
	newKubeUpgrader func(out io.Writer, kubeConfigPath string, log debugLog) (kubernetesUpgrader, error)
	newInfraApplier func(context.Context) (cloudApplier, func(), error)
}

/*
apply updates a Constellation cluster by applying a user's config.
The control flow is as follows:

	                          ┌───────▼───────┐
	                          │Parse Flags    │
	                          │               │
	                          │Read Config    │
	                          │               │
	                          │Read State-File│
	                          │               │
	                          │Validate input │
	                          └───────┬───────┘
	                                  │                          ───┐
	               ┌──────────────────▼───────────────────┐         │
	               │Check if Terraform state is up to date│         │
	               └──────────────────┬──┬────────────────┘         │
	                                  │  │Not up to date            │
	                                  │  │(Diff from Terraform plan)│
	                                  │  └────────────┐             │
	                                  │               │             |Infrastructure
	                                  │  ┌────────────▼──────────┐  │Phase
	                                  │  │Apply Terraform updates│  │
	                                  │  └────────────┬──────────┘  │
	                                  │               │             │
	                                  │  ┌────────────┘             │
	                                  │  │                       ───┘
	               ┌──────────────────▼──▼────────────┐
	               │Check for constellation-admin.conf│
	               └───────────────┬──┬───────────────┘
	            File does not exist│  │
	               ┌───────────────┘  │                          ───┐
	               │                  │                             │
	  ┌────────────▼────────────┐     │                             │
	  │Run Bootstrapper Init RPC│     │                             │
	  └────────────┬────────────┘     │File does exist              │
	               │                  │                             │
	┌──────────────▼───────────────┐  │                             │Init
	│Write constellation-admin.conf│  │                             │Phase
	└──────────────┬───────────────┘  │                             │
	               │                  │                             │
	┌──────────────▼───────────────┐  │                             │
	│Prepare "Init Success" Message│  │                             │
	└──────────────┬───────────────┘  │                             │
	               │                  │                             │
	               └───────────────┐  │                          ───┘
	                               │  │                          ───┐
	                    ┌──────────▼──▼──────────┐                  │AttestationConfig
	                    │Apply Attestation Config│                  │Phase
	                    └─────────────┬──────────┘               ───┘
	                                  │                          ───┐
	                   ┌──────────────▼────────────┐                │CertSANs
	                   │Extend API Server Cert SANs│                │Phase
	                   └──────────────┬────────────┘             ───┘
	                                  │                          ───┐
	                       ┌──────────▼────────┐                    │Helm
	                       │ Apply Helm Charts │                    │Phase
	                       └──────────┬────────┘                 ───┘
	                                  │                          ───┐
	                    ┌─────────────▼────────────┐                │
	     Can be skipped │Upgrade NodeVersion object│                │K8s/Image
	  if we ran Init RP │  (Image and K8s update)  │                │Phase
	                    └─────────────┬────────────┘                │
	                                  │                          ───┘
	                        ┌─────────▼──────────┐
	                        │Write success output│
	                        └────────────────────┘
*/
func (a *applyCmd) apply(cmd *cobra.Command, configFetcher attestationconfigapi.Fetcher, upgradeDir string) error {
	// Validate inputs
	conf, stateFile, err := a.validateInputs(cmd, configFetcher)
	if err != nil {
		return err
	}

	// Now start actually running the apply command

	// Check current Terraform state, if it exists and infrastructure upgrades are not skipped,
	// and apply migrations if necessary.
	if !a.flags.skipPhases.contains(skipInfrastructurePhase) {
		if err := a.runTerraformApply(cmd, conf, stateFile, upgradeDir); err != nil {
			return fmt.Errorf("applying Terraform configuration : %w", err)
		}
	}

	bufferedOutput := &bytes.Buffer{}
	// Run init RPC if required
	if !a.flags.skipPhases.contains(skipInitPhase) {
		bufferedOutput, err = a.runInit(cmd, conf, stateFile)
		if err != nil {
			return err
		}
	}

	// From now on we can assume a valid Kubernetes admin config file exists
	a.log.Debugf("Creating Kubernetes client using %s", a.flags.pathPrefixer.PrefixPrintablePath(constants.AdminConfFilename))
	kubeUpgrader, err := a.newKubeUpgrader(cmd.OutOrStdout(), constants.AdminConfFilename, a.log)
	if err != nil {
		return err
	}

	// Apply Attestation Config
	if !a.flags.skipPhases.contains(skipAttestationConfigPhase) {
		a.log.Debugf("Applying new attestation config to cluster")
		if err := a.applyJoinConfig(cmd, kubeUpgrader, conf.GetAttestationConfig(), stateFile.ClusterValues.MeasurementSalt); err != nil {
			return fmt.Errorf("applying attestation config: %w", err)
		}
	}

	// Extend API Server Cert SANs
	if !a.flags.skipPhases.contains(skipCertSANsPhase) {
		sans := append([]string{stateFile.Infrastructure.ClusterEndpoint, conf.CustomEndpoint}, stateFile.Infrastructure.APIServerCertSANs...)
		if err := kubeUpgrader.ExtendClusterConfigCertSANs(cmd.Context(), sans); err != nil {
			return fmt.Errorf("extending cert SANs: %w", err)
		}
	}

	// Apply Helm Charts
	if !a.flags.skipPhases.contains(skipHelmPhase) {
		if err := a.runHelmApply(cmd, conf, stateFile, kubeUpgrader, upgradeDir); err != nil {
			return err
		}
	}

	// Upgrade NodeVersion object
	// This can be skipped if we ran the init RPC, as the NodeVersion object is already up to date
	if !(a.flags.skipPhases.contains(skipK8sPhase) && a.flags.skipPhases.contains(skipImagePhase)) &&
		a.flags.skipPhases.contains(skipInitPhase) {
		if err := a.runK8sUpgrade(cmd, conf, kubeUpgrader); err != nil {
			return err
		}
	}

	// Write success output
	cmd.Print(bufferedOutput.String())

	return nil
}

func (a *applyCmd) validateInputs(cmd *cobra.Command, configFetcher attestationconfigapi.Fetcher) (*config.Config, *state.State, error) {
	// Read user's config and state file
	a.log.Debugf("Reading config from %s", a.flags.pathPrefixer.PrefixPrintablePath(constants.ConfigFilename))
	conf, err := config.New(a.fileHandler, constants.ConfigFilename, configFetcher, a.flags.force)
	var configValidationErr *config.ValidationError
	if errors.As(err, &configValidationErr) {
		cmd.PrintErrln(configValidationErr.LongMessage())
	}
	if err != nil {
		return nil, nil, err
	}

	// Read and validate state file
	a.log.Debugf("Reading state file from %s", a.flags.pathPrefixer.PrefixPrintablePath(constants.StateFilename))
	stateFile, err := state.ReadFromFile(a.fileHandler, constants.StateFilename)
	if err != nil {
		return nil, nil, err
	}
	if a.flags.skipPhases.contains(skipInitPhase) {
		// If the skipInit flag is set, we are in a state where the cluster
		// has already been initialized and check against the respective constraints.
		if err := stateFile.Validate(state.PostInit, conf.GetProvider()); err != nil {
			return nil, nil, err
		}
	} else {
		// The cluster has not been initialized yet, so we check against the pre-init constraints.
		if err := stateFile.Validate(state.PreInit, conf.GetProvider()); err != nil {
			return nil, nil, err
		}
	}

	// Check license
	a.log.Debugf("Running license check")
	checker := license.NewChecker(a.quotaChecker, a.fileHandler)
	if err := checker.CheckLicense(cmd.Context(), conf.GetProvider(), conf.Provider, cmd.Printf); err != nil {
		cmd.PrintErrf("License check failed: %v", err)
	}
	a.log.Debugf("Checked license")

	// Check if we already have a running Kubernetes cluster
	// by checking if the Kubernetes admin config file exists
	// If it exist, we skip the init phase
	// If it does not exist, we need to run the init RPC first
	// This may break things further down the line
	// It is the user's responsibility to make sure the cluster is in a valid state
	a.log.Debugf("Checking if %s exists", a.flags.pathPrefixer.PrefixPrintablePath(constants.AdminConfFilename))
	if _, err := a.fileHandler.Stat(constants.AdminConfFilename); err == nil {
		a.flags.skipPhases.add(skipInitPhase)
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, nil, fmt.Errorf("checking for %s: %w", a.flags.pathPrefixer.PrefixPrintablePath(constants.AdminConfFilename), err)
	}
	a.log.Debugf("Init RPC required: %t", !a.flags.skipPhases.contains(skipInitPhase))

	// Validate input arguments

	// Validate Kubernetes version as set in the user's config
	// If we need to run the init RPC, the version has to be valid
	// Otherwise, we are able to use an outdated version, meaning we skip the K8s upgrade
	// We skip version validation if the user explicitly skips the Kubernetes phase
	a.log.Debugf("Validating Kubernetes version %s", conf.KubernetesVersion)
	validVersion, err := versions.NewValidK8sVersion(string(conf.KubernetesVersion), true)
	if err != nil && !a.flags.skipPhases.contains(skipK8sPhase) {
		a.log.Debugf("Kubernetes version not valid: %s", err)
		if !a.flags.skipPhases.contains(skipInitPhase) {
			return nil, nil, err
		}
		a.log.Debugf("Checking if user wants to continue anyway")
		if !a.flags.yes {
			confirmed, err := askToConfirm(cmd,
				fmt.Sprintf(
					"WARNING: The Kubernetes patch version %s is not supported. If you continue, Kubernetes upgrades will be skipped. Do you want to continue anyway?",
					validVersion,
				),
			)
			if err != nil {
				return nil, nil, fmt.Errorf("asking for confirmation: %w", err)
			}
			if !confirmed {
				return nil, nil, fmt.Errorf("aborted by user")
			}
		}
		a.flags.skipPhases.add(skipK8sPhase)
		a.log.Debugf("Outdated Kubernetes version accepted, Kubernetes upgrade will be skipped")
	}
	if versions.IsPreviewK8sVersion(validVersion) {
		cmd.PrintErrf("Warning: Constellation with Kubernetes %s is still in preview. Use only for evaluation purposes.\n", validVersion)
	}
	conf.KubernetesVersion = validVersion
	a.log.Debugf("Target Kubernetes version set to %s", conf.KubernetesVersion)

	// Validate microservice version (helm versions) in the user's config matches the version of the CLI
	// This makes sure we catch potential errors early, not just after we already ran Terraform migrations or the init RPC
	if !a.flags.force && !a.flags.skipPhases.contains(skipHelmPhase) && !a.flags.skipPhases.contains(skipInitPhase) {
		if err := validateCLIandConstellationVersionAreEqual(constants.BinaryVersion(), conf.Image, conf.MicroserviceVersion); err != nil {
			return nil, nil, err
		}
	}

	// Constellation on QEMU or OpenStack don't support upgrades
	// If using one of those providers, make sure the command is only used to initialize a cluster
	if !(conf.GetProvider() == cloudprovider.AWS || conf.GetProvider() == cloudprovider.Azure || conf.GetProvider() == cloudprovider.GCP) {
		if a.flags.skipPhases.contains(skipInitPhase) {
			return nil, nil, fmt.Errorf("upgrades are not supported for provider %s", conf.GetProvider())
		}
		// Skip Terraform phase
		a.log.Debugf("Skipping Infrastructure upgrade")
		a.flags.skipPhases.add(skipInfrastructurePhase)
	}

	// Check if Terraform state exists
	if tfStateExists, err := a.tfStateExists(); err != nil {
		return nil, nil, fmt.Errorf("checking Terraform state: %w", err)
	} else if !tfStateExists {
		a.flags.skipPhases.add(skipInfrastructurePhase)
		a.log.Debugf("No Terraform state found in current working directory. Assuming self-managed infrastructure. Infrastructure upgrades will not be performed.")
	}

	// Print warning about AWS attestation
	// TODO(derpsteb): remove once AWS fixes SEV-SNP attestation provisioning issues
	if !a.flags.skipPhases.contains(skipInitPhase) && conf.GetAttestationConfig().GetVariant().Equal(variant.AWSSEVSNP{}) {
		cmd.PrintErrln("WARNING: Attestation temporarily relies on AWS nitroTPM. See https://docs.edgeless.systems/constellation/workflows/config#choosing-a-vm-type for more information.")
	}

	return conf, stateFile, nil
}

// applyJoincConfig creates or updates the cluster's join config.
// If the config already exists, and is different from the new config, the user is asked to confirm the upgrade.
func (a *applyCmd) applyJoinConfig(
	cmd *cobra.Command, kubeUpgrader kubernetesUpgrader, newConfig config.AttestationCfg, measurementSalt []byte,
) error {
	clusterAttestationConfig, err := kubeUpgrader.GetClusterAttestationConfig(cmd.Context(), newConfig.GetVariant())
	if err != nil {
		a.log.Debugf("Getting cluster attestation config failed: %s", err)
		if k8serrors.IsNotFound(err) {
			a.log.Debugf("Creating new join config")
			return kubeUpgrader.ApplyJoinConfig(cmd.Context(), newConfig, measurementSalt)
		}
		return fmt.Errorf("getting cluster attestation config: %w", err)
	}

	// If the current config is equal, or there is an error when comparing the configs, we skip the upgrade.
	equal, err := newConfig.EqualTo(clusterAttestationConfig)
	if err != nil {
		return fmt.Errorf("comparing attestation configs: %w", err)
	}
	if equal {
		a.log.Debugf("Current attestation config is equal to the new config, nothing to do")
		return nil
	}

	cmd.Println("The configured attestation config is different from the attestation config in the cluster.")
	diffStr, err := diffAttestationCfg(clusterAttestationConfig, newConfig)
	if err != nil {
		return fmt.Errorf("diffing attestation configs: %w", err)
	}
	cmd.Println("The following changes will be applied to the attestation config:")
	cmd.Println(diffStr)
	if !a.flags.yes {
		ok, err := askToConfirm(cmd, "Are you sure you want to change your cluster's attestation config?")
		if err != nil {
			return fmt.Errorf("asking for confirmation: %w", err)
		}
		if !ok {
			return errors.New("aborting upgrade since attestation config is different")
		}
	}

	if err := kubeUpgrader.ApplyJoinConfig(cmd.Context(), newConfig, measurementSalt); err != nil {
		return fmt.Errorf("updating attestation config: %w", err)
	}
	cmd.Println("Successfully updated the cluster's attestation config")

	return nil
}

// runK8sUpgrade upgrades image and Kubernetes version of the Constellation cluster.
func (a *applyCmd) runK8sUpgrade(cmd *cobra.Command, conf *config.Config, kubeUpgrader kubernetesUpgrader,
) error {
	err := kubeUpgrader.UpgradeNodeVersion(
		cmd.Context(), conf, a.flags.force,
		a.flags.skipPhases.contains(skipImagePhase),
		a.flags.skipPhases.contains(skipK8sPhase),
	)

	var upgradeErr *compatibility.InvalidUpgradeError
	switch {
	case errors.Is(err, kubecmd.ErrInProgress):
		cmd.PrintErrln("Skipping image and Kubernetes upgrades. Another upgrade is in progress.")
	case errors.As(err, &upgradeErr):
		cmd.PrintErrln(err)
	case err != nil:
		return fmt.Errorf("upgrading NodeVersion: %w", err)
	}

	return nil
}

// tfStateExists checks whether a Constellation Terraform state exists in the current working directory.
func (a *applyCmd) tfStateExists() (bool, error) {
	_, err := a.fileHandler.Stat(constants.TerraformWorkingDir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, nil
		}
		return false, fmt.Errorf("reading Terraform state: %w", err)
	}
	return true, nil
}
