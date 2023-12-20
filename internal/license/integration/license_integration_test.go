//go:build integration

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package integration

import (
	"context"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/license"
	"github.com/stretchr/testify/assert"
)

func TestQuotaCheckIntegration(t *testing.T) {
	testCases := map[string]struct {
		license   string
		wantQuota int
		wantError bool
	}{
		"OSS license has quota 8": {
			license:   license.CommunityLicense,
			wantQuota: 8,
		},
		"Empty license assumes community": {
			license:   "",
			wantQuota: 8,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			client := license.NewChecker()

			quota, err := client.CheckLicense(context.Background(), cloudprovider.Unknown, "test", tc.license)

			if tc.wantError {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			assert.Equal(tc.wantQuota, quota)
		})
	}
}
