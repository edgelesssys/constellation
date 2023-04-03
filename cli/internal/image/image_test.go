/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package image

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/versionsapi"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestGetReference(t *testing.T) {
	testCases := map[string]struct {
		info          versionsapi.ImageInfo
		provider      cloudprovider.Provider
		variant       string
		wantReference string
		wantErr       bool
	}{
		"reference exists aws": {
			info:          versionsapi.ImageInfo{AWS: map[string]string{"someVariant": "someReference"}},
			provider:      cloudprovider.AWS,
			variant:       "someVariant",
			wantReference: "someReference",
		},
		"reference exists azure": {
			info:          versionsapi.ImageInfo{Azure: map[string]string{"someVariant": "someReference"}},
			provider:      cloudprovider.Azure,
			variant:       "someVariant",
			wantReference: "someReference",
		},
		"reference exists gcp": {
			info:          versionsapi.ImageInfo{GCP: map[string]string{"someVariant": "someReference"}},
			provider:      cloudprovider.GCP,
			variant:       "someVariant",
			wantReference: "someReference",
		},
		"reference exists openstack": {
			info:          versionsapi.ImageInfo{OpenStack: map[string]string{"someVariant": "someReference"}},
			provider:      cloudprovider.OpenStack,
			variant:       "someVariant",
			wantReference: "someReference",
		},
		"reference exists qemu": {
			info:          versionsapi.ImageInfo{QEMU: map[string]string{"someVariant": "someReference"}},
			provider:      cloudprovider.QEMU,
			variant:       "someVariant",
			wantReference: "someReference",
		},
		"csp does not exist": {
			info:     versionsapi.ImageInfo{AWS: map[string]string{"someVariant": "someReference"}},
			provider: cloudprovider.Unknown,
			variant:  "someVariant",
			wantErr:  true,
		},
		"variant does not exist": {
			info:     versionsapi.ImageInfo{AWS: map[string]string{"someVariant": "someReference"}},
			provider: cloudprovider.AWS,
			variant:  "nonExistingVariant",
			wantErr:  true,
		},
		"info is empty": {
			info:     versionsapi.ImageInfo{},
			provider: cloudprovider.AWS,
			variant:  "someVariant",
			wantErr:  true,
		},
		"csp is nil": {
			info:     versionsapi.ImageInfo{AWS: nil},
			provider: cloudprovider.AWS,
			variant:  "someVariant",
			wantErr:  true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			reference, err := getReferenceFromImageInfo(tc.provider, tc.variant, tc.info)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantReference, reference)
		})
	}
}

func TestImageVariant(t *testing.T) {
	testCases := map[string]struct {
		csp         cloudprovider.Provider
		config      *config.Config
		wantVariant string
		wantErr     bool
	}{
		"AWS region": {
			csp: cloudprovider.AWS,
			config: &config.Config{Image: "someImage", Provider: config.ProviderConfig{
				AWS: &config.AWSConfig{Region: "someRegion"},
			}},
			wantVariant: "someRegion",
		},
		"Azure cvm": {
			csp: cloudprovider.Azure,
			config: &config.Config{
				Image: "someImage", Provider: config.ProviderConfig{Azure: &config.AzureConfig{}},
				Attestation: config.AttestationConfig{AzureSEVSNP: &config.AzureSEVSNP{}},
			},
			wantVariant: "cvm",
		},
		"Azure trustedlaunch": {
			csp: cloudprovider.Azure,
			config: &config.Config{
				Image: "someImage", Provider: config.ProviderConfig{Azure: &config.AzureConfig{}},
				Attestation: config.AttestationConfig{AzureTrustedLaunch: &config.AzureTrustedLaunch{}},
			},
			wantVariant: "trustedlaunch",
		},
		"GCP": {
			csp: cloudprovider.GCP,
			config: &config.Config{Image: "someImage", Provider: config.ProviderConfig{
				GCP: &config.GCPConfig{},
			}},
			wantVariant: "sev-es",
		},
		"OpenStack": {
			csp: cloudprovider.OpenStack,
			config: &config.Config{Image: "someImage", Provider: config.ProviderConfig{
				OpenStack: &config.OpenStackConfig{},
			}},
			wantVariant: "sev",
		},
		"QEMU": {
			csp: cloudprovider.QEMU,
			config: &config.Config{Image: "someImage", Provider: config.ProviderConfig{
				QEMU: &config.QEMUConfig{},
			}},
			wantVariant: "default",
		},
		"invalid": {
			csp:     cloudprovider.Provider(9999),
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			vari, err := imageVariant(tc.csp, tc.config)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantVariant, vari)
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
			QEMU:    map[string]string{"default": "someReference"},
			AWS:     map[string]string{"foo": "bar"},
			Azure:   map[string]string{"foo": "bar"},
			GCP:     map[string]string{"foo": "bar"},
		}
	}
	imgInfoPath := imageInfoFilename(newImgInfo())

	testCases := map[string]struct {
		config           *config.Config
		imageInfoFetcher versionsAPIImageInfoFetcher
		localFile        []byte
		wantReference    string
		wantErr          bool
	}{
		"reference fetched remotely": {
			config: &config.Config{
				Image:    img,
				Provider: config.ProviderConfig{QEMU: &config.QEMUConfig{}},
			},
			imageInfoFetcher: &stubVersionsAPIImageFetcher{
				fetchImageInfoInfo: newImgInfo(),
			},
			wantReference: "someReference",
		},
		"reference fetched remotely fails": {
			config: &config.Config{
				Image:    img,
				Provider: config.ProviderConfig{QEMU: &config.QEMUConfig{}},
			},
			imageInfoFetcher: &stubVersionsAPIImageFetcher{
				fetchImageInfoErr: errors.New("failed"),
			},
			wantErr: true,
		},
		"reference fetched locally": {
			config: &config.Config{
				Image:    img,
				Provider: config.ProviderConfig{QEMU: &config.QEMUConfig{}},
			},
			localFile: func() []byte {
				info := newImgInfo()
				info.QEMU["default"] = "localOverrideReference"
				file, err := json.Marshal(info)
				require.NoError(t, err)
				return file
			}(),
			wantReference: "localOverrideReference",
		},
		"local file first": {
			config: &config.Config{
				Image:    img,
				Provider: config.ProviderConfig{QEMU: &config.QEMUConfig{}},
			},
			imageInfoFetcher: &stubVersionsAPIImageFetcher{
				fetchImageInfoInfo: newImgInfo(),
			},
			localFile: func() []byte {
				info := newImgInfo()
				info.QEMU["default"] = "localOverrideReference"
				file, err := json.Marshal(info)
				require.NoError(t, err)
				return file
			}(),
			wantReference: "localOverrideReference",
		},
		"local file is invalid": {
			config: &config.Config{
				Image:    img,
				Provider: config.ProviderConfig{QEMU: &config.QEMUConfig{}},
			},
			localFile: []byte("invalid"),
			wantErr:   true,
		},
		"local file has invalid image info": {
			config: &config.Config{
				Image:    img,
				Provider: config.ProviderConfig{QEMU: &config.QEMUConfig{}},
			},
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
			config: &config.Config{
				Image:    "nonExistingImageName",
				Provider: config.ProviderConfig{QEMU: &config.QEMUConfig{}},
			},
			wantErr: true,
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

			reference, err := fetcher.FetchReference(context.Background(), tc.config)

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
