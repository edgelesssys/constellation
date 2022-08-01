package client

import (
	"context"
	"errors"

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
		if isNotFoundError(err) {
			continue
		}
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

	if err := c.terminateVPC(ctx); err != nil {
		return err
	}

	return nil
}

// terminateVPC terminates a VPC network.
//
// If the network has firewall rules, these must be terminated first.
func (c *Client) terminateVPC(ctx context.Context) error {
	if c.network == "" {
		return nil
	}

	req := &computepb.DeleteNetworkRequest{
		Project: c.project,
		Network: c.network,
	}
	op, err := c.networksAPI.Delete(ctx, req)
	if err != nil && !isNotFoundError(err) {
		return err
	}
	if isNotFoundError(err) {
		c.network = ""
		return nil
	}

	if err := c.waitForOperations(ctx, []Operation{op}); err != nil {
		return err
	}

	c.network = ""
	return nil
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
	if err != nil && !isNotFoundError(err) {
		return err
	}
	if isNotFoundError(err) {
		c.subnetwork = ""
		return nil
	}

	if err := c.waitForOperations(ctx, []Operation{op}); err != nil {
		return err
	}

	c.subnetwork = ""
	return nil
}
