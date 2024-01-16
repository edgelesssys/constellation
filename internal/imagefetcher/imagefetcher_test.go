/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package imagefetcher

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/api/versionsapi"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m, goleak.IgnoreAnyFunction("github.com/bazelbuild/rules_go/go/tools/bzltestutil.RegisterTimeoutHandler.func1"))
}

func TestGetReference(t *testing.T) {
	testCases := map[string]struct {
		info          versionsapi.ImageInfo
		provider      cloudprovider.Provider
		variant       string
		filter        filter
		wantReference string
		wantErr       bool
	}{
		"reference exists with filter": {
			info: versionsapi.ImageInfo{
				List: []versionsapi.ImageInfoEntry{
					{CSP: "aws", AttestationVariant: "aws-nitro-tpm", Reference: "someReference"},
					{CSP: "aws", AttestationVariant: "aws-nitro-tpm", Reference: "someOtherReference", Region: "someRegion"},
				},
			},
			provider: cloudprovider.AWS,
			variant:  "aws-nitro-tpm",
			filter: func(entry versionsapi.ImageInfoEntry) bool {
				return entry.Region == "someRegion"
			},
			wantReference: "someOtherReference",
		},
		"reference exists aws": {
			info: versionsapi.ImageInfo{
				List: []versionsapi.ImageInfoEntry{
					{CSP: "aws", AttestationVariant: "aws-nitro-tpm", Reference: "someReference"},
				},
			},
			provider:      cloudprovider.AWS,
			variant:       "aws-nitro-tpm",
			wantReference: "someReference",
		},
		"reference exists azure": {
			info: versionsapi.ImageInfo{
				List: []versionsapi.ImageInfoEntry{
					{CSP: "azure", AttestationVariant: "azure-sev-snp", Reference: "someReference"},
				},
			},
			provider:      cloudprovider.Azure,
			variant:       "azure-sev-snp",
			wantReference: "someReference",
		},
		"reference exists gcp": {
			info: versionsapi.ImageInfo{
				List: []versionsapi.ImageInfoEntry{
					{CSP: "gcp", AttestationVariant: "gcp-sev-es", Reference: "someReference"},
				},
			},
			provider:      cloudprovider.GCP,
			variant:       "gcp-sev-es",
			wantReference: "someReference",
		},
		"reference exists openstack": {
			info: versionsapi.ImageInfo{
				List: []versionsapi.ImageInfoEntry{
					{CSP: "openstack", AttestationVariant: "qemu-vtpm", Reference: "someReference"},
				},
			},
			provider:      cloudprovider.OpenStack,
			variant:       "qemu-vtpm",
			wantReference: "someReference",
		},
		"reference exists qemu": {
			info: versionsapi.ImageInfo{
				List: []versionsapi.ImageInfoEntry{
					{CSP: "qemu", AttestationVariant: "qemu-vtpm", Reference: "someReference"},
				},
			},
			provider:      cloudprovider.QEMU,
			variant:       "qemu-vtpm",
			wantReference: "someReference",
		},
		"csp does not exist": {
			info:     versionsapi.ImageInfo{List: []versionsapi.ImageInfoEntry{}},
			provider: cloudprovider.Unknown,
			variant:  "someVariant",
			wantErr:  true,
		},
		"variant does not exist": {
			info: versionsapi.ImageInfo{
				List: []versionsapi.ImageInfoEntry{
					{CSP: "aws", AttestationVariant: "dummy", Reference: "someReference"},
				},
			},
			provider: cloudprovider.AWS,
			variant:  "aws-nitro-tpm",
			wantErr:  true,
		},
		"info is empty": {
			info:     versionsapi.ImageInfo{},
			provider: cloudprovider.AWS,
			variant:  "someVariant",
			wantErr:  true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			var filters []filter
			if tc.filter != nil {
				filters = []filter{tc.filter}
			}
			reference, err := getReferenceFromImageInfo(tc.provider, tc.variant, tc.info, filters...)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantReference, reference)
		})
	}
}

func TestFetchReference(t *testing.T) {
	img := "ref/abc/stream/nightly/v1.2.3"
	newImgInfo := func() versionsapi.ImageInfo {
		return versionsapi.ImageInfo{
			Ref:     "abc",
			Stream:  "nightly",
			Version: "v1.2.3",
			List: []versionsapi.ImageInfoEntry{
				{
					CSP:                "qemu",
					AttestationVariant: "dummy",
					Reference:          "someReference",
				},
			},
		}
	}
	imgInfoPath := imageInfoFilename(newImgInfo())

	testCases := map[string]struct {
		provider         cloudprovider.Provider
		image            string
		imageInfoFetcher versionsAPIImageInfoFetcher
		localFile        []byte
		wantReference    string
		wantErr          bool
	}{
		"reference fetched remotely": {
			provider: cloudprovider.QEMU,
			image:    img,
			imageInfoFetcher: &stubVersionsAPIImageFetcher{
				fetchImageInfoInfo: newImgInfo(),
			},
			wantReference: "someReference",
		},
		"reference fetched remotely fails": {
			provider: cloudprovider.QEMU,
			image:    img,
			imageInfoFetcher: &stubVersionsAPIImageFetcher{
				fetchImageInfoErr: errors.New("failed"),
			},
			wantErr: true,
		},
		"reference fetched locally": {
			provider: cloudprovider.QEMU,
			image:    img,
			localFile: func() []byte {
				info := newImgInfo()
				info.List[0].Reference = "localOverrideReference"
				file, err := json.Marshal(info)
				require.NoError(t, err)
				return file
			}(),
			wantReference: "localOverrideReference",
		},
		"local file first": {
			provider: cloudprovider.QEMU,
			image:    img,
			imageInfoFetcher: &stubVersionsAPIImageFetcher{
				fetchImageInfoInfo: newImgInfo(),
			},
			localFile: func() []byte {
				info := newImgInfo()
				info.List[0].Reference = "localOverrideReference"
				file, err := json.Marshal(info)
				require.NoError(t, err)
				return file
			}(),
			wantReference: "localOverrideReference",
		},
		"local file is invalid": {
			provider:  cloudprovider.QEMU,
			image:     img,
			localFile: []byte("invalid"),
			wantErr:   true,
		},
		"local file has invalid image info": {
			provider: cloudprovider.QEMU,
			image:    img,
			localFile: func() []byte {
				info := newImgInfo()
				info.Ref = ""
				file, err := json.Marshal(info)
				require.NoError(t, err)
				return file
			}(),
			wantErr: true,
		},
		"image version does not exist": {
			provider: cloudprovider.QEMU,
			image:    "nonExistingImageName",
			wantErr:  true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			fs := afero.NewMemMapFs()
			af := &afero.Afero{Fs: fs}
			if tc.localFile != nil {
				fh := file.NewHandler(af)
				require.NoError(fh.Write(imgInfoPath, tc.localFile))
			}

			fetcher := &Fetcher{
				fetcher: tc.imageInfoFetcher,
				fs:      af,
			}

			reference, err := fetcher.FetchReference(context.Background(), tc.provider, variant.Dummy{},
				tc.image, "someRegion", false)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantReference, reference)
		})
	}
}

type stubVersionsAPIImageFetcher struct {
	fetchImageInfoInfo versionsapi.ImageInfo
	fetchImageInfoErr  error
}

func (f *stubVersionsAPIImageFetcher) FetchImageInfo(_ context.Context, _ versionsapi.ImageInfo) (
	versionsapi.ImageInfo, error,
) {
	return f.fetchImageInfoInfo, f.fetchImageInfoErr
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

// roundTripFunc .
type roundTripFunc func(req *http.Request) *http.Response

// RoundTrip .
func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

// newTestClient returns *http.Client with Transport replaced to avoid making real calls.
func newTestClient(fn roundTripFunc) *http.Client {
	return &http.Client{
		Transport: fn,
	}
}
