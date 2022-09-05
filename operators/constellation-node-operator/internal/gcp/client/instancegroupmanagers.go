/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package client

import (
	"context"
	"fmt"
	"regexp"
)

var (
	instanceGroupIDRegex               = regexp.MustCompile(`^projects/([^/]+)/zones/([^/]+)/instanceGroupManagers/([^/]+)$`)
	controlPlaneInstanceGroupNameRegex = regexp.MustCompile(`^(.*)control-plane(.*)$`)
	workerInstanceGroupNameRegex       = regexp.MustCompile(`^(.*)worker(.*)$`)
)

func (c *Client) canonicalInstanceGroupID(ctx context.Context, instanceGroupID string) (string, error) {
	project, zone, instanceGroup, err := splitInstanceGroupID(uriNormalize(instanceGroupID))
	if err != nil {
		return "", err
	}
	project, err = c.canonicalProjectID(ctx, project)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/zones/%s/instanceGroupManagers/%s", project, zone, instanceGroup), nil
}

// splitInstanceGroupID splits an instance group ID into core components.
func splitInstanceGroupID(instanceGroupID string) (project, zone, instanceGroup string, err error) {
	matches := instanceGroupIDRegex.FindStringSubmatch(instanceGroupID)
	if len(matches) != 4 {
		return "", "", "", fmt.Errorf("error splitting instanceGroupID: %v", instanceGroupID)
	}
	return matches[1], matches[2], matches[3], nil
}

// isControlPlaneInstanceGroup returns true if the instance group is a control plane instance group.
func isControlPlaneInstanceGroup(instanceGroupName string) bool {
	return controlPlaneInstanceGroupNameRegex.MatchString(instanceGroupName)
}

// isWorkerInstanceGroup returns true if the instance group is a worker instance group.
func isWorkerInstanceGroup(instanceGroupName string) bool {
	return workerInstanceGroupNameRegex.MatchString(instanceGroupName)
}

// generateInstanceName generates a random instance name.
func generateInstanceName(baseInstanceName string, random prng) string {
	letters := []byte("abcdefghijklmnopqrstuvwxyz0123456789")
	const uidLen = 4
	uid := make([]byte, 0, uidLen)
	for i := 0; i < uidLen; i++ {
		n := random.Intn(len(letters))
		uid = append(uid, letters[n])
	}
	return baseInstanceName + "-" + string(uid)
}
