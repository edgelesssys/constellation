//go:build enterprise

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package measurements

// Regenerate the measurements by running go generate.
// The directive can be found in the file measurements.go since it does not have
// a build tag.
// The enterprise build tag is required to validate the measurements using production
// sigstore certificates.

// revive:disable:var-naming
var (
	aws_AWSNitroTPM          = M{0: {Expected: []byte{0x73, 0x7f, 0x76, 0x7a, 0x12, 0xf5, 0x4e, 0x70, 0xee, 0xcb, 0xc8, 0x68, 0x40, 0x11, 0x32, 0x3a, 0xe2, 0xfe, 0x2d, 0xd9, 0xf9, 0x07, 0x85, 0x57, 0x79, 0x69, 0xd7, 0xa2, 0x01, 0x3e, 0x8c, 0x12}, ValidationOpt: WarnOnly}, 2: {Expected: []byte{0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea, 0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d, 0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a, 0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69}, ValidationOpt: WarnOnly}, 3: {Expected: []byte{0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea, 0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d, 0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a, 0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69}, ValidationOpt: WarnOnly}, 4: {Expected: []byte{0x44, 0x76, 0x39, 0x7d, 0x45, 0x1e, 0x30, 0x7c, 0x5e, 0x91, 0xd5, 0x4c, 0xbf, 0xaa, 0xc6, 0x12, 0x40, 0x14, 0x77, 0xfd, 0x67, 0xfc, 0x58, 0x10, 0x23, 0x2d, 0xe6, 0x1c, 0xc8, 0x08, 0x3a, 0xd9}, ValidationOpt: Enforce}, 6: {Expected: []byte{0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea, 0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d, 0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a, 0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69}, ValidationOpt: WarnOnly}, 7: {Expected: []byte{0xfb, 0x71, 0xe5, 0xe5, 0x5c, 0xef, 0xba, 0x9e, 0x2b, 0x39, 0x6d, 0x17, 0x60, 0x4d, 0xe0, 0xfe, 0x6e, 0x18, 0x41, 0xa7, 0x67, 0x58, 0x85, 0x6a, 0x12, 0x08, 0x33, 0xe3, 0xad, 0x1c, 0x40, 0xa3}, ValidationOpt: WarnOnly}, 8: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 9: {Expected: []byte{0x62, 0xbf, 0xf5, 0x8e, 0xfb, 0xbc, 0x30, 0xe9, 0x89, 0xb8, 0xc5, 0xe7, 0xc2, 0x40, 0x25, 0x4a, 0xce, 0x18, 0x1b, 0x0d, 0xb4, 0x7b, 0xfd, 0x60, 0x4a, 0xba, 0xee, 0xa9, 0xdb, 0xaf, 0xd8, 0xce}, ValidationOpt: Enforce}, 11: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 12: {Expected: []byte{0xe3, 0xf7, 0x9e, 0x78, 0xb8, 0x08, 0x45, 0xa9, 0x5a, 0x61, 0x0c, 0x00, 0x51, 0xba, 0xd0, 0x6d, 0x84, 0xb1, 0x1b, 0x60, 0x79, 0x37, 0xd7, 0x31, 0x83, 0x07, 0xe4, 0x0a, 0xea, 0xac, 0xaf, 0x1e}, ValidationOpt: Enforce}, 13: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 14: {Expected: []byte{0xd7, 0xc4, 0xcc, 0x7f, 0xf7, 0x93, 0x30, 0x22, 0xf0, 0x13, 0xe0, 0x3b, 0xde, 0xe8, 0x75, 0xb9, 0x17, 0x20, 0xb5, 0xb8, 0x6c, 0xf1, 0x75, 0x3c, 0xad, 0x83, 0x0f, 0x95, 0xe7, 0x91, 0x92, 0x6f}, ValidationOpt: WarnOnly}, 15: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}}
	aws_AWSSEVSNP            = M{0: {Expected: []byte{0x73, 0x7f, 0x76, 0x7a, 0x12, 0xf5, 0x4e, 0x70, 0xee, 0xcb, 0xc8, 0x68, 0x40, 0x11, 0x32, 0x3a, 0xe2, 0xfe, 0x2d, 0xd9, 0xf9, 0x07, 0x85, 0x57, 0x79, 0x69, 0xd7, 0xa2, 0x01, 0x3e, 0x8c, 0x12}, ValidationOpt: WarnOnly}, 2: {Expected: []byte{0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea, 0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d, 0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a, 0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69}, ValidationOpt: WarnOnly}, 3: {Expected: []byte{0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea, 0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d, 0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a, 0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69}, ValidationOpt: WarnOnly}, 4: {Expected: []byte{0xba, 0x0e, 0xe0, 0x01, 0x39, 0xbf, 0xbf, 0x6f, 0x3a, 0x6f, 0xde, 0x5b, 0x4c, 0x33, 0xf3, 0x28, 0x72, 0x0a, 0xb7, 0x8a, 0xe8, 0x61, 0xf2, 0x03, 0x63, 0xa9, 0xda, 0x71, 0x4c, 0x67, 0x40, 0x76}, ValidationOpt: Enforce}, 6: {Expected: []byte{0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea, 0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d, 0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a, 0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69}, ValidationOpt: WarnOnly}, 7: {Expected: []byte{0xfb, 0x71, 0xe5, 0xe5, 0x5c, 0xef, 0xba, 0x9e, 0x2b, 0x39, 0x6d, 0x17, 0x60, 0x4d, 0xe0, 0xfe, 0x6e, 0x18, 0x41, 0xa7, 0x67, 0x58, 0x85, 0x6a, 0x12, 0x08, 0x33, 0xe3, 0xad, 0x1c, 0x40, 0xa3}, ValidationOpt: WarnOnly}, 8: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 9: {Expected: []byte{0x62, 0xbf, 0xf5, 0x8e, 0xfb, 0xbc, 0x30, 0xe9, 0x89, 0xb8, 0xc5, 0xe7, 0xc2, 0x40, 0x25, 0x4a, 0xce, 0x18, 0x1b, 0x0d, 0xb4, 0x7b, 0xfd, 0x60, 0x4a, 0xba, 0xee, 0xa9, 0xdb, 0xaf, 0xd8, 0xce}, ValidationOpt: Enforce}, 11: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 12: {Expected: []byte{0x34, 0x26, 0x77, 0x15, 0xca, 0x77, 0x7a, 0xa6, 0xaf, 0x62, 0x44, 0x8d, 0x3b, 0x2c, 0x11, 0x01, 0x86, 0x2f, 0xfc, 0x04, 0x9f, 0xda, 0x8d, 0x2e, 0x32, 0x34, 0x87, 0x8b, 0xa9, 0x66, 0x83, 0x90}, ValidationOpt: Enforce}, 13: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 14: {Expected: []byte{0xd7, 0xc4, 0xcc, 0x7f, 0xf7, 0x93, 0x30, 0x22, 0xf0, 0x13, 0xe0, 0x3b, 0xde, 0xe8, 0x75, 0xb9, 0x17, 0x20, 0xb5, 0xb8, 0x6c, 0xf1, 0x75, 0x3c, 0xad, 0x83, 0x0f, 0x95, 0xe7, 0x91, 0x92, 0x6f}, ValidationOpt: WarnOnly}, 15: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}}
	azure_AzureSEVSNP        = M{1: {Expected: []byte{0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea, 0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d, 0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a, 0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69}, ValidationOpt: WarnOnly}, 2: {Expected: []byte{0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea, 0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d, 0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a, 0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69}, ValidationOpt: WarnOnly}, 3: {Expected: []byte{0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea, 0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d, 0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a, 0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69}, ValidationOpt: WarnOnly}, 4: {Expected: []byte{0x63, 0x2b, 0xb2, 0x90, 0xb5, 0xca, 0xbf, 0x56, 0x3d, 0xcf, 0x98, 0x65, 0xfd, 0xba, 0x9c, 0xd4, 0x26, 0x8f, 0x13, 0x37, 0x32, 0xfa, 0xcf, 0x2e, 0xde, 0xc9, 0xa6, 0x52, 0xf2, 0x9b, 0x2d, 0x1a}, ValidationOpt: Enforce}, 7: {Expected: []byte{0x34, 0x65, 0x47, 0xa8, 0xce, 0x59, 0x57, 0xaf, 0x27, 0xe5, 0x52, 0x42, 0x7d, 0x6b, 0x9e, 0x6d, 0x9c, 0xb5, 0x02, 0xf0, 0x15, 0x6e, 0x91, 0x55, 0x38, 0x04, 0x51, 0xee, 0xa1, 0xb3, 0xf0, 0xed}, ValidationOpt: WarnOnly}, 8: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 9: {Expected: []byte{0xcb, 0x5d, 0xbc, 0xdf, 0xc8, 0x29, 0x1a, 0x8e, 0x07, 0x42, 0xf7, 0x9f, 0xfe, 0x4e, 0xf5, 0xa7, 0x56, 0xc2, 0x51, 0x3a, 0x6e, 0xaa, 0x4a, 0x7b, 0x99, 0x54, 0x86, 0x32, 0x92, 0xd8, 0x98, 0x15}, ValidationOpt: Enforce}, 11: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 12: {Expected: []byte{0x46, 0xd1, 0x1b, 0x42, 0xc1, 0x96, 0xf2, 0x00, 0x44, 0xe2, 0xf4, 0x64, 0x63, 0x1d, 0x91, 0xfb, 0x58, 0x54, 0x63, 0xcc, 0x39, 0xa2, 0x36, 0x00, 0x45, 0xda, 0x73, 0x81, 0xa7, 0x07, 0x0a, 0xc8}, ValidationOpt: Enforce}, 13: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 14: {Expected: []byte{0xd7, 0xc4, 0xcc, 0x7f, 0xf7, 0x93, 0x30, 0x22, 0xf0, 0x13, 0xe0, 0x3b, 0xde, 0xe8, 0x75, 0xb9, 0x17, 0x20, 0xb5, 0xb8, 0x6c, 0xf1, 0x75, 0x3c, 0xad, 0x83, 0x0f, 0x95, 0xe7, 0x91, 0x92, 0x6f}, ValidationOpt: WarnOnly}, 15: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}}
	azure_AzureTrustedLaunch M
	gcp_GCPSEVES             = M{1: {Expected: []byte{0x74, 0x5f, 0x2f, 0xb4, 0x23, 0x5e, 0x46, 0x47, 0xaa, 0x0a, 0xd5, 0xac, 0xe7, 0x81, 0xcd, 0x92, 0x9e, 0xb6, 0x8c, 0x28, 0x87, 0x0e, 0x7d, 0xd5, 0xd1, 0xa1, 0x53, 0x58, 0x54, 0x32, 0x5e, 0x56}, ValidationOpt: WarnOnly}, 2: {Expected: []byte{0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea, 0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d, 0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a, 0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69}, ValidationOpt: WarnOnly}, 3: {Expected: []byte{0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea, 0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d, 0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a, 0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69}, ValidationOpt: WarnOnly}, 4: {Expected: []byte{0x8e, 0xa3, 0xee, 0x10, 0x73, 0x22, 0x9e, 0xef, 0x59, 0x57, 0x10, 0x22, 0xc8, 0x2f, 0xab, 0xeb, 0x86, 0xfa, 0xef, 0x8c, 0x0d, 0xac, 0x09, 0x5d, 0x27, 0x07, 0xaa, 0xe5, 0x8d, 0x5d, 0xa6, 0x17}, ValidationOpt: Enforce}, 6: {Expected: []byte{0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea, 0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d, 0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a, 0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69}, ValidationOpt: WarnOnly}, 7: {Expected: []byte{0xa1, 0xd1, 0x93, 0xdb, 0xfc, 0x3d, 0xa1, 0xa5, 0xe9, 0x3f, 0xe7, 0xb1, 0x38, 0x44, 0x27, 0xfb, 0x78, 0xfe, 0xef, 0xfc, 0xb0, 0x66, 0x75, 0xa0, 0xcf, 0x84, 0x0e, 0xc9, 0x94, 0x06, 0xf2, 0x37}, ValidationOpt: WarnOnly}, 8: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 9: {Expected: []byte{0x62, 0xbf, 0xf5, 0x8e, 0xfb, 0xbc, 0x30, 0xe9, 0x89, 0xb8, 0xc5, 0xe7, 0xc2, 0x40, 0x25, 0x4a, 0xce, 0x18, 0x1b, 0x0d, 0xb4, 0x7b, 0xfd, 0x60, 0x4a, 0xba, 0xee, 0xa9, 0xdb, 0xaf, 0xd8, 0xce}, ValidationOpt: Enforce}, 11: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 12: {Expected: []byte{0x7c, 0x62, 0x9a, 0x1d, 0x08, 0xe3, 0x69, 0xae, 0xab, 0xcc, 0x14, 0x72, 0xad, 0x8d, 0x03, 0xc9, 0xbf, 0x26, 0xe9, 0xfa, 0x7c, 0x03, 0xbd, 0x7f, 0xee, 0x45, 0x22, 0xaa, 0x2c, 0xb4, 0xc4, 0x47}, ValidationOpt: Enforce}, 13: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 14: {Expected: []byte{0xd7, 0xc4, 0xcc, 0x7f, 0xf7, 0x93, 0x30, 0x22, 0xf0, 0x13, 0xe0, 0x3b, 0xde, 0xe8, 0x75, 0xb9, 0x17, 0x20, 0xb5, 0xb8, 0x6c, 0xf1, 0x75, 0x3c, 0xad, 0x83, 0x0f, 0x95, 0xe7, 0x91, 0x92, 0x6f}, ValidationOpt: WarnOnly}, 15: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}}
	qemu_QEMUTDX             M
	qemu_QEMUVTPM            = M{4: {Expected: []byte{0x13, 0x4a, 0xef, 0x0f, 0xe1, 0x8a, 0x1c, 0x94, 0x95, 0x4c, 0x86, 0x03, 0x4f, 0x56, 0x7c, 0xb6, 0xc3, 0x20, 0x31, 0xe8, 0xc1, 0x6a, 0x72, 0x7e, 0xc0, 0x02, 0x60, 0xee, 0x07, 0x81, 0xe1, 0x9d}, ValidationOpt: Enforce}, 8: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 9: {Expected: []byte{0x62, 0xbf, 0xf5, 0x8e, 0xfb, 0xbc, 0x30, 0xe9, 0x89, 0xb8, 0xc5, 0xe7, 0xc2, 0x40, 0x25, 0x4a, 0xce, 0x18, 0x1b, 0x0d, 0xb4, 0x7b, 0xfd, 0x60, 0x4a, 0xba, 0xee, 0xa9, 0xdb, 0xaf, 0xd8, 0xce}, ValidationOpt: Enforce}, 11: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 12: {Expected: []byte{0xe4, 0xb5, 0x81, 0xd3, 0xdd, 0x9f, 0x86, 0xdd, 0xd3, 0x6b, 0x28, 0x77, 0x48, 0xec, 0x29, 0x47, 0x10, 0x2c, 0x1a, 0xad, 0xe5, 0x87, 0x73, 0x7d, 0x2f, 0x7b, 0x1d, 0x4d, 0xca, 0x00, 0xdc, 0xd8}, ValidationOpt: Enforce}, 13: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 15: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}}
)
