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
	cmd.PersistentFlags().StringP("version", "v", "", "Name of the version to delete (without .json suffix)")
	must(enforceRequiredFlags(cmd, "version"))
	return cmd
}

type deleteCmd struct {
	attestationclient client.Client
	closefn           client.CloseFunc
}

func (d deleteCmd) delete(ctx context.Context, cmd *cobra.Command) error {
	defer func() {
		if d.closefn != nil {
			if err := d.closefn(ctx); err != nil {
				fmt.Printf("close client: %s\n", err.Error())
			}
		}
	}()
	version := cmd.Flag("version").Value.String()
	return d.attestationclient.DeleteAzureSEVSNVersion(ctx, version)
}

func runDelete(cmd *cobra.Command, _ []string) error {
	ctx := context.Background()
	cfg := staticupload.Config{
		Bucket: awsBucket,
		Region: awsRegion,
	}
	repo, closefn, err := client.New(ctx, cfg, []byte(cosignPwd), []byte(privateKey))
	if err != nil {
		return fmt.Errorf("create attestation client: %w", err)
	}
	deleteCmd := deleteCmd{
		attestationclient: repo,
		closefn:           closefn,
	}
	return deleteCmd.delete(ctx, cmd)
}
