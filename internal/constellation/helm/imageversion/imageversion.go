/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package imageversion contains the pinned container images for the helm charts.
package imageversion

import "github.com/edgelesssys/constellation/v2/internal/containerimage"

// TODO(malt3): Migrate third-party images from versions.go.

// JoinService is the image of the join service.
// registry and prefix can be optionally set to use a different source.
func JoinService(registry, prefix string) string {
	return containerimage.NewBuilder(defaultJoinService, registry, prefix).Build().String()
}

// KeyService is the image of the key service.
// registry and prefix can be optionally set to use a different source.
func KeyService(registry, prefix string) string {
	return containerimage.NewBuilder(defaultKeyService, registry, prefix).Build().String()
}

// VerificationService is the image of the verification service.
// registry and prefix can be optionally set to use a different source.
func VerificationService(registry, prefix string) string {
	return containerimage.NewBuilder(defaultVerificationService, registry, prefix).Build().String()
}

// ConstellationNodeOperator is the image of the constellation node operator.
// registry and prefix can be optionally set to use a different source.
func ConstellationNodeOperator(registry, prefix string) string {
	return containerimage.NewBuilder(defaultNodeOperator, registry, prefix).Build().String()
}

var (
	defaultJoinService = containerimage.Image{
		Registry: joinServiceRegistry,
		Prefix:   joinServicePrefix,
		Name:     joinServiceName,
		Tag:      joinServiceTag,
		Digest:   joinServiceDigest,
	}
	defaultKeyService = containerimage.Image{
		Registry: keyServiceRegistry,
		Prefix:   keyServicePrefix,
		Name:     keyServiceName,
		Tag:      keyServiceTag,
		Digest:   keyServiceDigest,
	}
	defaultVerificationService = containerimage.Image{
		Registry: verificationServiceRegistry,
		Prefix:   verificationServicePrefix,
		Name:     verificationServiceName,
		Tag:      verificationServiceTag,
		Digest:   verificationServiceDigest,
	}
	defaultNodeOperator = containerimage.Image{
		Registry: constellationNodeOperatorRegistry,
		Prefix:   constellationNodeOperatorPrefix,
		Name:     constellationNodeOperatorName,
		Tag:      constellationNodeOperatorTag,
		Digest:   constellationNodeOperatorDigest,
	}
)
