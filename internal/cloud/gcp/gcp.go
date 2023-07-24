/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
Implements interaction with the GCP API.

Instance metadata is retrieved from the [GCP metadata API].

Retrieving metadata of other instances is done by using the GCP compute API, and requires GCP credentials.

[GCP metadata API]: https://cloud.google.com/compute/docs/storing-retrieving-metadata
*/
package gcp

import (
	"context"
	"errors"
	"fmt"
	"path"
	"regexp"
	"strings"
	"sync"
	"time"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	imds "cloud.google.com/go/compute/metadata"
	"github.com/edgelesssys/constellation/v2/internal/cloud"
	"github.com/edgelesssys/constellation/v2/internal/cloud/gcpshared"
	"github.com/edgelesssys/constellation/v2/internal/cloud/metadata"
	"github.com/edgelesssys/constellation/v2/internal/role"
	"google.golang.org/api/iterator"
	"google.golang.org/protobuf/proto"
)

const (
	// tagUsage is a label key used to indicate the use of the resource.
	tagUsage = "constellation-use"
	// maxCacheAgeInvariantResource is the maximum age of cached metadata for invariant resources.
	maxCacheAgeInvariantResource = 24 * time.Hour // 1 day
)

var (
	zoneFromRegionRegex = regexp.MustCompile("([a-z]*-[a-z]*[0-9])")
	errNoForwardingRule = errors.New("no forwarding rule found")
)

// Cloud provides GCP cloud metadata information and API access.
type Cloud struct {
	globalForwardingRulesAPI   globalForwardingRulesAPI
	regionalForwardingRulesAPI regionalForwardingRulesAPI
	imds                       imdsAPI
	instanceAPI                instanceAPI
	subnetAPI                  subnetAPI
	zoneAPI                    zoneAPI

	closers []func() error

	// cached metadata
	cacheMux        sync.Mutex
	regionCache     string
	regionCacheTime time.Time
	zoneCache       map[string]struct {
		zones         []string
		zoneCacheTime time.Time
	}
}

// New creates and initializes Cloud.
// The Close method should be called when Cloud is no longer needed.
func New(ctx context.Context) (cloud *Cloud, err error) {
	var closers []func() error

	insAPI, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		return nil, err
	}
	closers = append(closers, insAPI.Close)

	globalForwardingRulesAPI, err := compute.NewGlobalForwardingRulesRESTClient(ctx)
	if err != nil {
		return nil, err
	}
	closers = append(closers, globalForwardingRulesAPI.Close)

	regionalForwardingRulesAPI, err := compute.NewForwardingRulesRESTClient(ctx)
	if err != nil {
		return nil, err
	}
	closers = append(closers, regionalForwardingRulesAPI.Close)

	subnetAPI, err := compute.NewSubnetworksRESTClient(ctx)
	if err != nil {
		return nil, err
	}
	closers = append(closers, subnetAPI.Close)

	zoneAPI, err := compute.NewZonesRESTClient(ctx)
	if err != nil {
		return nil, err
	}
	closers = append(closers, zoneAPI.Close)

	return &Cloud{
		imds:                       imds.NewClient(nil),
		instanceAPI:                &instanceClient{insAPI},
		globalForwardingRulesAPI:   &globalForwardingRulesClient{globalForwardingRulesAPI},
		regionalForwardingRulesAPI: &regionalForwardingRulesClient{regionalForwardingRulesAPI},
		subnetAPI:                  subnetAPI,
		zoneAPI:                    &zoneClient{zoneAPI},
		closers:                    closers,
	}, nil
}

// Close closes all connections to the GCP API server.
func (c *Cloud) Close() {
	for _, close := range c.closers {
		_ = close()
	}
}

// GetLoadBalancerEndpoint returns the endpoint of the load balancer.
func (c *Cloud) GetLoadBalancerEndpoint(ctx context.Context) (host, port string, err error) {
	project, zone, instanceName, err := c.retrieveInstanceInfo()
	if err != nil {
		return "", "", err
	}
	uid, err := c.uid(ctx, project, zone, instanceName)
	if err != nil {
		return "", "", err
	}

	// First try to find a global forwarding rule.
	host, port, err = c.getGlobalForwardingRule(ctx, project, uid)
	if err != nil && !errors.Is(err, errNoForwardingRule) {
		return "", "", fmt.Errorf("getting global forwarding rule: %w", err)
	} else if err == nil {
		return host, port, nil
	}

	// If no global forwarding rule was found, try to find a regional forwarding rule.
	region := zoneFromRegionRegex.FindString(zone)
	if region == "" {
		return "", "", fmt.Errorf("invalid zone %s", zone)
	}

	host, port, err = c.getRegionalForwardingRule(ctx, project, uid, region)
	if err != nil && !errors.Is(err, errNoForwardingRule) {
		return "", "", fmt.Errorf("getting regional forwarding rule: %w", err)
	} else if err != nil {
		return "", "", fmt.Errorf("kubernetes load balancer with UID %s not found: %w", uid, err)
	}

	return host, port, nil
}

// getGlobalForwardingRule returns the endpoint of the load balancer if it is a global load balancer.
// It returns the host, port and optionally an error.
// This functions returns ErrNoForwardingRule if no forwarding rule was found.
func (c *Cloud) getGlobalForwardingRule(ctx context.Context, project, uid string) (string, string, error) {
	var resp *computepb.ForwardingRule
	var err error
	iter := c.globalForwardingRulesAPI.List(ctx, &computepb.ListGlobalForwardingRulesRequest{
		Project: project,
		Filter:  proto.String(fmt.Sprintf("(labels.%s:%s) AND (labels.%s:kubernetes)", cloud.TagUID, uid, tagUsage)),
	})
	for resp, err = iter.Next(); err == nil; resp, err = iter.Next() {
		if resp.PortRange == nil {
			continue
		}
		if resp.IPAddress == nil {
			continue
		}
		portRange := strings.Split(*resp.PortRange, "-")
		return *resp.IPAddress, portRange[0], nil
	}
	if err != iterator.Done {
		return "", "", fmt.Errorf("error listing global forwarding rules with UID %s: %w", uid, err)
	}
	return "", "", errNoForwardingRule
}

// getRegionalForwardingRule returns the endpoint of the load balancer if it is a regional load balancer.
// It returns the host, port and optionally an error.
// This functions returns ErrNoForwardingRule if no forwarding rule was found.
func (c *Cloud) getRegionalForwardingRule(ctx context.Context, project, uid, region string) (host string, port string, err error) {
	var resp *computepb.ForwardingRule
	iter := c.regionalForwardingRulesAPI.List(ctx, &computepb.ListForwardingRulesRequest{
		Project: project,
		Region:  region,
		Filter:  proto.String(fmt.Sprintf("(labels.%s:%s) AND (labels.%s:kubernetes)", cloud.TagUID, uid, tagUsage)),
	})
	for resp, err = iter.Next(); err == nil; resp, err = iter.Next() {
		if resp.PortRange == nil {
			continue
		}
		if resp.IPAddress == nil {
			continue
		}
		if resp.Region == nil {
			continue
		}
		portRange := strings.Split(*resp.PortRange, "-")
		return *resp.IPAddress, portRange[0], nil
	}
	if err != iterator.Done {
		return "", "", fmt.Errorf("error listing global forwarding rules with UID %s: %w", uid, err)
	}
	return "", "", errNoForwardingRule
}

// List retrieves all instances belonging to the current constellation.
// On GCP, this is done by listing all instances in a region by requesting all instances in each zone.
func (c *Cloud) List(ctx context.Context) ([]metadata.InstanceMetadata, error) {
	project, zone, instanceName, err := c.retrieveInstanceInfo()
	if err != nil {
		return nil, err
	}
	uid, err := c.uid(ctx, project, zone, instanceName)
	if err != nil {
		return nil, err
	}

	region, err := c.region()
	if err != nil {
		return nil, fmt.Errorf("getting region: %w", err)
	}

	zones, err := c.zones(ctx, project, region)
	if err != nil {
		return nil, fmt.Errorf("getting zones: %w", err)
	}

	var instances []metadata.InstanceMetadata
	for _, zone := range zones {
		zoneInstances, err := c.listInZone(ctx, project, zone, uid)
		if err != nil {
			return nil, fmt.Errorf("listing instances in zone %s: %w", zone, err)
		}
		instances = append(instances, zoneInstances...)
	}
	return instances, nil
}

// listInZone retrieves all instances belonging to the current constellation in a given zone.
func (c *Cloud) listInZone(ctx context.Context, project, zone, uid string) ([]metadata.InstanceMetadata, error) {
	var instances []metadata.InstanceMetadata
	var resp *computepb.Instance
	var err error
	iter := c.instanceAPI.List(ctx, &computepb.ListInstancesRequest{
		Filter:  proto.String(fmt.Sprintf("labels.%s:%s", cloud.TagUID, uid)),
		Project: project,
		Zone:    zone,
	})
	for resp, err = iter.Next(); err == nil; resp, err = iter.Next() {
		instance, err := convertToInstanceMetadata(resp, project, zone)
		if err != nil {
			return nil, fmt.Errorf("retrieving instance list from GCP: failed to convert instance: %w", err)
		}

		// convertToInstanceMetadata already checks for nil resp.NetworkInterfaces
		if len(resp.NetworkInterfaces) == 0 || resp.NetworkInterfaces[0] == nil ||
			resp.NetworkInterfaces[0].Subnetwork == nil {
			return nil, errors.New("retrieving compute instance: received invalid instance")
		}

		subnetCIDR, err := c.retrieveSubnetworkAliasCIDR(ctx, project, zone, *resp.NetworkInterfaces[0].Subnetwork)
		if err != nil {
			return nil, fmt.Errorf("retrieving compute instance: failed to retrieve subnet CIDR: %w", err)
		}
		instance.SecondaryIPRange = subnetCIDR

		instances = append(instances, instance)
	}
	if errors.Is(err, iterator.Done) {
		return instances, nil
	}
	return nil, fmt.Errorf("retrieving instance list from GCP: %w", err)
}

// ProviderID returns the providerID of the current instance.
func (c *Cloud) ProviderID(_ context.Context) (string, error) {
	project, zone, instanceName, err := c.retrieveInstanceInfo()
	if err != nil {
		return "", err
	}
	return gcpshared.JoinProviderID(project, zone, instanceName), nil
}

// Self retrieves the current instance.
func (c *Cloud) Self(ctx context.Context) (metadata.InstanceMetadata, error) {
	project, zone, instanceName, err := c.retrieveInstanceInfo()
	if err != nil {
		return metadata.InstanceMetadata{}, err
	}
	return c.getInstance(ctx, project, zone, instanceName)
}

// UID retrieves the UID of the constellation.
func (c *Cloud) UID(ctx context.Context) (string, error) {
	project, zone, instanceName, err := c.retrieveInstanceInfo()
	if err != nil {
		return "", err
	}
	return c.uid(ctx, project, zone, instanceName)
}

// InitSecretHash retrieves the InitSecretHash of the current instance.
func (c *Cloud) InitSecretHash(ctx context.Context) ([]byte, error) {
	project, zone, instanceName, err := c.retrieveInstanceInfo()
	if err != nil {
		return nil, err
	}
	initSecretHash, err := c.initSecretHash(ctx, project, zone, instanceName)
	if err != nil {
		return nil, fmt.Errorf("retrieving init secret hash: %w", err)
	}
	return []byte(initSecretHash), nil
}

// getInstance retrieves an instance using its project, zone and name, and parses it to metadata.InstanceMetadata.
func (c *Cloud) getInstance(ctx context.Context, project, zone, instanceName string) (metadata.InstanceMetadata, error) {
	gcpInstance, err := c.instanceAPI.Get(ctx, &computepb.GetInstanceRequest{
		Project:  project,
		Zone:     zone,
		Instance: instanceName,
	})
	if err != nil {
		return metadata.InstanceMetadata{}, fmt.Errorf("retrieving compute instance: %w", err)
	}

	if gcpInstance == nil || gcpInstance.NetworkInterfaces == nil || len(gcpInstance.NetworkInterfaces) == 0 ||
		gcpInstance.NetworkInterfaces[0] == nil || gcpInstance.NetworkInterfaces[0].Subnetwork == nil {
		return metadata.InstanceMetadata{}, errors.New("retrieving compute instance: received invalid instance")
	}
	subnetCIDR, err := c.retrieveSubnetworkAliasCIDR(ctx, project, zone, *gcpInstance.NetworkInterfaces[0].Subnetwork)
	if err != nil {
		return metadata.InstanceMetadata{}, err
	}

	instance, err := convertToInstanceMetadata(gcpInstance, project, zone)
	if err != nil {
		return metadata.InstanceMetadata{}, fmt.Errorf("converting instance: %w", err)
	}

	instance.SecondaryIPRange = subnetCIDR

	return instance, nil
}

// retrieveInstanceInfo retrieves the project, zone and instance name of the current instance using the imds API.
func (c *Cloud) retrieveInstanceInfo() (project, zone, instanceName string, err error) {
	project, err = c.imds.ProjectID()
	if err != nil {
		return "", "", "", fmt.Errorf("retrieving project ID from imds: %w", err)
	}
	zone, err = c.imds.Zone()
	if err != nil {
		return "", "", "", fmt.Errorf("retrieving zone from imds: %w", err)
	}
	instanceName, err = c.imds.InstanceName()
	if err != nil {
		return "", "", "", fmt.Errorf("retrieving instance name from imds: %w", err)
	}
	return project, zone, instanceName, nil
}

// retrieveSubnetworkAliasCIDR retrieves the secondary IP range CIDR of the subnetwork,
// identified by project, zone and subnetworkURI.
func (c *Cloud) retrieveSubnetworkAliasCIDR(ctx context.Context, project, zone, subnetworkURI string) (string, error) {
	// convert:
	//           zone --> region
	// europe-west3-b --> europe-west3
	region := zoneFromRegionRegex.FindString(zone)
	if region == "" {
		return "", fmt.Errorf("invalid zone %s", zone)
	}

	req := &computepb.GetSubnetworkRequest{
		Project:    project,
		Region:     region,
		Subnetwork: path.Base(subnetworkURI),
	}
	subnetwork, err := c.subnetAPI.Get(ctx, req)
	if err != nil {
		return "", fmt.Errorf("retrieving subnetwork alias CIDR failed: %w", err)
	}
	if subnetwork == nil || len(subnetwork.SecondaryIpRanges) == 0 ||
		subnetwork.SecondaryIpRanges[0] == nil || subnetwork.SecondaryIpRanges[0].IpCidrRange == nil {
		return "", fmt.Errorf("retrieving subnetwork alias CIDR failed: received invalid subnetwork")
	}

	return *subnetwork.SecondaryIpRanges[0].IpCidrRange, nil
}

// uid retrieves the UID of the instance identified by project, zone and instanceName.
// The UID is retrieved from the instance's labels.
func (c *Cloud) uid(ctx context.Context, project, zone, instanceName string) (string, error) {
	instance, err := c.instanceAPI.Get(ctx, &computepb.GetInstanceRequest{
		Project:  project,
		Zone:     zone,
		Instance: instanceName,
	})
	if err != nil {
		return "", fmt.Errorf("retrieving compute instance: %w", err)
	}
	if instance == nil || instance.Labels == nil {
		return "", errors.New("retrieving compute instance: received instance with invalid labels")
	}
	uid, ok := instance.Labels[cloud.TagUID]
	if !ok {
		return "", errors.New("retrieving compute instance: received instance with no UID label")
	}
	return uid, nil
}

// initSecretHash retrieves the init secret hash of the instance identified by project, zone and instanceName.
// The init secret hash is retrieved from the instance's labels.
func (c *Cloud) initSecretHash(ctx context.Context, project, zone, instanceName string) (string, error) {
	instance, err := c.instanceAPI.Get(ctx, &computepb.GetInstanceRequest{
		Project:  project,
		Zone:     zone,
		Instance: instanceName,
	})
	if err != nil {
		return "", fmt.Errorf("retrieving compute instance: %w", err)
	}
	if instance == nil || instance.Metadata == nil {
		return "", errors.New("retrieving compute instance: received instance with invalid metadata")
	}
	if len(instance.Metadata.Items) == 0 {
		return "", errors.New("retrieving compute instance: received instance with empty metadata")
	}
	for _, item := range instance.Metadata.Items {
		if item == nil || item.Key == nil || item.Value == nil {
			return "", errors.New("retrieving compute instance: received instance with invalid metadata item")
		}
		if *item.Key == cloud.TagInitSecretHash {
			return *item.Value, nil
		}
	}
	return "", errors.New("retrieving compute instance: received instance with no init secret hash label")
}

// region retrieves the region that this instance is located in.
func (c *Cloud) region() (string, error) {
	c.cacheMux.Lock()
	defer c.cacheMux.Unlock()
	// try to retrieve from cache first
	if c.regionCache != "" &&
		time.Since(c.regionCacheTime) < maxCacheAgeInvariantResource {
		return c.regionCache, nil
	}
	zone, err := c.imds.Zone()
	if err != nil {
		return "", fmt.Errorf("retrieving zone from imds: %w", err)
	}
	region, err := regionFromZone(zone)
	if err != nil {
		return "", fmt.Errorf("retrieving region from zone: %w", err)
	}
	c.regionCache = region
	c.regionCacheTime = time.Now()
	return region, nil
}

// zones retrieves all zones that are within a region.
func (c *Cloud) zones(ctx context.Context, project, region string) ([]string, error) {
	c.cacheMux.Lock()
	defer c.cacheMux.Unlock()
	// try to retrieve from cache first
	if cachedZones, ok := c.zoneCache[region]; ok &&
		time.Since(cachedZones.zoneCacheTime) < maxCacheAgeInvariantResource {
		return cachedZones.zones, nil
	}
	req := &computepb.ListZonesRequest{
		Project: project,
		Filter:  proto.String(fmt.Sprintf("name = \"%s*\"", region)),
	}
	zonesIter := c.zoneAPI.List(ctx, req)
	var zones []string
	var resp *computepb.Zone
	var err error
	for resp, err = zonesIter.Next(); err == nil; resp, err = zonesIter.Next() {
		if resp == nil || resp.Name == nil {
			continue
		}
		zones = append(zones, *resp.Name)
	}
	if err != nil && err != iterator.Done {
		return nil, fmt.Errorf("listing zones: %w", err)
	}
	if c.zoneCache == nil {
		c.zoneCache = make(map[string]struct {
			zones         []string
			zoneCacheTime time.Time
		})
	}
	c.zoneCache[region] = struct {
		zones         []string
		zoneCacheTime time.Time
	}{
		zones:         zones,
		zoneCacheTime: time.Now(),
	}
	return zones, nil
}

// convertToInstanceMetadata converts a *computepb.Instance to a metadata.InstanceMetadata.
func convertToInstanceMetadata(in *computepb.Instance, project string, zone string) (metadata.InstanceMetadata, error) {
	if in.Name == nil {
		return metadata.InstanceMetadata{}, fmt.Errorf("missing instance name")
	}

	var vpcIP string
	var ips []string
	for _, interf := range in.NetworkInterfaces {
		if interf == nil {
			continue
		}

		// use private IP from the default interface
		if interf.NetworkIP != nil && interf.Name != nil && *interf.Name == "nic0" {
			vpcIP = *interf.NetworkIP
		}

		if interf.AliasIpRanges == nil {
			continue
		}
		for _, aliasIP := range interf.AliasIpRanges {
			if aliasIP != nil && aliasIP.IpCidrRange != nil {
				ips = append(ips, *aliasIP.IpCidrRange)
			}
		}
	}

	return metadata.InstanceMetadata{
		Name:          *in.Name,
		ProviderID:    gcpshared.JoinProviderID(project, zone, *in.Name),
		Role:          role.FromString(in.Labels[cloud.TagRole]),
		VPCIP:         vpcIP,
		AliasIPRanges: ips,
	}, nil
}

func regionFromZone(zone string) (string, error) {
	zoneParts := strings.Split(zone, "-")
	if len(zoneParts) != 3 {
		return "", fmt.Errorf("invalid zone format: %s", zone)
	}
	return fmt.Sprintf("%s-%s", zoneParts[0], zoneParts[1]), nil
}
