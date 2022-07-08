package client

import (
	"context"
	"errors"
	"fmt"

	"github.com/edgelesssys/constellation/internal/cloud/cloudtypes"
	computepb "google.golang.org/genproto/googleapis/cloud/compute/v1"
	"google.golang.org/protobuf/proto"
)

const (
	SubnetCIDR    = "192.168.178.0/24"
	SubnetExtCIDR = "10.10.0.0/16"
)

// CreateFirewall creates a set of firewall rules for the client's network.
//
// The client must have a VPC network to set firewall rules.
func (c *Client) CreateFirewall(ctx context.Context, input FirewallInput) error {
	if c.network == "" {
		return errors.New("client has not network")
	}
	firewallRules, err := input.Ingress.GCP()
	if err != nil {
		return err
	}
	var ops []Operation
	for _, rule := range firewallRules {
		c.firewalls = append(c.firewalls, rule.GetName())

		rule.Network = proto.String("global/networks/" + c.network)
		rule.Name = proto.String(rule.GetName() + "-" + c.uid)
		req := &computepb.InsertFirewallRequest{
			FirewallResource: rule,
			Project:          c.project,
		}
		resp, err := c.firewallsAPI.Insert(ctx, req)
		if err != nil {
			return err
		}
		if resp.Proto().Name == nil {
			return errors.New("operation name is nil")
		}
		ops = append(ops, resp)
	}
	return c.waitForOperations(ctx, ops)
}

// TerminateFirewall deletes firewall rules from the client's network.
//
// The client must have a VPC network to set firewall rules.
func (c *Client) TerminateFirewall(ctx context.Context) error {
	if len(c.firewalls) == 0 {
		return nil
	}
	var ops []Operation
	for _, name := range c.firewalls {
		ruleName := name + "-" + c.uid
		req := &computepb.DeleteFirewallRequest{
			Firewall: ruleName,
			Project:  c.project,
		}
		resp, err := c.firewallsAPI.Delete(ctx, req)
		if err != nil {
			return err
		}
		if resp.Proto().Name == nil {
			return errors.New("operation name is nil")
		}
		ops = append(ops, resp)
	}
	if err := c.waitForOperations(ctx, ops); err != nil {
		return err
	}
	c.firewalls = []string{}
	return nil
}

// FirewallInput defines firewall rules to be set.
type FirewallInput struct {
	Ingress cloudtypes.Firewall
	Egress  cloudtypes.Firewall
}

// CreateVPCs creates all necessary VPC networks.
func (c *Client) CreateVPCs(ctx context.Context) error {
	c.network = c.name + "-" + c.uid

	op, err := c.createVPC(ctx, c.network)
	if err != nil {
		return err
	}
	if err := c.waitForOperations(ctx, []Operation{op}); err != nil {
		return err
	}

	if err := c.createSubnets(ctx); err != nil {
		return err
	}

	return nil
}

// createVPC creates a VPC network.
func (c *Client) createVPC(ctx context.Context, name string) (Operation, error) {
	req := &computepb.InsertNetworkRequest{
		NetworkResource: &computepb.Network{
			AutoCreateSubnetworks: proto.Bool(false),
			Description:           proto.String("Constellation VPC"),
			Name:                  proto.String(name),
		},
		Project: c.project,
	}
	return c.networksAPI.Insert(ctx, req)
}

// TerminateVPCs terminates all VPC networks.
//
// If the any network has firewall rules, these must be terminated first.
func (c *Client) TerminateVPCs(ctx context.Context) error {
	if len(c.firewalls) != 0 {
		return errors.New("client has firewalls, which must be deleted first")
	}

	if err := c.terminateSubnet(ctx); err != nil {
		return err
	}

	var op Operation
	var err error
	if c.network != "" {
		op, err = c.terminateVPC(ctx, c.network)
		if err != nil {
			return err
		}
		c.network = ""
	}

	return c.waitForOperations(ctx, []Operation{op})
}

// terminateVPC terminates a VPC network.
//
// If the network has firewall rules, these must be terminated first.
func (c *Client) terminateVPC(ctx context.Context, network string) (Operation, error) {
	req := &computepb.DeleteNetworkRequest{
		Project: c.project,
		Network: network,
	}
	return c.networksAPI.Delete(ctx, req)
}

func (c *Client) createSubnets(ctx context.Context) error {
	c.subnetwork = "node-net-" + c.uid
	c.secondarySubnetworkRange = "net-ext" + c.uid

	op, err := c.createSubnet(ctx, c.subnetwork, c.network, c.secondarySubnetworkRange)
	if err != nil {
		return err
	}

	return c.waitForOperations(ctx, []Operation{op})
}

func (c *Client) createSubnet(ctx context.Context, name, network, secondaryRangeName string) (Operation, error) {
	req := &computepb.InsertSubnetworkRequest{
		Project: c.project,
		Region:  c.region,
		SubnetworkResource: &computepb.Subnetwork{
			IpCidrRange: proto.String(SubnetCIDR),
			Name:        proto.String(name),
			Network:     proto.String("projects/" + c.project + "/global/networks/" + network),
			SecondaryIpRanges: []*computepb.SubnetworkSecondaryRange{
				{
					RangeName:   proto.String(secondaryRangeName),
					IpCidrRange: proto.String(SubnetExtCIDR),
				},
			},
		},
	}
	return c.subnetworksAPI.Insert(ctx, req)
}

func (c *Client) terminateSubnet(ctx context.Context) error {
	if c.subnetwork == "" {
		return nil
	}
	req := &computepb.DeleteSubnetworkRequest{
		Project:    c.project,
		Region:     c.region,
		Subnetwork: c.subnetwork,
	}
	op, err := c.subnetworksAPI.Delete(ctx, req)
	if err != nil {
		return err
	}
	return c.waitForOperations(ctx, []Operation{op})
}

// CreateLoadBalancer creates a load balancer.
func (c *Client) CreateLoadBalancer(ctx context.Context) error {
	c.healthCheck = c.name + "-" + c.uid
	resp, err := c.healthChecksAPI.Insert(ctx, &computepb.InsertRegionHealthCheckRequest{
		Project: c.project,
		Region:  c.region,
		HealthCheckResource: &computepb.HealthCheck{
			Name:             proto.String(c.healthCheck),
			Type:             proto.String(computepb.HealthCheck_Type_name[int32(computepb.HealthCheck_HTTPS)]),
			CheckIntervalSec: proto.Int32(1),
			TimeoutSec:       proto.Int32(1),
			HttpsHealthCheck: &computepb.HTTPSHealthCheck{
				Host:        proto.String(""),
				Port:        proto.Int32(6443),
				RequestPath: proto.String("/readyz"),
			},
		},
	})
	if err != nil {
		return err
	}
	if err := c.waitForOperations(ctx, []Operation{resp}); err != nil {
		return err
	}

	c.backendService = c.name + "-" + c.uid
	resp, err = c.backendServicesAPI.Insert(ctx, &computepb.InsertRegionBackendServiceRequest{
		Project: c.project,
		Region:  c.region,
		BackendServiceResource: &computepb.BackendService{
			Name:                proto.String(c.backendService),
			Protocol:            proto.String(computepb.BackendService_Protocol_name[int32(computepb.BackendService_TCP)]),
			LoadBalancingScheme: proto.String(computepb.BackendService_LoadBalancingScheme_name[int32(computepb.BackendService_EXTERNAL)]),
			TimeoutSec:          proto.Int32(10),
			HealthChecks:        []string{"https://www.googleapis.com/compute/v1/projects/" + c.project + "/regions/" + c.region + "/healthChecks/" + c.healthCheck},
			Backends: []*computepb.Backend{
				{
					BalancingMode: proto.String(computepb.Backend_BalancingMode_name[int32(computepb.Backend_CONNECTION)]),
					Group:         proto.String("https://www.googleapis.com/compute/v1/projects/" + c.project + "/zones/" + c.zone + "/instanceGroups/" + c.controlPlaneInstanceGroup),
				},
			},
		},
	})
	if err != nil {
		return err
	}
	if err := c.waitForOperations(ctx, []Operation{resp}); err != nil {
		return err
	}

	c.forwardingRule = c.name + "-" + c.uid
	resp, err = c.forwardingRulesAPI.Insert(ctx, &computepb.InsertForwardingRuleRequest{
		Project: c.project,
		Region:  c.region,
		ForwardingRuleResource: &computepb.ForwardingRule{
			Name:                proto.String(c.forwardingRule),
			IPProtocol:          proto.String(computepb.ForwardingRule_IPProtocolEnum_name[int32(computepb.ForwardingRule_TCP)]),
			LoadBalancingScheme: proto.String(computepb.ForwardingRule_LoadBalancingScheme_name[int32(computepb.ForwardingRule_EXTERNAL)]),
			Ports:               []string{"6443", "9000"},
			BackendService:      proto.String("https://www.googleapis.com/compute/v1/projects/" + c.project + "/regions/" + c.region + "/backendServices/" + c.backendService),
		},
	})
	if err != nil {
		return err
	}
	if err := c.waitForOperations(ctx, []Operation{resp}); err != nil {
		return err
	}

	forwardingRule, err := c.forwardingRulesAPI.Get(ctx, &computepb.GetForwardingRuleRequest{
		Project:        c.project,
		Region:         c.region,
		ForwardingRule: c.forwardingRule,
	})
	if err != nil {
		return err
	}
	if forwardingRule.LabelFingerprint == nil {
		return fmt.Errorf("forwarding rule %s has no label fingerprint", c.forwardingRule)
	}

	resp, err = c.forwardingRulesAPI.SetLabels(ctx, &computepb.SetLabelsForwardingRuleRequest{
		Project:  c.project,
		Region:   c.region,
		Resource: c.forwardingRule,
		RegionSetLabelsRequestResource: &computepb.RegionSetLabelsRequest{
			Labels:           map[string]string{"constellation-uid": c.uid},
			LabelFingerprint: forwardingRule.LabelFingerprint,
		},
	})
	if err != nil {
		return err
	}

	return c.waitForOperations(ctx, []Operation{resp})
}

// TerminateLoadBalancer removes the load balancer and its associated resources.
func (c *Client) TerminateLoadBalancer(ctx context.Context) error {
	resp, err := c.forwardingRulesAPI.Delete(ctx, &computepb.DeleteForwardingRuleRequest{
		Project:        c.project,
		Region:         c.region,
		ForwardingRule: c.forwardingRule,
	})
	if err != nil {
		return err
	}
	if err := c.waitForOperations(ctx, []Operation{resp}); err != nil {
		return err
	}

	resp, err = c.backendServicesAPI.Delete(ctx, &computepb.DeleteRegionBackendServiceRequest{
		Project:        c.project,
		Region:         c.region,
		BackendService: c.backendService,
	})
	if err != nil {
		return err
	}
	if err := c.waitForOperations(ctx, []Operation{resp}); err != nil {
		return err
	}

	resp, err = c.healthChecksAPI.Delete(ctx, &computepb.DeleteRegionHealthCheckRequest{
		Project:     c.project,
		Region:      c.region,
		HealthCheck: c.healthCheck,
	})
	if err != nil {
		return err
	}

	return c.waitForOperations(ctx, []Operation{resp})
}
