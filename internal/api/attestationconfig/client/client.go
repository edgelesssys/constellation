/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path"
	"sort"
	"time"

	s3manager "github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/edgelesssys/constellation/v2/internal/api/attestationconfig"
	"github.com/edgelesssys/constellation/v2/internal/api/attestationconfig/fetcher"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/kms/storage"
	"github.com/edgelesssys/constellation/v2/internal/sigstore"
	"github.com/edgelesssys/constellation/v2/internal/staticupload"
	"github.com/edgelesssys/constellation/v2/internal/variant"
)

// Client is the client for the attestationconfig API.
type Client interface {
	UploadAzureSEVSNP(ctx context.Context, versions attestationconfig.AzureSEVSNPVersion, date time.Time) error
	List(ctx context.Context, attestation variant.Variant) ([]string, error)
	DeleteList(ctx context.Context, attestation variant.Variant) error
	DeleteAzureSEVSNVersion(ctx context.Context, versionStr string) error
}

// client manages (modifies) the version information for the attestation variants.
type client struct {
	s3Client
	s3ClientClose func(ctx context.Context) error
	bucketID      string
	cosignPwd     []byte // used to decrypt the cosign private key
	privKey       []byte // used to sign
	fetcher       fetcher.AttestationConfigAPIFetcher
}

// New returns a new Client.
func New(ctx context.Context, cfg staticupload.Config, cosignPwd, privateKey []byte) (Client, CloseFunc, error) {
	staticclient, clientClose, err := staticupload.New(ctx, cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create s3 storage: %w", err)
	}

	repo := &client{
		s3Client:      staticclient,
		s3ClientClose: clientClose,
		bucketID:      cfg.Bucket,
		cosignPwd:     cosignPwd,
		privKey:       privateKey,
		fetcher:       fetcher.New(),
	}
	return repo, repo.Close, nil
}

// Close closes the Client.
func (a client) Close(ctx context.Context) error {
	if a.s3ClientClose == nil {
		return nil
	}
	return a.s3ClientClose(ctx)
}

// UploadAzureSEVSNP uploads the latest version numbers of the Azure SEVSNP.
func (a client) UploadAzureSEVSNP(ctx context.Context, versions attestationconfig.AzureSEVSNPVersion, date time.Time) error {
	versionBytes, err := json.Marshal(versions)
	if err != nil {
		return err
	}
	variant := variant.AzureSEVSNP{}
	fname := date.Format("2006-01-02-15-04") + ".json"

	filePath := fmt.Sprintf("%s/%s/%s", constants.CDNAttestationConfigPrefixV1, variant.String(), fname)
	err = put(ctx, a.s3Client, a.bucketID, filePath, versionBytes)
	if err != nil {
		return err
	}

	err = a.createAndUploadSignature(ctx, versionBytes, filePath)
	if err != nil {
		return err
	}
	return a.addVersionToList(ctx, variant, fname)
}

// createAndUploadSignature signs the given content and uploads the signature to the given filePath with the .sig suffix.
func (a client) createAndUploadSignature(ctx context.Context, content []byte, filePath string) error {
	signature, err := sigstore.SignContent(a.cosignPwd, a.privKey, content)
	if err != nil {
		return fmt.Errorf("sign version file: %w", err)
	}
	err = put(ctx, a.s3Client, a.bucketID, filePath+".sig", signature)
	if err != nil {
		return fmt.Errorf("upload signature: %w", err)
	}
	return nil
}

// List returns the list of versions for the given attestation type.
func (a client) List(ctx context.Context, attestation variant.Variant) ([]string, error) {
	key := path.Join(constants.CDNAttestationConfigPrefixV1, attestation.String(), "list")
	bt, err := get(ctx, a.s3Client, a.bucketID, key)
	if err != nil {
		return nil, err
	}
	var versions []string
	if err := json.Unmarshal(bt, &versions); err != nil {
		return nil, err
	}
	return versions, nil
}

// DeleteList empties the list of versions for the given attestation type.
func (a client) DeleteList(ctx context.Context, attestation variant.Variant) error {
	versions := []string{}
	bt, err := json.Marshal(&versions)
	if err != nil {
		return err
	}
	return put(ctx, a.s3Client, a.bucketID, path.Join(constants.CDNAttestationConfigPrefixV1, attestation.String(), "list"), bt)
}

func (a client) deleteAzureSEVSNVersion(versions attestationconfig.AzureSEVSNPVersionList, versionStr string) (ops []crudOP, err error) {
	versionStr = versionStr + ".json"
	variant := variant.AzureSEVSNP{}
	// delete version file
	filePath := fmt.Sprintf("%s/%s/%s", constants.CDNAttestationConfigPrefixV1, variant.String(), versionStr)
	ops = append(ops, deleteCmd{
		path: filePath,
	})

	// delete signature file
	sigPath := fmt.Sprintf("%s/%s/%s", constants.CDNAttestationConfigPrefixV1, variant.String(), versionStr+".sig")
	ops = append(ops, deleteCmd{
		path: sigPath,
	})

	// delete version from /list
	removedVersions, err := removeVersion(versions, versionStr)
	if err != nil {
		return nil, err
	}
	versionBt, err := json.Marshal(removedVersions)
	if err != nil {
		return nil, fmt.Errorf("marshal updated version list: %w", err)
	}
	ops = append(ops, putCmd{
		data: versionBt,
		path: path.Join(constants.CDNAttestationConfigPrefixV1, variant.String(), "list"),
	})
	return ops, nil
}

// DeleteAzureSEVSNPVersion deletes the given version (without .json suffix) from the API.
func (a client) DeleteAzureSEVSNVersion(ctx context.Context, versionStr string) error {
	versions, err := a.fetcher.FetchAzureSEVSNPVersionList(ctx, attestationconfig.AzureSEVSNPVersionList{})
	if err != nil {
		return fmt.Errorf("fetch version list: %w", err)
	}
	ops, err := a.deleteAzureSEVSNVersion(versions, versionStr)
	if err != nil {
		return err
	}
	for _, op := range ops {
		if err := op.Execute(ctx, a.s3Client, a.bucketID); err != nil {
			return fmt.Errorf("execute operation %+v: %w", op, err)
		}
	}
	return nil
}

func (a client) addVersionToList(ctx context.Context, attestation variant.Variant, fname string) error {
	versions := []string{}
	key := path.Join(constants.CDNAttestationConfigPrefixV1, attestation.String(), "list")
	bt, err := get(ctx, a.s3Client, a.bucketID, key)
	if err == nil {
		if err := json.Unmarshal(bt, &versions); err != nil {
			return err
		}
	} else if !errors.Is(err, storage.ErrDEKUnset) {
		return err
	}
	versions = append(versions, fname)
	versions = variant.RemoveDuplicate(versions)
	sort.Sort(sort.Reverse(sort.StringSlice(versions)))
	json, err := json.Marshal(versions)
	if err != nil {
		return err
	}
	return put(ctx, a.s3Client, a.bucketID, key, json)
}

// get is a convenience method.
func get(ctx context.Context, client s3Client, bucket, path string) ([]byte, error) {
	getObjectInput := &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &path,
	}
	output, err := client.GetObject(ctx, getObjectInput)
	if err != nil {
		return nil, fmt.Errorf("getting object: %w", err)
	}
	return io.ReadAll(output.Body)
}

// put is a convenience method.
func put(ctx context.Context, client s3Client, bucket, path string, data []byte) error {
	putObjectInput := &s3.PutObjectInput{
		Bucket: &bucket,
		Key:    &path,
		Body:   bytes.NewReader(data),
	}
	_, err := client.Upload(ctx, putObjectInput)
	return err
}

// delete is a convenience method.
func delete(ctx context.Context, client s3Client, bucket, path string) error {
	deleteObjectInput := &s3.DeleteObjectInput{
		Bucket: &bucket,
		Key:    &path,
	}
	_, err := client.DeleteObject(ctx, deleteObjectInput)
	return err
}

type s3Client interface {
	GetObject(
		ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options),
	) (*s3.GetObjectOutput, error)
	Upload(
		ctx context.Context, input *s3.PutObjectInput, opts ...func(*s3manager.Uploader),
	) (*s3manager.UploadOutput, error)
	DeleteObject(ctx context.Context, params *s3.DeleteObjectInput,
		optFns ...func(*s3.Options),
	) (*s3.DeleteObjectOutput, error)
}

// CloseFunc is a function that closes the client.
type CloseFunc func(ctx context.Context) error

func removeVersion(versions attestationconfig.AzureSEVSNPVersionList, versionStr string) (removedVersions attestationconfig.AzureSEVSNPVersionList, err error) {
	for i, v := range versions {
		if v == versionStr {
			removedVersions = append(versions[:i], versions[i+1:]...)
			return removedVersions, nil
		}
	}
	return nil, fmt.Errorf("version %s not found in list %v", versionStr, versions)
}

type deleteCmd struct {
	path string
}

func (d deleteCmd) Execute(ctx context.Context, c s3Client, bucketID string) error {
	return delete(ctx, c, bucketID, d.path)
}

type putCmd struct {
	data []byte
	path string
}

func (p putCmd) Execute(ctx context.Context, c s3Client, bucketID string) error {
	return put(ctx, c, bucketID, p.path, p.data)
}

type crudOP interface {
	Execute(ctx context.Context, c s3Client, bucketID string) error
}
