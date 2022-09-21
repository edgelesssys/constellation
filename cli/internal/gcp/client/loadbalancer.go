/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package client

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/edgelesssys/constellation/v2/internal/constants"
	"go.uber.org/multierr"
	computepb "google.golang.org/genproto/googleapis/cloud/compute/v1"
	"google.golang.org/protobuf/proto"
)

type loadBalancer struct {
	name string

	// For creation.
	ip              string
	frontendPort    int
	backendPortName string
	healthCheck     computepb.HealthCheck_Type
	label           bool

	// For resource management.
	hasHealthCheck     bool
	hasBackendService  bool
	hasForwardingRules bool
	hasTargetTCPProxy  bool
}

// CreateLoadBalancers creates all necessary load balancers.
func (c *Client) CreateLoadBalancers(ctx context.Context, isDebugCluster bool) error {
	if err := c.createIPAddr(ctx); err != nil {
		return fmt.Errorf("creating load balancer IP address: %w", err)
	}

	//
	// LoadBalancer definitions.
	//
	// LoadBalancers added here also need to be referenced in instances.go:*Client.CreateInstances

	c.loadbalancers = append(c.loadbalancers, &loadBalancer{
		name:            c.buildResourceName("kube"),
		ip:              c.loadbalancerIP,
		frontendPort:    constants.KubernetesPort,
		backendPortName: "kubernetes",
		healthCheck:     computepb.HealthCheck_HTTPS,
		label:           true, // Label, so bootstrapper can find kube-apiserver.
	})

	c.loadbalancers = append(c.loadbalancers, &loadBalancer{
		name:            c.buildResourceName("boot"),
		ip:              c.loadbalancerIPname,
		frontendPort:    constants.BootstrapperPort,
		backendPortName: "bootstrapper",
		healthCheck:     computepb.HealthCheck_TCP,
	})

	c.loadbalancers = append(c.loadbalancers, &loadBalancer{
		name:            c.buildResourceName("verify"),
		ip:              c.loadbalancerIPname,
		frontendPort:    constants.VerifyServiceNodePortGRPC,
		backendPortName: "verify",
		healthCheck:     computepb.HealthCheck_TCP,
	})

	c.loadbalancers = append(c.loadbalancers, &loadBalancer{
		name:            c.buildResourceName("konnectivity"),
		ip:              c.loadbalancerIPname,
		frontendPort:    constants.KonnectivityPort,
		backendPortName: "konnectivity",
		healthCheck:     computepb.HealthCheck_TCP,
	})

	c.loadbalancers = append(c.loadbalancers, &loadBalancer{
		name:            c.buildResourceName("recovery"),
		ip:              c.loadbalancerIPname,
		frontendPort:    constants.RecoveryPort,
		backendPortName: "recovery",
		healthCheck:     computepb.HealthCheck_TCP,
	})

	// Only create when the debug cluster flag is set in the Constellation config
	if isDebugCluster {
		c.loadbalancers = append(c.loadbalancers, &loadBalancer{
			name:            c.buildResourceName("debugd"),
			ip:              c.loadbalancerIPname,
			frontendPort:    constants.DebugdPort,
			backendPortName: "debugd",
			healthCheck:     computepb.HealthCheck_TCP,
		})
	}

	// Load balancer creation.

	errC := make(chan error)
	createLB := func(ctx context.Context, lb *loadBalancer) {
		errC <- c.createLoadBalancer(ctx, lb)
	}

	for _, lb := range c.loadbalancers {
		go createLB(ctx, lb)
	}

	var err error
	for i := 0; i < len(c.loadbalancers); i++ {
		err = multierr.Append(err, <-errC)
	}

	return err
}

// createLoadBalancer creates a load balancer.
func (c *Client) createLoadBalancer(ctx context.Context, lb *loadBalancer) error {
	if err := c.createHealthCheck(ctx, lb); err != nil {
		return fmt.Errorf("creating health checks: %w", err)
	}
	if err := c.createBackendService(ctx, lb); err != nil {
		return fmt.Errorf("creating backend services: %w", err)
	}
	if err := c.createTargetTCPProxy(ctx, lb); err != nil {
		return fmt.Errorf("creating target TCP proxies: %w", err)
	}
	if err := c.createForwardingRules(ctx, lb); err != nil {
		return fmt.Errorf("creating forwarding rules: %w", err)
	}
	return nil
}

func (c *Client) createHealthCheck(ctx context.Context, lb *loadBalancer) error {
	req := &computepb.InsertHealthCheckRequest{
		Project: c.project,
		HealthCheckResource: &computepb.HealthCheck{
			Name:             proto.String(lb.name),
			Type:             proto.String(computepb.HealthCheck_Type_name[int32(lb.healthCheck)]),
			CheckIntervalSec: proto.Int32(1),
			TimeoutSec:       proto.Int32(1),
		},
	}
	switch lb.healthCheck {
	case computepb.HealthCheck_HTTPS:
		req.HealthCheckResource.HttpsHealthCheck = newHealthCheckHTTPS(lb.frontendPort)
	case computepb.HealthCheck_TCP:
		req.HealthCheckResource.TcpHealthCheck = newHealthCheckTCP(lb.frontendPort)
	}
	resp, err := c.healthChecksAPI.Insert(ctx, req)
	if err != nil {
		return fmt.Errorf("inserting health check: %w", err)
	}

	if err := c.waitForOperations(ctx, []Operation{resp}); err != nil {
		return fmt.Errorf("waiting for health check creation: %w", err)
	}

	lb.hasHealthCheck = true
	return nil
}

func (c *Client) createBackendService(ctx context.Context, lb *loadBalancer) error {
	req := &computepb.InsertBackendServiceRequest{
		Project: c.project,
		BackendServiceResource: &computepb.BackendService{
			Name:                proto.String(lb.name),
			Protocol:            proto.String(computepb.BackendService_Protocol_name[int32(computepb.BackendService_TCP)]),
			LoadBalancingScheme: proto.String(computepb.BackendService_LoadBalancingScheme_name[int32(computepb.BackendService_EXTERNAL)]),
			HealthChecks:        []string{c.resourceURI(scopeGlobal, "healthChecks", lb.name)},
			PortName:            proto.String(lb.backendPortName),
			Backends: []*computepb.Backend{
				{
					BalancingMode: proto.String(computepb.Backend_BalancingMode_name[int32(computepb.Backend_UTILIZATION)]),
					Group:         proto.String(c.resourceURI(scopeZone, "instanceGroups", c.controlPlaneInstanceGroup)),
				},
			},
		},
	}
	resp, err := c.backendServicesAPI.Insert(ctx, req)
	if err != nil {
		return fmt.Errorf("inserting backend services: %w", err)
	}

	if err := c.waitForOperations(ctx, []Operation{resp}); err != nil {
		return fmt.Errorf("waiting for backend services creation: %w", err)
	}

	lb.hasBackendService = true
	return nil
}

func (c *Client) createForwardingRules(ctx context.Context, lb *loadBalancer) error {
	req := &computepb.InsertGlobalForwardingRuleRequest{
		Project: c.project,
		ForwardingRuleResource: &computepb.ForwardingRule{
			Name:                proto.String(lb.name),
			IPAddress:           proto.String(c.resourceURI(scopeGlobal, "addresses", c.loadbalancerIPname)),
			IPProtocol:          proto.String(computepb.ForwardingRule_IPProtocolEnum_name[int32(computepb.ForwardingRule_TCP)]),
			LoadBalancingScheme: proto.String(computepb.ForwardingRule_LoadBalancingScheme_name[int32(computepb.ForwardingRule_EXTERNAL)]),
			PortRange:           proto.String(strconv.Itoa(lb.frontendPort)),

			Target: proto.String(c.resourceURI(scopeGlobal, "targetTcpProxies", lb.name)),
		},
	}
	resp, err := c.forwardingRulesAPI.Insert(ctx, req)
	if err != nil {
		return fmt.Errorf("inserting forwarding rules: %w", err)
	}

	if err := c.waitForOperations(ctx, []Operation{resp}); err != nil {
		return err
	}
	lb.hasForwardingRules = true

	if lb.label {
		return c.labelLoadBalancer(ctx, lb.name)
	}
	return nil
}

// labelLoadBalancer labels a load balancer (its forwarding rules) so that it can be found by applications in the cluster.
func (c *Client) labelLoadBalancer(ctx context.Context, name string) error {
	forwardingRule, err := c.forwardingRulesAPI.Get(ctx, &computepb.GetGlobalForwardingRuleRequest{
		Project:        c.project,
		ForwardingRule: name,
	})
	if err != nil {
		return fmt.Errorf("getting forwarding rule: %w", err)
	}
	if forwardingRule.LabelFingerprint == nil {
		return fmt.Errorf("forwarding rule %s has no label fingerprint", name)
	}

	resp, err := c.forwardingRulesAPI.SetLabels(ctx, &computepb.SetLabelsGlobalForwardingRuleRequest{
		Project:  c.project,
		Resource: name,
		GlobalSetLabelsRequestResource: &computepb.GlobalSetLabelsRequest{
			Labels:           map[string]string{"constellation-uid": c.uid},
			LabelFingerprint: forwardingRule.LabelFingerprint,
		},
	})
	if err != nil {
		return fmt.Errorf("setting forwarding rule labels: %w", err)
	}

	return c.waitForOperations(ctx, []Operation{resp})
}

func (c *Client) createTargetTCPProxy(ctx context.Context, lb *loadBalancer) error {
	req := &computepb.InsertTargetTcpProxyRequest{
		Project: c.project,
		TargetTcpProxyResource: &computepb.TargetTcpProxy{
			Name:    proto.String(lb.name),
			Service: proto.String(c.resourceURI(scopeGlobal, "backendServices", lb.name)),
		},
	}
	resp, err := c.targetTCPProxiesAPI.Insert(ctx, req)
	if err != nil {
		return fmt.Errorf("inserting target tcp proxy: %w", err)
	}

	if err := c.waitForOperations(ctx, []Operation{resp}); err != nil {
		return err
	}
	lb.hasTargetTCPProxy = true
	return nil
}

// TerminateLoadBalancers terminates all load balancers.
func (c *Client) TerminateLoadBalancers(ctx context.Context) error {
	errC := make(chan error)

	terminateLB := func(ctx context.Context, lb *loadBalancer) {
		errC <- c.terminateLoadBalancer(ctx, lb)
	}

	for _, lb := range c.loadbalancers {
		go terminateLB(ctx, lb)
	}

	var err error
	for i := 0; i < len(c.loadbalancers); i++ {
		err = multierr.Append(err, <-errC)
	}
	if err != nil && !isNotFoundError(err) {
		return err
	}

	if err := c.deleteIPAddr(ctx); err != nil {
		return fmt.Errorf("deleting load balancer IP address: %w", err)
	}

	c.loadbalancers = nil
	return nil
}

// terminateLoadBalancer removes the load balancer and its associated resources.
func (c *Client) terminateLoadBalancer(ctx context.Context, lb *loadBalancer) error {
	if lb == nil {
		return nil
	}
	if lb.name == "" {
		return errors.New("load balancer name is empty")
	}

	if lb.hasForwardingRules {
		if err := c.terminateForwadingRules(ctx, lb); err != nil {
			return fmt.Errorf("terminating forwarding rules: %w", err)
		}
	}

	if lb.hasTargetTCPProxy {
		if err := c.terminateTargetTCPProxy(ctx, lb); err != nil {
			return fmt.Errorf("terminating target tcp proxy: %w", err)
		}
	}

	if lb.hasBackendService {
		if err := c.terminateBackendService(ctx, lb); err != nil {
			return fmt.Errorf("terminating backend services: %w", err)
		}
	}

	if lb.hasHealthCheck {
		if err := c.terminateHealthCheck(ctx, lb); err != nil {
			return fmt.Errorf("terminating health checks: %w", err)
		}
	}

	lb.name = ""
	return nil
}

func (c *Client) terminateForwadingRules(ctx context.Context, lb *loadBalancer) error {
	resp, err := c.forwardingRulesAPI.Delete(ctx, &computepb.DeleteGlobalForwardingRuleRequest{
		Project:        c.project,
		ForwardingRule: lb.name,
	})
	if isNotFoundError(err) {
		lb.hasForwardingRules = false
		return nil
	}
	if err != nil {
		return fmt.Errorf("deleting forwarding rules: %w", err)
	}

	if err := c.waitForOperations(ctx, []Operation{resp}); err != nil {
		return err
	}

	lb.hasForwardingRules = false
	return nil
}

func (c *Client) terminateTargetTCPProxy(ctx context.Context, lb *loadBalancer) error {
	resp, err := c.targetTCPProxiesAPI.Delete(ctx, &computepb.DeleteTargetTcpProxyRequest{
		Project:        c.project,
		TargetTcpProxy: lb.name,
	})
	if isNotFoundError(err) {
		lb.hasTargetTCPProxy = false
		return nil
	}
	if err != nil {
		return fmt.Errorf("deleting target tcp proxy: %w", err)
	}

	if err := c.waitForOperations(ctx, []Operation{resp}); err != nil {
		return err
	}

	lb.hasTargetTCPProxy = false
	return nil
}

func (c *Client) terminateBackendService(ctx context.Context, lb *loadBalancer) error {
	resp, err := c.backendServicesAPI.Delete(ctx, &computepb.DeleteBackendServiceRequest{
		Project:        c.project,
		BackendService: lb.name,
	})
	if isNotFoundError(err) {
		lb.hasBackendService = false
		return nil
	}
	if err != nil {
		return fmt.Errorf("deleting backend services: %w", err)
	}

	if err := c.waitForOperations(ctx, []Operation{resp}); err != nil {
		return err
	}

	lb.hasBackendService = false
	return nil
}

func (c *Client) terminateHealthCheck(ctx context.Context, lb *loadBalancer) error {
	resp, err := c.healthChecksAPI.Delete(ctx, &computepb.DeleteHealthCheckRequest{
		Project:     c.project,
		HealthCheck: lb.name,
	})
	if isNotFoundError(err) {
		lb.hasHealthCheck = false
		return nil
	}
	if err != nil {
		return fmt.Errorf("deleting health checks: %w", err)
	}

	if err := c.waitForOperations(ctx, []Operation{resp}); err != nil {
		return err
	}

	lb.hasHealthCheck = false
	return nil
}

func (c *Client) createIPAddr(ctx context.Context) error {
	ipName := c.buildResourceName()
	insertReq := &computepb.InsertGlobalAddressRequest{
		Project: c.project,
		AddressResource: &computepb.Address{
			Name: proto.String(ipName),
		},
	}
	op, err := c.addressesAPI.Insert(ctx, insertReq)
	if err != nil {
		return fmt.Errorf("inserting address: %w", err)
	}
	if err := c.waitForOperations(ctx, []Operation{op}); err != nil {
		return err
	}
	c.loadbalancerIPname = ipName

	getReq := &computepb.GetGlobalAddressRequest{
		Project: c.project,
		Address: c.loadbalancerIPname,
	}
	addr, err := c.addressesAPI.Get(ctx, getReq)
	if err != nil {
		return fmt.Errorf("getting address: %w", err)
	}
	if addr.Address == nil {
		return fmt.Errorf("address response without address: %q", addr)
	}

	c.loadbalancerIP = *addr.Address
	return nil
}

func (c *Client) deleteIPAddr(ctx context.Context) error {
	if c.loadbalancerIPname == "" {
		return nil
	}

	req := &computepb.DeleteGlobalAddressRequest{
		Project: c.project,
		Address: c.loadbalancerIPname,
	}
	op, err := c.addressesAPI.Delete(ctx, req)
	if isNotFoundError(err) {
		c.loadbalancerIPname = ""
		return nil
	}
	if err != nil {
		return fmt.Errorf("deleting address: %w", err)
	}

	if err := c.waitForOperations(ctx, []Operation{op}); err != nil {
		return err
	}

	c.loadbalancerIPname = ""
	return nil
}

func newHealthCheckHTTPS(port int) *computepb.HTTPSHealthCheck {
	return &computepb.HTTPSHealthCheck{
		Host:        proto.String(""),
		Port:        proto.Int32(int32(port)),
		RequestPath: proto.String("/readyz"),
	}
}

func newHealthCheckTCP(port int) *computepb.TCPHealthCheck {
	return &computepb.TCPHealthCheck{
		Port: proto.Int32(int32(port)),
	}
}
