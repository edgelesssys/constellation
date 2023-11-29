/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package license

import (
	"encoding/base64"
	"fmt"
)

// FromBytes reads the given license bytes and returns it as a string.
func FromBytes(license []byte) (string, error) {
	maxSize := base64.StdEncoding.DecodedLen(len(license))
	decodedLicense := make([]byte, maxSize)
	n, err := base64.StdEncoding.Decode(decodedLicense, license)
	if err != nil {
		return "", fmt.Errorf("unable to base64 decode license file: %w", err)
	}
	if n != 36 { // length of UUID
		return "", fmt.Errorf("license file corrupt: wrong length")
	}
	decodedLicense = decodedLicense[:n]
	return string(decodedLicense), nil
}
