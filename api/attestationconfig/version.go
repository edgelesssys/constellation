/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package attestationconfig

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

// SEVSNPVersion tracks the latest version of each component for SEV-SNP.
type SEVSNPVersion struct {
	// Bootloader is the latest version of the SEV-SNP bootloader.
	Bootloader uint8 `json:"bootloader"`
	// TEE is the latest version of the SEV-SNP TEE.
	TEE uint8 `json:"tee"`
	// SNP is the latest version of the SEV-SNP SNP.
	SNP uint8 `json:"snp"`
	// Microcode is the latest version of the SEV-SNP microcode.
	Microcode uint8 `json:"microcode"`
}

// TDXVersion tracks the latest version of each component for TDX.
type TDXVersion struct {
	// QESVN is the latest QE security version number.
	QESVN uint16 `json:"qeSVN"`
	// PCESVN is the latest PCE security version number.
	PCESVN uint16 `json:"pceSVN"`
	// TEETCBSVN are the latest component-wise security version numbers for the TEE.
	TEETCBSVN [16]byte `json:"teeTCBSVN"`
	// QEVendorID is the latest QE vendor ID.
	QEVendorID [16]byte `json:"qeVendorID"`
	// XFAM is the latest XFAM field.
	XFAM [8]byte `json:"xfam"`
}

// Entry is the request to get the version information of the specific version in the config api.
//
// TODO: Because variant is not part of the marshalled JSON, fetcher and client methods need to fill the variant property.
// In API v2 we should embed the variant in the object and remove some code from fetcher & client.
// That would remove the possibility of some fetcher/client code forgetting to set the variant.
type Entry struct {
	Version string          `json:"-"`
	Variant variant.Variant `json:"-"`
	SEVSNPVersion
	TDXVersion
}

// JSONPath returns the path to the JSON file for the request to the config api.
func (i Entry) JSONPath() string {
	return path.Join(AttestationURLPath, i.Variant.String(), i.Version)
}

// ValidateRequest validates the request.
func (i Entry) ValidateRequest() error {
	if !strings.HasSuffix(i.Version, ".json") {
		return fmt.Errorf("version has no .json suffix")
	}
	return nil
}

// Validate is a No-Op at the moment.
func (i Entry) Validate() error {
	return nil
}

// List is the request to retrieve of all versions in the API for one attestation variant.
//
// TODO: Because variant is not part of the marshalled JSON, fetcher and client methods need to fill the variant property.
// In API v2 we should embed the variant in the object and remove some code from fetcher & client.
// That would remove the possibility of some fetcher/client code forgetting to set the variant.
type List struct {
	Variant variant.Variant
	List    []string
}

// MarshalJSON marshals the i's list property to JSON.
func (i List) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.List)
}

// UnmarshalJSON unmarshals a list of strings into i's list property.
func (i *List) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &i.List)
}

// JSONPath returns the path to the JSON file for the request to the config api.
func (i List) JSONPath() string {
	return path.Join(AttestationURLPath, i.Variant.String(), "list")
}

// ValidateRequest is a NoOp as there is no input.
func (i List) ValidateRequest() error {
	return nil
}

// SortReverse sorts the list of versions in reverse order.
func (i *List) SortReverse() {
	sort.Sort(sort.Reverse(sort.StringSlice(i.List)))
}

// AddVersion adds new to i's list and sorts the element in descending order.
func (i *List) AddVersion(new string) {
	i.List = append(i.List, new)
	i.List = variant.RemoveDuplicate(i.List)

	i.SortReverse()
}

// Validate validates the response.
func (i List) Validate() error {
	if len(i.List) < 1 {
		return fmt.Errorf("no versions found in /list")
	}
	return nil
}
