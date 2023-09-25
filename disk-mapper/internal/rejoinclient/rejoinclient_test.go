/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package rejoinclient

import (
	"context"
	"errors"
	"net"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/cloud/metadata"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/grpc/atlscredentials"
	"github.com/edgelesssys/constellation/v2/internal/grpc/dialer"
	"github.com/edgelesssys/constellation/v2/internal/grpc/testdialer"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/role"
	"github.com/edgelesssys/constellation/v2/joinservice/joinproto"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
	"google.golang.org/grpc"
	testclock "k8s.io/utils/clock/testing"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestStartCancel(t *testing.T) {
	netDialer := testdialer.NewBufconnDialer()
	dialer := dialer.New(nil, nil, netDialer)

	clock := testclock.NewFakeClock(time.Time{})

	metaAPI := &stubMetadataAPI{
		instances: []metadata.InstanceMetadata{
			{
				Role:  role.ControlPlane,
				VPCIP: "192.0.2.1",
			},
			{
				Role:  role.ControlPlane,
				VPCIP: "192.0.2.1",
			},
		},
	}

	client := &RejoinClient{
		dialer:      dialer,
		nodeInfo:    metadata.InstanceMetadata{Role: role.Worker},
		metadataAPI: metaAPI,
		log:         logger.NewTest(t),
		timeout:     time.Second * 30,
		interval:    time.Second,
		clock:       clock,
	}

	serverCreds := atlscredentials.New(nil, nil)
	rejoinServer := grpc.NewServer(grpc.Creds(serverCreds))
	rejoinServiceAPI := &stubRejoinServiceAPI{err: errors.New("error")}
	joinproto.RegisterAPIServer(rejoinServer, rejoinServiceAPI)
	port := strconv.Itoa(constants.JoinServiceNodePort)
	listener := netDialer.GetListener(net.JoinHostPort("192.0.2.1", port))
	go rejoinServer.Serve(listener)
	defer rejoinServer.GracefulStop()

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		client.Start(ctx, "uuid")
	}()

	clock.Step(time.Millisecond)
	cancel()
	wg.Wait()
	assert.Equal(t, client.diskUUID, "uuid")
}

func TestRemoveSelfFromEndpoints(t *testing.T) {
	testCases := map[string]struct {
		self      string
		endpoints []string
	}{
		"self is not in endpoints": {
			self: "192.0.2.1",
			endpoints: []string{
				"192.0.2.2:30090",
				"192.0.2.3:30090",
				"192.0.2.4:30090",
				"192.0.2.5:30090",
				"192.0.2.6:30090",
			},
		},
		"self is in endpoints": {
			self: "192.0.2.1",
			endpoints: []string{
				"192.0.2.2:30090",
				"192.0.2.3:30090",
				"192.0.2.4:30090",
				"192.0.2.5:30090",
				"192.0.2.6:30090",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			got := removeSelfFromEndpoints(tc.self, tc.endpoints)
			assert.NotContains(got, tc.self)
		})
	}
}

func TestGetJoinEndpoints(t *testing.T) {
	testInstances := []metadata.InstanceMetadata{
		{
			Role:  role.ControlPlane,
			VPCIP: "192.0.2.2",
		},
		{
			Role:  role.ControlPlane,
			VPCIP: "192.0.2.3",
		},
		{
			Role:  role.ControlPlane,
			VPCIP: "192.0.2.4",
		},
		{
			Role:  role.Worker,
			VPCIP: "192.0.2.12",
		},
		{
			Role:  role.Worker,
			VPCIP: "192.0.2.13",
		},
		{
			Role:  role.Worker,
			VPCIP: "192.0.2.14",
		},
	}

	testCases := map[string]struct {
		nodeInfo      metadata.InstanceMetadata
		meta          stubMetadataAPI
		wantEndpoints int
		wantErr       bool
	}{
		"worker node": {
			nodeInfo: metadata.InstanceMetadata{
				Role:  role.Worker,
				VPCIP: "192.0.2.1",
			},
			meta: stubMetadataAPI{
				instances:  testInstances,
				lbEndpoint: "192.0.2.100",
			},
			wantEndpoints: 4,
		},
		"control-plane node not in list": {
			nodeInfo: metadata.InstanceMetadata{
				Role:  role.ControlPlane,
				VPCIP: "192.0.2.1",
			},
			meta: stubMetadataAPI{
				instances:  testInstances,
				lbEndpoint: "192.0.2.100",
			},
			wantEndpoints: 4,
		},
		"control-plane node in list": {
			nodeInfo: metadata.InstanceMetadata{
				Role:  role.ControlPlane,
				VPCIP: "192.0.2.2",
			},
			meta: stubMetadataAPI{
				instances:  testInstances,
				lbEndpoint: "192.0.2.100",
			},
			wantEndpoints: 3,
		},
		"metadata list error": {
			nodeInfo: metadata.InstanceMetadata{
				Role:  role.ControlPlane,
				VPCIP: "192.0.2.1",
			},
			meta: stubMetadataAPI{
				listErr: assert.AnError,
			},
			wantErr: true,
		},
		"metadata load balancer error": {
			nodeInfo: metadata.InstanceMetadata{
				Role:  role.ControlPlane,
				VPCIP: "192.0.2.1",
			},
			meta: stubMetadataAPI{
				getLoadBalancerEndpointErr: assert.AnError,
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			client := New(nil, tc.nodeInfo, tc.meta, logger.NewTest(t))

			endpoints, err := client.getJoinEndpoints()
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.NotContains(endpoints, tc.nodeInfo.VPCIP)
				// +1 for the load balancer endpoint
				assert.Len(endpoints, tc.wantEndpoints)
			}
		})
	}
}

func TestStart(t *testing.T) {
	testCases := map[string]struct {
		nodeInfo metadata.InstanceMetadata
	}{
		"worker node": {
			nodeInfo: metadata.InstanceMetadata{
				Role:  role.Worker,
				VPCIP: "192.0.2.99",
			},
		},
		"control-plane node": {
			nodeInfo: metadata.InstanceMetadata{
				Role:  role.ControlPlane,
				VPCIP: "192.0.2.99",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			diskKey := []byte("disk-key")
			measurementSecret := []byte("measurement-secret")
			netDialer := testdialer.NewBufconnDialer()
			dialer := dialer.New(nil, nil, netDialer)
			serverCreds := atlscredentials.New(nil, nil)
			rejoinServer := grpc.NewServer(grpc.Creds(serverCreds))
			rejoinServiceAPI := &stubRejoinServiceAPI{
				rejoinTicketResponse: &joinproto.IssueRejoinTicketResponse{
					StateDiskKey:      diskKey,
					MeasurementSecret: measurementSecret,
				},
			}
			joinproto.RegisterAPIServer(rejoinServer, rejoinServiceAPI)
			port := strconv.Itoa(constants.JoinServiceNodePort)
			listener := netDialer.GetListener(net.JoinHostPort("192.0.2.1", port))
			go rejoinServer.Serve(listener)
			defer rejoinServer.GracefulStop()

			meta := stubMetadataAPI{
				instances: []metadata.InstanceMetadata{
					{
						Role:  role.ControlPlane,
						VPCIP: "192.0.2.1",
					},
					{
						Role:  role.ControlPlane,
						VPCIP: "192.0.2.2",
					},
					{
						Role:  role.Worker,
						VPCIP: "192.0.2.13",
					},
					{
						Role:  role.Worker,
						VPCIP: "192.0.2.14",
					},
				},
			}

			client := New(dialer, tc.nodeInfo, meta, logger.NewTest(t))

			passphrase, secret := client.Start(context.Background(), "uuid")
			assert.Equal(diskKey, passphrase)
			assert.Equal(measurementSecret, secret)
		})
	}
}

type stubMetadataAPI struct {
	instances                  []metadata.InstanceMetadata
	lbEndpoint                 string
	getLoadBalancerEndpointErr error
	listErr                    error
}

func (s stubMetadataAPI) List(context.Context) ([]metadata.InstanceMetadata, error) {
	return s.instances, s.listErr
}

func (s stubMetadataAPI) GetLoadBalancerEndpoint(_ context.Context) (string, string, error) {
	return s.lbEndpoint, "", s.getLoadBalancerEndpointErr
}

type stubRejoinServiceAPI struct {
	rejoinTicketResponse *joinproto.IssueRejoinTicketResponse
	err                  error
	joinproto.UnimplementedAPIServer
}

func (s *stubRejoinServiceAPI) IssueRejoinTicket(context.Context, *joinproto.IssueRejoinTicketRequest,
) (*joinproto.IssueRejoinTicketResponse, error) {
	return s.rejoinTicketResponse, s.err
}
