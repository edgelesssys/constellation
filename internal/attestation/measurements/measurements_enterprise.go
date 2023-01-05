//go:build enterprise

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package measurements

import "github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"

// Regenerate the measurements by running go generate.
// The enterprise build tag is required to validate the measurements using production
// sigstore certificates.
//go:generate go run -tags enterprise measurement-generator/generate.go

// DefaultsFor provides the default measurements for given cloud provider.
func DefaultsFor(provider cloudprovider.Provider) M {
	switch provider {
	case cloudprovider.AWS:
		return M{
			0: {
				Expected: [32]byte{
					0x73, 0x7f, 0x76, 0x7a, 0x12, 0xf5, 0x4e, 0x70,
					0xee, 0xcb, 0xc8, 0x68, 0x40, 0x11, 0x32, 0x3a,
					0xe2, 0xfe, 0x2d, 0xd9, 0xf9, 0x07, 0x85, 0x57,
					0x79, 0x69, 0xd7, 0xa2, 0x01, 0x3e, 0x8c, 0x12,
				},
				WarnOnly: true,
			},
			2: {
				Expected: [32]byte{
					0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea,
					0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d,
					0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a,
					0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69,
				},
				WarnOnly: true,
			},
			3: {
				Expected: [32]byte{
					0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea,
					0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d,
					0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a,
					0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69,
				},
				WarnOnly: true,
			},
			4: {
				Expected: [32]byte{
					0x58, 0x75, 0x6b, 0x91, 0x89, 0x84, 0xeb, 0x5a,
					0xbc, 0x0d, 0xf8, 0xd7, 0x76, 0xea, 0xd6, 0x1b,
					0x56, 0x39, 0xa6, 0x6c, 0x27, 0x79, 0xa4, 0x78,
					0x72, 0x69, 0xb4, 0x71, 0x5d, 0x97, 0x12, 0x7d,
				},
				WarnOnly: false,
			},
			5: {
				Expected: [32]byte{
					0x98, 0x09, 0x28, 0xe9, 0x26, 0x20, 0x78, 0x2b,
					0xff, 0xfe, 0xc4, 0x37, 0x5d, 0x44, 0x66, 0x22,
					0xc6, 0xdd, 0x5b, 0xd1, 0x98, 0x33, 0xfb, 0xda,
					0xf5, 0x58, 0x2b, 0xb9, 0x38, 0x92, 0xbc, 0xee,
				},
				WarnOnly: true,
			},
			6: {
				Expected: [32]byte{
					0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea,
					0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d,
					0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a,
					0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69,
				},
				WarnOnly: true,
			},
			7: {
				Expected: [32]byte{
					0x12, 0x0e, 0x49, 0x8d, 0xb2, 0xa2, 0x24, 0xbd,
					0x51, 0x2b, 0x6e, 0xfc, 0x9b, 0x02, 0x34, 0xf8,
					0x43, 0xe1, 0x0b, 0xf0, 0x61, 0xeb, 0x7a, 0x76,
					0xec, 0xca, 0x55, 0x09, 0xa2, 0x23, 0x89, 0x01,
				},
				WarnOnly: true,
			},
			8: {
				Expected: [32]byte{
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				},
				WarnOnly: false,
			},
			9: {
				Expected: [32]byte{
					0x56, 0x32, 0x73, 0x5e, 0xcc, 0x9d, 0x45, 0x33,
					0xf5, 0x2d, 0xc3, 0x57, 0xd1, 0xf0, 0xd8, 0x4e,
					0x86, 0xf4, 0xfd, 0xc2, 0x89, 0x09, 0x6e, 0xb2,
					0x99, 0xa8, 0x59, 0x0f, 0xd2, 0xf0, 0x44, 0x8a,
				},
				WarnOnly: false,
			},
			11: {
				Expected: [32]byte{
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				},
				WarnOnly: false,
			},
			12: {
				Expected: [32]byte{
					0xd5, 0x04, 0x41, 0xce, 0x5f, 0xea, 0xa5, 0x1a,
					0x17, 0x94, 0x2d, 0x7f, 0xe0, 0x3c, 0xce, 0x5f,
					0x25, 0x1e, 0x7f, 0x53, 0x81, 0xba, 0x78, 0x3f,
					0x55, 0x31, 0xc2, 0xf8, 0xfb, 0xb6, 0xfa, 0x15,
				},
				WarnOnly: false,
			},
			13: {
				Expected: [32]byte{
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				},
				WarnOnly: false,
			},
			14: {
				Expected: [32]byte{
					0xd7, 0xc4, 0xcc, 0x7f, 0xf7, 0x93, 0x30, 0x22,
					0xf0, 0x13, 0xe0, 0x3b, 0xde, 0xe8, 0x75, 0xb9,
					0x17, 0x20, 0xb5, 0xb8, 0x6c, 0xf1, 0x75, 0x3c,
					0xad, 0x83, 0x0f, 0x95, 0xe7, 0x91, 0x92, 0x6f,
				},
				WarnOnly: true,
			},
			15: {
				Expected: [32]byte{
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				},
				WarnOnly: false,
			},
		}
	case cloudprovider.Azure:
		return M{
			1: {
				Expected: [32]byte{
					0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea,
					0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d,
					0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a,
					0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69,
				},
				WarnOnly: true,
			},
			2: {
				Expected: [32]byte{
					0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea,
					0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d,
					0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a,
					0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69,
				},
				WarnOnly: true,
			},
			3: {
				Expected: [32]byte{
					0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea,
					0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d,
					0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a,
					0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69,
				},
				WarnOnly: true,
			},
			4: {
				Expected: [32]byte{
					0x7b, 0x33, 0x94, 0xd0, 0x81, 0x79, 0x41, 0x4b,
					0x01, 0x34, 0x14, 0x61, 0xbc, 0xcd, 0xa4, 0x5b,
					0xc6, 0x2b, 0x2d, 0x4b, 0x0e, 0x47, 0x04, 0xc8,
					0x81, 0xe8, 0x5f, 0xb6, 0xa8, 0xe8, 0xcb, 0x61,
				},
				WarnOnly: false,
			},
			5: {
				Expected: [32]byte{
					0xd2, 0x6b, 0x55, 0x6b, 0x72, 0xed, 0x82, 0xe1,
					0x27, 0x83, 0x9d, 0x96, 0xf8, 0xe0, 0x7c, 0x36,
					0xc7, 0x46, 0xa1, 0xdb, 0x4e, 0xab, 0x10, 0x2e,
					0x9d, 0x46, 0x2f, 0xbc, 0x3b, 0xa6, 0x56, 0x60,
				},
				WarnOnly: true,
			},
			7: {
				Expected: [32]byte{
					0x34, 0x65, 0x47, 0xa8, 0xce, 0x59, 0x57, 0xaf,
					0x27, 0xe5, 0x52, 0x42, 0x7d, 0x6b, 0x9e, 0x6d,
					0x9c, 0xb5, 0x02, 0xf0, 0x15, 0x6e, 0x91, 0x55,
					0x38, 0x04, 0x51, 0xee, 0xa1, 0xb3, 0xf0, 0xed,
				},
				WarnOnly: true,
			},
			8: {
				Expected: [32]byte{
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				},
				WarnOnly: false,
			},
			9: {
				Expected: [32]byte{
					0x8d, 0x0e, 0x87, 0x01, 0x0f, 0xa3, 0x2f, 0x17,
					0xc3, 0xa1, 0xbe, 0x2e, 0xd2, 0x58, 0xa6, 0xd5,
					0x29, 0x07, 0xc9, 0xc2, 0xfb, 0x0e, 0xab, 0xf1,
					0xb7, 0x10, 0x50, 0x92, 0x6c, 0xc3, 0xe6, 0x1b,
				},
				WarnOnly: false,
			},
			11: {
				Expected: [32]byte{
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				},
				WarnOnly: false,
			},
			12: {
				Expected: [32]byte{
					0x7c, 0x79, 0x3c, 0xdf, 0x4e, 0xcc, 0x2a, 0x0c,
					0xdf, 0xba, 0x01, 0x82, 0xbc, 0x3b, 0x24, 0x88,
					0xe0, 0x06, 0xbc, 0x85, 0x50, 0x58, 0x82, 0x4b,
					0x6a, 0x6e, 0xd8, 0xd8, 0x78, 0x39, 0xca, 0xe5,
				},
				WarnOnly: false,
			},
			13: {
				Expected: [32]byte{
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				},
				WarnOnly: false,
			},
			14: {
				Expected: [32]byte{
					0xd7, 0xc4, 0xcc, 0x7f, 0xf7, 0x93, 0x30, 0x22,
					0xf0, 0x13, 0xe0, 0x3b, 0xde, 0xe8, 0x75, 0xb9,
					0x17, 0x20, 0xb5, 0xb8, 0x6c, 0xf1, 0x75, 0x3c,
					0xad, 0x83, 0x0f, 0x95, 0xe7, 0x91, 0x92, 0x6f,
				},
				WarnOnly: true,
			},
			15: {
				Expected: [32]byte{
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				},
				WarnOnly: false,
			},
		}
	case cloudprovider.GCP:
		return M{
			0: {
				Expected: [32]byte{
					0x0f, 0x35, 0xc2, 0x14, 0x60, 0x8d, 0x93, 0xc7,
					0xa6, 0xe6, 0x8a, 0xe7, 0x35, 0x9b, 0x4a, 0x8b,
					0xe5, 0xa0, 0xe9, 0x9e, 0xea, 0x91, 0x07, 0xec,
					0xe4, 0x27, 0xc4, 0xde, 0xa4, 0xe4, 0x39, 0xcf,
				},
				WarnOnly: false,
			},
			1: {
				Expected: [32]byte{
					0x74, 0x5f, 0x2f, 0xb4, 0x23, 0x5e, 0x46, 0x47,
					0xaa, 0x0a, 0xd5, 0xac, 0xe7, 0x81, 0xcd, 0x92,
					0x9e, 0xb6, 0x8c, 0x28, 0x87, 0x0e, 0x7d, 0xd5,
					0xd1, 0xa1, 0x53, 0x58, 0x54, 0x32, 0x5e, 0x56,
				},
				WarnOnly: true,
			},
			2: {
				Expected: [32]byte{
					0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea,
					0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d,
					0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a,
					0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69,
				},
				WarnOnly: true,
			},
			3: {
				Expected: [32]byte{
					0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea,
					0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d,
					0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a,
					0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69,
				},
				WarnOnly: true,
			},
			4: {
				Expected: [32]byte{
					0x74, 0x0d, 0x4b, 0x09, 0x6a, 0x9b, 0x8f, 0x46,
					0x9c, 0xe2, 0x5f, 0x93, 0x62, 0xdd, 0x3f, 0xa6,
					0x88, 0x28, 0x45, 0xfd, 0xe5, 0xb4, 0x2e, 0x56,
					0xac, 0x06, 0x69, 0x11, 0x49, 0x49, 0x20, 0xbf,
				},
				WarnOnly: false,
			},
			5: {
				Expected: [32]byte{
					0x62, 0x88, 0x87, 0x63, 0x7b, 0xd1, 0x6e, 0xcd,
					0xe7, 0x9f, 0x9c, 0x79, 0x64, 0xe4, 0x12, 0xc6,
					0x88, 0x83, 0xfb, 0x57, 0x1c, 0xe6, 0xdc, 0x74,
					0xde, 0xff, 0x11, 0x8c, 0xe5, 0xaa, 0xe8, 0x35,
				},
				WarnOnly: true,
			},
			6: {
				Expected: [32]byte{
					0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea,
					0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d,
					0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a,
					0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69,
				},
				WarnOnly: true,
			},
			7: {
				Expected: [32]byte{
					0xb1, 0xe9, 0xb3, 0x05, 0x32, 0x5c, 0x51, 0xb9,
					0x3d, 0xa5, 0x8c, 0xbf, 0x7f, 0x92, 0x51, 0x2d,
					0x8e, 0xeb, 0xfa, 0x01, 0x14, 0x3e, 0x4d, 0x88,
					0x44, 0xe4, 0x0e, 0x06, 0x2e, 0x9b, 0x6c, 0xd5,
				},
				WarnOnly: true,
			},
			8: {
				Expected: [32]byte{
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				},
				WarnOnly: false,
			},
			9: {
				Expected: [32]byte{
					0x6a, 0x8d, 0xa1, 0x07, 0xe5, 0xfb, 0xca, 0x6c,
					0x00, 0x8e, 0xf8, 0x84, 0x58, 0xe3, 0x35, 0x3f,
					0x27, 0x09, 0x98, 0x45, 0xda, 0x23, 0x36, 0xc4,
					0xda, 0x59, 0x2d, 0xdb, 0x94, 0x79, 0x70, 0xc9,
				},
				WarnOnly: false,
			},
			10: {
				Expected: [32]byte{
					0x66, 0x03, 0x42, 0x8e, 0x8f, 0xab, 0x0c, 0xed,
					0x71, 0xd2, 0x64, 0xec, 0x6a, 0xa2, 0xc4, 0x36,
					0x3f, 0xb8, 0x06, 0xe4, 0xe1, 0xb2, 0x98, 0xfc,
					0xe4, 0x54, 0xd9, 0xcf, 0x4d, 0x44, 0x52, 0x15,
				},
				WarnOnly: true,
			},
			11: {
				Expected: [32]byte{
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				},
				WarnOnly: false,
			},
			12: {
				Expected: [32]byte{
					0x3a, 0x92, 0x94, 0x70, 0x27, 0x12, 0x5a, 0x16,
					0xfd, 0x78, 0x10, 0x2b, 0x5f, 0x81, 0x05, 0xad,
					0x98, 0xdc, 0x7b, 0xd8, 0x0d, 0x9d, 0x03, 0x1e,
					0xdf, 0xb2, 0x13, 0xbb, 0x50, 0xa5, 0x85, 0x5d,
				},
				WarnOnly: false,
			},
			13: {
				Expected: [32]byte{
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				},
				WarnOnly: false,
			},
			14: {
				Expected: [32]byte{
					0xd7, 0xc4, 0xcc, 0x7f, 0xf7, 0x93, 0x30, 0x22,
					0xf0, 0x13, 0xe0, 0x3b, 0xde, 0xe8, 0x75, 0xb9,
					0x17, 0x20, 0xb5, 0xb8, 0x6c, 0xf1, 0x75, 0x3c,
					0xad, 0x83, 0x0f, 0x95, 0xe7, 0x91, 0x92, 0x6f,
				},
				WarnOnly: true,
			},
			15: {
				Expected: [32]byte{
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				},
				WarnOnly: false,
			},
		}
	case cloudprovider.QEMU:
		return M{
			4: {
				Expected: [32]byte{
					0xeb, 0xb4, 0xa7, 0xb0, 0xc0, 0x67, 0x24, 0x53,
					0xdb, 0xb4, 0x11, 0x5c, 0x0d, 0xb9, 0xae, 0x8c,
					0x2c, 0x60, 0x8b, 0xaa, 0xa6, 0x82, 0x37, 0xa3,
					0xa7, 0x97, 0x44, 0xe3, 0x6c, 0xf5, 0xa9, 0x55,
				},
				WarnOnly: false,
			},
			8: {
				Expected: [32]byte{
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				},
				WarnOnly: false,
			},
			9: {
				Expected: [32]byte{
					0x1d, 0x89, 0x7a, 0xeb, 0x6a, 0xe5, 0xb3, 0x66,
					0xf2, 0xd1, 0x75, 0x15, 0xfb, 0x11, 0x3d, 0x85,
					0x12, 0x94, 0x5d, 0x4d, 0xba, 0x26, 0x26, 0xb6,
					0x92, 0xef, 0xcc, 0x59, 0x05, 0xe3, 0x55, 0x16,
				},
				WarnOnly: false,
			},
			11: {
				Expected: [32]byte{
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				},
				WarnOnly: false,
			},
			12: {
				Expected: [32]byte{
					0x51, 0xd1, 0x13, 0x57, 0xea, 0x78, 0xf4, 0x12,
					0xc9, 0xc6, 0x5d, 0xd1, 0x02, 0x25, 0x48, 0x67,
					0x4f, 0xb6, 0x50, 0x48, 0x81, 0xc0, 0xcf, 0xba,
					0x84, 0x03, 0x6c, 0xf9, 0x3b, 0x7e, 0x79, 0xaf,
				},
				WarnOnly: false,
			},
			13: {
				Expected: [32]byte{
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				},
				WarnOnly: false,
			},
			15: {
				Expected: [32]byte{
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				},
				WarnOnly: false,
			},
		}

	default:
		return nil
	}
}
