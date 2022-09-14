/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package client

import (
	"context"
	"errors"
	"fmt"
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
		addrAPI        addressesAPI
		healthAPI      healthChecksAPI
		backendAPI     backendServicesAPI
		proxyAPI       targetTCPProxiesAPI
		forwardAPI     forwardingRulesAPI
		operationAPI   operationGlobalAPI
		isDebugCluster bool
		wantErr        bool
	}{
		"successful create": {
			addrAPI:      &stubAddressesAPI{getAddr: proto.String("192.0.2.1")},
			healthAPI:    &stubHealthChecksAPI{},
			backendAPI:   &stubBackendServicesAPI{},
			proxyAPI:     &stubTargetTCPProxiesAPI{},
			forwardAPI:   &stubForwardingRulesAPI{forwardingRule: forwardingRule},
			operationAPI: stubOperationGlobalAPI{},
		},
		"successful create (debug cluster)": {
			addrAPI:        &stubAddressesAPI{getAddr: proto.String("192.0.2.1")},
			healthAPI:      &stubHealthChecksAPI{},
			backendAPI:     &stubBackendServicesAPI{},
			proxyAPI:       &stubTargetTCPProxiesAPI{},
			forwardAPI:     &stubForwardingRulesAPI{forwardingRule: forwardingRule},
			operationAPI:   stubOperationGlobalAPI{},
			isDebugCluster: true,
		},
		"createIPAddr fails": {
			addrAPI:      &stubAddressesAPI{insertErr: someErr},
			healthAPI:    &stubHealthChecksAPI{},
			backendAPI:   &stubBackendServicesAPI{},
			proxyAPI:     &stubTargetTCPProxiesAPI{},
			forwardAPI:   &stubForwardingRulesAPI{forwardingRule: forwardingRule},
			operationAPI: stubOperationGlobalAPI{},
			wantErr:      true,
		},
		"createLB fails": {
			addrAPI:      &stubAddressesAPI{},
			healthAPI:    &stubHealthChecksAPI{},
			backendAPI:   &stubBackendServicesAPI{insertErr: someErr},
			proxyAPI:     &stubTargetTCPProxiesAPI{},
			forwardAPI:   &stubForwardingRulesAPI{forwardingRule: forwardingRule},
			operationAPI: stubOperationGlobalAPI{},
			wantErr:      true,
		},
		"createTcpProxy fails": {
			addrAPI:      &stubAddressesAPI{getAddr: proto.String("192.0.2.1")},
			healthAPI:    &stubHealthChecksAPI{},
			backendAPI:   &stubBackendServicesAPI{},
			proxyAPI:     &stubTargetTCPProxiesAPI{insertErr: someErr},
			forwardAPI:   &stubForwardingRulesAPI{forwardingRule: forwardingRule},
			operationAPI: stubOperationGlobalAPI{},
			wantErr:      true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			ctx := context.Background()
			client := Client{
				project:             "project",
				zone:                "zone",
				name:                "name",
				uid:                 "uid",
				addressesAPI:        tc.addrAPI,
				targetTCPProxiesAPI: tc.proxyAPI,
				healthChecksAPI:     tc.healthAPI,
				backendServicesAPI:  tc.backendAPI,
				forwardingRulesAPI:  tc.forwardAPI,
				operationGlobalAPI:  tc.operationAPI,
			}

			err := client.CreateLoadBalancers(ctx, tc.isDebugCluster)

			// In case we expect an error, check for the error and continue otherwise.
			if tc.wantErr {
				assert.Error(err)
				return
			}

			// If we don't expect an error, check if the resources have been successfully created.
			assert.NoError(err)
			assert.NotEmpty(client.loadbalancerIPname)

			var foundDebugdLB bool
			for _, lb := range client.loadbalancers {
				// Expect load balancer name to have the format of "name-serviceName-uid" which is what buildResourceName does currently.
				if lb.name == fmt.Sprintf("%s-debugd-%s", client.name, client.uid) {
					foundDebugdLB = true
					break
				}
			}

			if tc.isDebugCluster {
				assert.Equal(6, len(client.loadbalancers))
				assert.True(foundDebugdLB, "debugd loadbalancer not found in debug-mode")
			} else {
				assert.Equal(5, len(client.loadbalancers))
				assert.False(foundDebugdLB, "debugd loadbalancer found in non-debug mode")
			}
		})
	}
}

func TestCreateLoadBalancer(t *testing.T) {
	someErr := errors.New("failed")
	testCases := map[string]struct {
		operationGlobalAPI  operationGlobalAPI
		healthChecksAPI     healthChecksAPI
		backendServicesAPI  backendServicesAPI
		forwardingRulesAPI  forwardingRulesAPI
		targetTCPProxiesAPI targetTCPProxiesAPI
		wantErr             bool
		wantLB              *loadBalancer
	}{
		"successful create": {
			healthChecksAPI:     stubHealthChecksAPI{},
			backendServicesAPI:  stubBackendServicesAPI{},
			targetTCPProxiesAPI: stubTargetTCPProxiesAPI{},
			forwardingRulesAPI:  stubForwardingRulesAPI{forwardingRule: &compute.ForwardingRule{LabelFingerprint: proto.String("fingerprint")}},
			operationGlobalAPI:  stubOperationGlobalAPI{},
			wantLB: &loadBalancer{
				name:               "name",
				frontendPort:       1234,
				backendPortName:    "testport",
				hasHealthCheck:     true,
				hasTargetTCPProxy:  true,
				hasBackendService:  true,
				hasForwardingRules: true,
			},
		},
		"successful create with label": {
			healthChecksAPI:     stubHealthChecksAPI{},
			backendServicesAPI:  stubBackendServicesAPI{},
			targetTCPProxiesAPI: stubTargetTCPProxiesAPI{},
			forwardingRulesAPI:  stubForwardingRulesAPI{forwardingRule: &compute.ForwardingRule{LabelFingerprint: proto.String("fingerprint")}},
			operationGlobalAPI:  stubOperationGlobalAPI{},
			wantLB: &loadBalancer{
				name:               "name",
				frontendPort:       1234,
				backendPortName:    "testport",
				label:              true,
				hasHealthCheck:     true,
				hasTargetTCPProxy:  true,
				hasBackendService:  true,
				hasForwardingRules: true,
			},
		},
		"CreateLoadBalancer fails when getting forwarding rule": {
			healthChecksAPI:     stubHealthChecksAPI{},
			backendServicesAPI:  stubBackendServicesAPI{},
			targetTCPProxiesAPI: stubTargetTCPProxiesAPI{},
			forwardingRulesAPI:  stubForwardingRulesAPI{getErr: someErr},
			operationGlobalAPI:  stubOperationGlobalAPI{},
			wantErr:             true,
			wantLB: &loadBalancer{
				name:               "name",
				frontendPort:       1234,
				backendPortName:    "testport",
				label:              true,
				hasHealthCheck:     true,
				hasTargetTCPProxy:  true,
				hasBackendService:  true,
				hasForwardingRules: true,
			},
		},
		"CreateLoadBalancer fails when label fingerprint is missing": {
			healthChecksAPI:     stubHealthChecksAPI{},
			backendServicesAPI:  stubBackendServicesAPI{},
			targetTCPProxiesAPI: stubTargetTCPProxiesAPI{},
			forwardingRulesAPI:  stubForwardingRulesAPI{forwardingRule: &compute.ForwardingRule{}},
			operationGlobalAPI:  stubOperationGlobalAPI{},
			wantErr:             true,
			wantLB: &loadBalancer{
				name:               "name",
				frontendPort:       1234,
				backendPortName:    "testport",
				label:              true,
				hasHealthCheck:     true,
				hasTargetTCPProxy:  true,
				hasBackendService:  true,
				hasForwardingRules: true,
			},
		},
		"CreateLoadBalancer fails when creating health check": {
			healthChecksAPI:     stubHealthChecksAPI{insertErr: someErr},
			backendServicesAPI:  stubBackendServicesAPI{},
			targetTCPProxiesAPI: stubTargetTCPProxiesAPI{},
			forwardingRulesAPI:  stubForwardingRulesAPI{forwardingRule: &compute.ForwardingRule{LabelFingerprint: proto.String("fingerprint")}},
			operationGlobalAPI:  stubOperationGlobalAPI{},
			wantErr:             true,
			wantLB: &loadBalancer{
				name:               "name",
				frontendPort:       1234,
				backendPortName:    "testport",
				hasHealthCheck:     false,
				hasTargetTCPProxy:  false,
				hasBackendService:  false,
				hasForwardingRules: false,
			},
		},
		"CreateLoadBalancer fails when creating backend service": {
			healthChecksAPI:     stubHealthChecksAPI{},
			targetTCPProxiesAPI: stubTargetTCPProxiesAPI{},
			backendServicesAPI:  stubBackendServicesAPI{insertErr: someErr},
			forwardingRulesAPI:  stubForwardingRulesAPI{},
			operationGlobalAPI:  stubOperationGlobalAPI{},
			wantErr:             true,
			wantLB: &loadBalancer{
				name:               "name",
				frontendPort:       1234,
				backendPortName:    "testport",
				hasHealthCheck:     true,
				hasBackendService:  false,
				hasTargetTCPProxy:  false,
				hasForwardingRules: false,
			},
		},
		"CreateLoadBalancer fails when creating forwarding rule": {
			healthChecksAPI:     stubHealthChecksAPI{},
			backendServicesAPI:  stubBackendServicesAPI{},
			targetTCPProxiesAPI: stubTargetTCPProxiesAPI{},
			forwardingRulesAPI:  stubForwardingRulesAPI{insertErr: someErr},
			operationGlobalAPI:  stubOperationGlobalAPI{},
			wantErr:             true,
			wantLB: &loadBalancer{
				name:               "name",
				frontendPort:       1234,
				backendPortName:    "testport",
				hasHealthCheck:     true,
				hasBackendService:  true,
				hasTargetTCPProxy:  true,
				hasForwardingRules: false,
			},
		},
		"CreateLoadBalancer fails when creating target proxy rule": {
			healthChecksAPI:     stubHealthChecksAPI{},
			backendServicesAPI:  stubBackendServicesAPI{},
			targetTCPProxiesAPI: stubTargetTCPProxiesAPI{insertErr: someErr},
			forwardingRulesAPI:  stubForwardingRulesAPI{},
			operationGlobalAPI:  stubOperationGlobalAPI{},
			wantErr:             true,
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
			healthChecksAPI:     stubHealthChecksAPI{},
			backendServicesAPI:  stubBackendServicesAPI{},
			targetTCPProxiesAPI: stubTargetTCPProxiesAPI{},
			forwardingRulesAPI:  stubForwardingRulesAPI{forwardingRule: &compute.ForwardingRule{LabelFingerprint: proto.String("fingerprint")}},
			operationGlobalAPI:  stubOperationGlobalAPI{waitErr: someErr},
			wantErr:             true,
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
				project:             "project",
				zone:                "zone",
				name:                "name",
				uid:                 "uid",
				backendServicesAPI:  tc.backendServicesAPI,
				forwardingRulesAPI:  tc.forwardingRulesAPI,
				targetTCPProxiesAPI: tc.targetTCPProxiesAPI,
				healthChecksAPI:     tc.healthChecksAPI,
				operationGlobalAPI:  tc.operationGlobalAPI,
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
			hasTargetTCPProxy:  true,
			hasForwardingRules: true,
		}
	}

	testCases := map[string]struct {
		addrAPI     addressesAPI
		healthAPI   healthChecksAPI
		backendAPI  backendServicesAPI
		targetAPI   targetTCPProxiesAPI
		forwardAPI  forwardingRulesAPI
		opGlobalAPI operationGlobalAPI
		wantErr     bool
	}{
		"successful terminate": {
			addrAPI:     &stubAddressesAPI{},
			healthAPI:   &stubHealthChecksAPI{},
			backendAPI:  &stubBackendServicesAPI{},
			targetAPI:   &stubTargetTCPProxiesAPI{},
			forwardAPI:  &stubForwardingRulesAPI{},
			opGlobalAPI: stubOperationGlobalAPI{},
		},
		"deleteIPAddr fails": {
			addrAPI:     &stubAddressesAPI{deleteErr: someErr},
			healthAPI:   &stubHealthChecksAPI{},
			backendAPI:  &stubBackendServicesAPI{},
			targetAPI:   &stubTargetTCPProxiesAPI{},
			forwardAPI:  &stubForwardingRulesAPI{},
			opGlobalAPI: stubOperationGlobalAPI{},
			wantErr:     true,
		},
		"deleteLB fails": {
			addrAPI:     &stubAddressesAPI{},
			healthAPI:   &stubHealthChecksAPI{},
			backendAPI:  &stubBackendServicesAPI{deleteErr: someErr},
			targetAPI:   &stubTargetTCPProxiesAPI{},
			forwardAPI:  &stubForwardingRulesAPI{},
			opGlobalAPI: stubOperationGlobalAPI{},
			wantErr:     true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			ctx := context.Background()
			client := Client{
				project:             "project",
				zone:                "zone",
				name:                "name",
				uid:                 "uid",
				addressesAPI:        tc.addrAPI,
				healthChecksAPI:     tc.healthAPI,
				backendServicesAPI:  tc.backendAPI,
				targetTCPProxiesAPI: tc.targetAPI,
				forwardingRulesAPI:  tc.forwardAPI,
				operationGlobalAPI:  tc.opGlobalAPI,
				loadbalancerIPname:  "loadbalancerIPid",
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
			hasTargetTCPProxy:  true,
			hasBackendService:  true,
			hasForwardingRules: true,
		}
	}

	testCases := map[string]struct {
		lb                  *loadBalancer
		opGlobalAPI         operationGlobalAPI
		healthChecksAPI     healthChecksAPI
		backendServicesAPI  backendServicesAPI
		targetTCPProxiesAPI targetTCPProxiesAPI
		forwardingRulesAPI  forwardingRulesAPI
		wantErr             bool
		wantLB              *loadBalancer
	}{
		"successful terminate": {
			lb:                  newRunningLB(),
			healthChecksAPI:     stubHealthChecksAPI{},
			backendServicesAPI:  stubBackendServicesAPI{},
			forwardingRulesAPI:  stubForwardingRulesAPI{},
			targetTCPProxiesAPI: stubTargetTCPProxiesAPI{},
			opGlobalAPI:         stubOperationGlobalAPI{},
			wantLB:              &loadBalancer{},
		},
		"terminate partially created loadbalancer": {
			lb: &loadBalancer{
				name:               "name",
				hasHealthCheck:     true,
				hasBackendService:  true,
				hasForwardingRules: false,
			},
			healthChecksAPI:     stubHealthChecksAPI{},
			backendServicesAPI:  stubBackendServicesAPI{},
			forwardingRulesAPI:  stubForwardingRulesAPI{deleteErr: someErr},
			targetTCPProxiesAPI: stubTargetTCPProxiesAPI{},
			opGlobalAPI:         stubOperationGlobalAPI{},
			wantLB:              &loadBalancer{},
		},
		"terminate partially created loadbalancer 2": {
			lb: &loadBalancer{
				name:               "name",
				hasHealthCheck:     true,
				hasBackendService:  false,
				hasForwardingRules: false,
			},
			healthChecksAPI:     stubHealthChecksAPI{},
			backendServicesAPI:  stubBackendServicesAPI{deleteErr: someErr},
			forwardingRulesAPI:  stubForwardingRulesAPI{deleteErr: someErr},
			targetTCPProxiesAPI: stubTargetTCPProxiesAPI{},
			opGlobalAPI:         stubOperationGlobalAPI{},
			wantLB:              &loadBalancer{},
		},
		"no-op for nil loadbalancer": {
			lb: nil,
		},
		"health check not found": {
			lb:                  newRunningLB(),
			healthChecksAPI:     stubHealthChecksAPI{deleteErr: notFoundErr},
			backendServicesAPI:  stubBackendServicesAPI{},
			forwardingRulesAPI:  stubForwardingRulesAPI{},
			targetTCPProxiesAPI: stubTargetTCPProxiesAPI{},
			opGlobalAPI:         stubOperationGlobalAPI{},
			wantLB:              &loadBalancer{},
		},
		"backend service not found": {
			lb:                  newRunningLB(),
			healthChecksAPI:     stubHealthChecksAPI{},
			backendServicesAPI:  stubBackendServicesAPI{deleteErr: notFoundErr},
			forwardingRulesAPI:  stubForwardingRulesAPI{},
			targetTCPProxiesAPI: stubTargetTCPProxiesAPI{},
			opGlobalAPI:         stubOperationGlobalAPI{},
			wantLB:              &loadBalancer{},
		},
		"forwarding rules not found": {
			lb:                  newRunningLB(),
			healthChecksAPI:     stubHealthChecksAPI{},
			backendServicesAPI:  stubBackendServicesAPI{},
			forwardingRulesAPI:  stubForwardingRulesAPI{deleteErr: notFoundErr},
			targetTCPProxiesAPI: stubTargetTCPProxiesAPI{},
			opGlobalAPI:         stubOperationGlobalAPI{},
			wantLB:              &loadBalancer{},
		},
		"fails for loadbalancer without name": {
			lb:      &loadBalancer{},
			wantErr: true,
			wantLB:  &loadBalancer{},
		},
		"fails when deleting health check": {
			lb:                  newRunningLB(),
			healthChecksAPI:     stubHealthChecksAPI{deleteErr: someErr},
			backendServicesAPI:  stubBackendServicesAPI{},
			forwardingRulesAPI:  stubForwardingRulesAPI{},
			targetTCPProxiesAPI: stubTargetTCPProxiesAPI{},
			opGlobalAPI:         stubOperationGlobalAPI{},
			wantErr:             true,
			wantLB: &loadBalancer{
				name:               "name",
				hasHealthCheck:     true,
				hasBackendService:  false,
				hasForwardingRules: false,
				hasTargetTCPProxy:  false,
			},
		},
		"fails when deleting backend service": {
			lb:                  newRunningLB(),
			healthChecksAPI:     stubHealthChecksAPI{},
			backendServicesAPI:  stubBackendServicesAPI{deleteErr: someErr},
			forwardingRulesAPI:  stubForwardingRulesAPI{},
			targetTCPProxiesAPI: stubTargetTCPProxiesAPI{},
			opGlobalAPI:         stubOperationGlobalAPI{},
			wantErr:             true,
			wantLB: &loadBalancer{
				name:               "name",
				hasHealthCheck:     true,
				hasBackendService:  true,
				hasForwardingRules: false,
				hasTargetTCPProxy:  false,
			},
		},
		"fails when deleting forwarding rule": {
			lb:                  newRunningLB(),
			healthChecksAPI:     stubHealthChecksAPI{},
			backendServicesAPI:  stubBackendServicesAPI{},
			forwardingRulesAPI:  stubForwardingRulesAPI{deleteErr: someErr},
			targetTCPProxiesAPI: stubTargetTCPProxiesAPI{},
			opGlobalAPI:         stubOperationGlobalAPI{},
			wantErr:             true,
			wantLB: &loadBalancer{
				name:               "name",
				hasHealthCheck:     true,
				hasBackendService:  true,
				hasForwardingRules: true,
				hasTargetTCPProxy:  true,
			},
		},
		"fails when deleting tcp proxy rule": {
			lb:                  newRunningLB(),
			healthChecksAPI:     stubHealthChecksAPI{},
			backendServicesAPI:  stubBackendServicesAPI{},
			forwardingRulesAPI:  stubForwardingRulesAPI{},
			targetTCPProxiesAPI: stubTargetTCPProxiesAPI{deleteErr: someErr},
			opGlobalAPI:         stubOperationGlobalAPI{},
			wantErr:             true,
			wantLB: &loadBalancer{
				name:               "name",
				hasHealthCheck:     true,
				hasBackendService:  true,
				hasForwardingRules: false,
				hasTargetTCPProxy:  true,
			},
		},
		"fails when waiting on operation": {
			lb:                  newRunningLB(),
			healthChecksAPI:     stubHealthChecksAPI{},
			backendServicesAPI:  stubBackendServicesAPI{},
			forwardingRulesAPI:  stubForwardingRulesAPI{},
			targetTCPProxiesAPI: stubTargetTCPProxiesAPI{},
			opGlobalAPI:         stubOperationGlobalAPI{waitErr: someErr},
			wantErr:             true,
			wantLB: &loadBalancer{
				name:               "name",
				hasHealthCheck:     true,
				hasBackendService:  true,
				hasForwardingRules: true,
				hasTargetTCPProxy:  true,
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			ctx := context.Background()
			client := Client{
				project:             "project",
				zone:                "zone",
				name:                "name",
				uid:                 "uid",
				backendServicesAPI:  tc.backendServicesAPI,
				forwardingRulesAPI:  tc.forwardingRulesAPI,
				healthChecksAPI:     tc.healthChecksAPI,
				targetTCPProxiesAPI: tc.targetTCPProxiesAPI,
				operationGlobalAPI:  tc.opGlobalAPI,
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
		opAPI   operationGlobalAPI
		wantErr bool
	}{
		"successful create": {
			addrAPI: stubAddressesAPI{getAddr: proto.String("test-ip")},
			opAPI:   stubOperationGlobalAPI{},
		},
		"insert fails": {
			addrAPI: stubAddressesAPI{insertErr: someErr},
			opAPI:   stubOperationGlobalAPI{},
			wantErr: true,
		},
		"get fails": {
			addrAPI: stubAddressesAPI{getErr: someErr},
			opAPI:   stubOperationGlobalAPI{},
			wantErr: true,
		},
		"get address nil": {
			addrAPI: stubAddressesAPI{getAddr: nil},
			opAPI:   stubOperationGlobalAPI{},
			wantErr: true,
		},
		"wait fails": {
			addrAPI: stubAddressesAPI{},
			opAPI:   stubOperationGlobalAPI{waitErr: someErr},
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
				operationGlobalAPI: tc.opAPI,
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
		opAPI   operationGlobalAPI
		addrID  string
		wantErr bool
	}{
		"successful delete": {
			addrAPI: stubAddressesAPI{},
			opAPI:   stubOperationGlobalAPI{},
			addrID:  "name",
		},
		"not found": {
			addrAPI: stubAddressesAPI{deleteErr: notFoundErr},
			opAPI:   stubOperationGlobalAPI{},
			addrID:  "name",
		},
		"empty is no-op": {
			addrAPI: stubAddressesAPI{},
			opAPI:   stubOperationGlobalAPI{},
		},
		"delete fails": {
			addrAPI: stubAddressesAPI{deleteErr: someErr},
			opAPI:   stubOperationGlobalAPI{},
			addrID:  "name",
			wantErr: true,
		},
		"wait fails": {
			addrAPI: stubAddressesAPI{},
			opAPI:   stubOperationGlobalAPI{waitErr: someErr},
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
				operationGlobalAPI: tc.opAPI,
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
