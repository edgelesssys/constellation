/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/
package main

import (
	"errors"
	"fmt"
	"path"

	"github.com/edgelesssys/constellation/v2/internal/api/attestationconfigapi"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/staticupload"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// newDeleteCmd creates the delete command.
func newDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete {azure|aws} {snp-report|guest-firmware} <version>",
		Short:   "Delete an object from the attestationconfig API",
		Long:    "Delete a specific object version from the config api. <version> is the name of the object to delete (without .json suffix)",
		Example: "COSIGN_PASSWORD=$CPW COSIGN_PRIVATE_KEY=$CKEY cli delete azure snp-report 1.0.0",
		Args:    cobra.MatchAll(cobra.ExactArgs(3), isCloudProvider(0), isValidKind(1)),
		PreRunE: envCheck,
		RunE:    runDelete,
	}

	recursivelyCmd := &cobra.Command{
		Use:     "recursive {azure|aws}",
		Short:   "delete all objects from the API path constellation/v1/attestation/<csp>",
		Long:    "Delete all objects from the API path constellation/v1/attestation/<csp>",
		Example: "COSIGN_PASSWORD=$CPW COSIGN_PRIVATE_KEY=$CKEY cli delete recursive azure",
		Args:    cobra.MatchAll(cobra.ExactArgs(1), isCloudProvider(0)),
		RunE:    runRecursiveDelete,
	}

	cmd.AddCommand(recursivelyCmd)

	return cmd
}

func runDelete(cmd *cobra.Command, args []string) (retErr error) {
	log := logger.New(logger.PlainLog, zap.DebugLevel).Named("attestationconfigapi")

	deleteCfg, err := newDeleteConfig(cmd, ([3]string)(args[:3]))
	if err != nil {
		return fmt.Errorf("creating delete config: %w", err)
	}

	cfg := staticupload.Config{
		Bucket:         deleteCfg.bucket,
		Region:         deleteCfg.region,
		DistributionID: deleteCfg.distribution,
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

	switch deleteCfg.provider {
	case cloudprovider.AWS:
		return deleteAWS(cmd.Context(), client, deleteCfg)
	case cloudprovider.Azure:
		return deleteAzure(cmd.Context(), client, deleteCfg)
	default:
		return fmt.Errorf("unsupported cloud provider: %s", deleteCfg.provider)
	}
}

func runRecursiveDelete(cmd *cobra.Command, args []string) (retErr error) {
	// newDeleteConfig expects 3 args, so we pass "all" for the version argument and "snp-report" as kind.
	args = append(args, "snp-report")
	args = append(args, "all")
	deleteCfg, err := newDeleteConfig(cmd, ([3]string)(args[:3]))
	if err != nil {
		return fmt.Errorf("creating delete config: %w", err)
	}

	log := logger.New(logger.PlainLog, zap.DebugLevel).Named("attestationconfigapi")
	client, closeFn, err := staticupload.New(cmd.Context(), staticupload.Config{
		Bucket:         deleteCfg.bucket,
		Region:         deleteCfg.region,
		DistributionID: deleteCfg.distribution,
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

	var deletePath string
	switch deleteCfg.provider {
	case cloudprovider.AWS:
		deletePath = path.Join(attestationconfigapi.AttestationURLPath, variant.AWSSEVSNP{}.String())
	case cloudprovider.Azure:
		deletePath = path.Join(attestationconfigapi.AttestationURLPath, variant.AzureSEVSNP{}.String())
	default:
		return fmt.Errorf("unsupported cloud provider: %s", deleteCfg.provider)
	}

	return deleteRecursive(cmd.Context(), deletePath, client, deleteCfg)
}

type deleteConfig struct {
	provider        cloudprovider.Provider
	kind            objectKind
	version         string
	region          string
	bucket          string
	url             string
	distribution    string
	cosignPublicKey string
}

func newDeleteConfig(cmd *cobra.Command, args [3]string) (deleteConfig, error) {
	region, err := cmd.Flags().GetString("region")
	if err != nil {
		return deleteConfig{}, fmt.Errorf("getting region: %w", err)
	}

	bucket, err := cmd.Flags().GetString("bucket")
	if err != nil {
		return deleteConfig{}, fmt.Errorf("getting bucket: %w", err)
	}

	testing, err := cmd.Flags().GetBool("testing")
	if err != nil {
		return deleteConfig{}, fmt.Errorf("getting testing flag: %w", err)
	}
	apiCfg := getAPIEnvironment(testing)

	provider := cloudprovider.FromString(args[0])
	kind := kindFromString(args[1])
	version := args[2]

	return deleteConfig{
		provider:        provider,
		kind:            kind,
		version:         version,
		region:          region,
		bucket:          bucket,
		url:             apiCfg.url,
		distribution:    apiCfg.distribution,
		cosignPublicKey: apiCfg.cosignPublicKey,
	}, nil
}
