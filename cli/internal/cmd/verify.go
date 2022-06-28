package cmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/edgelesssys/constellation/cli/internal/cloudcmd"
	"github.com/edgelesssys/constellation/coordinator/util"
	"github.com/edgelesssys/constellation/internal/atls"
	"github.com/edgelesssys/constellation/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/edgelesssys/constellation/internal/file"
	"github.com/edgelesssys/constellation/internal/grpc/dialer"
	"github.com/edgelesssys/constellation/verify/verifyproto"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

// NewVerifyCmd returns a new cobra.Command for the verify command.
func NewVerifyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verify {aws|azure|gcp}",
		Short: "Verify the confidential properties of a Constellation cluster",
		Long:  "Verify the confidential properties of a Constellation cluster.",
		Args: cobra.MatchAll(
			cobra.ExactArgs(1),
			isCloudProvider(0),
			warnAWS(0),
		),
		RunE: runVerify,
	}
	cmd.Flags().String("owner-id", "", "verify using the owner identity derived from the master secret")
	cmd.Flags().String("unique-id", "", "verify using the unique cluster identity")
	cmd.Flags().StringP("node-endpoint", "e", "", "endpoint of the node to verify, passed as HOST[:PORT] (required)")
	must(cmd.MarkFlagRequired("node-endpoint"))
	return cmd
}

func runVerify(cmd *cobra.Command, args []string) error {
	provider := cloudprovider.FromString(args[0])
	fileHandler := file.NewHandler(afero.NewOsFs())
	verifyClient := &constellationVerifier{dialer: dialer.New(nil, nil, &net.Dialer{})}
	return verify(cmd, provider, fileHandler, verifyClient)
}

func verify(
	cmd *cobra.Command, provider cloudprovider.Provider, fileHandler file.Handler, verifyClient verifyClient,
) error {
	flags, err := parseVerifyFlags(cmd)
	if err != nil {
		return err
	}

	config, err := readConfig(cmd.OutOrStdout(), fileHandler, flags.configPath, provider)
	if err != nil {
		return fmt.Errorf("reading and validating config: %w", err)
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

	nonce, err := util.GenerateRandomBytes(32)
	if err != nil {
		return err
	}
	userData, err := util.GenerateRandomBytes(32)
	if err != nil {
		return err
	}

	if err := verifyClient.Verify(
		cmd.Context(),
		flags.endpoint,
		&verifyproto.GetAttestationRequest{
			Nonce:    nonce,
			UserData: userData,
		},
		validators.V()[0],
	); err != nil {
		return err
	}

	cmd.Println("OK")
	return nil
}

func parseVerifyFlags(cmd *cobra.Command) (verifyFlags, error) {
	ownerID, err := cmd.Flags().GetString("owner-id")
	if err != nil {
		return verifyFlags{}, fmt.Errorf("parsing owner-id argument: %w", err)
	}
	clusterID, err := cmd.Flags().GetString("unique-id")
	if err != nil {
		return verifyFlags{}, fmt.Errorf("parsing unique-id argument: %w", err)
	}
	if ownerID == "" && clusterID == "" {
		return verifyFlags{}, errors.New("neither owner-id nor unique-id provided to verify the cluster")
	}

	endpoint, err := cmd.Flags().GetString("node-endpoint")
	if err != nil {
		return verifyFlags{}, fmt.Errorf("parsing node-endpoint argument: %w", err)
	}
	endpoint, err = validateEndpoint(endpoint, constants.VerifyServiceNodePortGRPC)
	if err != nil {
		return verifyFlags{}, fmt.Errorf("validating endpoint argument: %w", err)
	}

	configPath, err := cmd.Flags().GetString("config")
	if err != nil {
		return verifyFlags{}, fmt.Errorf("parsing config path argument: %w", err)
	}

	return verifyFlags{
		endpoint:   endpoint,
		configPath: configPath,
		ownerID:    ownerID,
		clusterID:  clusterID,
	}, nil
}

type verifyFlags struct {
	endpoint   string
	ownerID    string
	clusterID  string
	configPath string
}

// verifyCompletion handles the completion of CLI arguments. It is frequently called
// while the user types arguments of the command to suggest completion.
func verifyCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	switch len(args) {
	case 0:
		return []string{"gcp", "azure"}, cobra.ShellCompDirectiveNoFileComp
	default:
		return []string{}, cobra.ShellCompDirectiveError
	}
}

type constellationVerifier struct {
	dialer grpcDialer
}

// Verify retrieves an attestation statement from the Constellation and verifies it using the validator.
func (v *constellationVerifier) Verify(
	ctx context.Context, endpoint string, req *verifyproto.GetAttestationRequest, validator atls.Validator,
) error {
	conn, err := v.dialer.DialInsecure(ctx, endpoint)
	if err != nil {
		return fmt.Errorf("dialing init server: %w", err)
	}
	defer conn.Close()

	client := verifyproto.NewAPIClient(conn)

	resp, err := client.GetAttestation(ctx, req)
	if err != nil {
		return fmt.Errorf("getting attestation: %w", err)
	}

	signedData, err := validator.Validate(resp.Attestation, req.Nonce)
	if err != nil {
		return fmt.Errorf("validating attestation: %w", err)
	}

	if !bytes.Equal(signedData, req.UserData) {
		return errors.New("signed data in attestation does not match provided user data")
	}
	return nil
}

type verifyClient interface {
	Verify(ctx context.Context, endpoint string, req *verifyproto.GetAttestationRequest, validator atls.Validator) error
}

type grpcDialer interface {
	DialInsecure(ctx context.Context, endpoint string) (conn *grpc.ClientConn, err error)
}
