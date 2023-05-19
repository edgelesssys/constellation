package versionsapi

import (
	"fmt"
	"net/url"
	"path"

	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/variant"
)

// AttestationPath is the path to the attestation versions.
const AttestationPath = "constellation/v1/attestation" // TODO already in attestationonapi but import cycle otherwise

// AzureSEVSNPVersion tracks the latest version of each component of the Azure SEVSNP.
type AzureSEVSNPVersion struct {
	// Bootloader is the latest version of the Azure SEVSNP bootloader.
	Bootloader uint8 `json:"bootloader"`
	// TEE is the latest version of the Azure SEVSNP TEE.
	TEE uint8 `json:"tee"`
	// SNP is the latest version of the Azure SEVSNP SNP.
	SNP uint8 `json:"snp"`
	// Microcode is the latest version of the Azure SEVSNP microcode.
	Microcode uint8 `json:"microcode"`
}

type AzureSEVSNPVersionGet struct {
	Version string `json:"-"`
	AzureSEVSNPVersion
}

func (i AzureSEVSNPVersionGet) URL() (string, error) {
	url, err := url.Parse(constants.CDNRepositoryURL)
	if err != nil {
		return "", fmt.Errorf("parsing CDN URL: %w", err)
	}
	url.Path = i.JSONPath()
	return url.String(), nil
}

func (i AzureSEVSNPVersionGet) JSONPath() string {
	return path.Join(AttestationPath, variant.AzureSEVSNP{}.String(), i.Version)
}

func (i AzureSEVSNPVersionGet) ValidateRequest() error {
	return nil
}

func (i AzureSEVSNPVersionGet) Validate() error {
	return nil
}

type AzureSEVSNPVersionList ([]string)

func (i AzureSEVSNPVersionList) URL() (string, error) {
	url, err := url.Parse(constants.CDNRepositoryURL)
	if err != nil {
		return "", fmt.Errorf("parsing CDN URL: %w", err)
	}
	url.Path = i.JSONPath()
	return url.String(), nil
}

func (i AzureSEVSNPVersionList) JSONPath() string {
	return path.Join(AttestationPath, variant.AzureSEVSNP{}.String(), "list")
}

func (i AzureSEVSNPVersionList) ValidateRequest() error {
	return nil
}

func (i AzureSEVSNPVersionList) Validate() error {
	return nil
}
