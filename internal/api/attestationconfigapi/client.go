/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/
package attestationconfigapi

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	apiclient "github.com/edgelesssys/constellation/v2/internal/api/client"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/sigstore"
	"github.com/edgelesssys/constellation/v2/internal/staticupload"
)

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

// UploadAzureSEVSNP uploads the latest version numbers of the Azure SEVSNP.
func (a Client) UploadAzureSEVSNP(ctx context.Context, version AzureSEVSNPVersion, date time.Time) error {
	versions, err := a.List(ctx, variant.AzureSEVSNP{})
	if err != nil {
		return fmt.Errorf("fetch version list: %w", err)
	}
	ops, err := a.uploadAzureSEVSNP(version, versions, date)
	if err != nil {
		return err
	}
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

// List returns the list of versions for the given attestation type.
func (a Client) List(ctx context.Context, attestation variant.Variant) ([]string, error) {
	if attestation.Equal(variant.AzureSEVSNP{}) {
		versions, err := apiclient.Fetch(ctx, a.s3Client, AzureSEVSNPVersionList{})
		if err != nil {
			return nil, err
		}
		return versions, nil
	}
	return nil, fmt.Errorf("unsupported attestation type: %s", attestation)
}

func (a Client) deleteAzureSEVSNPVersion(versions AzureSEVSNPVersionList, versionStr string) (ops []crudCmd, err error) {
	versionStr = versionStr + ".json"
	ops = append(ops, deleteCmd{
		apiObject: AzureSEVSNPVersionAPI{
			Version: versionStr,
		},
	})

	ops = append(ops, deleteCmd{
		apiObject: AzureSEVSNPVersionSignature{
			Version: versionStr,
		},
	})

	removedVersions, err := removeVersion(versions, versionStr)
	if err != nil {
		return nil, err
	}
	ops = append(ops, putCmd{
		apiObject: removedVersions,
	})
	return ops, nil
}

func (a Client) uploadAzureSEVSNP(versions AzureSEVSNPVersion, versionNames []string, date time.Time) (res []crudCmd, err error) {
	dateStr := date.Format("2006-01-02-15-04") + ".json"

	res = append(res, putCmd{AzureSEVSNPVersionAPI{Version: dateStr, AzureSEVSNPVersion: versions}})

	versionBytes, err := json.Marshal(versions)
	if err != nil {
		return res, err
	}
	signature, err := a.createSignature(versionBytes, dateStr)
	if err != nil {
		return res, err
	}
	res = append(res, putCmd{signature})
	newVersions := addVersion(versionNames, dateStr)
	res = append(res, putCmd{AzureSEVSNPVersionList(newVersions)})
	return
}

func (a Client) createSignature(content []byte, dateStr string) (res AzureSEVSNPVersionSignature, err error) {
	signature, err := a.signer.Sign(content)
	if err != nil {
		return res, fmt.Errorf("sign version file: %w", err)
	}
	return AzureSEVSNPVersionSignature{
		Signature: signature,
		Version:   dateStr,
	}, nil
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
	return apiclient.Delete(ctx, c, d.apiObject)
}

type putCmd struct {
	apiObject apiclient.APIObject
}

func (p putCmd) Execute(ctx context.Context, c *apiclient.Client) error {
	return apiclient.Update(ctx, c, p.apiObject)
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
	sort.Sort(sort.Reverse(sort.StringSlice(versions)))
	return versions
}
