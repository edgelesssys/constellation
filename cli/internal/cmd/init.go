/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/compatibility"

	"github.com/edgelesssys/constellation/v2/bootstrapper/initproto"
	"github.com/edgelesssys/constellation/v2/cli/internal/cloudcmd"
	"github.com/edgelesssys/constellation/v2/cli/internal/clusterid"
	"github.com/edgelesssys/constellation/v2/cli/internal/helm"
	"github.com/edgelesssys/constellation/v2/internal/cloud/azureshared"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/cloud/gcpshared"
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
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/clientcmd"
	clientcodec "k8s.io/client-go/tools/clientcmd/api/latest"
	"sigs.k8s.io/yaml"
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
	cmd.Flags().String("master-secret", "", "path to base64-encoded master secret")
	cmd.Flags().Bool("conformance", false, "enable conformance mode")
	cmd.Flags().Bool("merge-kubeconfig", false, "merge Constellation kubeconfig file with default kubeconfig file in $HOME/.kube/config")
	return cmd
}

type initCmd struct {
	log     debugLog
	merger  configMerger
	spinner spinnerInterf
}

// runInitialize runs the initialize command.
func runInitialize(cmd *cobra.Command, _ []string) error {
	log, err := newCLILogger(cmd)
	if err != nil {
		return fmt.Errorf("creating logger: %w", err)
	}
	defer log.Sync()
	fileHandler := file.NewHandler(afero.NewOsFs())
	newDialer := func(validator *cloudcmd.Validator) *dialer.Dialer {
		return dialer.New(nil, validator.V(cmd), &net.Dialer{})
	}

	spinner, err := newSpinnerOrStderr(cmd)
	if err != nil {
		return err
	}
	defer spinner.Stop()

	ctx, cancel := context.WithTimeout(cmd.Context(), time.Hour)
	defer cancel()
	cmd.SetContext(ctx)
	i := &initCmd{log: log, spinner: spinner, merger: &kubeconfigMerger{log: log}}
	return i.initialize(cmd, newDialer, fileHandler, license.NewClient())
}

// initialize initializes a Constellation.
func (i *initCmd) initialize(cmd *cobra.Command, newDialer func(validator *cloudcmd.Validator) *dialer.Dialer,
	fileHandler file.Handler, quotaChecker license.QuotaChecker,
) error {
	flags, err := i.evalFlagArgs(cmd)
	if err != nil {
		return err
	}
	i.log.Debugf("Using flags: %+v", flags)
	i.log.Debugf("Loading configuration file from %q", flags.configPath)
	conf, err := config.New(fileHandler, flags.configPath, flags.force)
	var configValidationErr *config.ValidationError
	if errors.As(err, &configValidationErr) {
		cmd.PrintErrln(configValidationErr.LongMessage())
	}
	if err != nil {
		return err
	}

	i.log.Debugf("Checking cluster ID file")
	var idFile clusterid.File
	if err := fileHandler.ReadJSON(constants.ClusterIDsFileName, &idFile); err != nil {
		return fmt.Errorf("reading cluster ID file: %w", err)
	}

	k8sVersion, err := versions.NewValidK8sVersion(compatibility.EnsurePrefixV(conf.KubernetesVersion))
	if err != nil {
		return fmt.Errorf("validating kubernetes version: %w", err)
	}
	i.log.Debugf("Validated k8s version as %s", k8sVersion)
	if versions.IsPreviewK8sVersion(k8sVersion) {
		cmd.PrintErrf("Warning: Constellation with Kubernetes %v is still in preview. Use only for evaluation purposes.\n", k8sVersion)
	}

	provider := conf.GetProvider()
	i.log.Debugf("Got provider %s", provider.String())
	checker := license.NewChecker(quotaChecker, fileHandler)
	if err := checker.CheckLicense(cmd.Context(), provider, conf.Provider, cmd.Printf); err != nil {
		cmd.PrintErrf("License check failed: %v", err)
	}
	i.log.Debugf("Checked license")
	validator, err := cloudcmd.NewValidator(conf, i.log)
	if err != nil {
		return err
	}
	i.log.Debugf("Created a new validator")
	serviceAccURI, err := i.getMarshaledServiceAccountURI(provider, conf, fileHandler)
	if err != nil {
		return err
	}
	i.log.Debugf("Successfully marshaled service account URI")
	masterSecret, err := i.readOrGenerateMasterSecret(cmd.OutOrStdout(), fileHandler, flags.masterSecretPath)
	if err != nil {
		return fmt.Errorf("parsing or generating master secret from file %s: %w", flags.masterSecretPath, err)
	}
	helmLoader := helm.NewLoader(provider, k8sVersion)
	i.log.Debugf("Created new Helm loader")
	helmDeployments, err := helmLoader.Load(conf, flags.conformance, masterSecret.Key, masterSecret.Salt)
	i.log.Debugf("Loaded Helm deployments")
	if err != nil {
		return fmt.Errorf("loading Helm charts: %w", err)
	}

	clusterName := conf.Name + "-" + idFile.UID
	i.log.Debugf("Setting cluster name to %s", clusterName)

	cmd.PrintErrln("Note: If you just created the cluster, it can take a few minutes to connect.")
	i.spinner.Start("Connecting ", false)
	req := &initproto.InitRequest{
		KmsUri:                 masterSecret.EncodeToURI(),
		StorageUri:             uri.NoStoreURI,
		CloudServiceAccountUri: serviceAccURI,
		KubernetesVersion:      versions.VersionConfigs[k8sVersion].ClusterVersion,
		KubernetesComponents:   versions.VersionConfigs[k8sVersion].KubernetesComponents.ToInitProto(),
		HelmDeployments:        helmDeployments,
		EnforcedPcrs:           conf.EnforcedPCRs(),
		EnforceIdkeydigest:     conf.EnforcesIDKeyDigest(),
		ConformanceMode:        flags.conformance,
		InitSecret:             idFile.InitSecret,
		ClusterName:            clusterName,
	}
	i.log.Debugf("Sending initialization request")
	resp, err := i.initCall(cmd.Context(), newDialer(validator), idFile.IP, req)
	i.spinner.Stop()
	if err != nil {
		var nonRetriable *nonRetriableError
		if errors.As(err, &nonRetriable) {
			cmd.PrintErrln("Cluster initialization failed. This error is not recoverable.")
			cmd.PrintErrln("Terminate your cluster and try again.")
		}
		return err
	}
	i.log.Debugf("Initialization request succeeded")
	i.log.Debugf("Writing Constellation ID file")
	idFile.CloudProvider = provider

	return i.writeOutput(idFile, resp, flags.mergeConfigs, cmd.OutOrStdout(), fileHandler)
}

func (i *initCmd) initCall(ctx context.Context, dialer grpcDialer, ip string, req *initproto.InitRequest) (*initproto.InitResponse, error) {
	doer := &initDoer{
		dialer:   dialer,
		endpoint: net.JoinHostPort(ip, strconv.Itoa(constants.BootstrapperPort)),
		req:      req,
		log:      i.log,
		spinner:  i.spinner,
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
	resp          *initproto.InitResponse
	log           debugLog
	spinner       spinnerInterf
	connectedOnce bool
}

func (d *initDoer) Do(ctx context.Context) error {
	// connectedOnce is set in handleGRPCStateChanges when a connection was established in one retry attempt.
	// This should cancel any other retry attempts when the connection is lost since the bootstrapper likely won't accept any new attempts anymore.
	if d.connectedOnce {
		return &nonRetriableError{errors.New("init already connected to the remote server in a previous attempt - resumption is not supported")}
	}

	conn, err := d.dialer.Dial(ctx, d.endpoint)
	if err != nil {
		d.log.Debugf("Dialing init server failed: %w. Retrying...", err)
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
		return &nonRetriableError{fmt.Errorf("init call: %w", err)}
	}
	d.resp = resp
	return nil
}

func (d *initDoer) handleGRPCStateChanges(ctx context.Context, wg *sync.WaitGroup, conn *grpc.ClientConn) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		state := conn.GetState()
		d.log.Debugf("Connection state started as %s", state)
		for ; state != connectivity.Ready && conn.WaitForStateChange(ctx, state); state = conn.GetState() {
			d.log.Debugf("Connection state changed to %s", state)
		}
		if state == connectivity.Ready {
			d.log.Debugf("Connection ready")
			d.connectedOnce = true
			d.spinner.Stop()
			d.spinner.Start("Initializing cluster ", false)
		} else {
			d.log.Debugf("Connection state ended with %s", state)
		}
	}()
}

func (i *initCmd) writeOutput(
	idFile clusterid.File, resp *initproto.InitResponse, mergeConfig bool, wr io.Writer, fileHandler file.Handler,
) error {
	fmt.Fprint(wr, "Your Constellation cluster was successfully initialized.\n\n")

	ownerID := hex.EncodeToString(resp.OwnerId)
	// i.log.Debugf("Owner id is %s", ownerID)
	clusterID := hex.EncodeToString(resp.ClusterId)

	tw := tabwriter.NewWriter(wr, 0, 0, 2, ' ', 0)
	// writeRow(tw, "Constellation cluster's owner identifier", ownerID)
	writeRow(tw, "Constellation cluster identifier", clusterID)
	writeRow(tw, "Kubernetes configuration", constants.AdminConfFilename)
	tw.Flush()
	fmt.Fprintln(wr)

	if err := fileHandler.Write(constants.AdminConfFilename, resp.Kubeconfig, file.OptNone); err != nil {
		return fmt.Errorf("writing kubeconfig: %w", err)
	}
	i.log.Debugf("Kubeconfig written to %s", constants.AdminConfFilename)

	if mergeConfig {
		if err := i.merger.mergeConfigs(constants.AdminConfFilename, fileHandler); err != nil {
			writeRow(tw, "Failed to automatically merge kubeconfig", err.Error())
			mergeConfig = false // Set to false so we don't print the wrong message below.
		} else {
			writeRow(tw, "Kubernetes configuration merged with default config", "")
		}
	}

	idFile.OwnerID = ownerID
	idFile.ClusterID = clusterID

	if err := fileHandler.WriteJSON(constants.ClusterIDsFileName, idFile, file.OptOverwrite); err != nil {
		return fmt.Errorf("writing Constellation ID file: %w", err)
	}
	i.log.Debugf("Constellation ID file written to %s", constants.ClusterIDsFileName)

	if !mergeConfig {
		fmt.Fprintln(wr, "You can now connect to your cluster by executing:")
		fmt.Fprintf(wr, "\texport KUBECONFIG=\"$PWD/%s\"\n", constants.AdminConfFilename)
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
	masterSecretPath, err := cmd.Flags().GetString("master-secret")
	if err != nil {
		return initFlags{}, fmt.Errorf("parsing master-secret path flag: %w", err)
	}
	i.log.Debugf("Master secret path flag value is %q", masterSecretPath)
	conformance, err := cmd.Flags().GetBool("conformance")
	if err != nil {
		return initFlags{}, fmt.Errorf("parsing conformance flag: %w", err)
	}
	i.log.Debugf("Conformance flag is %t", conformance)
	configPath, err := cmd.Flags().GetString("config")
	if err != nil {
		return initFlags{}, fmt.Errorf("parsing config path flag: %w", err)
	}
	i.log.Debugf("Configuration path flag is %q", configPath)
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
		configPath:       configPath,
		conformance:      conformance,
		masterSecretPath: masterSecretPath,
		force:            force,
		mergeConfigs:     mergeConfigs,
	}, nil
}

// initFlags are the resulting values of flag preprocessing.
type initFlags struct {
	configPath       string
	masterSecretPath string
	conformance      bool
	force            bool
	mergeConfigs     bool
}

// readOrGenerateMasterSecret reads a base64 encoded master secret from file or generates a new 32 byte secret.
func (i *initCmd) readOrGenerateMasterSecret(outWriter io.Writer, fileHandler file.Handler, filename string) (uri.MasterSecret, error) {
	if filename != "" {
		i.log.Debugf("Reading master secret from file %q", filename)
		var secret uri.MasterSecret
		if err := fileHandler.ReadJSON(filename, &secret); err != nil {
			return uri.MasterSecret{}, err
		}

		if len(secret.Key) < crypto.MasterSecretLengthMin {
			return uri.MasterSecret{}, fmt.Errorf("provided master secret is smaller than the required minimum of %d Bytes", crypto.MasterSecretLengthMin)
		}
		if len(secret.Salt) < crypto.RNGLengthDefault {
			return uri.MasterSecret{}, fmt.Errorf("provided salt is smaller than the required minimum of %d Bytes", crypto.RNGLengthDefault)
		}
		return secret, nil
	}

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
	if err := fileHandler.WriteJSON(constants.MasterSecretFilename, secret, file.OptNone); err != nil {
		return uri.MasterSecret{}, err
	}
	fmt.Fprintf(outWriter, "Your Constellation master secret was successfully written to ./%s\n", constants.MasterSecretFilename)
	return secret, nil
}

func readIPFromIDFile(fileHandler file.Handler) (string, error) {
	var idFile clusterid.File
	if err := fileHandler.ReadJSON(constants.ClusterIDsFileName, &idFile); err != nil {
		return "", err
	}
	if idFile.IP == "" {
		return "", fmt.Errorf("missing IP address in %q", constants.ClusterIDsFileName)
	}
	return idFile.IP, nil
}

func (i *initCmd) getMarshaledServiceAccountURI(provider cloudprovider.Provider, config *config.Config, fileHandler file.Handler) (string, error) {
	i.log.Debugf("Getting service account URI")
	switch provider {
	case cloudprovider.GCP:
		i.log.Debugf("Handling case for GCP")
		path := config.Provider.GCP.ServiceAccountKeyPath
		i.log.Debugf("GCP service account key path %s", path)

		var key gcpshared.ServiceAccountKey
		if err := fileHandler.ReadJSON(path, &key); err != nil {
			return "", fmt.Errorf("reading service account key from path %q: %w", path, err)
		}
		i.log.Debugf("Read GCP service account key from path")
		return key.ToCloudServiceAccountURI(), nil

	case cloudprovider.AWS:
		i.log.Debugf("Handling case for AWS")
		return "", nil // AWS does not need a service account URI
	case cloudprovider.Azure:
		i.log.Debugf("Handling case for Azure")
		creds := azureshared.ApplicationCredentials{
			TenantID:          config.Provider.Azure.TenantID,
			AppClientID:       config.Provider.Azure.AppClientID,
			ClientSecretValue: config.Provider.Azure.ClientSecretValue,
			Location:          config.Provider.Azure.Location,
		}
		return creds.ToCloudServiceAccountURI(), nil

	case cloudprovider.QEMU:
		i.log.Debugf("Handling case for QEMU")
		return "", nil // QEMU does not use service account keys

	default:
		return "", fmt.Errorf("unsupported cloud provider %q", provider)
	}
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
	err error
}

// Error returns the error message.
func (e *nonRetriableError) Error() string {
	return e.err.Error()
}

// Unwrap returns the wrapped error.
func (e *nonRetriableError) Unwrap() error {
	return e.err
}
