package cmd

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"strconv"
	"text/tabwriter"
	"time"

	"github.com/edgelesssys/constellation/bootstrapper/initproto"
	"github.com/edgelesssys/constellation/cli/internal/azure"
	"github.com/edgelesssys/constellation/cli/internal/cloudcmd"
	"github.com/edgelesssys/constellation/cli/internal/gcp"
	"github.com/edgelesssys/constellation/cli/internal/helm"
	"github.com/edgelesssys/constellation/internal/azureshared"
	"github.com/edgelesssys/constellation/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/internal/cloud/cloudtypes"
	"github.com/edgelesssys/constellation/internal/config"
	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/edgelesssys/constellation/internal/crypto"
	"github.com/edgelesssys/constellation/internal/deploy/ssh"
	"github.com/edgelesssys/constellation/internal/file"
	"github.com/edgelesssys/constellation/internal/gcpshared"
	"github.com/edgelesssys/constellation/internal/grpc/dialer"
	grpcRetry "github.com/edgelesssys/constellation/internal/grpc/retry"
	"github.com/edgelesssys/constellation/internal/license"
	"github.com/edgelesssys/constellation/internal/retry"
	"github.com/edgelesssys/constellation/internal/state"
	kms "github.com/edgelesssys/constellation/kms/setup"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

// NewInitCmd returns a new cobra.Command for the init command.
func NewInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "init",
		Short:             "Initialize the Constellation cluster",
		Long:              "Initialize the Constellation cluster. Start your confidential Kubernetes.",
		ValidArgsFunction: initCompletion,
		Args:              cobra.ExactArgs(0),
		RunE:              runInitialize,
	}
	cmd.Flags().String("master-secret", "", "path to base64-encoded master secret")
	cmd.Flags().String("endpoint", "", "endpoint of the bootstrapper, passed as HOST[:PORT]")
	cmd.Flags().Bool("autoscale", false, "enable Kubernetes cluster-autoscaler")
	return cmd
}

// runInitialize runs the initialize command.
func runInitialize(cmd *cobra.Command, args []string) error {
	fileHandler := file.NewHandler(afero.NewOsFs())
	newDialer := func(validator *cloudcmd.Validator) *dialer.Dialer {
		return dialer.New(nil, validator.V(cmd), &net.Dialer{})
	}
	helmLoader := &helm.ChartLoader{}
	return initialize(cmd, newDialer, fileHandler, helmLoader, license.NewClient())
}

// initialize initializes a Constellation.
func initialize(cmd *cobra.Command, newDialer func(validator *cloudcmd.Validator) *dialer.Dialer,
	fileHandler file.Handler, helmLoader helmLoader, quotaChecker license.QuotaChecker,
) error {
	flags, err := evalFlagArgs(cmd, fileHandler)
	if err != nil {
		return err
	}

	var stat state.ConstellationState
	err = fileHandler.ReadJSON(constants.StateFilename, &stat)
	if errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("missing Constellation state file: %w. Please do 'constellation create ...' before 'constellation init'", err)
	} else if err != nil {
		return fmt.Errorf("loading Constellation state file: %w", err)
	}

	provider := cloudprovider.FromString(stat.CloudProvider)

	config, err := readConfig(cmd.OutOrStdout(), fileHandler, flags.configPath, provider)
	if err != nil {
		return fmt.Errorf("reading and validating config: %w", err)
	}

	checker := license.NewChecker(quotaChecker, fileHandler)
	if err := checker.CheckLicense(cmd.Context(), cmd.Printf); err != nil {
		cmd.Printf("License check failed: %v", err)
	}

	var sshUsers []*ssh.UserKey
	for _, user := range config.SSHUsers {
		sshUsers = append(sshUsers, &ssh.UserKey{
			Username:  user.Username,
			PublicKey: user.PublicKey,
		})
	}

	validator, err := cloudcmd.NewValidator(provider, config)
	if err != nil {
		return err
	}

	serviceAccURI, err := getMarschaledServiceAccountURI(provider, config, fileHandler)
	if err != nil {
		return err
	}

	workers, err := getScalingGroupsFromState(stat, config)
	if err != nil {
		return err
	}

	var autoscalingNodeGroups []string
	if flags.autoscale {
		autoscalingNodeGroups = append(autoscalingNodeGroups, workers.GroupID)
	}

	cmd.Println("Loading Helm charts ...")
	helmDeployments, err := helmLoader.Load(stat.CloudProvider)
	if err != nil {
		return fmt.Errorf("loading Helm charts: %w", err)
	}

	cmd.Println("Initializing cluster ...")
	req := &initproto.InitRequest{
		AutoscalingNodeGroups:  autoscalingNodeGroups,
		MasterSecret:           flags.masterSecret.Key,
		Salt:                   flags.masterSecret.Salt,
		KmsUri:                 kms.ClusterKMSURI,
		StorageUri:             kms.NoStoreURI,
		KeyEncryptionKeyId:     "",
		UseExistingKek:         false,
		CloudServiceAccountUri: serviceAccURI,
		KubernetesVersion:      config.KubernetesVersion,
		SshUserKeys:            ssh.ToProtoSlice(sshUsers),
		HelmDeployments:        helmDeployments,
		EnforcedPcrs:           getEnforcedMeasurements(provider, config),
	}
	resp, err := initCall(cmd.Context(), newDialer(validator), flags.endpoint, req)
	if err != nil {
		return err
	}

	if err := writeOutput(resp, flags.endpoint, cmd.OutOrStdout(), fileHandler); err != nil {
		return err
	}

	return nil
}

func initCall(ctx context.Context, dialer grpcDialer, ip string, req *initproto.InitRequest) (*initproto.InitResponse, error) {
	doer := &initDoer{
		dialer:   dialer,
		endpoint: net.JoinHostPort(ip, strconv.Itoa(constants.BootstrapperPort)),
		req:      req,
	}
	retrier := retry.NewIntervalRetrier(doer, 30*time.Second, grpcRetry.ServiceIsUnavailable)
	if err := retrier.Do(ctx); err != nil {
		return nil, err
	}
	return doer.resp, nil
}

type initDoer struct {
	dialer   grpcDialer
	endpoint string
	req      *initproto.InitRequest
	resp     *initproto.InitResponse
}

func (d *initDoer) Do(ctx context.Context) error {
	conn, err := d.dialer.Dial(ctx, d.endpoint)
	if err != nil {
		return fmt.Errorf("dialing init server: %w", err)
	}
	defer conn.Close()
	protoClient := initproto.NewAPIClient(conn)
	resp, err := protoClient.Init(ctx, d.req)
	if err != nil {
		return fmt.Errorf("init call: %w", err)
	}
	d.resp = resp
	return nil
}

func writeOutput(resp *initproto.InitResponse, ip string, wr io.Writer, fileHandler file.Handler) error {
	fmt.Fprint(wr, "Your Constellation cluster was successfully initialized.\n\n")

	ownerID := base64.StdEncoding.EncodeToString(resp.OwnerId)
	clusterID := base64.StdEncoding.EncodeToString(resp.ClusterId)

	tw := tabwriter.NewWriter(wr, 0, 0, 2, ' ', 0)
	// writeRow(tw, "Constellation cluster's owner identifier", ownerID)
	writeRow(tw, "Constellation cluster identifier", clusterID)
	writeRow(tw, "Kubernetes configuration", constants.AdminConfFilename)
	tw.Flush()
	fmt.Fprintln(wr)

	if err := fileHandler.Write(constants.AdminConfFilename, resp.Kubeconfig, file.OptNone); err != nil {
		return fmt.Errorf("writing kubeconfig: %w", err)
	}

	idFile := clusterIDsFile{
		ClusterID: clusterID,
		OwnerID:   ownerID,
		IP:        ip,
	}
	if err := fileHandler.WriteJSON(constants.ClusterIDsFileName, idFile, file.OptOverwrite); err != nil {
		return fmt.Errorf("writing Constellation id file: %w", err)
	}

	fmt.Fprintln(wr, "You can now connect to your cluster by executing:")
	fmt.Fprintf(wr, "\texport KUBECONFIG=\"$PWD/%s\"\n", constants.AdminConfFilename)
	return nil
}

func writeRow(wr io.Writer, col1 string, col2 string) {
	fmt.Fprint(wr, col1, "\t", col2, "\n")
}

func getEnforcedMeasurements(provider cloudprovider.Provider, config *config.Config) []uint32 {
	switch provider {
	case cloudprovider.Azure:
		return config.Provider.Azure.EnforcedMeasurements
	case cloudprovider.GCP:
		return config.Provider.GCP.EnforcedMeasurements
	case cloudprovider.QEMU:
		return config.Provider.QEMU.EnforcedMeasurements
	default:
		return nil
	}
}

// evalFlagArgs gets the flag values and does preprocessing of these values like
// reading the content from file path flags and deriving other values from flag combinations.
func evalFlagArgs(cmd *cobra.Command, fileHandler file.Handler) (initFlags, error) {
	masterSecretPath, err := cmd.Flags().GetString("master-secret")
	if err != nil {
		return initFlags{}, fmt.Errorf("parsing master-secret path flag: %w", err)
	}
	masterSecret, err := readOrGenerateMasterSecret(cmd.OutOrStdout(), fileHandler, masterSecretPath)
	if err != nil {
		return initFlags{}, fmt.Errorf("parsing or generating master mastersecret from file %s: %w", masterSecretPath, err)
	}
	endpoint, err := cmd.Flags().GetString("endpoint")
	if err != nil {
		return initFlags{}, fmt.Errorf("parsing endpoint flag: %w", err)
	}
	if endpoint == "" {
		endpoint, err = readIPFromIDFile(fileHandler)
		if err != nil {
			return initFlags{}, fmt.Errorf("getting bootstrapper endpoint: %w", err)
		}
	}
	autoscale, err := cmd.Flags().GetBool("autoscale")
	if err != nil {
		return initFlags{}, fmt.Errorf("parsing autoscale flag: %w", err)
	}
	configPath, err := cmd.Flags().GetString("config")
	if err != nil {
		return initFlags{}, fmt.Errorf("parsing config path flag: %w", err)
	}

	return initFlags{
		configPath:   configPath,
		endpoint:     endpoint,
		autoscale:    autoscale,
		masterSecret: masterSecret,
	}, nil
}

// initFlags are the resulting values of flag preprocessing.
type initFlags struct {
	configPath   string
	masterSecret masterSecret
	endpoint     string
	autoscale    bool
}

// masterSecret holds the master key and salt for deriving keys.
type masterSecret struct {
	Key  []byte `json:"key"`
	Salt []byte `json:"salt"`
}

// readOrGenerateMasterSecret reads a base64 encoded master secret from file or generates a new 32 byte secret.
func readOrGenerateMasterSecret(writer io.Writer, fileHandler file.Handler, filename string) (masterSecret, error) {
	if filename != "" {
		var secret masterSecret
		if err := fileHandler.ReadJSON(filename, &secret); err != nil {
			return masterSecret{}, err
		}

		if len(secret.Key) < crypto.MasterSecretLengthMin {
			return masterSecret{}, fmt.Errorf("provided master secret is smaller than the required minimum of %d Bytes", crypto.MasterSecretLengthMin)
		}
		if len(secret.Salt) < crypto.RNGLengthDefault {
			return masterSecret{}, fmt.Errorf("provided salt is smaller than the required minimum of %d Bytes", crypto.RNGLengthDefault)
		}
		return secret, nil
	}

	// No file given, generate a new secret, and save it to disk
	key, err := crypto.GenerateRandomBytes(crypto.MasterSecretLengthDefault)
	if err != nil {
		return masterSecret{}, err
	}
	salt, err := crypto.GenerateRandomBytes(crypto.RNGLengthDefault)
	if err != nil {
		return masterSecret{}, err
	}
	secret := masterSecret{
		Key:  key,
		Salt: salt,
	}

	if err := fileHandler.WriteJSON(constants.MasterSecretFilename, secret, file.OptNone); err != nil {
		return masterSecret{}, err
	}
	fmt.Fprintf(writer, "Your Constellation master secret was successfully written to ./%s\n", constants.MasterSecretFilename)
	return secret, nil
}

func readIPFromIDFile(fileHandler file.Handler) (string, error) {
	var idFile clusterIDsFile
	if err := fileHandler.ReadJSON(constants.ClusterIDsFileName, &idFile); err != nil {
		return "", err
	}
	if idFile.IP == "" {
		return "", fmt.Errorf("missing IP address in %q", constants.ClusterIDsFileName)
	}
	return idFile.IP, nil
}

func getMarschaledServiceAccountURI(provider cloudprovider.Provider, config *config.Config, fileHandler file.Handler) (string, error) {
	switch provider {
	case cloudprovider.GCP:
		path := config.Provider.GCP.ServiceAccountKeyPath

		var key gcpshared.ServiceAccountKey
		if err := fileHandler.ReadJSON(path, &key); err != nil {
			return "", fmt.Errorf("reading service account key from path %q: %w", path, err)
		}

		return key.ToCloudServiceAccountURI(), nil

	case cloudprovider.Azure:
		creds := azureshared.ApplicationCredentials{
			TenantID:          config.Provider.Azure.TenantID,
			AppClientID:       config.Provider.Azure.AppClientID,
			ClientSecretValue: config.Provider.Azure.ClientSecretValue,
			Location:          config.Provider.Azure.Location,
		}
		return creds.ToCloudServiceAccountURI(), nil

	case cloudprovider.QEMU:
		return "", nil // QEMU does not use service account keys

	default:
		return "", fmt.Errorf("unsupported cloud provider %q", provider)
	}
}

func getScalingGroupsFromState(stat state.ConstellationState, config *config.Config) (workers cloudtypes.ScalingGroup, err error) {
	switch cloudprovider.FromString(stat.CloudProvider) {
	case cloudprovider.GCP:
		return cloudtypes.ScalingGroup{
			GroupID: gcp.AutoscalingNodeGroup(stat.GCPProject, stat.GCPZone, stat.GCPWorkerInstanceGroup, config.AutoscalingNodeGroupMin, config.AutoscalingNodeGroupMax),
		}, nil
	case cloudprovider.Azure:
		return cloudtypes.ScalingGroup{
			GroupID: azure.AutoscalingNodeGroup(stat.AzureWorkerScaleSet, config.AutoscalingNodeGroupMin, config.AutoscalingNodeGroupMax),
		}, nil
	case cloudprovider.QEMU:
		return cloudtypes.ScalingGroup{GroupID: ""}, nil
	default:
		return cloudtypes.ScalingGroup{}, errors.New("unknown cloud provider")
	}
}

// initCompletion handels the completion of CLI arguments. It is frequently called
// while the user types arguments of the command to suggest completion.
func initCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return []string{}, cobra.ShellCompDirectiveError
	}
	return []string{}, cobra.ShellCompDirectiveDefault
}

type grpcDialer interface {
	Dial(ctx context.Context, target string) (*grpc.ClientConn, error)
}
