/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package gcp

import (
	"encoding/json"
	"errors"
	"io"

	"cloud.google.com/go/compute/metadata"
	"github.com/edgelesssys/constellation/v2/internal/attestation/vtpm"
	"github.com/edgelesssys/constellation/v2/internal/oid"
	tpmclient "github.com/google/go-tpm-tools/client"
	"github.com/google/go-tpm-tools/proto/attest"
)

// Issuer for GCP confidential VM attestation.
type Issuer struct {
	oid.GCP
	*vtpm.Issuer
}

// NewIssuer initializes a new GCP Issuer.
func NewIssuer(log vtpm.AttestationLogger) *Issuer {
	return &Issuer{
		Issuer: vtpm.NewIssuer(
			vtpm.OpenVTPM,
			tpmclient.GceAttestationKeyRSA,
			getGCEInstanceInfo(metadataClient{}),
			log,
		),
	}
}

// getGCEInstanceInfo fetches VM metadata used for attestation.
func getGCEInstanceInfo(client gcpMetadataClient) func(io.ReadWriteCloser) ([]byte, error) {
	// Ideally we would want to use the endorsement public key certificate
	// However, this is not available on GCE instances
	// Workaround: Provide ShieldedVM instance info
	// The attestating party can request the VMs signing key using Google's API
	return func(io.ReadWriteCloser) ([]byte, error) {
		projectID, err := client.projectID()
		if err != nil {
			return nil, errors.New("unable to fetch projectID")
		}
		zone, err := client.zone()
		if err != nil {
			return nil, errors.New("unable to fetch zone")
		}
		instanceName, err := client.instanceName()
		if err != nil {
			return nil, errors.New("unable to fetch instance name")
		}

		return json.Marshal(attest.GCEInstanceInfo{
			Zone:         zone,
			ProjectId:    projectID,
			InstanceName: instanceName,
		})
	}
}

type gcpMetadataClient interface {
	projectID() (string, error)
	instanceName() (string, error)
	zone() (string, error)
}

type metadataClient struct{}

func (c metadataClient) projectID() (string, error) {
	return metadata.ProjectID()
}

func (c metadataClient) instanceName() (string, error) {
	return metadata.InstanceName()
}

func (c metadataClient) zone() (string, error) {
	return metadata.Zone()
}
