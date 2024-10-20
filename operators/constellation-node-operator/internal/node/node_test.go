/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package node

import (
	"testing"

	updatev1alpha1 "github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/api/v1alpha1"
	"github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/internal/constants"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestReady(t *testing.T) {
	testCases := map[string]struct {
		node      corev1.Node
		wantReady bool
	}{
		"node without status conditions": {},
		"node with NodeReady set to false": {
			node: corev1.Node{
				Status: corev1.NodeStatus{
					Conditions: []corev1.NodeCondition{
						{Type: corev1.NodeReady, Status: corev1.ConditionFalse},
					},
				},
			},
		},
		"node with NodeReady set to unknown": {
			node: corev1.Node{
				Status: corev1.NodeStatus{
					Conditions: []corev1.NodeCondition{
						{Type: corev1.NodeReady, Status: corev1.ConditionUnknown},
					},
				},
			},
		},
		"node with NodeReady set to true": {
			node: corev1.Node{
				Status: corev1.NodeStatus{
					Conditions: []corev1.NodeCondition{
						{Type: corev1.NodeReady, Status: corev1.ConditionTrue},
					},
				},
			},
			wantReady: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			assert.Equal(tc.wantReady, Ready(&tc.node))
		})
	}
}

func TestIsControlPlaneNode(t *testing.T) {
	testCases := map[string]struct {
		node             corev1.Node
		wantControlPlane bool
	}{
		"node without labels": {},
		"node with control-plane role label": {
			node: corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{constants.ControlPlaneRoleLabel: ""},
				},
			},
			wantControlPlane: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			assert.Equal(tc.wantControlPlane, IsControlPlaneNode(&tc.node))
		})
	}
}

func TestVPCIP(t *testing.T) {
	testCases := map[string]struct {
		node      corev1.Node
		wantVPCIP string
		wantErr   bool
	}{
		"node without addresses": {
			wantErr: true,
		},
		"node with multiple addresses": {
			node: corev1.Node{
				Status: corev1.NodeStatus{
					Addresses: []corev1.NodeAddress{
						{Type: corev1.NodeExternalIP, Address: "external"},
						{Type: corev1.NodeInternalIP, Address: "192.0.2.1"},
						{Type: corev1.NodeInternalIP, Address: "second-internal"},
					},
				},
			},
			wantVPCIP: "192.0.2.1",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)
			gotIP, err := VPCIP(&tc.node)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantVPCIP, gotIP)
		})
	}
}

func TestFindPending(t *testing.T) {
	testCases := map[string]struct {
		pendingNodes []updatev1alpha1.PendingNode
		node         *corev1.Node
		wantPending  *updatev1alpha1.PendingNode
	}{
		"everything nil": {},
		"node nil": {
			pendingNodes: pendingNodes,
		},
		"node is not in pending nodes list": {
			pendingNodes: pendingNodes,
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "doesnotexist",
				},
			},
		},
		"pending node is leaving": {
			pendingNodes: pendingNodes,
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "leavingnode",
				},
			},
		},
		"pending node is not ready": {
			pendingNodes: pendingNodes,
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "unreadynode",
				},
			},
		},
		"pending node is found": {
			pendingNodes: pendingNodes,
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "joiningnode",
				},
			},
			wantPending: &pendingNodes[0],
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			pending := FindPending(tc.pendingNodes, tc.node)
			if tc.wantPending == nil {
				assert.Nil(pending)
				return
			}
			assert.Equal(*tc.wantPending, *pending)
		})
	}
}

func TestFilterLabels(t *testing.T) {
	labels := map[string]string{
		"key":                                      "value",
		"app.kubernetes.io/component":              "component",
		"app.kubernetes.io/created-by":             "created-by",
		"app.kubernetes.io/instance":               "instance",
		"app.kubernetes.io/managed-by":             "managed-by",
		"app.kubernetes.io/name":                   "name",
		"app.kubernetes.io/part-of":                "part-of",
		"app.kubernetes.io/version":                "version",
		"kubernetes.io/arch":                       "arch",
		"kubernetes.io/os":                         "os",
		"beta.kubernetes.io/arch":                  "arch",
		"beta.kubernetes.io/os":                    "os",
		"kubernetes.io/hostname":                   "hostname",
		"kubernetes.io/change-cause":               "change-cause",
		"kubernetes.io/description":                "description",
		"node.kubernetes.io/instance-type":         "instance-type",
		"failure-domain.beta.kubernetes.io/region": "region",
		"failure-domain.beta.kubernetes.io/zone":   "zone",
		"topology.kubernetes.io/region":            "region",
		"topology.kubernetes.io/zone":              "zone",
		"node.kubernetes.io/windows-build":         "windows-build",
		"node-role.kubernetes.io/control-plane":    "control-plane",
	}
	wantFiltered := map[string]string{
		"key": "value",
	}
	assert := assert.New(t)
	assert.Equal(wantFiltered, FilterLabels(labels))
}

var pendingNodes = []updatev1alpha1.PendingNode{
	{
		Spec: updatev1alpha1.PendingNodeSpec{
			NodeName: "joiningnode",
			Goal:     updatev1alpha1.NodeGoalJoin,
		},
		Status: updatev1alpha1.PendingNodeStatus{
			CSPNodeState: updatev1alpha1.NodeStateReady,
		},
	},
	{
		Spec: updatev1alpha1.PendingNodeSpec{
			NodeName: "unreadynode",
			Goal:     updatev1alpha1.NodeGoalJoin,
		},
		Status: updatev1alpha1.PendingNodeStatus{
			CSPNodeState: updatev1alpha1.NodeStateCreating,
		},
	},
	{
		Spec: updatev1alpha1.PendingNodeSpec{
			NodeName: "leavingnode",
			Goal:     updatev1alpha1.NodeGoalLeave,
		},
	},
}
