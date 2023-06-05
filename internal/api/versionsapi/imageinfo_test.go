/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package versionsapi

import (
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/stretchr/testify/assert"
)

func TestImageInfoJSONPath(t *testing.T) {
	testCases := map[string]struct {
		info     ImageInfo
		wantPath string
	}{
		"image info": {
			info: ImageInfo{
				Ref:     "test-ref",
				Stream:  "nightly",
				Version: "v1.0.0",
			},
			wantPath: constants.CDNAPIPrefixV2 + "/ref/test-ref/stream/nightly/v1.0.0/image/info.json",
		},
		"image info release": {
			info: ImageInfo{
				Ref:     ReleaseRef,
				Stream:  "stable",
				Version: "v1.0.0",
			},
			wantPath: constants.CDNAPIPrefixV2 + "/ref/-/stream/stable/v1.0.0/image/info.json",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.wantPath, tc.info.JSONPath())
		})
	}
}

func TestImageInfoURL(t *testing.T) {
	testCases := map[string]struct {
		info     ImageInfo
		wantURL  string
		wantPath string
	}{
		"image info": {
			info: ImageInfo{
				Ref:     "test-ref",
				Stream:  "nightly",
				Version: "v1.0.0",
			},
			wantURL: constants.CDNRepositoryURL + "/" + constants.CDNAPIPrefixV2 + "/ref/test-ref/stream/nightly/v1.0.0/image/info.json",
		},
		"image info release": {
			info: ImageInfo{
				Ref:     ReleaseRef,
				Stream:  "stable",
				Version: "v1.0.0",
			},
			wantURL: constants.CDNRepositoryURL + "/" + constants.CDNAPIPrefixV2 + "/ref/-/stream/stable/v1.0.0/image/info.json",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			url, err := tc.info.URL()
			assert.NoError(err)
			assert.Equal(tc.wantURL, url)
		})
	}
}

func TestImageInfoValidate(t *testing.T) {
	testCases := map[string]struct {
		info    ImageInfo
		wantErr bool
	}{
		"valid image info": {
			info: ImageInfo{
				Ref:     "test-ref",
				Stream:  "nightly",
				Version: "v1.0.0",
				List: []ImageInfoEntry{
					{
						CSP:                "aws",
						AttestationVariant: "aws-nitro-tpm",
						Reference:          "ami-123",
						Region:             "us-east-1",
					},
					{
						CSP:                "aws",
						AttestationVariant: "aws-nitro-tpm",
						Reference:          "ami-123",
						Region:             "us-east-2",
					},
					{
						CSP:                "gcp",
						AttestationVariant: "gcp-sev-es",
						Reference:          "gcp-123",
					},
					{
						CSP:                "azure",
						AttestationVariant: "azure-sev-snp",
						Reference:          "azure-123",
					},
					{
						CSP:                "qemu",
						AttestationVariant: "qemu-vtpm",
						Reference:          "https://example.com/qemu-123/image.raw",
					},
				},
			},
		},
		"invalid ref": {
			info: ImageInfo{
				Ref:     "",
				Stream:  "nightly",
				Version: "v1.0.0",
				List: []ImageInfoEntry{
					{
						CSP:                "aws",
						AttestationVariant: "aws-nitro-tpm",
						Reference:          "ami-123",
						Region:             "us-east-1",
					},
				},
			},
			wantErr: true,
		},
		"invalid stream": {
			info: ImageInfo{
				Ref:     "test-ref",
				Stream:  "",
				Version: "v1.0.0",
				List: []ImageInfoEntry{
					{
						CSP:                "aws",
						AttestationVariant: "aws-nitro-tpm",
						Reference:          "ami-123",
						Region:             "us-east-1",
					},
				},
			},
			wantErr: true,
		},
		"invalid version": {
			info: ImageInfo{
				Ref:     "test-ref",
				Stream:  "nightly",
				Version: "",
				List: []ImageInfoEntry{
					{
						CSP:                "aws",
						AttestationVariant: "aws-nitro-tpm",
						Reference:          "ami-123",
						Region:             "us-east-1",
					},
				},
			},
			wantErr: true,
		},
		"no entries in list": {
			info: ImageInfo{
				Ref:     "test-ref",
				Stream:  "nightly",
				Version: "v1.0.0",
			},
			wantErr: true,
		},
		"multiple errors": {
			info: ImageInfo{
				Ref:     "",
				Stream:  "",
				Version: "",
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			err := tc.info.Validate()

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

func TestImageInfoValidateRequest(t *testing.T) {
	testCases := map[string]struct {
		info    ImageInfo
		wantErr bool
	}{
		"valid image info": {
			info: ImageInfo{
				Ref:     "test-ref",
				Stream:  "nightly",
				Version: "v1.0.0",
			},
		},
		"invalid ref": {
			info: ImageInfo{
				Ref:     "",
				Stream:  "nightly",
				Version: "v1.0.0",
			},
			wantErr: true,
		},
		"invalid stream": {
			info: ImageInfo{
				Ref:     "test-ref",
				Stream:  "",
				Version: "v1.0.0",
			},
			wantErr: true,
		},
		"invalid version": {
			info: ImageInfo{
				Ref:     "test-ref",
				Stream:  "nightly",
				Version: "",
			},
			wantErr: true,
		},
		"request contains entries": {
			info: ImageInfo{
				Ref:     "test-ref",
				Stream:  "nightly",
				Version: "v1.0.0",
				List: []ImageInfoEntry{
					{
						CSP:                "aws",
						AttestationVariant: "aws-nitro-tpm",
						Reference:          "ami-123",
						Region:             "us-east-1",
					},
				},
			},
			wantErr: true,
		},
		"multiple errors": {
			info: ImageInfo{
				Ref:     "",
				Stream:  "",
				Version: "",
				List: []ImageInfoEntry{
					{
						CSP:                "aws",
						AttestationVariant: "aws-nitro-tpm",
						Reference:          "ami-123",
						Region:             "us-east-1",
					},
				},
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			err := tc.info.ValidateRequest()

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

func TestMergeImageInfos(t *testing.T) {
	testCases := map[string]struct {
		infos    []ImageInfo
		wantInfo ImageInfo
		wantErr  bool
	}{
		"only one element": {
			infos: []ImageInfo{
				{
					Ref:     "test-ref",
					Stream:  "nightly",
					Version: "v1.0.0",
					List: []ImageInfoEntry{
						{
							CSP:                "aws",
							AttestationVariant: "aws-nitro-tpm",
							Reference:          "ami-123",
							Region:             "us-east-1",
						},
					},
				},
			},
			wantInfo: ImageInfo{
				Ref:     "test-ref",
				Stream:  "nightly",
				Version: "v1.0.0",
				List: []ImageInfoEntry{
					{
						CSP:                "aws",
						AttestationVariant: "aws-nitro-tpm",
						Reference:          "ami-123",
						Region:             "us-east-1",
					},
				},
			},
		},
		"valid image info": {
			infos: []ImageInfo{
				{
					Ref:     "test-ref",
					Stream:  "nightly",
					Version: "v1.0.0",
					List: []ImageInfoEntry{
						{
							CSP:                "aws",
							AttestationVariant: "aws-nitro-tpm",
							Reference:          "ami-123",
							Region:             "us-east-1",
						},
					},
				},
				{
					Ref:     "test-ref",
					Stream:  "nightly",
					Version: "v1.0.0",
					List: []ImageInfoEntry{
						{
							CSP:                "gcp",
							AttestationVariant: "gcp-sev-es",
							Reference:          "image-123",
						},
					},
				},
			},
			wantInfo: ImageInfo{
				Ref:     "test-ref",
				Stream:  "nightly",
				Version: "v1.0.0",
				List: []ImageInfoEntry{
					{
						CSP:                "aws",
						AttestationVariant: "aws-nitro-tpm",
						Reference:          "ami-123",
						Region:             "us-east-1",
					},
					{
						CSP:                "gcp",
						AttestationVariant: "gcp-sev-es",
						Reference:          "image-123",
					},
				},
			},
		},
		"sorting": {
			infos: []ImageInfo{
				{
					Ref:     "test-ref",
					Stream:  "nightly",
					Version: "v1.0.0",
					List: []ImageInfoEntry{
						{
							CSP:                "gcp",
							AttestationVariant: "gcp-sev-es",
							Reference:          "image-123",
						},
					},
				},
				{
					Ref:     "test-ref",
					Stream:  "nightly",
					Version: "v1.0.0",
					List: []ImageInfoEntry{
						{
							CSP:                "aws",
							AttestationVariant: "aws-nitro-tpm",
							Reference:          "ami-123",
							Region:             "us-east-1",
						},
					},
				},
			},
			wantInfo: ImageInfo{
				Ref:     "test-ref",
				Stream:  "nightly",
				Version: "v1.0.0",
				List: []ImageInfoEntry{
					{
						CSP:                "aws",
						AttestationVariant: "aws-nitro-tpm",
						Reference:          "ami-123",
						Region:             "us-east-1",
					},
					{
						CSP:                "gcp",
						AttestationVariant: "gcp-sev-es",
						Reference:          "image-123",
					},
				},
			},
		},
		"mismatch in base info": {
			infos: []ImageInfo{
				{
					Ref:     "test-ref",
					Stream:  "nightly",
					Version: "v1.0.0",
					List: []ImageInfoEntry{
						{
							CSP:                "gcp",
							AttestationVariant: "gcp-sev-es",
							Reference:          "image-123",
						},
					},
				},
				{
					Ref:     "other-ref",
					Stream:  "stable",
					Version: "v2.0.0",
					List: []ImageInfoEntry{
						{
							CSP:                "aws",
							AttestationVariant: "aws-nitro-tpm",
							Reference:          "ami-123",
							Region:             "us-east-1",
						},
					},
				},
			},
			wantErr: true,
		},
		"empty list": {
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			gotInfo, err := MergeImageInfos(tc.infos...)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			assert.Equal(tc.wantInfo, gotInfo)
		})
	}
}
