package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/genproto/googleapis/cloud/compute/v1"
	"google.golang.org/protobuf/proto"
)

func TestGetMetadataByKey(t *testing.T) {
	testCases := map[string]struct {
		metadata  *compute.Metadata
		key       string
		wantValue string
	}{
		"metadata has key": {
			metadata: &compute.Metadata{
				Items: []*compute.Items{
					{Key: proto.String("key"), Value: proto.String("value")},
				},
			},
			key:       "key",
			wantValue: "value",
		},
		"metadata does not have key": {
			metadata: &compute.Metadata{
				Items: []*compute.Items{
					{Key: proto.String("otherkey"), Value: proto.String("value")},
				},
			},
			key:       "key",
			wantValue: "",
		},
		"metadata contains invalid item": {
			metadata: &compute.Metadata{
				Items: []*compute.Items{
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
