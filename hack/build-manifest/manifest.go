/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package main

import "encoding/json"

// Manifest contains all Constellation releases.
type Manifest struct {
	releases map[string]Images
}

// Images for all supported cloud providers.
type Images struct {
	AzureOSImage string `json:"AzureOSImage"`
	GCPOSImage   string `json:"GCPOSImage"`
}

// OldManifests provides Constellation releases to image mapping. These are the
// default images configured for each release.
func OldManifests() Manifest {
	return Manifest{
		releases: map[string]Images{
			"v1.0.0": {
				AzureOSImage: "/subscriptions/0d202bbb-4fa7-4af8-8125-58c269a05435/resourceGroups/CONSTELLATION-IMAGES/providers/Microsoft.Compute/galleries/Constellation/images/constellation-coreos/versions/0.0.1651150807",
				GCPOSImage:   "constellation-coreos-1651150807",
			},
			"v1.1.0": {
				AzureOSImage: "/subscriptions/0d202bbb-4fa7-4af8-8125-58c269a05435/resourceGroups/CONSTELLATION-IMAGES/providers/Microsoft.Compute/galleries/Constellation/images/constellation-coreos/versions/0.0.1654096948",
				GCPOSImage:   "projects/constellation-images/global/images/constellation-coreos-1654096948",
			},
			"v1.2.0": {
				AzureOSImage: "/subscriptions/0d202bbb-4fa7-4af8-8125-58c269a05435/resourceGroups/CONSTELLATION-IMAGES/providers/Microsoft.Compute/galleries/Constellation/images/constellation-coreos/versions/0.0.1654162332",
				GCPOSImage:   "projects/constellation-images/global/images/constellation-coreos-1654162332",
			},
			"v1.3.0": {
				AzureOSImage: "/subscriptions/0d202bbb-4fa7-4af8-8125-58c269a05435/resourceGroups/CONSTELLATION-IMAGES/providers/Microsoft.Compute/galleries/Constellation/images/constellation-coreos/versions/0.0.1654162332",
				GCPOSImage:   "projects/constellation-images/global/images/constellation-coreos-1654162332",
			},
			"v1.3.1": {
				AzureOSImage: "/subscriptions/0d202bbb-4fa7-4af8-8125-58c269a05435/resourceGroups/CONSTELLATION-IMAGES/providers/Microsoft.Compute/galleries/Constellation/images/constellation-coreos/versions/0.0.1657199013",
				GCPOSImage:   "projects/constellation-images/global/images/constellation-coreos-1657199013",
			},
			"v1.4.0": {
				AzureOSImage: "/subscriptions/0d202bbb-4fa7-4af8-8125-58c269a05435/resourceGroups/CONSTELLATION-IMAGES/providers/Microsoft.Compute/galleries/Constellation/images/constellation-coreos/versions/0.0.1659453699",
				GCPOSImage:   "projects/constellation-images/global/images/constellation-coreos-1659453699",
			},
		},
	}
}

// MarshalJSON marshals releases to JSON.
func (m *Manifest) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.releases)
}

// SetAzureImage for a given version.
func (m *Manifest) SetAzureImage(version string, image string) {
	if release, ok := m.releases[version]; !ok {
		images := Images{AzureOSImage: image}
		m.releases[version] = images
	} else {
		release.AzureOSImage = image
		m.releases[version] = release
	}
}

// SetGCPImage for a given version.
func (m *Manifest) SetGCPImage(version string, image string) {
	if release, ok := m.releases[version]; !ok {
		images := Images{GCPOSImage: image}
		m.releases[version] = images
	} else {
		release.GCPOSImage = image
		m.releases[version] = release
	}
}
