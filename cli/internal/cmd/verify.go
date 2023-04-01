/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
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
		Long: `Verify the confidential properties of a Constellation cluster.\n` +
			`If arguments aren't specified, values are read from ` + "`" + constants.ClusterIDsFileName + "`.",
		Args: cobra.ExactArgs(0),
		RunE: runVerify,
	}
	cmd.Flags().String("cluster-id", "", "expected cluster identifier")
	cmd.Flags().Bool("raw", false, "print raw attestation document")
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

func (c *verifyCmd) verify(cmd *cobra.Command, fileHandler file.Handler, verifyClient verifyClient) error {
	flags, err := c.parseVerifyFlags(cmd, fileHandler)
	if err != nil {
		return fmt.Errorf("parsing flags: %w", err)
	}
	c.log.Debugf("Using flags: %+v", flags)

	c.log.Debugf("Loading configuration file from %q", flags.configPath)
	conf, err := config.New(fileHandler, flags.configPath, flags.force)
	var configValidationErr *config.ValidationError
	if errors.As(err, &configValidationErr) {
		cmd.PrintErrln(configValidationErr.LongMessage())
	}
	if err != nil {
		return fmt.Errorf("loading config file: %w", err)
	}

	c.log.Debugf("Creating aTLS Validator for %s", conf.AttestationVariant)
	validators, err := cloudcmd.NewValidator(conf, flags.maaURL, c.log)
	if err != nil {
		return fmt.Errorf("creating aTLS validator: %w", err)
	}

	c.log.Debugf("Updating expected PCRs")
	if err := validators.UpdateInitPCRs(flags.ownerID, flags.clusterID); err != nil {
		return fmt.Errorf("updating expected PCRs: %w", err)
	}

	nonce, err := crypto.GenerateRandomBytes(32)
	if err != nil {
		return fmt.Errorf("generating random nonce: %w", err)
	}
	c.log.Debugf("Generated random nonce: %x", nonce)

	rawAttestationDoc, err := verifyClient.Verify(
		cmd.Context(),
		flags.endpoint,
		&verifyproto.GetAttestationRequest{
			Nonce: nonce,
		},
		validators.V(cmd),
	)
	if err != nil {
		return fmt.Errorf("verifying: %w", err)
	}

	cmd.Printf("Cluster Endpoint:\n\t%s\n", flags.endpoint)
	if flags.maaURL != "" {
		cmd.Printf("MAA URL:\n\t%s\n", flags.maaURL)
	}
	cmd.Println("Expected PCRs:")
	for i, pcr := range validators.PCRS() {
		cmd.Printf("\tPCR %d:\t%x - Strict: %t\n", i, pcr.Expected, !pcr.ValidationOpt)
	}
	cmd.Println("Attestation Document:")
	if err := c.printAttestationDoc(cmd, rawAttestationDoc, flags.rawOutput); err != nil {
		return fmt.Errorf("printing attestation document: %w", err)
	}

	cmd.Println("\nVerification OK")
	return nil
}

func (c *verifyCmd) parseVerifyFlags(cmd *cobra.Command, fileHandler file.Handler) (verifyFlags, error) {
	configPath, err := cmd.Flags().GetString("config")
	if err != nil {
		return verifyFlags{}, fmt.Errorf("parsing config path argument: %w", err)
	}
	c.log.Debugf("Flag 'config' set to %q", configPath)

	ownerID := ""
	clusterID, err := cmd.Flags().GetString("cluster-id")
	if err != nil {
		return verifyFlags{}, fmt.Errorf("parsing cluster-id argument: %w", err)
	}
	c.log.Debugf("Flag 'cluster-id' set to %q", clusterID)

	endpoint, err := cmd.Flags().GetString("node-endpoint")
	if err != nil {
		return verifyFlags{}, fmt.Errorf("parsing node-endpoint argument: %w", err)
	}
	c.log.Debugf("Flag 'node-endpoint' set to %q", endpoint)

	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return verifyFlags{}, fmt.Errorf("parsing force argument: %w", err)
	}
	c.log.Debugf("Flag 'force' set to %t", force)

	raw, err := cmd.Flags().GetBool("raw")
	if err != nil {
		return verifyFlags{}, fmt.Errorf("parsing raw argument: %w", err)
	}
	c.log.Debugf("Flag 'raw' set to %t", force)

	var idFile clusterid.File
	if err := fileHandler.ReadJSON(constants.ClusterIDsFileName, &idFile); err != nil && !errors.Is(err, afero.ErrFileNotFound) {
		return verifyFlags{}, fmt.Errorf("reading cluster ID file: %w", err)
	}

	// Get empty values from ID file
	emptyEndpoint := endpoint == ""
	emptyIDs := ownerID == "" && clusterID == ""
	if emptyEndpoint || emptyIDs {
		c.log.Debugf("Trying to supplement empty flag values from %q", constants.ClusterIDsFileName)
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
		rawOutput:  raw,
		force:      force,
	}, nil
}

type verifyFlags struct {
	endpoint   string
	ownerID    string
	clusterID  string
	configPath string
	maaURL     string
	rawOutput  bool
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

// printAttestationDoc prints the attestation document to the command output.
func (c *verifyCmd) printAttestationDoc(cmd *cobra.Command, rawDoc string, rawOutput bool) error {
	if rawOutput {
		cmd.Println(rawDoc)
		return nil
	}

	var doc attestationDoc
	if err := json.Unmarshal([]byte(rawDoc), &doc); err != nil {
		return fmt.Errorf("unmarshal attestation document: %w", err)
	}

	pcrs256 := doc.Attestation.Quotes[len(doc.Attestation.Quotes)-1].Pcrs.Pcrs
	for i, pcr := range pcrs256 {
		cmd.Printf("\tPCR %d:\t%x", i, pcr)
	}

	instanceInfoString, err := base64.StdEncoding.DecodeString(doc.InstanceInfo)
	if err != nil {
		return fmt.Errorf("decode instance info: %w", err)
	}

	var instanceInfo azureInstanceInfo
	if err := json.Unmarshal(instanceInfoString, &instanceInfo); err != nil {
		return fmt.Errorf("unmarshal instance info: %w", err)
	}

	if err := c.parseAndPrintCert(cmd, "VCEK certificate", instanceInfo.Vcek); err != nil {
		return fmt.Errorf("print VCEK certificate: %w", err)
	}
	if err := c.parseAndPrintCert(cmd, "Certificate chain", instanceInfo.CertChain); err != nil {
		return fmt.Errorf("print certificate chain: %w", err)
	}

	return nil
}

// parseAndPrintCerts parses the base64 encoded PEM certificates and prints their details to the command output.
func (c *verifyCmd) parseAndPrintCert(cmd *cobra.Command, certTypeName string, encCertString string) error {
	certBytes, err := base64.StdEncoding.DecodeString(encCertString)
	if err != nil {
		return fmt.Errorf("decode %s: %w", certTypeName, err)
	}

	i := 1
	block, rest := pem.Decode(certBytes)
	for block != nil {
		if block.Type != "CERTIFICATE" {
			return fmt.Errorf("parse %s: expected PEM block type 'CERTIFICATE', got '%s'", certTypeName, block.Type)
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return fmt.Errorf("parse %s: %w", certTypeName, err)
		}

		cmd.Printf("\t%s (%d):\n", certTypeName, i)
		cmd.Printf("\t\tSerial Number: %s\n", cert.SerialNumber)
		cmd.Printf("\t\tSubject: %s\n", cert.Subject)
		cmd.Printf("\t\tIssuer: %s\n", cert.Issuer)
		cmd.Printf("\t\tNot Before: %s\n", cert.NotBefore)
		cmd.Printf("\t\tNot After: %s\n", cert.NotAfter)
		cmd.Printf("\t\tSignature Algorithm: %s\n", cert.SignatureAlgorithm)
		cmd.Printf("\t\tPublic Key Algorithm: %s\n", cert.PublicKeyAlgorithm)

		block, rest = pem.Decode(rest)
		i++
	}

	return nil
}

type attestationDoc struct {
	Attestation struct {
		AkPub  string `json:"ak_pub"`
		Quotes []struct {
			Quote  string `json:"quote"`
			RawSig string `json:"raw_sig"`
			Pcrs   struct {
				Hash int               `json:"hash"`
				Pcrs map[string]string `json:"pcrs"`
			} `json:"pcrs"`
		} `json:"quotes"`
		EventLog       string      `json:"event_log"`
		TeeAttestation interface{} `json:"TeeAttestation"`
	} `json:"Attestation"`
	InstanceInfo string `json:"InstanceInfo"`
	UserData     string `json:"UserData"`
}

type azureInstanceInfo struct {
	Vcek              string `json:"Vcek"`
	CertChain         string `json:"CertChain"`
	AttestationReport string `json:"AttestationReport"`
	RuntimeData       string `json:"RuntimeData"`
}

type constellationVerifier struct {
	dialer grpcInsecureDialer
	log    debugLog
}

// Verify retrieves an attestation statement from the Constellation and verifies it using the validator.
func (v *constellationVerifier) Verify(
	ctx context.Context, endpoint string, req *verifyproto.GetAttestationRequest, validator atls.Validator,
) (string, error) {
	v.log.Debugf("Dialing endpoint: %q", endpoint)
	conn, err := v.dialer.DialInsecure(ctx, endpoint)
	if err != nil {
		return "", fmt.Errorf("dialing init server: %w", err)
	}
	defer conn.Close()

	client := verifyproto.NewAPIClient(conn)

	v.log.Debugf("Sending attestation request")
	resp, err := client.GetAttestation(ctx, req)
	if err != nil {
		return "", fmt.Errorf("getting attestation: %w", err)
	}

	v.log.Debugf("Verifying attestation")
	signedData, err := validator.Validate(ctx, resp.Attestation, req.Nonce)
	if err != nil {
		return "", fmt.Errorf("validating attestation: %w", err)
	}

	if !bytes.Equal(signedData, []byte(constants.ConstellationVerifyServiceUserData)) {
		return "", errors.New("signed data in attestation does not match expected user data")
	}

	return string(resp.Attestation), nil
}

type verifyClient interface {
	Verify(ctx context.Context, endpoint string, req *verifyproto.GetAttestationRequest, validator atls.Validator) (string, error)
}

type grpcInsecureDialer interface {
	DialInsecure(ctx context.Context, endpoint string) (conn *grpc.ClientConn, err error)
}
