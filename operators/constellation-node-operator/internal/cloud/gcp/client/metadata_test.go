/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package client

import (
	"testing"

	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestGetMetadataByKey(t *testing.T) {
	testCases := map[string]struct {
		metadata  *computepb.Metadata
		key       string
		wantValue string
	}{
		"metadata has key": {
			metadata: &computepb.Metadata{
				Items: []*computepb.Items{
					{Key: proto.String("key"), Value: proto.String("value")},
				},
			},
			key:       "key",
			wantValue: "value",
		},
		"metadata does not have key": {
			metadata: &computepb.Metadata{
				Items: []*computepb.Items{
					{Key: proto.String("otherkey"), Value: proto.String("value")},
				},
			},
			key:       "key",
			wantValue: "",
		},
		"metadata contains invalid item": {
			metadata: &computepb.Metadata{
				Items: []*computepb.Items{
					{},
					{Key: proto.String("key"), Value: proto.String("value")},
				},
			},
			key:       "key",
			wantValue: "value",
		},
		"metadata is nil": {
			key:       "key",
			wantValue: "",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			assert.Equal(tc.wantValue, getMetadataByKey(tc.metadata, tc.key))
		})
	}
}
