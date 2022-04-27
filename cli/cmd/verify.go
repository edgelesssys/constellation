package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/edgelesssys/constellation/cli/cloud/cloudcmd"
	"github.com/edgelesssys/constellation/cli/cloudprovider"
	"github.com/edgelesssys/constellation/cli/file"
	"github.com/edgelesssys/constellation/cli/proto"
	"github.com/edgelesssys/constellation/internal/config"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	rpcStatus "google.golang.org/grpc/status"
)

func newVerifyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verify {azure|gcp} IP PORT",
		Short: "Verify the confidential properties of your Constellation.",
		Long:  "Verify the confidential properties of your Constellation.",
		Args: cobra.MatchAll(
			cobra.ExactArgs(3),
			isCloudProvider(0),
			isIP(1),
			isPort(2),
		),
		RunE: runVerify,
	}
	cmd.Flags().String("owner-id", "", "verify the Constellation using the owner identity derived from the master secret.")
	cmd.Flags().String("unique-id", "", "verify the Constellation using the unique cluster identity.")
	return cmd
}

func runVerify(cmd *cobra.Command, args []string) error {
	provider := cloudprovider.FromString(args[0])
	ip := args[1]
	port := args[2]
	fileHandler := file.NewHandler(afero.NewOsFs())
	protoClient := &proto.Client{}
	defer protoClient.Close()
	return verify(cmd.Context(), cmd, provider, ip, port, fileHandler, protoClient)
}

func verify(ctx context.Context, cmd *cobra.Command, provider cloudprovider.Provider, ip, port string, fileHandler file.Handler, protoClient protoClient) error {
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

	if err := protoClient.Connect(ip, port, validators.V()); err != nil {
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
		return verifyFlags{}, errors.New("neither owner ID nor unique ID provided to verify the Constellation")
	}

	devConfigPath, err := cmd.Flags().GetString("dev-config")
	if err != nil {
		return verifyFlags{}, err
	}

	return verifyFlags{
		devConfigPath: devConfigPath,
		ownerID:       ownerID,
		clusterID:     clusterID,
	}, nil
}

type verifyFlags struct {
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
	case 1, 2:
		return []string{}, cobra.ShellCompDirectiveNoFileComp
	default:
		return []string{}, cobra.ShellCompDirectiveError
	}
}
