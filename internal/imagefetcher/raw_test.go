/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package imagefetcher

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"path"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestShouldDownload(t *testing.T) {
	testCases := map[string]struct {
		partfile, destination string
		wantDownload          bool
	}{
		"no files exist yet": {
			wantDownload: true,
		},
		"partial download": {
			partfile:     "some data",
			wantDownload: true,
		},
		"download succeeded": {
			destination: "all of the data",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			downloader := &Downloader{
				fs: newDownloaderStubFs(t, "someVersion", tc.partfile, tc.destination),
			}
			gotDownload := downloader.shouldDownload("someVersion.raw")
			assert.Equal(tc.wantDownload, gotDownload)
		})
	}
}

func TestDownloadWithProgress(t *testing.T) {
	rawImage := "raw image"
	client := newTestClient(func(req *http.Request) *http.Response {
		if req.URL.String() == "https://cdn.example.com/image.raw" {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(rawImage)),
				Header:     make(http.Header),
			}
		}
		return &http.Response{
			StatusCode: http.StatusNotFound,
			Body:       io.NopCloser(bytes.NewBufferString("Not found.")),
			Header:     make(http.Header),
		}
	})

	testCases := map[string]struct {
		source  string
		wantErr bool
	}{
		"correct file requested": {
			source: "https://cdn.example.com/image.raw",
		},
		"incorrect file requested": {
			source:  "https://cdn.example.com/incorrect.raw",
			wantErr: true,
		},
		"invalid scheme": {
			source:  "xyz://",
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			fs := newDownloaderStubFs(t, "someVersion", "", "")
			downloader := &Downloader{
				httpc: client,
				fs:    fs,
			}
			var outBuffer bytes.Buffer
			err := downloader.downloadWithProgress(context.Background(), &outBuffer, false, tc.source, "someVersion.raw")
			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			out, err := fs.ReadFile("someVersion.raw")
			assert.NoError(err)
			assert.Equal(rawImage, string(out))
		})
	}
}

func TestDownload(t *testing.T) {
	rawImage := "raw image"
	cwd, err := os.Getwd()
	assert.NoError(t, err)
	wantDestination := path.Join(cwd, "someVersion.raw")
	client := newTestClient(func(req *http.Request) *http.Response {
		if req.URL.String() == "https://cdn.example.com/image.raw" {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(rawImage)),
				Header:     make(http.Header),
			}
		}
		return &http.Response{
			StatusCode: http.StatusNotFound,
			Body:       io.NopCloser(bytes.NewBufferString("Not found.")),
			Header:     make(http.Header),
		}
	})

	testCases := map[string]struct {
		source       string
		destination  string
		overrideFile string
		wantErr      bool
	}{
		"correct file requested": {
			source: "https://cdn.example.com/image.raw",
		},
		"file url": {
			source:       "file:///override.raw",
			overrideFile: "override image",
		},
		"file exists": {
			source:      "https://cdn.example.com/image.raw",
			destination: "already exists",
		},
		"incorrect file requested": {
			source:  "https://cdn.example.com/incorrect.raw",
			wantErr: true,
		},
		"invalid scheme": {
			source:  "xyz://",
			wantErr: true,
		},
		"invalid URL": {
			source:  "\x00",
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			fs := newDownloaderStubFs(t, cwd+"/someVersion", "", tc.destination)
			if tc.overrideFile != "" {
				must(t, fs.WriteFile("/override.raw", []byte(tc.overrideFile), os.ModePerm))
			}
			downloader := &Downloader{
				httpc: client,
				fs:    fs,
			}
			var outBuffer bytes.Buffer
			gotDestination, err := downloader.Download(context.Background(), &outBuffer, false, tc.source, "someVersion")
			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			if tc.overrideFile == "" {
				assert.Equal(wantDestination, gotDestination)
			} else {
				assert.Equal("/override.raw", gotDestination)
			}
			out, err := fs.ReadFile(gotDestination)
			assert.NoError(err)
			switch {
			case tc.overrideFile != "":
				assert.Equal(tc.overrideFile, string(out))
			case tc.destination != "":
				assert.Equal(tc.destination, string(out))
			default:
				assert.Equal(rawImage, string(out))
			}
		})
	}
}

func newDownloaderStubFs(t *testing.T, version, partfile, destination string) *afero.Afero {
	fs := afero.NewMemMapFs()
	if partfile != "" {
		must(t, afero.WriteFile(fs, version+".raw.part", []byte(partfile), os.ModePerm))
	}
	if destination != "" {
		must(t, afero.WriteFile(fs, version+".raw", []byte(destination), os.ModePerm))
	}
	return &afero.Afero{Fs: fs}
}
