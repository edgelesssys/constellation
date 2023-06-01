/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/
package configapi

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

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/edgelesssys/constellation/v2/internal/kms/storage"
	"github.com/edgelesssys/constellation/v2/internal/sigstore"
	"github.com/edgelesssys/constellation/v2/internal/staticupload"
	"github.com/edgelesssys/constellation/v2/internal/variant"
)

// AttestationVersionRepo manages (modifies) the version information for the attestation variants.
type AttestationVersionRepo struct {
	*staticupload.Client
	cosignPwd []byte // used to decrypt the cosign private key
	privKey   []byte // used to sign
}

// NewAttestationVersionRepo returns a new AttestationVersionRepo.
func NewAttestationVersionRepo(ctx context.Context, cfg staticupload.Config, cosignPwd, privateKey []byte) (*AttestationVersionRepo, error) {
	client, err := staticupload.New(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create s3 storage: %w", err)
	}
	return &AttestationVersionRepo{client, cosignPwd, privateKey}, nil
}

// UploadAzureSEVSNP uploads the latest version numbers of the Azure SEVSNP.
func (a AttestationVersionRepo) UploadAzureSEVSNP(ctx context.Context, versions AzureSEVSNPVersion, date time.Time) error {
	versionBytes, err := json.Marshal(versions)
	if err != nil {
		return err
	}
	variant := variant.AzureSEVSNP{}
	fname := date.Format("2006-01-02-15-04") + ".json"

	filePath := fmt.Sprintf("%s/%s/%s", attestationURLPath, variant.String(), fname)
	err = put(ctx, a.Client, filePath, versionBytes)
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
func (a AttestationVersionRepo) createAndUploadSignature(ctx context.Context, content []byte, filePath string) error {
	signature, err := sigstore.SignContent(a.cosignPwd, a.privKey, content)
	if err != nil {
		return fmt.Errorf("sign version file: %w", err)
	}
	err = put(ctx, a.Client, filePath+".sig", signature)
	if err != nil {
		return fmt.Errorf("upload signature: %w", err)
	}
	return nil
}

// List returns the list of versions for the given attestation type.
func (a AttestationVersionRepo) List(ctx context.Context, attestation variant.Variant) ([]string, error) {
	key := path.Join(attestationURLPath, attestation.String(), "list")
	bt, err := get(ctx, a.Client, key)
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
func (a AttestationVersionRepo) DeleteList(ctx context.Context, attestation variant.Variant) error {
	versions := []string{}
	bt, err := json.Marshal(&versions)
	if err != nil {
		return err
	}
	return put(ctx, a.Client, path.Join(attestationURLPath, attestation.String(), "list"), bt)
}

func (a AttestationVersionRepo) addVersionToList(ctx context.Context, attestation variant.Variant, fname string) error {
	versions := []string{}
	key := path.Join(attestationURLPath, attestation.String(), "list")
	bt, err := get(ctx, a.Client, key)
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
	return put(ctx, a.Client, key, json)
}

// get is a convenience method.
func get(ctx context.Context, client *staticupload.Client, path string) ([]byte, error) {
	getObjectInput := &s3.GetObjectInput{
		Bucket: &client.BucketID,
		Key:    &path,
	}
	output, err := client.GetObject(ctx, getObjectInput)
	if err != nil {
		return nil, fmt.Errorf("getting object: %w", err)
	}
	return io.ReadAll(output.Body)
}

// put is a convenience method.
func put(ctx context.Context, client *staticupload.Client, path string, data []byte) error {
	putObjectInput := &s3.PutObjectInput{
		Bucket: &client.BucketID,
		Key:    &path,
		Body:   bytes.NewReader(data),
	}
	_, err := client.Upload(ctx, putObjectInput)
	return err
}
