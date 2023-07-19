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
	aws_AWSNitroTPM          = M{0: {Expected: []byte{0x73, 0x7f, 0x76, 0x7a, 0x12, 0xf5, 0x4e, 0x70, 0xee, 0xcb, 0xc8, 0x68, 0x40, 0x11, 0x32, 0x3a, 0xe2, 0xfe, 0x2d, 0xd9, 0xf9, 0x07, 0x85, 0x57, 0x79, 0x69, 0xd7, 0xa2, 0x01, 0x3e, 0x8c, 0x12}, ValidationOpt: WarnOnly}, 2: {Expected: []byte{0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea, 0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d, 0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a, 0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69}, ValidationOpt: WarnOnly}, 3: {Expected: []byte{0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea, 0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d, 0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a, 0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69}, ValidationOpt: WarnOnly}, 4: {Expected: []byte{0xb4, 0x38, 0xdc, 0xf4, 0x8c, 0xd3, 0xa6, 0x61, 0x53, 0xe2, 0x8b, 0xa6, 0x5b, 0xb9, 0x35, 0x00, 0x8a, 0xa2, 0xa5, 0xe3, 0xbd, 0xc4, 0xbf, 0xe1, 0x83, 0x69, 0x42, 0x60, 0xde, 0x94, 0x1b, 0x08}, ValidationOpt: Enforce}, 6: {Expected: []byte{0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea, 0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d, 0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a, 0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69}, ValidationOpt: WarnOnly}, 8: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 9: {Expected: []byte{0x88, 0x9b, 0x24, 0xb4, 0xa1, 0x9d, 0x93, 0x72, 0x85, 0x32, 0xe6, 0x76, 0x6a, 0x0b, 0xb2, 0x8b, 0x23, 0xe8, 0x9b, 0x50, 0xfc, 0xa8, 0x40, 0x04, 0xe2, 0xfe, 0xc8, 0x13, 0x78, 0x26, 0xbe, 0x00}, ValidationOpt: Enforce}, 11: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 12: {Expected: []byte{0x3a, 0x4b, 0x6d, 0x79, 0x79, 0x08, 0x00, 0xd0, 0x4e, 0x07, 0x52, 0xd3, 0x84, 0x1b, 0xe0, 0xf8, 0xad, 0x84, 0x50, 0xa6, 0x06, 0x63, 0x66, 0x37, 0x2b, 0x5d, 0xab, 0xc5, 0xe9, 0x7a, 0xf4, 0xf7}, ValidationOpt: Enforce}, 13: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 14: {Expected: []byte{0xd7, 0xc4, 0xcc, 0x7f, 0xf7, 0x93, 0x30, 0x22, 0xf0, 0x13, 0xe0, 0x3b, 0xde, 0xe8, 0x75, 0xb9, 0x17, 0x20, 0xb5, 0xb8, 0x6c, 0xf1, 0x75, 0x3c, 0xad, 0x83, 0x0f, 0x95, 0xe7, 0x91, 0x92, 0x6f}, ValidationOpt: WarnOnly}, 15: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}}
	aws_AWSSEVSNP            = M{0: {Expected: []byte{0x7b, 0x06, 0x8c, 0x0c, 0x3a, 0xc2, 0x9a, 0xfe, 0x26, 0x41, 0x34, 0x53, 0x6b, 0x9b, 0xe2, 0x6f, 0x1d, 0x4c, 0xcd, 0x57, 0x5b, 0x88, 0xd3, 0xc3, 0xce, 0xab, 0xf3, 0x6a, 0xc9, 0x9c, 0x02, 0x78}, ValidationOpt: WarnOnly}, 2: {Expected: []byte{0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea, 0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d, 0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a, 0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69}, ValidationOpt: WarnOnly}, 3: {Expected: []byte{0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea, 0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d, 0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a, 0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69}, ValidationOpt: WarnOnly}, 4: {Expected: []byte{0xb1, 0x39, 0xe0, 0x25, 0xd8, 0xd9, 0x53, 0xcc, 0x12, 0xd0, 0x6b, 0x25, 0x4d, 0xd5, 0x08, 0x6f, 0xea, 0xb1, 0x91, 0x97, 0xee, 0x46, 0x50, 0xc1, 0x3b, 0x39, 0xff, 0x7e, 0xd8, 0x4d, 0xbb, 0xff}, ValidationOpt: Enforce}, 6: {Expected: []byte{0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea, 0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d, 0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a, 0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69}, ValidationOpt: WarnOnly}, 8: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 9: {Expected: []byte{0x88, 0x9b, 0x24, 0xb4, 0xa1, 0x9d, 0x93, 0x72, 0x85, 0x32, 0xe6, 0x76, 0x6a, 0x0b, 0xb2, 0x8b, 0x23, 0xe8, 0x9b, 0x50, 0xfc, 0xa8, 0x40, 0x04, 0xe2, 0xfe, 0xc8, 0x13, 0x78, 0x26, 0xbe, 0x00}, ValidationOpt: Enforce}, 11: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 12: {Expected: []byte{0xe0, 0xdb, 0x22, 0xa9, 0xaf, 0xa9, 0x4e, 0x7d, 0x3c, 0x04, 0x34, 0xf3, 0x4a, 0x76, 0x78, 0x98, 0x27, 0xd4, 0x82, 0x06, 0x4a, 0x92, 0xd0, 0xba, 0x65, 0xf4, 0x5b, 0x0b, 0xba, 0x71, 0x8f, 0xb3}, ValidationOpt: Enforce}, 13: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 14: {Expected: []byte{0xd7, 0xc4, 0xcc, 0x7f, 0xf7, 0x93, 0x30, 0x22, 0xf0, 0x13, 0xe0, 0x3b, 0xde, 0xe8, 0x75, 0xb9, 0x17, 0x20, 0xb5, 0xb8, 0x6c, 0xf1, 0x75, 0x3c, 0xad, 0x83, 0x0f, 0x95, 0xe7, 0x91, 0x92, 0x6f}, ValidationOpt: WarnOnly}, 15: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}}
	azure_AzureSEVSNP        = M{1: {Expected: []byte{0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea, 0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d, 0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a, 0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69}, ValidationOpt: WarnOnly}, 2: {Expected: []byte{0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea, 0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d, 0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a, 0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69}, ValidationOpt: WarnOnly}, 3: {Expected: []byte{0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea, 0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d, 0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a, 0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69}, ValidationOpt: WarnOnly}, 4: {Expected: []byte{0x97, 0x51, 0x65, 0x16, 0x3c, 0xac, 0xef, 0x41, 0x8a, 0xdb, 0xd5, 0x5e, 0x5c, 0x59, 0xc2, 0x4f, 0xdd, 0x35, 0x37, 0x48, 0x38, 0xcb, 0xa3, 0xae, 0xa0, 0x5e, 0xb6, 0x85, 0x83, 0xdb, 0x97, 0x1b}, ValidationOpt: Enforce}, 8: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 9: {Expected: []byte{0x6f, 0x92, 0xe4, 0xf8, 0x27, 0x79, 0x52, 0x99, 0xc4, 0x50, 0xcc, 0x70, 0xeb, 0xce, 0xa8, 0xbb, 0x7c, 0x66, 0x0d, 0x28, 0x6f, 0x49, 0xda, 0xd0, 0x3e, 0x44, 0xaf, 0x98, 0x6a, 0x70, 0x17, 0x6e}, ValidationOpt: Enforce}, 11: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 12: {Expected: []byte{0x96, 0xc5, 0x0c, 0x36, 0x50, 0x5b, 0xa4, 0x15, 0x1a, 0x81, 0xe5, 0x52, 0xe6, 0x88, 0x95, 0xed, 0xa7, 0x49, 0x9d, 0xd5, 0x40, 0x9a, 0x8f, 0x38, 0xd2, 0xd0, 0x5b, 0x45, 0x85, 0x52, 0x85, 0xae}, ValidationOpt: Enforce}, 13: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 14: {Expected: []byte{0xd7, 0xc4, 0xcc, 0x7f, 0xf7, 0x93, 0x30, 0x22, 0xf0, 0x13, 0xe0, 0x3b, 0xde, 0xe8, 0x75, 0xb9, 0x17, 0x20, 0xb5, 0xb8, 0x6c, 0xf1, 0x75, 0x3c, 0xad, 0x83, 0x0f, 0x95, 0xe7, 0x91, 0x92, 0x6f}, ValidationOpt: WarnOnly}, 15: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}}
	azure_AzureTrustedLaunch M
	gcp_GCPSEVES             = M{1: {Expected: []byte{0x74, 0x5f, 0x2f, 0xb4, 0x23, 0x5e, 0x46, 0x47, 0xaa, 0x0a, 0xd5, 0xac, 0xe7, 0x81, 0xcd, 0x92, 0x9e, 0xb6, 0x8c, 0x28, 0x87, 0x0e, 0x7d, 0xd5, 0xd1, 0xa1, 0x53, 0x58, 0x54, 0x32, 0x5e, 0x56}, ValidationOpt: WarnOnly}, 2: {Expected: []byte{0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea, 0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d, 0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a, 0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69}, ValidationOpt: WarnOnly}, 3: {Expected: []byte{0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea, 0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d, 0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a, 0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69}, ValidationOpt: WarnOnly}, 4: {Expected: []byte{0x29, 0x9d, 0x9c, 0x9f, 0x11, 0xad, 0x92, 0x79, 0x4f, 0x9a, 0x7e, 0x26, 0xee, 0xf9, 0xd8, 0x5a, 0x29, 0x03, 0xc2, 0x16, 0x9b, 0x45, 0xdd, 0x83, 0x1f, 0x32, 0xd8, 0x97, 0x15, 0x05, 0xea, 0x98}, ValidationOpt: Enforce}, 6: {Expected: []byte{0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea, 0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d, 0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a, 0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69}, ValidationOpt: WarnOnly}, 8: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 9: {Expected: []byte{0xed, 0xe7, 0xef, 0x32, 0x31, 0x47, 0x66, 0xef, 0x4f, 0x41, 0xde, 0xe0, 0xf0, 0xae, 0x88, 0x8a, 0xa5, 0x64, 0x86, 0x36, 0x03, 0x26, 0x33, 0xa0, 0xb4, 0x23, 0xb4, 0x7b, 0x7a, 0xfc, 0x5f, 0xf2}, ValidationOpt: Enforce}, 11: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 12: {Expected: []byte{0x0c, 0x43, 0xb2, 0x3e, 0x3f, 0xc5, 0x32, 0x5b, 0x24, 0x5b, 0xcc, 0xec, 0xe6, 0xc5, 0xee, 0x46, 0x1b, 0xa8, 0xc9, 0xa2, 0x37, 0xa2, 0x94, 0x39, 0x51, 0x72, 0x0a, 0xb5, 0x2c, 0xe8, 0xd0, 0x63}, ValidationOpt: Enforce}, 13: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 14: {Expected: []byte{0xd7, 0xc4, 0xcc, 0x7f, 0xf7, 0x93, 0x30, 0x22, 0xf0, 0x13, 0xe0, 0x3b, 0xde, 0xe8, 0x75, 0xb9, 0x17, 0x20, 0xb5, 0xb8, 0x6c, 0xf1, 0x75, 0x3c, 0xad, 0x83, 0x0f, 0x95, 0xe7, 0x91, 0x92, 0x6f}, ValidationOpt: WarnOnly}, 15: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}}
	qemu_QEMUTDX             M
	qemu_QEMUVTPM            = M{4: {Expected: []byte{0x98, 0xc3, 0x08, 0xf4, 0xb4, 0x7d, 0x83, 0xcc, 0xed, 0x4b, 0x0a, 0x1c, 0x16, 0x90, 0xfa, 0x8d, 0x2f, 0x5c, 0x07, 0x86, 0xfb, 0xa8, 0x85, 0x13, 0x1f, 0x05, 0x36, 0x8b, 0x16, 0x26, 0xae, 0x95}, ValidationOpt: Enforce}, 8: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 9: {Expected: []byte{0xed, 0xe7, 0xef, 0x32, 0x31, 0x47, 0x66, 0xef, 0x4f, 0x41, 0xde, 0xe0, 0xf0, 0xae, 0x88, 0x8a, 0xa5, 0x64, 0x86, 0x36, 0x03, 0x26, 0x33, 0xa0, 0xb4, 0x23, 0xb4, 0x7b, 0x7a, 0xfc, 0x5f, 0xf2}, ValidationOpt: Enforce}, 11: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 12: {Expected: []byte{0xe3, 0x41, 0x54, 0x64, 0xb0, 0x06, 0x49, 0x13, 0x53, 0x76, 0xac, 0xc2, 0x28, 0x82, 0xf3, 0xfb, 0x73, 0xef, 0x5c, 0xd9, 0x20, 0xc5, 0xe7, 0x43, 0xb7, 0x00, 0xd3, 0xc2, 0xe0, 0xf6, 0xf6, 0xa6}, ValidationOpt: Enforce}, 13: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 15: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}}
)
