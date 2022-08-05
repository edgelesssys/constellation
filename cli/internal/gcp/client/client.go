package client

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"

	compute "cloud.google.com/go/compute/apiv1"
	admin "cloud.google.com/go/iam/admin/apiv1"
	resourcemanager "cloud.google.com/go/resourcemanager/apiv3"
	"github.com/edgelesssys/constellation/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/internal/cloud/cloudtypes"
	"github.com/edgelesssys/constellation/internal/state"
	"go.uber.org/multierr"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/googleapi"
)

// Client is a client for the Google Compute Engine.
type Client struct {
	instanceAPI
	operationRegionAPI
	operationZoneAPI
	operationGlobalAPI
	networksAPI
	subnetworksAPI
	firewallsAPI
	forwardingRulesAPI
	backendServicesAPI
	healthChecksAPI
	instanceTemplateAPI
	instanceGroupManagersAPI
	iamAPI
	projectsAPI
	addressesAPI

	workers       cloudtypes.Instances
	controlPlanes cloudtypes.Instances

	workerInstanceGroup       string
	controlPlaneInstanceGroup string
	controlPlaneTemplate      string
	workerTemplate            string
	network                   string
	subnetwork                string
	secondarySubnetworkRange  string
	firewalls                 []string
	name                      string
	project                   string
	uid                       string
	zone                      string
	region                    string
	serviceAccount            string

	// loadbalancer
	loadbalancerIP     string
	loadbalancerIPname string
	loadbalancers      []*loadBalancer
}

// NewFromDefault creates an uninitialized client.
func NewFromDefault(ctx context.Context) (*Client, error) {
	var closers []closer
	insAPI, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		return nil, err
	}
	closers = append(closers, insAPI)
	opZoneAPI, err := compute.NewZoneOperationsRESTClient(ctx)
	if err != nil {
		_ = closeAll(closers)
		return nil, err
	}
	closers = append(closers, opZoneAPI)
	opRegionAPI, err := compute.NewRegionOperationsRESTClient(ctx)
	if err != nil {
		_ = closeAll(closers)
		return nil, err
	}
	closers = append(closers, opRegionAPI)
	opGlobalAPI, err := compute.NewGlobalOperationsRESTClient(ctx)
	if err != nil {
		_ = closeAll(closers)
		return nil, err
	}
	closers = append(closers, opGlobalAPI)
	netAPI, err := compute.NewNetworksRESTClient(ctx)
	if err != nil {
		_ = closeAll(closers)
		return nil, err
	}
	closers = append(closers, netAPI)
	subnetAPI, err := compute.NewSubnetworksRESTClient(ctx)
	if err != nil {
		_ = closeAll(closers)
		return nil, err
	}
	closers = append(closers, subnetAPI)
	fwAPI, err := compute.NewFirewallsRESTClient(ctx)
	if err != nil {
		_ = closeAll(closers)
		return nil, err
	}
	closers = append(closers, subnetAPI)
	forwardingRulesAPI, err := compute.NewForwardingRulesRESTClient(ctx)
	if err != nil {
		_ = closeAll(closers)
		return nil, err
	}
	closers = append(closers, forwardingRulesAPI)
	backendServicesAPI, err := compute.NewRegionBackendServicesRESTClient(ctx)
	if err != nil {
		_ = closeAll(closers)
		return nil, err
	}
	closers = append(closers, backendServicesAPI)
	targetPoolsAPI, err := compute.NewTargetPoolsRESTClient(ctx)
	if err != nil {
		_ = closeAll(closers)
		return nil, err
	}
	closers = append(closers, targetPoolsAPI)
	healthChecksAPI, err := compute.NewRegionHealthChecksRESTClient(ctx)
	if err != nil {
		_ = closeAll(closers)
		return nil, err
	}
	closers = append(closers, healthChecksAPI)
	templAPI, err := compute.NewInstanceTemplatesRESTClient(ctx)
	if err != nil {
		_ = closeAll(closers)
		return nil, err
	}
	closers = append(closers, templAPI)
	groupAPI, err := compute.NewInstanceGroupManagersRESTClient(ctx)
	if err != nil {
		_ = closeAll(closers)
		return nil, err
	}
	closers = append(closers, groupAPI)
	iamAPI, err := admin.NewIamClient(ctx)
	if err != nil {
		_ = closeAll(closers)
		return nil, err
	}
	closers = append(closers, iamAPI)
	projectsAPI, err := resourcemanager.NewProjectsClient(ctx)
	if err != nil {
		_ = closeAll(closers)
		return nil, err
	}
	closers = append(closers, projectsAPI)
	addressesAPI, err := compute.NewAddressesRESTClient(ctx)
	if err != nil {
		_ = closeAll(closers)
		return nil, err
	}

	return &Client{
		instanceAPI:              &instanceClient{insAPI},
		operationRegionAPI:       opRegionAPI,
		operationZoneAPI:         opZoneAPI,
		operationGlobalAPI:       opGlobalAPI,
		networksAPI:              &networksClient{netAPI},
		subnetworksAPI:           &subnetworksClient{subnetAPI},
		firewallsAPI:             &firewallsClient{fwAPI},
		forwardingRulesAPI:       &forwardingRulesClient{forwardingRulesAPI},
		backendServicesAPI:       &backendServicesClient{backendServicesAPI},
		healthChecksAPI:          &healthChecksClient{healthChecksAPI},
		instanceTemplateAPI:      &instanceTemplateClient{templAPI},
		instanceGroupManagersAPI: &instanceGroupManagersClient{groupAPI},
		iamAPI:                   &iamClient{iamAPI},
		projectsAPI:              &projectsClient{projectsAPI},
		addressesAPI:             &addressesClient{addressesAPI},
		workers:                  make(cloudtypes.Instances),
		controlPlanes:            make(cloudtypes.Instances),
	}, nil
}

// NewInitialized creates an initialized client.
func NewInitialized(ctx context.Context, project, zone, region, name string) (*Client, error) {
	// check if ADC are configured for the same project as the cluster
	var defaultProject string
	creds, err := google.FindDefaultCredentials(ctx)
	if err != nil {
		return nil, err
	}
	// if the CLI is run by a service account, use the project of the service account
	defaultProject = creds.ProjectID
	// if the CLI is run by a user directly projectID will be empty, use the quota project id of the user instead
	if defaultProject == "" {
		var projectID struct {
			ProjectID string `json:"quota_project_id"`
		}
		if err := json.Unmarshal(creds.JSON, &projectID); err != nil {
			return nil, err
		}
		defaultProject = projectID.ProjectID
	}
	if defaultProject != project {
		return nil, fmt.Errorf("application default credentials are configured for project %q, but the cluster is configured for project %q", defaultProject, project)
	}

	client, err := NewFromDefault(ctx)
	if err != nil {
		return nil, err
	}
	err = client.init(project, zone, region, name)
	return client, err
}

// Close closes the client's connection.
func (c *Client) Close() error {
	closers := []closer{
		c.instanceAPI,
		c.operationRegionAPI,
		c.operationZoneAPI,
		c.operationGlobalAPI,
		c.networksAPI,
		c.subnetworksAPI,
		c.firewallsAPI,
		c.forwardingRulesAPI,
		c.backendServicesAPI,
		c.healthChecksAPI,
		c.instanceTemplateAPI,
		c.instanceGroupManagersAPI,
		c.iamAPI,
		c.projectsAPI,
		c.addressesAPI,
	}
	return closeAll(closers)
}

// init initializes the client.
func (c *Client) init(project, zone, region, name string) error {
	c.project = project
	c.zone = zone
	c.name = name
	c.region = region

	uid, err := c.generateUID()
	if err != nil {
		return err
	}
	c.uid = uid
	return nil
}

// GetState returns the state of the client as ConstellationState.
func (c *Client) GetState() state.ConstellationState {
	stat := state.ConstellationState{
		Name:                            c.name,
		UID:                             c.uid,
		CloudProvider:                   cloudprovider.GCP.String(),
		LoadBalancerIP:                  c.loadbalancerIP,
		GCPProject:                      c.project,
		GCPZone:                         c.zone,
		GCPRegion:                       c.region,
		GCPWorkerInstances:              c.workers,
		GCPWorkerInstanceGroup:          c.workerInstanceGroup,
		GCPWorkerInstanceTemplate:       c.workerTemplate,
		GCPControlPlaneInstances:        c.controlPlanes,
		GCPControlPlaneInstanceGroup:    c.controlPlaneInstanceGroup,
		GCPControlPlaneInstanceTemplate: c.controlPlaneTemplate,
		GCPFirewalls:                    c.firewalls,
		GCPNetwork:                      c.network,
		GCPSubnetwork:                   c.subnetwork,
		GCPServiceAccount:               c.serviceAccount,
		GCPLoadbalancerIPname:           c.loadbalancerIPname,
	}
	for _, lb := range c.loadbalancers {
		stat.GCPLoadbalancers = append(stat.GCPLoadbalancers, lb.name)
	}
	return stat
}

// SetState sets the state of the client to the handed ConstellationState.
func (c *Client) SetState(stat state.ConstellationState) {
	c.workers = stat.GCPWorkerInstances
	c.controlPlanes = stat.GCPControlPlaneInstances
	c.workerInstanceGroup = stat.GCPWorkerInstanceGroup
	c.controlPlaneInstanceGroup = stat.GCPControlPlaneInstanceGroup
	c.project = stat.GCPProject
	c.zone = stat.GCPZone
	c.region = stat.GCPRegion
	c.name = stat.Name
	c.uid = stat.UID
	c.firewalls = stat.GCPFirewalls
	c.network = stat.GCPNetwork
	c.subnetwork = stat.GCPSubnetwork
	c.workerTemplate = stat.GCPWorkerInstanceTemplate
	c.controlPlaneTemplate = stat.GCPControlPlaneInstanceTemplate
	c.loadbalancerIPname = stat.GCPLoadbalancerIPname
	c.loadbalancerIP = stat.LoadBalancerIP
	c.serviceAccount = stat.GCPServiceAccount
	for _, lbName := range stat.GCPLoadbalancers {
		lb := &loadBalancer{
			name:               lbName,
			hasForwardingRules: true,
			hasBackendService:  true,
			hasHealthCheck:     true,
		}
		c.loadbalancers = append(c.loadbalancers, lb)
	}
}

func (c *Client) generateUID() (string, error) {
	letters := []byte("abcdefghijklmnopqrstuvwxyz0123456789")

	const uidLen = 5
	uid := make([]byte, uidLen)
	for i := 0; i < uidLen; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		uid[i] = letters[n.Int64()]
	}
	return string(uid), nil
}

func (c *Client) resourceURI(scope resourceScope, resourceType, resourceName string) string {
	const baseURI = "https://www.googleapis.com/compute/v1/projects/"

	builder := strings.Builder{}

	builder.WriteString(baseURI)
	builder.WriteString(c.project)

	switch scope {
	case scopeGlobal:
		builder.WriteString("/global/")
	case scopeRegion:
		builder.WriteString("/regions/")
		builder.WriteString(c.region)
		builder.WriteString("/")
	case scopeZone:
		builder.WriteString("/zones/")
		builder.WriteString(c.zone)
		builder.WriteString("/")
	default:
		panic("unknown scope")
	}

	builder.WriteString(resourceType)
	builder.WriteString("/")
	builder.WriteString(resourceName)

	return builder.String()
}

type resourceScope string

const (
	scopeGlobal resourceScope = "global"
	scopeRegion resourceScope = "region"
	scopeZone   resourceScope = "zone"
)

type closer interface {
	Close() error
}

// closeAll closes all closers, even if an error occurs.
//
// Errors are collected and a composed error is returned.
func closeAll(closers []closer) error {
	// Since this function is intended to be deferred, it will always call all
	// close operations, even if a previous operation failed. The if multiple
	// errors occur, the returned error will be composed of the error messages
	// of those errors.
	var err error
	for _, closer := range closers {
		err = multierr.Append(err, closer.Close())
	}
	return err
}

func isNotFoundError(err error) bool {
	var gAPIErr *googleapi.Error
	if !errors.As(err, &gAPIErr) {
		return false
	}
	return gAPIErr.Code == http.StatusNotFound
}
