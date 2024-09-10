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
	"path"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/edgelesssys/constellation/v2/api/attestationconfig"
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

	log *slog.Logger
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
		log:             log,
	}
	return repo, clientClose, nil
}

// DeleteVersion deletes the given version (without .json suffix) from the API.
func (c Client) DeleteVersion(ctx context.Context, attestation variant.Variant, versionStr string) error {
	versions, err := c.List(ctx, attestation)
	if err != nil {
		return fmt.Errorf("fetch version list: %w", err)
	}

	ops, err := c.deleteVersion(versions, versionStr)
	if err != nil {
		return err
	}
	return executeAllCmds(ctx, c.s3Client, ops)
}

// List returns the list of versions for the given attestation variant.
func (c Client) List(ctx context.Context, attestation variant.Variant) (attestationconfig.List, error) {
	versions, err := apiclient.Fetch(ctx, c.s3Client, attestationconfig.List{Variant: attestation})
	if err != nil {
		var notFoundErr *apiclient.NotFoundError
		if errors.As(err, &notFoundErr) {
			return attestationconfig.List{Variant: attestation}, nil
		}
		return attestationconfig.List{}, err
	}

	versions.Variant = attestation

	return versions, nil
}

func (c Client) deleteVersion(versions attestationconfig.List, versionStr string) (ops []crudCmd, err error) {
	versionStr = versionStr + ".json"
	ops = append(ops, deleteCmd{
		apiObject: attestationconfig.Entry{
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
		signer:    c.signer,
	})
	return ops, nil
}

func (c Client) listCachedVersions(ctx context.Context, attestation variant.Variant) ([]string, error) {
	list, err := c.s3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(c.bucketID),
		Prefix: aws.String(reportVersionDir(attestation)),
	})
	if err != nil {
		return nil, fmt.Errorf("list objects: %w", err)
	}

	var dates []string
	for _, obj := range list.Contents {
		fileName := path.Base(*obj.Key)

		// The cache contains signature and json files
		// We only want the json files
		if date, ok := strings.CutSuffix(fileName, ".json"); ok {
			dates = append(dates, date)
		}
	}
	return dates, nil
}

func removeVersion(list attestationconfig.List, versionStr string) (removedVersions attestationconfig.List, err error) {
	versions := list.List
	for i, v := range versions {
		if v == versionStr {
			if i == len(versions)-1 {
				removedVersions = attestationconfig.List{List: versions[:i], Variant: list.Variant}
			} else {
				removedVersions = attestationconfig.List{List: append(versions[:i], versions[i+1:]...), Variant: list.Variant}
			}
			return removedVersions, nil
		}
	}
	return attestationconfig.List{}, fmt.Errorf("version %s not found in list %v", versionStr, versions)
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
