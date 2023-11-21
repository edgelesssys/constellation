/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"path/filepath"
	"strconv"
	"text/tabwriter"
	"time"

	"github.com/edgelesssys/constellation/v2/bootstrapper/initproto"
	"github.com/edgelesssys/constellation/v2/internal/cloudcmd"
	"github.com/edgelesssys/constellation/v2/internal/state"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/crypto"
	"github.com/edgelesssys/constellation/v2/internal/file"
	grpcRetry "github.com/edgelesssys/constellation/v2/internal/grpc/retry"
	"github.com/edgelesssys/constellation/v2/internal/kms/uri"
	"github.com/edgelesssys/constellation/v2/internal/retry"
	"github.com/edgelesssys/constellation/v2/internal/versions"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
)

// runInit runs the init RPC to set up the Kubernetes cluster.
// This function only needs to be run once per cluster.
// On success, it writes the Kubernetes admin config file to disk.
// Therefore it is skipped if the Kubernetes admin config file already exists.
func (a *applyCmd) runInit(cmd *cobra.Command, conf *config.Config, stateFile *state.State) (*bytes.Buffer, error) {
	a.log.Debugf("Running init RPC")
	a.log.Debugf("Creating aTLS Validator for %s", conf.GetAttestationConfig().GetVariant())
	validator, err := cloudcmd.NewValidator(cmd, conf.GetAttestationConfig(), a.log)
	if err != nil {
		return nil, fmt.Errorf("creating new validator: %w", err)
	}

	a.log.Debugf("Generating master secret")
	masterSecret, err := a.generateAndPersistMasterSecret(cmd.OutOrStdout())
	if err != nil {
		return nil, fmt.Errorf("generating master secret: %w", err)
	}
	a.log.Debugf("Generated master secret key and salt values")

	a.log.Debugf("Generating measurement salt")
	measurementSalt, err := crypto.GenerateRandomBytes(crypto.RNGLengthDefault)
	if err != nil {
		return nil, fmt.Errorf("generating measurement salt: %w", err)
	}

	a.spinner.Start("Connecting ", false)
	req := &initproto.InitRequest{
		KmsUri:               masterSecret.EncodeToURI(),
		StorageUri:           uri.NoStoreURI,
		MeasurementSalt:      measurementSalt,
		KubernetesVersion:    versions.VersionConfigs[conf.KubernetesVersion].ClusterVersion,
		KubernetesComponents: versions.VersionConfigs[conf.KubernetesVersion].KubernetesComponents.ToInitProto(),
		ConformanceMode:      a.flags.conformance,
		InitSecret:           stateFile.Infrastructure.InitSecret,
		ClusterName:          stateFile.Infrastructure.Name,
		ApiserverCertSans:    stateFile.Infrastructure.APIServerCertSANs,
	}
	a.log.Debugf("Sending initialization request")
	resp, err := a.initCall(cmd.Context(), a.newDialer(validator), stateFile.Infrastructure.ClusterEndpoint, req)
	a.spinner.Stop()
	a.log.Debugf("Initialization request finished")

	if err != nil {
		var nonRetriable *nonRetriableError
		if errors.As(err, &nonRetriable) {
			cmd.PrintErrln("Cluster initialization failed. This error is not recoverable.")
			cmd.PrintErrln("Terminate your cluster and try again.")
			if nonRetriable.logCollectionErr != nil {
				cmd.PrintErrf("Failed to collect logs from bootstrapper: %s\n", nonRetriable.logCollectionErr)
			} else {
				cmd.PrintErrf("Fetched bootstrapper logs are stored in %q\n", a.flags.pathPrefixer.PrefixPrintablePath(constants.ErrorLog))
			}
		}
		return nil, err
	}
	a.log.Debugf("Initialization request successful")

	a.log.Debugf("Buffering init success message")
	bufferedOutput := &bytes.Buffer{}
	if err := a.writeInitOutput(stateFile, resp, a.flags.mergeConfigs, bufferedOutput, measurementSalt); err != nil {
		return nil, err
	}

	return bufferedOutput, nil
}

// initCall performs the gRPC call to the bootstrapper to initialize the cluster.
func (a *applyCmd) initCall(ctx context.Context, dialer grpcDialer, ip string, req *initproto.InitRequest) (*initproto.InitSuccessResponse, error) {
	doer := &initDoer{
		dialer:   dialer,
		endpoint: net.JoinHostPort(ip, strconv.Itoa(constants.BootstrapperPort)),
		req:      req,
		log:      a.log,
		spinner:  a.spinner,
		fh:       file.NewHandler(afero.NewOsFs()),
	}

	// Create a wrapper function that allows logging any returned error from the retrier before checking if it's the expected retriable one.
	serviceIsUnavailable := func(err error) bool {
		isServiceUnavailable := grpcRetry.ServiceIsUnavailable(err)
		a.log.Debugf("Encountered error (retriable: %t): %s", isServiceUnavailable, err)
		return isServiceUnavailable
	}

	a.log.Debugf("Making initialization call, doer is %+v", doer)
	retrier := retry.NewIntervalRetrier(doer, 30*time.Second, serviceIsUnavailable)
	if err := retrier.Do(ctx); err != nil {
		return nil, err
	}
	return doer.resp, nil
}

// generateAndPersistMasterSecret generates a 32 byte master secret and saves it to disk.
func (a *applyCmd) generateAndPersistMasterSecret(outWriter io.Writer) (uri.MasterSecret, error) {
	// No file given, generate a new secret, and save it to disk
	key, err := crypto.GenerateRandomBytes(crypto.MasterSecretLengthDefault)
	if err != nil {
		return uri.MasterSecret{}, err
	}
	salt, err := crypto.GenerateRandomBytes(crypto.RNGLengthDefault)
	if err != nil {
		return uri.MasterSecret{}, err
	}
	secret := uri.MasterSecret{
		Key:  key,
		Salt: salt,
	}
	if err := a.fileHandler.WriteJSON(constants.MasterSecretFilename, secret, file.OptNone); err != nil {
		return uri.MasterSecret{}, err
	}
	fmt.Fprintf(outWriter, "Your Constellation master secret was successfully written to %q\n", a.flags.pathPrefixer.PrefixPrintablePath(constants.MasterSecretFilename))
	return secret, nil
}

// writeInitOutput writes the output of a cluster initialization to the
// state- / kubeconfig-file and saves it to disk.
func (a *applyCmd) writeInitOutput(
	stateFile *state.State, initResp *initproto.InitSuccessResponse,
	mergeConfig bool, wr io.Writer, measurementSalt []byte,
) error {
	fmt.Fprint(wr, "Your Constellation cluster was successfully initialized.\n\n")

	ownerID := hex.EncodeToString(initResp.GetOwnerId())
	clusterID := hex.EncodeToString(initResp.GetClusterId())

	stateFile.SetClusterValues(state.ClusterValues{
		MeasurementSalt: measurementSalt,
		OwnerID:         ownerID,
		ClusterID:       clusterID,
	})

	tw := tabwriter.NewWriter(wr, 0, 0, 2, ' ', 0)
	writeRow(tw, "Constellation cluster identifier", clusterID)
	writeRow(tw, "Kubernetes configuration", a.flags.pathPrefixer.PrefixPrintablePath(constants.AdminConfFilename))
	tw.Flush()
	fmt.Fprintln(wr)

	a.log.Debugf("Rewriting cluster server address in kubeconfig to %s", stateFile.Infrastructure.ClusterEndpoint)
	kubeconfig, err := clientcmd.Load(initResp.GetKubeconfig())
	if err != nil {
		return fmt.Errorf("loading kubeconfig: %w", err)
	}
	if len(kubeconfig.Clusters) != 1 {
		return fmt.Errorf("expected exactly one cluster in kubeconfig, got %d", len(kubeconfig.Clusters))
	}
	for _, cluster := range kubeconfig.Clusters {
		kubeEndpoint, err := url.Parse(cluster.Server)
		if err != nil {
			return fmt.Errorf("parsing kubeconfig server URL: %w", err)
		}
		kubeEndpoint.Host = net.JoinHostPort(stateFile.Infrastructure.ClusterEndpoint, kubeEndpoint.Port())
		cluster.Server = kubeEndpoint.String()
	}
	kubeconfigBytes, err := clientcmd.Write(*kubeconfig)
	if err != nil {
		return fmt.Errorf("marshaling kubeconfig: %w", err)
	}

	if err := a.fileHandler.Write(constants.AdminConfFilename, kubeconfigBytes, file.OptNone); err != nil {
		return fmt.Errorf("writing kubeconfig: %w", err)
	}
	a.log.Debugf("Kubeconfig written to %s", a.flags.pathPrefixer.PrefixPrintablePath(constants.AdminConfFilename))

	if mergeConfig {
		if err := a.merger.mergeConfigs(constants.AdminConfFilename, a.fileHandler); err != nil {
			writeRow(tw, "Failed to automatically merge kubeconfig", err.Error())
			mergeConfig = false // Set to false so we don't print the wrong message below.
		} else {
			writeRow(tw, "Kubernetes configuration merged with default config", "")
		}
	}

	if err := stateFile.WriteToFile(a.fileHandler, constants.StateFilename); err != nil {
		return fmt.Errorf("writing Constellation state file: %w", err)
	}

	a.log.Debugf("Constellation state file written to %s", a.flags.pathPrefixer.PrefixPrintablePath(constants.StateFilename))

	if !mergeConfig {
		fmt.Fprintln(wr, "You can now connect to your cluster by executing:")

		exportPath, err := filepath.Abs(constants.AdminConfFilename)
		if err != nil {
			return fmt.Errorf("getting absolute path to kubeconfig: %w", err)
		}

		fmt.Fprintf(wr, "\texport KUBECONFIG=%q\n", exportPath)
	} else {
		fmt.Fprintln(wr, "Constellation kubeconfig merged with default config.")

		if a.merger.kubeconfigEnvVar() != "" {
			fmt.Fprintln(wr, "Warning: KUBECONFIG environment variable is set.")
			fmt.Fprintln(wr, "You may need to unset it to use the default config and connect to your cluster.")
		} else {
			fmt.Fprintln(wr, "You can now connect to your cluster.")
		}
	}
	fmt.Fprintln(wr) // add final newline
	return nil
}
