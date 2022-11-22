/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"strconv"
	"text/tabwriter"
	"time"

	"github.com/edgelesssys/constellation/v2/bootstrapper/initproto"
	"github.com/edgelesssys/constellation/v2/cli/internal/cloudcmd"
	"github.com/edgelesssys/constellation/v2/cli/internal/clusterid"
	"github.com/edgelesssys/constellation/v2/cli/internal/helm"
	"github.com/edgelesssys/constellation/v2/internal/cloud/azureshared"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/cloud/gcpshared"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/crypto"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/grpc/dialer"
	grpcRetry "github.com/edgelesssys/constellation/v2/internal/grpc/retry"
	"github.com/edgelesssys/constellation/v2/internal/license"
	"github.com/edgelesssys/constellation/v2/internal/retry"
	"github.com/edgelesssys/constellation/v2/internal/versions"
	kms "github.com/edgelesssys/constellation/v2/kms/setup"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

// NewInitCmd returns a new cobra.Command for the init command.
func NewInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize the Constellation cluster",
		Long:  "Initialize the Constellation cluster. Start your confidential Kubernetes.",
		Args:  cobra.ExactArgs(0),
		RunE:  runInitialize,
	}
	cmd.Flags().String("master-secret", "", "path to base64-encoded master secret")
	cmd.Flags().Bool("conformance", false, "enable conformance mode")
	return cmd
}

// runInitialize runs the initialize command.
func runInitialize(cmd *cobra.Command, args []string) error {
	fileHandler := file.NewHandler(afero.NewOsFs())
	newDialer := func(validator *cloudcmd.Validator) *dialer.Dialer {
		return dialer.New(nil, validator.V(cmd), &net.Dialer{})
	}

	spinner := newSpinner(cmd.ErrOrStderr())
	defer spinner.Stop()

	ctx, cancel := context.WithTimeout(cmd.Context(), time.Hour)
	defer cancel()
	cmd.SetContext(ctx)

	return initialize(cmd, newDialer, fileHandler, license.NewClient(), spinner)
}

// initialize initializes a Constellation.
func initialize(cmd *cobra.Command, newDialer func(validator *cloudcmd.Validator) *dialer.Dialer,
	fileHandler file.Handler, quotaChecker license.QuotaChecker, spinner spinnerInterf,
) error {
	flags, err := evalFlagArgs(cmd)
	if err != nil {
		return err
	}

	conf, err := config.New(fileHandler, flags.configPath)
	if err != nil {
		return displayConfigValidationErrors(cmd.ErrOrStderr(), err)
	}

	var idFile clusterid.File
	if err := fileHandler.ReadJSON(constants.ClusterIDsFileName, &idFile); err != nil {
		return fmt.Errorf("reading cluster ID file: %w", err)
	}

	k8sVersion, err := versions.NewValidK8sVersion(conf.KubernetesVersion)
	if err != nil {
		return fmt.Errorf("validating kubernetes version: %w", err)
	}
	if versions.IsPreviewK8sVersion(k8sVersion) {
		cmd.PrintErrf("Warning: Constellation with Kubernetes %v is still in preview. Use only for evaluation purposes.\n", k8sVersion)
	}

	provider := conf.GetProvider()
	checker := license.NewChecker(quotaChecker, fileHandler)
	if err := checker.CheckLicense(cmd.Context(), provider, conf.Provider, cmd.Printf); err != nil {
		cmd.PrintErrf("License check failed: %v", err)
	}

	validator, err := cloudcmd.NewValidator(provider, conf)
	if err != nil {
		return err
	}

	serviceAccURI, err := getMarshaledServiceAccountURI(provider, conf, fileHandler)
	if err != nil {
		return err
	}

	masterSecret, err := readOrGenerateMasterSecret(cmd.OutOrStdout(), fileHandler, flags.masterSecretPath)
	if err != nil {
		return fmt.Errorf("parsing or generating master secret from file %s: %w", flags.masterSecretPath, err)
	}
	helmLoader := helm.New(provider, k8sVersion)
	helmDeployments, err := helmLoader.Load(conf, flags.conformance, masterSecret.Key, masterSecret.Salt)
	if err != nil {
		return fmt.Errorf("loading Helm charts: %w", err)
	}

	spinner.Start("Initializing cluster ", false)
	req := &initproto.InitRequest{
		MasterSecret:           masterSecret.Key,
		Salt:                   masterSecret.Salt,
		KmsUri:                 kms.ClusterKMSURI,
		StorageUri:             kms.NoStoreURI,
		KeyEncryptionKeyId:     "",
		UseExistingKek:         false,
		CloudServiceAccountUri: serviceAccURI,
		KubernetesVersion:      conf.KubernetesVersion,
		KubernetesComponents:   versions.VersionConfigs[k8sVersion].KubernetesComponents.ToProto(),
		HelmDeployments:        helmDeployments,
		EnforcedPcrs:           conf.EnforcedPCRs(),
		EnforceIdkeydigest:     conf.EnforcesIDKeyDigest(),
		ConformanceMode:        flags.conformance,
	}
	resp, err := initCall(cmd.Context(), newDialer(validator), idFile.IP, req)
	spinner.Stop()
	if err != nil {
		return err
	}

	idFile.CloudProvider = provider
	if err := writeOutput(idFile, resp, cmd.OutOrStdout(), fileHandler); err != nil {
		return err
	}

	return nil
}

func initCall(ctx context.Context, dialer grpcDialer, ip string, req *initproto.InitRequest) (*initproto.InitResponse, error) {
	doer := &initDoer{
		dialer:   dialer,
		endpoint: net.JoinHostPort(ip, strconv.Itoa(constants.BootstrapperPort)),
		req:      req,
	}
	retrier := retry.NewIntervalRetrier(doer, 30*time.Second, grpcRetry.ServiceIsUnavailable)
	if err := retrier.Do(ctx); err != nil {
		return nil, err
	}
	return doer.resp, nil
}

type initDoer struct {
	dialer   grpcDialer
	endpoint string
	req      *initproto.InitRequest
	resp     *initproto.InitResponse
}

func (d *initDoer) Do(ctx context.Context) error {
	conn, err := d.dialer.Dial(ctx, d.endpoint)
	if err != nil {
		return fmt.Errorf("dialing init server: %w", err)
	}
	defer conn.Close()
	protoClient := initproto.NewAPIClient(conn)
	resp, err := protoClient.Init(ctx, d.req)
	if err != nil {
		return fmt.Errorf("init call: %w", err)
	}
	d.resp = resp
	return nil
}

func writeOutput(idFile clusterid.File, resp *initproto.InitResponse, wr io.Writer, fileHandler file.Handler) error {
	fmt.Fprint(wr, "Your Constellation cluster was successfully initialized.\n\n")

	ownerID := hex.EncodeToString(resp.OwnerId)
	clusterID := hex.EncodeToString(resp.ClusterId)

	tw := tabwriter.NewWriter(wr, 0, 0, 2, ' ', 0)
	// writeRow(tw, "Constellation cluster's owner identifier", ownerID)
	writeRow(tw, "Constellation cluster identifier", clusterID)
	writeRow(tw, "Kubernetes configuration", constants.AdminConfFilename)
	tw.Flush()
	fmt.Fprintln(wr)

	if err := fileHandler.Write(constants.AdminConfFilename, resp.Kubeconfig, file.OptNone); err != nil {
		return fmt.Errorf("writing kubeconfig: %w", err)
	}

	idFile.OwnerID = ownerID
	idFile.ClusterID = clusterID

	if err := fileHandler.WriteJSON(constants.ClusterIDsFileName, idFile, file.OptOverwrite); err != nil {
		return fmt.Errorf("writing Constellation id file: %w", err)
	}

	fmt.Fprintln(wr, "You can now connect to your cluster by executing:")
	fmt.Fprintf(wr, "\texport KUBECONFIG=\"$PWD/%s\"\n", constants.AdminConfFilename)
	return nil
}

func writeRow(wr io.Writer, col1 string, col2 string) {
	fmt.Fprint(wr, col1, "\t", col2, "\n")
}

// evalFlagArgs gets the flag values and does preprocessing of these values like
// reading the content from file path flags and deriving other values from flag combinations.
func evalFlagArgs(cmd *cobra.Command) (initFlags, error) {
	masterSecretPath, err := cmd.Flags().GetString("master-secret")
	if err != nil {
		return initFlags{}, fmt.Errorf("parsing master-secret path flag: %w", err)
	}
	conformance, err := cmd.Flags().GetBool("conformance")
	if err != nil {
		return initFlags{}, fmt.Errorf("parsing autoscale flag: %w", err)
	}
	configPath, err := cmd.Flags().GetString("config")
	if err != nil {
		return initFlags{}, fmt.Errorf("parsing config path flag: %w", err)
	}

	return initFlags{
		configPath:       configPath,
		conformance:      conformance,
		masterSecretPath: masterSecretPath,
	}, nil
}

// initFlags are the resulting values of flag preprocessing.
type initFlags struct {
	configPath       string
	masterSecretPath string
	conformance      bool
}

// masterSecret holds the master key and salt for deriving keys.
type masterSecret struct {
	Key  []byte `json:"key"`
	Salt []byte `json:"salt"`
}

// readOrGenerateMasterSecret reads a base64 encoded master secret from file or generates a new 32 byte secret.
func readOrGenerateMasterSecret(outWriter io.Writer, fileHandler file.Handler, filename string) (masterSecret, error) {
	if filename != "" {
		var secret masterSecret
		if err := fileHandler.ReadJSON(filename, &secret); err != nil {
			return masterSecret{}, err
		}

		if len(secret.Key) < crypto.MasterSecretLengthMin {
			return masterSecret{}, fmt.Errorf("provided master secret is smaller than the required minimum of %d Bytes", crypto.MasterSecretLengthMin)
		}
		if len(secret.Salt) < crypto.RNGLengthDefault {
			return masterSecret{}, fmt.Errorf("provided salt is smaller than the required minimum of %d Bytes", crypto.RNGLengthDefault)
		}
		return secret, nil
	}

	// No file given, generate a new secret, and save it to disk
	key, err := crypto.GenerateRandomBytes(crypto.MasterSecretLengthDefault)
	if err != nil {
		return masterSecret{}, err
	}
	salt, err := crypto.GenerateRandomBytes(crypto.RNGLengthDefault)
	if err != nil {
		return masterSecret{}, err
	}
	secret := masterSecret{
		Key:  key,
		Salt: salt,
	}

	if err := fileHandler.WriteJSON(constants.MasterSecretFilename, secret, file.OptNone); err != nil {
		return masterSecret{}, err
	}
	fmt.Fprintf(outWriter, "Your Constellation master secret was successfully written to ./%s\n", constants.MasterSecretFilename)
	return secret, nil
}

func readIPFromIDFile(fileHandler file.Handler) (string, error) {
	var idFile clusterid.File
	if err := fileHandler.ReadJSON(constants.ClusterIDsFileName, &idFile); err != nil {
		return "", err
	}
	if idFile.IP == "" {
		return "", fmt.Errorf("missing IP address in %q", constants.ClusterIDsFileName)
	}
	return idFile.IP, nil
}

func getMarshaledServiceAccountURI(provider cloudprovider.Provider, config *config.Config, fileHandler file.Handler) (string, error) {
	switch provider {
	case cloudprovider.GCP:
		path := config.Provider.GCP.ServiceAccountKeyPath

		var key gcpshared.ServiceAccountKey
		if err := fileHandler.ReadJSON(path, &key); err != nil {
			return "", fmt.Errorf("reading service account key from path %q: %w", path, err)
		}

		return key.ToCloudServiceAccountURI(), nil

	case cloudprovider.AWS:
		return "", nil // AWS does not need a service account URI
	case cloudprovider.Azure:
		creds := azureshared.ApplicationCredentials{
			TenantID:          config.Provider.Azure.TenantID,
			AppClientID:       config.Provider.Azure.AppClientID,
			ClientSecretValue: config.Provider.Azure.ClientSecretValue,
			Location:          config.Provider.Azure.Location,
		}
		return creds.ToCloudServiceAccountURI(), nil

	case cloudprovider.QEMU:
		return "", nil // QEMU does not use service account keys

	default:
		return "", fmt.Errorf("unsupported cloud provider %q", provider)
	}
}

type grpcDialer interface {
	Dial(ctx context.Context, target string) (*grpc.ClientConn, error)
}
