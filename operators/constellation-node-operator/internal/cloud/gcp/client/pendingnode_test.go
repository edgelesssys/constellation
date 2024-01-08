/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package client

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"cloud.google.com/go/compute/apiv1/computepb"
	updatev1alpha1 "github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/googleapi"
	"google.golang.org/protobuf/proto"
)

func TestGetNodeState(t *testing.T) {
	testCases := map[string]struct {
		providerID     string
		getInstanceErr error
		instanceStatus *string
		wantNodeState  updatev1alpha1.CSPNodeState
		wantErr        bool
	}{
		"node is deleted and API returns 404": {
			providerID: "gce://project/zone/instance-name",
			getInstanceErr: &googleapi.Error{
				Code: http.StatusNotFound,
			},
			wantNodeState: updatev1alpha1.NodeStateTerminated,
		},
		"splitting providerID fails": {
			providerID: "invalid",
			wantErr:    true,
		},
		"node is deleted and API returns other error": {
			providerID:     "gce://project/zone/instance-name",
			getInstanceErr: errors.New("get instance error"),
			wantErr:        true,
		},
		"instance has no status": {
			providerID:    "gce://project/zone/instance-name",
			wantNodeState: updatev1alpha1.NodeStateUnknown,
		},
		"instance is provisioning": {
			providerID:     "gce://project/zone/instance-name",
			instanceStatus: proto.String("PROVISIONING"),
			wantNodeState:  updatev1alpha1.NodeStateCreating,
		},
		"instance is staging": {
			providerID:     "gce://project/zone/instance-name",
			instanceStatus: proto.String("STAGING"),
			wantNodeState:  updatev1alpha1.NodeStateCreating,
		},
		"instance is running": {
			providerID:     "gce://project/zone/instance-name",
			instanceStatus: proto.String("RUNNING"),
			wantNodeState:  updatev1alpha1.NodeStateReady,
		},
		"instance is stopping": {
			providerID:     "gce://project/zone/instance-name",
			instanceStatus: proto.String("STOPPING"),
			wantNodeState:  updatev1alpha1.NodeStateStopped,
		},
		"instance is suspending": {
			providerID:     "gce://project/zone/instance-name",
			instanceStatus: proto.String("SUSPENDING"),
			wantNodeState:  updatev1alpha1.NodeStateStopped,
		},
		"instance is suspended": {
			providerID:     "gce://project/zone/instance-name",
			instanceStatus: proto.String("SUSPENDED"),
			wantNodeState:  updatev1alpha1.NodeStateStopped,
		},
		"instance is repairing": {
			providerID:     "gce://project/zone/instance-name",
			instanceStatus: proto.String("REPAIRING"),
			wantNodeState:  updatev1alpha1.NodeStateStopped,
		},
		"instance terminated": {
			providerID:     "gce://project/zone/instance-name",
			instanceStatus: proto.String("TERMINATED"),
			wantNodeState:  updatev1alpha1.NodeStateStopped,
		},
		"instance state unknown": {
			providerID:     "gce://project/zone/instance-name",
			instanceStatus: proto.String("unknown"),
			wantNodeState:  updatev1alpha1.NodeStateUnknown,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			client := Client{
				instanceAPI: &stubInstanceAPI{
					getErr: tc.getInstanceErr,
					instance: &computepb.Instance{
						Status: tc.instanceStatus,
					},
				},
			}
			nodeState, err := client.GetNodeState(context.Background(), tc.providerID)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantNodeState, nodeState)
		})
	}
}
