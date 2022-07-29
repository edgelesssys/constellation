package client

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"

	compute "cloud.google.com/go/compute/apiv1"
	admin "cloud.google.com/go/iam/admin/apiv1"
	resourcemanager "cloud.google.com/go/resourcemanager/apiv3"
	"github.com/edgelesssys/constellation/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/internal/cloud/cloudtypes"
	"github.com/edgelesssys/constellation/internal/state"
	"golang.org/x/oauth2/google"
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
	healthCheck    string
	backendService string
	forwardingRule string
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
func (c *Client) GetState() (state.ConstellationState, error) {
	var stat state.ConstellationState
	stat.CloudProvider = cloudprovider.GCP.String()
	if len(c.workers) == 0 {
		return state.ConstellationState{}, errors.New("client has no workers")
	}
	stat.GCPWorkerInstances = c.workers

	if len(c.controlPlanes) == 0 {
		return state.ConstellationState{}, errors.New("client has no controlPlanes")
	}
	stat.GCPControlPlaneInstances = c.controlPlanes
	publicIPs := c.controlPlanes.PublicIPs()
	if len(publicIPs) == 0 {
		return state.ConstellationState{}, errors.New("client has no bootstrapper endpoint")
	}
	stat.BootstrapperHost = publicIPs[0]

	if c.workerInstanceGroup == "" {
		return state.ConstellationState{}, errors.New("client has no workerInstanceGroup")
	}
	stat.GCPWorkerInstanceGroup = c.workerInstanceGroup

	if c.controlPlaneInstanceGroup == "" {
		return state.ConstellationState{}, errors.New("client has no controlPlaneInstanceGroup")
	}
	stat.GCPControlPlaneInstanceGroup = c.controlPlaneInstanceGroup

	if c.project == "" {
		return state.ConstellationState{}, errors.New("client has no project")
	}
	stat.GCPProject = c.project

	if c.zone == "" {
		return state.ConstellationState{}, errors.New("client has no zone")
	}
	stat.GCPZone = c.zone

	if c.region == "" {
		return state.ConstellationState{}, errors.New("client has no region")
	}
	stat.GCPRegion = c.region

	if c.name == "" {
		return state.ConstellationState{}, errors.New("client has no name")
	}
	stat.Name = c.name

	if c.uid == "" {
		return state.ConstellationState{}, errors.New("client has no uid")
	}
	stat.UID = c.uid

	if len(c.firewalls) == 0 {
		return state.ConstellationState{}, errors.New("client has no firewalls")
	}
	stat.GCPFirewalls = c.firewalls

	if c.network == "" {
		return state.ConstellationState{}, errors.New("client has no network")
	}
	stat.GCPNetwork = c.network

	if c.subnetwork == "" {
		return state.ConstellationState{}, errors.New("client has no subnetwork")
	}
	stat.GCPSubnetwork = c.subnetwork

	if c.workerTemplate == "" {
		return state.ConstellationState{}, errors.New("client has no worker instance template")
	}
	stat.GCPWorkerInstanceTemplate = c.workerTemplate

	if c.controlPlaneTemplate == "" {
		return state.ConstellationState{}, errors.New("client has no controlPlane instance template")
	}
	stat.GCPControlPlaneInstanceTemplate = c.controlPlaneTemplate

	if c.healthCheck == "" {
		return state.ConstellationState{}, errors.New("client has no health check")
	}
	stat.GCPHealthCheck = c.healthCheck

	if c.backendService == "" {
		return state.ConstellationState{}, errors.New("client has no backend service")
	}
	stat.GCPBackendService = c.backendService

	if c.forwardingRule == "" {
		return state.ConstellationState{}, errors.New("client has no forwarding rule")
	}
	stat.GCPForwardingRule = c.forwardingRule

	// service account does not have to be set at all times
	stat.GCPServiceAccount = c.serviceAccount

	return stat, nil
}

// SetState sets the state of the client to the handed ConstellationState.
func (c *Client) SetState(stat state.ConstellationState) error {
	if stat.CloudProvider != cloudprovider.GCP.String() {
		return errors.New("state is not gcp state")
	}
	if len(stat.GCPWorkerInstances) == 0 {
		return errors.New("state has no workers")
	}
	c.workers = stat.GCPWorkerInstances

	if len(stat.GCPControlPlaneInstances) == 0 {
		return errors.New("state has no controlPlane")
	}
	c.controlPlanes = stat.GCPControlPlaneInstances

	if stat.GCPWorkerInstanceGroup == "" {
		return errors.New("state has no workerInstanceGroup")
	}
	c.workerInstanceGroup = stat.GCPWorkerInstanceGroup

	if stat.GCPControlPlaneInstanceGroup == "" {
		return errors.New("state has no controlPlaneInstanceGroup")
	}
	c.controlPlaneInstanceGroup = stat.GCPControlPlaneInstanceGroup

	if stat.GCPProject == "" {
		return errors.New("state has no project")
	}
	c.project = stat.GCPProject

	if stat.GCPZone == "" {
		return errors.New("state has no zone")
	}
	c.zone = stat.GCPZone

	if stat.GCPRegion == "" {
		return errors.New("state has no region")
	}
	c.region = stat.GCPRegion

	if stat.Name == "" {
		return errors.New("state has no name")
	}
	c.name = stat.Name

	if stat.UID == "" {
		return errors.New("state has no uid")
	}
	c.uid = stat.UID

	if len(stat.GCPFirewalls) == 0 {
		return errors.New("state has no firewalls")
	}
	c.firewalls = stat.GCPFirewalls

	if stat.GCPNetwork == "" {
		return errors.New("state has no network")
	}
	c.network = stat.GCPNetwork

	if stat.GCPSubnetwork == "" {
		return errors.New("state has no subnetwork")
	}
	c.subnetwork = stat.GCPSubnetwork

	if stat.GCPWorkerInstanceTemplate == "" {
		return errors.New("state has no worker instance template")
	}
	c.workerTemplate = stat.GCPWorkerInstanceTemplate

	if stat.GCPControlPlaneInstanceTemplate == "" {
		return errors.New("state has no controlPlane instance template")
	}
	c.controlPlaneTemplate = stat.GCPControlPlaneInstanceTemplate

	if stat.GCPHealthCheck == "" {
		return errors.New("state has no health check")
	}
	c.healthCheck = stat.GCPHealthCheck

	if stat.GCPBackendService == "" {
		return errors.New("state has no backend service")
	}
	c.backendService = stat.GCPBackendService

	if stat.GCPForwardingRule == "" {
		return errors.New("state has no forwarding rule")
	}
	c.forwardingRule = stat.GCPForwardingRule

	// service account does not have to be set at all times
	c.serviceAccount = stat.GCPServiceAccount

	return nil
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
	var errs []error
	for _, closer := range closers {
		errs = append(errs, closer.Close())
	}
	return composeErr(errs)
}

// composeErr composes a list of errors to a single error.
//
// If all errs are nil, the returned error is also nil.
func composeErr(errs []error) error {
	var composed strings.Builder
	for i, err := range errs {
		if err != nil {
			composed.WriteString(fmt.Sprintf("%d: %s", i, err.Error()))
		}
	}
	if composed.Len() != 0 {
		return errors.New(composed.String())
	}
	return nil
}
