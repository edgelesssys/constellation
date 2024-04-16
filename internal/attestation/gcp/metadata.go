/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package gcp

import (
	"context"
	"encoding/json"
	"errors"
	"io"

	"cloud.google.com/go/compute/metadata"
	"github.com/google/go-tpm-tools/proto/attest"
)

// GCEInstanceInfo fetches VM metadata used for attestation from the GCE Metadata API.
func GCEInstanceInfo(client gcpMetadataClient) func(context.Context, io.ReadWriteCloser, []byte) ([]byte, error) {
	// Ideally we would want to use the endorsement public key certificate
	// However, this is not available on GCE instances
	// Workaround: Provide ShieldedVM instance info
	// The attestating party can request the VMs signing key using Google's API
	return func(context.Context, io.ReadWriteCloser, []byte) ([]byte, error) {
		projectID, err := client.ProjectID()
		if err != nil {
			return nil, errors.New("unable to fetch projectID")
		}
		zone, err := client.Zone()
		if err != nil {
			return nil, errors.New("unable to fetch zone")
		}
		instanceName, err := client.InstanceName()
		if err != nil {
			return nil, errors.New("unable to fetch instance name")
		}

		return json.Marshal(&attest.GCEInstanceInfo{
			Zone:         zone,
			ProjectId:    projectID,
			InstanceName: instanceName,
		})
	}
}

type gcpMetadataClient interface {
	ProjectID() (string, error)
	InstanceName() (string, error)
	Zone() (string, error)
}

// A MetadataClient fetches metadata from the GCE Metadata API.
type MetadataClient struct{}

// ProjectID returns the project ID of the GCE instance.
func (c MetadataClient) ProjectID() (string, error) {
	return metadata.ProjectID()
}

// InstanceName returns the instance name of the GCE instance.
func (c MetadataClient) InstanceName() (string, error) {
	return metadata.InstanceName()
}

// Zone returns the zone the GCE instance is located in.
func (c MetadataClient) Zone() (string, error) {
	return metadata.Zone()
}
