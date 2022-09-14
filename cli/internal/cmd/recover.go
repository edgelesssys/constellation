/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/edgelesssys/constellation/cli/internal/cloudcmd"
	"github.com/edgelesssys/constellation/disk-mapper/recoverproto"
	"github.com/edgelesssys/constellation/internal/attestation"
	"github.com/edgelesssys/constellation/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/edgelesssys/constellation/internal/crypto"
	"github.com/edgelesssys/constellation/internal/file"
	"github.com/edgelesssys/constellation/internal/grpc/dialer"
	grpcRetry "github.com/edgelesssys/constellation/internal/grpc/retry"
	"github.com/edgelesssys/constellation/internal/retry"
	"github.com/edgelesssys/constellation/internal/state"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"go.uber.org/multierr"
)

// NewRecoverCmd returns a new cobra.Command for the recover command.
func NewRecoverCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "recover",
		Short: "Recover a completely stopped Constellation cluster",
		Long: "Recover a Constellation cluster by sending a recovery key to an instance in the boot stage." +
			"\nThis is only required if instances restart without other instances available for bootstrapping.",
		Args: cobra.ExactArgs(0),
		RunE: runRecover,
	}
	cmd.Flags().StringP("endpoint", "e", "", "endpoint of the instance, passed as HOST[:PORT] (required)")
	must(cmd.MarkFlagRequired("endpoint"))
	cmd.Flags().String("master-secret", constants.MasterSecretFilename, "path to master secret file")
	return cmd
}

func runRecover(cmd *cobra.Command, _ []string) error {
	fileHandler := file.NewHandler(afero.NewOsFs())
	newDialer := func(validator *cloudcmd.Validator) *dialer.Dialer {
		return dialer.New(nil, validator.V(cmd), &net.Dialer{})
	}
	return recover(cmd, fileHandler, newDialer)
}

func recover(cmd *cobra.Command, fileHandler file.Handler, newDialer func(validator *cloudcmd.Validator) *dialer.Dialer) error {
	flags, err := parseRecoverFlags(cmd)
	if err != nil {
		return err
	}

	var masterSecret masterSecret
	if err := fileHandler.ReadJSON(flags.secretPath, &masterSecret); err != nil {
		return err
	}

	var stat state.ConstellationState
	if err := fileHandler.ReadJSON(constants.StateFilename, &stat); err != nil {
		return err
	}

	provider := cloudprovider.FromString(stat.CloudProvider)
	config, err := readConfig(cmd.OutOrStdout(), fileHandler, flags.configPath)
	if err != nil {
		return fmt.Errorf("reading and validating config: %w", err)
	}

	validator, err := cloudcmd.NewValidator(provider, config)
	if err != nil {
		return err
	}

	if err := recoverCall(cmd.Context(), newDialer(validator), flags.endpoint, masterSecret.Key, masterSecret.Salt); err != nil {
		return fmt.Errorf("recovering cluster: %w", err)
	}

	cmd.Println("Pushed recovery key.")
	return nil
}

func recoverCall(ctx context.Context, dialer grpcDialer, endpoint string, key, salt []byte) error {
	measurementSecret, err := attestation.DeriveMeasurementSecret(key, salt)
	if err != nil {
		return err
	}
	doer := &recoverDoer{
		dialer:            dialer,
		endpoint:          endpoint,
		getDiskKey:        getStateDiskKeyFunc(key, salt),
		measurementSecret: measurementSecret,
	}
	retrier := retry.NewIntervalRetrier(doer, 30*time.Second, grpcRetry.ServiceIsUnavailable)
	if err := retrier.Do(ctx); err != nil {
		return err
	}
	return nil
}

type recoverDoer struct {
	dialer            grpcDialer
	endpoint          string
	measurementSecret []byte
	getDiskKey        func(uuid string) (key []byte, err error)
}

func (d *recoverDoer) Do(ctx context.Context) (retErr error) {
	conn, err := d.dialer.Dial(ctx, d.endpoint)
	if err != nil {
		return fmt.Errorf("dialing recovery server: %w", err)
	}
	defer conn.Close()

	// set up streaming client
	protoClient := recoverproto.NewAPIClient(conn)
	recoverclient, err := protoClient.Recover(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err := recoverclient.CloseSend(); err != nil {
			multierr.AppendInto(&retErr, err)
		}
	}()

	// send measurement secret as first message
	if err := recoverclient.Send(&recoverproto.RecoverMessage{
		Request: &recoverproto.RecoverMessage_MeasurementSecret{
			MeasurementSecret: d.measurementSecret,
		},
	}); err != nil {
		return err
	}

	// receive disk uuid
	res, err := recoverclient.Recv()
	if err != nil {
		return err
	}
	stateDiskKey, err := d.getDiskKey(res.DiskUuid)
	if err != nil {
		return err
	}

	// send disk key
	if err := recoverclient.Send(&recoverproto.RecoverMessage{
		Request: &recoverproto.RecoverMessage_StateDiskKey{
			StateDiskKey: stateDiskKey,
		},
	}); err != nil {
		return err
	}

	if _, err := recoverclient.Recv(); err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	return nil
}

func parseRecoverFlags(cmd *cobra.Command) (recoverFlags, error) {
	endpoint, err := cmd.Flags().GetString("endpoint")
	if err != nil {
		return recoverFlags{}, fmt.Errorf("parsing endpoint argument: %w", err)
	}
	endpoint, err = addPortIfMissing(endpoint, constants.RecoveryPort)
	if err != nil {
		return recoverFlags{}, fmt.Errorf("validating endpoint argument: %w", err)
	}

	masterSecretPath, err := cmd.Flags().GetString("master-secret")
	if err != nil {
		return recoverFlags{}, fmt.Errorf("parsing master-secret path argument: %w", err)
	}

	configPath, err := cmd.Flags().GetString("config")
	if err != nil {
		return recoverFlags{}, fmt.Errorf("parsing config path argument: %w", err)
	}

	return recoverFlags{
		endpoint:   endpoint,
		secretPath: masterSecretPath,
		configPath: configPath,
	}, nil
}

type recoverFlags struct {
	endpoint   string
	secretPath string
	configPath string
}

func getStateDiskKeyFunc(masterKey, salt []byte) func(uuid string) ([]byte, error) {
	return func(uuid string) ([]byte, error) {
		return crypto.DeriveKey(masterKey, salt, []byte(crypto.HKDFInfoPrefix+uuid), crypto.StateDiskKeyLength)
	}
}
