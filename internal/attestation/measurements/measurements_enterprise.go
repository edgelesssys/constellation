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
//
// To add measurements for a new variant, add a new entry as `<csp>_<variant> = M{}` and run the generate tool.
// Entries defined as `<csp>_<variant> M` are ignored.

// revive:disable:var-naming
var (
	aws_AWSNitroTPM          = M{0: {Expected: []byte{0x73, 0x7f, 0x76, 0x7a, 0x12, 0xf5, 0x4e, 0x70, 0xee, 0xcb, 0xc8, 0x68, 0x40, 0x11, 0x32, 0x3a, 0xe2, 0xfe, 0x2d, 0xd9, 0xf9, 0x07, 0x85, 0x57, 0x79, 0x69, 0xd7, 0xa2, 0x01, 0x3e, 0x8c, 0x12}, ValidationOpt: WarnOnly}, 2: {Expected: []byte{0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea, 0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d, 0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a, 0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69}, ValidationOpt: WarnOnly}, 3: {Expected: []byte{0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea, 0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d, 0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a, 0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69}, ValidationOpt: WarnOnly}, 4: {Expected: []byte{0xd8, 0xed, 0x88, 0x2d, 0x9f, 0xbd, 0x13, 0x77, 0xfa, 0x67, 0xa3, 0x11, 0xd5, 0xf5, 0xf3, 0x5f, 0xd5, 0x44, 0xee, 0xe9, 0x9f, 0x60, 0x2c, 0xed, 0x6d, 0x72, 0xd7, 0xa1, 0x04, 0x7d, 0x2a, 0xd8}, ValidationOpt: Enforce}, 6: {Expected: []byte{0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea, 0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d, 0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a, 0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69}, ValidationOpt: WarnOnly}, 8: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 9: {Expected: []byte{0x45, 0xe7, 0xd6, 0x7e, 0x7f, 0xb0, 0x18, 0xca, 0x76, 0xcc, 0x27, 0x8b, 0x7a, 0xd7, 0x15, 0x0e, 0xa8, 0x9d, 0x2a, 0xf0, 0x37, 0xc9, 0x9d, 0x43, 0x95, 0x9a, 0xaf, 0x46, 0xe2, 0x0f, 0x62, 0xec}, ValidationOpt: Enforce}, 11: {Expected: []byte{0xdf, 0x18, 0x72, 0x7c, 0x1e, 0xb5, 0x2c, 0xf2, 0xb7, 0xc0, 0x30, 0x22, 0x3b, 0xe6, 0x5b, 0x6e, 0x2a, 0xb8, 0xb0, 0x6d, 0x20, 0x4f, 0xb9, 0xa6, 0xee, 0x12, 0xa2, 0x4d, 0xa7, 0x7b, 0x84, 0xc9}, ValidationOpt: Enforce}, 12: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 13: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 14: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: WarnOnly}, 15: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}}
	aws_AWSSEVSNP            = M{0: {Expected: []byte{0xd6, 0xdf, 0x85, 0x53, 0x58, 0xf5, 0xb1, 0x0f, 0x06, 0xf0, 0xfa, 0xb3, 0xf4, 0x08, 0xad, 0x26, 0xcd, 0x16, 0x5a, 0x29, 0x49, 0xba, 0xd6, 0x9e, 0x2c, 0xc7, 0x56, 0x92, 0x52, 0x9e, 0x66, 0x2a}, ValidationOpt: WarnOnly}, 2: {Expected: []byte{0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea, 0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d, 0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a, 0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69}, ValidationOpt: WarnOnly}, 3: {Expected: []byte{0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea, 0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d, 0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a, 0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69}, ValidationOpt: WarnOnly}, 4: {Expected: []byte{0xe7, 0x16, 0xc8, 0x9f, 0x32, 0xbe, 0x42, 0x40, 0x2c, 0x48, 0x64, 0x6d, 0xc0, 0xd0, 0x49, 0x6b, 0x06, 0x69, 0x5b, 0xfc, 0xb4, 0xda, 0x79, 0xbf, 0x23, 0xa7, 0xee, 0x5f, 0x32, 0x5f, 0x69, 0xec}, ValidationOpt: Enforce}, 6: {Expected: []byte{0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea, 0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d, 0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a, 0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69}, ValidationOpt: WarnOnly}, 8: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 9: {Expected: []byte{0x3b, 0x48, 0xbf, 0x17, 0x24, 0x9b, 0xc6, 0x1a, 0x83, 0xe5, 0x99, 0x16, 0x9a, 0xbe, 0x95, 0xa8, 0x3f, 0xf5, 0x86, 0xd8, 0xf2, 0xc1, 0x98, 0x49, 0xba, 0xec, 0x63, 0x2d, 0x13, 0x83, 0xe5, 0xbf}, ValidationOpt: Enforce}, 11: {Expected: []byte{0xf7, 0x0d, 0xc7, 0x5e, 0xd8, 0x9b, 0x88, 0x53, 0x40, 0x91, 0xd6, 0xa1, 0x91, 0xde, 0x16, 0x5c, 0x48, 0x54, 0xa8, 0x24, 0xd9, 0x6b, 0xb4, 0x66, 0xe8, 0xa4, 0x05, 0xba, 0x55, 0x46, 0xb7, 0x52}, ValidationOpt: Enforce}, 12: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 13: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 14: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: WarnOnly}, 15: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}}
	azure_AzureSEVSNP        = M{1: {Expected: []byte{0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea, 0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d, 0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a, 0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69}, ValidationOpt: WarnOnly}, 2: {Expected: []byte{0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea, 0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d, 0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a, 0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69}, ValidationOpt: WarnOnly}, 3: {Expected: []byte{0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea, 0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d, 0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a, 0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69}, ValidationOpt: WarnOnly}, 4: {Expected: []byte{0xe5, 0x23, 0x66, 0x0a, 0xc1, 0x81, 0x6c, 0x01, 0x04, 0x62, 0xd3, 0x57, 0x2d, 0xb3, 0x8b, 0xc0, 0xfa, 0x80, 0xb3, 0xb1, 0x06, 0x9b, 0x79, 0x65, 0x93, 0xc4, 0xd5, 0x7d, 0xe1, 0x14, 0x90, 0x68}, ValidationOpt: Enforce}, 8: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 9: {Expected: []byte{0xe5, 0xa2, 0x2e, 0x1f, 0xe8, 0x16, 0x8e, 0xc6, 0xa7, 0xac, 0xee, 0xf3, 0x4a, 0x94, 0x97, 0x15, 0xdb, 0x94, 0xb4, 0xd7, 0x8e, 0xe7, 0xf6, 0xf6, 0x6c, 0xf5, 0x50, 0xf7, 0xfb, 0xa9, 0xb6, 0x46}, ValidationOpt: Enforce}, 11: {Expected: []byte{0xae, 0xbf, 0x34, 0xc4, 0x10, 0xb4, 0xd9, 0xa4, 0x12, 0x94, 0x71, 0xa5, 0x7d, 0x3c, 0xc0, 0x7e, 0xb0, 0x01, 0x48, 0x9f, 0x7c, 0x7f, 0x44, 0xd6, 0xc2, 0x1a, 0x00, 0xd9, 0x2c, 0x1f, 0xbb, 0x9b}, ValidationOpt: Enforce}, 12: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 13: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 14: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: WarnOnly}, 15: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}}
	azure_AzureTDX           = M{1: {Expected: []byte{0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea, 0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d, 0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a, 0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69}, ValidationOpt: WarnOnly}, 2: {Expected: []byte{0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea, 0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d, 0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a, 0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69}, ValidationOpt: WarnOnly}, 3: {Expected: []byte{0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea, 0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d, 0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a, 0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69}, ValidationOpt: WarnOnly}, 4: {Expected: []byte{0x8a, 0x0d, 0x52, 0x7a, 0xf7, 0x1f, 0x31, 0x2a, 0x46, 0xab, 0x37, 0x06, 0xea, 0xa5, 0xd7, 0x41, 0xa8, 0x95, 0xdb, 0x4a, 0xbd, 0xba, 0x44, 0x7d, 0x1d, 0xa9, 0x44, 0xcd, 0x55, 0xdc, 0xb7, 0x6d}, ValidationOpt: Enforce}, 8: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 9: {Expected: []byte{0x70, 0xbf, 0x3e, 0x57, 0x62, 0x2f, 0x36, 0xa1, 0xc6, 0xbc, 0xb7, 0x67, 0xdc, 0x2f, 0x72, 0x2d, 0x84, 0x4c, 0x63, 0x9d, 0xd1, 0xae, 0x30, 0xd9, 0xc1, 0x7f, 0xfa, 0xb8, 0xa2, 0xeb, 0x7b, 0x2b}, ValidationOpt: Enforce}, 11: {Expected: []byte{0x9f, 0x77, 0x46, 0xa8, 0x09, 0x12, 0x4f, 0x90, 0xe7, 0xfa, 0x4b, 0x05, 0xce, 0x93, 0x17, 0x7f, 0xf7, 0x3a, 0x06, 0x86, 0x39, 0xfe, 0x45, 0xd1, 0x7d, 0xa0, 0x91, 0xbb, 0xd9, 0x7b, 0xa4, 0xc5}, ValidationOpt: Enforce}, 12: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 13: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 14: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: WarnOnly}, 15: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}}
	azure_AzureTrustedLaunch M
	gcp_GCPSEVES             = M{1: {Expected: []byte{0x36, 0x95, 0xdc, 0xc5, 0x5e, 0x3a, 0xa3, 0x40, 0x27, 0xc2, 0x77, 0x93, 0xc8, 0x5c, 0x72, 0x3c, 0x69, 0x7d, 0x70, 0x8c, 0x42, 0xd1, 0xf7, 0x3b, 0xd6, 0xfa, 0x4f, 0x26, 0x60, 0x8a, 0x5b, 0x24}, ValidationOpt: WarnOnly}, 2: {Expected: []byte{0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea, 0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d, 0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a, 0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69}, ValidationOpt: WarnOnly}, 3: {Expected: []byte{0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea, 0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d, 0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a, 0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69}, ValidationOpt: WarnOnly}, 4: {Expected: []byte{0xc3, 0x9e, 0x72, 0x77, 0x9c, 0xa5, 0x90, 0x2b, 0x69, 0x87, 0xf6, 0xbc, 0x4b, 0x00, 0xd0, 0x0e, 0x30, 0x0a, 0x80, 0x18, 0xbd, 0x41, 0x0f, 0xe5, 0xd3, 0x34, 0x11, 0xfd, 0x79, 0x95, 0x05, 0x59}, ValidationOpt: Enforce}, 6: {Expected: []byte{0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea, 0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d, 0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a, 0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69}, ValidationOpt: WarnOnly}, 8: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 9: {Expected: []byte{0xdf, 0x13, 0x61, 0x47, 0x14, 0xb6, 0x8e, 0x4f, 0xf2, 0x88, 0x91, 0x97, 0x98, 0xd9, 0xb1, 0xb0, 0xf3, 0x34, 0x67, 0xda, 0x14, 0x9a, 0x08, 0x06, 0x74, 0x52, 0x6c, 0xa9, 0x54, 0xd7, 0xdc, 0xd9}, ValidationOpt: Enforce}, 11: {Expected: []byte{0x5b, 0xc5, 0xcd, 0xea, 0x8d, 0xd3, 0x23, 0x7e, 0x67, 0x47, 0x23, 0xb2, 0x41, 0xfb, 0x5b, 0x92, 0x65, 0x31, 0xc1, 0x1f, 0x73, 0x4e, 0x5e, 0xae, 0x83, 0x99, 0xbb, 0xc1, 0x5b, 0x3f, 0x49, 0xca}, ValidationOpt: Enforce}, 12: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 13: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 14: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: WarnOnly}, 15: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}}
	gcp_GCPSEVSNP            = M{1: {Expected: []byte{0x36, 0x95, 0xdc, 0xc5, 0x5e, 0x3a, 0xa3, 0x40, 0x27, 0xc2, 0x77, 0x93, 0xc8, 0x5c, 0x72, 0x3c, 0x69, 0x7d, 0x70, 0x8c, 0x42, 0xd1, 0xf7, 0x3b, 0xd6, 0xfa, 0x4f, 0x26, 0x60, 0x8a, 0x5b, 0x24}, ValidationOpt: WarnOnly}, 2: {Expected: []byte{0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea, 0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d, 0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a, 0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69}, ValidationOpt: WarnOnly}, 3: {Expected: []byte{0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea, 0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d, 0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a, 0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69}, ValidationOpt: WarnOnly}, 4: {Expected: []byte{0x93, 0x79, 0x42, 0xd1, 0x21, 0xfc, 0x66, 0x5d, 0xb8, 0xa9, 0x9b, 0xfb, 0x56, 0x32, 0x74, 0x84, 0xf0, 0x0b, 0xff, 0x00, 0x53, 0xa5, 0x5c, 0x2f, 0xde, 0xa8, 0xa1, 0x95, 0x40, 0xf6, 0x9c, 0xba}, ValidationOpt: Enforce}, 6: {Expected: []byte{0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea, 0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d, 0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a, 0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69}, ValidationOpt: WarnOnly}, 8: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 9: {Expected: []byte{0x74, 0xbf, 0xc4, 0x68, 0x27, 0x39, 0x52, 0xbe, 0xf7, 0x42, 0x30, 0x78, 0x33, 0x6b, 0x91, 0x51, 0x67, 0x0a, 0x0a, 0x70, 0x69, 0xac, 0xa9, 0xf1, 0xad, 0x3c, 0xc3, 0xbb, 0x69, 0xd2, 0x97, 0x51}, ValidationOpt: Enforce}, 11: {Expected: []byte{0x95, 0x89, 0x21, 0x49, 0xf3, 0xf3, 0xf5, 0x9e, 0xb7, 0xf2, 0x80, 0x35, 0x1d, 0x86, 0xfb, 0x36, 0x63, 0x23, 0xae, 0xbf, 0x88, 0xd3, 0xf1, 0x48, 0xb5, 0x5f, 0xbe, 0x37, 0xc6, 0x5b, 0x55, 0x02}, ValidationOpt: Enforce}, 12: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 13: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 14: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: WarnOnly}, 15: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}}
	openstack_QEMUVTPM       = M{4: {Expected: []byte{0x36, 0x41, 0x15, 0x9b, 0x85, 0xf7, 0x15, 0xd3, 0xe9, 0x00, 0x1b, 0x3f, 0x1d, 0x4d, 0x83, 0xcd, 0x11, 0xff, 0x87, 0xb6, 0xc3, 0x9d, 0xa7, 0x26, 0x91, 0x2c, 0xd4, 0xed, 0x39, 0x97, 0x73, 0xe9}, ValidationOpt: Enforce}, 8: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 9: {Expected: []byte{0xc7, 0xac, 0x2b, 0x72, 0xb5, 0x68, 0x95, 0xc0, 0xa8, 0xcc, 0xae, 0xab, 0xce, 0x57, 0x10, 0xef, 0x38, 0x2e, 0x63, 0x0f, 0xdc, 0x55, 0x4c, 0x5c, 0xaa, 0x13, 0x9e, 0x4a, 0x1d, 0x87, 0x65, 0xd1}, ValidationOpt: Enforce}, 11: {Expected: []byte{0x44, 0xda, 0xdc, 0xe6, 0x09, 0xf0, 0xd8, 0xb1, 0xf1, 0x16, 0xb5, 0x4a, 0x68, 0xf3, 0x7f, 0x40, 0x5e, 0xa4, 0x50, 0x00, 0x88, 0x27, 0x58, 0x1c, 0x54, 0x15, 0x0b, 0x39, 0x92, 0xf5, 0x03, 0x5e}, ValidationOpt: Enforce}, 12: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 13: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 14: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: WarnOnly}, 15: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}}
	qemu_QEMUTDX             M
	qemu_QEMUVTPM            = M{4: {Expected: []byte{0x53, 0x56, 0x19, 0x85, 0x3a, 0x2a, 0x7e, 0xb2, 0x95, 0xf1, 0xa5, 0x11, 0x9e, 0x0e, 0x62, 0xbd, 0xc3, 0xca, 0x16, 0xa4, 0xd0, 0x66, 0x7e, 0x3a, 0xc9, 0x2a, 0x7f, 0xb6, 0xd5, 0xd6, 0x13, 0x81}, ValidationOpt: Enforce}, 8: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 9: {Expected: []byte{0x41, 0x4e, 0x74, 0xc2, 0x24, 0x41, 0x80, 0x22, 0xa1, 0xf1, 0xa9, 0xd3, 0x9f, 0x20, 0xc9, 0x36, 0x95, 0x26, 0xef, 0x94, 0x36, 0x2a, 0xff, 0xa9, 0x1d, 0xde, 0xdb, 0xe9, 0x97, 0xc3, 0x48, 0xd5}, ValidationOpt: Enforce}, 11: {Expected: []byte{0x23, 0xba, 0xe8, 0xf2, 0xe1, 0x17, 0x34, 0xa5, 0x46, 0x99, 0xad, 0x35, 0xb1, 0x46, 0xa9, 0x45, 0x72, 0x6c, 0x92, 0x5b, 0x39, 0x74, 0xaf, 0x80, 0x33, 0xc5, 0x40, 0xb8, 0x0c, 0x33, 0xbb, 0xc3}, ValidationOpt: Enforce}, 12: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 13: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 15: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}}
)
