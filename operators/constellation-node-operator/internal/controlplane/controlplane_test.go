/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package controlplane

import (
	"context"
	"errors"
	"testing"

	"github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/internal/constants"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestListControlPlaneIPs(t *testing.T) {
	testCases := map[string]struct {
		nodes   []corev1.Node
		listErr error
		wantIPs []string
		wantErr bool
	}{
		"listing works": {
			nodes: []corev1.Node{
				{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{constants.ControlPlaneRoleLabel: ""},
					},
					Status: corev1.NodeStatus{Addresses: []corev1.NodeAddress{{
						Type:    corev1.NodeInternalIP,
						Address: "192.0.2.1",
					}}},
				},
				{
					Status: corev1.NodeStatus{Addresses: []corev1.NodeAddress{{
						Type:    corev1.NodeInternalIP,
						Address: "192.0.2.2",
					}}},
				},
			},
			wantIPs: []string{"192.0.2.1"},
		},
		"listing fails": {
			listErr: errors.New("listing failed"),
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			client := &stubClient{
				nodes:   tc.nodes,
				listErr: tc.listErr,
			}
			gotIPs, err := ListControlPlaneIPs(client)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.ElementsMatch(tc.wantIPs, gotIPs)
		})
	}
}

type stubClient struct {
	nodes   []corev1.Node
	listErr error
	client.Client
}

func (c *stubClient) List(_ context.Context, list client.ObjectList, _ ...client.ListOption) error {
	list.(*corev1.NodeList).Items = c.nodes
	return c.listErr
}
