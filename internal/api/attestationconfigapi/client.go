/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/
package attestationconfigapi

import (
	"context"
	"fmt"
	"time"

	apiclient "github.com/edgelesssys/constellation/v2/internal/api/client"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/sigstore"
	"github.com/edgelesssys/constellation/v2/internal/staticupload"
)

// VersionFormat is the format of the version name in the S3 bucket.
const VersionFormat = "2006-01-02-15-04"

// Client manages (modifies) the version information for the attestation variants.
type Client struct {
	s3Client      *apiclient.Client
	s3ClientClose func(ctx context.Context) error
	bucketID      string
	signer        sigstore.Signer
}

// NewClient returns a new Client.
func NewClient(ctx context.Context, cfg staticupload.Config, cosignPwd, privateKey []byte, dryRun bool, log *logger.Logger) (*Client, apiclient.CloseFunc, error) {
	s3Client, clientClose, err := apiclient.NewClient(ctx, cfg.Region, cfg.Bucket, cfg.DistributionID, dryRun, log)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create s3 storage: %w", err)
	}

	repo := &Client{
		s3Client:      s3Client,
		s3ClientClose: clientClose,
		signer:        sigstore.NewSigner(cosignPwd, privateKey),
		bucketID:      cfg.Bucket,
	}
	return repo, clientClose, nil
}

// UploadAzureSEVSNPVersion uploads the latest version numbers of the Azure SEVSNP. Then version name is the UTC timestamp of the date. The /list entry stores the version name + .json suffix.
func (a Client) UploadAzureSEVSNPVersion(ctx context.Context, version AzureSEVSNPVersion, date time.Time) error {
	versions, err := a.List(ctx, variant.AzureSEVSNP{})
	if err != nil {
		return fmt.Errorf("fetch version list: %w", err)
	}
	ops := a.constructUploadCmd(version, versions, date)

	return executeAllCmds(ctx, a.s3Client, ops)
}

// DeleteAzureSEVSNPVersion deletes the given version (without .json suffix) from the API.
func (a Client) DeleteAzureSEVSNPVersion(ctx context.Context, versionStr string) error {
	versions, err := a.List(ctx, variant.AzureSEVSNP{})
	if err != nil {
		return fmt.Errorf("fetch version list: %w", err)
	}
	ops, err := a.deleteAzureSEVSNPVersion(versions, versionStr)
	if err != nil {
		return err
	}
	return executeAllCmds(ctx, a.s3Client, ops)
}

// List returns the list of versions for the given attestation variant.
func (a Client) List(ctx context.Context, attestation variant.Variant) ([]string, error) {
	if attestation.Equal(variant.AzureSEVSNP{}) {
		versions, err := apiclient.Fetch(ctx, a.s3Client, AzureSEVSNPVersionList{})
		if err != nil {
			return nil, err
		}
		return versions, nil
	}
	return nil, fmt.Errorf("unsupported attestation variant: %s", attestation)
}

func (a Client) deleteAzureSEVSNPVersion(versions AzureSEVSNPVersionList, versionStr string) (ops []crudCmd, err error) {
	versionStr = versionStr + ".json"
	ops = append(ops, deleteCmd{
		apiObject: AzureSEVSNPVersionAPI{
			Version: versionStr,
		},
	})

	removedVersions, err := removeVersion(versions, versionStr)
	if err != nil {
		return nil, err
	}
	ops = append(ops, putCmd{
		apiObject: removedVersions,
		signer:    a.signer,
	})
	return ops, nil
}

func (a Client) constructUploadCmd(versions AzureSEVSNPVersion, versionNames []string, date time.Time) []crudCmd {
	dateStr := date.Format(VersionFormat) + ".json"
	var res []crudCmd

	res = append(res, putCmd{
		apiObject: AzureSEVSNPVersionAPI{Version: dateStr, AzureSEVSNPVersion: versions},
		signer:    a.signer,
	})

	newVersions := addVersion(versionNames, dateStr)
	res = append(res, putCmd{
		apiObject: AzureSEVSNPVersionList(newVersions),
		signer:    a.signer,
	})

	return res
}

func removeVersion(versions AzureSEVSNPVersionList, versionStr string) (removedVersions AzureSEVSNPVersionList, err error) {
	for i, v := range versions {
		if v == versionStr {
			if i == len(versions)-1 {
				removedVersions = versions[:i]
			} else {
				removedVersions = append(versions[:i], versions[i+1:]...)
			}
			return removedVersions, nil
		}
	}
	return nil, fmt.Errorf("version %s not found in list %v", versionStr, versions)
}

type crudCmd interface {
	Execute(ctx context.Context, c *apiclient.Client) error
}

type deleteCmd struct {
	apiObject apiclient.APIObject
}

func (d deleteCmd) Execute(ctx context.Context, c *apiclient.Client) error {
	return apiclient.DeleteWithSignature(ctx, c, d.apiObject)
}

type putCmd struct {
	apiObject apiclient.APIObject
	signer    sigstore.Signer
}

func (p putCmd) Execute(ctx context.Context, c *apiclient.Client) error {
	return apiclient.SignAndUpdate(ctx, c, p.apiObject, p.signer)
}

func executeAllCmds(ctx context.Context, client *apiclient.Client, cmds []crudCmd) error {
	for _, cmd := range cmds {
		if err := cmd.Execute(ctx, client); err != nil {
			return fmt.Errorf("execute operation %+v: %w", cmd, err)
		}
	}
	return nil
}

func addVersion(versions []string, newVersion string) []string {
	versions = append(versions, newVersion)
	versions = variant.RemoveDuplicate(versions)
	SortAzureSEVSNPVersionList(versions)
	return versions
}
