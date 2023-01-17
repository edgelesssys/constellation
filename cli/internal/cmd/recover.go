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
	"sync"
	"time"

	"github.com/edgelesssys/constellation/v2/cli/internal/cloudcmd"
	"github.com/edgelesssys/constellation/v2/disk-mapper/recoverproto"
	"github.com/edgelesssys/constellation/v2/internal/attestation"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/crypto"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/grpc/dialer"
	grpcRetry "github.com/edgelesssys/constellation/v2/internal/grpc/retry"
	"github.com/edgelesssys/constellation/v2/internal/retry"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

// NewRecoverCmd returns a new cobra.Command for the recover command.
func NewRecoverCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "recover",
		Short: "Recover a completely stopped Constellation cluster",
		Long: "Recover a Constellation cluster by sending a recovery key to an instance in the boot stage.\n\n" +
			"This is only required if instances restart without other instances available for bootstrapping.",
		Args: cobra.ExactArgs(0),
		RunE: runRecover,
	}
	cmd.Flags().StringP("endpoint", "e", "", "endpoint of the instance, passed as HOST[:PORT]")
	cmd.Flags().String("master-secret", constants.MasterSecretFilename, "path to master secret file")
	return cmd
}

type recoverCmd struct {
	log debugLog
}

func runRecover(cmd *cobra.Command, _ []string) error {
	log, err := newCLILogger(cmd)
	if err != nil {
		return fmt.Errorf("creating logger: %w", err)
	}
	defer log.Sync()
	fileHandler := file.NewHandler(afero.NewOsFs())
	newDialer := func(validator *cloudcmd.Validator) *dialer.Dialer {
		return dialer.New(nil, validator.V(cmd), &net.Dialer{})
	}
	r := &recoverCmd{log: log}
	return r.recover(cmd, fileHandler, 5*time.Second, &recoverDoer{log: r.log}, newDialer)
}

func (r *recoverCmd) recover(
	cmd *cobra.Command, fileHandler file.Handler, interval time.Duration,
	doer recoverDoerInterface, newDialer func(validator *cloudcmd.Validator) *dialer.Dialer,
) error {
	flags, err := r.parseRecoverFlags(cmd, fileHandler)
	if err != nil {
		return err
	}
	r.log.Debugf("Using flags: %+v", flags)

	var masterSecret masterSecret
	r.log.Debugf("Loading master secret file from %s", flags.secretPath)
	if err := fileHandler.ReadJSON(flags.secretPath, &masterSecret); err != nil {
		return err
	}

	r.log.Debugf("Loading configuration file from %s", flags.configPath)
	conf, err := config.New(fileHandler, flags.configPath)
	if err != nil {
		return displayConfigValidationErrors(cmd.ErrOrStderr(), err)
	}
	provider := conf.GetProvider()
	r.log.Debugf("Got provider %s", provider.String())
	if provider == cloudprovider.Azure {
		interval = 20 * time.Second // Azure LB takes a while to remove unhealthy instances
	}

	validator, err := cloudcmd.NewValidator(provider, conf)
	if err != nil {
		return err
	}
	r.log.Debugf("Created a new validator")
	doer.setDialer(newDialer(validator), flags.endpoint)
	r.log.Debugf("Set dialer for endpoint %s", flags.endpoint)
	measurementSecret, err := attestation.DeriveMeasurementSecret(masterSecret.Key, masterSecret.Salt)
	r.log.Debugf("Derived measurementSecret")
	if err != nil {
		return err
	}
	doer.setSecrets(getStateDiskKeyFunc(masterSecret.Key, masterSecret.Salt), measurementSecret)
	r.log.Debugf("Set secrets")
	if err := r.recoverCall(cmd.Context(), cmd.OutOrStdout(), interval, doer); err != nil {
		if grpcRetry.ServiceIsUnavailable(err) {
			return nil
		}
		return fmt.Errorf("recovering cluster: %w", err)
	}
	return nil
}

func (r *recoverCmd) recoverCall(ctx context.Context, out io.Writer, interval time.Duration, doer recoverDoerInterface) error {
	var err error
	ctr := 0
	for {
		once := sync.Once{}
		retryOnceOnFailure := func(err error) bool {
			// retry transient GCP LB errors
			if grpcRetry.LoadbalancerIsNotReady(err) {
				return true
			}
			retry := false

			// retry connection errors once
			// this is necessary because Azure's LB takes a while to remove unhealthy instances
			once.Do(func() {
				retry = grpcRetry.ServiceIsUnavailable(err)
			})
			return retry
		}

		retrier := retry.NewIntervalRetrier(doer, interval, retryOnceOnFailure)
		r.log.Debugf("Created new interval retrier")
		err = retrier.Do(ctx)
		if err != nil {
			break
		}
		fmt.Fprintln(out, "Pushed recovery key.")
		ctr++
	}
	r.log.Debugf("Retry counter is %d", ctr)
	if ctr > 0 {
		fmt.Fprintf(out, "Recovered %d control-plane nodes.\n", ctr)
	} else if grpcRetry.ServiceIsUnavailable(err) {
		fmt.Fprintln(out, "No control-plane nodes in need of recovery found. Exiting.")
		return nil
	}
	return err
}

type recoverDoerInterface interface {
	Do(ctx context.Context) error
	setDialer(dialer grpcDialer, endpoint string)
	setSecrets(getDiskKey func(uuid string) ([]byte, error), measurementSecret []byte)
}

type recoverDoer struct {
	dialer            grpcDialer
	endpoint          string
	measurementSecret []byte
	getDiskKey        func(uuid string) (key []byte, err error)
	log               debugLog
}

// Do performs the recover streaming rpc.
func (d *recoverDoer) Do(ctx context.Context) (retErr error) {
	conn, err := d.dialer.Dial(ctx, d.endpoint)
	if err != nil {
		return fmt.Errorf("dialing recovery server: %w", err)
	}
	d.log.Debugf("Dialed recovery server")
	defer conn.Close()

	// set up streaming client
	protoClient := recoverproto.NewAPIClient(conn)
	d.log.Debugf("Created protoClient")
	recoverclient, err := protoClient.Recover(ctx)
	d.log.Debugf("Created recoverclient")
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}
	defer func() {
		_ = recoverclient.CloseSend()
	}()

	// send measurement secret as first message
	if err := recoverclient.Send(&recoverproto.RecoverMessage{
		Request: &recoverproto.RecoverMessage_MeasurementSecret{
			MeasurementSecret: d.measurementSecret,
		},
	}); err != nil {
		return fmt.Errorf("sending measurement secret: %w", err)
	}
	d.log.Debugf("Sent measurement secret")

	// receive disk uuid
	res, err := recoverclient.Recv()
	if err != nil {
		return fmt.Errorf("receiving disk uuid: %w", err)
	}
	d.log.Debugf("Received disk uuid")
	stateDiskKey, err := d.getDiskKey(res.DiskUuid)
	if err != nil {
		return fmt.Errorf("getting state disk key: %w", err)
	}
	d.log.Debugf("Got state disk key")

	// send disk key
	if err := recoverclient.Send(&recoverproto.RecoverMessage{
		Request: &recoverproto.RecoverMessage_StateDiskKey{
			StateDiskKey: stateDiskKey,
		},
	}); err != nil {
		return fmt.Errorf("sending state disk key: %w", err)
	}
	d.log.Debugf("Sent state disk key")

	if _, err := recoverclient.Recv(); err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("receiving confirmation: %w", err)
	}
	d.log.Debugf("Received confirmation")
	return nil
}

func (d *recoverDoer) setDialer(dialer grpcDialer, endpoint string) {
	d.dialer = dialer
	d.endpoint = endpoint
}

func (d *recoverDoer) setSecrets(getDiskKey func(string) ([]byte, error), measurementSecret []byte) {
	d.getDiskKey = getDiskKey
	d.measurementSecret = measurementSecret
}

type recoverFlags struct {
	endpoint   string
	secretPath string
	configPath string
}

func (r *recoverCmd) parseRecoverFlags(cmd *cobra.Command, fileHandler file.Handler) (recoverFlags, error) {
	endpoint, err := cmd.Flags().GetString("endpoint")
	r.log.Debugf("Endpoint flag is %s", endpoint)
	if err != nil {
		return recoverFlags{}, fmt.Errorf("parsing endpoint argument: %w", err)
	}
	if endpoint == "" {
		endpoint, err = readIPFromIDFile(fileHandler)
		if err != nil {
			return recoverFlags{}, fmt.Errorf("getting recovery endpoint: %w", err)
		}
	}
	endpoint, err = addPortIfMissing(endpoint, constants.RecoveryPort)
	if err != nil {
		return recoverFlags{}, fmt.Errorf("validating endpoint argument: %w", err)
	}
	r.log.Debugf("Endpoint value after parsing is %s", endpoint)
	masterSecretPath, err := cmd.Flags().GetString("master-secret")
	if err != nil {
		return recoverFlags{}, fmt.Errorf("parsing master-secret path argument: %w", err)
	}
	r.log.Debugf("Master secret flag is %s", masterSecretPath)
	configPath, err := cmd.Flags().GetString("config")
	if err != nil {
		return recoverFlags{}, fmt.Errorf("parsing config path argument: %w", err)
	}
	r.log.Debugf("Config path flag is %s", configPath)

	return recoverFlags{
		endpoint:   endpoint,
		secretPath: masterSecretPath,
		configPath: configPath,
	}, nil
}

func getStateDiskKeyFunc(masterKey, salt []byte) func(uuid string) ([]byte, error) {
	return func(uuid string) ([]byte, error) {
		return crypto.DeriveKey(masterKey, salt, []byte(crypto.HKDFInfoPrefix+uuid), crypto.StateDiskKeyLength)
	}
}
