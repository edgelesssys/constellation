/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package client

import (
	"fmt"
	"regexp"
)

var providerIDRegex = regexp.MustCompile(`^gce://([^/]+)/([^/]+)/([^/]+)$`)

// splitProviderID splits a provider's id into core components.
// A providerID is build after the schema 'gce://<project-id>/<zone>/<instance-name>'
func splitProviderID(providerID string) (project, zone, instance string, err error) {
	matches := providerIDRegex.FindStringSubmatch(providerID)

	if len(matches) != 4 {
		return "", "", "", fmt.Errorf("splitting providerID: %q. matches: %v", providerID, matches)
	}
	return matches[1], matches[2], matches[3], nil
}

// joinProviderID builds a k8s provider ID for GCP instances.
// A providerID is build after the schema 'gce://<project-id>/<zone>/<instance-name>'
func joinProviderID(project, zone, instanceName string) string {
	return fmt.Sprintf("gce://%v/%v/%v", project, zone, instanceName)
}

// joinInstanceID builds a gcp instance ID from the zone and instance name.
func joinInstanceID(zone, instanceName string) string {
	return fmt.Sprintf("zones/%v/instances/%v", zone, instanceName)
}
