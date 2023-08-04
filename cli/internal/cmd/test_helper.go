/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/api/attestationconfigapi"
	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
)

var testCfg = attestationconfigapi.AzureSEVSNPVersion{
	Microcode:  93,
	TEE:        0,
	SNP:        6,
	Bootloader: 2,
}

type StubAttestationFetcher struct{}

func (f StubAttestationFetcher) FetchAzureSEVSNPVersionList(_ context.Context, _ attestationconfigapi.AzureSEVSNPVersionList) (attestationconfigapi.AzureSEVSNPVersionList, error) {
	return attestationconfigapi.AzureSEVSNPVersionList(
		[]string{},
	), nil
}

func (f StubAttestationFetcher) FetchAzureSEVSNPVersion(_ context.Context, _ attestationconfigapi.AzureSEVSNPVersionAPI) (attestationconfigapi.AzureSEVSNPVersionAPI, error) {
	return attestationconfigapi.AzureSEVSNPVersionAPI{
		AzureSEVSNPVersion: testCfg,
	}, nil
}

func (f StubAttestationFetcher) FetchAzureSEVSNPVersionLatest(_ context.Context, _ time.Time) (attestationconfigapi.AzureSEVSNPVersionAPI, error) {
	return attestationconfigapi.AzureSEVSNPVersionAPI{
		AzureSEVSNPVersion: testCfg,
	}, nil
}

type NopSpinner struct {
	io.Writer
}

func (s *NopSpinner) Start(string, bool) {}
func (s *NopSpinner) Stop()              {}
func (s *NopSpinner) Write(p []byte) (n int, err error) {
	return s.Writer.Write(p)
}

func DefaultConfigWithExpectedMeasurements(t *testing.T, conf *config.Config, csp cloudprovider.Provider) *config.Config {
	t.Helper()

	conf.RemoveProviderAndAttestationExcept(csp)

	conf.Image = constants.BinaryVersion().String()
	conf.Name = "kubernetes"

	var zone, instanceType, diskType string
	switch csp {
	case cloudprovider.Azure:
		conf.Provider.Azure.SubscriptionID = "01234567-0123-0123-0123-0123456789ab"
		conf.Provider.Azure.TenantID = "01234567-0123-0123-0123-0123456789ab"
		conf.Provider.Azure.Location = "test-location"
		conf.Provider.Azure.UserAssignedIdentity = "test-identity"
		conf.Provider.Azure.ResourceGroup = "test-resource-group"
		conf.Attestation.AzureSEVSNP.Measurements[4] = measurements.WithAllBytes(0x44, measurements.Enforce, measurements.PCRMeasurementLength)
		conf.Attestation.AzureSEVSNP.Measurements[9] = measurements.WithAllBytes(0x11, measurements.Enforce, measurements.PCRMeasurementLength)
		conf.Attestation.AzureSEVSNP.Measurements[12] = measurements.WithAllBytes(0xcc, measurements.Enforce, measurements.PCRMeasurementLength)
		instanceType = "Standard_DC4as_v5"
		diskType = "StandardSSD_LRS"
	case cloudprovider.GCP:
		conf.Provider.GCP.Region = "test-region"
		conf.Provider.GCP.Project = "test-project"
		conf.Provider.GCP.Zone = "test-zone"
		conf.Provider.GCP.ServiceAccountKeyPath = "test-key-path"
		conf.Attestation.GCPSEVES.Measurements[4] = measurements.WithAllBytes(0x44, measurements.Enforce, measurements.PCRMeasurementLength)
		conf.Attestation.GCPSEVES.Measurements[9] = measurements.WithAllBytes(0x11, measurements.Enforce, measurements.PCRMeasurementLength)
		conf.Attestation.GCPSEVES.Measurements[12] = measurements.WithAllBytes(0xcc, measurements.Enforce, measurements.PCRMeasurementLength)
		zone = "europe-west3-b"
		instanceType = "n2d-standard-4"
		diskType = "pd-ssd"
	case cloudprovider.QEMU:
		conf.Attestation.QEMUVTPM.Measurements[4] = measurements.WithAllBytes(0x44, measurements.Enforce, measurements.PCRMeasurementLength)
		conf.Attestation.QEMUVTPM.Measurements[9] = measurements.WithAllBytes(0x11, measurements.Enforce, measurements.PCRMeasurementLength)
		conf.Attestation.QEMUVTPM.Measurements[12] = measurements.WithAllBytes(0xcc, measurements.Enforce, measurements.PCRMeasurementLength)
	}

	for groupName, group := range conf.NodeGroups {
		group.Zone = zone
		group.InstanceType = instanceType
		group.StateDiskType = diskType
		conf.NodeGroups[groupName] = group
	}

	return conf
}
