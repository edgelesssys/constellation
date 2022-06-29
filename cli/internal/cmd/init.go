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

	"github.com/edgelesssys/constellation/cli/internal/azure"
	"github.com/edgelesssys/constellation/cli/internal/cloudcmd"
	"github.com/edgelesssys/constellation/cli/internal/gcp"
	"github.com/edgelesssys/constellation/coordinator/initproto"
	"github.com/edgelesssys/constellation/coordinator/kms"
	"github.com/edgelesssys/constellation/coordinator/util"
	"github.com/edgelesssys/constellation/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/internal/cloud/cloudtypes"
	"github.com/edgelesssys/constellation/internal/config"
	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/edgelesssys/constellation/internal/deploy/ssh"
	"github.com/edgelesssys/constellation/internal/file"
	"github.com/edgelesssys/constellation/internal/grpc/dialer"
	"github.com/edgelesssys/constellation/internal/grpc/retry"
	"github.com/edgelesssys/constellation/internal/state"
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
	cmd.Flags().Bool("autoscale", false, "enable Kubernetes cluster-autoscaler")
	return cmd
}

// runInitialize runs the initialize command.
func runInitialize(cmd *cobra.Command, args []string) error {
	fileHandler := file.NewHandler(afero.NewOsFs())
	serviceAccountCreator := cloudcmd.NewServiceAccountCreator()
	dialer := dialer.New(nil, nil, &net.Dialer{})
	return initialize(cmd, dialer, serviceAccountCreator, fileHandler)
}

// initialize initializes a Constellation. Coordinator instances are activated as contole-plane nodes and will
// themself activate the other peers as workers.
func initialize(cmd *cobra.Command, dialer grpcDialer, serviceAccCreator serviceAccountCreator,
	fileHandler file.Handler,
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

	var sshUsers []*ssh.UserKey
	for _, user := range config.SSHUsers {
		sshUsers = append(sshUsers, &ssh.UserKey{
			Username:  user.Username,
			PublicKey: user.PublicKey,
		})
	}

	validators, err := cloudcmd.NewValidators(provider, config)
	if err != nil {
		return err
	}
	cmd.Print(validators.WarningsIncludeInit())

	cmd.Println("Creating service account ...")
	serviceAccount, stat, err := serviceAccCreator.Create(cmd.Context(), stat, config)
	if err != nil {
		return err
	}
	if err := fileHandler.WriteJSON(constants.StateFilename, stat, file.OptOverwrite); err != nil {
		return err
	}

	coordinators, nodes, err := getScalingGroupsFromConfig(stat, config)
	if err != nil {
		return err
	}

	var autoscalingNodeGroups []string
	if flags.autoscale {
		autoscalingNodeGroups = append(autoscalingNodeGroups, nodes.GroupID)
	}

	req := &initproto.InitRequest{
		AutoscalingNodeGroups:  autoscalingNodeGroups,
		MasterSecret:           flags.masterSecret,
		KmsUri:                 kms.ClusterKMSURI,
		StorageUri:             kms.NoStoreURI,
		KeyEncryptionKeyId:     "",
		UseExistingKek:         false,
		CloudServiceAccountUri: serviceAccount,
		KubernetesVersion:      "1.23.6",
		SshUserKeys:            ssh.ToProtoSlice(sshUsers),
	}
	resp, err := initCall(cmd.Context(), dialer, coordinators.PublicIPs()[0], req)
	if err != nil {
		return err
	}

	if err := writeOutput(resp, cmd.OutOrStdout(), fileHandler); err != nil {
		return err
	}

	return nil
}

func initCall(ctx context.Context, dialer grpcDialer, ip string, req *initproto.InitRequest) (*initproto.InitResponse, error) {
	doer := &initDoer{
		dialer:   dialer,
		endpoint: net.JoinHostPort(ip, strconv.Itoa(constants.CoordinatorPort)),
		req:      req,
	}
	retrier := retry.NewIntervalRetrier(doer, 30*time.Second)
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
	protoClient := initproto.NewAPIClient(conn)
	resp, err := protoClient.Init(ctx, d.req)
	if err != nil {
		return fmt.Errorf("marshalling VPN config: %w", err)
	}
	d.resp = resp
	return nil
}

func writeOutput(resp *initproto.InitResponse, wr io.Writer, fileHandler file.Handler) error {
	fmt.Fprint(wr, "Your Constellation cluster was successfully initialized.\n\n")

	tw := tabwriter.NewWriter(wr, 0, 0, 2, ' ', 0)
	writeRow(tw, "Constellation cluster's owner identifier", string(resp.OwnerId))
	writeRow(tw, "Constellation cluster's unique identifier", string(resp.ClusterId))
	writeRow(tw, "Kubernetes configuration", constants.AdminConfFilename)
	tw.Flush()
	fmt.Fprintln(wr)

	if err := fileHandler.Write(constants.AdminConfFilename, resp.Kubeconfig, file.OptNone); err != nil {
		return fmt.Errorf("write kubeconfig: %w", err)
	}

	idFile := clusterIDsFile{ClusterID: r.clusterID, OwnerID: r.ownerID, Endpoint: r.coordinatorPubIP}
	if err := fileHandler.WriteJSON(constants.ClusterIDsFileName, idFile, file.OptNone); err != nil {
		return fmt.Errorf("writing Constellation id file: %w", err)
	}

	fmt.Fprintln(wr, "You can now connect to your cluster by executing:")
	fmt.Fprintf(wr, "\texport KUBECONFIG=\"$PWD/%s\"\n", constants.AdminConfFilename)
	return nil
}

func writeRow(wr io.Writer, col1 string, col2 string) {
	fmt.Fprint(wr, col1, "\t", col2, "\n")
}

// evalFlagArgs gets the flag values and does preprocessing of these values like
// reading the content from file path flags and deriving other values from flag combinations.
func evalFlagArgs(cmd *cobra.Command, fileHandler file.Handler) (initFlags, error) {
	masterSecretPath, err := cmd.Flags().GetString("master-secret")
	if err != nil {
		return initFlags{}, fmt.Errorf("parsing master-secret path argument: %w", err)
	}
	masterSecret, err := readOrGenerateMasterSecret(cmd.OutOrStdout(), fileHandler, masterSecretPath)
	if err != nil {
		return initFlags{}, fmt.Errorf("parsing or generating master mastersecret from file %s: %w", masterSecretPath, err)
	}
	autoscale, err := cmd.Flags().GetBool("autoscale")
	if err != nil {
		return initFlags{}, fmt.Errorf("parsing autoscale argument: %w", err)
	}
	configPath, err := cmd.Flags().GetString("config")
	if err != nil {
		return initFlags{}, fmt.Errorf("parsing config path argument: %w", err)
	}

	return initFlags{
		configPath:   configPath,
		autoscale:    autoscale,
		masterSecret: masterSecret,
	}, nil
}

// initFlags are the resulting values of flag preprocessing.
type initFlags struct {
	configPath   string
	masterSecret []byte
	autoscale    bool
}

// readOrGenerateMasterSecret reads a base64 encoded master secret from file or generates a new 32 byte secret.
func readOrGenerateMasterSecret(writer io.Writer, fileHandler file.Handler, filename string) ([]byte, error) {
	if filename != "" {
		// Try to read the base64 secret from file
		encodedSecret, err := fileHandler.Read(filename)
		if err != nil {
			return nil, err
		}
		decoded, err := base64.StdEncoding.DecodeString(string(encodedSecret))
		if err != nil {
			return nil, err
		}
		if len(decoded) < constants.MasterSecretLengthMin {
			return nil, errors.New("provided master secret is smaller than the required minimum of 16 Bytes")
		}
		return decoded, nil
	}

	// No file given, generate a new secret, and save it to disk
	masterSecret, err := util.GenerateRandomBytes(constants.MasterSecretLengthDefault)
	if err != nil {
		return nil, err
	}
	if err := fileHandler.Write(constants.MasterSecretFilename, []byte(base64.StdEncoding.EncodeToString(masterSecret)), file.OptNone); err != nil {
		return nil, err
	}
	fmt.Fprintf(writer, "Your Constellation master secret was successfully written to ./%s\n", constants.MasterSecretFilename)
	return masterSecret, nil
}

func getScalingGroupsFromConfig(stat state.ConstellationState, config *config.Config) (coordinators, nodes cloudtypes.ScalingGroup, err error) {
	switch {
	case len(stat.GCPCoordinators) != 0:
		return getGCPInstances(stat, config)
	case len(stat.AzureCoordinators) != 0:
		return getAzureInstances(stat, config)
	case len(stat.QEMUCoordinators) != 0:
		return getQEMUInstances(stat, config)
	default:
		return cloudtypes.ScalingGroup{}, cloudtypes.ScalingGroup{}, errors.New("no instances to initialize")
	}
}

func getGCPInstances(stat state.ConstellationState, config *config.Config) (coordinators, nodes cloudtypes.ScalingGroup, err error) {
	if len(stat.GCPCoordinators) == 0 {
		return cloudtypes.ScalingGroup{}, cloudtypes.ScalingGroup{}, errors.New("no control-plane nodes available, can't create Constellation without any instance")
	}

	// GroupID of coordinators is empty, since they currently do not scale.
	coordinators = cloudtypes.ScalingGroup{
		Instances: stat.GCPCoordinators,
		GroupID:   "",
	}

	if len(stat.GCPNodes) == 0 {
		return cloudtypes.ScalingGroup{}, cloudtypes.ScalingGroup{}, errors.New("no worker nodes available, can't create Constellation with one instance")
	}

	// TODO: make min / max configurable and abstract autoscaling for different cloud providers
	nodes = cloudtypes.ScalingGroup{
		Instances: stat.GCPNodes,
		GroupID:   gcp.AutoscalingNodeGroup(stat.GCPProject, stat.GCPZone, stat.GCPNodeInstanceGroup, config.AutoscalingNodeGroupMin, config.AutoscalingNodeGroupMax),
	}

	return
}

func getAzureInstances(stat state.ConstellationState, config *config.Config) (coordinators, nodes cloudtypes.ScalingGroup, err error) {
	if len(stat.AzureCoordinators) == 0 {
		return cloudtypes.ScalingGroup{}, cloudtypes.ScalingGroup{}, errors.New("no control-plane nodes available, can't create Constellation cluster without any instance")
	}

	// GroupID of coordinators is empty, since they currently do not scale.
	coordinators = cloudtypes.ScalingGroup{
		Instances: stat.AzureCoordinators,
		GroupID:   "",
	}

	if len(stat.AzureNodes) == 0 {
		return cloudtypes.ScalingGroup{}, cloudtypes.ScalingGroup{}, errors.New("no worker nodes available, can't create Constellation cluster with one instance")
	}

	// TODO: make min / max configurable and abstract autoscaling for different cloud providers
	nodes = cloudtypes.ScalingGroup{
		Instances: stat.AzureNodes,
		GroupID:   azure.AutoscalingNodeGroup(stat.AzureNodesScaleSet, config.AutoscalingNodeGroupMin, config.AutoscalingNodeGroupMax),
	}
	return
}

func getQEMUInstances(stat state.ConstellationState, config *config.Config) (coordinators, nodes cloudtypes.ScalingGroup, err error) {
	coordinatorMap := stat.QEMUCoordinators
	if len(coordinatorMap) == 0 {
		return cloudtypes.ScalingGroup{}, cloudtypes.ScalingGroup{}, errors.New("no coordinators available, can't create Constellation without any instance")
	}

	// QEMU does not support autoscaling
	coordinators = cloudtypes.ScalingGroup{
		Instances: stat.QEMUCoordinators,
		GroupID:   "",
	}

	if len(stat.QEMUNodes) == 0 {
		return cloudtypes.ScalingGroup{}, cloudtypes.ScalingGroup{}, errors.New("no nodes available, can't create Constellation with one instance")
	}

	// QEMU does not support autoscaling
	nodes = cloudtypes.ScalingGroup{
		Instances: stat.QEMUNodes,
		GroupID:   "",
	}
	return
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
