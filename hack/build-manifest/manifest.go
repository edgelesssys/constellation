/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package main

import "encoding/json"

type Manifest struct {
	releases map[string]Images
}

type Images struct {
	AzureCoreosImage string `json:"AzureCoreOSImage"`
	GCPCoreOSImage   string `json:"GCPCoreOSImage"`
}

// OldManifests provides Constellation releases to image mapping. These are the
// default images configured for each release.
func OldManifests() Manifest {
	return Manifest{
		releases: map[string]Images{
			"v1.0.0": {
				AzureCoreosImage: "/subscriptions/0d202bbb-4fa7-4af8-8125-58c269a05435/resourceGroups/CONSTELLATION-IMAGES/providers/Microsoft.Compute/galleries/Constellation/images/constellation-coreos/versions/0.0.1651150807",
				GCPCoreOSImage:   "constellation-coreos-1651150807",
			},
			"v1.1.0": {
				AzureCoreosImage: "/subscriptions/0d202bbb-4fa7-4af8-8125-58c269a05435/resourceGroups/CONSTELLATION-IMAGES/providers/Microsoft.Compute/galleries/Constellation/images/constellation-coreos/versions/0.0.1654096948",
				GCPCoreOSImage:   "projects/constellation-images/global/images/constellation-coreos-1654096948",
			},
			"v1.2.0": {
				AzureCoreosImage: "/subscriptions/0d202bbb-4fa7-4af8-8125-58c269a05435/resourceGroups/CONSTELLATION-IMAGES/providers/Microsoft.Compute/galleries/Constellation/images/constellation-coreos/versions/0.0.1654162332",
				GCPCoreOSImage:   "projects/constellation-images/global/images/constellation-coreos-1654162332",
			},
			"v1.3.0": {
				AzureCoreosImage: "/subscriptions/0d202bbb-4fa7-4af8-8125-58c269a05435/resourceGroups/CONSTELLATION-IMAGES/providers/Microsoft.Compute/galleries/Constellation/images/constellation-coreos/versions/0.0.1654162332",
				GCPCoreOSImage:   "projects/constellation-images/global/images/constellation-coreos-1654162332",
			},
			"v1.3.1": {
				AzureCoreosImage: "/subscriptions/0d202bbb-4fa7-4af8-8125-58c269a05435/resourceGroups/CONSTELLATION-IMAGES/providers/Microsoft.Compute/galleries/Constellation/images/constellation-coreos/versions/0.0.1657199013",
				GCPCoreOSImage:   "projects/constellation-images/global/images/constellation-coreos-1657199013",
			},
			"v1.4.0": {
				AzureCoreosImage: "/subscriptions/0d202bbb-4fa7-4af8-8125-58c269a05435/resourceGroups/CONSTELLATION-IMAGES/providers/Microsoft.Compute/galleries/Constellation/images/constellation-coreos/versions/0.0.1659453699",
				GCPCoreOSImage:   "projects/constellation-images/global/images/constellation-coreos-1659453699",
			},
		},
	}
}

func (m *Manifest) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.releases)
}

func (m *Manifest) SetAzureImage(version string, image string) {
	if release, ok := m.releases[version]; !ok {
		images := Images{AzureCoreosImage: image}
		m.releases[version] = images
	} else {
		release.AzureCoreosImage = image
		m.releases[version] = release
	}
}

func (m *Manifest) SetGCPImage(version string, image string) {
	if release, ok := m.releases[version]; !ok {
		images := Images{GCPCoreOSImage: image}
		m.releases[version] = images
	} else {
		release.GCPCoreOSImage = image
		m.releases[version] = release
	}
}
