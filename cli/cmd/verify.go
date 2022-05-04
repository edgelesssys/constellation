package cmd

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/edgelesssys/constellation/cli/cloud/cloudcmd"
	"github.com/edgelesssys/constellation/cli/cloudprovider"
	"github.com/edgelesssys/constellation/cli/file"
	"github.com/edgelesssys/constellation/cli/proto"
	"github.com/edgelesssys/constellation/internal/config"
	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	rpcStatus "google.golang.org/grpc/status"
)

func newVerifyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verify {aws|azure|gcp}",
		Short: "Verify the confidential properties of your Constellation cluster.",
		Long:  "Verify the confidential properties of your Constellation cluster.",
		Args: cobra.MatchAll(
			cobra.ExactArgs(1),
			isCloudProvider(0),
			warnAWS(0),
		),
		RunE: runVerify,
	}
	cmd.Flags().String("owner-id", "", "Verify using the owner identity derived from the master secret.")
	cmd.Flags().String("unique-id", "", "Verify using the unique cluster identity.")
	cmd.Flags().StringP("node-endpoint", "e", "", "Endpoint of the node to verify. Form: HOST[:PORT]")
	must(cmd.MarkFlagRequired("node-endpoint"))
	return cmd
}

func runVerify(cmd *cobra.Command, args []string) error {
	provider := cloudprovider.FromString(args[0])
	fileHandler := file.NewHandler(afero.NewOsFs())
	protoClient := &proto.Client{}
	defer protoClient.Close()
	return verify(cmd.Context(), cmd, provider, fileHandler, protoClient)
}

func verify(ctx context.Context, cmd *cobra.Command, provider cloudprovider.Provider, fileHandler file.Handler, protoClient protoClient) error {
	flags, err := parseVerifyFlags(cmd)
	if err != nil {
		return err
	}

	config, err := config.FromFile(fileHandler, flags.devConfigPath)
	if err != nil {
		return err
	}

	validators, err := cloudcmd.NewValidators(provider, config)
	if err != nil {
		return err
	}

	if err := validators.UpdateInitPCRs(flags.ownerID, flags.clusterID); err != nil {
		return err
	}
	if validators.Warnings() != "" {
		cmd.Print(validators.Warnings())
	}

	if err := protoClient.Connect(flags.nodeHost, flags.nodePort, validators.V()); err != nil {
		return err
	}
	if _, err := protoClient.GetState(ctx); err != nil {
		if err, ok := rpcStatus.FromError(err); ok {
			return fmt.Errorf("unable to verify Constellation cluster: %s", err.Message())
		}
		return err
	}

	cmd.Println("OK")
	return nil
}

func parseVerifyFlags(cmd *cobra.Command) (verifyFlags, error) {
	ownerID, err := cmd.Flags().GetString("owner-id")
	if err != nil {
		return verifyFlags{}, err
	}
	clusterID, err := cmd.Flags().GetString("unique-id")
	if err != nil {
		return verifyFlags{}, err
	}
	if ownerID == "" && clusterID == "" {
		return verifyFlags{}, errors.New("neither owner ID nor unique ID provided to verify the cluster")
	}

	endpoint, err := cmd.Flags().GetString("node-endpoint")
	if err != nil {
		return verifyFlags{}, err
	}
	host, port, err := net.SplitHostPort(endpoint)
	if err != nil {
		if !strings.Contains(err.Error(), "missing port in address") {
			return verifyFlags{}, err
		}
		host = endpoint
		port = strconv.Itoa(constants.CoordinatorPort)
	}

	devConfigPath, err := cmd.Flags().GetString("dev-config")
	if err != nil {
		return verifyFlags{}, err
	}

	return verifyFlags{
		nodeHost:      host,
		nodePort:      port,
		devConfigPath: devConfigPath,
		ownerID:       ownerID,
		clusterID:     clusterID,
	}, nil
}

type verifyFlags struct {
	nodeHost      string
	nodePort      string
	ownerID       string
	clusterID     string
	devConfigPath string
}

// verifyCompletion handels the completion of CLI arguments. It is frequently called
// while the user types arguments of the command to suggest completion.
func verifyCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	switch len(args) {
	case 0:
		return []string{"gcp", "azure"}, cobra.ShellCompDirectiveNoFileComp
	default:
		return []string{}, cobra.ShellCompDirectiveError
	}
}
