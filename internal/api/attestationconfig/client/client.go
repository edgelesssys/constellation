/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/api/attestationconfig"
	"github.com/edgelesssys/constellation/v2/internal/api/attestationconfig/fetcher"
	apiclient "github.com/edgelesssys/constellation/v2/internal/api/client"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/sigstore"
	"github.com/edgelesssys/constellation/v2/internal/staticupload"
	"github.com/edgelesssys/constellation/v2/internal/variant"
)

// Client manages (modifies) the version information for the attestation variants.
type Client struct {
	s3Client      *apiclient.Client
	s3ClientClose func(ctx context.Context) error
	bucketID      string
	signer        sigstore.Signer
	fetcher       fetcher.AttestationConfigAPIFetcher
}

// New returns a new Client.
func New(ctx context.Context, cfg staticupload.Config, cosignPwd, privateKey []byte, dryRun bool, log *logger.Logger) (*Client, apiclient.CloseFunc, error) {
	s3Client, clientClose, err := apiclient.NewClient(ctx, cfg.Region, cfg.Bucket, cfg.DistributionID, dryRun, log)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create s3 storage: %w", err)
	}

	repo := &Client{
		s3Client:      s3Client,
		s3ClientClose: clientClose,
		signer:        sigstore.NewSigner(cosignPwd, privateKey),
		bucketID:      cfg.Bucket,
		fetcher:       fetcher.New(),
	}
	return repo, clientClose, nil
}

func (a Client) uploadAzureSEVSNP(versions attestationconfig.AzureSEVSNPVersion, versionNames []string, date time.Time) (res []putCmd, err error) {
	dateStr := date.Format("2006-01-02-15-04") + ".json"

	res = append(res, putCmd{attestationconfig.AzureSEVSNPVersionAPI{Version: dateStr, AzureSEVSNPVersion: versions}})

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
	res = append(res, putCmd{attestationconfig.AzureSEVSNPVersionList(newVersions)})
	return
}

// UploadAzureSEVSNP uploads the latest version numbers of the Azure SEVSNP.
func (a Client) UploadAzureSEVSNP(ctx context.Context, version attestationconfig.AzureSEVSNPVersion, date time.Time) error {
	variant := variant.AzureSEVSNP{}

	dateStr := date.Format("2006-01-02-15-04") + ".json"
	err := apiclient.Update(ctx, a.s3Client, attestationconfig.AzureSEVSNPVersionAPI{Version: dateStr, AzureSEVSNPVersion: version})
	if err != nil {
		return err
	}

	versionBytes, err := json.Marshal(version)
	if err != nil {
		return err
	}
	filePath := fmt.Sprintf("%s/%s/%s", constants.CDNAttestationConfigPrefixV1, variant.String(), dateStr)
	err = a.createAndUploadSignature(ctx, versionBytes, filePath)
	if err != nil {
		return err
	}

	return a.addVersionToList(ctx, variant, dateStr)
}

func (a Client) createSignature(content []byte, dateStr string) (res attestationconfig.AzureSEVSNPVersionSignature, err error) {
	signature, err := a.signer.Sign(content)
	if err != nil {
		return res, fmt.Errorf("sign version file: %w", err)
	}
	return attestationconfig.AzureSEVSNPVersionSignature{
		Signature: signature,
		Version:   dateStr,
	}, nil
}

// createAndUploadSignature signs the given content and uploads the signature to the given filePath with the .sig suffix.
func (a Client) createAndUploadSignature(ctx context.Context, content []byte, filePath string) error {
	signature, err := a.createSignature(content, filePath)
	if err != nil {
		return err
	}
	if err := apiclient.Update(ctx, a.s3Client, signature); err != nil {
		return fmt.Errorf("upload signature: %w", err)
	}
	return nil
}

// List returns the list of versions for the given attestation type.
func (a Client) List(ctx context.Context, attestation variant.Variant) ([]string, error) {
	if attestation.Equal(variant.AzureSEVSNP{}) {
		versions, err := apiclient.Fetch(ctx, a.s3Client, attestationconfig.AzureSEVSNPVersionList{})
		if err != nil {
			return nil, err
		}
		return versions, nil
	}
	return nil, fmt.Errorf("unsupported attestation type: %s", attestation)
}

// DeleteList empties the list of versions for the given attestation type.
func (a Client) DeleteList(ctx context.Context, attestation variant.Variant) error {
	if attestation.Equal(variant.AzureSEVSNP{}) {
		return apiclient.Update(ctx, a.s3Client, attestationconfig.AzureSEVSNPVersionList{})
	}
	return fmt.Errorf("unsupported attestation type: %s", attestation)
}

func (a Client) deleteAzureSEVSNPVersion(versions attestationconfig.AzureSEVSNPVersionList, versionStr string) (ops []crudOPNew, err error) {
	versionStr = versionStr + ".json"
	ops = append(ops, deleteCmd{
		apiObject: attestationconfig.AzureSEVSNPVersionAPI{
			Version: versionStr,
		},
	})

	ops = append(ops, deleteCmd{
		apiObject: attestationconfig.AzureSEVSNPVersionSignature{
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
	for _, op := range ops {
		if err := op.Execute(ctx, a.s3Client); err != nil {
			return fmt.Errorf("execute operation %+v: %w", op, err)
		}
	}
	return nil
}

func (a Client) addVersionToList(ctx context.Context, attestation variant.Variant, fname string) error {
	versions, err := a.List(ctx, attestation)
	if err != nil {
		return err
	}
	versions = append(versions, fname)
	versions = variant.RemoveDuplicate(versions)
	sort.Sort(sort.Reverse(sort.StringSlice(versions)))
	return apiclient.Update(ctx, a.s3Client, attestationconfig.AzureSEVSNPVersionList(versions))
}

func removeVersion(versions attestationconfig.AzureSEVSNPVersionList, versionStr string) (removedVersions attestationconfig.AzureSEVSNPVersionList, err error) {
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

type crudOPNew interface {
	Execute(ctx context.Context, c *apiclient.Client) error
}

func addVersion(versions []string, newVersion string) []string {
	versions = append(versions, newVersion)
	versions = variant.RemoveDuplicate(versions)
	sort.Sort(sort.Reverse(sort.StringSlice(versions)))
	return versions
}
