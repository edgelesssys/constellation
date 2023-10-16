/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/
package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/edgelesssys/constellation/v2/internal/api/attestationconfigapi"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/staticupload"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// newDeleteCmd creates the delete command.
func newDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete",
		Short:   "delete a specific version from the config api",
		PreRunE: envCheck,
		RunE:    runDelete,
	}
	cmd.Flags().StringP("version", "v", "", "Name of the version to delete (without .json suffix)")
	must(cmd.MarkFlagRequired("version"))

	recursivelyCmd := &cobra.Command{
		Use:   "recursive",
		Short: "delete all objects from the API path",
		RunE:  runRecursiveDelete,
	}
	cmd.AddCommand(recursivelyCmd)
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

func runDelete(cmd *cobra.Command, _ []string) (retErr error) {
	log := logger.New(logger.PlainLog, zap.DebugLevel).Named("attestationconfigapi")

	region, err := cmd.Flags().GetString("region")
	if err != nil {
		return fmt.Errorf("getting region: %w", err)
	}

	bucket, err := cmd.Flags().GetString("bucket")
	if err != nil {
		return fmt.Errorf("getting bucket: %w", err)
	}

	testing, err := cmd.Flags().GetBool("testing")
	if err != nil {
		return fmt.Errorf("getting testing flag: %w", err)
	}
	apiCfg := getAPIEnvironment(testing)

	cfg := staticupload.Config{
		Bucket:         bucket,
		Region:         region,
		DistributionID: apiCfg.distribution,
	}
	client, clientClose, err := attestationconfigapi.NewClient(cmd.Context(), cfg,
		[]byte(cosignPwd), []byte(privateKey), false, 1, log)
	if err != nil {
		return fmt.Errorf("create attestation client: %w", err)
	}
	defer func() {
		err := clientClose(cmd.Context())
		if err != nil {
			retErr = errors.Join(retErr, fmt.Errorf("failed to invalidate cache: %w", err))
		}
	}()

	deleteCmd := deleteCmd{
		attestationClient: client,
	}
	return deleteCmd.delete(cmd)
}

func runRecursiveDelete(cmd *cobra.Command, _ []string) (retErr error) {
	region, err := cmd.Flags().GetString("region")
	if err != nil {
		return fmt.Errorf("getting region: %w", err)
	}

	bucket, err := cmd.Flags().GetString("bucket")
	if err != nil {
		return fmt.Errorf("getting bucket: %w", err)
	}

	testing, err := cmd.Flags().GetBool("testing")
	if err != nil {
		return fmt.Errorf("getting testing flag: %w", err)
	}
	apiCfg := getAPIEnvironment(testing)

	log := logger.New(logger.PlainLog, zap.DebugLevel).Named("attestationconfigapi")
	client, closeFn, err := staticupload.New(cmd.Context(), staticupload.Config{
		Bucket:         bucket,
		Region:         region,
		DistributionID: apiCfg.distribution,
	}, log)
	if err != nil {
		return fmt.Errorf("create static upload client: %w", err)
	}
	defer func() {
		err := closeFn(cmd.Context())
		if err != nil {
			retErr = errors.Join(retErr, fmt.Errorf("failed to close client: %w", err))
		}
	}()
	path := "constellation/v1/attestation/azure-sev-snp"
	resp, err := client.ListObjectsV2(cmd.Context(), &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(path),
	})
	if err != nil {
		return err
	}

	// Delete all objects in the path.
	objIDs := make([]s3types.ObjectIdentifier, len(resp.Contents))
	for i, obj := range resp.Contents {
		objIDs[i] = s3types.ObjectIdentifier{Key: obj.Key}
	}
	if len(objIDs) > 0 {
		_, err = client.DeleteObjects(cmd.Context(), &s3.DeleteObjectsInput{
			Bucket: aws.String(bucket),
			Delete: &s3types.Delete{
				Objects: objIDs,
				Quiet:   true,
			},
		})
		if err != nil {
			return err
		}
	}
	return nil
}
