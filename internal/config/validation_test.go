/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package config

import (
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/semver"
	"github.com/stretchr/testify/assert"
)

// TestValidateVersionCompatibilityHelper checks that basic version and image short paths are correctly validated.
func TestValidateVersionCompatibilityHelper(t *testing.T) {
	testCases := map[string]struct {
		cli       semver.Semver
		target    string
		wantError bool
	}{
		"full version works": {
			cli:    semver.NewFromInt(0, 1, 0, ""),
			target: "v0.0.0",
		},
		"short path works": {
			cli:    semver.NewFromInt(0, 1, 0, ""),
			target: "ref/main/stream/debug/v0.0.0-pre.0.20230109121528-d24fac00f018",
		},
		"minor version difference > 1": {
			cli:       semver.NewFromInt(0, 0, 0, ""),
			target:    "ref/main/stream/debug/v0.2.0-pre.0.20230109121528-d24fac00f018",
			wantError: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			err := validateImageCompatibilityHelper(tc.cli, "image", tc.target)
			if tc.wantError {
				assert.Error(err)
				return
			}
			assert.NoError(err)
		})
	}
}

func TestValidateMicroserviceVersion(t *testing.T) {
	testCases := map[string]struct {
		cli       semver.Semver
		services  semver.Semver
		wantError bool
	}{
		"success": {
			cli:      semver.NewFromInt(0, 1, 0, ""),
			services: semver.NewFromInt(0, 0, 0, ""),
		},
		"minor version difference > 1": {
			cli:       semver.NewFromInt(0, 0, 0, ""),
			services:  semver.NewFromInt(0, 2, 0, "pre.0.20230109121528-d24fac00f018"),
			wantError: true,
		},
		"major version difference": {
			cli:       semver.NewFromInt(0, 0, 0, ""),
			services:  semver.NewFromInt(1, 0, 0, ""),
			wantError: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			err := ValidateMicroserviceVersion(tc.cli, tc.services)
			if tc.wantError {
				assert.Error(err)
				return
			}
			assert.NoError(err)
		})
	}
}
