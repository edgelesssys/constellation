/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package sigstore

import (
	"testing"

	"github.com/sigstore/rekor/pkg/generated/models"
	hashedrekord "github.com/sigstore/rekor/pkg/types/hashedrekord/v0.0.1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsEntrySignedBy(t *testing.T) {
	testCases := map[string]struct {
		entry       *hashedrekord.V001Entry
		key         string
		wantSuccess bool
	}{
		"valid key": {
			entry: &hashedrekord.V001Entry{
				HashedRekordObj: models.HashedrekordV001Schema{
					Signature: &models.HashedrekordV001SchemaSignature{
						PublicKey: &models.HashedrekordV001SchemaSignaturePublicKey{
							Content: []byte("my key"),
						},
					},
				},
			},
			key:         "bXkga2V5", // "my key" in base64
			wantSuccess: true,
		},
		"nil rekord": {
			entry:       nil,
			wantSuccess: false,
		},
		"nil signature": {
			entry: &hashedrekord.V001Entry{
				HashedRekordObj: models.HashedrekordV001Schema{
					Signature: nil,
				},
			},
			wantSuccess: false,
		},
		"nil pub key": {
			entry: &hashedrekord.V001Entry{
				HashedRekordObj: models.HashedrekordV001Schema{
					Signature: &models.HashedrekordV001SchemaSignature{
						PublicKey: nil,
					},
				},
			},
			wantSuccess: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			assert.Equal(tc.wantSuccess, isEntrySignedBy(tc.entry, tc.key))
		})
	}
}

func TestNewRekor(t *testing.T) {
	assert := assert.New(t)
	rekor, err := NewRekor()
	assert.NoError(err)
	assert.NotNil(rekor)
}

func TestHashedRekordFromEntry(t *testing.T) {
	testCases := map[string]struct {
		jsonEntry string
		wantError bool
	}{
		"invalid base64": {
			jsonEntry: "{\"body\":\"abc!\"}",
			wantError: true,
		},
		"valid base64, but invalid json": {
			jsonEntry: "{\"body\":\"aGVsbG8K\"}", // base64(hello)
			wantError: true,
		},
		"valid v001Entry": {
			jsonEntry: "{\"body\":\"eyJhcGlWZXJzaW9uIjoiMC4wLjEiLCJraW5kIjoiaGFzaGVkcmVrb3JkIiwic3BlYyI6eyJkYXRhIjp7Imhhc2giOnsiYWxnb3JpdGhtIjoic2hhMjU2IiwidmFsdWUiOiI0MGUxMzdiOWI5YjgyMDRkNjcyNjQyZmQxZTE4MWM2ZDVjY2I1MGNmYzVjYzdmY2JiMDZhOGMyYzc4ZjQ0YWZmIn19LCJzaWduYXR1cmUiOnsiY29udGVudCI6Ik1FVUNJUUNTRVIzbUdqK2o1UHIya09YVGxDSUhRQzNnVDMwSTdxa0xyOUF3dDZlVVVRSWdjTFVLUklsWTUwVU44Skd3VmVOZ2tCWnlZRDhITXh3Qy9MRlJXb01uMTgwPSIsInB1YmxpY0tleSI6eyJjb250ZW50IjoiTFMwdExTMUNSVWRKVGlCUVZVSk1TVU1nUzBWWkxTMHRMUzBLVFVacmQwVjNXVWhMYjFwSmVtb3dRMEZSV1VsTGIxcEplbW93UkVGUlkwUlJaMEZGWmpoR01XaHdiWGRGSzFsRFJsaDZha2QwWVZGamNrdzJXRnBXVkFwS2JVVmxOV2xUVEhaSE1WTjVVVk5CWlhjM1YyUk5TMFkyYnpsME9HVXlWRVoxUTJ0NmJFOW9hR3gzY3pKUFNGZGlhVVphYmtaWFEwWjNQVDBLTFMwdExTMUZUa1FnVUZWQ1RFbERJRXRGV1MwdExTMHRDZz09In19fX0=\"}", // base64("hello")
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)
			var entry models.LogEntryAnon
			err := entry.UnmarshalBinary([]byte(tc.jsonEntry))
			require.NoError(err)

			_, err = hashedRekordFromEntry(entry)
			if tc.wantError {
				assert.Error(err)
				return
			}
			assert.NoError(err)
		})
	}
}
