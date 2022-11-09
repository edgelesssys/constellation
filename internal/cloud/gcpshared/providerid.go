/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package gcpshared

import (
	"fmt"
	"regexp"
)

var providerIDRegex = regexp.MustCompile(`^gce://([^/]+)/([^/]+)/([^/]+)$`)

// SplitProviderID splits a k8s provider ID for GCP instances into its core components.
// A provider ID is build after the schema 'gce://<project-id>/<zone>/<instance-name>'
func SplitProviderID(providerID string) (project, zone, instance string, err error) {
	matches := providerIDRegex.FindStringSubmatch(providerID)
	if len(matches) != 4 {
		return "", "", "", fmt.Errorf("error splitting providerID: %v", providerID)
	}
	return matches[1], matches[2], matches[3], nil
}

// JoinProviderID builds a k8s provider ID for GCP instances.
// A providerID is build after the schema 'gce://<project-id>/<zone>/<instance-name>'
func JoinProviderID(project, zone, instanceName string) string {
	return fmt.Sprintf("gce://%v/%v/%v", project, zone, instanceName)
}
