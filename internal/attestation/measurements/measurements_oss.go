//go:build !enterprise

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package measurements

import "github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"

// DefaultsFor provides the default measurements for given cloud provider.
func DefaultsFor(provider cloudprovider.Provider) M {
	switch provider {
	case cloudprovider.AWS:
		return M{
			4:                         PlaceHolderMeasurement(),
			8:                         WithAllBytes(0x00, Enforce),
			9:                         PlaceHolderMeasurement(),
			11:                        WithAllBytes(0x00, Enforce),
			12:                        PlaceHolderMeasurement(),
			13:                        WithAllBytes(0x00, Enforce),
			uint32(PCRIndexClusterID): WithAllBytes(0x00, Enforce),
		}
	case cloudprovider.Azure:
		return M{
			4:                         PlaceHolderMeasurement(),
			8:                         WithAllBytes(0x00, Enforce),
			9:                         PlaceHolderMeasurement(),
			11:                        WithAllBytes(0x00, Enforce),
			12:                        PlaceHolderMeasurement(),
			13:                        WithAllBytes(0x00, Enforce),
			uint32(PCRIndexClusterID): WithAllBytes(0x00, Enforce),
		}
	case cloudprovider.GCP:
		return M{
			4:                         PlaceHolderMeasurement(),
			8:                         WithAllBytes(0x00, Enforce),
			9:                         PlaceHolderMeasurement(),
			11:                        WithAllBytes(0x00, Enforce),
			12:                        PlaceHolderMeasurement(),
			13:                        WithAllBytes(0x00, Enforce),
			uint32(PCRIndexClusterID): WithAllBytes(0x00, Enforce),
		}
	case cloudprovider.QEMU:
		return M{
			4:                         PlaceHolderMeasurement(),
			8:                         WithAllBytes(0x00, Enforce),
			9:                         PlaceHolderMeasurement(),
			11:                        WithAllBytes(0x00, Enforce),
			12:                        PlaceHolderMeasurement(),
			13:                        WithAllBytes(0x00, Enforce),
			uint32(PCRIndexClusterID): WithAllBytes(0x00, Enforce),
		}
	case cloudprovider.OpenStack:
		return M{
			4:                         PlaceHolderMeasurement(),
			8:                         WithAllBytes(0x00, Enforce),
			9:                         PlaceHolderMeasurement(),
			11:                        WithAllBytes(0x00, Enforce),
			12:                        PlaceHolderMeasurement(),
			13:                        WithAllBytes(0x00, Enforce),
			uint32(PCRIndexClusterID): WithAllBytes(0x00, Enforce),
		}
	default:
		return nil
	}
}
