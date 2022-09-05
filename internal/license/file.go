/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package license

import (
	"encoding/base64"
	"fmt"

	"github.com/edgelesssys/constellation/internal/file"
)

func FromFile(fileHandler file.Handler, path string) (string, error) {
	readBytes, err := fileHandler.Read(path)
	if err != nil {
		return "", fmt.Errorf("unable to read from '%s': %w", path, err)
	}

	maxSize := base64.StdEncoding.DecodedLen(len(readBytes))
	decodedLicense := make([]byte, maxSize)
	n, err := base64.StdEncoding.Decode(decodedLicense, readBytes)
	if err != nil {
		return "", fmt.Errorf("unable to base64 decode license file: %w", err)
	}
	if n != 36 { // length of UUID
		return "", fmt.Errorf("license file corrupt: wrong length")
	}
	decodedLicense = decodedLicense[:n]

	return string(decodedLicense), nil
}
