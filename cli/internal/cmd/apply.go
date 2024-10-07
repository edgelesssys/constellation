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
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/edgelesssys/constellation/v2/cli/internal/cloudcmd"
	"github.com/edgelesssys/constellation/v2/internal/api/attestationconfigapi"
	"github.com/edgelesssys/constellation/v2/internal/api/versionsapi"
	"github.com/edgelesssys/constellation/v2/internal/atls"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/compatibility"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/constellation"
	"github.com/edgelesssys/constellation/v2/internal/constellation/helm"
	"github.com/edgelesssys/constellation/v2/internal/constellation/kubecmd"
	"github.com/edgelesssys/constellation/v2/internal/constellation/state"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/grpc/dialer"
	"github.com/edgelesssys/constellation/v2/internal/imagefetcher"
	"github.com/edgelesssys/constellation/v2/internal/kms/uri"
	"github.com/edgelesssys/constellation/v2/internal/semver"
	"github.com/edgelesssys/constellation/v2/internal/versions"
	slogmulti "github.com/samber/slog-multi"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	xsemver "golang.org/x/mod/semver"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
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
func allPhases(except ...skipPhase) []string {
	phases := []string{
		string(skipInfrastructurePhase),
		string(skipInitPhase),
		string(skipAttestationConfigPhase),
		string(skipCertSANsPhase),
		string(skipHelmPhase),
		string(skipImagePhase),
		string(skipK8sPhase),
	}

	var returnedPhases []string
	for idx, phase := range phases {
		if !slices.Contains(except, skipPhase(phase)) {
			returnedPhases = append(returnedPhases, phases[idx])
		}
	}
	return returnedPhases
}

// formatSkipPhases returns a formatted string of all phases that can be skipped.
func formatSkipPhases() string {
	return fmt.Sprintf("{ %s }", strings.Join(allPhases(), " | "))
}

// skipPhase is a phase of the upgrade process that can be skipped.
type skipPhase string

// skipPhases is a list of phases that can be skipped during the upgrade process.
type skipPhases map[skipPhase]struct{}

// contains returns true if skipPhases contains all of the given phases.
func (s skipPhases) contains(phases ...skipPhase) bool {
	for _, phase := range phases {
		if _, ok := s[skipPhase(strings.ToLower(string(phase)))]; !ok {
			return false
		}
	}
	return true
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
	cmd.Flags().Duration("helm-timeout", 10*time.Minute, "change helm install/upgrade timeout\n"+
		"Might be useful for slow connections or big clusters.")
	cmd.Flags().StringSlice("skip-phases", nil, "comma-separated list of upgrade phases to skip\n"+
		fmt.Sprintf("one or multiple of %s", formatSkipPhases()))

	must(cmd.Flags().MarkHidden("helm-timeout"))

	must(cmd.RegisterFlagCompletionFunc("skip-phases", skipPhasesCompletion))
	return cmd
}

// applyFlags defines the flags for the apply command.
type applyFlags struct {
	rootFlags
	yes          bool
	conformance  bool
	mergeConfigs bool
	helmTimeout  time.Duration
	helmWaitMode helm.WaitMode
	skipPhases   skipPhases
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

	f.helmTimeout, err = flags.GetDuration("helm-timeout")
	if err != nil {
		return fmt.Errorf("getting 'helm-timeout' flag: %w", err)
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
	logger, err := newDebugFileLogger(cmd, fileHandler)
	if err != nil {
		return err
	}

	newDialer := func(validator atls.Validator) *dialer.Dialer {
		return dialer.New(nil, validator, &net.Dialer{})
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

	applier := constellation.NewApplier(log, spinner, constellation.ApplyContextCLI, newDialer)

	apply := &applyCmd{
		fileHandler:     fileHandler,
		flags:           flags,
		log:             logger,
		wLog:            &warnLogger{cmd: cmd, log: log},
		spinner:         spinner,
		merger:          &kubeconfigMerger{log: log},
		newInfraApplier: newInfraApplier,
		imageFetcher:    imagefetcher.New(),
		applier:         applier,
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
	wLog    warnLog
	spinner spinnerInterf

	merger configMerger

	imageFetcher imageFetcher
	applier      applier

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
	 if we ran Init RPC │  (Image and K8s update)  │                │Phase
	                    └─────────────┬────────────┘                │
	                                  │                          ───┘
	                        ┌─────────▼──────────┐
	                        │Write success output│
	                        └────────────────────┘
*/
func (a *applyCmd) apply(
	cmd *cobra.Command, configFetcher attestationconfigapi.Fetcher, upgradeDir string,
) error {
	// Validate inputs
	conf, stateFile, err := a.validateInputs(cmd, configFetcher)
	if err != nil {
		return err
	}

	// Check license
	a.checkLicenseFile(cmd, conf.GetProvider(), conf.UseMarketplaceImage())

	// Now start actually running the apply command

	// Check current Terraform state, if it exists and infrastructure upgrades are not skipped,
	// and apply migrations if necessary.
	if !a.flags.skipPhases.contains(skipInfrastructurePhase) {
		if err := a.runTerraformApply(cmd, conf, stateFile, upgradeDir); err != nil {
			return fmt.Errorf("applying Terraform configuration: %w", err)
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

	if a.flags.skipPhases.contains(skipAttestationConfigPhase, skipCertSANsPhase, skipHelmPhase, skipK8sPhase, skipImagePhase) {
		cmd.Print(bufferedOutput.String())
		return nil
	}

	// From now on we can assume a valid Kubernetes admin config file exists
	kubeConfig, err := a.fileHandler.Read(constants.AdminConfFilename)
	if err != nil {
		return fmt.Errorf("reading kubeconfig: %w", err)
	}
	if err := a.applier.SetKubeConfig(kubeConfig); err != nil {
		return err
	}

	// Apply Attestation Config
	if !a.flags.skipPhases.contains(skipAttestationConfigPhase) {
		a.log.Debug("Applying new attestation config to cluster")
		if err := a.applyJoinConfig(cmd, conf.GetAttestationConfig(), stateFile.ClusterValues.MeasurementSalt); err != nil {
			return fmt.Errorf("applying attestation config: %w", err)
		}
	}

	// Extend API Server Cert SANs
	if !a.flags.skipPhases.contains(skipCertSANsPhase) {
		if err := a.applier.ExtendClusterConfigCertSANs(
			cmd.Context(),
			stateFile.Infrastructure.ClusterEndpoint,
			conf.CustomEndpoint,
			stateFile.Infrastructure.APIServerCertSANs,
		); err != nil {
			return fmt.Errorf("extending cert SANs: %w", err)
		}
	}

	// Apply Helm Charts
	if !a.flags.skipPhases.contains(skipHelmPhase) {
		if err := a.applier.AnnotateCoreDNSResources(cmd.Context()); err != nil {
			return fmt.Errorf("annotating CoreDNS: %w", err)
		}
		if err := a.runHelmApply(cmd, conf, stateFile, upgradeDir); err != nil {
			return err
		}
		if err := a.applier.CleanupCoreDNSResources(cmd.Context()); err != nil {
			return fmt.Errorf("cleaning up CoreDNS: %w", err)
		}
	}

	// Upgrade node image
	if !a.flags.skipPhases.contains(skipImagePhase) {
		if err := a.runNodeImageUpgrade(cmd, conf); err != nil {
			return err
		}
	}

	// Upgrade Kubernetes version
	if !a.flags.skipPhases.contains(skipK8sPhase) {
		if err := a.runK8sVersionUpgrade(cmd, conf); err != nil {
			return err
		}
	}

	// Write success output
	cmd.Print(bufferedOutput.String())

	return nil
}

func (a *applyCmd) validateInputs(cmd *cobra.Command, configFetcher attestationconfigapi.Fetcher) (*config.Config, *state.State, error) {
	// Read user's config and state file
	a.log.Debug(fmt.Sprintf("Reading config from %q", a.flags.pathPrefixer.PrefixPrintablePath(constants.ConfigFilename)))
	conf, err := config.New(a.fileHandler, constants.ConfigFilename, configFetcher, a.flags.force)
	var configValidationErr *config.ValidationError
	if errors.As(err, &configValidationErr) {
		cmd.PrintErrln(configValidationErr.LongMessage())
	}
	if err != nil {
		return nil, nil, err
	}

	a.log.Debug(fmt.Sprintf("Reading state file from %q", a.flags.pathPrefixer.PrefixPrintablePath(constants.StateFilename)))
	stateFile, err := state.CreateOrRead(a.fileHandler, constants.StateFilename)
	if err != nil {
		return nil, nil, err
	}

	// Validate the state file and set flags accordingly
	//
	// We don't run "hard" verification of skip-phases flags and state file here,
	// a user may still end up skipping phases that could result in errors later on.
	// However, we perform basic steps, like ensuring init phase is not skipped if
	a.log.Debug("Validating state file")
	preCreateValidateErr := stateFile.Validate(state.PreCreate, conf.GetAttestationConfig().GetVariant())
	preInitValidateErr := stateFile.Validate(state.PreInit, conf.GetAttestationConfig().GetVariant())
	postInitValidateErr := stateFile.Validate(state.PostInit, conf.GetAttestationConfig().GetVariant())

	// If the state file is in a pre-create state, we need to create the cluster,
	// in which case the workspace has to be clean
	if preCreateValidateErr == nil {
		// We can't skip the infrastructure phase if no infrastructure has been defined
		a.log.Debug("State file is in pre-create state, checking workspace")
		if a.flags.skipPhases.contains(skipInfrastructurePhase) {
			return nil, nil, preInitValidateErr
		}

		if err := a.checkCreateFilesClean(); err != nil {
			return nil, nil, err
		}

		a.log.Debug("No Terraform state found in current working directory. Preparing to create a new cluster.")
		printCreateWarnings(cmd.ErrOrStderr(), conf)
	}

	// Check if the state file is in a pre-init OR
	// if in pre-create state and init should not be skipped
	// If so, we need to run the init RPC
	if preInitValidateErr == nil || (preCreateValidateErr == nil && !a.flags.skipPhases.contains(skipInitPhase)) {
		// We can't skip the init phase if the init RPC hasn't been run yet
		a.log.Debug("State file is in pre-init state, checking workspace")
		if a.flags.skipPhases.contains(skipInitPhase) {
			return nil, nil, postInitValidateErr
		}

		if err := a.checkInitFilesClean(); err != nil {
			return nil, nil, err
		}

		// Skip image and k8s phase, since they are covered by the init RPC
		a.flags.skipPhases.add(skipImagePhase, skipK8sPhase)
	}

	// If the state file is in a post-init state,
	// we need to make sure specific files exist in the workspace
	if postInitValidateErr == nil {
		a.log.Debug("State file is in post-init state, checking workspace")
		if err := a.checkPostInitFilesExist(); err != nil {
			return nil, nil, err
		}

		// Skip init phase, since the init RPC has already been run
		a.flags.skipPhases.add(skipInitPhase)
	} else if preCreateValidateErr != nil && preInitValidateErr != nil {
		return nil, nil, postInitValidateErr
	}

	// Validate Kubernetes version as set in the user's config
	// If we need to run the init RPC, the version has to be valid
	// Otherwise, we are able to use an outdated version, meaning we skip the K8s upgrade
	// We skip version validation if the user explicitly skips the Kubernetes phase
	a.log.Debug(fmt.Sprintf("Validating Kubernetes version %q", conf.KubernetesVersion))
	validVersion, err := versions.NewValidK8sVersion(string(conf.KubernetesVersion), true)
	if err != nil {
		a.log.Debug(fmt.Sprintf("Kubernetes version not valid: %q", err))
		if !a.flags.skipPhases.contains(skipInitPhase) {
			return nil, nil, err
		}

		if !a.flags.skipPhases.contains(skipK8sPhase) {
			a.log.Debug("Checking if user wants to continue anyway")
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
			a.log.Debug("Outdated Kubernetes version accepted, Kubernetes upgrade will be skipped")
		}

		validVersionString, err := versions.ResolveK8sPatchVersion(xsemver.MajorMinor(string(conf.KubernetesVersion)))
		if err != nil {
			return nil, nil, fmt.Errorf("resolving Kubernetes patch version: %w", err)
		}
		validVersion, err = versions.NewValidK8sVersion(validVersionString, true)
		if err != nil {
			return nil, nil, fmt.Errorf("parsing Kubernetes version: %w", err)
		}
	}
	if versions.IsPreviewK8sVersion(validVersion) {
		cmd.PrintErrf("Warning: Constellation with Kubernetes %s is still in preview. Use only for evaluation purposes.\n", validVersion)
	}
	conf.KubernetesVersion = validVersion
	a.log.Debug(fmt.Sprintf("Target Kubernetes version set to %q", conf.KubernetesVersion))

	// Validate microservice version (helm versions) in the user's config matches the version of the CLI
	// This makes sure we catch potential errors early, not just after we already ran Terraform migrations or the init RPC
	if !a.flags.force && !a.flags.skipPhases.contains(skipHelmPhase, skipInitPhase) {
		if err := validateCLIandConstellationVersionAreEqual(constants.BinaryVersion(), conf.Image, conf.MicroserviceVersion); err != nil {
			return nil, nil, err
		}
	}

	// Constellation does not support image upgrades on all CSPs. Not supported are: QEMU, OpenStack
	// If using one of those providers, print a warning when trying to upgrade the image
	if !(conf.GetProvider() == cloudprovider.AWS || conf.GetProvider() == cloudprovider.Azure || conf.GetProvider() == cloudprovider.GCP) &&
		!a.flags.skipPhases.contains(skipImagePhase) {
		cmd.PrintErrf("Image upgrades are not supported for provider %s\n", conf.GetProvider())
		cmd.PrintErrln("Image phase will be skipped")
		a.flags.skipPhases.add(skipImagePhase)
	}

	return conf, stateFile, nil
}

// applyJoinConfig creates or updates the cluster's join config.
// If the config already exists, and is different from the new config, the user is asked to confirm the upgrade.
func (a *applyCmd) applyJoinConfig(cmd *cobra.Command, newConfig config.AttestationCfg, measurementSalt []byte,
) error {
	clusterAttestationConfig, err := a.applier.GetClusterAttestationConfig(cmd.Context(), newConfig.GetVariant())
	if err != nil {
		a.log.Debug(fmt.Sprintf("Getting cluster attestation config failed: %q", err))
		if k8serrors.IsNotFound(err) {
			a.log.Debug("Creating new join config")
			return a.applier.ApplyJoinConfig(cmd.Context(), newConfig, measurementSalt)
		}
		return fmt.Errorf("getting cluster attestation config: %w", err)
	}

	// If the current config is equal, or there is an error when comparing the configs, we skip the upgrade.
	equal, err := newConfig.EqualTo(clusterAttestationConfig)
	if err != nil {
		return fmt.Errorf("comparing attestation configs: %w", err)
	}
	if equal {
		a.log.Debug("Current attestation config is equal to the new config, nothing to do")
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

	if err := a.applier.ApplyJoinConfig(cmd.Context(), newConfig, measurementSalt); err != nil {
		return fmt.Errorf("updating attestation config: %w", err)
	}
	cmd.Println("Successfully updated the cluster's attestation config")

	return nil
}

func (a *applyCmd) runNodeImageUpgrade(cmd *cobra.Command, conf *config.Config) error {
	provider := conf.GetProvider()
	attestationVariant := conf.GetAttestationConfig().GetVariant()
	region := conf.GetRegion()
	imageReference, err := a.imageFetcher.FetchReference(cmd.Context(), provider, attestationVariant, conf.Image, region, conf.UseMarketplaceImage())
	if err != nil {
		return fmt.Errorf("fetching image reference: %w", err)
	}

	imageVersionInfo, err := versionsapi.NewVersionFromShortPath(conf.Image, versionsapi.VersionKindImage)
	if err != nil {
		return fmt.Errorf("parsing version from image short path: %w", err)
	}
	imageVersion, err := semver.New(imageVersionInfo.Version())
	if err != nil {
		return fmt.Errorf("parsing image version: %w", err)
	}

	err = a.applier.UpgradeNodeImage(cmd.Context(), imageVersion, imageReference, a.flags.force)
	var upgradeErr *compatibility.InvalidUpgradeError
	switch {
	case errors.Is(err, kubecmd.ErrInProgress):
		cmd.PrintErrln("Skipping image upgrade: Another upgrade is already in progress.")
	case errors.As(err, &upgradeErr):
		cmd.PrintErrln(err)
	case err != nil:
		return fmt.Errorf("upgrading NodeVersion: %w", err)
	}

	return nil
}

func (a *applyCmd) runK8sVersionUpgrade(cmd *cobra.Command, conf *config.Config) error {
	err := a.applier.UpgradeKubernetesVersion(cmd.Context(), conf.KubernetesVersion, a.flags.force)
	var upgradeErr *compatibility.InvalidUpgradeError
	switch {
	case errors.As(err, &upgradeErr):
		cmd.PrintErrln(err)
	case err != nil:
		return fmt.Errorf("upgrading Kubernetes version: %w", err)
	}

	return nil
}

// checkCreateFilesClean ensures that the workspace is clean before creating a new cluster.
func (a *applyCmd) checkCreateFilesClean() error {
	if err := a.checkInitFilesClean(); err != nil {
		return err
	}
	a.log.Debug("Checking Terraform state")
	if _, err := a.fileHandler.Stat(constants.TerraformWorkingDir); err == nil {
		return fmt.Errorf(
			"terraform state %q already exists in working directory, run 'constellation terminate' before creating a new cluster",
			a.flags.pathPrefixer.PrefixPrintablePath(constants.TerraformWorkingDir),
		)
	} else if !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("checking for %s: %w", a.flags.pathPrefixer.PrefixPrintablePath(constants.TerraformWorkingDir), err)
	}

	return nil
}

// checkInitFilesClean ensures that the workspace is clean before running the init RPC.
func (a *applyCmd) checkInitFilesClean() error {
	a.log.Debug("Checking admin configuration file")
	if _, err := a.fileHandler.Stat(constants.AdminConfFilename); err == nil {
		return fmt.Errorf(
			"file %q already exists in working directory, run 'constellation terminate' before creating a new cluster",
			a.flags.pathPrefixer.PrefixPrintablePath(constants.AdminConfFilename),
		)
	} else if !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("checking for %q: %w", a.flags.pathPrefixer.PrefixPrintablePath(constants.AdminConfFilename), err)
	}
	a.log.Debug("Checking master secrets file")
	if _, err := a.fileHandler.Stat(constants.MasterSecretFilename); err == nil {
		return fmt.Errorf(
			"file %q already exists in working directory. Constellation won't overwrite previous master secrets. Move it somewhere or delete it before creating a new cluster",
			a.flags.pathPrefixer.PrefixPrintablePath(constants.MasterSecretFilename),
		)
	} else if !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("checking for %q: %w", a.flags.pathPrefixer.PrefixPrintablePath(constants.MasterSecretFilename), err)
	}

	return nil
}

// checkPostInitFilesExist ensures that the workspace contains the files from a previous init RPC.
func (a *applyCmd) checkPostInitFilesExist() error {
	if _, err := a.fileHandler.Stat(constants.AdminConfFilename); err != nil {
		return fmt.Errorf("checking for %q: %w", a.flags.pathPrefixer.PrefixPrintablePath(constants.AdminConfFilename), err)
	}
	if _, err := a.fileHandler.Stat(constants.MasterSecretFilename); err != nil {
		return fmt.Errorf("checking for %q: %w", a.flags.pathPrefixer.PrefixPrintablePath(constants.MasterSecretFilename), err)
	}
	return nil
}

func printCreateWarnings(out io.Writer, conf *config.Config) {
	var printedAWarning bool
	if !conf.IsReleaseImage() {
		fmt.Fprintln(out, "Configured image doesn't look like a released production image. Double check image before deploying to production.")
		printedAWarning = true
	}

	if conf.IsNamedLikeDebugImage() && !conf.IsDebugCluster() {
		fmt.Fprintln(out, "WARNING: A debug image is used but debugCluster is false.")
		printedAWarning = true
	}

	if conf.IsDebugCluster() {
		fmt.Fprintln(out, "WARNING: Creating a debug cluster. This cluster is not secure and should only be used for debugging purposes.")
		fmt.Fprintln(out, "DO NOT USE THIS CLUSTER IN PRODUCTION.")
		printedAWarning = true
	}

	if conf.GetAttestationConfig().GetVariant().Equal(variant.AzureTrustedLaunch{}) {
		fmt.Fprintln(out, "Disabling Confidential VMs is insecure. Use only for evaluation purposes.")
		printedAWarning = true
	}

	// Print an extra new line later to separate warnings from the prompt message of the create command
	if printedAWarning {
		fmt.Fprintln(out, "")
	}
}

// skipPhasesCompletion returns suggestions for the skip-phases flag.
// We suggest completion for all phases that can be skipped.
// The phases may be given in any order, as a comma-separated list.
// For example, "skip-phases helm,init" should suggest all phases but "helm" and "init".
func skipPhasesCompletion(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	skippedPhases := strings.Split(toComplete, ",")
	if skippedPhases[0] == "" {
		// No phases were typed yet, so suggest all phases
		return allPhases(), cobra.ShellCompDirectiveNoFileComp
	}

	// Determine what phases have already been typed by the user
	phases := make(map[string]struct{})
	for _, phase := range allPhases() {
		phases[phase] = struct{}{}
	}
	for _, phase := range skippedPhases {
		delete(phases, phase)
	}

	// Get the last phase typed by the user
	// This is the phase we want to complete
	lastPhase := skippedPhases[len(skippedPhases)-1]
	fullyTypedPhases := strings.TrimSuffix(toComplete, lastPhase)

	// Add all phases that have not been typed yet to the suggestions
	// The suggestion is the fully typed phases + the phase that is being completed
	var suggestions []string
	for phase := range phases {
		if strings.HasPrefix(phase, lastPhase) {
			suggestions = append(suggestions, fmt.Sprintf("%s%s", fullyTypedPhases, phase))
		}
	}

	return suggestions, cobra.ShellCompDirectiveNoFileComp
}

// warnLogger implements logging of warnings for validators.
type warnLogger struct {
	cmd *cobra.Command
	log debugLog
}

// Info messages are reduced to debug messages, since we don't want
// the extra info when using the CLI without setting the debug flag.
func (wl warnLogger) Info(msg string, args ...any) {
	wl.log.Debug(msg, args...)
}

// Warn prints a formatted warning from the validator.
func (wl warnLogger) Warn(msg string, args ...any) {
	wl.cmd.PrintErrf("Warning: %s %s\n", msg, fmt.Sprint(args...))
}

type warnLog interface {
	Warn(msg string, args ...any)
	Info(msg string, args ...any)
}

// applier is used to run the different phases of the apply command.
type applier interface {
	SetKubeConfig(kubeConfig []byte) error
	CheckLicense(ctx context.Context, csp cloudprovider.Provider, initRequest bool, licenseID string) (int, error)

	// methods required by "init"

	GenerateMasterSecret() (uri.MasterSecret, error)
	GenerateMeasurementSalt() ([]byte, error)
	Init(
		ctx context.Context, validator atls.Validator, state *state.State,
		clusterLogWriter io.Writer, payload constellation.InitPayload,
	) (constellation.InitOutput, error)

	// methods required to install/upgrade Helm charts

	AnnotateCoreDNSResources(context.Context) error
	CleanupCoreDNSResources(context.Context) error
	PrepareHelmCharts(
		flags helm.Options, state *state.State, serviceAccURI string, masterSecret uri.MasterSecret,
	) (helm.Applier, bool, error)

	// methods to interact with Kubernetes

	ExtendClusterConfigCertSANs(ctx context.Context, clusterEndpoint, customEndpoint string, additionalAPIServerCertSANs []string) error
	GetClusterAttestationConfig(ctx context.Context, variant variant.Variant) (config.AttestationCfg, error)
	ApplyJoinConfig(ctx context.Context, newAttestConfig config.AttestationCfg, measurementSalt []byte) error
	UpgradeNodeImage(ctx context.Context, imageVersion semver.Semver, imageReference string, force bool) error
	UpgradeKubernetesVersion(ctx context.Context, kubernetesVersion versions.ValidK8sVersion, force bool) error
	BackupCRDs(ctx context.Context, fileHandler file.Handler, upgradeDir string) ([]apiextensionsv1.CustomResourceDefinition, error)
	BackupCRs(ctx context.Context, fileHandler file.Handler, crds []apiextensionsv1.CustomResourceDefinition, upgradeDir string) error
}

// imageFetcher gets an image reference from the versionsapi.
type imageFetcher interface {
	FetchReference(ctx context.Context,
		provider cloudprovider.Provider, attestationVariant variant.Variant,
		image, region string, useMarketplaceImage bool,
	) (string, error)
}

func newDebugFileLogger(cmd *cobra.Command, fileHandler file.Handler) (debugLog, error) {
	logLvl := slog.LevelInfo
	debugLog, err := cmd.Flags().GetBool("debug")
	if err != nil {
		return nil, err
	}
	if debugLog {
		logLvl = slog.LevelDebug
	}

	fileWriter := &fileWriter{
		fileHandler: fileHandler,
	}
	return slog.New(
		slogmulti.Fanout(
			slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{AddSource: true, Level: logLvl}),           // first handler: stderr at log level
			slog.NewJSONHandler(fileWriter, &slog.HandlerOptions{AddSource: true, Level: slog.LevelDebug}), // second handler: debug JSON log to file
		),
	), nil
}

type fileWriter struct {
	fileHandler file.Handler
}

// Write satisfies the io.Writer interface by writing a message to file.
func (l *fileWriter) Write(msg []byte) (int, error) {
	err := l.fileHandler.Write(constants.CLIDebugLogFile, msg, file.OptAppend)
	return len(msg), err
}
