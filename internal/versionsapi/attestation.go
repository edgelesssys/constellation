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

// AzureSEVSNPVersionGet is the request to get the version information of the specific version in the config api.
type AzureSEVSNPVersionGet struct {
	Version string `json:"-"`
	AzureSEVSNPVersion
}

// URL returns the URL for the request to the config api.
func (i AzureSEVSNPVersionGet) URL() (string, error) {
	url, err := url.Parse(constants.CDNRepositoryURL)
	if err != nil {
		return "", fmt.Errorf("parsing CDN URL: %w", err)
	}
	url.Path = i.JSONPath()
	return url.String(), nil
}

// JSONPath returns the path to the JSON file for the request to the config api.
func (i AzureSEVSNPVersionGet) JSONPath() string {
	return path.Join(AttestationPath, variant.AzureSEVSNP{}.String(), i.Version)
}

// ValidateRequest validates the request.
func (i AzureSEVSNPVersionGet) ValidateRequest() error {
	return nil
}

// Validate validates the request.
func (i AzureSEVSNPVersionGet) Validate() error {
	return nil
}

// AzureSEVSNPVersionList is the request to list all versions in the config api.
type AzureSEVSNPVersionList ([]string)

// URL returns the URL for the request to the config api.
func (i AzureSEVSNPVersionList) URL() (string, error) {
	url, err := url.Parse(constants.CDNRepositoryURL)
	if err != nil {
		return "", fmt.Errorf("parsing CDN URL: %w", err)
	}
	url.Path = i.JSONPath()
	return url.String(), nil
}

// JSONPath returns the path to the JSON file for the request to the config api.
func (i AzureSEVSNPVersionList) JSONPath() string {
	return path.Join(AttestationPath, variant.AzureSEVSNP{}.String(), "list")
}

// ValidateRequest validates the request.
func (i AzureSEVSNPVersionList) ValidateRequest() error {
	return nil
}

// Validate validates the request.
func (i AzureSEVSNPVersionList) Validate() error {
	return nil
}
