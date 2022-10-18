/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package aws

// func testTpmEnabled(t *testing.T) {
//
// 	mockInstanceIdentityDocument := InstanceIdentityDocument{
// 		[1]string{"devpayProductCodes"},
// 		[1]string{"marketplaceProductCodes"},
// 		"privateIp",
// 		"version",
// 		"region",
// 		"instanceId",
// 		[1]string{"billingProducts",
// 		"instanceType",
// 		"accountId",
// 		time.Now(),
// 		"imageId",
// 		"kernelId",
// 		"ramdiskId",
// 		"architecture",
// 	}
//
// 	testCases := map[string]struct {
// 		idDocument imds.GetInstanceIdentityDocumentOutput
// 		wantErr bool
// 	}{
// 		"is enabled": {
// 			idDoc: imds.GetInstanceIdentityDocumentOutput{
// 				&mockInstanceIdentityDocument,
// 				nil,
// 			},
// 		},
// 		"is disabled": {
// 			iDoc
// 		}
// 	}
// }
