/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/
package configapi

import (
	"fmt"
	"net/url"
	"path"

	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/variant"
)

// BaseURL is the base URL of the config api.
var BaseURL = constants.CDNRepositoryURL

// AttestationURLPath is the URL path to the attestation versions.
const AttestationURLPath = "constellation/v1/attestation"

// AzureSEVSNPVersionType is the type of the version to be requested.
type AzureSEVSNPVersionType (string)

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
	url, err := url.Parse(BaseURL)
	if err != nil {
		return "", fmt.Errorf("parsing CDN URL: %w", err)
	}
	url.Path = i.JSONPath()
	return url.String(), nil
}

// JSONPath returns the path to the JSON file for the request to the config api.
func (i AzureSEVSNPVersionGet) JSONPath() string {
	return path.Join(AttestationURLPath, variant.AzureSEVSNP{}.String(), i.Version)
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
type AzureSEVSNPVersionList []string

// URL returns the URL for the request to the config api.
func (i AzureSEVSNPVersionList) URL() (string, error) {
	url, err := url.Parse(BaseURL)
	if err != nil {
		return "", fmt.Errorf("parsing CDN URL: %w", err)
	}
	url.Path = i.JSONPath()
	return url.String(), nil
}

// JSONPath returns the path to the JSON file for the request to the config api.
func (i AzureSEVSNPVersionList) JSONPath() string {
	return path.Join(AttestationURLPath, variant.AzureSEVSNP{}.String(), "list")
}

// ValidateRequest validates the request.
func (i AzureSEVSNPVersionList) ValidateRequest() error {
	return nil
}

// Validate validates the request.
func (i AzureSEVSNPVersionList) Validate() error {
	return nil
}

// GetVersionByType returns the requested version of the given type.
func GetVersionByType(res AzureSEVSNPVersion, t AzureSEVSNPVersionType) uint8 {
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
		panic("unknown version type") // TODO
	}
}
