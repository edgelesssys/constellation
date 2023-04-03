/*
Copyright (c) Edgeless Systems GmbH
SPDX-License-Identifier: AGPL-3.0-only
*/

package extract

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
)

var digestRegexp = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)

const (
	supportedSchemaVersion = 2
	supportedMediaType     = "application/vnd.oci.image.index.v1+json"
)

// Digest extracts the digest from an OCI index.
func Digest(index io.Reader) (string, error) {
	var oci ociIndex
	if err := json.NewDecoder(index).Decode(&oci); err != nil {
		return "", fmt.Errorf("decoding oci index: %w", err)
	}
	if oci.SchemaVersion != supportedSchemaVersion {
		return "", fmt.Errorf("unsupported schema version %d", oci.SchemaVersion)
	}
	if oci.MediaType != supportedMediaType {
		return "", fmt.Errorf("unsupported media type %q", oci.MediaType)
	}
	if len(oci.Manifests) != 1 {
		return "", fmt.Errorf("expected 1 manifest, got %d", len(oci.Manifests))
	}
	digest := oci.Manifests[0].Digest
	if matched := digestRegexp.MatchString(digest); !matched {
		return "", fmt.Errorf("malformed digest %q", digest)
	}

	return digest, nil
}

type ociIndex struct {
	SchemaVersion int    `json:"schemaVersion"`
	MediaType     string `json:"mediaType"`
	Manifests     []struct {
		MediaType string `json:"mediaType"`
		Digest    string `json:"digest"`
	} `json:"manifests"`
}
