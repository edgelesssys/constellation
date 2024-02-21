//go:build !enterprise

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package measurements

// revive:disable:var-naming
var (
	aws_AWSNitroTPM = M{
		4:                         PlaceHolderMeasurement(PCRMeasurementLength),
		8:                         WithAllBytes(0x00, Enforce, PCRMeasurementLength),
		9:                         PlaceHolderMeasurement(PCRMeasurementLength),
		11:                        WithAllBytes(0x00, Enforce, PCRMeasurementLength),
		12:                        PlaceHolderMeasurement(PCRMeasurementLength),
		13:                        WithAllBytes(0x00, Enforce, PCRMeasurementLength),
		uint32(PCRIndexClusterID): WithAllBytes(0x00, Enforce, PCRMeasurementLength),
	}
	aws_AWSSEVSNP = M{
		4:                         PlaceHolderMeasurement(PCRMeasurementLength),
		8:                         WithAllBytes(0x00, Enforce, PCRMeasurementLength),
		9:                         PlaceHolderMeasurement(PCRMeasurementLength),
		11:                        WithAllBytes(0x00, Enforce, PCRMeasurementLength),
		12:                        PlaceHolderMeasurement(PCRMeasurementLength),
		13:                        WithAllBytes(0x00, Enforce, PCRMeasurementLength),
		uint32(PCRIndexClusterID): WithAllBytes(0x00, Enforce, PCRMeasurementLength),
	}
	azure_AzureSEVSNP = M{
		4:                         PlaceHolderMeasurement(PCRMeasurementLength),
		8:                         WithAllBytes(0x00, Enforce, PCRMeasurementLength),
		9:                         PlaceHolderMeasurement(PCRMeasurementLength),
		11:                        WithAllBytes(0x00, Enforce, PCRMeasurementLength),
		12:                        PlaceHolderMeasurement(PCRMeasurementLength),
		13:                        WithAllBytes(0x00, Enforce, PCRMeasurementLength),
		uint32(PCRIndexClusterID): WithAllBytes(0x00, Enforce, PCRMeasurementLength),
	}
	azure_AzureTDX = M{
		4:                         PlaceHolderMeasurement(PCRMeasurementLength),
		8:                         WithAllBytes(0x00, Enforce, PCRMeasurementLength),
		9:                         PlaceHolderMeasurement(PCRMeasurementLength),
		11:                        WithAllBytes(0x00, Enforce, PCRMeasurementLength),
		12:                        PlaceHolderMeasurement(PCRMeasurementLength),
		13:                        WithAllBytes(0x00, Enforce, PCRMeasurementLength),
		uint32(PCRIndexClusterID): WithAllBytes(0x00, Enforce, PCRMeasurementLength),
	}
	azure_AzureTrustedLaunch = M{
		4:                         PlaceHolderMeasurement(PCRMeasurementLength),
		8:                         WithAllBytes(0x00, Enforce, PCRMeasurementLength),
		9:                         PlaceHolderMeasurement(PCRMeasurementLength),
		11:                        WithAllBytes(0x00, Enforce, PCRMeasurementLength),
		12:                        PlaceHolderMeasurement(PCRMeasurementLength),
		13:                        WithAllBytes(0x00, Enforce, PCRMeasurementLength),
		uint32(PCRIndexClusterID): WithAllBytes(0x00, Enforce, PCRMeasurementLength),
	}
	gcp_GCPSEVES = M{
		4:                         PlaceHolderMeasurement(PCRMeasurementLength),
		8:                         WithAllBytes(0x00, Enforce, PCRMeasurementLength),
		9:                         PlaceHolderMeasurement(PCRMeasurementLength),
		11:                        WithAllBytes(0x00, Enforce, PCRMeasurementLength),
		12:                        PlaceHolderMeasurement(PCRMeasurementLength),
		13:                        WithAllBytes(0x00, Enforce, PCRMeasurementLength),
		uint32(PCRIndexClusterID): WithAllBytes(0x00, Enforce, PCRMeasurementLength),
	}
	openstack_QEMUVTPM = M{
		4:                         PlaceHolderMeasurement(PCRMeasurementLength),
		8:                         WithAllBytes(0x00, Enforce, PCRMeasurementLength),
		9:                         PlaceHolderMeasurement(PCRMeasurementLength),
		11:                        WithAllBytes(0x00, Enforce, PCRMeasurementLength),
		12:                        PlaceHolderMeasurement(PCRMeasurementLength),
		13:                        WithAllBytes(0x00, Enforce, PCRMeasurementLength),
		uint32(PCRIndexClusterID): WithAllBytes(0x00, Enforce, PCRMeasurementLength),
	}
	qemu_QEMUTDX = M{
		0:                         PlaceHolderMeasurement(TDXMeasurementLength),
		1:                         PlaceHolderMeasurement(TDXMeasurementLength),
		2:                         PlaceHolderMeasurement(TDXMeasurementLength),
		uint32(TDXIndexClusterID): WithAllBytes(0x00, Enforce, TDXMeasurementLength),
		4:                         PlaceHolderMeasurement(TDXMeasurementLength),
	}
	qemu_QEMUVTPM = M{
		4:                         PlaceHolderMeasurement(PCRMeasurementLength),
		8:                         WithAllBytes(0x00, Enforce, PCRMeasurementLength),
		9:                         PlaceHolderMeasurement(PCRMeasurementLength),
		11:                        WithAllBytes(0x00, Enforce, PCRMeasurementLength),
		12:                        PlaceHolderMeasurement(PCRMeasurementLength),
		13:                        WithAllBytes(0x00, Enforce, PCRMeasurementLength),
		uint32(PCRIndexClusterID): WithAllBytes(0x00, Enforce, PCRMeasurementLength),
	}
)
