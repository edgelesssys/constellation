/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package client

import (
	"context"
	"crypto/rand"
	"errors"
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
	targetTCPProxiesAPI
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
	forwardingRulesAPI, err := compute.NewGlobalForwardingRulesRESTClient(ctx)
	if err != nil {
		_ = closeAll(closers)
		return nil, err
	}
	closers = append(closers, forwardingRulesAPI)
	backendServicesAPI, err := compute.NewBackendServicesRESTClient(ctx)
	if err != nil {
		_ = closeAll(closers)
		return nil, err
	}
	closers = append(closers, backendServicesAPI)
	targetTCPProxiesAPI, err := compute.NewTargetTcpProxiesRESTClient(ctx)
	if err != nil {
		_ = closeAll(closers)
		return nil, err
	}
	closers = append(closers, targetTCPProxiesAPI)
	targetPoolsAPI, err := compute.NewTargetPoolsRESTClient(ctx)
	if err != nil {
		_ = closeAll(closers)
		return nil, err
	}
	closers = append(closers, targetPoolsAPI)
	healthChecksAPI, err := compute.NewHealthChecksRESTClient(ctx)
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
	addressesAPI, err := compute.NewGlobalAddressesRESTClient(ctx)
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
		targetTCPProxiesAPI:      &targetTCPProxiesClient{targetTCPProxiesAPI},
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
	for _, lbName := range stat.GCPLoadbalancers {
		lb := &loadBalancer{
			name:               lbName,
			hasForwardingRules: true,
			hasBackendService:  true,
			hasHealthCheck:     true,
			hasTargetTCPProxy:  true,
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

// buildInstanceName returns a formatted name string.
// The names are joined with a '-'.
// If names is empty, the returned value is c.name + "-" + c.uid.
func (c *Client) buildResourceName(names ...string) string {
	builder := strings.Builder{}

	builder.WriteString(c.name)
	builder.WriteRune('-')
	for _, name := range names {
		builder.WriteString(name)
		builder.WriteRune('-')
	}
	builder.WriteString(c.uid)

	return builder.String()
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
		builder.WriteRune('/')
	case scopeZone:
		builder.WriteString("/zones/")
		builder.WriteString(c.zone)
		builder.WriteRune('/')
	default:
		panic("unknown scope")
	}

	builder.WriteString(resourceType)
	builder.WriteRune('/')
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
