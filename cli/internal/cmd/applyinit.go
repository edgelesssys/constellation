/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"path/filepath"
	"text/tabwriter"

	"github.com/edgelesssys/constellation/v2/bootstrapper/initproto"
	"github.com/edgelesssys/constellation/v2/internal/attestation/choose"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/constellation"
	"github.com/edgelesssys/constellation/v2/internal/constellation/state"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/kms/uri"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
)

// runInit runs the init RPC to set up the Kubernetes cluster.
// This function only needs to be run once per cluster.
// On success, it writes the Kubernetes admin config file to disk.
// Therefore it is skipped if the Kubernetes admin config file already exists.
func (a *applyCmd) runInit(cmd *cobra.Command, conf *config.Config, stateFile *state.State) (*bytes.Buffer, error) {
	a.log.Debugf("Creating aTLS Validator for %s", conf.GetAttestationConfig().GetVariant())
	validator, err := choose.Validator(conf.GetAttestationConfig(), a.wLog)
	if err != nil {
		return nil, fmt.Errorf("creating validator: %w", err)
	}

	a.log.Debugf("Running init RPC")
	masterSecret, err := a.generateAndPersistMasterSecret(cmd.OutOrStdout())
	if err != nil {
		return nil, fmt.Errorf("generating master secret: %w", err)
	}

	measurementSalt, err := a.applier.GenerateMeasurementSalt()
	if err != nil {
		return nil, fmt.Errorf("generating measurement salt: %w", err)
	}

	clusterLogs := &bytes.Buffer{}
	resp, err := a.applier.Init(
		cmd.Context(), validator, stateFile, clusterLogs,
		constellation.InitPayload{
			MasterSecret:    masterSecret,
			MeasurementSalt: measurementSalt,
			K8sVersion:      conf.KubernetesVersion,
			ConformanceMode: a.flags.conformance,
			ServiceCIDR:     conf.ServiceCIDR,
		})
	if len(clusterLogs.Bytes()) > 0 {
		if err := a.fileHandler.Write(constants.ErrorLog, clusterLogs.Bytes(), file.OptAppend); err != nil {
			return nil, fmt.Errorf("writing bootstrapper logs: %w", err)
		}
	}
	if err != nil {
		var nonRetriable *constellation.NonRetriableInitError
		if errors.As(err, &nonRetriable) {
			cmd.PrintErrln("Cluster initialization failed. This error is not recoverable.")
			cmd.PrintErrln("Terminate your cluster and try again.")
			if nonRetriable.LogCollectionErr != nil {
				cmd.PrintErrf("Failed to collect logs from bootstrapper: %s\n", nonRetriable.LogCollectionErr)
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

// generateAndPersistMasterSecret generates a 32 byte master secret and saves it to disk.
func (a *applyCmd) generateAndPersistMasterSecret(outWriter io.Writer) (uri.MasterSecret, error) {
	secret, err := a.applier.GenerateMasterSecret()
	if err != nil {
		return uri.MasterSecret{}, fmt.Errorf("generating master secret: %w", err)
	}
	if err := a.fileHandler.WriteJSON(constants.MasterSecretFilename, secret, file.OptNone); err != nil {
		return uri.MasterSecret{}, fmt.Errorf("writing master secret: %w", err)
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
