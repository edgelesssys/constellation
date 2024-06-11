/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
package client contains code to manage CVM versions in Constellation's CDN API.
It is used to upload and delete "latest" versions for AMD SEV-SNP and Intel TDX.
*/
package client

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/api/attestationconfigapi"
	apiclient "github.com/edgelesssys/constellation/v2/internal/api/client"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/sigstore"

	"github.com/edgelesssys/constellation/v2/internal/staticupload"
)

// VersionFormat is the format of the version name in the S3 bucket.
const VersionFormat = "2006-01-02-15-04"

// Client manages (modifies) the version information for the attestation variants.
type Client struct {
	s3Client        *apiclient.Client
	s3ClientClose   func(ctx context.Context) error
	bucketID        string
	signer          sigstore.Signer
	cacheWindowSize int
}

// New returns a new Client.
func New(ctx context.Context, cfg staticupload.Config, cosignPwd, privateKey []byte, dryRun bool, versionWindowSize int, log *slog.Logger) (*Client, apiclient.CloseFunc, error) {
	s3Client, clientClose, err := apiclient.NewClient(ctx, cfg.Region, cfg.Bucket, cfg.DistributionID, dryRun, log)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create s3 storage: %w", err)
	}

	repo := &Client{
		s3Client:        s3Client,
		s3ClientClose:   clientClose,
		signer:          sigstore.NewSigner(cosignPwd, privateKey),
		bucketID:        cfg.Bucket,
		cacheWindowSize: versionWindowSize,
	}
	return repo, clientClose, nil
}

// uploadSEVSNPVersion uploads the latest version numbers of the SEVSNP. Then version name is the UTC timestamp of the date. The /list entry stores the version name + .json suffix.
func (a Client) uploadSEVSNPVersion(ctx context.Context, attestation variant.Variant, version attestationconfigapi.SEVSNPVersion, date time.Time) error {
	versions, err := a.List(ctx, attestation)
	if err != nil {
		return fmt.Errorf("fetch version list: %w", err)
	}
	ops := a.constructUploadCmd(attestation, version, versions, date)

	return executeAllCmds(ctx, a.s3Client, ops)
}

// DeleteSEVSNPVersion deletes the given version (without .json suffix) from the API.
func (a Client) DeleteSEVSNPVersion(ctx context.Context, attestation variant.Variant, versionStr string) error {
	versions, err := a.List(ctx, attestation)
	if err != nil {
		return fmt.Errorf("fetch version list: %w", err)
	}

	ops, err := a.deleteSEVSNPVersion(versions, versionStr)
	if err != nil {
		return err
	}
	return executeAllCmds(ctx, a.s3Client, ops)
}

// List returns the list of versions for the given attestation variant.
func (a Client) List(ctx context.Context, attestation variant.Variant) (attestationconfigapi.SEVSNPVersionList, error) {
	if !attestation.Equal(variant.AzureSEVSNP{}) &&
		!attestation.Equal(variant.AWSSEVSNP{}) &&
		!attestation.Equal(variant.GCPSEVSNP{}) {
		return attestationconfigapi.SEVSNPVersionList{}, fmt.Errorf("unsupported attestation variant: %s", attestation)
	}

	versions, err := apiclient.Fetch(ctx, a.s3Client, attestationconfigapi.SEVSNPVersionList{Variant: attestation})
	if err != nil {
		var notFoundErr *apiclient.NotFoundError
		if errors.As(err, &notFoundErr) {
			return attestationconfigapi.SEVSNPVersionList{Variant: attestation}, nil
		}
		return attestationconfigapi.SEVSNPVersionList{}, err
	}

	versions.Variant = attestation

	return versions, nil
}

func (a Client) deleteSEVSNPVersion(versions attestationconfigapi.SEVSNPVersionList, versionStr string) (ops []crudCmd, err error) {
	versionStr = versionStr + ".json"
	ops = append(ops, deleteCmd{
		apiObject: attestationconfigapi.SEVSNPVersionAPI{
			Variant: versions.Variant,
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

func (a Client) constructUploadCmd(attestation variant.Variant, version attestationconfigapi.SEVSNPVersion, versionNames attestationconfigapi.SEVSNPVersionList, date time.Time) []crudCmd {
	if !attestation.Equal(versionNames.Variant) {
		return nil
	}

	dateStr := date.Format(VersionFormat) + ".json"
	var res []crudCmd

	res = append(res, putCmd{
		apiObject: attestationconfigapi.SEVSNPVersionAPI{Version: dateStr, Variant: attestation, SEVSNPVersion: version},
		signer:    a.signer,
	})

	versionNames.AddVersion(dateStr)

	res = append(res, putCmd{
		apiObject: versionNames,
		signer:    a.signer,
	})

	return res
}

func removeVersion(list attestationconfigapi.SEVSNPVersionList, versionStr string) (removedVersions attestationconfigapi.SEVSNPVersionList, err error) {
	versions := list.List
	for i, v := range versions {
		if v == versionStr {
			if i == len(versions)-1 {
				removedVersions = attestationconfigapi.SEVSNPVersionList{List: versions[:i], Variant: list.Variant}
			} else {
				removedVersions = attestationconfigapi.SEVSNPVersionList{List: append(versions[:i], versions[i+1:]...), Variant: list.Variant}
			}
			return removedVersions, nil
		}
	}
	return attestationconfigapi.SEVSNPVersionList{}, fmt.Errorf("version %s not found in list %v", versionStr, versions)
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
