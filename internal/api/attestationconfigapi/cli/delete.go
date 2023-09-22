/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/
package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
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

	distribution, err := cmd.Flags().GetString("distribution")
	if err != nil {
		return fmt.Errorf("getting distribution: %w", err)
	}

	cfg := staticupload.Config{
		Bucket:         bucket,
		Region:         region,
		DistributionID: distribution,
	}
	client, clientClose, err := attestationconfigapi.NewClient(cmd.Context(), cfg, []byte(cosignPwd), []byte(privateKey), false, log)
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

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		return
	}

	// Create an S3 client.
	svc := s3.New(sess)

	path := "constellation/v1/attestation/azure-sev-snp"
	// List all objects in the path.
	resp, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(path),
	})
	if err != nil {
		fmt.Println("Error listing objects:", err)
		os.Exit(1)
	}

	// Delete all objects in the path.
	var keys []*s3.ObjectIdentifier
	for _, obj := range resp.Contents {
		keys = append(keys, &s3.ObjectIdentifier{
			Key: obj.Key,
		})
	}
	if len(keys) > 0 {
		_, err = svc.DeleteObjects(&s3.DeleteObjectsInput{
			Bucket: aws.String(bucket),
			Delete: &s3.Delete{
				Objects: keys,
				Quiet:   aws.Bool(true),
			},
		})
		if err != nil {
			return err
		}
	}
	return nil
}
