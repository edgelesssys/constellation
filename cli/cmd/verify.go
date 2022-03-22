package cmd

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/edgelesssys/constellation/cli/status"
	"github.com/edgelesssys/constellation/coordinator/atls"
	"github.com/edgelesssys/constellation/coordinator/attestation/vtpm"
	"github.com/edgelesssys/constellation/coordinator/pubapi/pubproto"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	rpcStatus "google.golang.org/grpc/status"
)

func newVerifyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verify azure|gcp",
		Short: "Verify the confidential properties of your Constellation.",
		Long:  "Verify the confidential properties of your Constellation.",
	}

	cmd.PersistentFlags().String("owner-id", "", "verify the Constellation using the owner identity derived from the master secret.")
	cmd.PersistentFlags().String("unique-id", "", "verify the Constellation using the unique cluster identity.")

	cmd.AddCommand(newVerifyGCPCmd())
	cmd.AddCommand(newVerifyAzureCmd())
	cmd.AddCommand(newVerifyGCPNonCVMCmd())

	return cmd
}

func runVerify(cmd *cobra.Command, args []string, pcrs map[uint32][]byte, validator atls.Validator) error {
	if err := warnAboutPCRs(cmd, pcrs, false); err != nil {
		return err
	}

	verifier := verifier{
		newConn:   newVerifiedConn,
		newClient: pubproto.NewAPIClient,
	}
	return verify(cmd.Context(), cmd.OutOrStdout(), net.JoinHostPort(args[0], args[1]), []atls.Validator{validator}, verifier)
}

func verify(ctx context.Context, w io.Writer, target string, validators []atls.Validator, verifier verifier) error {
	conn, err := verifier.newConn(ctx, target, validators)
	if err != nil {
		return err
	}
	defer conn.Close()
	client := verifier.newClient(conn)

	if _, err := client.GetState(ctx, &pubproto.GetStateRequest{}); err != nil {
		if err, ok := rpcStatus.FromError(err); ok {
			return fmt.Errorf("unable to verify Constellation cluster: %s", err.Message())
		}
		return err
	}
	fmt.Fprintln(w, "OK")
	return nil
}

// prepareValidator parses parameters and updates the PCR map.
func prepareValidator(cmd *cobra.Command, pcrs map[uint32][]byte) error {
	ownerID, err := cmd.Flags().GetString("owner-id")
	if err != nil {
		return err
	}
	clusterID, err := cmd.Flags().GetString("unique-id")
	if err != nil {
		return err
	}
	if ownerID == "" && clusterID == "" {
		return errors.New("neither owner identity nor unique identity provided to verify the Constellation")
	}

	return updatePCRMap(pcrs, ownerID, clusterID)
}

func updatePCRMap(pcrs map[uint32][]byte, ownerID, clusterID string) error {
	if err := addOrSkipPCR(pcrs, uint32(vtpm.PCRIndexOwnerID), ownerID); err != nil {
		return err
	}
	return addOrSkipPCR(pcrs, uint32(vtpm.PCRIndexClusterID), clusterID)
}

// addOrSkipPCR adds a new entry to the map, or removes the key if the input is an empty string.
//
// When adding, the input is first decoded from base64.
// We then calculate the expected PCR by hashing the input using SHA256,
// appending expected PCR for initialization, and then hashing once more.
func addOrSkipPCR(toAdd map[uint32][]byte, pcrIndex uint32, encoded string) error {
	if encoded == "" {
		delete(toAdd, pcrIndex)
		return nil
	}
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return fmt.Errorf("input [%s] is not base64 encoded: %w", encoded, err)
	}
	// new_pcr_value := hash(old_pcr_value || data_to_extend)
	// Since we use the TPM2_PCR_Event call to extend the PCR, data_to_extend is the hash of our input
	hashedInput := sha256.Sum256(decoded)
	expectedPcr := sha256.Sum256(append(toAdd[pcrIndex], hashedInput[:]...))
	toAdd[pcrIndex] = expectedPcr[:]
	return nil
}

type verifier struct {
	newConn   func(context.Context, string, []atls.Validator) (status.ClientConn, error)
	newClient func(cc grpc.ClientConnInterface) pubproto.APIClient
}

// newVerifiedConn creates a grpc over aTLS connection to the target, using the provided PCR values to verify the server.
func newVerifiedConn(ctx context.Context, target string, validators []atls.Validator) (status.ClientConn, error) {
	tlsConfig, err := atls.CreateAttestationClientTLSConfig(validators)
	if err != nil {
		return nil, err
	}

	return grpc.DialContext(
		ctx, target, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)),
	)
}

// verifyCompletion handels the completion of CLI arguments. It is frequently called
// while the user types arguments of the command to suggest completion.
func verifyCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	switch len(args) {
	case 0, 1:
		return []string{}, cobra.ShellCompDirectiveNoFileComp
	default:
		return []string{}, cobra.ShellCompDirectiveError
	}
}
