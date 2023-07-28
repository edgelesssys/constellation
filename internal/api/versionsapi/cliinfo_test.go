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

func TestCLIInfoJSONPath(t *testing.T) {
	testCases := map[string]struct {
		info     CLIInfo
		wantPath string
	}{
		"cli info": {
			info: CLIInfo{
				Ref:     "test-ref",
				Stream:  "nightly",
				Version: "v1.0.0",
			},
			wantPath: constants.CDNAPIPrefix + "/ref/test-ref/stream/nightly/v1.0.0/cli/info.json",
		},
		"cli info release": {
			info: CLIInfo{
				Ref:     constants.ReleaseRef,
				Stream:  "stable",
				Version: "v1.0.0",
			},
			wantPath: constants.CDNAPIPrefix + "/ref/-/stream/stable/v1.0.0/cli/info.json",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.wantPath, tc.info.JSONPath())
		})
	}
}

func TestCLIInfoURL(t *testing.T) {
	testCases := map[string]struct {
		info     CLIInfo
		wantURL  string
		wantPath string
	}{
		"cli info": {
			info: CLIInfo{
				Ref:     "test-ref",
				Stream:  "nightly",
				Version: "v1.0.0",
			},
			wantURL: constants.CDNRepositoryURL + "/" + constants.CDNAPIPrefix + "/ref/test-ref/stream/nightly/v1.0.0/cli/info.json",
		},
		"cli info release": {
			info: CLIInfo{
				Ref:     constants.ReleaseRef,
				Stream:  "stable",
				Version: "v1.0.0",
			},
			wantURL: constants.CDNRepositoryURL + "/" + constants.CDNAPIPrefix + "/ref/-/stream/stable/v1.0.0/cli/info.json",
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

func TestCLIInfoValidate(t *testing.T) {
	testCases := map[string]struct {
		info    CLIInfo
		wantErr bool
	}{
		"valid cli info": {
			info: CLIInfo{
				Ref:        "test-ref",
				Stream:     "nightly",
				Version:    "v1.0.0",
				Kubernetes: []string{"v1.26.1", "v1.3.3", "v1.32"},
			},
		},
		"invalid ref": {
			info: CLIInfo{
				Ref:        "",
				Stream:     "nightly",
				Version:    "v1.0.0",
				Kubernetes: []string{"v1.26.1", "v1.3.3", "v1.32"},
			},
			wantErr: true,
		},
		"invalid stream": {
			info: CLIInfo{
				Ref:        "test-ref",
				Stream:     "",
				Version:    "v1.0.0",
				Kubernetes: []string{"v1.26.1", "v1.3.3", "v1.32"},
			},
			wantErr: true,
		},
		"invalid version": {
			info: CLIInfo{
				Ref:        "test-ref",
				Stream:     "nightly",
				Version:    "",
				Kubernetes: []string{"v1.26.1", "v1.3.3", "v1.32"},
			},
			wantErr: true,
		},
		"invalid k8s versions": {
			info: CLIInfo{
				Ref:        "test-ref",
				Stream:     "nightly",
				Version:    "v1.0.0",
				Kubernetes: []string{"1", "", "1.32"},
			},
			wantErr: true,
		},
		"multiple errors": {
			info: CLIInfo{
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

func TestCLIInfoValidateRequest(t *testing.T) {
	testCases := map[string]struct {
		info    CLIInfo
		wantErr bool
	}{
		"valid cli info": {
			info: CLIInfo{
				Ref:     "test-ref",
				Stream:  "nightly",
				Version: "v1.0.0",
			},
		},
		"invalid ref": {
			info: CLIInfo{
				Ref:     "",
				Stream:  "nightly",
				Version: "v1.0.0",
			},
			wantErr: true,
		},
		"invalid stream": {
			info: CLIInfo{
				Ref:     "test-ref",
				Stream:  "",
				Version: "v1.0.0",
			},
			wantErr: true,
		},
		"invalid version": {
			info: CLIInfo{
				Ref:     "test-ref",
				Stream:  "nightly",
				Version: "",
			},
			wantErr: true,
		},
		"invalid k8s versions": {
			info: CLIInfo{
				Ref:        "test-ref",
				Stream:     "nightly",
				Version:    "v1.0.0",
				Kubernetes: []string{"v1.26.1", "v1.3.3", "v1.32"},
			},
			wantErr: true,
		},
		"multiple errors": {
			info: CLIInfo{
				Ref:        "",
				Stream:     "",
				Version:    "",
				Kubernetes: []string{"v1.26.1", "v1.3.3", "v1.32"},
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
