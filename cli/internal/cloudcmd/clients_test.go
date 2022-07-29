package cloudcmd

import (
	"context"
	"errors"
	"strconv"
	"testing"

	azurecl "github.com/edgelesssys/constellation/cli/internal/azure/client"
	gcpcl "github.com/edgelesssys/constellation/cli/internal/gcp/client"
	"github.com/edgelesssys/constellation/internal/azureshared"
	"github.com/edgelesssys/constellation/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/internal/cloud/cloudtypes"
	"github.com/edgelesssys/constellation/internal/gcpshared"
	"github.com/edgelesssys/constellation/internal/state"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m,
		// https://github.com/census-instrumentation/opencensus-go/issues/1262
		goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start"),
	)
}

type fakeAzureClient struct {
	workers       cloudtypes.Instances
	controlPlanes cloudtypes.Instances

	resourceGroup        string
	name                 string
	uid                  string
	location             string
	subscriptionID       string
	tenantID             string
	subnetID             string
	loadBalancerName     string
	controlPlaneScaleSet string
	workerScaleSet       string
	networkSecurityGroup string
	adAppObjectID        string
}

func (c *fakeAzureClient) GetState() (state.ConstellationState, error) {
	stat := state.ConstellationState{
		CloudProvider:              cloudprovider.Azure.String(),
		AzureWorkerInstances:       c.workers,
		AzureControlPlaneInstances: c.controlPlanes,
		Name:                       c.name,
		UID:                        c.uid,
		AzureResourceGroup:         c.resourceGroup,
		AzureLocation:              c.location,
		AzureSubscription:          c.subscriptionID,
		AzureTenant:                c.tenantID,
		AzureSubnet:                c.subnetID,
		AzureNetworkSecurityGroup:  c.networkSecurityGroup,
		AzureWorkerScaleSet:        c.workerScaleSet,
		AzureControlPlaneScaleSet:  c.controlPlaneScaleSet,
		AzureADAppObjectID:         c.adAppObjectID,
	}
	return stat, nil
}

func (c *fakeAzureClient) SetState(stat state.ConstellationState) error {
	c.workers = stat.AzureWorkerInstances
	c.controlPlanes = stat.AzureControlPlaneInstances
	c.name = stat.Name
	c.uid = stat.UID
	c.resourceGroup = stat.AzureResourceGroup
	c.location = stat.AzureLocation
	c.subscriptionID = stat.AzureSubscription
	c.tenantID = stat.AzureTenant
	c.subnetID = stat.AzureSubnet
	c.networkSecurityGroup = stat.AzureNetworkSecurityGroup
	c.workerScaleSet = stat.AzureWorkerScaleSet
	c.controlPlaneScaleSet = stat.AzureControlPlaneScaleSet
	c.adAppObjectID = stat.AzureADAppObjectID
	return nil
}

func (c *fakeAzureClient) CreateApplicationInsight(ctx context.Context) error {
	return nil
}

func (c *fakeAzureClient) CreateResourceGroup(ctx context.Context) error {
	c.resourceGroup = "resource-group"
	return nil
}

func (c *fakeAzureClient) CreateVirtualNetwork(ctx context.Context) error {
	c.subnetID = "subnet"
	return nil
}

func (c *fakeAzureClient) CreateExternalLoadBalancer(ctx context.Context) error {
	c.loadBalancerName = "loadBalancer"
	return nil
}

func (c *fakeAzureClient) CreateSecurityGroup(ctx context.Context, input azurecl.NetworkSecurityGroupInput) error {
	c.networkSecurityGroup = "network-security-group"
	return nil
}

func (c *fakeAzureClient) CreateInstances(ctx context.Context, input azurecl.CreateInstancesInput) error {
	c.controlPlaneScaleSet = "controlplanes-scale-set"
	c.workerScaleSet = "workers-scale-set"
	c.workers = make(cloudtypes.Instances)
	for i := 0; i < input.CountWorkers; i++ {
		id := "id-" + strconv.Itoa(i)
		c.workers[id] = cloudtypes.Instance{PublicIP: "192.0.2.1", PrivateIP: "192.0.2.1"}
	}
	c.controlPlanes = make(cloudtypes.Instances)
	for i := 0; i < input.CountControlPlanes; i++ {
		id := "id-" + strconv.Itoa(i)
		c.controlPlanes[id] = cloudtypes.Instance{PublicIP: "192.0.2.1", PrivateIP: "192.0.2.1"}
	}
	return nil
}

// TODO: deprecate as soon as scale sets are available.
func (c *fakeAzureClient) CreateInstancesVMs(ctx context.Context, input azurecl.CreateInstancesInput) error {
	c.workers = make(cloudtypes.Instances)
	for i := 0; i < input.CountWorkers; i++ {
		id := "id-" + strconv.Itoa(i)
		c.workers[id] = cloudtypes.Instance{PublicIP: "192.0.2.1", PrivateIP: "192.0.2.1"}
	}
	c.controlPlanes = make(cloudtypes.Instances)
	for i := 0; i < input.CountControlPlanes; i++ {
		id := "id-" + strconv.Itoa(i)
		c.controlPlanes[id] = cloudtypes.Instance{PublicIP: "192.0.2.1", PrivateIP: "192.0.2.1"}
	}
	return nil
}

func (c *fakeAzureClient) CreateServicePrincipal(ctx context.Context) (string, error) {
	c.adAppObjectID = "00000000-0000-0000-0000-000000000001"
	return azureshared.ApplicationCredentials{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
	}.ToCloudServiceAccountURI(), nil
}

func (c *fakeAzureClient) TerminateResourceGroup(ctx context.Context) error {
	if c.resourceGroup == "" {
		return nil
	}
	c.workers = nil
	c.controlPlanes = nil
	c.resourceGroup = ""
	c.subnetID = ""
	c.networkSecurityGroup = ""
	c.workerScaleSet = ""
	c.controlPlaneScaleSet = ""
	return nil
}

func (c *fakeAzureClient) TerminateServicePrincipal(ctx context.Context) error {
	if c.adAppObjectID == "" {
		return nil
	}
	c.adAppObjectID = ""
	return nil
}

type stubAzureClient struct {
	terminateResourceGroupCalled    bool
	terminateServicePrincipalCalled bool

	getStateErr                  error
	setStateErr                  error
	createApplicationInsightErr  error
	createResourceGroupErr       error
	createVirtualNetworkErr      error
	createSecurityGroupErr       error
	createLoadBalancerErr        error
	createInstancesErr           error
	createServicePrincipalErr    error
	terminateResourceGroupErr    error
	terminateServicePrincipalErr error
}

func (c *stubAzureClient) GetState() (state.ConstellationState, error) {
	return state.ConstellationState{}, c.getStateErr
}

func (c *stubAzureClient) SetState(state.ConstellationState) error {
	return c.setStateErr
}

func (c *stubAzureClient) CreateExternalLoadBalancer(ctx context.Context) error {
	return c.createLoadBalancerErr
}

func (c *stubAzureClient) CreateApplicationInsight(ctx context.Context) error {
	return c.createApplicationInsightErr
}

func (c *stubAzureClient) CreateResourceGroup(ctx context.Context) error {
	return c.createResourceGroupErr
}

func (c *stubAzureClient) CreateVirtualNetwork(ctx context.Context) error {
	return c.createVirtualNetworkErr
}

func (c *stubAzureClient) CreateSecurityGroup(ctx context.Context, input azurecl.NetworkSecurityGroupInput) error {
	return c.createSecurityGroupErr
}

func (c *stubAzureClient) CreateInstances(ctx context.Context, input azurecl.CreateInstancesInput) error {
	return c.createInstancesErr
}

// TODO: deprecate as soon as scale sets are available.
func (c *stubAzureClient) CreateInstancesVMs(ctx context.Context, input azurecl.CreateInstancesInput) error {
	return c.createInstancesErr
}

func (c *stubAzureClient) CreateServicePrincipal(ctx context.Context) (string, error) {
	return azureshared.ApplicationCredentials{
		ClientID:     "00000000-0000-0000-0000-000000000000",
		ClientSecret: "secret",
	}.ToCloudServiceAccountURI(), c.createServicePrincipalErr
}

func (c *stubAzureClient) TerminateResourceGroup(ctx context.Context) error {
	c.terminateResourceGroupCalled = true
	return c.terminateResourceGroupErr
}

func (c *stubAzureClient) TerminateServicePrincipal(ctx context.Context) error {
	c.terminateServicePrincipalCalled = true
	return c.terminateServicePrincipalErr
}

type fakeGcpClient struct {
	workers       cloudtypes.Instances
	controlPlanes cloudtypes.Instances

	workerInstanceGroup       string
	controlPlaneInstanceGroup string
	controlPlaneTemplate      string
	workerTemplate            string
	network                   string
	subnetwork                string
	firewalls                 []string
	project                   string
	uid                       string
	name                      string
	zone                      string
	serviceAccount            string

	// loadbalancer
	healthCheck    string
	backendService string
	forwardingRule string
}

func (c *fakeGcpClient) GetState() (state.ConstellationState, error) {
	stat := state.ConstellationState{
		CloudProvider:                   cloudprovider.GCP.String(),
		GCPWorkerInstances:              c.workers,
		GCPControlPlaneInstances:        c.controlPlanes,
		GCPWorkerInstanceGroup:          c.workerInstanceGroup,
		GCPControlPlaneInstanceGroup:    c.controlPlaneInstanceGroup,
		GCPWorkerInstanceTemplate:       c.workerTemplate,
		GCPControlPlaneInstanceTemplate: c.controlPlaneTemplate,
		GCPNetwork:                      c.network,
		GCPSubnetwork:                   c.subnetwork,
		GCPFirewalls:                    c.firewalls,
		GCPBackendService:               c.backendService,
		GCPHealthCheck:                  c.healthCheck,
		GCPForwardingRule:               c.forwardingRule,
		GCPProject:                      c.project,
		Name:                            c.name,
		UID:                             c.uid,
		GCPZone:                         c.zone,
		GCPServiceAccount:               c.serviceAccount,
	}
	return stat, nil
}

func (c *fakeGcpClient) SetState(stat state.ConstellationState) error {
	c.workers = stat.GCPWorkerInstances
	c.controlPlanes = stat.GCPControlPlaneInstances
	c.workerInstanceGroup = stat.GCPWorkerInstanceGroup
	c.controlPlaneInstanceGroup = stat.GCPControlPlaneInstanceGroup
	c.workerTemplate = stat.GCPWorkerInstanceTemplate
	c.controlPlaneTemplate = stat.GCPControlPlaneInstanceTemplate
	c.network = stat.GCPNetwork
	c.subnetwork = stat.GCPSubnetwork
	c.firewalls = stat.GCPFirewalls
	c.project = stat.GCPProject
	c.name = stat.Name
	c.uid = stat.UID
	c.zone = stat.GCPZone
	c.serviceAccount = stat.GCPServiceAccount
	c.healthCheck = stat.GCPHealthCheck
	c.backendService = stat.GCPBackendService
	c.forwardingRule = stat.GCPForwardingRule
	return nil
}

func (c *fakeGcpClient) CreateVPCs(ctx context.Context) error {
	c.network = "network"
	c.subnetwork = "subnetwork"
	return nil
}

func (c *fakeGcpClient) CreateFirewall(ctx context.Context, input gcpcl.FirewallInput) error {
	if c.network == "" {
		return errors.New("client has not network")
	}
	for _, rule := range input.Ingress {
		c.firewalls = append(c.firewalls, rule.Name)
	}
	return nil
}

func (c *fakeGcpClient) CreateInstances(ctx context.Context, input gcpcl.CreateInstancesInput) error {
	c.controlPlaneInstanceGroup = "controlplane-group"
	c.workerInstanceGroup = "workers-group"
	c.workerTemplate = "worker-template"
	c.controlPlaneTemplate = "controlplane-template"
	c.workers = make(cloudtypes.Instances)
	for i := 0; i < input.CountWorkers; i++ {
		id := "id-" + strconv.Itoa(i)
		c.workers[id] = cloudtypes.Instance{PublicIP: "192.0.2.1", PrivateIP: "192.0.2.1"}
	}
	c.controlPlanes = make(cloudtypes.Instances)
	for i := 0; i < input.CountControlPlanes; i++ {
		id := "id-" + strconv.Itoa(i)
		c.controlPlanes[id] = cloudtypes.Instance{PublicIP: "192.0.2.1", PrivateIP: "192.0.2.1"}
	}
	return nil
}

func (c *fakeGcpClient) CreateServiceAccount(ctx context.Context, input gcpcl.ServiceAccountInput) (string, error) {
	c.serviceAccount = "service-account@" + c.project + ".iam.gserviceaccount.com"
	return gcpshared.ServiceAccountKey{
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
	}.ToCloudServiceAccountURI(), nil
}

func (c *fakeGcpClient) CreateLoadBalancer(ctx context.Context) error {
	c.healthCheck = "health-check"
	c.backendService = "backend-service"
	c.forwardingRule = "forwarding-rule"
	return nil
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
	c.workerTemplate = ""
	c.controlPlaneTemplate = ""
	c.workerInstanceGroup = ""
	c.controlPlaneInstanceGroup = ""
	c.workers = nil
	c.controlPlanes = nil
	return nil
}

func (c *fakeGcpClient) TerminateServiceAccount(context.Context) error {
	c.serviceAccount = ""
	return nil
}

func (c *fakeGcpClient) TerminateLoadBalancer(context.Context) error {
	c.healthCheck = ""
	c.backendService = ""
	c.forwardingRule = ""
	return nil
}

func (c *fakeGcpClient) Close() error {
	return nil
}

type stubGcpClient struct {
	terminateFirewallCalled       bool
	terminateInstancesCalled      bool
	terminateVPCsCalled           bool
	terminateServiceAccountCalled bool
	closeCalled                   bool

	getStateErr                error
	setStateErr                error
	createVPCsErr              error
	createFirewallErr          error
	createInstancesErr         error
	createServiceAccountErr    error
	createLoadBalancerErr      error
	terminateFirewallErr       error
	terminateVPCsErr           error
	terminateInstancesErr      error
	terminateServiceAccountErr error
	terminateLoadBalancerErr   error
	closeErr                   error
}

func (c *stubGcpClient) GetState() (state.ConstellationState, error) {
	return state.ConstellationState{}, c.getStateErr
}

func (c *stubGcpClient) SetState(state.ConstellationState) error {
	return c.setStateErr
}

func (c *stubGcpClient) CreateVPCs(ctx context.Context) error {
	return c.createVPCsErr
}

func (c *stubGcpClient) CreateFirewall(ctx context.Context, input gcpcl.FirewallInput) error {
	return c.createFirewallErr
}

func (c *stubGcpClient) CreateInstances(ctx context.Context, input gcpcl.CreateInstancesInput) error {
	return c.createInstancesErr
}

func (c *stubGcpClient) CreateServiceAccount(ctx context.Context, input gcpcl.ServiceAccountInput) (string, error) {
	return gcpshared.ServiceAccountKey{}.ToCloudServiceAccountURI(), c.createServiceAccountErr
}

func (c *stubGcpClient) CreateLoadBalancer(ctx context.Context) error {
	return c.createLoadBalancerErr
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
	c.terminateServiceAccountCalled = true
	return c.terminateServiceAccountErr
}

func (c *stubGcpClient) TerminateLoadBalancer(context.Context) error {
	return c.terminateLoadBalancerErr
}

func (c *stubGcpClient) Close() error {
	c.closeCalled = true
	return c.closeErr
}
