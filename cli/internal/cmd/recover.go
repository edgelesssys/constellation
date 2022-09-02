/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"context"
	"fmt"
	"io"

	"github.com/edgelesssys/constellation/cli/internal/cloudcmd"
	"github.com/edgelesssys/constellation/cli/internal/proto"
	"github.com/edgelesssys/constellation/internal/atls"
	"github.com/edgelesssys/constellation/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/edgelesssys/constellation/internal/file"
	"github.com/edgelesssys/constellation/internal/state"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

type recoveryClient interface {
	Connect(endpoint string, validators atls.Validator) error
	Recover(ctx context.Context, masterSecret, salt []byte) error
	io.Closer
}

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
	recoveryClient := &proto.RecoverClient{}
	defer recoveryClient.Close()
	return recover(cmd, fileHandler, recoveryClient)
}

func recover(cmd *cobra.Command, fileHandler file.Handler, recoveryClient recoveryClient) error {
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

	config, err := readConfig(cmd.OutOrStdout(), fileHandler, flags.configPath, provider)
	if err != nil {
		return fmt.Errorf("reading and validating config: %w", err)
	}

	validators, err := cloudcmd.NewValidator(provider, config)
	if err != nil {
		return err
	}

	if err := recoveryClient.Connect(flags.endpoint, validators.V(cmd)); err != nil {
		return err
	}

	if err := recoveryClient.Recover(cmd.Context(), masterSecret.Key, masterSecret.Salt); err != nil {
		return err
	}

	cmd.Println("Pushed recovery key.")
	return nil
}

func parseRecoverFlags(cmd *cobra.Command) (recoverFlags, error) {
	endpoint, err := cmd.Flags().GetString("endpoint")
	if err != nil {
		return recoverFlags{}, fmt.Errorf("parsing endpoint argument: %w", err)
	}
	endpoint, err = addPortIfMissing(endpoint, constants.BootstrapperPort)
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
