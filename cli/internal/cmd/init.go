/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/api/attestationconfigapi"
	"github.com/edgelesssys/constellation/v2/internal/atls"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/clientcmd"
	clientcodec "k8s.io/client-go/tools/clientcmd/api/latest"
	"sigs.k8s.io/yaml"

	"github.com/edgelesssys/constellation/v2/bootstrapper/initproto"
	"github.com/edgelesssys/constellation/v2/cli/internal/cloudcmd"
	"github.com/edgelesssys/constellation/v2/cli/internal/clusterid"
	"github.com/edgelesssys/constellation/v2/cli/internal/cmd/pathprefix"
	"github.com/edgelesssys/constellation/v2/cli/internal/helm"
	"github.com/edgelesssys/constellation/v2/cli/internal/kubecmd"
	"github.com/edgelesssys/constellation/v2/cli/internal/state"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/crypto"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/grpc/dialer"
	"github.com/edgelesssys/constellation/v2/internal/grpc/grpclog"
	grpcRetry "github.com/edgelesssys/constellation/v2/internal/grpc/retry"
	"github.com/edgelesssys/constellation/v2/internal/kms/uri"
	"github.com/edgelesssys/constellation/v2/internal/license"
	"github.com/edgelesssys/constellation/v2/internal/retry"
	"github.com/edgelesssys/constellation/v2/internal/versions"
)

// NewInitCmd returns a new cobra.Command for the init command.
func NewInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize the Constellation cluster",
		Long: "Initialize the Constellation cluster.\n\n" +
			"Start your confidential Kubernetes.",
		Args: cobra.ExactArgs(0),
		RunE: runInitialize,
	}
	cmd.Flags().Bool("conformance", false, "enable conformance mode")
	cmd.Flags().Bool("skip-helm-wait", false, "install helm charts without waiting for deployments to be ready")
	cmd.Flags().Bool("merge-kubeconfig", false, "merge Constellation kubeconfig file with default kubeconfig file in $HOME/.kube/config")
	return cmd
}

type initCmd struct {
	log         debugLog
	merger      configMerger
	spinner     spinnerInterf
	fileHandler file.Handler
	pf          pathprefix.PathPrefixer
}

func newInitCmd(fileHandler file.Handler, spinner spinnerInterf, merger configMerger, log debugLog) *initCmd {
	return &initCmd{
		log:         log,
		merger:      merger,
		spinner:     spinner,
		fileHandler: fileHandler,
	}
}

// runInitialize runs the initialize command.
func runInitialize(cmd *cobra.Command, _ []string) error {
	log, err := newCLILogger(cmd)
	if err != nil {
		return fmt.Errorf("creating logger: %w", err)
	}
	defer log.Sync()
	fileHandler := file.NewHandler(afero.NewOsFs())
	newDialer := func(validator atls.Validator) *dialer.Dialer {
		return dialer.New(nil, validator, &net.Dialer{})
	}

	spinner, err := newSpinnerOrStderr(cmd)
	if err != nil {
		return err
	}
	defer spinner.Stop()

	ctx, cancel := context.WithTimeout(cmd.Context(), time.Hour)
	defer cancel()
	cmd.SetContext(ctx)

	i := newInitCmd(fileHandler, spinner, &kubeconfigMerger{log: log}, log)
	fetcher := attestationconfigapi.NewFetcher()
	newAttestationApplier := func(w io.Writer, kubeConfig string, log debugLog) (attestationConfigApplier, error) {
		return kubecmd.New(w, kubeConfig, fileHandler, log)
	}
	newHelmClient := func(kubeConfigPath string, log debugLog) (helmApplier, error) {
		return helm.NewClient(kubeConfigPath, log)
	} // need to defer helm client instantiation until kubeconfig is available

	return i.initialize(cmd, newDialer, license.NewClient(), fetcher, newAttestationApplier, newHelmClient)
}

// initialize initializes a Constellation.
func (i *initCmd) initialize(
	cmd *cobra.Command, newDialer func(validator atls.Validator) *dialer.Dialer,
	quotaChecker license.QuotaChecker, configFetcher attestationconfigapi.Fetcher,
	newAttestationApplier func(io.Writer, string, debugLog) (attestationConfigApplier, error),
	newHelmClient func(kubeConfigPath string, log debugLog) (helmApplier, error),
) error {
	flags, err := i.evalFlagArgs(cmd)
	if err != nil {
		return err
	}
	i.log.Debugf("Using flags: %+v", flags)
	i.log.Debugf("Loading configuration file from %q", i.pf.PrefixPrintablePath(constants.ConfigFilename))
	conf, err := config.New(i.fileHandler, constants.ConfigFilename, configFetcher, flags.force)
	var configValidationErr *config.ValidationError
	if errors.As(err, &configValidationErr) {
		cmd.PrintErrln(configValidationErr.LongMessage())
	}
	if err != nil {
		return err
	}
	// cfg validation does not check k8s patch version since upgrade may accept an outdated patch version.
	k8sVersion, err := versions.NewValidK8sVersion(string(conf.KubernetesVersion), true)
	if err != nil {
		return err
	}
	if !flags.force {
		if err := validateCLIandConstellationVersionAreEqual(constants.BinaryVersion(), conf.Image, conf.MicroserviceVersion); err != nil {
			return err
		}
	}
	if conf.GetAttestationConfig().GetVariant().Equal(variant.AWSSEVSNP{}) {
		cmd.PrintErrln("WARNING: Attestation temporarily relies on AWS nitroTPM. See https://docs.edgeless.systems/constellation/workflows/config#choosing-a-vm-type for more information.")
	}

	// TODO(msanft): Remove IDFile as per AB#3354
	i.log.Debugf("Checking cluster ID file")
	var idFile clusterid.File
	if err := i.fileHandler.ReadJSON(constants.ClusterIDsFilename, &idFile); err != nil {
		return fmt.Errorf("reading cluster ID file: %w", err)
	}

	stateFile, err := state.ReadFromFile(i.fileHandler, constants.StateFilename)
	if err != nil {
		return fmt.Errorf("reading state file: %w", err)
	}

	i.log.Debugf("Validated k8s version as %s", k8sVersion)
	if versions.IsPreviewK8sVersion(k8sVersion) {
		cmd.PrintErrf("Warning: Constellation with Kubernetes %v is still in preview. Use only for evaluation purposes.\n", k8sVersion)
	}

	provider := conf.GetProvider()
	i.log.Debugf("Got provider %s", provider.String())
	checker := license.NewChecker(quotaChecker, i.fileHandler)
	if err := checker.CheckLicense(cmd.Context(), provider, conf.Provider, cmd.Printf); err != nil {
		cmd.PrintErrf("License check failed: %v", err)
	}
	i.log.Debugf("Checked license")

	conf.UpdateMAAURL(idFile.AttestationURL)
	i.log.Debugf("Creating aTLS Validator for %s", conf.GetAttestationConfig().GetVariant())
	validator, err := cloudcmd.NewValidator(cmd, conf.GetAttestationConfig(), i.log)
	if err != nil {
		return fmt.Errorf("creating new validator: %w", err)
	}
	i.log.Debugf("Created a new validator")
	serviceAccURI, err := cloudcmd.GetMarshaledServiceAccountURI(provider, conf, i.pf, i.log, i.fileHandler)
	if err != nil {
		return err
	}
	i.log.Debugf("Successfully marshaled service account URI")

	i.log.Debugf("Generating master secret")
	masterSecret, err := i.generateMasterSecret(cmd.OutOrStdout())
	if err != nil {
		return fmt.Errorf("generating master secret: %w", err)
	}
	i.log.Debugf("Generated measurement salt")
	measurementSalt, err := crypto.GenerateRandomBytes(crypto.RNGLengthDefault)
	if err != nil {
		return fmt.Errorf("generating measurement salt: %w", err)
	}
	idFile.MeasurementSalt = measurementSalt

	stateFile.SetClusterValues(state.ClusterValues{
		MeasurementSalt: base64.StdEncoding.EncodeToString(measurementSalt),
	})

	clusterName := clusterid.GetClusterName(conf, idFile)
	i.log.Debugf("Setting cluster name to %s", clusterName)

	cmd.PrintErrln("Note: If you just created the cluster, it can take a few minutes to connect.")
	i.spinner.Start("Connecting ", false)
	req := &initproto.InitRequest{
		KmsUri:               masterSecret.EncodeToURI(),
		StorageUri:           uri.NoStoreURI,
		MeasurementSalt:      measurementSalt,
		KubernetesVersion:    versions.VersionConfigs[k8sVersion].ClusterVersion,
		KubernetesComponents: versions.VersionConfigs[k8sVersion].KubernetesComponents.ToInitProto(),
		ConformanceMode:      flags.conformance,
		InitSecret:           idFile.InitSecret,
		ClusterName:          clusterName,
		ApiserverCertSans:    idFile.APIServerCertSANs,
	}
	i.log.Debugf("Sending initialization request")
	resp, err := i.initCall(cmd.Context(), newDialer(validator), idFile.IP, req)
	i.spinner.Stop()

	if err != nil {
		var nonRetriable *nonRetriableError
		if errors.As(err, &nonRetriable) {
			cmd.PrintErrln("Cluster initialization failed. This error is not recoverable.")
			cmd.PrintErrln("Terminate your cluster and try again.")
			if nonRetriable.logCollectionErr != nil {
				cmd.PrintErrf("Failed to collect logs from bootstrapper: %s\n", nonRetriable.logCollectionErr)
			} else {
				cmd.PrintErrf("Fetched bootstrapper logs are stored in %q\n", i.pf.PrefixPrintablePath(constants.ErrorLog))
			}
		}
		return err
	}
	i.log.Debugf("Initialization request succeeded")

	// TODO(msanft): Remove IDFile as per AB#3354
	i.log.Debugf("Writing Constellation ID file")
	idFile.CloudProvider = provider

	bufferedOutput := &bytes.Buffer{}
	if err := i.writeOutput(idFile, stateFile, resp, flags.mergeConfigs, bufferedOutput); err != nil {
		return err
	}

	attestationApplier, err := newAttestationApplier(cmd.OutOrStdout(), constants.AdminConfFilename, i.log)
	if err != nil {
		return err
	}
	if err := attestationApplier.ApplyJoinConfig(cmd.Context(), conf.GetAttestationConfig(), measurementSalt); err != nil {
		return fmt.Errorf("applying attestation config: %w", err)
	}

	i.spinner.Start("Installing Kubernetes components ", false)
	options := helm.Options{
		Force:            flags.force,
		Conformance:      flags.conformance,
		HelmWaitMode:     flags.helmWaitMode,
		AllowDestructive: helm.DenyDestructive,
	}
	helmApplier, err := newHelmClient(constants.AdminConfFilename, i.log)
	if err != nil {
		return fmt.Errorf("creating Helm client: %w", err)
	}
	executor, includesUpgrades, err := helmApplier.PrepareApply(conf, stateFile, options, serviceAccURI, masterSecret)
	if err != nil {
		return fmt.Errorf("getting Helm chart executor: %w", err)
	}
	if includesUpgrades {
		return errors.New("init: helm tried to upgrade charts instead of installing them")
	}
	if err := executor.Apply(cmd.Context()); err != nil {
		return fmt.Errorf("applying Helm charts: %w", err)
	}
	i.spinner.Stop()
	i.log.Debugf("Helm deployment installation succeeded")
	cmd.Println(bufferedOutput.String())
	return nil
}

func (i *initCmd) initCall(ctx context.Context, dialer grpcDialer, ip string, req *initproto.InitRequest) (*initproto.InitSuccessResponse, error) {
	doer := &initDoer{
		dialer:   dialer,
		endpoint: net.JoinHostPort(ip, strconv.Itoa(constants.BootstrapperPort)),
		req:      req,
		log:      i.log,
		spinner:  i.spinner,
		fh:       file.NewHandler(afero.NewOsFs()),
	}

	// Create a wrapper function that allows logging any returned error from the retrier before checking if it's the expected retriable one.
	serviceIsUnavailable := func(err error) bool {
		isServiceUnavailable := grpcRetry.ServiceIsUnavailable(err)
		i.log.Debugf("Encountered error (retriable: %t): %s", isServiceUnavailable, err)
		return isServiceUnavailable
	}

	i.log.Debugf("Making initialization call, doer is %+v", doer)
	retrier := retry.NewIntervalRetrier(doer, 30*time.Second, serviceIsUnavailable)
	if err := retrier.Do(ctx); err != nil {
		return nil, err
	}
	return doer.resp, nil
}

type initDoer struct {
	dialer        grpcDialer
	endpoint      string
	req           *initproto.InitRequest
	resp          *initproto.InitSuccessResponse
	log           debugLog
	spinner       spinnerInterf
	connectedOnce bool
	fh            file.Handler
}

func (d *initDoer) Do(ctx context.Context) error {
	// connectedOnce is set in handleGRPCStateChanges when a connection was established in one retry attempt.
	// This should cancel any other retry attempts when the connection is lost since the bootstrapper likely won't accept any new attempts anymore.
	if d.connectedOnce {
		return &nonRetriableError{
			logCollectionErr: errors.New("init already connected to the remote server in a previous attempt - resumption is not supported"),
			err:              errors.New("init already connected to the remote server in a previous attempt - resumption is not supported"),
		}
	}

	conn, err := d.dialer.Dial(ctx, d.endpoint)
	if err != nil {
		d.log.Debugf("Dialing init server failed: %s. Retrying...", err)
		return fmt.Errorf("dialing init server: %w", err)
	}
	defer conn.Close()

	var wg sync.WaitGroup
	defer wg.Wait()

	grpcStateLogCtx, grpcStateLogCancel := context.WithCancel(ctx)
	defer grpcStateLogCancel()
	d.handleGRPCStateChanges(grpcStateLogCtx, &wg, conn)

	protoClient := initproto.NewAPIClient(conn)
	d.log.Debugf("Created protoClient")
	resp, err := protoClient.Init(ctx, d.req)
	if err != nil {
		return &nonRetriableError{
			logCollectionErr: errors.New("rpc failed before first response was received - no logs available"),
			err:              fmt.Errorf("init call: %w", err),
		}
	}

	res, err := resp.Recv() // get first response, either success or failure
	if err != nil {
		if e := d.getLogs(resp); e != nil {
			d.log.Debugf("Failed to collect logs: %s", e)
			return &nonRetriableError{
				logCollectionErr: e,
				err:              err,
			}
		}
		return &nonRetriableError{err: err}
	}

	switch res.Kind.(type) {
	case *initproto.InitResponse_InitFailure:
		if e := d.getLogs(resp); e != nil {
			d.log.Debugf("Failed to get logs from cluster: %s", e)
			return &nonRetriableError{
				logCollectionErr: e,
				err:              errors.New(res.GetInitFailure().GetError()),
			}
		}
		return &nonRetriableError{err: errors.New(res.GetInitFailure().GetError())}
	case *initproto.InitResponse_InitSuccess:
		d.resp = res.GetInitSuccess()
	case nil:
		d.log.Debugf("Cluster returned nil response type")
		err = errors.New("empty response from cluster")
		if e := d.getLogs(resp); e != nil {
			d.log.Debugf("Failed to collect logs: %s", e)
			return &nonRetriableError{
				logCollectionErr: e,
				err:              err,
			}
		}
		return &nonRetriableError{err: err}
	default:
		d.log.Debugf("Cluster returned unknown response type")
		err = errors.New("unknown response from cluster")
		if e := d.getLogs(resp); e != nil {
			d.log.Debugf("Failed to collect logs: %s", e)
			return &nonRetriableError{
				logCollectionErr: e,
				err:              err,
			}
		}
		return &nonRetriableError{err: err}
	}

	return nil
}

func (d *initDoer) getLogs(resp initproto.API_InitClient) error {
	d.log.Debugf("Attempting to collect cluster logs")
	for {
		res, err := resp.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		switch res.Kind.(type) {
		case *initproto.InitResponse_InitFailure:
			return errors.New("trying to collect logs: received init failure response, expected log response")
		case *initproto.InitResponse_InitSuccess:
			return errors.New("trying to collect logs: received init success response, expected log response")
		case nil:
			return errors.New("trying to collect logs: received nil response, expected log response")
		}

		log := res.GetLog().GetLog()
		if log == nil {
			return errors.New("received empty logs")
		}

		if err := d.fh.Write(constants.ErrorLog, log, file.OptAppend); err != nil {
			return err
		}
	}
	return nil
}

func (d *initDoer) handleGRPCStateChanges(ctx context.Context, wg *sync.WaitGroup, conn *grpc.ClientConn) {
	grpclog.LogStateChangesUntilReady(ctx, conn, d.log, wg, func() {
		d.connectedOnce = true
		d.spinner.Stop()
		d.spinner.Start("Initializing cluster ", false)
	})
}

func (i *initCmd) writeOutput(
	idFile clusterid.File, stateFile *state.State, initResp *initproto.InitSuccessResponse, mergeConfig bool, wr io.Writer,
) error {
	fmt.Fprint(wr, "Your Constellation cluster was successfully initialized.\n\n")

	ownerID := hex.EncodeToString(initResp.GetOwnerId())
	// i.log.Debugf("Owner id is %s", ownerID)
	clusterID := hex.EncodeToString(initResp.GetClusterId())

	tw := tabwriter.NewWriter(wr, 0, 0, 2, ' ', 0)
	// writeRow(tw, "Constellation cluster's owner identifier", ownerID)
	writeRow(tw, "Constellation cluster identifier", clusterID)
	writeRow(tw, "Kubernetes configuration", i.pf.PrefixPrintablePath(constants.AdminConfFilename))
	tw.Flush()
	fmt.Fprintln(wr)

	i.log.Debugf("Rewriting cluster server address in kubeconfig to %s", idFile.IP)
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
		kubeEndpoint.Host = net.JoinHostPort(idFile.IP, kubeEndpoint.Port())
		cluster.Server = kubeEndpoint.String()
	}
	kubeconfigBytes, err := clientcmd.Write(*kubeconfig)
	if err != nil {
		return fmt.Errorf("marshaling kubeconfig: %w", err)
	}

	if err := i.fileHandler.Write(constants.AdminConfFilename, kubeconfigBytes, file.OptNone); err != nil {
		return fmt.Errorf("writing kubeconfig: %w", err)
	}
	i.log.Debugf("Kubeconfig written to %s", i.pf.PrefixPrintablePath(constants.AdminConfFilename))

	if mergeConfig {
		if err := i.merger.mergeConfigs(constants.AdminConfFilename, i.fileHandler); err != nil {
			writeRow(tw, "Failed to automatically merge kubeconfig", err.Error())
			mergeConfig = false // Set to false so we don't print the wrong message below.
		} else {
			writeRow(tw, "Kubernetes configuration merged with default config", "")
		}
	}

	idFile.OwnerID = ownerID
	idFile.ClusterID = clusterID

	stateFile.ClusterValues.OwnerID = ownerID
	stateFile.ClusterValues.ClusterID = clusterID

	if err := stateFile.WriteToFile(i.fileHandler, constants.StateFilename); err != nil {
		return fmt.Errorf("writing state file: %w", err)
	}

	if err := i.fileHandler.WriteJSON(constants.ClusterIDsFilename, idFile, file.OptOverwrite); err != nil {
		return fmt.Errorf("writing Constellation ID file: %w", err)
	}
	i.log.Debugf("Constellation ID file written to %s", i.pf.PrefixPrintablePath(constants.ClusterIDsFilename))

	if !mergeConfig {
		fmt.Fprintln(wr, "You can now connect to your cluster by executing:")

		exportPath, err := filepath.Abs(constants.AdminConfFilename)
		if err != nil {
			return fmt.Errorf("getting absolute path to kubeconfig: %w", err)
		}

		fmt.Fprintf(wr, "\texport KUBECONFIG=%q\n", exportPath)
	} else {
		fmt.Fprintln(wr, "Constellation kubeconfig merged with default config.")

		if i.merger.kubeconfigEnvVar() != "" {
			fmt.Fprintln(wr, "Warning: KUBECONFIG environment variable is set.")
			fmt.Fprintln(wr, "You may need to unset it to use the default config and connect to your cluster.")
		} else {
			fmt.Fprintln(wr, "You can now connect to your cluster.")
		}
	}
	return nil
}

func writeRow(wr io.Writer, col1 string, col2 string) {
	fmt.Fprint(wr, col1, "\t", col2, "\n")
}

// evalFlagArgs gets the flag values and does preprocessing of these values like
// reading the content from file path flags and deriving other values from flag combinations.
func (i *initCmd) evalFlagArgs(cmd *cobra.Command) (initFlags, error) {
	conformance, err := cmd.Flags().GetBool("conformance")
	if err != nil {
		return initFlags{}, fmt.Errorf("parsing conformance flag: %w", err)
	}
	i.log.Debugf("Conformance flag is %t", conformance)
	skipHelmWait, err := cmd.Flags().GetBool("skip-helm-wait")
	if err != nil {
		return initFlags{}, fmt.Errorf("parsing skip-helm-wait flag: %w", err)
	}
	helmWaitMode := helm.WaitModeAtomic
	if skipHelmWait {
		helmWaitMode = helm.WaitModeNone
	}
	i.log.Debugf("Helm wait flag is %t", skipHelmWait)
	workDir, err := cmd.Flags().GetString("workspace")
	if err != nil {
		return initFlags{}, fmt.Errorf("parsing config path flag: %w", err)
	}
	i.pf = pathprefix.New(workDir)

	mergeConfigs, err := cmd.Flags().GetBool("merge-kubeconfig")
	if err != nil {
		return initFlags{}, fmt.Errorf("parsing merge-kubeconfig flag: %w", err)
	}
	i.log.Debugf("Merge kubeconfig flag is %t", mergeConfigs)

	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return initFlags{}, fmt.Errorf("parsing force argument: %w", err)
	}
	i.log.Debugf("force flag is %t", force)

	return initFlags{
		conformance:  conformance,
		helmWaitMode: helmWaitMode,
		force:        force,
		mergeConfigs: mergeConfigs,
	}, nil
}

// initFlags are the resulting values of flag preprocessing.
type initFlags struct {
	conformance  bool
	helmWaitMode helm.WaitMode
	force        bool
	mergeConfigs bool
}

// generateMasterSecret reads a base64 encoded master secret from file or generates a new 32 byte secret.
func (i *initCmd) generateMasterSecret(outWriter io.Writer) (uri.MasterSecret, error) {
	// No file given, generate a new secret, and save it to disk
	i.log.Debugf("Generating new master secret")
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
	i.log.Debugf("Generated master secret key and salt values")
	if err := i.fileHandler.WriteJSON(constants.MasterSecretFilename, secret, file.OptNone); err != nil {
		return uri.MasterSecret{}, err
	}
	fmt.Fprintf(outWriter, "Your Constellation master secret was successfully written to %q\n", i.pf.PrefixPrintablePath(constants.MasterSecretFilename))
	return secret, nil
}

type configMerger interface {
	mergeConfigs(configPath string, fileHandler file.Handler) error
	kubeconfigEnvVar() string
}

type kubeconfigMerger struct {
	log debugLog
}

func (c *kubeconfigMerger) mergeConfigs(configPath string, fileHandler file.Handler) error {
	constellConfig, err := clientcmd.LoadFromFile(configPath)
	if err != nil {
		return fmt.Errorf("loading admin kubeconfig: %w", err)
	}

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.Precedence = []string{
		clientcmd.RecommendedHomeFile,
		configPath, // our config should overwrite the default config
	}
	c.log.Debugf("Kubeconfig file loading precedence: %v", loadingRules.Precedence)

	// merge the kubeconfigs
	cfg, err := loadingRules.Load()
	if err != nil {
		return fmt.Errorf("loading merged kubeconfig: %w", err)
	}

	// Set the current context to the cluster we just created
	cfg.CurrentContext = constellConfig.CurrentContext
	c.log.Debugf("Set current context to %s", cfg.CurrentContext)

	json, err := runtime.Encode(clientcodec.Codec, cfg)
	if err != nil {
		return fmt.Errorf("encoding merged kubeconfig: %w", err)
	}

	mergedKubeconfig, err := yaml.JSONToYAML(json)
	if err != nil {
		return fmt.Errorf("converting merged kubeconfig to YAML: %w", err)
	}

	if err := fileHandler.Write(clientcmd.RecommendedHomeFile, mergedKubeconfig, file.OptOverwrite); err != nil {
		return fmt.Errorf("writing merged kubeconfig to file: %w", err)
	}
	c.log.Debugf("Merged kubeconfig into default config file: %s", clientcmd.RecommendedHomeFile)
	return nil
}

func (c *kubeconfigMerger) kubeconfigEnvVar() string {
	return os.Getenv(clientcmd.RecommendedConfigPathEnvVar)
}

type grpcDialer interface {
	Dial(ctx context.Context, target string) (*grpc.ClientConn, error)
}

type nonRetriableError struct {
	logCollectionErr error
	err              error
}

// Error returns the error message.
func (e *nonRetriableError) Error() string {
	return e.err.Error()
}

// Unwrap returns the wrapped error.
func (e *nonRetriableError) Unwrap() error {
	return e.err
}

type attestationConfigApplier interface {
	ApplyJoinConfig(ctx context.Context, newAttestConfig config.AttestationCfg, measurementSalt []byte) error
}

type helmApplier interface {
	PrepareApply(conf *config.Config, stateFile *state.State,
		flags helm.Options, serviceAccURI string, masterSecret uri.MasterSecret) (
		helm.Applier, bool, error)
}
