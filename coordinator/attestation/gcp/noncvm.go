package gcp

import (
	"context"
	"fmt"
	"time"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/metadata"
	"github.com/edgelesssys/constellation/coordinator/attestation/vtpm"
	"github.com/edgelesssys/constellation/coordinator/oid"
	tpmclient "github.com/google/go-tpm-tools/client"
	computepb "google.golang.org/genproto/googleapis/cloud/compute/v1"
)

// NonCVMValidator is a validator for regular GCP VMs with vTPM.
// TODO: Remove once we no longer use non cvms.
type NonCVMValidator struct {
	oid.GCPNonCVM
	*vtpm.Validator
}

// NewNonCVMValidator initializes a new non CVM GCP validator with the provided PCR values.
// TODO: Remove once we no longer use non cvms.
func NewNonCVMValidator(pcrs map[uint32][]byte) *NonCVMValidator {
	return &NonCVMValidator{
		Validator: vtpm.NewValidator(
			pcrs,
			trustedKeyFromGCEAPI(newInstanceClient),
			func(attestation vtpm.AttestationDocument) error { return nil },
			vtpm.VerifyPKCS1v15,
		),
	}
}

// NonCVNMIssuer for GCP confindetial VM attestation.
// TODO: Remove once we no longer use non cvms.
type NonCVMIssuer struct {
	oid.GCPNonCVM
	*vtpm.Issuer
}

// NewNonCVNMIssuer initializes a new GCP Issuer.
// TODO: Remove once we no longer use non cvms.
func NewNonCVMIssuer() *NonCVMIssuer {
	return &NonCVMIssuer{
		Issuer: vtpm.NewIssuer(
			vtpm.OpenVTPM,
			tpmclient.GceAttestationKeyRSA,
			getGCEInstanceInfo(metadataClient{}),
		),
	}
}

// IsCVM returns true if the VM has confidential computing capabilities enabled.
func IsCVM() (bool, error) {
	project, err := metadata.ProjectID()
	if err != nil {
		return false, err
	}
	zone, err := metadata.Zone()
	if err != nil {
		return false, err
	}
	instance, err := metadata.InstanceName()
	if err != nil {
		return false, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	client, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		return false, err
	}
	defer client.Close()
	infos, err := client.Get(ctx, &computepb.GetInstanceRequest{
		Instance: instance,
		Project:  project,
		Zone:     zone,
	})
	if err != nil {
		return false, err
	}

	if infos.ConfidentialInstanceConfig == nil {
		return false, fmt.Errorf("received empty confidential instance config")
	}

	return *infos.ConfidentialInstanceConfig.EnableConfidentialCompute, nil
}
