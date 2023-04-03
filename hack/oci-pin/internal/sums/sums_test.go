/*
Copyright (c) Edgeless Systems GmbH
SPDX-License-Identifier: AGPL-3.0-only
*/

package sums

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateParse(t *testing.T) {
	testCases := map[string]struct {
		refs     []PinnedImageReference
		wantErr  bool
		wantOut  string
		wantRefs []PinnedImageReference
	}{
		"single image": {
			refs: []PinnedImageReference{
				{
					Registry: "registry.example.com",
					Prefix:   "staging",
					Name:     "foo-service",
					Tag:      "v1.2.3",
					Digest:   "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
				},
			},
			wantOut: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855  registry.example.com/staging/foo-service:v1.2.3\n",
		},
		"no prefix": {
			refs: []PinnedImageReference{
				{
					Registry: "registry.example.com",
					Name:     "foo-service",
					Tag:      "v1.2.3",
					Digest:   "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
				},
			},
			wantOut: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855  registry.example.com/foo-service:v1.2.3\n",
		},
		"no tag": {
			refs: []PinnedImageReference{
				{
					Registry: "registry.example.com",
					Prefix:   "staging",
					Name:     "foo-service",
					Digest:   "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
				},
			},
			wantOut: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855  registry.example.com/staging/foo-service\n",
		},
		"multiple images": {
			refs: []PinnedImageReference{
				{
					Registry: "registry.example.com",
					Prefix:   "staging",
					Name:     "foo-service",
					Tag:      "v1.2.3",
					Digest:   "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
				},
				{
					Registry: "registry.example.com",
					Prefix:   "staging",
					Name:     "bar-service",
					Tag:      "v1.2.3",
					Digest:   "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
				},
				{
					Registry: "registry.example.com",
					Prefix:   "production",
					Name:     "baz-service",
					Tag:      "v1.2.3",
					Digest:   "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
				},
				{
					Registry: "registry.example.com",
					Prefix:   "staging",
					Name:     "foo-service",
					Tag:      "v2.0.0",
					Digest:   "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
				},
				{
					Registry: "backup.example.com",
					Prefix:   "staging",
					Name:     "foo-service",
					Tag:      "v1.2.3",
					Digest:   "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
				},
			},
			wantRefs: []PinnedImageReference{
				{
					Registry: "backup.example.com",
					Prefix:   "staging",
					Name:     "foo-service",
					Tag:      "v1.2.3",
					Digest:   "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
				},
				{
					Registry: "registry.example.com",
					Prefix:   "production",
					Name:     "baz-service",
					Tag:      "v1.2.3",
					Digest:   "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
				},
				{
					Registry: "registry.example.com",
					Prefix:   "staging",
					Name:     "bar-service",
					Tag:      "v1.2.3",
					Digest:   "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
				},
				{
					Registry: "registry.example.com",
					Prefix:   "staging",
					Name:     "foo-service",
					Tag:      "v1.2.3",
					Digest:   "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
				},
				{
					Registry: "registry.example.com",
					Prefix:   "staging",
					Name:     "foo-service",
					Tag:      "v2.0.0",
					Digest:   "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
				},
			},
			wantOut: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855  backup.example.com/staging/foo-service:v1.2.3\n" +
				"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855  registry.example.com/production/baz-service:v1.2.3\n" +
				"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855  registry.example.com/staging/bar-service:v1.2.3\n" +
				"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855  registry.example.com/staging/foo-service:v1.2.3\n" +
				"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855  registry.example.com/staging/foo-service:v2.0.0\n",
		},
		"duplicate images": {
			refs: []PinnedImageReference{
				{
					Registry: "registry.example.com",
					Prefix:   "staging",
					Name:     "foo-service",
					Tag:      "v1.2.3",
					Digest:   "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
				},
				{
					Registry: "registry.example.com",
					Prefix:   "staging",
					Name:     "foo-service",
					Tag:      "v1.2.3",
					Digest:   "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
				},
			},
			wantRefs: []PinnedImageReference{
				{
					Registry: "registry.example.com",
					Prefix:   "staging",
					Name:     "foo-service",
					Tag:      "v1.2.3",
					Digest:   "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
				},
			},
			wantOut: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855  registry.example.com/staging/foo-service:v1.2.3\n",
		},
		"duplicate images with different hashes": {
			refs: []PinnedImageReference{
				{
					Registry: "registry.example.com",
					Prefix:   "staging",
					Name:     "foo-service",
					Tag:      "v1.2.3",
					Digest:   "sha256:1111111111111111111111111111111111111111111111111111111111111111",
				},
				{
					Registry: "registry.example.com",
					Prefix:   "staging",
					Name:     "foo-service",
					Tag:      "v1.2.3",
					Digest:   "sha256:0000000000000000000000000000000000000000000000000000000000000000",
				},
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name+"_create", func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			var buf bytes.Buffer
			err := Create(tc.refs, &buf)
			if tc.wantErr {
				require.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantOut, buf.String())
		})
	}

	for name, tc := range testCases {
		if tc.wantErr {
			continue // skip inverse test cases where the forward test case is expected to fail
		}
		t.Run(name+"_parse", func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			buf := bytes.NewBufferString(tc.wantOut)
			gotRefs, err := Parse(buf)
			if tc.wantErr {
				require.Error(err)
				return
			}
			require.NoError(err)
			wantRefs := tc.refs
			if len(tc.wantRefs) > 0 {
				wantRefs = tc.wantRefs
			}
			assert.Equal(wantRefs, gotRefs)
		})
	}
}

func TestMerge(t *testing.T) {
	testCases := map[string]struct {
		refs    [][]PinnedImageReference
		wantErr bool
		wantOut string
	}{
		"different images": {
			refs: [][]PinnedImageReference{
				{
					{
						Registry: "registry.example.com",
						Prefix:   "staging",
						Name:     "foo-service",
						Tag:      "v1.2.3",
						Digest:   "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
					},
				},
				{
					{
						Registry: "registry.example.com",
						Prefix:   "staging",
						Name:     "bar-service",
						Tag:      "v1.2.3",
						Digest:   "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
					},
				},
			},
			wantOut: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855  registry.example.com/staging/bar-service:v1.2.3\n" + "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855  registry.example.com/staging/foo-service:v1.2.3\n",
		},
		"duplicate images": {
			refs: [][]PinnedImageReference{
				{
					{
						Registry: "registry.example.com",
						Prefix:   "staging",
						Name:     "foo-service",
						Tag:      "v1.2.3",
						Digest:   "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
					},
				},
				{
					{
						Registry: "registry.example.com",
						Prefix:   "staging",
						Name:     "foo-service",
						Tag:      "v1.2.3",
						Digest:   "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
					},
				},
			},
			wantOut: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855  registry.example.com/staging/foo-service:v1.2.3\n",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			var buf bytes.Buffer
			err := Merge(tc.refs, &buf)
			if tc.wantErr {
				require.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantOut, buf.String())
		})
	}
}

// TestParse has additional test cases that are not covered by TestCreateParse.
func TestParse(t *testing.T) {
	testCases := map[string]struct {
		in       string
		wantErr  bool
		wantRefs []PinnedImageReference
	}{
		"empty line": {
			in: "\n\n",
		},
		"line too short": {
			in:      "short",
			wantErr: true,
		},
		"malformed digest": {
			in:      "malformed-digest-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa  registry.example.com/staging/foo-service:v1.2.3",
			wantErr: true,
		},
		"missing registry": {
			in:      "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855  foo-service:v1.2.3",
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			buf := bytes.NewBufferString(tc.in)
			gotRefs, err := Parse(buf)
			if tc.wantErr {
				require.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantRefs, gotRefs)
		})
	}
}
