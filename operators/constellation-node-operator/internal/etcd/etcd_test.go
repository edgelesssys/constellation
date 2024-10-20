/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package etcd

import (
	"context"
	"errors"
	"testing"

	"github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/internal/constants"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	pb "go.etcd.io/etcd/api/v3/etcdserverpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestClose(t *testing.T) {
	client := Client{etcdClient: &stubEtcdClient{}}
	assert.NoError(t, client.Close())
}

func TestRemoveEtcdMemberFromCluster(t *testing.T) {
	testCases := map[string]struct {
		vpcIP         string
		memberListErr error
		wantErr       bool
	}{
		"removing member works": {
			vpcIP: "192.0.2.1",
		},
		"member already removed": {
			vpcIP: "192.0.2.2",
		},
		"listing members fails": {
			memberListErr: errors.New("listing members failed"),
			wantErr:       true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			client := Client{etcdClient: &stubEtcdClient{
				members: []*pb.Member{
					{ID: 1, PeerURLs: []string{"https://192.0.2.1:2380"}},
				},
				listErr: tc.memberListErr,
			}}
			err := client.RemoveEtcdMemberFromCluster(context.Background(), tc.vpcIP)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
		})
	}
}

func TestGetMemberID(t *testing.T) {
	testCases := map[string]struct {
		members       []*pb.Member
		memberListErr error
		wantMemberID  uint64
		wantErr       bool
	}{
		"getting member id works": {
			members: []*pb.Member{
				{ID: 1, PeerURLs: []string{"https://192.0.2.1:2380"}},
			},
			wantMemberID: 1,
		},
		"vpc ip has no corresponding etcd member": {
			members: []*pb.Member{
				{ID: 1, PeerURLs: []string{"https://192.0.2.2:2380"}},
			},
			wantErr: true,
		},
		"listing members fails": {
			memberListErr: errors.New("listing members failed"),
			wantErr:       true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			client := Client{etcdClient: &stubEtcdClient{
				members: tc.members,
				listErr: tc.memberListErr,
			}}
			gotMemberID, err := client.getMemberID(context.Background(), "192.0.2.1")
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantMemberID, gotMemberID)
		})
	}
}

func TestPeerURL(t *testing.T) {
	assert.Equal(t, "https://host:2380", peerURL("host", etcdListenPeerPort))
}

func TestGetInitialEndpoints(t *testing.T) {
	testCases := map[string]struct {
		nodes         []corev1.Node
		listErr       error
		wantEndpoints []string
		wantErr       bool
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
			wantEndpoints: []string{"192.0.2.1:2379"},
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

			client := &stubK8sClient{
				nodes:   tc.nodes,
				listErr: tc.listErr,
			}
			gotEndpoints, err := getInitialEndpoints(client)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.ElementsMatch(tc.wantEndpoints, gotEndpoints)
		})
	}
}

type stubK8sClient struct {
	nodes   []corev1.Node
	listErr error
	client.Client
}

func (c *stubK8sClient) List(_ context.Context, list client.ObjectList, _ ...client.ListOption) error {
	list.(*corev1.NodeList).Items = c.nodes
	return c.listErr
}

type stubEtcdClient struct {
	members   []*pb.Member
	listErr   error
	removeErr error
	syncErr   error
	closeErr  error
}

func (c *stubEtcdClient) MemberList(_ context.Context) (*clientv3.MemberListResponse, error) {
	return &clientv3.MemberListResponse{
		Members: c.members,
	}, c.listErr
}

func (c *stubEtcdClient) MemberRemove(_ context.Context, _ uint64) (*clientv3.MemberRemoveResponse, error) {
	return &clientv3.MemberRemoveResponse{
		Members: c.members,
	}, c.removeErr
}

func (c *stubEtcdClient) Sync(_ context.Context) error {
	return c.syncErr
}

func (c *stubEtcdClient) Close() error {
	return c.closeErr
}
