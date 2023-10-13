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
	"net"
	"os"
	"path/filepath"
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

type applyFlags struct {
	rootFlags
	yes            bool
	conformance    bool
	mergeConfigs   bool
	upgradeTimeout time.Duration
	helmWaitMode   helm.WaitMode
	skipPhases     skipPhases
}

func (f *applyFlags) parse(flags *pflag.FlagSet) error {
	if err := f.rootFlags.parse(flags); err != nil {
		return err
	}

	rawSkipPhases, err := flags.GetStringSlice("skip-phases")
	if err != nil {
		return fmt.Errorf("getting 'skip-phases' flag: %w", err)
	}
	var skipPhases []skipPhase
	for _, phase := range rawSkipPhases {
		switch skipPhase(phase) {
		case skipInfrastructurePhase, skipHelmPhase, skipImagePhase, skipK8sPhase:
			skipPhases = append(skipPhases, skipPhase(phase))
		default:
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
	clusterUpgrader, err := cloudcmd.NewClusterUpgrader(
		cmd.Context(),
		constants.TerraformWorkingDir,
		upgradeDir,
		flags.tfLogLevel,
		fileHandler,
	)
	if err != nil {
		return fmt.Errorf("setting up cluster upgrader: %w", err)
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
		clusterUpgrader: clusterUpgrader,
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
	newKubeUpgrader func(io.Writer, string, debugLog) (kubernetesUpgrader, error)
	clusterUpgrader clusterUpgrader
}

/*
		 ~~~~~~~~~~~~~~             ┌───────▼───────┐
		   Apply Flow               │Parse Flags    │
		 ~~~~~~~~~~~~~~             │               │
		                            │Read Config    │
		                            │               │
		                            │Read State-File│
		                            └───────┬───────┘
		                                    │                          ───┐
		                 ┌──────────────────▼───────────────────┐         │
		                 │Check if Terraform state is up to date│         │
		                 └──────────────────┬──┬────────────────┘         │
		                                    │  │Not up to date            │
		                                    │  │(Diff from Terraform plan)│
		                                    │  └────────────┐             │
		                                    │               │             │Terraform
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
		                                 │  │
		                      ┌──────────▼──▼──────────┐
		                      │Apply Attestation Config│
		                      └─────────────┬──────────┘
		                                    │
		                     ┌──────────────▼────────────┐
		                     │Extend API Server Cert SANs│
		                     └──────────────┬────────────┘
		                                    │                          ───┐
		                         ┌──────────▼────────┐                    │Helm
		                         │ Apply Helm Charts │                    │Phase
		                         └──────────┬────────┘                 ───┘
		                                    │                          ───┐
		                      ┌─────────────▼────────────┐                │
		Can   be skipped if we│Upgrade NodeVersion object│                │K8s/Image
	  ran Init RPC (time save)│  (Image and K8s update)  │                │Phase
		                      └─────────────┬────────────┘                │
							                │                          ───┘
							      ┌─────────▼──────────┐
							      │Write success output│
							      └────────────────────┘
*/
func (a *applyCmd) apply(cmd *cobra.Command, configFetcher attestationconfigapi.Fetcher, upgradeDir string) error {
	// Read user's config and state file
	a.log.Debugf("Reading config from %s", a.flags.pathPrefixer.PrefixPrintablePath(constants.ConfigFilename))
	conf, err := config.New(a.fileHandler, constants.ConfigFilename, configFetcher, a.flags.force)
	var configValidationErr *config.ValidationError
	if errors.As(err, &configValidationErr) {
		cmd.PrintErrln(configValidationErr.LongMessage())
	}
	if err != nil {
		return err
	}

	a.log.Debugf("Reading state file from %s", a.flags.pathPrefixer.PrefixPrintablePath(constants.StateFilename))
	stateFile, err := state.ReadFromFile(a.fileHandler, constants.StateFilename)
	if err != nil {
		return err
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
	// If not, we need to run the init RPC first
	a.log.Debugf("Checking if %s exists", a.flags.pathPrefixer.PrefixPrintablePath(constants.AdminConfFilename))
	initRequired := false
	if _, err := a.fileHandler.Stat(constants.AdminConfFilename); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("checking for %s: %w", a.flags.pathPrefixer.PrefixPrintablePath(constants.AdminConfFilename), err)
		}
		// Only run init RPC if we are not skipping the init phase
		// This may break things further down the line
		// It is the user's responsibility to make sure the cluster is in a valid state
		initRequired = true && !a.flags.skipPhases.contains(skipInitPhase)
	}
	a.log.Debugf("Init RPC required: %t", initRequired)

	// Validate input arguments

	// Validate Kubernetes version as set in the user's config
	// If we need to run the init RPC, the version has to be valid
	// Otherwise, we are able to use an outdated version, meaning we skip the K8s upgrade
	a.log.Debugf("Validating Kubernetes version %s", conf.KubernetesVersion)
	validVersion, err := versions.NewValidK8sVersion(string(conf.KubernetesVersion), true)
	if err != nil {
		a.log.Debugf("Kubernetes version not valid: %s", err)
		if initRequired {
			return err
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
				return fmt.Errorf("asking for confirmation: %w", err)
			}
			if !confirmed {
				return fmt.Errorf("aborted by user")
			}
		}
		a.flags.skipPhases = append(a.flags.skipPhases, skipK8sPhase)
		a.log.Debugf("Outdated Kubernetes version accepted, Kubernetes upgrade will be skipped")
	}
	if versions.IsPreviewK8sVersion(validVersion) {
		cmd.PrintErrf("Warning: Constellation with Kubernetes %s is still in preview. Use only for evaluation purposes.\n", validVersion)
	}
	conf.KubernetesVersion = validVersion
	a.log.Debugf("Target Kubernetes version set to %s", conf.KubernetesVersion)

	// Validate microservice version (helm versions) in the user's config matches the version of the CLI
	// This makes sure we catch potential errors early, not just after we already ran Terraform migrations or the init RPC
	if !a.flags.force {
		if err := validateCLIandConstellationVersionAreEqual(constants.BinaryVersion(), conf.Image, conf.MicroserviceVersion); err != nil {
			return err
		}
	}

	// Constellation on QEMU or OpenStack don't support upgrades
	// If using one of those providers, make sure the command is only used to initialize a cluster
	if !(conf.GetProvider() == cloudprovider.AWS || conf.GetProvider() == cloudprovider.Azure || conf.GetProvider() == cloudprovider.GCP) {
		if !initRequired {
			return fmt.Errorf("upgrades are not supported for provider %s", conf.GetProvider())
		}
		// Skip Terraform phase
		a.log.Debugf("Skipping Infrastructure phase for provider %s", conf.GetProvider())
		a.flags.skipPhases = append(a.flags.skipPhases, skipInfrastructurePhase)
	}

	// Print warning about AWS attestation
	// TODO(derpsteb): remove once AWS fixes SEV-SNP attestation provisioning issues
	if initRequired && conf.GetAttestationConfig().GetVariant().Equal(variant.AWSSEVSNP{}) {
		cmd.PrintErrln("WARNING: Attestation temporarily relies on AWS nitroTPM. See https://docs.edgeless.systems/constellation/workflows/config#choosing-a-vm-type for more information.")
	}

	// Now start actually running the apply command

	// Check if Terraform state is up to date and apply potential upgrades
	if !a.flags.skipPhases.contains(skipInfrastructurePhase) {
		if err := a.runTerraformApply(cmd, conf, stateFile, upgradeDir); err != nil {
			return err
		}
	}

	bufferedOutput := &bytes.Buffer{}
	// Run init RPC if required
	if initRequired {
		bufferedOutput, err = a.runInit(cmd, conf, stateFile)
		if err != nil {
			return err
		}
	}

	// From now on we can assume a valid Kubernetes admin config file exists
	kubeUpgrader, err := a.newKubeUpgrader(cmd.OutOrStdout(), constants.AdminConfFilename, a.log)
	if err != nil {
		return err
	}

	// Apply Attestation Config
	a.log.Debugf("Creating Kubernetes client using %s", a.flags.pathPrefixer.PrefixPrintablePath(constants.AdminConfFilename))
	a.log.Debugf("Applying new attestation config to cluster")
	if err := a.applyJoinConfig(cmd, kubeUpgrader, conf.GetAttestationConfig(), stateFile.ClusterValues.MeasurementSalt); err != nil {
		return fmt.Errorf("applying attestation config: %w", err)
	}

	// Extend API Server Cert SANs
	sans := append([]string{stateFile.Infrastructure.ClusterEndpoint, conf.CustomEndpoint}, stateFile.Infrastructure.APIServerCertSANs...)
	if err := kubeUpgrader.ExtendClusterConfigCertSANs(cmd.Context(), sans); err != nil {
		return fmt.Errorf("extending cert SANs: %w", err)
	}

	// Apply Helm Charts
	if !a.flags.skipPhases.contains(skipHelmPhase) {
		if err := a.runHelmApply(cmd, conf, stateFile, kubeUpgrader, upgradeDir, initRequired); err != nil {
			return err
		}
	}

	// Upgrade NodeVersion object
	// This can be skipped if we ran the init RPC, as the NodeVersion object is already up to date
	if !(a.flags.skipPhases.contains(skipK8sPhase) && a.flags.skipPhases.contains(skipImagePhase)) && !initRequired {
		if err := a.runK8sUpgrade(cmd, conf, kubeUpgrader); err != nil {
			return err
		}
	}

	// Write success output
	cmd.Print(bufferedOutput.String())

	return nil
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
		a.flags.skipPhases.contains(skipK8sPhase),
		a.flags.skipPhases.contains(skipImagePhase),
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
