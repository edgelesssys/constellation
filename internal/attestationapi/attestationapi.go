package attestationapi

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
	"github.com/edgelesssys/constellation/v2/internal/variant"
	"github.com/edgelesssys/constellation/v2/internal/versionsapi"
	"github.com/edgelesssys/constellation/v2/internal/versionsapi/fetcher"
)

const (
	Bootloader Type = "bootloader" // Bootloader is the version of the Azure SEVSNP bootloader.
	TEE        Type = "tee"        // TEE is the version of the Azure SEVSNP TEE.
	SNP        Type = "snp"        // SNP is the version of the Azure SEVSNP SNP.
	Microcode  Type = "microcode"  // Microcode is the version of the Azure SEVSNP microcode.
)

// AttestationPath is the path to the attestation versions.
const AttestationPath = "constellation/v1/attestation"

// AzureSEVSNP is the latest version of each component of the Azure SEVSNP.
// used for testing only
var AzureSEVSNP = versionsapi.AzureSEVSNPVersion{
	Bootloader: 2,
	TEE:        0,
	SNP:        6,
	Microcode:  93,
}

// Type is the type of the version to be requested.
type Type (string)

// AttestationVersionRepo manages (modifies) the version information for the attestation variants.
type AttestationVersionRepo struct {
	*awss3.Storage
}

// NewAttestationVersionRepo returns a new AttestationVersionRepo.
func NewAttestationVersionRepo(ctx context.Context, cfg uri.AWSS3Config) (*AttestationVersionRepo, error) {
	s3, err := awss3.New(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create s3 storage: %w", err)
	}
	return &AttestationVersionRepo{s3}, nil
}

// UploadAzureSEVSNP uploads the latest version numbers of the Azure SEVSNP.
func (a AttestationVersionRepo) UploadAzureSEVSNP(ctx context.Context, versions versionsapi.AzureSEVSNPVersion, date time.Time) error {
	bt, err := json.Marshal(versions)
	if err != nil {
		return err
	}
	variant := variant.AzureSEVSNP{}
	fname := date.Format("2006-01-02-15-04") + ".json"

	err = a.Put(ctx, fmt.Sprintf("%s/%s/%s", AttestationPath, variant.String(), fname), bt)
	if err != nil {
		return err
	}
	return a.addVersionToList(ctx, variant, fname)
}

func (a AttestationVersionRepo) addVersionToList(ctx context.Context, attestation variant.Variant, fname string) error {
	versions := []string{}
	key := path.Join(AttestationPath, attestation.String(), "list")
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

// List returns the list of versions for the given attestation type.
func (a AttestationVersionRepo) List(ctx context.Context, attestation variant.Variant) ([]string, error) {
	key := path.Join(AttestationPath, attestation.String(), "list")
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
	return a.Put(ctx, path.Join(AttestationPath, attestation.String(), "list"), bt)
}

func GetVersionByType(res versionsapi.AzureSEVSNPVersion, t Type) uint8 {
	switch t {
	case Bootloader:
		return res.Bootloader
	case TEE:
		return res.TEE
	case SNP:
		return res.SNP
	case Microcode:
		return res.Microcode
	default:
		return 1
	}
}

// GetAzureSEVSNPVersion returns the requested version of the given type.
func GetAzureSEVSNPVersion(ctx context.Context) (res versionsapi.AzureSEVSNPVersion, err error) {
	var versions versionsapi.AzureSEVSNPVersionList
	fetcher := fetcher.NewFetcher()
	versions, err = fetcher.FetchAttestationList(ctx, versions)
	if err != nil {
		return res, fmt.Errorf("failed fetching versions list: %w", err)
	}
	if len(versions) < 1 {
		return res, errors.New("no versions found in /list")
	}
	get := versionsapi.AzureSEVSNPVersionGet{Version: versions[0]} // get latest version (as sorted reversely alphanumerically)
	get, err = fetcher.FetchAttestationVersion(ctx, get)
	if err != nil {
		return res, fmt.Errorf("failed fetching version: %w", err)
	}
	return get.AzureSEVSNPVersion, nil
}
