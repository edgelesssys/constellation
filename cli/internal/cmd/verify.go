/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"

	tpmProto "github.com/google/go-tpm-tools/proto/tpm"

	"github.com/edgelesssys/constellation/v2/internal/api/attestationconfigapi"
	"github.com/edgelesssys/constellation/v2/internal/atls"
	"github.com/edgelesssys/constellation/v2/internal/attestation/choose"
	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/attestation/snp"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/attestation/vtpm"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/constellation/state"
	"github.com/edgelesssys/constellation/v2/internal/crypto"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/grpc/dialer"
	"github.com/edgelesssys/constellation/v2/internal/verify"
	"github.com/edgelesssys/constellation/v2/verify/verifyproto"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/grpc"
)

// NewVerifyCmd returns a new cobra.Command for the verify command.
func NewVerifyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verify",
		Short: "Verify the confidential properties of a Constellation cluster",
		Long: "Verify the confidential properties of a Constellation cluster.\n" +
			"If arguments aren't specified, values are read from `" + constants.StateFilename + "`.",
		Args: cobra.ExactArgs(0),
		RunE: runVerify,
	}
	cmd.Flags().String("cluster-id", "", "expected cluster identifier")
	cmd.Flags().StringP("output", "o", "", "print the attestation document in the output format {json|raw}")
	cmd.Flags().StringP("node-endpoint", "e", "", "endpoint of the node to verify, passed as HOST[:PORT]")
	return cmd
}

type verifyFlags struct {
	rootFlags
	endpoint  string
	ownerID   string
	clusterID string
	output    string
}

func (f *verifyFlags) parse(flags *pflag.FlagSet) error {
	if err := f.rootFlags.parse(flags); err != nil {
		return err
	}

	var err error
	f.output, err = flags.GetString("output")
	if err != nil {
		return fmt.Errorf("getting 'output' flag: %w", err)
	}
	f.endpoint, err = flags.GetString("node-endpoint")
	if err != nil {
		return fmt.Errorf("getting 'node-endpoint' flag: %w", err)
	}
	f.clusterID, err = flags.GetString("cluster-id")
	if err != nil {
		return fmt.Errorf("getting 'cluster-id' flag: %w", err)
	}
	return nil
}

type verifyCmd struct {
	fileHandler file.Handler
	flags       verifyFlags
	log         debugLog
}

func runVerify(cmd *cobra.Command, _ []string) error {
	log, err := newCLILogger(cmd)
	if err != nil {
		return fmt.Errorf("creating logger: %w", err)
	}

	fileHandler := file.NewHandler(afero.NewOsFs())
	verifyClient := &constellationVerifier{
		dialer: dialer.New(nil, nil, &net.Dialer{}),
		log:    log,
	}
	formatterFactory := func(output string, attestation variant.Variant, log debugLog) (attestationDocFormatter, error) {
		if output == "json" && (!attestation.Equal(variant.AzureSEVSNP{}) && !attestation.Equal(variant.AWSSEVSNP{})) {
			return nil, errors.New("json output is only supported for Azure SEV-SNP and AWS SEV-SNP")
		}
		switch output {
		case "json":
			return &jsonAttestationDocFormatter{log}, nil
		case "raw":
			return &rawAttestationDocFormatter{log}, nil
		case "":
			return &defaultAttestationDocFormatter{log}, nil
		default:
			return nil, fmt.Errorf("invalid output value for formatter: %s", output)
		}
	}
	v := &verifyCmd{
		fileHandler: fileHandler,
		log:         log,
	}
	if err := v.flags.parse(cmd.Flags()); err != nil {
		return err
	}
	v.log.Debug("Using flags: %+v", v.flags)
	fetcher := attestationconfigapi.NewFetcher()
	return v.verify(cmd, verifyClient, formatterFactory, fetcher)
}

type formatterFactory func(output string, attestation variant.Variant, log debugLog) (attestationDocFormatter, error)

func (c *verifyCmd) verify(cmd *cobra.Command, verifyClient verifyClient, factory formatterFactory, configFetcher attestationconfigapi.Fetcher) error {
	c.log.Debug("Loading configuration file from %q", c.flags.pathPrefixer.PrefixPrintablePath(constants.ConfigFilename))
	conf, err := config.New(c.fileHandler, constants.ConfigFilename, configFetcher, c.flags.force)
	var configValidationErr *config.ValidationError
	if errors.As(err, &configValidationErr) {
		cmd.PrintErrln(configValidationErr.LongMessage())
	}
	if err != nil {
		return fmt.Errorf("loading config file: %w", err)
	}

	stateFile, err := state.ReadFromFile(c.fileHandler, constants.StateFilename)
	if err != nil {
		return fmt.Errorf("reading state file: %w", err)
	}
	if err := stateFile.Validate(state.PostInit, conf.GetAttestationConfig().GetVariant()); err != nil {
		return fmt.Errorf("validating state file: %w", err)
	}

	ownerID, clusterID, err := c.validateIDFlags(cmd, stateFile)
	if err != nil {
		return err
	}
	endpoint, err := c.validateEndpointFlag(cmd, stateFile)
	if err != nil {
		return err
	}

	var maaURL string
	if stateFile.Infrastructure.Azure != nil {
		maaURL = stateFile.Infrastructure.Azure.AttestationURL
	}
	conf.UpdateMAAURL(maaURL)

	c.log.Debug("Updating expected PCRs")
	attConfig := conf.GetAttestationConfig()
	if err := updateInitMeasurements(attConfig, ownerID, clusterID); err != nil {
		return fmt.Errorf("updating expected PCRs: %w", err)
	}

	c.log.Debug("Creating aTLS Validator for %s", conf.GetAttestationConfig().GetVariant())
	validator, err := choose.Validator(attConfig, warnLogger{cmd: cmd, log: c.log})
	if err != nil {
		return fmt.Errorf("creating aTLS validator: %w", err)
	}

	nonce, err := crypto.GenerateRandomBytes(32)
	if err != nil {
		return fmt.Errorf("generating random nonce: %w", err)
	}
	c.log.Debug("Generated random nonce: %x", nonce)

	rawAttestationDoc, err := verifyClient.Verify(
		cmd.Context(),
		endpoint,
		&verifyproto.GetAttestationRequest{
			Nonce: nonce,
		},
		validator,
	)
	if err != nil {
		return fmt.Errorf("verifying: %w", err)
	}

	// certificates are only available for Azure SEV-SNP and AWS SEV-SNP
	formatter, err := factory(c.flags.output, conf.GetAttestationConfig().GetVariant(), c.log)
	if err != nil {
		return fmt.Errorf("creating formatter: %w", err)
	}
	attDocOutput, err := formatter.format(
		cmd.Context(),
		rawAttestationDoc,
		(!attConfig.GetVariant().Equal(variant.AzureSEVSNP{}) && !attConfig.GetVariant().Equal(variant.AWSSEVSNP{})),
		attConfig,
	)
	if err != nil {
		return fmt.Errorf("printing attestation document: %w", err)
	}
	cmd.Println(attDocOutput)
	cmd.PrintErrln("Verification OK")

	return nil
}

func (c *verifyCmd) validateIDFlags(cmd *cobra.Command, stateFile *state.State) (ownerID, clusterID string, err error) {
	ownerID, clusterID = c.flags.ownerID, c.flags.clusterID
	if c.flags.clusterID == "" {
		cmd.PrintErrf("Using ID from %q. Specify --cluster-id to override this.\n", c.flags.pathPrefixer.PrefixPrintablePath(constants.StateFilename))
		clusterID = stateFile.ClusterValues.ClusterID
	}
	if ownerID == "" {
		// We don't want to print warnings until this is implemented again
		// cmd.PrintErrf("Using ID from %q. Specify --owner-id to override this.\n", c.flags.pathPrefixer.PrefixPrintablePath(constants.StateFilename))
		ownerID = stateFile.ClusterValues.OwnerID
	}

	// Validate
	if ownerID == "" && clusterID == "" {
		return "", "", errors.New("cluster-id not provided to verify the cluster")
	}

	return ownerID, clusterID, nil
}

func (c *verifyCmd) validateEndpointFlag(cmd *cobra.Command, stateFile *state.State) (string, error) {
	endpoint := c.flags.endpoint
	if endpoint == "" {
		cmd.PrintErrf("Using endpoint from %q. Specify --node-endpoint to override this.\n", c.flags.pathPrefixer.PrefixPrintablePath(constants.StateFilename))
		endpoint = stateFile.Infrastructure.ClusterEndpoint
	}
	endpoint, err := addPortIfMissing(endpoint, constants.VerifyServiceNodePortGRPC)
	if err != nil {
		return "", fmt.Errorf("validating endpoint argument: %w", err)
	}
	return endpoint, nil
}

// an attestationDocFormatter formats the attestation document.
type attestationDocFormatter interface {
	// format returns the raw or formatted attestation doc depending on the rawOutput argument.
	format(ctx context.Context, docString string, PCRsOnly bool, attestationCfg config.AttestationCfg) (string, error)
}

type jsonAttestationDocFormatter struct {
	log debugLog
}

// format returns the json formatted attestation doc.
func (f *jsonAttestationDocFormatter) format(ctx context.Context, docString string, _ bool,
	attestationCfg config.AttestationCfg,
) (string, error) {
	var doc attestationDoc
	if err := json.Unmarshal([]byte(docString), &doc); err != nil {
		return "", fmt.Errorf("unmarshal attestation document: %w", err)
	}

	instanceInfo, err := extractInstanceInfo(doc)
	if err != nil {
		return "", fmt.Errorf("unmarshalling instance info: %w", err)
	}
	report, err := verify.NewReport(ctx, instanceInfo, attestationCfg, f.log)
	if err != nil {
		return "", fmt.Errorf("parsing SNP report: %w", err)
	}

	jsonBytes, err := json.Marshal(report)

	return string(jsonBytes), err
}

type rawAttestationDocFormatter struct {
	log debugLog
}

// format returns the raw attestation doc.
func (f *rawAttestationDocFormatter) format(_ context.Context, docString string, _ bool,
	_ config.AttestationCfg,
) (string, error) {
	b := &strings.Builder{}
	b.WriteString("Attestation Document:\n")
	b.WriteString(fmt.Sprintf("%s\n", docString))
	return b.String(), nil
}

type defaultAttestationDocFormatter struct {
	log debugLog
}

// format returns the formatted attestation doc.
func (f *defaultAttestationDocFormatter) format(ctx context.Context, docString string, PCRsOnly bool,
	attestationCfg config.AttestationCfg,
) (string, error) {
	b := &strings.Builder{}
	b.WriteString("Attestation Document:\n")

	var doc attestationDoc
	if err := json.Unmarshal([]byte(docString), &doc); err != nil {
		return "", fmt.Errorf("unmarshal attestation document: %w", err)
	}

	if err := f.parseQuotes(b, doc.Attestation.Quotes, attestationCfg.GetMeasurements()); err != nil {
		return "", fmt.Errorf("parse quote: %w", err)
	}
	if PCRsOnly {
		return b.String(), nil
	}

	instanceInfo, err := extractInstanceInfo(doc)
	if err != nil {
		return "", fmt.Errorf("unmarshalling instance info: %w", err)
	}

	report, err := verify.NewReport(ctx, instanceInfo, attestationCfg, f.log)
	if err != nil {
		return "", fmt.Errorf("parsing SNP report: %w", err)
	}

	return report.FormatString(b)
}

// parseQuotes parses the base64-encoded quotes and writes their details to the output builder.
func (f *defaultAttestationDocFormatter) parseQuotes(b *strings.Builder, quotes []*tpmProto.Quote, expectedPCRs measurements.M) error {
	writeIndentfln(b, 1, "Quote:")

	var pcrNumbers []uint32
	for pcrNum := range expectedPCRs {
		pcrNumbers = append(pcrNumbers, pcrNum)
	}
	sort.Slice(pcrNumbers, func(i, j int) bool { return pcrNumbers[i] < pcrNumbers[j] })

	for _, pcrNum := range pcrNumbers {
		expectedPCR := expectedPCRs[pcrNum]
		pcrIdx, err := vtpm.GetSHA256QuoteIndex(quotes)
		if err != nil {
			return fmt.Errorf("get SHA256 quote index: %w", err)
		}

		actualPCR, ok := quotes[pcrIdx].Pcrs.Pcrs[pcrNum]
		if !ok {
			return fmt.Errorf("PCR %d not found in quote", pcrNum)
		}
		writeIndentfln(b, 2, "PCR %d (Strict: %t):", pcrNum, !expectedPCR.ValidationOpt)
		writeIndentfln(b, 3, "Expected:\t%x", expectedPCR.Expected)
		writeIndentfln(b, 3, "Actual:\t\t%x", actualPCR)
	}
	return nil
}

// attestationDoc is the attestation document returned by the verifier.
type attestationDoc struct {
	Attestation struct {
		AkPub          string            `json:"ak_pub"`
		Quotes         []*tpmProto.Quote `json:"quotes"`
		EventLog       string            `json:"event_log"`
		TeeAttestation interface{}       `json:"TeeAttestation"`
	} `json:"Attestation"`
	InstanceInfo string `json:"InstanceInfo"`
	UserData     string `json:"UserData"`
}

type constellationVerifier struct {
	dialer grpcInsecureDialer
	log    debugLog
}

// Verify retrieves an attestation statement from the Constellation and verifies it using the validator.
func (v *constellationVerifier) Verify(
	ctx context.Context, endpoint string, req *verifyproto.GetAttestationRequest, validator atls.Validator,
) (string, error) {
	v.log.Debug("Dialing endpoint: %q", endpoint)
	conn, err := v.dialer.DialInsecure(ctx, endpoint)
	if err != nil {
		return "", fmt.Errorf("dialing init server: %w", err)
	}
	defer conn.Close()

	client := verifyproto.NewAPIClient(conn)

	v.log.Debug("Sending attestation request")
	resp, err := client.GetAttestation(ctx, req)
	if err != nil {
		return "", fmt.Errorf("getting attestation: %w", err)
	}

	v.log.Debug("Verifying attestation")
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

// writeIndentfln writes a formatted string to the builder with the given indentation level
// and a newline at the end.
func writeIndentfln(b *strings.Builder, indentLvl int, format string, args ...any) {
	for i := 0; i < indentLvl; i++ {
		b.WriteByte('\t')
	}
	b.WriteString(fmt.Sprintf(format+"\n", args...))
}

func extractInstanceInfo(doc attestationDoc) (snp.InstanceInfo, error) {
	instanceInfoString, err := base64.StdEncoding.DecodeString(doc.InstanceInfo)
	if err != nil {
		return snp.InstanceInfo{}, fmt.Errorf("decode instance info: %w", err)
	}

	var instanceInfo snp.InstanceInfo
	if err := json.Unmarshal(instanceInfoString, &instanceInfo); err != nil {
		return snp.InstanceInfo{}, fmt.Errorf("unmarshal instance info: %w", err)
	}
	return instanceInfo, nil
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

// UpdateInitMeasurements sets the owner and cluster measurement values in the attestation config depending on the
// attestation variant.
func updateInitMeasurements(config config.AttestationCfg, ownerID, clusterID string) error {
	m := config.GetMeasurements()

	switch config.GetVariant() {
	case variant.AWSNitroTPM{}, variant.AWSSEVSNP{},
		variant.AzureTrustedLaunch{}, variant.AzureSEVSNP{}, variant.AzureTDX{}, // AzureTDX also uses a vTPM for measurements
		variant.GCPSEVES{},
		variant.QEMUVTPM{}:
		if err := updateMeasurementTPM(m, uint32(measurements.PCRIndexOwnerID), ownerID); err != nil {
			return err
		}
		return updateMeasurementTPM(m, uint32(measurements.PCRIndexClusterID), clusterID)
	case variant.QEMUTDX{}:
		// Measuring ownerID is currently not implemented for Constellation
		// Since adding support for measuring ownerID to TDX would require additional code changes,
		// the current implementation does not support it, but can be changed if we decide to add support in the future
		return updateMeasurementTDX(m, uint32(measurements.TDXIndexClusterID), clusterID)
	default:
		return errors.New("selecting attestation variant: unknown attestation variant")
	}
}

// updateMeasurementTDX updates the TDX measurement value in the attestation config for the given measurement index.
func updateMeasurementTDX(m measurements.M, measurementIdx uint32, encoded string) error {
	if encoded == "" {
		delete(m, measurementIdx)
		return nil
	}
	decoded, err := decodeMeasurement(encoded)
	if err != nil {
		return err
	}

	// new_measurement_value := hash(old_measurement_value || data_to_extend)
	// Since we use the DG.MR.RTMR.EXTEND call to extend the register, data_to_extend is the hash of our input
	hashedInput := sha512.Sum384(decoded)
	oldExpected := m[measurementIdx].Expected
	expectedMeasurementSum := sha512.Sum384(append(oldExpected[:], hashedInput[:]...))
	m[measurementIdx] = measurements.Measurement{
		Expected:      expectedMeasurementSum[:],
		ValidationOpt: m[measurementIdx].ValidationOpt,
	}
	return nil
}

// updateMeasurementTPM updates the TPM measurement value in the attestation config for the given measurement index.
func updateMeasurementTPM(m measurements.M, measurementIdx uint32, encoded string) error {
	if encoded == "" {
		delete(m, measurementIdx)
		return nil
	}
	decoded, err := decodeMeasurement(encoded)
	if err != nil {
		return err
	}

	// new_pcr_value := hash(old_pcr_value || data_to_extend)
	// Since we use the TPM2_PCR_Event call to extend the PCR, data_to_extend is the hash of our input
	hashedInput := sha256.Sum256(decoded)
	oldExpected := m[measurementIdx].Expected
	expectedMeasurement := sha256.Sum256(append(oldExpected[:], hashedInput[:]...))
	m[measurementIdx] = measurements.Measurement{
		Expected:      expectedMeasurement[:],
		ValidationOpt: m[measurementIdx].ValidationOpt,
	}
	return nil
}

// decodeMeasurement is a utility function that decodes the given string as hex or base64.
func decodeMeasurement(encoded string) ([]byte, error) {
	decoded, err := hex.DecodeString(encoded)
	if err != nil {
		hexErr := err
		decoded, err = base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return nil, fmt.Errorf("input [%s] could neither be hex decoded (%w) nor base64 decoded (%w)", encoded, hexErr, err)
		}
	}
	return decoded, nil
}
