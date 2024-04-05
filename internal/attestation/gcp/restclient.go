/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package gcp

import (
	"context"
	"crypto"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/edgelesssys/constellation/v2/internal/attestation/snp"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/attestation/vtpm"
	"github.com/google/go-tpm-tools/proto/attest"
	"github.com/googleapis/gax-go/v2"
	"google.golang.org/api/option"
)

// RESTClient is a client for the GCE API.
type RESTClient struct {
	*compute.InstancesClient
}

// NewRESTClient creates a new RESTClient.
func NewRESTClient(ctx context.Context, opts ...option.ClientOption) (CVMRestClient, error) {
	c, err := compute.NewInstancesRESTClient(ctx, opts...)
	if err != nil {
		return nil, err
	}
	return &RESTClient{c}, nil
}

// CVMRestClient is the interface a GCP REST client for a CVM must implement.
type CVMRestClient interface {
	GetShieldedInstanceIdentity(ctx context.Context, req *computepb.GetShieldedInstanceIdentityInstanceRequest, opts ...gax.CallOption) (*computepb.ShieldedInstanceIdentity, error)
	Close() error
}

// TrustedKeyGetter returns a function that queries the GCE API for a shieldedVM's public signing key.
// This key can be used to verify attestation statements issued by the VM.
func TrustedKeyGetter(
	attestationVariant variant.Variant,
	newRESTClient func(ctx context.Context, opts ...option.ClientOption) (CVMRestClient, error),
) (func(ctx context.Context, attDoc vtpm.AttestationDocument, _ []byte) (crypto.PublicKey, error), error) {
	return func(ctx context.Context, attDoc vtpm.AttestationDocument, _ []byte) (crypto.PublicKey, error) {
		client, err := newRESTClient(ctx)
		if err != nil {
			return nil, fmt.Errorf("creating GCE client: %w", err)
		}
		defer client.Close()

		var gceInstanceInfo attest.GCEInstanceInfo
		switch attestationVariant {
		case variant.GCPSEVES{}:
			if err := json.Unmarshal(attDoc.InstanceInfo, &gceInstanceInfo); err != nil {
				return nil, err
			}
		case variant.GCPSEVSNP{}:
			var instanceInfo snp.InstanceInfo
			if err := json.Unmarshal(attDoc.InstanceInfo, &instanceInfo); err != nil {
				return nil, err
			}
			gceInstanceInfo = attest.GCEInstanceInfo{
				InstanceName: instanceInfo.GCP.InstanceName,
				ProjectId:    instanceInfo.GCP.ProjectId,
				Zone:         instanceInfo.GCP.Zone,
			}
		default:
			return nil, fmt.Errorf("unsupported attestation variant: %v", attestationVariant)
		}

		instance, err := client.GetShieldedInstanceIdentity(ctx, &computepb.GetShieldedInstanceIdentityInstanceRequest{
			Instance: gceInstanceInfo.GetInstanceName(),
			Project:  gceInstanceInfo.GetProjectId(),
			Zone:     gceInstanceInfo.GetZone(),
		})
		if err != nil {
			return nil, fmt.Errorf("retrieving VM identity: %w", err)
		}

		if instance.SigningKey == nil || instance.SigningKey.EkPub == nil {
			return nil, fmt.Errorf("received no signing key from GCP API")
		}

		// Parse the signing key return by GetShieldedInstanceIdentity
		block, _ := pem.Decode([]byte(*instance.SigningKey.EkPub))
		if block == nil || block.Type != "PUBLIC KEY" {
			return nil, fmt.Errorf("failed to decode PEM block containing public key")
		}

		return x509.ParsePKIXPublicKey(block.Bytes)
	}, nil
}
