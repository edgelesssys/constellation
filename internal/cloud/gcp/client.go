/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package gcp

import (
	"context"
	"errors"
	"fmt"
	"net"
	"regexp"
	"strings"

	compute "cloud.google.com/go/compute/apiv1"
	"github.com/edgelesssys/constellation/v2/internal/cloud/metadata"
	"github.com/edgelesssys/constellation/v2/internal/gcpshared"
	"google.golang.org/api/iterator"
	computepb "google.golang.org/genproto/googleapis/cloud/compute/v1"
	"google.golang.org/protobuf/proto"
)

const (
	gcpSSHMetadataKey           = "ssh-keys"
	constellationUIDMetadataKey = "constellation-uid"
)

var zoneFromRegionRegex = regexp.MustCompile("([a-z]*-[a-z]*[0-9])")

// Client implements the gcp.API interface.
type Client struct {
	instanceAPI
	subnetworkAPI
	metadataAPI
	forwardingRulesAPI
}

// NewClient creates a new Client.
func NewClient(ctx context.Context) (*Client, error) {
	insAPI, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		return nil, err
	}
	subnetAPI, err := compute.NewSubnetworksRESTClient(ctx)
	if err != nil {
		return nil, err
	}
	forwardingRulesAPI, err := compute.NewGlobalForwardingRulesRESTClient(ctx)
	if err != nil {
		return nil, err
	}
	return &Client{
		instanceAPI:        &instanceClient{insAPI},
		subnetworkAPI:      &subnetworkClient{subnetAPI},
		forwardingRulesAPI: &forwardingRulesClient{forwardingRulesAPI},
		metadataAPI:        &metadataClient{},
	}, nil
}

// RetrieveInstances returns list of instances including their ips and metadata.
func (c *Client) RetrieveInstances(ctx context.Context, project string) ([]metadata.InstanceMetadata, error) {
	uid, err := c.UID()
	if err != nil {
		return nil, err
	}
	req := &computepb.AggregatedListInstancesRequest{
		Project: project,
		Filter:  proto.String(fmt.Sprintf("name eq \".+-%s-.+\"", uid)), // filter by constellation UID
	}
	listPairIterator := c.instanceAPI.AggregatedList(ctx, req)

	instancesMetadataList := []metadata.InstanceMetadata{}
	for {
		resp, err := listPairIterator.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("retrieving instances scoped list pair from compute API client: %w", err)
		}
		if resp.Value == nil {
			return nil, fmt.Errorf("retrieving instances scoped list pair from compute API client returned invalid results")
		}

		for _, instance := range resp.Value.Instances {
			instanceMetadata, err := convertToMetadataInstance(instance, project)
			if err != nil {
				return nil, fmt.Errorf("extracting metadata of instance: %w", err)
			}
			instancesMetadataList = append(instancesMetadataList, instanceMetadata)
		}
	}
	return instancesMetadataList, nil
}

// RetrieveInstance returns a an instance including ips and metadata.
func (c *Client) RetrieveInstance(ctx context.Context, project, zone, instanceName string) (metadata.InstanceMetadata, error) {
	instance, err := c.getComputeInstance(ctx, project, zone, instanceName)
	if err != nil {
		return metadata.InstanceMetadata{}, err
	}

	return convertToMetadataInstance(instance, project)
}

// RetrieveProjectID retrieves the GCP projectID containing the current instance.
func (c *Client) RetrieveProjectID() (string, error) {
	value, err := c.metadataAPI.ProjectID()
	if err != nil {
		return "", fmt.Errorf("requesting GCP projectID failed %w", err)
	}
	return value, nil
}

// RetrieveZone retrieves the GCP zone containing the current instance.
func (c *Client) RetrieveZone() (string, error) {
	value, err := c.metadataAPI.Zone()
	if err != nil {
		return "", fmt.Errorf("requesting GCP zone failed %w", err)
	}
	return value, nil
}

func (c *Client) RetrieveInstanceName() (string, error) {
	value, err := c.metadataAPI.InstanceName()
	if err != nil {
		return "", fmt.Errorf("requesting GCP instanceName failed %w", err)
	}
	return value, nil
}

func (c *Client) RetrieveInstanceMetadata(attr string) (string, error) {
	value, err := c.metadataAPI.InstanceAttributeValue(attr)
	if err != nil {
		return "", fmt.Errorf("requesting GCP instance metadata: %w", err)
	}
	return value, nil
}

// SetInstanceMetadata modifies a key value pair of metadata for the instance specified by project, zone and instanceName.
func (c *Client) SetInstanceMetadata(ctx context.Context, project, zone, instanceName, key, value string) error {
	instance, err := c.getComputeInstance(ctx, project, zone, instanceName)
	if err != nil {
		return fmt.Errorf("retrieving instance metadata: %w", err)
	}
	if instance == nil || instance.Metadata == nil {
		return fmt.Errorf("retrieving instance metadata returned invalid results")
	}

	// convert instance metadata to map to handle duplicate keys correctly
	metadataMap := extractInstanceMetadata(instance.Metadata, key, false)
	metadataMap[key] = value
	// convert instance metadata back to flat list
	metadata := flattenInstanceMetadata(metadataMap, instance.Metadata.Fingerprint, instance.Metadata.Kind)

	if err := c.updateInstanceMetadata(ctx, project, zone, instanceName, metadata); err != nil {
		return fmt.Errorf("setting instance metadata %v: %v: %w", key, value, err)
	}
	return nil
}

// UnsetInstanceMetadata modifies a key value pair of metadata for the instance specified by project, zone and instanceName.
func (c *Client) UnsetInstanceMetadata(ctx context.Context, project, zone, instanceName, key string) error {
	instance, err := c.getComputeInstance(ctx, project, zone, instanceName)
	if err != nil {
		return fmt.Errorf("retrieving instance metadata: %w", err)
	}
	if instance == nil || instance.Metadata == nil {
		return fmt.Errorf("retrieving instance metadata returned invalid results")
	}

	// convert instance metadata to map to handle duplicate keys correctly
	// and skip the key to be removed
	metadataMap := extractInstanceMetadata(instance.Metadata, key, true)

	// convert instance metadata back to flat list
	metadata := flattenInstanceMetadata(metadataMap, instance.Metadata.Fingerprint, instance.Metadata.Kind)

	if err := c.updateInstanceMetadata(ctx, project, zone, instanceName, metadata); err != nil {
		return fmt.Errorf("unsetting instance metadata key %v: %w", key, err)
	}
	return nil
}

// RetrieveSubnetworkAliasCIDR returns the alias CIDR of the subnetwork specified by project, zone and subnetworkName.
func (c *Client) RetrieveSubnetworkAliasCIDR(ctx context.Context, project, zone, instanceName string) (string, error) {
	instance, err := c.getComputeInstance(ctx, project, zone, instanceName)
	if err != nil {
		return "", err
	}
	if instance == nil || instance.NetworkInterfaces == nil || len(instance.NetworkInterfaces) == 0 || instance.NetworkInterfaces[0].Subnetwork == nil {
		return "", fmt.Errorf("retrieving instance network interfaces failed")
	}
	subnetworkURL := *instance.NetworkInterfaces[0].Subnetwork
	subnetworkURLFragments := strings.Split(subnetworkURL, "/")
	subnetworkName := subnetworkURLFragments[len(subnetworkURLFragments)-1]

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
		Subnetwork: subnetworkName,
	}
	subnetwork, err := c.subnetworkAPI.Get(ctx, req)
	if err != nil {
		return "", fmt.Errorf("retrieving subnetwork alias CIDR failed: %w", err)
	}
	if subnetwork == nil || len(subnetwork.SecondaryIpRanges) == 0 || (subnetwork.SecondaryIpRanges[0]).IpCidrRange == nil {
		return "", fmt.Errorf("retrieving subnetwork alias CIDR returned invalid results")
	}

	return *(subnetwork.SecondaryIpRanges[0]).IpCidrRange, nil
}

// RetrieveLoadBalancerEndpoint returns the endpoint of the load balancer with the constellation-uid tag.
func (c *Client) RetrieveLoadBalancerEndpoint(ctx context.Context, project string) (string, error) {
	uid, err := c.UID()
	if err != nil {
		return "", err
	}

	req := &computepb.ListGlobalForwardingRulesRequest{
		Project: project,
	}
	iter := c.forwardingRulesAPI.List(ctx, req)
	for {
		resp, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return "", fmt.Errorf("retrieving load balancer IP failed: %w", err)
		}
		if resp.Labels["constellation-uid"] == uid {
			if resp.PortRange == nil {
				return "", errors.New("load balancer with searched UID has no ports")
			}
			portRange := strings.Split(*resp.PortRange, "-")
			return net.JoinHostPort(*resp.IPAddress, portRange[0]), nil
		}
	}

	return "", fmt.Errorf("retrieving load balancer IP failed: load balancer not found")
}

// Close closes the instanceAPI client.
func (c *Client) Close() error {
	if err := c.subnetworkAPI.Close(); err != nil {
		return err
	}
	if err := c.forwardingRulesAPI.Close(); err != nil {
		return err
	}
	return c.instanceAPI.Close()
}

func (c *Client) getComputeInstance(ctx context.Context, project, zone, instanceName string) (*computepb.Instance, error) {
	instanceGetReq := &computepb.GetInstanceRequest{
		Project:  project,
		Zone:     zone,
		Instance: instanceName,
	}
	instance, err := c.instanceAPI.Get(ctx, instanceGetReq)
	if err != nil {
		return nil, fmt.Errorf("retrieving compute instance: %w", err)
	}
	return instance, nil
}

// updateInstanceMetadata updates all instance metadata key-value pairs.
func (c *Client) updateInstanceMetadata(ctx context.Context, project, zone, instanceName string, metadata *computepb.Metadata) error {
	setMetadataReq := &computepb.SetMetadataInstanceRequest{
		Project:          project,
		Zone:             zone,
		Instance:         instanceName,
		MetadataResource: metadata,
	}

	if _, err := c.instanceAPI.SetMetadata(ctx, setMetadataReq); err != nil {
		return fmt.Errorf("updating instance metadata: %w", err)
	}
	return nil
}

// UID retrieves the current instances uid.
func (c *Client) UID() (string, error) {
	// API endpoint: http://metadata.google.internal/computeMetadata/v1/instance/attributes/constellation-uid
	uid, err := c.RetrieveInstanceMetadata(constellationUIDMetadataKey)
	if err != nil {
		return "", fmt.Errorf("retrieving constellation uid: %w", err)
	}
	return uid, nil
}

// extractVPCIP extracts the primary private IP from a list of interfaces.
func extractVPCIP(interfaces []*computepb.NetworkInterface) string {
	for _, interf := range interfaces {
		if interf == nil || interf.NetworkIP == nil || interf.Name == nil || *interf.Name != "nic0" {
			continue
		}
		// return private IP from the default interface
		return *interf.NetworkIP
	}
	return ""
}

// extractPublicIP extracts a public IP from a list of interfaces.
func extractPublicIP(interfaces []*computepb.NetworkInterface) string {
	for _, interf := range interfaces {
		if interf == nil || interf.AccessConfigs == nil || interf.Name == nil || *interf.Name != "nic0" {
			continue
		}

		// return public IP from the default interface
		// GCP only supports one type of access config, so returning the first IP should result in a valid public IP
		for _, accessConfig := range interf.AccessConfigs {
			if accessConfig == nil || accessConfig.NatIP == nil {
				continue
			}
			return *accessConfig.NatIP
		}
	}
	return ""
}

// extractAliasIPRanges extracts alias interface IPs from a list of interfaces.
func extractAliasIPRanges(interfaces []*computepb.NetworkInterface) []string {
	ips := []string{}
	for _, interf := range interfaces {
		if interf == nil || interf.AliasIpRanges == nil {
			continue
		}
		for _, aliasIP := range interf.AliasIpRanges {
			if aliasIP == nil || aliasIP.IpCidrRange == nil {
				continue
			}
			ips = append(ips, *aliasIP.IpCidrRange)
		}
	}
	return ips
}

// extractSSHKeys extracts SSH keys from GCP instance metadata.
// reference: https://cloud.google.com/compute/docs/connect/add-ssh-keys .
func extractSSHKeys(metadata map[string]string) map[string][]string {
	sshKeysRaw, ok := metadata[gcpSSHMetadataKey]
	if !ok {
		// ignore missing metadata entry
		return map[string][]string{}
	}

	sshKeyLines := strings.Split(sshKeysRaw, "\n")
	keys := map[string][]string{}
	for _, sshKeyRaw := range sshKeyLines {
		keyParts := strings.SplitN(sshKeyRaw, ":", 2)
		if len(keyParts) != 2 {
			continue
		}
		username := keyParts[0]
		keyParts = strings.SplitN(keyParts[1], " ", 3)
		if len(keyParts) < 2 {
			continue
		}
		keyValue := fmt.Sprintf("%s %s", keyParts[0], keyParts[1])
		keys[username] = append(keys[username], keyValue)
	}
	return keys
}

func extractZone(in *computepb.Instance) (string, error) {
	if in == nil {
		return "", errors.New("instance is nil")
	}
	if in.Zone == nil {
		return "", errors.New("instance zone is nil")
	}
	zoneURL := *in.Zone
	zoneParts := strings.Split(zoneURL, "/")
	if zoneParts[len(zoneParts)-2] != "zones" {
		return "", fmt.Errorf("invalid zone URL: %s", zoneURL)
	}
	return zoneParts[len(zoneParts)-1], nil
}

// convertToMetadataInstance converts a *computepb.Instance to a core.Instance.
func convertToMetadataInstance(in *computepb.Instance, project string) (metadata.InstanceMetadata, error) {
	if in.Name == nil {
		return metadata.InstanceMetadata{}, fmt.Errorf("retrieving instance from compute API client returned invalid instance name: %v", in.Name)
	}
	zone, err := extractZone(in)
	if err != nil {
		return metadata.InstanceMetadata{}, fmt.Errorf("extracting zone from instance: %w", err)
	}
	mdata := extractInstanceMetadata(in.Metadata, "", false)
	return metadata.InstanceMetadata{
		Name:          *in.Name,
		ProviderID:    gcpshared.JoinProviderID(project, zone, *in.Name),
		Role:          extractRole(mdata),
		VPCIP:         extractVPCIP(in.NetworkInterfaces),
		PublicIP:      extractPublicIP(in.NetworkInterfaces),
		AliasIPRanges: extractAliasIPRanges(in.NetworkInterfaces),
		SSHKeys:       extractSSHKeys(mdata),
	}, nil
}

// extractInstanceMetadata will extract the list of instance metadata key-value pairs into a map.
// If "skipKey" is true, "key" will be skipped.
func extractInstanceMetadata(in *computepb.Metadata, key string, skipKey bool) map[string]string {
	metadataMap := map[string]string{}
	for _, item := range in.Items {
		if item == nil || item.Key == nil || item.Value == nil {
			continue
		}
		if skipKey && *item.Key == key {
			continue
		}
		metadataMap[*item.Key] = *item.Value
	}
	return metadataMap
}

// flattenInstanceMetadata takes a map of metadata key-value pairs and returns a flat list of computepb.Items inside computepb.Metadata.
func flattenInstanceMetadata(metadataMap map[string]string, fingerprint, kind *string) *computepb.Metadata {
	metadata := &computepb.Metadata{
		Fingerprint: fingerprint,
		Kind:        kind,
		Items:       make([]*computepb.Items, len(metadataMap)),
	}
	i := 0
	for mapKey, mapValue := range metadataMap {
		metadata.Items[i] = &computepb.Items{Key: proto.String(mapKey), Value: proto.String(mapValue)}
		i++
	}
	return metadata
}
