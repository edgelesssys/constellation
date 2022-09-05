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

	"github.com/stretchr/testify/assert"
	"google.golang.org/api/googleapi"
	"google.golang.org/genproto/googleapis/cloud/compute/v1"
	"google.golang.org/protobuf/proto"
)

func TestCreateLoadBalancers(t *testing.T) {
	someErr := errors.New("failed")
	forwardingRule := &compute.ForwardingRule{LabelFingerprint: proto.String("fingerprint")}

	testCases := map[string]struct {
		addrAPI    addressesAPI
		healthAPI  healthChecksAPI
		backendAPI backendServicesAPI
		forwardAPI forwardingRulesAPI
		opRegAPI   operationRegionAPI
		wantErr    bool
	}{
		"successful create": {
			addrAPI:    &stubAddressesAPI{getAddr: proto.String("192.0.2.1")},
			healthAPI:  &stubHealthChecksAPI{},
			backendAPI: &stubBackendServicesAPI{},
			forwardAPI: &stubForwardingRulesAPI{forwardingRule: forwardingRule},
			opRegAPI:   stubOperationRegionAPI{},
		},
		"createIPAddr fails": {
			addrAPI:    &stubAddressesAPI{insertErr: someErr},
			healthAPI:  &stubHealthChecksAPI{},
			backendAPI: &stubBackendServicesAPI{},
			forwardAPI: &stubForwardingRulesAPI{forwardingRule: forwardingRule},
			opRegAPI:   stubOperationRegionAPI{},
			wantErr:    true,
		},
		"createLB fails": {
			addrAPI:    &stubAddressesAPI{},
			healthAPI:  &stubHealthChecksAPI{},
			backendAPI: &stubBackendServicesAPI{insertErr: someErr},
			forwardAPI: &stubForwardingRulesAPI{forwardingRule: forwardingRule},
			opRegAPI:   stubOperationRegionAPI{},
			wantErr:    true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			ctx := context.Background()
			client := Client{
				project:            "project",
				zone:               "zone",
				name:               "name",
				uid:                "uid",
				addressesAPI:       tc.addrAPI,
				healthChecksAPI:    tc.healthAPI,
				backendServicesAPI: tc.backendAPI,
				forwardingRulesAPI: tc.forwardAPI,
				operationRegionAPI: tc.opRegAPI,
			}

			err := client.CreateLoadBalancers(ctx)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.NotEmpty(client.loadbalancerIPname)
				assert.Equal(4, len(client.loadbalancers))
			}
		})
	}
}

func TestCreateLoadBalancer(t *testing.T) {
	someErr := errors.New("failed")
	testCases := map[string]struct {
		operationRegionAPI operationRegionAPI
		healthChecksAPI    healthChecksAPI
		backendServicesAPI backendServicesAPI
		forwardingRulesAPI forwardingRulesAPI
		wantErr            bool
		wantLB             *loadBalancer
	}{
		"successful create": {
			healthChecksAPI:    stubHealthChecksAPI{},
			backendServicesAPI: stubBackendServicesAPI{},
			forwardingRulesAPI: stubForwardingRulesAPI{forwardingRule: &compute.ForwardingRule{LabelFingerprint: proto.String("fingerprint")}},
			operationRegionAPI: stubOperationRegionAPI{},
			wantLB: &loadBalancer{
				name:               "name",
				frontendPort:       1234,
				backendPortName:    "testport",
				hasHealthCheck:     true,
				hasBackendService:  true,
				hasForwardingRules: true,
			},
		},
		"successful create with label": {
			healthChecksAPI:    stubHealthChecksAPI{},
			backendServicesAPI: stubBackendServicesAPI{},
			forwardingRulesAPI: stubForwardingRulesAPI{forwardingRule: &compute.ForwardingRule{LabelFingerprint: proto.String("fingerprint")}},
			operationRegionAPI: stubOperationRegionAPI{},
			wantLB: &loadBalancer{
				name:               "name",
				frontendPort:       1234,
				backendPortName:    "testport",
				label:              true,
				hasHealthCheck:     true,
				hasBackendService:  true,
				hasForwardingRules: true,
			},
		},
		"CreateLoadBalancer fails when getting forwarding rule": {
			healthChecksAPI:    stubHealthChecksAPI{},
			backendServicesAPI: stubBackendServicesAPI{},
			forwardingRulesAPI: stubForwardingRulesAPI{getErr: someErr},
			operationRegionAPI: stubOperationRegionAPI{},
			wantErr:            true,
			wantLB: &loadBalancer{
				name:               "name",
				frontendPort:       1234,
				backendPortName:    "testport",
				label:              true,
				hasHealthCheck:     true,
				hasBackendService:  true,
				hasForwardingRules: true,
			},
		},
		"CreateLoadBalancer fails when label fingerprint is missing": {
			healthChecksAPI:    stubHealthChecksAPI{},
			backendServicesAPI: stubBackendServicesAPI{},
			forwardingRulesAPI: stubForwardingRulesAPI{forwardingRule: &compute.ForwardingRule{}},
			operationRegionAPI: stubOperationRegionAPI{},
			wantErr:            true,
			wantLB: &loadBalancer{
				name:               "name",
				frontendPort:       1234,
				backendPortName:    "testport",
				label:              true,
				hasHealthCheck:     true,
				hasBackendService:  true,
				hasForwardingRules: true,
			},
		},
		"CreateLoadBalancer fails when creating health check": {
			healthChecksAPI:    stubHealthChecksAPI{insertErr: someErr},
			backendServicesAPI: stubBackendServicesAPI{},
			forwardingRulesAPI: stubForwardingRulesAPI{forwardingRule: &compute.ForwardingRule{LabelFingerprint: proto.String("fingerprint")}},
			operationRegionAPI: stubOperationRegionAPI{},
			wantErr:            true,
			wantLB: &loadBalancer{
				name:               "name",
				frontendPort:       1234,
				backendPortName:    "testport",
				hasHealthCheck:     false,
				hasBackendService:  false,
				hasForwardingRules: false,
			},
		},
		"CreateLoadBalancer fails when creating backend service": {
			healthChecksAPI:    stubHealthChecksAPI{},
			backendServicesAPI: stubBackendServicesAPI{insertErr: someErr},
			forwardingRulesAPI: stubForwardingRulesAPI{},
			operationRegionAPI: stubOperationRegionAPI{},
			wantErr:            true,
			wantLB: &loadBalancer{
				name:               "name",
				frontendPort:       1234,
				backendPortName:    "testport",
				hasHealthCheck:     true,
				hasBackendService:  false,
				hasForwardingRules: false,
			},
		},
		"CreateLoadBalancer fails when creating forwarding rule": {
			healthChecksAPI:    stubHealthChecksAPI{},
			backendServicesAPI: stubBackendServicesAPI{},
			forwardingRulesAPI: stubForwardingRulesAPI{insertErr: someErr},
			operationRegionAPI: stubOperationRegionAPI{},
			wantErr:            true,
			wantLB: &loadBalancer{
				name:               "name",
				frontendPort:       1234,
				backendPortName:    "testport",
				hasHealthCheck:     true,
				hasBackendService:  true,
				hasForwardingRules: false,
			},
		},
		"CreateLoadBalancer fails when waiting on operation": {
			healthChecksAPI:    stubHealthChecksAPI{},
			backendServicesAPI: stubBackendServicesAPI{},
			forwardingRulesAPI: stubForwardingRulesAPI{forwardingRule: &compute.ForwardingRule{LabelFingerprint: proto.String("fingerprint")}},
			operationRegionAPI: stubOperationRegionAPI{waitErr: someErr},
			wantErr:            true,
			wantLB: &loadBalancer{
				name:               "name",
				frontendPort:       1234,
				backendPortName:    "testport",
				hasHealthCheck:     false,
				hasBackendService:  false,
				hasForwardingRules: false,
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			ctx := context.Background()
			client := Client{
				project:            "project",
				zone:               "zone",
				name:               "name",
				uid:                "uid",
				backendServicesAPI: tc.backendServicesAPI,
				forwardingRulesAPI: tc.forwardingRulesAPI,
				healthChecksAPI:    tc.healthChecksAPI,
				operationRegionAPI: tc.operationRegionAPI,
			}
			lb := &loadBalancer{
				name:            tc.wantLB.name,
				frontendPort:    tc.wantLB.frontendPort,
				backendPortName: tc.wantLB.backendPortName,
				label:           tc.wantLB.label,
			}

			err := client.createLoadBalancer(ctx, lb)

			if tc.wantErr {
				assert.Error(err)
				assert.Equal(tc.wantLB, lb)
			} else {
				assert.NoError(err)
				assert.Equal(tc.wantLB, lb)
			}
		})
	}
}

func TestTerminateLoadbalancers(t *testing.T) {
	someErr := errors.New("failed")
	newRunningLB := func() *loadBalancer {
		return &loadBalancer{
			name:               "name",
			hasHealthCheck:     true,
			hasBackendService:  true,
			hasForwardingRules: true,
		}
	}

	testCases := map[string]struct {
		addrAPI     addressesAPI
		healthAPI   healthChecksAPI
		backendAPI  backendServicesAPI
		forwardAPI  forwardingRulesAPI
		opRegionAPI operationRegionAPI
		wantErr     bool
	}{
		"successful terminate": {
			addrAPI:     &stubAddressesAPI{},
			healthAPI:   &stubHealthChecksAPI{},
			backendAPI:  &stubBackendServicesAPI{},
			forwardAPI:  &stubForwardingRulesAPI{},
			opRegionAPI: stubOperationRegionAPI{},
		},
		"deleteIPAddr fails": {
			addrAPI:     &stubAddressesAPI{deleteErr: someErr},
			healthAPI:   &stubHealthChecksAPI{},
			backendAPI:  &stubBackendServicesAPI{},
			forwardAPI:  &stubForwardingRulesAPI{},
			opRegionAPI: stubOperationRegionAPI{},
			wantErr:     true,
		},
		"deleteLB fails": {
			addrAPI:     &stubAddressesAPI{},
			healthAPI:   &stubHealthChecksAPI{},
			backendAPI:  &stubBackendServicesAPI{deleteErr: someErr},
			forwardAPI:  &stubForwardingRulesAPI{},
			opRegionAPI: stubOperationRegionAPI{},
			wantErr:     true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			ctx := context.Background()
			client := Client{
				project:            "project",
				zone:               "zone",
				name:               "name",
				uid:                "uid",
				addressesAPI:       tc.addrAPI,
				healthChecksAPI:    tc.healthAPI,
				backendServicesAPI: tc.backendAPI,
				forwardingRulesAPI: tc.forwardAPI,
				operationRegionAPI: tc.opRegionAPI,
				loadbalancerIPname: "loadbalancerIPid",
				loadbalancers: []*loadBalancer{
					newRunningLB(),
					newRunningLB(),
					newRunningLB(),
				},
			}

			err := client.TerminateLoadBalancers(ctx)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Empty(client.loadbalancerIPname)
				assert.Nil(client.loadbalancers)
			}
		})
	}
}

func TestTerminateLoadBalancer(t *testing.T) {
	someErr := errors.New("failed")
	notFoundErr := &googleapi.Error{Code: http.StatusNotFound}
	newRunningLB := func() *loadBalancer {
		return &loadBalancer{
			name:               "name",
			hasHealthCheck:     true,
			hasBackendService:  true,
			hasForwardingRules: true,
		}
	}

	testCases := map[string]struct {
		lb                 *loadBalancer
		opRegionAPI        operationRegionAPI
		healthChecksAPI    healthChecksAPI
		backendServicesAPI backendServicesAPI
		forwardingRulesAPI forwardingRulesAPI
		wantErr            bool
		wantLB             *loadBalancer
	}{
		"successful terminate": {
			lb:                 newRunningLB(),
			healthChecksAPI:    stubHealthChecksAPI{},
			backendServicesAPI: stubBackendServicesAPI{},
			forwardingRulesAPI: stubForwardingRulesAPI{},
			opRegionAPI:        stubOperationRegionAPI{},
			wantLB:             &loadBalancer{},
		},
		"terminate partially created loadbalancer": {
			lb: &loadBalancer{
				name:               "name",
				hasHealthCheck:     true,
				hasBackendService:  true,
				hasForwardingRules: false,
			},
			healthChecksAPI:    stubHealthChecksAPI{},
			backendServicesAPI: stubBackendServicesAPI{},
			forwardingRulesAPI: stubForwardingRulesAPI{deleteErr: someErr},
			opRegionAPI:        stubOperationRegionAPI{},
			wantLB:             &loadBalancer{},
		},
		"terminate partially created loadbalancer 2": {
			lb: &loadBalancer{
				name:               "name",
				hasHealthCheck:     true,
				hasBackendService:  false,
				hasForwardingRules: false,
			},
			healthChecksAPI:    stubHealthChecksAPI{},
			backendServicesAPI: stubBackendServicesAPI{deleteErr: someErr},
			forwardingRulesAPI: stubForwardingRulesAPI{deleteErr: someErr},
			opRegionAPI:        stubOperationRegionAPI{},
			wantLB:             &loadBalancer{},
		},
		"no-op for nil loadbalancer": {
			lb: nil,
		},
		"health check not found": {
			lb:                 newRunningLB(),
			healthChecksAPI:    stubHealthChecksAPI{deleteErr: notFoundErr},
			backendServicesAPI: stubBackendServicesAPI{},
			forwardingRulesAPI: stubForwardingRulesAPI{},
			opRegionAPI:        stubOperationRegionAPI{},
			wantLB:             &loadBalancer{},
		},
		"backend service not found": {
			lb:                 newRunningLB(),
			healthChecksAPI:    stubHealthChecksAPI{},
			backendServicesAPI: stubBackendServicesAPI{deleteErr: notFoundErr},
			forwardingRulesAPI: stubForwardingRulesAPI{},
			opRegionAPI:        stubOperationRegionAPI{},
			wantLB:             &loadBalancer{},
		},
		"forwarding rules not found": {
			lb:                 newRunningLB(),
			healthChecksAPI:    stubHealthChecksAPI{},
			backendServicesAPI: stubBackendServicesAPI{},
			forwardingRulesAPI: stubForwardingRulesAPI{deleteErr: notFoundErr},
			opRegionAPI:        stubOperationRegionAPI{},
			wantLB:             &loadBalancer{},
		},
		"fails for loadbalancer without name": {
			lb:      &loadBalancer{},
			wantErr: true,
			wantLB:  &loadBalancer{},
		},
		"fails when deleting health check": {
			lb:                 newRunningLB(),
			healthChecksAPI:    stubHealthChecksAPI{deleteErr: someErr},
			backendServicesAPI: stubBackendServicesAPI{},
			forwardingRulesAPI: stubForwardingRulesAPI{},
			opRegionAPI:        stubOperationRegionAPI{},
			wantErr:            true,
			wantLB: &loadBalancer{
				name:               "name",
				hasHealthCheck:     true,
				hasBackendService:  false,
				hasForwardingRules: false,
			},
		},
		"fails when deleting backend service": {
			lb:                 newRunningLB(),
			healthChecksAPI:    stubHealthChecksAPI{},
			backendServicesAPI: stubBackendServicesAPI{deleteErr: someErr},
			forwardingRulesAPI: stubForwardingRulesAPI{},
			opRegionAPI:        stubOperationRegionAPI{},
			wantErr:            true,
			wantLB: &loadBalancer{
				name:               "name",
				hasHealthCheck:     true,
				hasBackendService:  true,
				hasForwardingRules: false,
			},
		},
		"fails when deleting forwarding rule": {
			lb:                 newRunningLB(),
			healthChecksAPI:    stubHealthChecksAPI{},
			backendServicesAPI: stubBackendServicesAPI{},
			forwardingRulesAPI: stubForwardingRulesAPI{deleteErr: someErr},
			opRegionAPI:        stubOperationRegionAPI{},
			wantErr:            true,
			wantLB: &loadBalancer{
				name:               "name",
				hasHealthCheck:     true,
				hasBackendService:  true,
				hasForwardingRules: true,
			},
		},
		"fails when waiting on operation": {
			lb:                 newRunningLB(),
			healthChecksAPI:    stubHealthChecksAPI{},
			backendServicesAPI: stubBackendServicesAPI{},
			forwardingRulesAPI: stubForwardingRulesAPI{},
			opRegionAPI:        stubOperationRegionAPI{waitErr: someErr},
			wantErr:            true,
			wantLB: &loadBalancer{
				name:               "name",
				hasHealthCheck:     true,
				hasBackendService:  true,
				hasForwardingRules: true,
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			ctx := context.Background()
			client := Client{
				project:            "project",
				zone:               "zone",
				name:               "name",
				uid:                "uid",
				backendServicesAPI: tc.backendServicesAPI,
				forwardingRulesAPI: tc.forwardingRulesAPI,
				healthChecksAPI:    tc.healthChecksAPI,
				operationRegionAPI: tc.opRegionAPI,
			}

			err := client.terminateLoadBalancer(ctx, tc.lb)

			if tc.wantErr {
				assert.Error(err)
				assert.Equal(tc.wantLB, tc.lb)
			} else {
				assert.NoError(err)
				assert.Equal(tc.wantLB, tc.lb)
			}
		})
	}
}

func TestCreateIPAddr(t *testing.T) {
	someErr := errors.New("failed")

	testCases := map[string]struct {
		addrAPI addressesAPI
		opAPI   operationRegionAPI
		wantErr bool
	}{
		"successful create": {
			addrAPI: stubAddressesAPI{getAddr: proto.String("test-ip")},
			opAPI:   stubOperationRegionAPI{},
		},
		"insert fails": {
			addrAPI: stubAddressesAPI{insertErr: someErr},
			opAPI:   stubOperationRegionAPI{},
			wantErr: true,
		},
		"get fails": {
			addrAPI: stubAddressesAPI{getErr: someErr},
			opAPI:   stubOperationRegionAPI{},
			wantErr: true,
		},
		"get address nil": {
			addrAPI: stubAddressesAPI{getAddr: nil},
			opAPI:   stubOperationRegionAPI{},
			wantErr: true,
		},
		"wait fails": {
			addrAPI: stubAddressesAPI{},
			opAPI:   stubOperationRegionAPI{waitErr: someErr},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			ctx := context.Background()
			client := Client{
				project:            "project",
				zone:               "zone",
				name:               "name",
				uid:                "uid",
				addressesAPI:       tc.addrAPI,
				operationRegionAPI: tc.opAPI,
			}

			err := client.createIPAddr(ctx)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal("test-ip", client.loadbalancerIP)
				assert.Equal("name-uid", client.loadbalancerIPname)
			}
		})
	}
}

func TestDeleteIPAddr(t *testing.T) {
	someErr := errors.New("failed")
	notFoundErr := &googleapi.Error{Code: http.StatusNotFound}

	testCases := map[string]struct {
		addrAPI addressesAPI
		opAPI   operationRegionAPI
		addrID  string
		wantErr bool
	}{
		"successful delete": {
			addrAPI: stubAddressesAPI{},
			opAPI:   stubOperationRegionAPI{},
			addrID:  "name",
		},
		"not found": {
			addrAPI: stubAddressesAPI{deleteErr: notFoundErr},
			opAPI:   stubOperationRegionAPI{},
			addrID:  "name",
		},
		"empty is no-op": {
			addrAPI: stubAddressesAPI{},
			opAPI:   stubOperationRegionAPI{},
		},
		"delete fails": {
			addrAPI: stubAddressesAPI{deleteErr: someErr},
			opAPI:   stubOperationRegionAPI{},
			addrID:  "name",
			wantErr: true,
		},
		"wait fails": {
			addrAPI: stubAddressesAPI{},
			opAPI:   stubOperationRegionAPI{waitErr: someErr},
			addrID:  "name",
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			ctx := context.Background()
			client := Client{
				project:            "project",
				zone:               "zone",
				name:               "name",
				uid:                "uid",
				addressesAPI:       tc.addrAPI,
				operationRegionAPI: tc.opAPI,
				loadbalancerIPname: tc.addrID,
			}

			err := client.deleteIPAddr(ctx)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Empty(client.loadbalancerIPname)
			}
		})
	}
}
