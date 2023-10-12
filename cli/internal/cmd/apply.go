/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"text/tabwriter"
	"time"

	"github.com/edgelesssys/constellation/v2/bootstrapper/initproto"
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
	"github.com/edgelesssys/constellation/v2/internal/crypto"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/grpc/dialer"
	grpcRetry "github.com/edgelesssys/constellation/v2/internal/grpc/retry"
	"github.com/edgelesssys/constellation/v2/internal/kms/uri"
	"github.com/edgelesssys/constellation/v2/internal/license"
	"github.com/edgelesssys/constellation/v2/internal/retry"
	"github.com/edgelesssys/constellation/v2/internal/versions"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/clientcmd"
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

// runTerraformApply checks if changes to Terraform are required and applies them.
func (a *applyCmd) runTerraformApply(cmd *cobra.Command, conf *config.Config, stateFile *state.State, upgradeDir string) error {
	a.log.Debugf("Checking if Terraform migrations are required")
	migrationRequired, err := a.planTerraformMigration(cmd, conf)
	if err != nil {
		return fmt.Errorf("planning Terraform migrations: %w", err)
	}

	if !migrationRequired {
		a.log.Debugf("No changes to infrastructure required, skipping Terraform migrations")
		return nil
	}

	a.log.Debugf("Migrating terraform resources for infrastructure changes")
	postMigrationInfraState, err := a.migrateTerraform(cmd, conf, upgradeDir)
	if err != nil {
		return fmt.Errorf("performing Terraform migrations: %w", err)
	}

	// Merge the pre-upgrade state with the post-migration infrastructure values
	a.log.Debugf("Updating state file with new infrastructure state")
	if _, err := stateFile.Merge(
		// temporary state with post-migration infrastructure values
		state.New().SetInfrastructure(postMigrationInfraState),
	); err != nil {
		return fmt.Errorf("merging pre-upgrade state with post-migration infrastructure values: %w", err)
	}

	// Write the post-migration state to disk
	if err := stateFile.WriteToFile(a.fileHandler, constants.StateFilename); err != nil {
		return fmt.Errorf("writing state file: %w", err)
	}
	return nil
}

// planTerraformMigration checks if the Constellation version the cluster is being upgraded to requires a migration.
func (a *applyCmd) planTerraformMigration(cmd *cobra.Command, conf *config.Config) (bool, error) {
	a.log.Debugf("Planning Terraform migrations")
	vars, err := cloudcmd.TerraformUpgradeVars(conf)
	if err != nil {
		return false, fmt.Errorf("parsing upgrade variables: %w", err)
	}
	a.log.Debugf("Using Terraform variables:\n%+v", vars)

	// Check if there are any Terraform migrations to apply

	// Add manual migrations here if required
	//
	// var manualMigrations []terraform.StateMigration
	// for _, migration := range manualMigrations {
	// 	  u.log.Debugf("Adding manual Terraform migration: %s", migration.DisplayName)
	// 	  u.upgrader.AddManualStateMigration(migration)
	// }

	return a.clusterUpgrader.PlanClusterUpgrade(cmd.Context(), cmd.OutOrStdout(), vars, conf.GetProvider())
}

// migrateTerraform migrates an existing Terraform state and the post-migration infrastructure state is returned.
func (a *applyCmd) migrateTerraform(cmd *cobra.Command, conf *config.Config, upgradeDir string) (state.Infrastructure, error) {
	// Ask for confirmation first
	fmt.Fprintln(cmd.OutOrStdout(), "The upgrade requires a migration of Constellation cloud resources by applying an updated Terraform template. Please manually review the suggested changes below.")
	if !a.flags.yes {
		ok, err := askToConfirm(cmd, "Do you want to apply the Terraform migrations?")
		if err != nil {
			return state.Infrastructure{}, fmt.Errorf("asking for confirmation: %w", err)
		}
		if !ok {
			cmd.Println("Aborting upgrade.")
			// User doesn't expect to see any changes in his workspace after aborting an "upgrade apply",
			// therefore, roll back to the backed up state.
			if err := a.clusterUpgrader.RestoreClusterWorkspace(); err != nil {
				return state.Infrastructure{}, fmt.Errorf(
					"restoring Terraform workspace: %w, restore the Terraform workspace manually from %s ",
					err,
					filepath.Join(upgradeDir, constants.TerraformUpgradeBackupDir),
				)
			}
			return state.Infrastructure{}, fmt.Errorf("cluster upgrade aborted by user")
		}
	}
	a.log.Debugf("Applying Terraform migrations")

	a.spinner.Start("Migrating Terraform resources", false)
	infraState, err := a.clusterUpgrader.ApplyClusterUpgrade(cmd.Context(), conf.GetProvider())
	a.spinner.Stop()
	if err != nil {
		return state.Infrastructure{}, fmt.Errorf("applying terraform migrations: %w", err)
	}

	cmd.Printf("Infrastructure migrations applied successfully and output written to: %s\n"+
		"A backup of the pre-upgrade state has been written to: %s\n",
		a.flags.pathPrefixer.PrefixPrintablePath(constants.StateFilename),
		a.flags.pathPrefixer.PrefixPrintablePath(filepath.Join(upgradeDir, constants.TerraformUpgradeBackupDir)),
	)
	return infraState, nil
}

// runInit runs the init RPC to set up the Kubernetes cluster.
// This function only needs to be run once per cluster.
// On success, it writes the Kubernetes admin config file to disk.
// Therefore it is skipped if the Kubernetes admin config file already exists.
func (a *applyCmd) runInit(cmd *cobra.Command, conf *config.Config, stateFile *state.State) (*bytes.Buffer, error) {
	a.log.Debugf("Running init RPC")
	a.log.Debugf("Creating aTLS Validator for %s", conf.GetAttestationConfig().GetVariant())
	validator, err := cloudcmd.NewValidator(cmd, conf.GetAttestationConfig(), a.log)
	if err != nil {
		return nil, fmt.Errorf("creating new validator: %w", err)
	}

	a.log.Debugf("Generating master secret")
	masterSecret, err := a.generateMasterSecret(cmd.OutOrStdout())
	if err != nil {
		return nil, fmt.Errorf("generating master secret: %w", err)
	}
	a.log.Debugf("Generated master secret key and salt values")

	a.log.Debugf("Generating measurement salt")
	measurementSalt, err := crypto.GenerateRandomBytes(crypto.RNGLengthDefault)
	if err != nil {
		return nil, fmt.Errorf("generating measurement salt: %w", err)
	}

	a.spinner.Start("Connecting ", false)
	req := &initproto.InitRequest{
		KmsUri:               masterSecret.EncodeToURI(),
		StorageUri:           uri.NoStoreURI,
		MeasurementSalt:      measurementSalt,
		KubernetesVersion:    versions.VersionConfigs[conf.KubernetesVersion].ClusterVersion,
		KubernetesComponents: versions.VersionConfigs[conf.KubernetesVersion].KubernetesComponents.ToInitProto(),
		ConformanceMode:      a.flags.conformance,
		InitSecret:           stateFile.Infrastructure.InitSecret,
		ClusterName:          stateFile.Infrastructure.Name,
		ApiserverCertSans:    stateFile.Infrastructure.APIServerCertSANs,
	}
	a.log.Debugf("Sending initialization request")
	resp, err := a.initCall(cmd.Context(), a.newDialer(validator), stateFile.Infrastructure.ClusterEndpoint, req)
	a.spinner.Stop()
	a.log.Debugf("Initialization request finished")

	if err != nil {
		var nonRetriable *nonRetriableError
		if errors.As(err, &nonRetriable) {
			cmd.PrintErrln("Cluster initialization failed. This error is not recoverable.")
			cmd.PrintErrln("Terminate your cluster and try again.")
			if nonRetriable.logCollectionErr != nil {
				cmd.PrintErrf("Failed to collect logs from bootstrapper: %s\n", nonRetriable.logCollectionErr)
			} else {
				cmd.PrintErrf("Fetched bootstrapper logs are stored in %q\n", a.flags.pathPrefixer.PrefixPrintablePath(constants.ErrorLog))
			}
		}
		return nil, err
	}
	a.log.Debugf("Initialization request successful")

	a.log.Debugf("Buffering init success message")
	bufferedOutput := &bytes.Buffer{}
	if err := a.writeOutput(stateFile, resp, a.flags.mergeConfigs, bufferedOutput, measurementSalt); err != nil {
		return nil, err
	}

	return bufferedOutput, nil
}

// initCall performs the gRPC call to the bootstrapper to initialize the cluster.
func (a *applyCmd) initCall(ctx context.Context, dialer grpcDialer, ip string, req *initproto.InitRequest) (*initproto.InitSuccessResponse, error) {
	doer := &initDoer{
		dialer:   dialer,
		endpoint: net.JoinHostPort(ip, strconv.Itoa(constants.BootstrapperPort)),
		req:      req,
		log:      a.log,
		spinner:  a.spinner,
		fh:       file.NewHandler(afero.NewOsFs()),
	}

	// Create a wrapper function that allows logging any returned error from the retrier before checking if it's the expected retriable one.
	serviceIsUnavailable := func(err error) bool {
		isServiceUnavailable := grpcRetry.ServiceIsUnavailable(err)
		a.log.Debugf("Encountered error (retriable: %t): %s", isServiceUnavailable, err)
		return isServiceUnavailable
	}

	a.log.Debugf("Making initialization call, doer is %+v", doer)
	retrier := retry.NewIntervalRetrier(doer, 30*time.Second, serviceIsUnavailable)
	if err := retrier.Do(ctx); err != nil {
		return nil, err
	}
	return doer.resp, nil
}

// generateMasterSecret reads a base64 encoded master secret from file or generates a new 32 byte secret.
func (a *applyCmd) generateMasterSecret(outWriter io.Writer) (uri.MasterSecret, error) {
	// No file given, generate a new secret, and save it to disk
	key, err := crypto.GenerateRandomBytes(crypto.MasterSecretLengthDefault)
	if err != nil {
		return uri.MasterSecret{}, err
	}
	salt, err := crypto.GenerateRandomBytes(crypto.RNGLengthDefault)
	if err != nil {
		return uri.MasterSecret{}, err
	}
	secret := uri.MasterSecret{
		Key:  key,
		Salt: salt,
	}
	if err := a.fileHandler.WriteJSON(constants.MasterSecretFilename, secret, file.OptNone); err != nil {
		return uri.MasterSecret{}, err
	}
	fmt.Fprintf(outWriter, "Your Constellation master secret was successfully written to %q\n", a.flags.pathPrefixer.PrefixPrintablePath(constants.MasterSecretFilename))
	return secret, nil
}

// writeOutput writes the output of a cluster initialization to the
// state- / kubeconfig-file and saves it to disk.
func (a *applyCmd) writeOutput(
	stateFile *state.State, initResp *initproto.InitSuccessResponse,
	mergeConfig bool, wr io.Writer, measurementSalt []byte,
) error {
	fmt.Fprint(wr, "Your Constellation cluster was successfully initialized.\n\n")

	ownerID := hex.EncodeToString(initResp.GetOwnerId())
	clusterID := hex.EncodeToString(initResp.GetClusterId())

	stateFile.SetClusterValues(state.ClusterValues{
		MeasurementSalt: measurementSalt,
		OwnerID:         ownerID,
		ClusterID:       clusterID,
	})

	tw := tabwriter.NewWriter(wr, 0, 0, 2, ' ', 0)
	writeRow(tw, "Constellation cluster identifier", clusterID)
	writeRow(tw, "Kubernetes configuration", a.flags.pathPrefixer.PrefixPrintablePath(constants.AdminConfFilename))
	tw.Flush()
	fmt.Fprintln(wr)

	a.log.Debugf("Rewriting cluster server address in kubeconfig to %s", stateFile.Infrastructure.ClusterEndpoint)
	kubeconfig, err := clientcmd.Load(initResp.GetKubeconfig())
	if err != nil {
		return fmt.Errorf("loading kubeconfig: %w", err)
	}
	if len(kubeconfig.Clusters) != 1 {
		return fmt.Errorf("expected exactly one cluster in kubeconfig, got %d", len(kubeconfig.Clusters))
	}
	for _, cluster := range kubeconfig.Clusters {
		kubeEndpoint, err := url.Parse(cluster.Server)
		if err != nil {
			return fmt.Errorf("parsing kubeconfig server URL: %w", err)
		}
		kubeEndpoint.Host = net.JoinHostPort(stateFile.Infrastructure.ClusterEndpoint, kubeEndpoint.Port())
		cluster.Server = kubeEndpoint.String()
	}
	kubeconfigBytes, err := clientcmd.Write(*kubeconfig)
	if err != nil {
		return fmt.Errorf("marshaling kubeconfig: %w", err)
	}

	if err := a.fileHandler.Write(constants.AdminConfFilename, kubeconfigBytes, file.OptNone); err != nil {
		return fmt.Errorf("writing kubeconfig: %w", err)
	}
	a.log.Debugf("Kubeconfig written to %s", a.flags.pathPrefixer.PrefixPrintablePath(constants.AdminConfFilename))

	if mergeConfig {
		if err := a.merger.mergeConfigs(constants.AdminConfFilename, a.fileHandler); err != nil {
			writeRow(tw, "Failed to automatically merge kubeconfig", err.Error())
			mergeConfig = false // Set to false so we don't print the wrong message below.
		} else {
			writeRow(tw, "Kubernetes configuration merged with default config", "")
		}
	}

	if err := stateFile.WriteToFile(a.fileHandler, constants.StateFilename); err != nil {
		return fmt.Errorf("writing Constellation state file: %w", err)
	}

	a.log.Debugf("Constellation state file written to %s", a.flags.pathPrefixer.PrefixPrintablePath(constants.StateFilename))

	if !mergeConfig {
		fmt.Fprintln(wr, "You can now connect to your cluster by executing:")

		exportPath, err := filepath.Abs(constants.AdminConfFilename)
		if err != nil {
			return fmt.Errorf("getting absolute path to kubeconfig: %w", err)
		}

		fmt.Fprintf(wr, "\texport KUBECONFIG=%q\n", exportPath)
	} else {
		fmt.Fprintln(wr, "Constellation kubeconfig merged with default config.")

		if a.merger.kubeconfigEnvVar() != "" {
			fmt.Fprintln(wr, "Warning: KUBECONFIG environment variable is set.")
			fmt.Fprintln(wr, "You may need to unset it to use the default config and connect to your cluster.")
		} else {
			fmt.Fprintln(wr, "You can now connect to your cluster.")
		}
	}
	fmt.Fprintln(wr) // add final newline
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

// runHelmApply handles installing or upgrading helm charts for the cluster.
func (a *applyCmd) runHelmApply(
	cmd *cobra.Command, conf *config.Config, stateFile *state.State,
	kubeUpgrader kubernetesUpgrader, upgradeDir string, initRequired bool,
) error {
	a.log.Debugf("Installing or upgrading Helm charts")
	var masterSecret uri.MasterSecret
	if err := a.fileHandler.ReadJSON(constants.MasterSecretFilename, &masterSecret); err != nil {
		return fmt.Errorf("reading master secret: %w", err)
	}

	options := helm.Options{
		Force:            a.flags.force,
		Conformance:      a.flags.conformance,
		HelmWaitMode:     a.flags.helmWaitMode,
		AllowDestructive: helm.DenyDestructive,
	}
	helmApplier, err := a.newHelmClient(constants.AdminConfFilename, a.log)
	if err != nil {
		return fmt.Errorf("creating Helm client: %w", err)
	}

	a.log.Debugf("Getting service account URI")
	serviceAccURI, err := cloudcmd.GetMarshaledServiceAccountURI(conf, a.fileHandler)
	if err != nil {
		return err
	}

	a.log.Debugf("Preparing Helm charts")
	executor, includesUpgrades, err := helmApplier.PrepareApply(conf, stateFile, options, serviceAccURI, masterSecret)
	if errors.Is(err, helm.ErrConfirmationMissing) {
		if !a.flags.yes {
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
		options.AllowDestructive = helm.AllowDestructive
		executor, includesUpgrades, err = helmApplier.PrepareApply(conf, stateFile, options, serviceAccURI, masterSecret)
	}
	var upgradeErr *compatibility.InvalidUpgradeError
	if err != nil {
		if !errors.As(err, &upgradeErr) {
			return fmt.Errorf("preparing Helm charts: %w", err)
		}
		cmd.PrintErrln(err)
	}

	a.log.Debugf("Backing up Helm charts")
	if err := a.backupHelmCharts(cmd.Context(), kubeUpgrader, executor, includesUpgrades, upgradeDir); err != nil {
		return err
	}

	a.log.Debugf("Applying Helm charts")
	if initRequired {
		a.spinner.Start("Installing Kubernetes components ", false)
	} else {
		a.spinner.Start("Upgrading Kubernetes components ", false)
	}
	if err := executor.Apply(cmd.Context()); err != nil {
		return fmt.Errorf("applying Helm charts: %w", err)
	}
	a.spinner.Stop()

	if !initRequired {
		cmd.Println("Successfully upgraded Constellation services.")
	}

	return nil
}

// backupHelmCharts saves the Helm charts for the upgrade to disk and creates a backup of existing CRDs and CRs.
func (a *applyCmd) backupHelmCharts(
	ctx context.Context, kubeUpgrader kubernetesUpgrader, executor helm.Applier, includesUpgrades bool, upgradeDir string,
) error {
	// Save the Helm charts for the upgrade to disk
	chartDir := filepath.Join(upgradeDir, "helm-charts")
	if err := executor.SaveCharts(chartDir, a.fileHandler); err != nil {
		return fmt.Errorf("saving Helm charts to disk: %w", err)
	}
	a.log.Debugf("Helm charts saved to %s", a.flags.pathPrefixer.PrefixPrintablePath(chartDir))

	if includesUpgrades {
		a.log.Debugf("Creating backup of CRDs and CRs")
		crds, err := kubeUpgrader.BackupCRDs(ctx, upgradeDir)
		if err != nil {
			return fmt.Errorf("creating CRD backup: %w", err)
		}
		if err := kubeUpgrader.BackupCRs(ctx, crds, upgradeDir); err != nil {
			return fmt.Errorf("creating CR backup: %w", err)
		}
	}

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
