/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package client

import (
	"fmt"
	"regexp"

	"google.golang.org/genproto/googleapis/cloud/compute/v1"
)

var (
	diskSourceRegex = regexp.MustCompile(`^https://www.googleapis.com/compute/v1/projects/([^/]+)/zones/([^/]+)/disks/([^/]+)$`)
	computeAPIBase  = regexp.MustCompile(`^https://www.googleapis.com/compute/v1/(.+)$`)
)

// diskSourceToDiskReq converts a disk source URI to a disk request.
func diskSourceToDiskReq(diskSource string) (*compute.GetDiskRequest, error) {
	matches := diskSourceRegex.FindStringSubmatch(diskSource)
	if len(matches) != 4 {
		return nil, fmt.Errorf("error splitting diskSource: %v", diskSource)
	}
	return &compute.GetDiskRequest{
		Disk:    matches[3],
		Project: matches[1],
		Zone:    matches[2],
	}, nil
}

// uriNormalize normalizes a compute API URI by removing the optional URI prefix.
// for normalization, the prefix 'https://www.googleapis.com/compute/v1/' is removed.
func uriNormalize(imageURI string) string {
	matches := computeAPIBase.FindStringSubmatch(imageURI)
	if len(matches) != 2 {
		return imageURI
	}
	return matches[1]
}
