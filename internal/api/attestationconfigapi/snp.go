/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package attestationconfigapi

import (
	"encoding/json"
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
)

// AttestationURLPath is the URL path to the attestation versions.
const AttestationURLPath = "constellation/v1/attestation"

// SEVSNPVersion tracks the latest version of each component of the SEVSNP.
type SEVSNPVersion struct {
	// Bootloader is the latest version of the SEVSNP bootloader.
	Bootloader uint8 `json:"bootloader"`
	// TEE is the latest version of the SEVSNP TEE.
	TEE uint8 `json:"tee"`
	// SNP is the latest version of the SEVSNP SNP.
	SNP uint8 `json:"snp"`
	// Microcode is the latest version of the SEVSNP microcode.
	Microcode uint8 `json:"microcode"`
}

// SEVSNPVersionAPI is the request to get the version information of the specific version in the config api.
// Because variant is not part of the marshalled JSON, fetcher and client methods need to fill the variant property.
// Once we switch to v2 of the API we should embed the variant in the object.
// That would remove the possibility of some fetcher/client code forgetting to set the variant.
type SEVSNPVersionAPI struct {
	Version string          `json:"-"`
	Variant variant.Variant `json:"-"`
	SEVSNPVersion
}

// JSONPath returns the path to the JSON file for the request to the config api.
func (i SEVSNPVersionAPI) JSONPath() string {
	return path.Join(AttestationURLPath, i.Variant.String(), i.Version)
}

// ValidateRequest validates the request.
func (i SEVSNPVersionAPI) ValidateRequest() error {
	if !strings.HasSuffix(i.Version, ".json") {
		return fmt.Errorf("version has no .json suffix")
	}
	return nil
}

// Validate is a No-Op at the moment.
func (i SEVSNPVersionAPI) Validate() error {
	return nil
}

// SEVSNPVersionList is the request to list all versions in the config api.
// Because variant is not part of the marshalled JSON, fetcher and client methods need to fill the variant property.
// Once we switch to v2 of the API we could embed the variant in the object and remove some code from fetcher & client.
// That would remove the possibility of some fetcher/client code forgetting to set the variant.
type SEVSNPVersionList struct {
	Variant variant.Variant
	List    []string
}

// MarshalJSON marshals the i's list property to JSON.
func (i SEVSNPVersionList) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.List)
}

// UnmarshalJSON unmarshals a list of strings into i's list property.
func (i *SEVSNPVersionList) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &i.List)
}

// JSONPath returns the path to the JSON file for the request to the config api.
func (i SEVSNPVersionList) JSONPath() string {
	return path.Join(AttestationURLPath, i.Variant.String(), "list")
}

// ValidateRequest is a NoOp as there is no input.
func (i SEVSNPVersionList) ValidateRequest() error {
	return nil
}

// SortReverse sorts the list of versions in reverse order.
func (i *SEVSNPVersionList) SortReverse() {
	sort.Sort(sort.Reverse(sort.StringSlice(i.List)))
}

// AddVersion adds new to i's list and sorts the element in descending order.
func (i *SEVSNPVersionList) AddVersion(new string) {
	i.List = append(i.List, new)
	i.List = variant.RemoveDuplicate(i.List)

	i.SortReverse()
}

// Validate validates the response.
func (i SEVSNPVersionList) Validate() error {
	if len(i.List) < 1 {
		return fmt.Errorf("no versions found in /list")
	}
	return nil
}
