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
	"github.com/edgelesssys/constellation/v2/cli/internal/clusterid"
	"github.com/edgelesssys/constellation/v2/disk-mapper/recoverproto"
	"github.com/edgelesssys/constellation/v2/internal/atls"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/crypto"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/grpc/dialer"
	grpcRetry "github.com/edgelesssys/constellation/v2/internal/grpc/retry"
	"github.com/edgelesssys/constellation/v2/internal/kms/uri"
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
	newDialer := func(validator atls.Validator) *dialer.Dialer {
		return dialer.New(nil, validator, &net.Dialer{})
	}
	r := &recoverCmd{log: log}
	return r.recover(cmd, fileHandler, 5*time.Second, &recoverDoer{log: r.log}, newDialer)
}

func (r *recoverCmd) recover(
	cmd *cobra.Command, fileHandler file.Handler, interval time.Duration,
	doer recoverDoerInterface, newDialer func(validator atls.Validator) *dialer.Dialer,
) error {
	flags, err := r.parseRecoverFlags(cmd, fileHandler)
	if err != nil {
		return err
	}
	r.log.Debugf("Using flags: %+v", flags)

	var masterSecret uri.MasterSecret
	r.log.Debugf("Loading master secret file from %s", flags.secretPath)
	if err := fileHandler.ReadJSON(flags.secretPath, &masterSecret); err != nil {
		return err
	}

	r.log.Debugf("Loading configuration file from %q", flags.configPath)
	conf, err := config.New(fileHandler, flags.configPath, flags.force)
	var configValidationErr *config.ValidationError
	if errors.As(err, &configValidationErr) {
		cmd.PrintErrln(configValidationErr.LongMessage())
	}
	if err != nil {
		return err
	}

	provider := conf.GetProvider()
	r.log.Debugf("Got provider %s", provider.String())
	if provider == cloudprovider.Azure {
		interval = 20 * time.Second // Azure LB takes a while to remove unhealthy instances
	}

	conf.UpdateMAAURL(flags.maaURL)
	r.log.Debugf("Creating aTLS Validator for %s", conf.GetAttestationConfig().GetVariant())
	validator, err := cloudcmd.NewValidator(cmd, conf.GetAttestationConfig(), r.log)
	if err != nil {
		return fmt.Errorf("creating new validator: %w", err)
	}
	r.log.Debugf("Created a new validator")
	doer.setDialer(newDialer(validator), flags.endpoint)
	r.log.Debugf("Set dialer for endpoint %s", flags.endpoint)
	doer.setURIs(masterSecret.EncodeToURI(), uri.NoStoreURI)
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
			var retry bool
			// retry transient GCP LB errors
			if grpcRetry.LoadbalancerIsNotReady(err) {
				retry = true
			} else {
				// retry connection errors once
				// this is necessary because Azure's LB takes a while to remove unhealthy instances
				once.Do(func() {
					retry = grpcRetry.ServiceIsUnavailable(err)
				})
			}

			r.log.Debugf("Encountered error (retriable: %t): %s", retry, err)
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
	setURIs(kmsURI, storageURI string)
}

type recoverDoer struct {
	dialer     grpcDialer
	endpoint   string
	kmsURI     string // encodes masterSecret
	storageURI string
	log        debugLog
}

// Do performs the recover streaming rpc.
func (d *recoverDoer) Do(ctx context.Context) (retErr error) {
	conn, err := d.dialer.Dial(ctx, d.endpoint)
	if err != nil {
		return fmt.Errorf("dialing recovery server: %w", err)
	}
	d.log.Debugf("Dialed recovery server")
	defer conn.Close()

	protoClient := recoverproto.NewAPIClient(conn)
	d.log.Debugf("Created protoClient")

	req := &recoverproto.RecoverMessage{
		KmsUri:     d.kmsURI,
		StorageUri: d.storageURI,
	}

	_, err = protoClient.Recover(ctx, req)
	if err != nil {
		return fmt.Errorf("calling recover: %w", err)
	}

	d.log.Debugf("Received confirmation")
	return nil
}

func (d *recoverDoer) setDialer(dialer grpcDialer, endpoint string) {
	d.dialer = dialer
	d.endpoint = endpoint
}

func (d *recoverDoer) setURIs(kmsURI, storageURI string) {
	d.kmsURI = kmsURI
	d.storageURI = storageURI
}

type recoverFlags struct {
	endpoint   string
	secretPath string
	configPath string
	maaURL     string
	force      bool
}

func (r *recoverCmd) parseRecoverFlags(cmd *cobra.Command, fileHandler file.Handler) (recoverFlags, error) {
	var idFile clusterid.File
	if err := fileHandler.ReadJSON(constants.ClusterIDsFileName, &idFile); err != nil && !errors.Is(err, afero.ErrFileNotFound) {
		return recoverFlags{}, err
	}

	endpoint, err := cmd.Flags().GetString("endpoint")
	r.log.Debugf("Endpoint flag is %s", endpoint)
	if err != nil {
		return recoverFlags{}, fmt.Errorf("parsing endpoint argument: %w", err)
	}
	if endpoint == "" {
		endpoint = idFile.IP
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
	r.log.Debugf("Configuration path flag is %s", configPath)

	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return recoverFlags{}, fmt.Errorf("parsing force argument: %w", err)
	}

	return recoverFlags{
		endpoint:   endpoint,
		secretPath: masterSecretPath,
		configPath: configPath,
		maaURL:     idFile.AttestationURL,
		force:      force,
	}, nil
}

func getStateDiskKeyFunc(masterKey, salt []byte) func(uuid string) ([]byte, error) {
	return func(uuid string) ([]byte, error) {
		return crypto.DeriveKey(masterKey, salt, []byte(crypto.DEKPrefix+uuid), crypto.StateDiskKeyLength)
	}
}
