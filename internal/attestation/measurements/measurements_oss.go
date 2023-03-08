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
			8:                         WithAllBytes(0x00, false),
			9:                         PlaceHolderMeasurement(),
			11:                        WithAllBytes(0x00, false),
			12:                        PlaceHolderMeasurement(),
			13:                        WithAllBytes(0x00, false),
			uint32(PCRIndexClusterID): WithAllBytes(0x00, false),
		}
	case cloudprovider.Azure:
		return M{
			4:                         PlaceHolderMeasurement(),
			8:                         WithAllBytes(0x00, false),
			9:                         PlaceHolderMeasurement(),
			11:                        WithAllBytes(0x00, false),
			12:                        PlaceHolderMeasurement(),
			13:                        WithAllBytes(0x00, false),
			uint32(PCRIndexClusterID): WithAllBytes(0x00, false),
		}
	case cloudprovider.GCP:
		return M{
			0: {
				Expected: []byte{0x0F, 0x35, 0xC2, 0x14, 0x60, 0x8D, 0x93, 0xC7, 0xA6, 0xE6, 0x8A, 0xE7, 0x35, 0x9B, 0x4A, 0x8B, 0xE5, 0xA0, 0xE9, 0x9E, 0xEA, 0x91, 0x07, 0xEC, 0xE4, 0x27, 0xC4, 0xDE, 0xA4, 0xE4, 0x39, 0xCF},
				WarnOnly: false,
			},
			4:                         PlaceHolderMeasurement(),
			8:                         WithAllBytes(0x00, false),
			9:                         PlaceHolderMeasurement(),
			11:                        WithAllBytes(0x00, false),
			12:                        PlaceHolderMeasurement(),
			13:                        WithAllBytes(0x00, false),
			uint32(PCRIndexClusterID): WithAllBytes(0x00, false),
		}
	case cloudprovider.QEMU:
		return M{
			4:                         PlaceHolderMeasurement(),
			8:                         WithAllBytes(0x00, false),
			9:                         PlaceHolderMeasurement(),
			11:                        WithAllBytes(0x00, false),
			12:                        PlaceHolderMeasurement(),
			13:                        WithAllBytes(0x00, false),
			uint32(PCRIndexClusterID): WithAllBytes(0x00, false),
		}
	default:
		return nil
	}
}
