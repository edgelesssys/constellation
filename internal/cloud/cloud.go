/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
# Cloud

This package provides functions to interact with cloud providers.
This is mainly used to fetch information about the current instance, or other instances of the Constellation cluster.

Implementation of the cloud provider specific code is done in subpackages named after the CSP.
Code that is commonly used by other packages that do not require actual interaction with the CSP API,
such as CSP URI string parsing or data types, should go in a <CSP>shared package instead.

A cloud package should implement the following interface:

	type Cloud interface {
		List(ctx context.Context) ([]metadata.InstanceMetadata, error)
		Self(ctx context.Context) (metadata.InstanceMetadata, error)
		GetLoadBalancerEndpoint(ctx context.Context) (string, error)
		InitSecretHash(ctx context.Context) ([]byte, error)
		UID(ctx context.Context) (string, error)
	}
*/
package cloud

const (
	// TagRole is the tag/label key used to identify the role of a node.
	TagRole = "constellation-role"
	// TagUID is the tag/label key used to identify the UID of a cluster.
	TagUID = "constellation-uid"
	// TagInitSecretHash is the tag/label key used to identify the hash of the init secret.
	TagInitSecretHash = "constellation-init-secret-hash"
	// TagCustomEndpoint is the tag/label key used to identify the custom endpoint
	// or dns name that should be added to tls cert SANs.
	TagCustomEndpoint = "constellation-custom-endpoint"
)
