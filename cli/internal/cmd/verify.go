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
	"io"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	tpmProto "github.com/google/go-tpm-tools/proto/tpm"

	"github.com/edgelesssys/constellation/v2/cli/internal/cloudcmd"
	"github.com/edgelesssys/constellation/v2/cli/internal/clusterid"
	"github.com/edgelesssys/constellation/v2/cli/internal/cmd/pathprefix"
	"github.com/edgelesssys/constellation/v2/internal/api/attestationconfigapi"
	"github.com/edgelesssys/constellation/v2/internal/atls"
	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/attestation/vtpm"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/crypto"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/grpc/dialer"
	"github.com/edgelesssys/constellation/v2/internal/verify"
	"github.com/edgelesssys/constellation/v2/verify/verifyproto"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/go-sev-guest/abi"
	"github.com/google/go-sev-guest/kds"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

// NewVerifyCmd returns a new cobra.Command for the verify command.
func NewVerifyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verify",
		Short: "Verify the confidential properties of a Constellation cluster",
		Long: "Verify the confidential properties of a Constellation cluster.\n" +
			"If arguments aren't specified, values are read from `" + constants.ClusterIDsFilename + "`.",
		Args: cobra.ExactArgs(0),
		RunE: runVerify,
	}
	cmd.Flags().String("cluster-id", "", "expected cluster identifier")
	cmd.Flags().Bool("raw", false, "print raw attestation document")
	cmd.Flags().Bool("json", false, "print the attestation document as parsed json")
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
	formatterFactory := func(jsonOutput bool) attestationDocFormatter {
		if jsonOutput {
			return &jsonAttestationDocFormatter{log}
		}
		return &attestationDocFormatterImpl{log}
	}
	// TODO(elchead): unify the formatters to share the parsing logic
	v := &verifyCmd{log: log}
	fetcher := attestationconfigapi.NewFetcher()
	return v.verify(cmd, fileHandler, verifyClient, formatterFactory, fetcher)
}

type formatterFactory func(jsonOutput bool) attestationDocFormatter

func (c *verifyCmd) verify(cmd *cobra.Command, fileHandler file.Handler, verifyClient verifyClient, formatter formatterFactory, configFetcher attestationconfigapi.Fetcher) error {
	flags, err := c.parseVerifyFlags(cmd, fileHandler)
	if err != nil {
		return fmt.Errorf("parsing flags: %w", err)
	}
	c.log.Debugf("Using flags: %+v", flags)

	c.log.Debugf("Loading configuration file from %q", flags.pf.PrefixPrintablePath(constants.ConfigFilename))
	conf, err := config.New(fileHandler, constants.ConfigFilename, configFetcher, flags.force)
	var configValidationErr *config.ValidationError
	if errors.As(err, &configValidationErr) {
		cmd.PrintErrln(configValidationErr.LongMessage())
	}
	if err != nil {
		return fmt.Errorf("loading config file: %w", err)
	}

	conf.UpdateMAAURL(flags.maaURL)
	c.log.Debugf("Updating expected PCRs")
	attConfig := conf.GetAttestationConfig()
	if err := cloudcmd.UpdateInitMeasurements(attConfig, flags.ownerID, flags.clusterID); err != nil {
		return fmt.Errorf("updating expected PCRs: %w", err)
	}

	c.log.Debugf("Creating aTLS Validator for %s", conf.GetAttestationConfig().GetVariant())
	validator, err := cloudcmd.NewValidator(cmd, attConfig, c.log)
	if err != nil {
		return fmt.Errorf("creating aTLS validator: %w", err)
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
		validator,
	)
	if err != nil {
		return fmt.Errorf("verifying: %w", err)
	}

	// certificates are only available for Azure
	attDocOutput, err := formatter(flags.jsonOutput).format(
		cmd.Context(),
		rawAttestationDoc,
		conf.Provider.Azure == nil,
		flags.rawOutput,
		attConfig.GetMeasurements(),
		flags.maaURL,
	)
	if err != nil {
		return fmt.Errorf("printing attestation document: %w", err)
	}
	cmd.Println(attDocOutput)
	cmd.Println("Verification OK")

	return nil
}

func (c *verifyCmd) parseVerifyFlags(cmd *cobra.Command, fileHandler file.Handler) (verifyFlags, error) {
	workDir, err := cmd.Flags().GetString("workspace")
	if err != nil {
		return verifyFlags{}, fmt.Errorf("parsing config path argument: %w", err)
	}
	c.log.Debugf("Flag 'workspace' set to %q", workDir)
	pf := pathprefix.New(workDir)

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
	json, err := cmd.Flags().GetBool("json")
	if err != nil {
		return verifyFlags{}, fmt.Errorf("parsing raw argument: %w", err)
	}
	c.log.Debugf("Flag 'json' set to %t", force)

	var idFile clusterid.File
	if err := fileHandler.ReadJSON(constants.ClusterIDsFilename, &idFile); err != nil && !errors.Is(err, afero.ErrFileNotFound) {
		return verifyFlags{}, fmt.Errorf("reading cluster ID file: %w", err)
	}

	// Get empty values from ID file
	emptyEndpoint := endpoint == ""
	emptyIDs := ownerID == "" && clusterID == ""
	if emptyEndpoint || emptyIDs {
		c.log.Debugf("Trying to supplement empty flag values from %q", pf.PrefixPrintablePath(constants.ClusterIDsFilename))
		if emptyEndpoint {
			cmd.Printf("Using endpoint from %q. Specify --node-endpoint to override this.\n", pf.PrefixPrintablePath(constants.ClusterIDsFilename))
			endpoint = idFile.IP
		}
		if emptyIDs {
			cmd.Printf("Using ID from %q. Specify --cluster-id to override this.\n", pf.PrefixPrintablePath(constants.ClusterIDsFilename))
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

	if raw && json {
		return verifyFlags{}, errors.New("cannot specify both --raw and --json")
	}
	return verifyFlags{
		endpoint:   endpoint,
		pf:         pf,
		ownerID:    ownerID,
		clusterID:  clusterID,
		maaURL:     idFile.AttestationURL,
		rawOutput:  raw,
		jsonOutput: json,
		force:      force,
	}, nil
}

type verifyFlags struct {
	endpoint   string
	ownerID    string
	clusterID  string
	maaURL     string
	rawOutput  bool
	jsonOutput bool
	force      bool
	pf         pathprefix.PathPrefixer
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

// an attestationDocFormatter formats the attestation document.
// TODO(elchead): refactor the interface to be more generic (e.g. no rawOutput argument).
type attestationDocFormatter interface {
	// format returns the raw or formatted attestation doc depending on the rawOutput argument.
	format(ctx context.Context, docString string, PCRsOnly bool, rawOutput bool, expectedPCRs measurements.M,
		attestationServiceURL string) (string, error)
}

type jsonAttestationDocFormatter struct {
	log debugLog
}

// format returns the raw or formatted attestation doc depending on the rawOutput argument.
func (f *jsonAttestationDocFormatter) format(ctx context.Context, docString string, _ bool,
	_ bool, _ measurements.M, attestationServiceURL string,
) (string, error) {
	var doc attestationDoc
	if err := json.Unmarshal([]byte(docString), &doc); err != nil {
		return "", fmt.Errorf("unmarshal attestation document: %w", err)
	}

	instanceInfoString, err := base64.StdEncoding.DecodeString(doc.InstanceInfo)
	if err != nil {
		return "", fmt.Errorf("decode instance info: %w", err)
	}

	var instanceInfo azureInstanceInfo
	if err := json.Unmarshal(instanceInfoString, &instanceInfo); err != nil {
		return "", fmt.Errorf("unmarshal instance info: %w", err)
	}

	// TODO(elchead): omit quotes?
	snpReport, err := newSNPReport(instanceInfo.AttestationReport)
	if err != nil {
		return "", fmt.Errorf("parsing SNP report: %w", err)
	}

	vcek, err := newCertificates("VCEK certificate", instanceInfo.Vcek, f.log)
	if err != nil {
		return "", fmt.Errorf("parsing VCEK certificate: %w", err)
	}
	certChain, err := newCertificates("Certificate chain", instanceInfo.CertChain, f.log)
	if err != nil {
		return "", fmt.Errorf("parsing certificate chain: %w", err)
	}
	maaToken, err := newMAAToken(ctx, instanceInfo.MAAToken, attestationServiceURL)
	if err != nil {
		return "", fmt.Errorf("parsing MAA token: %w", err)
	}

	report := verify.Report{
		SNPReport: snpReport,
		VCEK:      vcek,
		CertChain: certChain,
		MAAToken:  maaToken,
	}
	jsonBytes, err := json.Marshal(report)

	return string(jsonBytes), err
}

type attestationDocFormatterImpl struct {
	log debugLog
}

// format returns the raw or formatted attestation doc depending on the rawOutput argument.
func (f *attestationDocFormatterImpl) format(ctx context.Context, docString string, PCRsOnly bool,
	rawOutput bool, expectedPCRs measurements.M, attestationServiceURL string,
) (string, error) {
	b := &strings.Builder{}
	b.WriteString("Attestation Document:\n")
	if rawOutput {
		b.WriteString(fmt.Sprintf("%s\n", docString))
		return b.String(), nil
	}

	var doc attestationDoc
	if err := json.Unmarshal([]byte(docString), &doc); err != nil {
		return "", fmt.Errorf("unmarshal attestation document: %w", err)
	}

	if err := f.parseQuotes(b, doc.Attestation.Quotes, expectedPCRs); err != nil {
		return "", fmt.Errorf("parse quote: %w", err)
	}
	if PCRsOnly {
		return b.String(), nil
	}

	instanceInfoString, err := base64.StdEncoding.DecodeString(doc.InstanceInfo)
	if err != nil {
		return "", fmt.Errorf("decode instance info: %w", err)
	}

	var instanceInfo azureInstanceInfo
	if err := json.Unmarshal(instanceInfoString, &instanceInfo); err != nil {
		return "", fmt.Errorf("unmarshal instance info: %w", err)
	}

	if err := f.parseCerts(b, "VCEK certificate", instanceInfo.Vcek); err != nil {
		return "", fmt.Errorf("print VCEK certificate: %w", err)
	}
	if err := f.parseCerts(b, "Certificate chain", instanceInfo.CertChain); err != nil {
		return "", fmt.Errorf("print certificate chain: %w", err)
	}
	if err := f.parseSNPReport(b, instanceInfo.AttestationReport); err != nil {
		return "", fmt.Errorf("print SNP report: %w", err)
	}
	if err := parseMAAToken(ctx, b, instanceInfo.MAAToken, attestationServiceURL); err != nil {
		return "", fmt.Errorf("print MAA token: %w", err)
	}

	return b.String(), nil
}

// parseCerts parses the PEM certificates and writes their details to the output builder.
func (f *attestationDocFormatterImpl) parseCerts(b *strings.Builder, certTypeName string, cert []byte) error {
	newlinesTrimmed := strings.TrimSpace(string(cert))
	formattedCert := strings.ReplaceAll(newlinesTrimmed, "\n", "\n\t\t") + "\n"
	b.WriteString(fmt.Sprintf("\tRaw %s:\n\t\t%s", certTypeName, formattedCert))

	f.log.Debugf("Decoding PEM certificate: %s", certTypeName)
	i := 1
	var rest []byte
	var block *pem.Block
	for block, rest = pem.Decode([]byte(newlinesTrimmed)); block != nil; block, rest = pem.Decode(rest) {
		f.log.Debugf("Parsing PEM block: %d", i)
		if block.Type != "CERTIFICATE" {
			return fmt.Errorf("parse %s: expected PEM block type 'CERTIFICATE', got '%s'", certTypeName, block.Type)
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return fmt.Errorf("parse %s: %w", certTypeName, err)
		}

		writeIndentfln(b, 1, "%s (%d):", certTypeName, i)
		writeIndentfln(b, 2, "Serial Number: %s", cert.SerialNumber)
		writeIndentfln(b, 2, "Subject: %s", cert.Subject)
		writeIndentfln(b, 2, "Issuer: %s", cert.Issuer)
		writeIndentfln(b, 2, "Not Before: %s", cert.NotBefore)
		writeIndentfln(b, 2, "Not After: %s", cert.NotAfter)
		writeIndentfln(b, 2, "Signature Algorithm: %s", cert.SignatureAlgorithm)
		writeIndentfln(b, 2, "Public Key Algorithm: %s", cert.PublicKeyAlgorithm)

		if certTypeName == "VCEK certificate" {
			// Extensions documented in Table 8 and Table 9 of
			// https://www.amd.com/system/files/TechDocs/57230.pdf
			vcekExts, err := kds.VcekCertificateExtensions(cert)
			if err != nil {
				return fmt.Errorf("parsing VCEK certificate extensions: %w", err)
			}

			writeIndentfln(b, 2, "Struct version: %d", vcekExts.StructVersion)
			writeIndentfln(b, 2, "Product name: %s", vcekExts.ProductName)
			tcb := kds.DecomposeTCBVersion(vcekExts.TCBVersion)
			writeIndentfln(b, 2, "Secure Processor bootloader SVN: %d", tcb.BlSpl)
			writeIndentfln(b, 2, "Secure Processor operating system SVN: %d", tcb.TeeSpl)
			writeIndentfln(b, 2, "SVN 4 (reserved): %d", tcb.Spl4)
			writeIndentfln(b, 2, "SVN 5 (reserved): %d", tcb.Spl5)
			writeIndentfln(b, 2, "SVN 6 (reserved): %d", tcb.Spl6)
			writeIndentfln(b, 2, "SVN 7 (reserved): %d", tcb.Spl7)
			writeIndentfln(b, 2, "SEV-SNP firmware SVN: %d", tcb.SnpSpl)
			writeIndentfln(b, 2, "Microcode SVN: %d", tcb.UcodeSpl)
			writeIndentfln(b, 2, "Hardware ID: %x", vcekExts.HWID)
		}

		i++
	}

	if i == 1 {
		return fmt.Errorf("parse %s: no PEM blocks found", certTypeName)
	}
	if len(rest) != 0 {
		return fmt.Errorf("parse %s: remaining PEM block is not a valid certificate: %s", certTypeName, rest)
	}

	return nil
}

// parseQuotes parses the base64-encoded quotes and writes their details to the output builder.
func (f *attestationDocFormatterImpl) parseQuotes(b *strings.Builder, quotes []*tpmProto.Quote, expectedPCRs measurements.M) error {
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

func (f *attestationDocFormatterImpl) parseSNPReport(b *strings.Builder, reportBytes []byte) error {
	report, err := abi.ReportToProto(reportBytes)
	if err != nil {
		return fmt.Errorf("parsing report to proto: %w", err)
	}

	policy, err := abi.ParseSnpPolicy(report.Policy)
	if err != nil {
		return fmt.Errorf("parsing policy: %w", err)
	}

	platformInfo, err := abi.ParseSnpPlatformInfo(report.PlatformInfo)
	if err != nil {
		return fmt.Errorf("parsing platform info: %w", err)
	}

	signature, err := abi.ReportToSignatureDER(reportBytes)
	if err != nil {
		return fmt.Errorf("parsing signature: %w", err)
	}

	signerInfo, err := abi.ParseSignerInfo(report.SignerInfo)
	if err != nil {
		return fmt.Errorf("parsing signer info: %w", err)
	}

	writeTCB := func(tcbVersion uint64) {
		tcb := kds.DecomposeTCBVersion(kds.TCBVersion(tcbVersion))
		writeIndentfln(b, 3, "Secure Processor bootloader SVN: %d", tcb.BlSpl)
		writeIndentfln(b, 3, "Secure Processor operating system SVN: %d", tcb.TeeSpl)
		writeIndentfln(b, 3, "SVN 4 (reserved): %d", tcb.Spl4)
		writeIndentfln(b, 3, "SVN 5 (reserved): %d", tcb.Spl5)
		writeIndentfln(b, 3, "SVN 6 (reserved): %d", tcb.Spl6)
		writeIndentfln(b, 3, "SVN 7 (reserved): %d", tcb.Spl7)
		writeIndentfln(b, 3, "SEV-SNP firmware SVN: %d", tcb.SnpSpl)
		writeIndentfln(b, 3, "Microcode SVN: %d", tcb.UcodeSpl)
	}

	writeIndentfln(b, 1, "SNP Report:")
	writeIndentfln(b, 2, "Version: %d", report.Version)
	writeIndentfln(b, 2, "Guest SVN: %d", report.GuestSvn)
	writeIndentfln(b, 2, "Policy:")
	writeIndentfln(b, 3, "ABI Minor: %d", policy.ABIMinor)
	writeIndentfln(b, 3, "ABI Major: %d", policy.ABIMajor)
	writeIndentfln(b, 3, "Symmetric Multithreading enabled: %t", policy.SMT)
	writeIndentfln(b, 3, "Migration agent enabled: %t", policy.MigrateMA)
	writeIndentfln(b, 3, "Debugging enabled (host decryption of VM): %t", policy.Debug)
	writeIndentfln(b, 3, "Single socket enabled: %t", policy.SingleSocket)
	writeIndentfln(b, 2, "Family ID: %x", report.FamilyId)
	writeIndentfln(b, 2, "Image ID: %x", report.ImageId)
	writeIndentfln(b, 2, "VMPL: %d", report.Vmpl)
	writeIndentfln(b, 2, "Signature Algorithm: %d", report.SignatureAlgo)
	writeIndentfln(b, 2, "Current TCB:")
	writeTCB(report.CurrentTcb)
	writeIndentfln(b, 2, "Platform Info:")
	writeIndentfln(b, 3, "Symmetric Multithreading enabled (SMT): %t", platformInfo.SMTEnabled)
	writeIndentfln(b, 3, "Transparent secure memory encryption (TSME): %t", platformInfo.TSMEEnabled)
	writeIndentfln(b, 2, "Signer Info:")
	writeIndentfln(b, 3, "Author Key Enabled: %t", signerInfo.AuthorKeyEn)
	writeIndentfln(b, 3, "Chip ID Masking: %t", signerInfo.MaskChipKey)
	writeIndentfln(b, 3, "Signing Type: %s", signerInfo.SigningKey)
	writeIndentfln(b, 2, "Report Data: %x", report.ReportData)
	writeIndentfln(b, 2, "Measurement: %x", report.Measurement)
	writeIndentfln(b, 2, "Host Data: %x", report.HostData)
	writeIndentfln(b, 2, "ID Key Digest: %x", report.IdKeyDigest)
	writeIndentfln(b, 2, "Author Key Digest: %x", report.AuthorKeyDigest)
	writeIndentfln(b, 2, "Report ID: %x", report.ReportId)
	writeIndentfln(b, 2, "Report ID MA: %x", report.ReportIdMa)
	writeIndentfln(b, 2, "Reported TCB:")
	writeTCB(report.ReportedTcb)
	writeIndentfln(b, 2, "Chip ID: %x", report.ChipId)
	writeIndentfln(b, 2, "Committed TCB:")
	writeTCB(report.CommittedTcb)
	writeIndentfln(b, 2, "Current Build: %d", report.CurrentBuild)
	writeIndentfln(b, 2, "Current Minor: %d", report.CurrentMinor)
	writeIndentfln(b, 2, "Current Major: %d", report.CurrentMajor)
	writeIndentfln(b, 2, "Committed Build: %d", report.CommittedBuild)
	writeIndentfln(b, 2, "Committed Minor: %d", report.CommittedMinor)
	writeIndentfln(b, 2, "Committed Major: %d", report.CommittedMajor)
	writeIndentfln(b, 2, "Launch TCB:")
	writeTCB(report.LaunchTcb)
	writeIndentfln(b, 2, "Signature (DER):")
	writeIndentfln(b, 3, "%x", signature)

	return nil
}

func parseMAAToken(ctx context.Context, b *strings.Builder, rawToken, attestationServiceURL string) error {
	var claims verify.MaaTokenClaims
	_, err := jwt.ParseWithClaims(rawToken, &claims, keyFromJKUFunc(ctx, attestationServiceURL), jwt.WithIssuedAt())
	if err != nil {
		return fmt.Errorf("parsing token: %w", err)
	}

	out, err := json.MarshalIndent(claims, "\t\t", "  ")
	if err != nil {
		return fmt.Errorf("marshaling claims: %w", err)
	}

	b.WriteString("\tMicrosoft Azure Attestation Token:\n\t")
	b.WriteString(string(out))
	return nil
}

// keyFromJKUFunc returns a function that gets the JSON Web Key URI from the token
// and fetches the key from that URI. The keys are then parsed, and the key with
// the kid that matches the token header is returned.
func keyFromJKUFunc(ctx context.Context, webKeysURLBase string) func(token *jwt.Token) (any, error) {
	return func(token *jwt.Token) (any, error) {
		webKeysURL, err := url.JoinPath(webKeysURLBase, "certs")
		if err != nil {
			return nil, fmt.Errorf("joining web keys base URL with path: %w", err)
		}

		if token.Header["alg"] != "RS256" {
			return nil, fmt.Errorf("invalid signing algorithm: %s", token.Header["alg"])
		}
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid kid: %v", token.Header["kid"])
		}
		jku, ok := token.Header["jku"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid jku: %v", token.Header["jku"])
		}
		if jku != webKeysURL {
			return nil, fmt.Errorf("jku from token (%s) does not match configured attestation service (%s)", jku, webKeysURL)
		}

		keySetBytes, err := httpGet(ctx, jku)
		if err != nil {
			return nil, fmt.Errorf("getting signing keys from jku %s: %w", jku, err)
		}

		var rawKeySet struct {
			Keys []struct {
				X5c [][]byte
				Kid string
			}
		}

		if err := json.Unmarshal(keySetBytes, &rawKeySet); err != nil {
			return nil, err
		}

		for _, key := range rawKeySet.Keys {
			if key.Kid != kid {
				continue
			}
			cert, err := x509.ParseCertificate(key.X5c[0])
			if err != nil {
				return nil, fmt.Errorf("parsing certificate: %w", err)
			}

			return cert.PublicKey, nil
		}

		return nil, fmt.Errorf("no key found for kid %s", kid)
	}
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

// azureInstanceInfo is the b64-decoded InstanceInfo field of the attestation document.
// as of now (2023-04-03), it only contains interesting data on Azure.
type azureInstanceInfo struct {
	Vcek              []byte
	CertChain         []byte
	AttestationReport []byte
	RuntimeData       []byte
	MAAToken          string
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

// writeIndentfln writes a formatted string to the builder with the given indentation level
// and a newline at the end.
func writeIndentfln(b *strings.Builder, indentLvl int, format string, args ...any) {
	for i := 0; i < indentLvl; i++ {
		b.WriteByte('\t')
	}
	b.WriteString(fmt.Sprintf(format+"\n", args...))
}

func httpGet(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func newCertificates(certTypeName string, cert []byte, log debugLog) (certs []verify.Certificate, err error) {
	newlinesTrimmed := strings.TrimSpace(string(cert))

	log.Debugf("Decoding PEM certificate: %s", certTypeName)
	i := 1
	var rest []byte
	var block *pem.Block
	for block, rest = pem.Decode([]byte(newlinesTrimmed)); block != nil; block, rest = pem.Decode(rest) {
		log.Debugf("Parsing PEM block: %d", i)
		if block.Type != "CERTIFICATE" {
			return certs, fmt.Errorf("parse %s: expected PEM block type 'CERTIFICATE', got '%s'", certTypeName, block.Type)
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return certs, fmt.Errorf("parse %s: %w", certTypeName, err)
		}
		if certTypeName == "VCEK certificate" {
			vcekExts, err := kds.VcekCertificateExtensions(cert)
			if err != nil {
				return certs, fmt.Errorf("parsing VCEK certificate extensions: %w", err)
			}
			certs = append(certs, verify.Certificate{
				Certificate:   cert,
				CertTypeName:  certTypeName,
				StructVersion: vcekExts.StructVersion,
				ProductName:   vcekExts.ProductName,
				TCBVersion:    newTCBVersion(vcekExts.TCBVersion),
				HardwareID:    vcekExts.HWID,
			})
		} else {
			certs = append(certs, verify.Certificate{
				Certificate:  cert,
				CertTypeName: certTypeName,
			})
		}
		i++
	}
	if i == 1 {
		return certs, fmt.Errorf("parse %s: no PEM blocks found", certTypeName)
	}
	if len(rest) != 0 {
		return certs, fmt.Errorf("parse %s: remaining PEM block is not a valid certificate: %s", certTypeName, rest)
	}
	return certs, nil
}

func newSNPReport(reportBytes []byte) (res verify.SNPReport, err error) {
	report, err := abi.ReportToProto(reportBytes)
	if err != nil {
		return res, fmt.Errorf("parsing report to proto: %w", err)
	}

	policy, err := abi.ParseSnpPolicy(report.Policy)
	if err != nil {
		return res, fmt.Errorf("parsing policy: %w", err)
	}

	platformInfo, err := abi.ParseSnpPlatformInfo(report.PlatformInfo)
	if err != nil {
		return res, fmt.Errorf("parsing platform info: %w", err)
	}

	signature, err := abi.ReportToSignatureDER(reportBytes)
	if err != nil {
		return res, fmt.Errorf("parsing signature: %w", err)
	}

	signerInfo, err := abi.ParseSignerInfo(report.SignerInfo)
	if err != nil {
		return res, fmt.Errorf("parsing signer info: %w", err)
	}
	return verify.SNPReport{
		Version:              report.Version,
		GuestSvn:             report.GuestSvn,
		PolicyABIMinor:       policy.ABIMinor,
		PolicyABIMajor:       policy.ABIMajor,
		PolicySMT:            policy.SMT,
		PolicyMigrationAgent: policy.MigrateMA,
		PolicyDebug:          policy.Debug,
		PolicySingleSocket:   policy.SingleSocket,
		FamilyID:             report.FamilyId,
		ImageID:              report.ImageId,
		Vmpl:                 report.Vmpl,
		SignatureAlgo:        report.SignatureAlgo,
		CurrentTCB:           newTCBVersion(kds.TCBVersion(report.CurrentTcb)),
		PlatformInfo: verify.PlatformInfo{
			SMT:  platformInfo.SMTEnabled,
			TSME: platformInfo.TSMEEnabled,
		},
		SignerInfo: verify.SignerInfo{
			AuthorKeyEn: signerInfo.AuthorKeyEn,
			MaskChipKey: signerInfo.MaskChipKey,
			SigningKey:  signerInfo.SigningKey,
		},
		ReportData:      report.ReportData,
		Measurement:     report.Measurement,
		HostData:        report.HostData,
		IDKeyDigest:     report.IdKeyDigest,
		AuthorKeyDigest: report.AuthorKeyDigest,
		ReportID:        report.ReportId,
		ReportIDMa:      report.ReportIdMa,
		ReportedTCB:     newTCBVersion(kds.TCBVersion(report.ReportedTcb)),
		ChipID:          report.ChipId,
		CommittedTCB:    newTCBVersion(kds.TCBVersion(report.CommittedTcb)),
		CurrentBuild:    report.CurrentBuild,
		CurrentMinor:    report.CurrentMinor,
		CurrentMajor:    report.CurrentMajor,
		CommittedBuild:  report.CommittedBuild,
		CommittedMinor:  report.CommittedMinor,
		CommittedMajor:  report.CommittedMajor,
		LaunchTCB:       newTCBVersion(kds.TCBVersion(report.LaunchTcb)),
		Signature:       signature,
	}, nil
}

func newMAAToken(ctx context.Context, rawToken, attestationServiceURL string) (verify.MaaTokenClaims, error) {
	var claims verify.MaaTokenClaims
	_, err := jwt.ParseWithClaims(rawToken, &claims, keyFromJKUFunc(ctx, attestationServiceURL), jwt.WithIssuedAt())
	return claims, err
}

func newTCBVersion(tcbVersion kds.TCBVersion) (res verify.TCBVersion) {
	tcb := kds.DecomposeTCBVersion(tcbVersion)
	return verify.TCBVersion{
		Bootloader: tcb.BlSpl,
		TEE:        tcb.TeeSpl,
		SNP:        tcb.SnpSpl,
		Microcode:  tcb.UcodeSpl,
		Spl4:       tcb.Spl4,
		Spl5:       tcb.Spl5,
		Spl6:       tcb.Spl6,
		Spl7:       tcb.Spl7,
		UcodeSpl:   tcb.UcodeSpl,
	}
}
