/*
Copyright (c) Edgeless Systems GmbH
SPDX-License-Identifier: AGPL-3.0-only
*/

package extract

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDigest(t *testing.T) {
	testCases := map[string]struct {
		ociIndex   string
		wantDigest string
		wantErr    bool
	}{
		"valid OCI index": {
			ociIndex:   validOCIIndex,
			wantDigest: "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			wantErr:    false,
		},
		"wrong version": {
			ociIndex: `{
				"schemaVersion": 1,
				"mediaType": "application/vnd.oci.image.index.v1+json",
				"manifests": [
					{
						"mediaType": "application/vnd.oci.image.manifest.v1+json",
						"digest": "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
						"size": 0
					}
				]
			}`,
			wantErr: true,
		},
		"wrong media type": {
			ociIndex: `{
				"schemaVersion": 2,
				"mediaType": "application/something-else",
				"manifests": [
					{
						"mediaType": "application/vnd.oci.image.manifest.v1+json",
						"digest": "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
						"size": 0
					}
				]
			}`,
			wantErr: true,
		},
		"incorrect manifest length": {
			ociIndex: `{
				"schemaVersion": 2,
				"mediaType": "application/vnd.oci.image.index.v1+json",
				"manifests": []
			}`,
			wantErr: true,
		},
		"incorrect manifest digest format": {
			ociIndex: `{
				"schemaVersion": 2,
				"mediaType": "application/vnd.oci.image.index.v1+json",
				"manifests": [
					{
						"mediaType": "application/vnd.oci.image.manifest.v1+json",
						"digest": "foo:bar",
						"size": 0
					}
				]
			}`,
			wantErr: true,
		},
		"malformed json": {
			ociIndex: `}`,
			wantErr:  true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)
			digest, err := Digest(bytes.NewBufferString(tc.ociIndex))
			if tc.wantErr {
				require.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantDigest, digest)
		})
	}
}

const (
	// This is a valid OCI index.
	// It has a schema version of 2, a media type of
	// "application/vnd.oci.image.index.v1+json", and a single manifest.
	// The manifest has a media type of
	// "application/vnd.oci.image.manifest.v1+json", and a digest.
	// The digest is a valid SHA256 hash.
	validOCIIndex = `{
		"schemaVersion": 2,
		"mediaType": "application/vnd.oci.image.index.v1+json",
		"manifests": [
			{
				"mediaType": "application/vnd.oci.image.manifest.v1+json",
				"digest": "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
				"size": 0
			}
		]
	}`
)
