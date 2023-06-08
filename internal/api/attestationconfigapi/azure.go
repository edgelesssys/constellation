/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package attestationconfigapi

import (
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/variant"
)

// attestationURLPath is the URL path to the attestation versions.
const attestationURLPath = "constellation/v1/attestation"

// AzureSEVSNPVersionType is the type of the version to be requested.
type AzureSEVSNPVersionType string

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

// AzureSEVSNPVersionSignature is the object to perform CRUD operations on the config api.
type AzureSEVSNPVersionSignature struct {
	Version   string `json:"-"`
	Signature []byte `json:"signature"`
}

// JSONPath returns the path to the JSON file for the request to the config api.
func (s AzureSEVSNPVersionSignature) JSONPath() string {
	return path.Join(attestationURLPath, variant.AzureSEVSNP{}.String(), s.Version+".sig")
}

// URL returns the URL for the request to the config api.
func (s AzureSEVSNPVersionSignature) URL() (string, error) {
	return getURL(s)
}

// ValidateRequest validates the request.
func (s AzureSEVSNPVersionSignature) ValidateRequest() error {
	if !strings.HasSuffix(s.Version, ".json") {
		return fmt.Errorf("%s version has no .json suffix", s.Version)
	}
	return nil
}

// Validate is a No-Op at the moment.
func (s AzureSEVSNPVersionSignature) Validate() error {
	return nil
}

// AzureSEVSNPVersionAPI is the request to get the version information of the specific version in the config api.
type AzureSEVSNPVersionAPI struct {
	Version string `json:"-"`
	AzureSEVSNPVersion
}

// URL returns the URL for the request to the config api.
func (i AzureSEVSNPVersionAPI) URL() (string, error) {
	return getURL(i)
}

// JSONPath returns the path to the JSON file for the request to the config api.
func (i AzureSEVSNPVersionAPI) JSONPath() string {
	return path.Join(attestationURLPath, variant.AzureSEVSNP{}.String(), i.Version)
}

// ValidateRequest validates the request.
func (i AzureSEVSNPVersionAPI) ValidateRequest() error {
	if !strings.HasSuffix(i.Version, ".json") {
		return fmt.Errorf("version has no .json suffix")
	}
	return nil
}

// Validate is a No-Op at the moment.
func (i AzureSEVSNPVersionAPI) Validate() error {
	return nil
}

// AzureSEVSNPVersionList is the request to list all versions in the config api.
type AzureSEVSNPVersionList []string

// URL returns the URL for the request to the config api.
func (i AzureSEVSNPVersionList) URL() (string, error) {
	return getURL(i)
}

// JSONPath returns the path to the JSON file for the request to the config api.
func (i AzureSEVSNPVersionList) JSONPath() string {
	return path.Join(attestationURLPath, variant.AzureSEVSNP{}.String(), "list")
}

// ValidateRequest is a NoOp as there is no input.
func (i AzureSEVSNPVersionList) ValidateRequest() error {
	return nil
}

// Validate validates the response.
func (i AzureSEVSNPVersionList) Validate() error {
	if len(i) < 1 {
		return fmt.Errorf("no versions found in /list")
	}
	return nil
}

func getURL(obj jsoPather) (string, error) {
	url, err := url.Parse(constants.CDNRepositoryURL)
	if err != nil {
		return "", fmt.Errorf("parsing CDN URL: %w", err)
	}
	url.Path = obj.JSONPath()
	return url.String(), nil
}

type jsoPather interface {
	JSONPath() string
}
