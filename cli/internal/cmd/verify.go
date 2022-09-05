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
	"net"
	"strconv"
	"strings"

	"github.com/edgelesssys/constellation/cli/internal/cloudcmd"
	"github.com/edgelesssys/constellation/internal/atls"
	"github.com/edgelesssys/constellation/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/edgelesssys/constellation/internal/crypto"
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
		Long: `Verify the confidential properties of a Constellation cluster.

If arguments aren't specified, values are read from ` + "`" + constants.ClusterIDsFileName + "`.",
		Args: cobra.MatchAll(
			cobra.ExactArgs(1),
			isCloudProvider(0),
			warnAWS(0),
		),
		RunE: runVerify,
	}
	cmd.Flags().String("owner-id", "", "verify using the owner identity derived from the master secret")
	cmd.Flags().String("cluster-id", "", "verify using Constellation's cluster identifier")
	cmd.Flags().StringP("node-endpoint", "e", "", "endpoint of the node to verify, passed as HOST[:PORT]")
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
	flags, err := parseVerifyFlags(cmd, fileHandler)
	if err != nil {
		return err
	}

	config, err := readConfig(cmd.OutOrStdout(), fileHandler, flags.configPath, provider)
	if err != nil {
		return fmt.Errorf("reading and validating config: %w", err)
	}

	validators, err := cloudcmd.NewValidator(provider, config)
	if err != nil {
		return err
	}

	if err := validators.UpdateInitPCRs(flags.ownerID, flags.clusterID); err != nil {
		return err
	}

	nonce, err := crypto.GenerateRandomBytes(32)
	if err != nil {
		return err
	}
	userData, err := crypto.GenerateRandomBytes(32)
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
		validators.V(cmd),
	); err != nil {
		return err
	}

	cmd.Println("OK")
	return nil
}

func parseVerifyFlags(cmd *cobra.Command, fileHandler file.Handler) (verifyFlags, error) {
	configPath, err := cmd.Flags().GetString("config")
	if err != nil {
		return verifyFlags{}, fmt.Errorf("parsing config path argument: %w", err)
	}
	ownerID, err := cmd.Flags().GetString("owner-id")
	if err != nil {
		return verifyFlags{}, fmt.Errorf("parsing owner-id argument: %w", err)
	}
	clusterID, err := cmd.Flags().GetString("cluster-id")
	if err != nil {
		return verifyFlags{}, fmt.Errorf("parsing cluster-id argument: %w", err)
	}
	endpoint, err := cmd.Flags().GetString("node-endpoint")
	if err != nil {
		return verifyFlags{}, fmt.Errorf("parsing node-endpoint argument: %w", err)
	}

	emptyEndpoint := endpoint == ""
	emptyIDs := ownerID == "" && clusterID == ""

	var idFile clusterIDsFile
	if emptyEndpoint || emptyIDs { // Get empty values from ID file
		if err := fileHandler.ReadJSON(constants.ClusterIDsFileName, &idFile); err != nil {
			return verifyFlags{}, fmt.Errorf("reading cluster ID file: %w", err)
		}
	}
	if emptyEndpoint {
		cmd.Printf("Using endpoint from %q. Specify --node-endpoint to override this.\n", constants.ClusterIDsFileName)
		endpoint = idFile.IP
	}
	if emptyIDs {
		cmd.Printf("Using IDs from %q. Specify --owner-id  and/or --cluster-id to override this.\n", constants.ClusterIDsFileName)
		ownerID = idFile.OwnerID
		clusterID = idFile.ClusterID
	}

	// Validate
	if ownerID == "" && clusterID == "" {
		return verifyFlags{}, errors.New("neither owner-id nor cluster-id provided to verify the cluster")
	}
	endpoint, err = addPortIfMissing(endpoint, constants.VerifyServiceNodePortGRPC)
	if err != nil {
		return verifyFlags{}, fmt.Errorf("validating endpoint argument: %w", err)
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

func addPortIfMissing(endpoint string, defaultPort int) (string, error) {
	if endpoint == "" {
		return "", errors.New("endpoint is empty")
	}

	_, _, err := net.SplitHostPort(endpoint)
	if err == nil {
		return endpoint, nil
	}

	if strings.Contains(err.Error(), "missing port in address") {
		return net.JoinHostPort(endpoint, strconv.Itoa(defaultPort)), nil
	}

	return "", err
}

// verifyCompletion handles the completion of CLI arguments. It is frequently called
// while the user types arguments of the command to suggest completion.
func verifyCompletion(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
	switch len(args) {
	case 0:
		return []string{"gcp", "azure"}, cobra.ShellCompDirectiveNoFileComp
	default:
		return []string{}, cobra.ShellCompDirectiveError
	}
}

type constellationVerifier struct {
	dialer grpcInsecureDialer
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

type grpcInsecureDialer interface {
	DialInsecure(ctx context.Context, endpoint string) (conn *grpc.ClientConn, err error)
}
