/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package aws

import "testing"

func TestGetAWSInstanceInfo(t *testing.T) {
	//t.Skip("aws validator not implemented")
	testCases := map[string]struct {
		client fakeMetadata
	}
}

type fakeMetadataClient struct {
	projectIDString    string
	instanceNameString string
	zoneString         string
	projecIDErr        error
	instanceNameErr    error
	zoneErr            error
}
