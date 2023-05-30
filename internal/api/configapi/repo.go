/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/
package configapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"sort"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/kms/storage"
	"github.com/edgelesssys/constellation/v2/internal/kms/storage/awss3"
	"github.com/edgelesssys/constellation/v2/internal/kms/uri"
	"github.com/edgelesssys/constellation/v2/internal/sigstore"
	"github.com/edgelesssys/constellation/v2/internal/variant"
)

// AttestationVersionRepo manages (modifies) the version information for the attestation variants.
type AttestationVersionRepo struct {
	*awss3.Storage
	cosignPwd []byte // used to decrypt the cosign private key
	privKey   []byte // used to sign
}

// NewAttestationVersionRepo returns a new AttestationVersionRepo.
func NewAttestationVersionRepo(ctx context.Context, cfg uri.AWSS3Config, cosignPwd, privateKey []byte) (*AttestationVersionRepo, error) {
	s3, err := awss3.New(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create s3 storage: %w", err)
	}
	return &AttestationVersionRepo{s3, cosignPwd, privateKey}, nil
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
	err = a.Put(ctx, filePath, versionBytes)
	if err != nil {
		return err
	}

	err = a.signAndUpload(ctx, versionBytes, filePath)
	if err != nil {
		return err
	}
	return a.addVersionToList(ctx, variant, fname)
}

// signAndUpload signs the given content and uploads it to the given filePath with the .sig suffix.
func (a AttestationVersionRepo) signAndUpload(ctx context.Context, content []byte, filePath string) error {
	signature, err := sigstore.SignContent(a.cosignPwd, a.privKey, content)
	if err != nil {
		return fmt.Errorf("sign version file: %w", err)
	}
	err = a.Put(ctx, filePath+".sig", signature)
	if err != nil {
		return fmt.Errorf("upload signature: %w", err)
	}
	return nil
}

// List returns the list of versions for the given attestation type.
func (a AttestationVersionRepo) List(ctx context.Context, attestation variant.Variant) ([]string, error) {
	key := path.Join(attestationURLPath, attestation.String(), "list")
	bt, err := a.Get(ctx, key)
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
	return a.Put(ctx, path.Join(attestationURLPath, attestation.String(), "list"), bt)
}

func (a AttestationVersionRepo) addVersionToList(ctx context.Context, attestation variant.Variant, fname string) error {
	versions := []string{}
	key := path.Join(attestationURLPath, attestation.String(), "list")
	bt, err := a.Get(ctx, key)
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
	return a.Put(ctx, key, json)
}
