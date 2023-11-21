/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cloudcmd

import (
	"testing"

	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/stretchr/testify/assert"
)

func TestNormalizeAzureURIs(t *testing.T) {
	testCases := map[string]struct {
		in   *terraform.AzureClusterVariables
		want *terraform.AzureClusterVariables
	}{
		"empty": {
			in:   &terraform.AzureClusterVariables{},
			want: &terraform.AzureClusterVariables{},
		},
		"no change": {
			in: &terraform.AzureClusterVariables{
				ImageID: "/communityGalleries/foo/images/constellation/versions/2.1.0",
			},
			want: &terraform.AzureClusterVariables{
				ImageID: "/communityGalleries/foo/images/constellation/versions/2.1.0",
			},
		},
		"fix image id": {
			in: &terraform.AzureClusterVariables{
				ImageID: "/CommunityGalleries/foo/Images/constellation/Versions/2.1.0",
			},
			want: &terraform.AzureClusterVariables{
				ImageID: "/communityGalleries/foo/images/constellation/versions/2.1.0",
			},
		},
		"fix resource group": {
			in: &terraform.AzureClusterVariables{
				UserAssignedIdentity: "/subscriptions/foo/resourcegroups/test/providers/Microsoft.ManagedIdentity/userAssignedIdentities/uai",
			},
			want: &terraform.AzureClusterVariables{
				UserAssignedIdentity: "/subscriptions/foo/resourceGroups/test/providers/Microsoft.ManagedIdentity/userAssignedIdentities/uai",
			},
		},
		"fix arbitrary casing": {
			in: &terraform.AzureClusterVariables{
				ImageID:              "/CoMMUnitygaLLeries/foo/iMAges/constellation/vERsions/2.1.0",
				UserAssignedIdentity: "/subsCRiptions/foo/resoURCegroups/test/proViDers/MICROsoft.mANAgedIdentity/USerASsignediDENtities/uai",
			},
			want: &terraform.AzureClusterVariables{
				ImageID:              "/communityGalleries/foo/images/constellation/versions/2.1.0",
				UserAssignedIdentity: "/subscriptions/foo/resourceGroups/test/providers/Microsoft.ManagedIdentity/userAssignedIdentities/uai",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			out := normalizeAzureURIs(tc.in)
			assert.Equal(tc.want, out)
		})
	}
}
