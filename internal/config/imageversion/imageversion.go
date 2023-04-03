/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package imageversion contains the pinned container images for the config.
package imageversion

import "github.com/edgelesssys/constellation/v2/internal/containerimage"

// QEMUMetadata is the image of the QEMU metadata api service.
func QEMUMetadata() string {
	return defaultQEMUMetadata.String()
}

// Libvirt is the image of the libvirt container.
func Libvirt() string {
	return defaultLibvirt.String()
}

var (
	defaultQEMUMetadata = containerimage.Image{
		Registry: qemuMetadataRegistry,
		Prefix:   qemuMetadataPrefix,
		Name:     qemuMetadataName,
		Tag:      qemuMetadataTag,
		Digest:   qemuMetadataDigest,
	}
	defaultLibvirt = containerimage.Image{
		Registry: libvirtRegistry,
		Prefix:   libvirtPrefix,
		Name:     libvirtName,
		Tag:      libvirtTag,
		Digest:   libvirtDigest,
	}
)
