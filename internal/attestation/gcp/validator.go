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
	"github.com/edgelesssys/constellation/v2/internal/attestation"
	"github.com/edgelesssys/constellation/v2/internal/attestation/vtpm"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/variant"
	"github.com/google/go-tpm-tools/proto/attest"
	"github.com/googleapis/gax-go/v2"
	"google.golang.org/api/option"
)

const minimumGceVersion = 1

// Validator for GCP confidential VM attestation.
type Validator struct {
	variant.GCPSEVES
	*vtpm.Validator

	restClient func(context.Context, ...option.ClientOption) (gcpRestClient, error)
}

// NewValidator initializes a new GCP validator with the provided PCR values.
func NewValidator(cfg *config.GCPSEVES, log attestation.Logger) *Validator {
	v := &Validator{
		restClient: newInstanceClient,
	}
	v.Validator = vtpm.NewValidator(
		cfg.Measurements,
		v.trustedKeyFromGCEAPI,
		validateCVM,
		log,
	)

	return v
}

type gcpRestClient interface {
	GetShieldedInstanceIdentity(ctx context.Context, req *computepb.GetShieldedInstanceIdentityInstanceRequest, opts ...gax.CallOption) (*computepb.ShieldedInstanceIdentity, error)
	Close() error
}

type instanceClient struct {
	*compute.InstancesClient
}

func newInstanceClient(ctx context.Context, opts ...option.ClientOption) (gcpRestClient, error) {
	c, err := compute.NewInstancesRESTClient(ctx, opts...)
	if err != nil {
		return nil, err
	}
	return &instanceClient{c}, nil
}

// trustedKeyFromGCEAPI queries the GCE API for a shieldedVM's public signing key.
// This key can be used to verify attestation statements issued by the VM.
func (v *Validator) trustedKeyFromGCEAPI(ctx context.Context, attDoc vtpm.AttestationDocument, _ []byte) (crypto.PublicKey, error) {
	client, err := v.restClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("creating GCE client: %w", err)
	}
	defer client.Close()

	var instanceInfo attest.GCEInstanceInfo
	if err := json.Unmarshal(attDoc.InstanceInfo, &instanceInfo); err != nil {
		return nil, err
	}

	instance, err := client.GetShieldedInstanceIdentity(ctx, &computepb.GetShieldedInstanceIdentityInstanceRequest{
		Instance: instanceInfo.GetInstanceName(),
		Project:  instanceInfo.GetProjectId(),
		Zone:     instanceInfo.GetZone(),
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
}

// validateCVM checks that the machine state represents a GCE AMD-SEV VM.
func validateCVM(_ vtpm.AttestationDocument, state *attest.MachineState) error {
	gceVersion := state.Platform.GetGceVersion()
	if gceVersion < minimumGceVersion {
		return fmt.Errorf("outdated GCE version: %v (require >= %v)", gceVersion, minimumGceVersion)
	}

	tech := state.Platform.Technology
	wantTech := attest.GCEConfidentialTechnology_AMD_SEV
	if tech != wantTech {
		return fmt.Errorf("unexpected confidential technology: %v (expected: %v)", tech, wantTech)
	}

	return nil
}
