/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package etcd

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"

	"github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/internal/controlplane"
	"go.etcd.io/etcd/client/pkg/v3/transport"
	clientv3 "go.etcd.io/etcd/client/v3"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// etcdListenClientPort defines the port etcd listen on for client traffic.
	etcdListenClientPort = "2379"
	// etcdListenPeerPort defines the port etcd listen on for peer traffic.
	etcdListenPeerPort = "2380"
	// etcdCACertName defines etcd's CA certificate name.
	etcdCACertName = "/etc/kubernetes/pki/etcd/ca.crt"
	// etcdPeerCertName defines etcd's peer certificate name.
	etcdPeerCertName = "/etc/kubernetes/pki/etcd/peer.crt"
	// etcdPeerKeyName defines etcd's peer key name.
	etcdPeerKeyName = "/etc/kubernetes/pki/etcd/peer.key"
)

var errMemberNotFound = errors.New("member not found")

// Client is an etcd client that can be used to remove a member from an etcd cluster.
type Client struct {
	etcdClient etcdClient
}

// New creates a new Client.
func New(k8sClient client.Client) (*Client, error) {
	initialEndpoints, err := getInitialEndpoints(k8sClient)
	if err != nil {
		return nil, err
	}
	tlsInfo := transport.TLSInfo{
		CertFile:      etcdPeerCertName,
		KeyFile:       etcdPeerKeyName,
		TrustedCAFile: etcdCACertName,
	}
	tlsConfig, err := tlsInfo.ClientConfig()
	if err != nil {
		return nil, err
	}
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints: initialEndpoints,
		TLS:       tlsConfig,
	})
	if err != nil {
		return nil, err
	}

	if err = etcdClient.Sync(context.TODO()); err != nil {
		return nil, fmt.Errorf("syncing endpoints with etcd: %w", err)
	}
	return &Client{
		etcdClient: etcdClient,
	}, nil
}

// Close shuts down the client's etcd connections.
func (c *Client) Close() error {
	return c.etcdClient.Close()
}

// RemoveEtcdMemberFromCluster removes an etcd member from the cluster.
func (c *Client) RemoveEtcdMemberFromCluster(ctx context.Context, vpcIP string) error {
	memberID, err := c.getMemberID(ctx, vpcIP)
	if err != nil {
		if err == errMemberNotFound {
			return nil
		}
		return err
	}
	_, err = c.etcdClient.MemberRemove(ctx, memberID)
	return err
}

// getMemberID returns the member ID of the member with the given vpcIP.
func (c *Client) getMemberID(ctx context.Context, vpcIP string) (uint64, error) {
	listResponse, err := c.etcdClient.MemberList(ctx)
	if err != nil {
		return 0, err
	}
	wantedPeerURL := peerURL(vpcIP, etcdListenPeerPort)
	for _, member := range listResponse.Members {
		for _, peerURL := range member.PeerURLs {
			if peerURL == wantedPeerURL {
				return member.ID, nil
			}
		}
	}
	return 0, errMemberNotFound
}

// peerURL returns the peer etcd URL for the given vpcIP and port.
func peerURL(host, port string) string {
	return (&url.URL{
		Scheme: "https",
		Host:   net.JoinHostPort(host, port),
	}).String()
}

// getInitialEndpoints returns the initial endpoints for the etcd cluster.
func getInitialEndpoints(k8sClient client.Client) ([]string, error) {
	ips, err := controlplane.ListControlPlaneIPs(k8sClient)
	if err != nil {
		return nil, err
	}
	etcdEndpoints := make([]string, len(ips))
	for i, ip := range ips {
		etcdEndpoints[i] = net.JoinHostPort(ip, etcdListenClientPort)
	}
	return etcdEndpoints, nil
}

type etcdClient interface {
	MemberList(ctx context.Context) (*clientv3.MemberListResponse, error)
	MemberRemove(ctx context.Context, memberID uint64) (*clientv3.MemberRemoveResponse, error)
	Sync(ctx context.Context) error
	Close() error
}
