/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/
package main

import (
	"context"
	"fmt"

	"github.com/edgelesssys/constellation/v2/internal/api/attestationconfigapi"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/staticupload"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// newDeleteCmd creates the delete command.
func newDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "delete a specific version from the config api",
		RunE:  runDelete,
	}
	cmd.Flags().StringP("version", "v", "", "Name of the version to delete (without .json suffix)")
	must(cmd.MarkFlagRequired("version"))
	return cmd
}

type deleteCmd struct {
	attestationClient deleteClient
}

type deleteClient interface {
	DeleteAzureSEVSNPVersion(ctx context.Context, versionStr string) error
}

func (d deleteCmd) delete(cmd *cobra.Command) error {
	version, err := cmd.Flags().GetString("version")
	if err != nil {
		return err
	}
	return d.attestationClient.DeleteAzureSEVSNPVersion(cmd.Context(), version)
}

func runDelete(cmd *cobra.Command, _ []string) error {
	log := logger.New(logger.PlainLog, zap.DebugLevel).Named("attestationconfigapi")

	region, err := cmd.Flags().GetString("region")
	if err != nil {
		return fmt.Errorf("getting region: %w", err)
	}

	bucket, err := cmd.Flags().GetString("bucket")
	if err != nil {
		return fmt.Errorf("getting bucket: %w", err)
	}

	cfg := staticupload.Config{
		Bucket: bucket,
		Region: region,
	}
	client, stop, err := attestationconfigapi.NewClient(cmd.Context(), cfg, []byte(cosignPwd), []byte(privateKey), false, log)
	if err != nil {
		return fmt.Errorf("create attestation client: %w", err)
	}
	defer func() {
		if err := stop(cmd.Context()); err != nil {
			cmd.Printf("stopping client: %s\n", err.Error())
		}
	}()
	deleteCmd := deleteCmd{
		attestationClient: client,
	}
	return deleteCmd.delete(cmd)
}
