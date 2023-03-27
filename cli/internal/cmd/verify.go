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

	"github.com/edgelesssys/constellation/v2/cli/internal/cloudcmd"
	"github.com/edgelesssys/constellation/v2/cli/internal/clusterid"
	"github.com/edgelesssys/constellation/v2/internal/atls"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/crypto"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/grpc/dialer"
	"github.com/edgelesssys/constellation/v2/verify/verifyproto"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

// NewVerifyCmd returns a new cobra.Command for the verify command.
func NewVerifyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verify",
		Short: "Verify the confidential properties of a Constellation cluster",
		Long: `Verify the confidential properties of a Constellation cluster.

If arguments aren't specified, values are read from ` + "`" + constants.ClusterIDsFileName + "`.",
		Args: cobra.ExactArgs(0),
		RunE: runVerify,
	}
	cmd.Flags().String("cluster-id", "", "expected cluster identifier")
	cmd.Flags().StringP("node-endpoint", "e", "", "endpoint of the node to verify, passed as HOST[:PORT]")
	return cmd
}

type verifyCmd struct {
	log debugLog
}

func runVerify(cmd *cobra.Command, _ []string) error {
	log, err := newCLILogger(cmd)
	if err != nil {
		return fmt.Errorf("creating logger: %w", err)
	}
	defer log.Sync()

	fileHandler := file.NewHandler(afero.NewOsFs())
	verifyClient := &constellationVerifier{
		dialer: dialer.New(nil, nil, &net.Dialer{}),
		log:    log,
	}

	v := &verifyCmd{log: log}
	return v.verify(cmd, fileHandler, verifyClient)
}

func (v *verifyCmd) verify(cmd *cobra.Command, fileHandler file.Handler, verifyClient verifyClient) error {
	flags, err := v.parseVerifyFlags(cmd, fileHandler)
	if err != nil {
		return err
	}
	v.log.Debugf("Using flags: %+v", flags)

	v.log.Debugf("Loading configuration file from %q", flags.configPath)
	conf, err := config.New(fileHandler, flags.configPath, flags.force)
	var configValidationErr *config.ValidationError
	if errors.As(err, &configValidationErr) {
		cmd.PrintErrln(configValidationErr.LongMessage())
	}
	if err != nil {
		return err
	}

	v.log.Debugf("Creating aTLS Validator for %s", conf.AttestationVariant)
	validators, err := cloudcmd.NewValidator(conf, flags.maaURL, v.log)
	if err != nil {
		return err
	}

	v.log.Debugf("Updating expected PCRs")
	if err := validators.UpdateInitPCRs(flags.ownerID, flags.clusterID); err != nil {
		return err
	}

	nonce, err := crypto.GenerateRandomBytes(32)
	if err != nil {
		return err
	}
	v.log.Debugf("Generated random nonce: %x", nonce)

	if err := verifyClient.Verify(
		cmd.Context(),
		flags.endpoint,
		&verifyproto.GetAttestationRequest{
			Nonce: nonce,
		},
		validators.V(cmd),
	); err != nil {
		return err
	}

	cmd.Println("OK")
	return nil
}

func (v *verifyCmd) parseVerifyFlags(cmd *cobra.Command, fileHandler file.Handler) (verifyFlags, error) {
	configPath, err := cmd.Flags().GetString("config")
	if err != nil {
		return verifyFlags{}, fmt.Errorf("parsing config path argument: %w", err)
	}
	v.log.Debugf("Flag 'config' set to %q", configPath)

	ownerID := ""
	clusterID, err := cmd.Flags().GetString("cluster-id")
	if err != nil {
		return verifyFlags{}, fmt.Errorf("parsing cluster-id argument: %w", err)
	}
	v.log.Debugf("Flag 'cluster-id' set to %q", clusterID)

	endpoint, err := cmd.Flags().GetString("node-endpoint")
	if err != nil {
		return verifyFlags{}, fmt.Errorf("parsing node-endpoint argument: %w", err)
	}
	v.log.Debugf("Flag 'node-endpoint' set to %q", endpoint)

	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return verifyFlags{}, fmt.Errorf("parsing force argument: %w", err)
	}
	v.log.Debugf("Flag 'force' set to %t", force)

	var idFile clusterid.File
	if err := fileHandler.ReadJSON(constants.ClusterIDsFileName, &idFile); err != nil && !errors.Is(err, afero.ErrFileNotFound) {
		return verifyFlags{}, fmt.Errorf("reading cluster ID file: %w", err)
	}

	// Get empty values from ID file
	emptyEndpoint := endpoint == ""
	emptyIDs := ownerID == "" && clusterID == ""
	if emptyEndpoint || emptyIDs {
		v.log.Debugf("Trying to supplement empty flag values from %q", constants.ClusterIDsFileName)
		if emptyEndpoint {
			cmd.Printf("Using endpoint from %q. Specify --node-endpoint to override this.\n", constants.ClusterIDsFileName)
			endpoint = idFile.IP
		}
		if emptyIDs {
			cmd.Printf("Using ID from %q. Specify --cluster-id to override this.\n", constants.ClusterIDsFileName)
			ownerID = idFile.OwnerID
			clusterID = idFile.ClusterID
		}
	}

	// Validate
	if ownerID == "" && clusterID == "" {
		return verifyFlags{}, errors.New("cluster-id not provided to verify the cluster")
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
		maaURL:     idFile.AttestationURL,
		force:      force,
	}, nil
}

type verifyFlags struct {
	endpoint   string
	ownerID    string
	clusterID  string
	configPath string
	maaURL     string
	force      bool
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

type constellationVerifier struct {
	dialer grpcInsecureDialer
	log    debugLog
}

// Verify retrieves an attestation statement from the Constellation and verifies it using the validator.
func (v *constellationVerifier) Verify(
	ctx context.Context, endpoint string, req *verifyproto.GetAttestationRequest, validator atls.Validator,
) error {
	v.log.Debugf("Dialing endpoint: %q", endpoint)
	conn, err := v.dialer.DialInsecure(ctx, endpoint)
	if err != nil {
		return fmt.Errorf("dialing init server: %w", err)
	}
	defer conn.Close()

	client := verifyproto.NewAPIClient(conn)

	v.log.Debugf("Sending attestation request")
	resp, err := client.GetAttestation(ctx, req)
	if err != nil {
		return fmt.Errorf("getting attestation: %w", err)
	}

	v.log.Debugf("Verifying attestation")
	signedData, err := validator.Validate(ctx, resp.Attestation, req.Nonce)
	if err != nil {
		return fmt.Errorf("validating attestation: %w", err)
	}

	if !bytes.Equal(signedData, []byte(constants.ConstellationVerifyServiceUserData)) {
		return errors.New("signed data in attestation does not match expected user data")
	}
	return nil
}

type verifyClient interface {
	Verify(ctx context.Context, endpoint string, req *verifyproto.GetAttestationRequest, validator atls.Validator) error
}

type grpcInsecureDialer interface {
	DialInsecure(ctx context.Context, endpoint string) (conn *grpc.ClientConn, err error)
}
