package cmd

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"text/tabwriter"

	"github.com/edgelesssys/constellation/cli/azure"
	"github.com/edgelesssys/constellation/cli/cloud/cloudcmd"
	"github.com/edgelesssys/constellation/cli/cloudprovider"
	"github.com/edgelesssys/constellation/cli/file"
	"github.com/edgelesssys/constellation/cli/gcp"
	"github.com/edgelesssys/constellation/cli/proto"
	"github.com/edgelesssys/constellation/cli/status"
	"github.com/edgelesssys/constellation/cli/vpn"
	"github.com/edgelesssys/constellation/coordinator/atls"
	coordinatorstate "github.com/edgelesssys/constellation/coordinator/state"
	"github.com/edgelesssys/constellation/coordinator/util"
	"github.com/edgelesssys/constellation/internal/config"
	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/edgelesssys/constellation/internal/state"
	"github.com/kr/text"
	wgquick "github.com/nmiculinic/wg-quick-go"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func newInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "init",
		Short:             "Initialize the Constellation. Start your confidential Kubernetes cluster.",
		Long:              "Initialize the Constellation. Start your confidential Kubernetes cluster.",
		ValidArgsFunction: initCompletion,
		Args:              cobra.ExactArgs(0),
		RunE:              runInitialize,
	}
	cmd.Flags().String("privatekey", "", "path to your private key.")
	cmd.Flags().String("master-secret", "", "path to base64 encoded master secret.")
	cmd.Flags().Bool("wg-autoconfig", false, "enable automatic configuration of WireGuard interface")
	cmd.Flags().Bool("autoscale", false, "enable kubernetes cluster-autoscaler")
	return cmd
}

// runInitialize runs the initialize command.
func runInitialize(cmd *cobra.Command, args []string) error {
	fileHandler := file.NewHandler(afero.NewOsFs())
	vpnHandler := vpn.NewConfigHandler()
	serviceAccountCreator := cloudcmd.NewServiceAccountCreator()
	waiter := status.NewWaiter()
	protoClient := &proto.Client{}
	defer protoClient.Close()

	// We have to parse the context separately, since cmd.Context()
	// returns nil during the tests otherwise.
	return initialize(cmd.Context(), cmd, protoClient, serviceAccountCreator, fileHandler, waiter, vpnHandler)
}

// initialize initializes a Constellation. Coordinator instances are activated as Coordinators and will
// themself activate the other peers as nodes.
func initialize(ctx context.Context, cmd *cobra.Command, protCl protoClient, serviceAccCreator serviceAccountCreator,
	fileHandler file.Handler, waiter statusWaiter, vpnHandler vpnHandler,
) error {
	flags, err := evalFlagArgs(cmd, fileHandler)
	if err != nil {
		return err
	}

	config, err := config.FromFile(fileHandler, flags.devConfigPath)
	if err != nil {
		return err
	}

	var stat state.ConstellationState
	err = fileHandler.ReadJSON(constants.StateFilename, &stat)
	if errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("nothing to initialize: %w", err)
	} else if err != nil {
		return err
	}

	validators, err := cloudcmd.NewValidators(cloudprovider.FromString(stat.CloudProvider), config)
	if err != nil {
		return err
	}
	cmd.Print(validators.WarningsIncludeInit())

	cmd.Println("Creating service account ...")
	serviceAccount, stat, err := serviceAccCreator.Create(ctx, stat, config)
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

	endpoints := ipsToEndpoints(append(coordinators.PublicIPs(), nodes.PublicIPs()...), *config.CoordinatorPort)

	cmd.Println("Waiting for cloud provider to finish resource creation ...")
	waiter.InitializeValidators(validators.V())
	if err := waiter.WaitForAll(ctx, endpoints, coordinatorstate.AcceptingInit); err != nil {
		return fmt.Errorf("failed to wait for peer status: %w", err)
	}

	var autoscalingNodeGroups []string
	if flags.autoscale {
		autoscalingNodeGroups = append(autoscalingNodeGroups, nodes.GroupID)
	}

	input := activationInput{
		coordinatorPubIP:       coordinators.PublicIPs()[0],
		pubKey:                 flags.userPubKey,
		masterSecret:           flags.masterSecret,
		nodePrivIPs:            nodes.PrivateIPs(),
		autoscalingNodeGroups:  autoscalingNodeGroups,
		cloudServiceAccountURI: serviceAccount,
	}
	result, err := activate(ctx, cmd, protCl, input, config, validators.V())
	if err != nil {
		return err
	}

	err = result.writeOutput(cmd.OutOrStdout(), fileHandler)
	if err != nil {
		return err
	}

	vpnConfig, err := vpnHandler.Create(result.coordinatorPubKey, result.coordinatorPubIP, string(flags.userPrivKey), result.clientVpnIP, wireguardAdminMTU)
	if err != nil {
		return err
	}

	if err := writeWGQuickFile(fileHandler, vpnHandler, vpnConfig); err != nil {
		return fmt.Errorf("write wg-quick file: %w", err)
	}

	if flags.autoconfigureWG {
		if err := vpnHandler.Apply(vpnConfig); err != nil {
			return err
		}
	}

	return nil
}

func activate(ctx context.Context, cmd *cobra.Command, client protoClient, input activationInput,
	config *config.Config, validators []atls.Validator,
) (activationResult, error) {
	err := client.Connect(input.coordinatorPubIP, *config.CoordinatorPort, validators)
	if err != nil {
		return activationResult{}, err
	}

	respCl, err := client.Activate(ctx, input.pubKey, input.masterSecret, input.nodePrivIPs, input.autoscalingNodeGroups, input.cloudServiceAccountURI)
	if err != nil {
		return activationResult{}, err
	}

	indentOut := text.NewIndentWriter(cmd.OutOrStdout(), []byte{'\t'})
	cmd.Println("Activating the cluster ...")
	if err := respCl.WriteLogStream(indentOut); err != nil {
		return activationResult{}, err
	}

	clientVpnIp, err := respCl.GetClientVpnIp()
	if err != nil {
		return activationResult{}, err
	}
	coordinatorPubKey, err := respCl.GetCoordinatorVpnKey()
	if err != nil {
		return activationResult{}, err
	}
	kubeconfig, err := respCl.GetKubeconfig()
	if err != nil {
		return activationResult{}, err
	}
	ownerID, err := respCl.GetOwnerID()
	if err != nil {
		return activationResult{}, err
	}
	clusterID, err := respCl.GetClusterID()
	if err != nil {
		return activationResult{}, err
	}

	return activationResult{
		clientVpnIP:       clientVpnIp,
		coordinatorPubKey: coordinatorPubKey,
		coordinatorPubIP:  input.coordinatorPubIP,
		kubeconfig:        kubeconfig,
		ownerID:           ownerID,
		clusterID:         clusterID,
	}, nil
}

type activationInput struct {
	coordinatorPubIP       string
	pubKey                 []byte
	masterSecret           []byte
	nodePrivIPs            []string
	autoscalingNodeGroups  []string
	cloudServiceAccountURI string
}

type activationResult struct {
	clientVpnIP       string
	coordinatorPubKey string
	coordinatorPubIP  string
	kubeconfig        string
	ownerID           string
	clusterID         string
}

// writeWGQuickFile writes the wg-quick file to the default path.
func writeWGQuickFile(fileHandler file.Handler, vpnHandler vpnHandler, vpnConfig *wgquick.Config) error {
	data, err := vpnHandler.Marshal(vpnConfig)
	if err != nil {
		return err
	}
	return fileHandler.Write(constants.WGQuickConfigFilename, data, file.OptNone)
}

func (r activationResult) writeOutput(wr io.Writer, fileHandler file.Handler) error {
	fmt.Fprint(wr, "Your Constellation was successfully initialized.\n\n")

	tw := tabwriter.NewWriter(wr, 0, 0, 2, ' ', 0)
	writeRow(tw, "Your WireGuard IP", r.clientVpnIP)
	writeRow(tw, "Coordinator's public IP", r.coordinatorPubIP)
	writeRow(tw, "Coordinator's public key", r.coordinatorPubKey)
	writeRow(tw, "Constellation's owner identifier", r.ownerID)
	writeRow(tw, "Constellation's unique identifier", r.clusterID)
	writeRow(tw, "WireGuard configuration file", constants.WGQuickConfigFilename)
	writeRow(tw, "Kubernetes configuration", constants.AdminConfFilename)
	tw.Flush()
	fmt.Fprintln(wr)

	if err := fileHandler.Write(constants.AdminConfFilename, []byte(r.kubeconfig), file.OptNone); err != nil {
		return fmt.Errorf("write kubeconfig: %w", err)
	}

	fmt.Fprintln(wr, "You can now connect to your Constellation by executing:")
	fmt.Fprintf(wr, "\twg-quick up ./%s\n", constants.WGQuickConfigFilename)
	fmt.Fprintf(wr, "\texport KUBECONFIG=\"$PWD/%s\"\n", constants.AdminConfFilename)
	return nil
}

func writeRow(wr io.Writer, col1 string, col2 string) {
	fmt.Fprint(wr, col1, "\t", col2, "\n")
}

// evalFlagArgs gets the flag values and does preprocessing of these values like
// reading the content from file path flags and deriving other values from flag combinations.
func evalFlagArgs(cmd *cobra.Command, fileHandler file.Handler) (initFlags, error) {
	userPrivKeyPath, err := cmd.Flags().GetString("privatekey")
	if err != nil {
		return initFlags{}, err
	}
	userPrivKey, userPubKey, err := readOrGenerateVPNKey(fileHandler, userPrivKeyPath)
	if err != nil {
		return initFlags{}, err
	}
	autoconfigureWG, err := cmd.Flags().GetBool("wg-autoconfig")
	if err != nil {
		return initFlags{}, err
	}
	masterSecretPath, err := cmd.Flags().GetString("master-secret")
	if err != nil {
		return initFlags{}, err
	}
	masterSecret, err := readOrGeneratedMasterSecret(cmd.OutOrStdout(), fileHandler, masterSecretPath)
	if err != nil {
		return initFlags{}, err
	}
	autoscale, err := cmd.Flags().GetBool("autoscale")
	if err != nil {
		return initFlags{}, err
	}
	devConfigPath, err := cmd.Flags().GetString("dev-config")
	if err != nil {
		return initFlags{}, err
	}

	return initFlags{
		devConfigPath:   devConfigPath,
		userPrivKey:     userPrivKey,
		userPubKey:      userPubKey,
		autoconfigureWG: autoconfigureWG,
		autoscale:       autoscale,
		masterSecret:    masterSecret,
	}, nil
}

// initFlags are the resulting values of flag preprocessing.
type initFlags struct {
	devConfigPath   string
	userPrivKey     []byte
	userPubKey      []byte
	masterSecret    []byte
	autoconfigureWG bool
	autoscale       bool
}

func readOrGenerateVPNKey(fileHandler file.Handler, privKeyPath string) (privKey, pubKey []byte, err error) {
	var privKeyParsed wgtypes.Key
	if privKeyPath == "" {
		privKeyParsed, err = wgtypes.GeneratePrivateKey()
		if err != nil {
			return nil, nil, err
		}
		privKey = []byte(privKeyParsed.String())
	} else {
		privKey, err = fileHandler.Read(privKeyPath)
		if err != nil {
			return nil, nil, err
		}
		privKeyParsed, err = wgtypes.ParseKey(string(privKey))
		if err != nil {
			return nil, nil, err
		}
	}

	pubKey = []byte(privKeyParsed.PublicKey().String())

	return privKey, pubKey, nil
}

func ipsToEndpoints(ips []string, port string) []string {
	var endpoints []string
	for _, ip := range ips {
		endpoints = append(endpoints, net.JoinHostPort(ip, port))
	}
	return endpoints
}

// readOrGeneratedMasterSecret reads a base64 encoded master secret from file or generates a new 32 byte secret.
func readOrGeneratedMasterSecret(w io.Writer, fileHandler file.Handler, filename string) ([]byte, error) {
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
		if len(decoded) < masterSecretLengthMin {
			return nil, errors.New("provided master secret is smaller than the required minimum of 16 Bytes")
		}
		return decoded, nil
	}

	// No file given, generate a new secret, and save it to disk
	masterSecret, err := util.GenerateRandomBytes(masterSecretLengthDefault)
	if err != nil {
		return nil, err
	}
	if err := fileHandler.Write(constants.MasterSecretFilename, []byte(base64.StdEncoding.EncodeToString(masterSecret)), file.OptNone); err != nil {
		return nil, err
	}
	fmt.Fprintf(w, "Your Constellation master secret was successfully written to ./%s\n", constants.MasterSecretFilename)
	return masterSecret, nil
}

func getScalingGroupsFromConfig(stat state.ConstellationState, config *config.Config) (coordinators, nodes ScalingGroup, err error) {
	switch {
	case len(stat.EC2Instances) != 0:
		return getAWSInstances(stat)
	case len(stat.GCPCoordinators) != 0:
		return getGCPInstances(stat, config)
	case len(stat.AzureCoordinators) != 0:
		return getAzureInstances(stat, config)
	default:
		return ScalingGroup{}, ScalingGroup{}, errors.New("no instances to init")
	}
}

func getAWSInstances(stat state.ConstellationState) (coordinators, nodes ScalingGroup, err error) {
	coordinatorID, coordinator, err := stat.EC2Instances.GetOne()
	if err != nil {
		return
	}
	// GroupID of coordinators is empty, since they currently do not scale.
	coordinators = ScalingGroup{Instances: Instances{Instance(coordinator)}, GroupID: ""}

	nodeMap := stat.EC2Instances.GetOthers(coordinatorID)
	if len(nodeMap) == 0 {
		return ScalingGroup{}, ScalingGroup{}, errors.New("no nodes available, can't create Constellation with one instance")
	}

	var nodeInstances Instances
	for _, node := range nodeMap {
		nodeInstances = append(nodeInstances, Instance(node))
	}

	// TODO: make min / max configurable and abstract autoscaling for different cloud providers
	// TODO: GroupID of nodes is empty, since they currently do not scale.
	nodes = ScalingGroup{Instances: nodeInstances, GroupID: ""}

	return
}

func getGCPInstances(stat state.ConstellationState, config *config.Config) (coordinators, nodes ScalingGroup, err error) {
	_, coordinator, err := stat.GCPCoordinators.GetOne()
	if err != nil {
		return
	}
	// GroupID of coordinators is empty, since they currently do not scale.
	coordinators = ScalingGroup{Instances: Instances{Instance(coordinator)}, GroupID: ""}

	nodeMap := stat.GCPNodes
	if len(nodeMap) == 0 {
		return ScalingGroup{}, ScalingGroup{}, errors.New("no nodes available, can't create Constellation with one instance")
	}

	var nodeInstances Instances
	for _, node := range nodeMap {
		nodeInstances = append(nodeInstances, Instance(node))
	}

	// TODO: make min / max configurable and abstract autoscaling for different cloud providers
	nodes = ScalingGroup{
		Instances: nodeInstances,
		GroupID:   gcp.AutoscalingNodeGroup(stat.GCPProject, stat.GCPZone, stat.GCPNodeInstanceGroup, *config.AutoscalingNodeGroupsMin, *config.AutoscalingNodeGroupsMax),
	}

	return
}

func getAzureInstances(stat state.ConstellationState, config *config.Config) (coordinators, nodes ScalingGroup, err error) {
	_, coordinator, err := stat.AzureCoordinators.GetOne()
	if err != nil {
		return
	}
	// GroupID of coordinators is empty, since they currently do not scale.
	coordinators = ScalingGroup{Instances: Instances{Instance(coordinator)}, GroupID: ""}

	nodeMap := stat.AzureNodes
	if len(nodeMap) == 0 {
		return ScalingGroup{}, ScalingGroup{}, errors.New("no nodes available, can't create Constellation with one instance")
	}

	var nodeInstances Instances
	for _, node := range nodeMap {
		nodeInstances = append(nodeInstances, Instance(node))
	}

	// TODO: make min / max configurable and abstract autoscaling for different cloud providers
	nodes = ScalingGroup{
		Instances: nodeInstances,
		GroupID:   azure.AutoscalingNodeGroup(stat.AzureNodesScaleSet, *config.AutoscalingNodeGroupsMin, *config.AutoscalingNodeGroupsMax),
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

//
// TODO: Code below is target of multicloud refactoring.
//

// Instance is a cloud instance.
type Instance struct {
	PublicIP  string
	PrivateIP string
}

type Instances []Instance

type ScalingGroup struct {
	Instances
	GroupID string
}

// PublicIPs returns the public IPs of all the instances.
func (i Instances) PublicIPs() []string {
	var ips []string
	for _, instance := range i {
		ips = append(ips, instance.PublicIP)
	}
	return ips
}

// PrivateIPs returns the private IPs of all the instances of the Constellation.
func (i Instances) PrivateIPs() []string {
	var ips []string
	for _, instance := range i {
		ips = append(ips, instance.PrivateIP)
	}
	return ips
}
