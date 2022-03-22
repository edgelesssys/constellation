package client

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"

	compute "cloud.google.com/go/compute/apiv1"
	admin "cloud.google.com/go/iam/admin/apiv1"
	resourcemanager "cloud.google.com/go/resourcemanager/apiv3"
	"github.com/edgelesssys/constellation/cli/cloudprovider"
	"github.com/edgelesssys/constellation/cli/gcp"
	"github.com/edgelesssys/constellation/internal/state"
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
	instanceTemplateAPI
	instanceGroupManagersAPI
	iamAPI
	projectsAPI

	nodes        gcp.Instances
	coordinators gcp.Instances

	nodesInstanceGroup       string
	coordinatorInstanceGroup string
	coordinatorTemplate      string
	nodeTemplate             string
	network                  string
	subnetwork               string
	secondarySubnetworkRange string
	firewalls                []string
	name                     string
	project                  string
	uid                      string
	zone                     string
	region                   string
	serviceAccount           string
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
	closers = append(closers, fwAPI)
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
		instanceTemplateAPI:      &instanceTemplateClient{templAPI},
		instanceGroupManagersAPI: &instanceGroupManagersClient{groupAPI},
		iamAPI:                   &iamClient{iamAPI},
		projectsAPI:              &projectsClient{projectsAPI},
		nodes:                    make(gcp.Instances),
		coordinators:             make(gcp.Instances),
	}, nil
}

// NewInitialized creates an initialized client.
func NewInitialized(ctx context.Context, project, zone, region, name string) (*Client, error) {
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
		c.operationZoneAPI,
		c.operationGlobalAPI,
		c.networksAPI,
		c.firewallsAPI,
		c.instanceTemplateAPI,
		c.instanceGroupManagersAPI,
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
	if len(c.nodes) == 0 {
		return state.ConstellationState{}, errors.New("client has no nodes")
	}
	stat.GCPNodes = c.nodes

	if len(c.coordinators) == 0 {
		return state.ConstellationState{}, errors.New("client has no coordinators")
	}
	stat.GCPCoordinators = c.coordinators

	if c.nodesInstanceGroup == "" {
		return state.ConstellationState{}, errors.New("client has no nodeInstanceGroup")
	}
	stat.GCPNodeInstanceGroup = c.nodesInstanceGroup

	if c.coordinatorInstanceGroup == "" {
		return state.ConstellationState{}, errors.New("client has no coordinatorInstanceGroup")
	}
	stat.GCPCoordinatorInstanceGroup = c.coordinatorInstanceGroup

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

	if c.nodeTemplate == "" {
		return state.ConstellationState{}, errors.New("client has no node instance template")
	}
	stat.GCPNodeInstanceTemplate = c.nodeTemplate

	if c.coordinatorTemplate == "" {
		return state.ConstellationState{}, errors.New("client has no coordinator instance template")
	}
	stat.GCPCoordinatorInstanceTemplate = c.coordinatorTemplate

	// service account does not have to be set at all times
	stat.GCPServiceAccount = c.serviceAccount

	return stat, nil
}

// SetState sets the state of the client to the handed ConstellationState.
func (c *Client) SetState(stat state.ConstellationState) error {
	if stat.CloudProvider != cloudprovider.GCP.String() {
		return errors.New("state is not gcp state")
	}
	if len(stat.GCPNodes) == 0 {
		return errors.New("state has no nodes")
	}
	c.nodes = stat.GCPNodes

	if len(stat.GCPCoordinators) == 0 {
		return errors.New("state has no coordinator")
	}
	c.coordinators = stat.GCPCoordinators

	if stat.GCPNodeInstanceGroup == "" {
		return errors.New("state has no nodeInstanceGroup")
	}
	c.nodesInstanceGroup = stat.GCPNodeInstanceGroup

	if stat.GCPCoordinatorInstanceGroup == "" {
		return errors.New("state has no coordinatorInstanceGroup")
	}
	c.coordinatorInstanceGroup = stat.GCPCoordinatorInstanceGroup

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

	if stat.GCPNodeInstanceTemplate == "" {
		return errors.New("state has no node instance template")
	}
	c.nodeTemplate = stat.GCPNodeInstanceTemplate

	if stat.GCPCoordinatorInstanceTemplate == "" {
		return errors.New("state has no coordinator instance template")
	}
	c.coordinatorTemplate = stat.GCPCoordinatorInstanceTemplate

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
