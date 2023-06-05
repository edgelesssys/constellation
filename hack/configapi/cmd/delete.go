/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/
package cmd

import (
	"context"
	"fmt"

	"github.com/edgelesssys/constellation/v2/internal/api/attestationconfig/client"
	"github.com/edgelesssys/constellation/v2/internal/staticupload"
	"github.com/spf13/cobra"
)

// newDeleteCmd creates the delete command.
func newDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "delete a specific version from the config api",
		RunE:  runDelete,
	}
	cmd.Flags().StringP("version", "v", "", "Name of the version to delete (without .json suffix)")
	must(enforceRequiredFlags(cmd, "version"))
	return cmd
}

type deleteCmd struct {
	attestationClient deleteClient
	closeFn           client.CloseFunc
}

type deleteClient interface {
	DeleteAzureSEVSNPVersion(ctx context.Context, versionStr string) error
}

func (d deleteCmd) delete(cmd *cobra.Command) error {
	defer func() {
		if d.closeFn != nil {
			if err := d.closeFn(cmd.Context()); err != nil {
				fmt.Printf("close client: %s\n", err.Error())
			}
		}
	}()
	version := cmd.Flag("version").Value.String()
	return d.attestationClient.DeleteAzureSEVSNPVersion(cmd.Context(), version)
}

func runDelete(cmd *cobra.Command, _ []string) error {
	cfg := staticupload.Config{
		Bucket: awsBucket,
		Region: awsRegion,
	}
	repo, closefn, err := client.New(cmd.Context(), cfg, []byte(cosignPwd), []byte(privateKey), false, log())
	if err != nil {
		return fmt.Errorf("create attestation client: %w", err)
	}
	deleteCmd := deleteCmd{
		attestationClient: repo,
		closeFn:           closefn,
	}
	return deleteCmd.delete(cmd)
}
