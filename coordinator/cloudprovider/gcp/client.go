package gcp

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	compute "cloud.google.com/go/compute/apiv1"
	"github.com/edgelesssys/constellation/coordinator/cloudprovider"
	"github.com/edgelesssys/constellation/coordinator/core"
	"google.golang.org/api/iterator"
	computepb "google.golang.org/genproto/googleapis/cloud/compute/v1"
	"google.golang.org/protobuf/proto"
)

const gcpSSHMetadataKey = "ssh-keys"

// Client implements the gcp.API interface.
type Client struct {
	instanceAPI
	metadataAPI
}

// NewClient creates a new Client.
func NewClient(ctx context.Context) (*Client, error) {
	insAPI, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		return nil, err
	}
	return &Client{instanceAPI: &instanceClient{insAPI}, metadataAPI: &metadataClient{}}, nil
}

// RetrieveInstances returns list of instances including their ips and metadata.
func (c *Client) RetrieveInstances(ctx context.Context, project, zone string) ([]core.Instance, error) {
	uid, err := c.uid()
	if err != nil {
		return nil, err
	}
	req := &computepb.ListInstancesRequest{
		Project: project,
		Zone:    zone,
	}
	instanceIterator := c.instanceAPI.List(ctx, req)

	instances := []core.Instance{}
	for {
		resp, err := instanceIterator.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("retrieving instance list from compute API client failed: %w", err)
		}
		metadata := extractInstanceMetadata(resp.Metadata, "", false)
		// skip instances not belonging to the current constellation
		if instanceUID, ok := metadata[core.ConstellationUIDMetadataKey]; !ok || instanceUID != uid {
			continue
		}
		instance, err := convertToCoreInstance(resp, project, zone)
		if err != nil {
			return nil, err
		}

		instances = append(instances, instance)
	}
	return instances, nil
}

// RetrieveInstance returns a an instance including ips and metadata.
func (c *Client) RetrieveInstance(ctx context.Context, project, zone, instanceName string) (core.Instance, error) {
	instance, err := c.getComputeInstance(ctx, project, zone, instanceName)
	if err != nil {
		return core.Instance{}, err
	}

	return convertToCoreInstance(instance, project, zone)
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
		return "", fmt.Errorf("requesting GCP instance metadata failed: %w", err)
	}
	return value, nil
}

// SetInstanceMetadata modifies a key value pair of metadata for the instance specified by project, zone and instanceName.
func (c *Client) SetInstanceMetadata(ctx context.Context, project, zone, instanceName, key, value string) error {
	instance, err := c.getComputeInstance(ctx, project, zone, instanceName)
	if err != nil {
		return fmt.Errorf("retrieving instance metadata failed: %w", err)
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
		return fmt.Errorf("setting instance metadata %v: %v failed with: %w", key, value, err)
	}
	return nil
}

// UnsetInstanceMetadata modifies a key value pair of metadata for the instance specified by project, zone and instanceName.
func (c *Client) UnsetInstanceMetadata(ctx context.Context, project, zone, instanceName, key string) error {
	instance, err := c.getComputeInstance(ctx, project, zone, instanceName)
	if err != nil {
		return fmt.Errorf("retrieving instance metadata failed: %w", err)
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
		return fmt.Errorf("unsetting instance metadata key %v failed with: %w", key, err)
	}
	return nil
}

// Close closes the instanceAPI client.
func (c *Client) Close() error {
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
		return nil, fmt.Errorf("retrieving compute instance failed: %w", err)
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
		return fmt.Errorf("updating instance metadata failed: %w", err)
	}
	return nil
}

// uid retrieves the current instances uid.
func (c *Client) uid() (string, error) {
	// API endpoint: http://metadata.google.internal/computeMetadata/v1/instance/attributes/constellation-uid
	uid, err := c.RetrieveInstanceMetadata(core.ConstellationUIDMetadataKey)
	if err != nil {
		return "", fmt.Errorf("retrieving constellation uid failed: %w", err)
	}
	return uid, nil
}

// extractIPs extracts private interface IPs from a list of interfaces.
func extractIPs(interfaces []*computepb.NetworkInterface) []string {
	ips := []string{}
	for _, interf := range interfaces {
		if interf == nil || interf.NetworkIP == nil {
			continue
		}
		ips = append(ips, *interf.NetworkIP)
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

// convertToCoreInstance converts a *computepb.Instance to a core.Instance.
func convertToCoreInstance(in *computepb.Instance, project string, zone string) (core.Instance, error) {
	if in.Name == nil {
		return core.Instance{}, fmt.Errorf("retrieving instance from compute API client returned invalid instance Name: %v", in.Name)
	}
	metadata := extractInstanceMetadata(in.Metadata, "", false)
	return core.Instance{
		Name:       *in.Name,
		ProviderID: joinProviderID(project, zone, *in.Name),
		Role:       cloudprovider.ExtractRole(metadata),
		IPs:        extractIPs(in.NetworkInterfaces),
		SSHKeys:    extractSSHKeys(metadata),
	}, nil
}

// joinProviderID builds a k8s provider ID for GCP instances.
// A providerID is build after the schema 'gce://<project-id>/<zone>/<instance-name>'
func joinProviderID(project, zone, instanceName string) string {
	return fmt.Sprintf("gce://%v/%v/%v", project, zone, instanceName)
}

// splitProviderID splits a provider's id into core components.
// A providerID is build after the schema 'gce://<project-id>/<zone>/<instance-name>'
func splitProviderID(providerID string) (project, zone, instance string, err error) {
	// providerIDregex is a regex matching a gce providerID with each part of the URI being a submatch.
	providerIDregex := regexp.MustCompile(`^gce://([^/]+)/([^/]+)/([^/]+)$`)
	matches := providerIDregex.FindStringSubmatch(providerID)
	if len(matches) != 4 {
		return "", "", "", errors.New("error splitting providerID")
	}
	return matches[1], matches[2], matches[3], nil
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
