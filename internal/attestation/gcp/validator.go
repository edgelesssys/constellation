/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package gcp

import (
	"bytes"
	"context"
	"crypto"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"time"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/attestation/vtpm"
	"github.com/edgelesssys/constellation/v2/internal/oid"
	"github.com/google/go-tpm-tools/proto/attest"
	"github.com/google/go-tpm-tools/server"
	"github.com/googleapis/gax-go/v2"
	"google.golang.org/api/option"
)

// Validator for GCP confidential VM attestation.
type Validator struct {
	oid.GCPSEVES
	*vtpm.Validator
}

// NewValidator initializes a new GCP validator with the provided PCR values.
func NewValidator(pcrs measurements.M, log vtpm.AttestationLogger) *Validator {
	return &Validator{
		Validator: vtpm.NewValidator(
			pcrs,
			trustedKeyFromGCEAPI(newInstanceClient),
			gceNonHostInfoEvent,
			log,
		),
	}
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
func trustedKeyFromGCEAPI(getClient func(ctx context.Context, opts ...option.ClientOption) (gcpRestClient, error)) func(akPub []byte, instanceInfoRaw []byte) (crypto.PublicKey, error) {
	return func(akPub, instanceInfoRaw []byte) (crypto.PublicKey, error) {
		var instanceInfo attest.GCEInstanceInfo
		if err := json.Unmarshal(instanceInfoRaw, &instanceInfo); err != nil {
			return nil, err
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		client, err := getClient(ctx)
		if err != nil {
			return nil, fmt.Errorf("creating GCE client: %w", err)
		}
		defer client.Close()

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
}

// gceNonHostInfoEvent looks for the GCE Non-Host info event in an event log.
// Returns an error if the event is not found, or if the event is missing the required flag to mark the VM confidential.
func gceNonHostInfoEvent(attDoc vtpm.AttestationDocument, _ *attest.MachineState) error {
	if attDoc.Attestation == nil {
		return errors.New("missing attestation in attestation document")
	}
	// The event log of a GCE VM contains the GCE Non-Host info event
	// This event is 32-bytes, followed by one byte 0x01 if it is confidential, 0x00 otherwise,
	// followed by 15 reserved bytes.
	// See https://pkg.go.dev/github.com/google/go-tpm-tools@v0.3.1/server#pkg-variables
	idx := bytes.Index(attDoc.Attestation.EventLog, server.GCENonHostInfoSignature)
	if idx <= 0 {
		return fmt.Errorf("event log is missing GCE Non-Host info event")
	}
	if attDoc.Attestation.EventLog[idx+len(server.GCENonHostInfoSignature)] != 0x01 {
		return fmt.Errorf("GCE Non-Host info is missing confidential bit")
	}
	return nil
}
