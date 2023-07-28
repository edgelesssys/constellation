/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package attestationconfigapi

import (
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/sev-snp-measure-go/ovmf"
)

// FirmwareMetadata contains information to precalculate the launchmeasurement of one firmware version.
type FirmwareMetadata struct {
	Data        ovmf.MetadataWrapper `json:"metadata"`
	FirstSeenOn time.Time            `json:"firstSeenOn"`
}

// AWSFirmwareMetadata tracks the latest version of each component of the Azure SEVSNP.
type AWSFirmwareMetadata struct {
	Version  string             `json:"-"`
	Metadata []FirmwareMetadata `json:"metadata"`
}

func NewAWSFirmwareMetadata() AWSFirmwareMetadata {
	return AWSFirmwareMetadata{
		Version: "1.0.json",
	}
}

// URL returns the URL for the request to the config api.
func (i AWSFirmwareMetadata) URL() (string, error) {
	return getURL(i)
}

// JSONPath returns the path to the JSON file for the request to the config api.
func (i AWSFirmwareMetadata) JSONPath() string {
	return path.Join(attestationURLPath, variant.AWSSEVSNP{}.String(), i.Version+".json")
}

// ValidateRequest validates the request.
func (i AWSFirmwareMetadata) ValidateRequest() error {
	if !strings.HasSuffix(i.Version, ".json") {
		return fmt.Errorf("version has no .json suffix")
	}
	return nil
}

// Validate is a No-Op at the moment.
func (i AWSFirmwareMetadata) Validate() error {
	return nil
}

// AWSFirmwareMetadataList is the request to list all versions in the config api.
type AWSFirmwareMetadataList []string

// URL returns the URL for the request to the config api.
func (i AWSFirmwareMetadataList) URL() (string, error) {
	return getURL(i)
}

// JSONPath returns the path to the JSON file for the request to the config api.
func (i AWSFirmwareMetadataList) JSONPath() string {
	return path.Join(attestationURLPath, variant.AzureSEVSNP{}.String(), "list")
}

// ValidateRequest is a NoOp as there is no input.
func (i AWSFirmwareMetadataList) ValidateRequest() error {
	return nil
}

// Validate validates the response.
func (i AWSFirmwareMetadataList) Validate() error {
	if len(i) < 1 {
		return fmt.Errorf("no versions found in /list")
	}
	return nil
}
