package cmd

import (
	"context"
	"errors"
	"strconv"

	"github.com/edgelesssys/constellation/cli/cloudprovider"
	"github.com/edgelesssys/constellation/cli/gcp"
	"github.com/edgelesssys/constellation/cli/gcp/client"
	"github.com/edgelesssys/constellation/internal/state"
)

type fakeGcpClient struct {
	nodes        gcp.Instances
	coordinators gcp.Instances

	nodesInstanceGroup       string
	coordinatorInstanceGroup string
	coordinatorTemplate      string
	nodeTemplate             string
	network                  string
	subnetwork               string
	firewalls                []string
	project                  string
	uid                      string
	name                     string
	zone                     string
	serviceAccount           string
}

func (c *fakeGcpClient) GetState() (state.ConstellationState, error) {
	stat := state.ConstellationState{
		CloudProvider:                  cloudprovider.GCP.String(),
		GCPNodes:                       c.nodes,
		GCPCoordinators:                c.coordinators,
		GCPNodeInstanceGroup:           c.nodesInstanceGroup,
		GCPCoordinatorInstanceGroup:    c.coordinatorInstanceGroup,
		GCPNodeInstanceTemplate:        c.nodeTemplate,
		GCPCoordinatorInstanceTemplate: c.coordinatorTemplate,
		GCPNetwork:                     c.network,
		GCPSubnetwork:                  c.subnetwork,
		GCPFirewalls:                   c.firewalls,
		GCPProject:                     c.project,
		Name:                           c.name,
		UID:                            c.uid,
		GCPZone:                        c.zone,
		GCPServiceAccount:              c.serviceAccount,
	}
	return stat, nil
}

func (c *fakeGcpClient) SetState(stat state.ConstellationState) error {
	c.nodes = stat.GCPNodes
	c.coordinators = stat.GCPCoordinators
	c.nodesInstanceGroup = stat.GCPNodeInstanceGroup
	c.coordinatorInstanceGroup = stat.GCPCoordinatorInstanceGroup
	c.nodeTemplate = stat.GCPNodeInstanceTemplate
	c.coordinatorTemplate = stat.GCPCoordinatorInstanceTemplate
	c.network = stat.GCPNetwork
	c.subnetwork = stat.GCPSubnetwork
	c.firewalls = stat.GCPFirewalls
	c.project = stat.GCPProject
	c.name = stat.Name
	c.uid = stat.UID
	c.zone = stat.GCPZone
	c.serviceAccount = stat.GCPServiceAccount
	return nil
}

func (c *fakeGcpClient) CreateVPCs(ctx context.Context, input client.VPCsInput) error {
	c.network = "network"
	c.subnetwork = "subnetwork"
	return nil
}

func (c *fakeGcpClient) CreateFirewall(ctx context.Context, input client.FirewallInput) error {
	if c.network == "" {
		return errors.New("client has not network")
	}
	var firewalls []string
	for _, rule := range input.Ingress {
		firewalls = append(firewalls, rule.Name)
	}
	c.firewalls = firewalls
	return nil
}

func (c *fakeGcpClient) CreateInstances(ctx context.Context, input client.CreateInstancesInput) error {
	c.coordinatorInstanceGroup = "coordinator-group"
	c.nodesInstanceGroup = "nodes-group"
	c.nodeTemplate = "node-template"
	c.coordinatorTemplate = "coordinator-template"
	c.nodes = make(gcp.Instances)
	for i := 0; i < input.CountNodes; i++ {
		id := "id-" + strconv.Itoa(i)
		c.nodes[id] = gcp.Instance{PublicIP: "192.0.2.1", PrivateIP: "192.0.2.1"}
	}
	c.coordinators = make(gcp.Instances)
	for i := 0; i < input.CountCoordinators; i++ {
		id := "id-" + strconv.Itoa(i)
		c.coordinators[id] = gcp.Instance{PublicIP: "192.0.2.1", PrivateIP: "192.0.2.1"}
	}
	return nil
}

func (c *fakeGcpClient) CreateServiceAccount(ctx context.Context, input client.ServiceAccountInput) (string, error) {
	c.serviceAccount = "service-account@" + c.project + ".iam.gserviceaccount.com"
	return client.ServiceAccountKey{
		Type:                    "service_account",
		ProjectID:               c.project,
		PrivateKeyID:            "key-id",
		PrivateKey:              "-----BEGIN PRIVATE KEY-----\nprivate-key\n-----END PRIVATE KEY-----\n",
		ClientEmail:             c.serviceAccount,
		ClientID:                "client-id",
		AuthURI:                 "https://accounts.google.com/o/oauth2/auth",
		TokenURI:                "https://accounts.google.com/o/oauth2/token",
		AuthProviderX509CertURL: "https://www.googleapis.com/oauth2/v1/certs",
		ClientX509CertURL:       "https://www.googleapis.com/robot/v1/metadata/x509/service-account-email",
	}.ConvertToCloudServiceAccountURI(), nil
}

func (c *fakeGcpClient) TerminateFirewall(ctx context.Context) error {
	if len(c.firewalls) == 0 {
		return nil
	}
	c.firewalls = nil
	return nil
}

func (c *fakeGcpClient) TerminateVPCs(context.Context) error {
	if len(c.firewalls) != 0 {
		return errors.New("client has firewalls, which must be deleted first")
	}
	c.network = ""
	c.subnetwork = ""
	return nil
}

func (c *fakeGcpClient) TerminateInstances(context.Context) error {
	c.nodeTemplate = ""
	c.coordinatorTemplate = ""
	c.nodesInstanceGroup = ""
	c.coordinatorInstanceGroup = ""
	c.nodes = nil
	c.coordinators = nil
	return nil
}

func (c *fakeGcpClient) TerminateServiceAccount(context.Context) error {
	c.serviceAccount = ""
	return nil
}

func (c *fakeGcpClient) Close() error {
	return nil
}

type stubGcpClient struct {
	terminateFirewallCalled  bool
	terminateInstancesCalled bool
	terminateVPCsCalled      bool

	getStateErr                error
	setStateErr                error
	createVPCsErr              error
	createFirewallErr          error
	createInstancesErr         error
	createServiceAccountErr    error
	terminateFirewallErr       error
	terminateVPCsErr           error
	terminateInstancesErr      error
	terminateServiceAccountErr error
	closeErr                   error
}

func (c *stubGcpClient) GetState() (state.ConstellationState, error) {
	return state.ConstellationState{}, c.getStateErr
}

func (c *stubGcpClient) SetState(state.ConstellationState) error {
	return c.setStateErr
}

func (c *stubGcpClient) CreateVPCs(ctx context.Context, input client.VPCsInput) error {
	return c.createVPCsErr
}

func (c *stubGcpClient) CreateFirewall(ctx context.Context, input client.FirewallInput) error {
	return c.createFirewallErr
}

func (c *stubGcpClient) CreateInstances(ctx context.Context, input client.CreateInstancesInput) error {
	return c.createInstancesErr
}

func (c *stubGcpClient) CreateServiceAccount(ctx context.Context, input client.ServiceAccountInput) (string, error) {
	return client.ServiceAccountKey{}.ConvertToCloudServiceAccountURI(), c.createServiceAccountErr
}

func (c *stubGcpClient) TerminateFirewall(ctx context.Context) error {
	c.terminateFirewallCalled = true
	return c.terminateFirewallErr
}

func (c *stubGcpClient) TerminateVPCs(context.Context) error {
	c.terminateVPCsCalled = true
	return c.terminateVPCsErr
}

func (c *stubGcpClient) TerminateInstances(context.Context) error {
	c.terminateInstancesCalled = true
	return c.terminateInstancesErr
}

func (c *stubGcpClient) TerminateServiceAccount(context.Context) error {
	return c.terminateServiceAccountErr
}

func (c *stubGcpClient) Close() error {
	return c.closeErr
}
