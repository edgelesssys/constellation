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
	aws_AWSNitroTPM          = M{0: {Expected: []byte{0x73, 0x7f, 0x76, 0x7a, 0x12, 0xf5, 0x4e, 0x70, 0xee, 0xcb, 0xc8, 0x68, 0x40, 0x11, 0x32, 0x3a, 0xe2, 0xfe, 0x2d, 0xd9, 0xf9, 0x07, 0x85, 0x57, 0x79, 0x69, 0xd7, 0xa2, 0x01, 0x3e, 0x8c, 0x12}, ValidationOpt: WarnOnly}, 2: {Expected: []byte{0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea, 0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d, 0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a, 0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69}, ValidationOpt: WarnOnly}, 3: {Expected: []byte{0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea, 0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d, 0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a, 0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69}, ValidationOpt: WarnOnly}, 4: {Expected: []byte{0xad, 0x5e, 0x0b, 0x65, 0x8d, 0x1d, 0x89, 0xa6, 0x4c, 0xba, 0x29, 0x39, 0x19, 0x59, 0x19, 0x1c, 0x59, 0x3f, 0xa6, 0xf5, 0xf7, 0x2d, 0x7f, 0x9d, 0xc0, 0xeb, 0x50, 0xcf, 0xd4, 0xc6, 0x50, 0x49}, ValidationOpt: Enforce}, 6: {Expected: []byte{0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea, 0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d, 0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a, 0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69}, ValidationOpt: WarnOnly}, 8: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 9: {Expected: []byte{0x4f, 0xcd, 0x6c, 0x24, 0xf7, 0x54, 0xff, 0x1a, 0xdf, 0xc6, 0xc6, 0xff, 0x98, 0x53, 0xcf, 0x45, 0x44, 0xb6, 0x34, 0x04, 0xeb, 0x1f, 0x67, 0xcd, 0x2a, 0x40, 0x51, 0xff, 0xc7, 0x4d, 0x51, 0xb5}, ValidationOpt: Enforce}, 11: {Expected: []byte{0xac, 0x4f, 0xdf, 0x33, 0xbd, 0x75, 0x90, 0xd0, 0x9f, 0x59, 0x28, 0xb8, 0x56, 0x43, 0xab, 0x81, 0xd4, 0xd6, 0x40, 0x93, 0xaf, 0xab, 0x68, 0xe7, 0x39, 0x69, 0x43, 0xee, 0x95, 0x32, 0x01, 0x1d}, ValidationOpt: Enforce}, 12: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 13: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 14: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: WarnOnly}, 15: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}}
	aws_AWSSEVSNP            = M{0: {Expected: []byte{0xd6, 0xdf, 0x85, 0x53, 0x58, 0xf5, 0xb1, 0x0f, 0x06, 0xf0, 0xfa, 0xb3, 0xf4, 0x08, 0xad, 0x26, 0xcd, 0x16, 0x5a, 0x29, 0x49, 0xba, 0xd6, 0x9e, 0x2c, 0xc7, 0x56, 0x92, 0x52, 0x9e, 0x66, 0x2a}, ValidationOpt: WarnOnly}, 2: {Expected: []byte{0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea, 0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d, 0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a, 0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69}, ValidationOpt: WarnOnly}, 3: {Expected: []byte{0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea, 0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d, 0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a, 0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69}, ValidationOpt: WarnOnly}, 4: {Expected: []byte{0xb4, 0x75, 0x00, 0xf6, 0x30, 0xb4, 0x54, 0x70, 0x0c, 0x87, 0xca, 0xec, 0x86, 0xc1, 0x25, 0xeb, 0x79, 0x34, 0xaa, 0x17, 0x4f, 0x9c, 0x35, 0x4d, 0x46, 0xca, 0xe3, 0x47, 0xe1, 0x40, 0xec, 0xde}, ValidationOpt: Enforce}, 6: {Expected: []byte{0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea, 0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d, 0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a, 0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69}, ValidationOpt: WarnOnly}, 8: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 9: {Expected: []byte{0x6b, 0xe2, 0xed, 0xee, 0xc9, 0xd5, 0x7a, 0x49, 0x8b, 0x6d, 0xa6, 0x9b, 0x4b, 0x96, 0xcb, 0x75, 0xc0, 0x0c, 0x38, 0x3a, 0x55, 0x36, 0x89, 0xd0, 0x39, 0xc2, 0xa1, 0xde, 0xb2, 0xb4, 0x0f, 0xe6}, ValidationOpt: Enforce}, 11: {Expected: []byte{0xf6, 0x4e, 0x2c, 0xc6, 0xdc, 0x91, 0xb0, 0x1a, 0x0c, 0x27, 0xaf, 0x0e, 0x54, 0x66, 0x27, 0x48, 0x00, 0xc8, 0xe9, 0x38, 0xac, 0x3b, 0x65, 0x41, 0x1b, 0x48, 0x92, 0x65, 0xdc, 0x55, 0x56, 0xe2}, ValidationOpt: Enforce}, 12: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 13: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 14: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: WarnOnly}, 15: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}}
	azure_AzureSEVSNP        = M{1: {Expected: []byte{0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea, 0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d, 0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a, 0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69}, ValidationOpt: WarnOnly}, 2: {Expected: []byte{0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea, 0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d, 0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a, 0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69}, ValidationOpt: WarnOnly}, 3: {Expected: []byte{0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea, 0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d, 0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a, 0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69}, ValidationOpt: WarnOnly}, 4: {Expected: []byte{0x98, 0x32, 0x19, 0x32, 0x89, 0x53, 0xfa, 0x2b, 0x75, 0x67, 0xb9, 0x3a, 0x79, 0x3d, 0x44, 0xe4, 0xdf, 0xed, 0xf6, 0xbe, 0x9c, 0xbe, 0x50, 0x08, 0x3e, 0x52, 0xa7, 0xc6, 0xcb, 0xb4, 0xd7, 0x96}, ValidationOpt: Enforce}, 8: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 9: {Expected: []byte{0xd7, 0x25, 0x4c, 0xc4, 0xe6, 0x9e, 0x04, 0x0b, 0x67, 0xf2, 0x65, 0x0c, 0xb9, 0xb8, 0xc5, 0x17, 0xcf, 0x0e, 0xcd, 0x0b, 0x55, 0x21, 0xb8, 0xb3, 0x01, 0x9a, 0x51, 0x2c, 0x19, 0xe6, 0xf4, 0x86}, ValidationOpt: Enforce}, 11: {Expected: []byte{0xda, 0x7c, 0x69, 0x6e, 0xcf, 0x06, 0x08, 0xfd, 0x6d, 0x71, 0xef, 0x86, 0x19, 0x9b, 0x69, 0x5b, 0x38, 0x21, 0x6c, 0xb8, 0xe3, 0x3a, 0x83, 0x57, 0x78, 0x0c, 0xef, 0x2b, 0x39, 0xb9, 0x9d, 0x25}, ValidationOpt: Enforce}, 12: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 13: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 14: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: WarnOnly}, 15: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}}
	azure_AzureTDX           = M{1: {Expected: []byte{0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea, 0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d, 0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a, 0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69}, ValidationOpt: WarnOnly}, 2: {Expected: []byte{0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea, 0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d, 0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a, 0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69}, ValidationOpt: WarnOnly}, 3: {Expected: []byte{0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea, 0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d, 0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a, 0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69}, ValidationOpt: WarnOnly}, 4: {Expected: []byte{0x75, 0xe6, 0x04, 0x4a, 0x09, 0x6b, 0x95, 0x34, 0xa5, 0xde, 0x69, 0x2d, 0xfa, 0xbc, 0x07, 0xaf, 0xf8, 0xb4, 0x46, 0xf1, 0x93, 0x9d, 0xb5, 0xa9, 0xb3, 0x21, 0x20, 0x29, 0x1d, 0xb5, 0x3f, 0xdf}, ValidationOpt: Enforce}, 8: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 9: {Expected: []byte{0x04, 0x20, 0x5c, 0x22, 0x11, 0x80, 0x26, 0xee, 0x3d, 0xa0, 0x31, 0xb7, 0xcc, 0x5b, 0x89, 0x68, 0xef, 0x24, 0x9b, 0x43, 0x88, 0x20, 0x0b, 0xa2, 0xe4, 0xdf, 0xaa, 0x0c, 0xe4, 0x90, 0x95, 0x3c}, ValidationOpt: Enforce}, 11: {Expected: []byte{0xd8, 0x8e, 0x86, 0xd3, 0x34, 0xce, 0x8b, 0xc7, 0x96, 0x30, 0x47, 0x4e, 0xfc, 0x10, 0x7b, 0x31, 0x48, 0x67, 0xab, 0xf3, 0x41, 0xa2, 0xb4, 0x7f, 0xf5, 0x23, 0x0b, 0x45, 0x33, 0x04, 0x15, 0x5e}, ValidationOpt: Enforce}, 12: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 13: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 14: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: WarnOnly}, 15: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}}
	azure_AzureTrustedLaunch M
	gcp_GCPSEVES             = M{1: {Expected: []byte{0x36, 0x95, 0xdc, 0xc5, 0x5e, 0x3a, 0xa3, 0x40, 0x27, 0xc2, 0x77, 0x93, 0xc8, 0x5c, 0x72, 0x3c, 0x69, 0x7d, 0x70, 0x8c, 0x42, 0xd1, 0xf7, 0x3b, 0xd6, 0xfa, 0x4f, 0x26, 0x60, 0x8a, 0x5b, 0x24}, ValidationOpt: WarnOnly}, 2: {Expected: []byte{0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea, 0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d, 0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a, 0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69}, ValidationOpt: WarnOnly}, 3: {Expected: []byte{0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea, 0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d, 0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a, 0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69}, ValidationOpt: WarnOnly}, 4: {Expected: []byte{0x77, 0x0a, 0x48, 0x26, 0x6b, 0xca, 0x94, 0xfe, 0xd6, 0x82, 0x39, 0x33, 0xed, 0x2c, 0xf2, 0x7c, 0xe2, 0xc7, 0xad, 0x18, 0xc7, 0xb9, 0xba, 0x2d, 0x80, 0x32, 0x58, 0x2f, 0x4d, 0xae, 0x2f, 0x18}, ValidationOpt: Enforce}, 6: {Expected: []byte{0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea, 0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d, 0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a, 0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69}, ValidationOpt: WarnOnly}, 8: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 9: {Expected: []byte{0xf6, 0x4f, 0x15, 0xfb, 0x8f, 0xeb, 0x57, 0xbf, 0xfc, 0xb1, 0x8a, 0xa8, 0xf0, 0xfc, 0x2b, 0x0d, 0x63, 0x12, 0x24, 0x38, 0x73, 0xe2, 0x00, 0x50, 0xfc, 0x8a, 0x46, 0x60, 0x90, 0x96, 0xd1, 0x3b}, ValidationOpt: Enforce}, 11: {Expected: []byte{0x7c, 0xb7, 0x64, 0x20, 0x8a, 0xd3, 0x55, 0xb1, 0x76, 0x25, 0x55, 0xd6, 0xe6, 0xd4, 0x88, 0x9e, 0x29, 0xc8, 0x27, 0x34, 0xb1, 0xe8, 0xd4, 0x33, 0x0b, 0x40, 0xdd, 0xcb, 0x44, 0x82, 0x1b, 0xdb}, ValidationOpt: Enforce}, 12: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 13: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 14: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: WarnOnly}, 15: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}}
	gcp_GCPSEVSNP            = M{1: {Expected: []byte{0x36, 0x95, 0xdc, 0xc5, 0x5e, 0x3a, 0xa3, 0x40, 0x27, 0xc2, 0x77, 0x93, 0xc8, 0x5c, 0x72, 0x3c, 0x69, 0x7d, 0x70, 0x8c, 0x42, 0xd1, 0xf7, 0x3b, 0xd6, 0xfa, 0x4f, 0x26, 0x60, 0x8a, 0x5b, 0x24}, ValidationOpt: WarnOnly}, 2: {Expected: []byte{0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea, 0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d, 0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a, 0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69}, ValidationOpt: WarnOnly}, 3: {Expected: []byte{0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea, 0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d, 0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a, 0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69}, ValidationOpt: WarnOnly}, 4: {Expected: []byte{0xc4, 0x0f, 0x5b, 0x30, 0x51, 0x2d, 0x43, 0x3d, 0x16, 0x2b, 0xac, 0x11, 0x71, 0x61, 0xe0, 0x4e, 0xd3, 0xae, 0xb5, 0x65, 0xe1, 0x4c, 0xbb, 0x11, 0x56, 0xd4, 0x30, 0x1b, 0x14, 0xac, 0x61, 0x5d}, ValidationOpt: Enforce}, 6: {Expected: []byte{0x3d, 0x45, 0x8c, 0xfe, 0x55, 0xcc, 0x03, 0xea, 0x1f, 0x44, 0x3f, 0x15, 0x62, 0xbe, 0xec, 0x8d, 0xf5, 0x1c, 0x75, 0xe1, 0x4a, 0x9f, 0xcf, 0x9a, 0x72, 0x34, 0xa1, 0x3f, 0x19, 0x8e, 0x79, 0x69}, ValidationOpt: WarnOnly}, 8: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 9: {Expected: []byte{0xda, 0x0e, 0x94, 0xaf, 0x2e, 0xbc, 0xb6, 0x0e, 0x9e, 0x5f, 0x40, 0x4f, 0x28, 0x90, 0x78, 0xba, 0xdf, 0x2a, 0x31, 0x5c, 0xb5, 0x6b, 0x40, 0x2a, 0x6d, 0xc4, 0xbb, 0xcf, 0x21, 0x61, 0xbb, 0x55}, ValidationOpt: Enforce}, 11: {Expected: []byte{0xea, 0x26, 0x71, 0x11, 0x2a, 0xb4, 0x00, 0xe5, 0x07, 0x94, 0x50, 0x50, 0x84, 0x63, 0xf1, 0x47, 0x61, 0x5e, 0xed, 0x97, 0xc0, 0xf8, 0x03, 0xf8, 0x01, 0xf7, 0x2f, 0xed, 0x56, 0x59, 0xe7, 0x1f}, ValidationOpt: Enforce}, 12: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 13: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 14: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: WarnOnly}, 15: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}}
	openstack_QEMUVTPM       = M{4: {Expected: []byte{0x0d, 0xb0, 0x94, 0x5c, 0x51, 0x98, 0xab, 0x89, 0xb8, 0x25, 0xd5, 0xcb, 0x4e, 0xd8, 0x34, 0x53, 0x20, 0x2b, 0x7a, 0xe9, 0x2e, 0x71, 0x90, 0xd7, 0x6d, 0x90, 0xd5, 0xc0, 0x8a, 0x1d, 0xbd, 0xa0}, ValidationOpt: Enforce}, 8: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 9: {Expected: []byte{0x71, 0x3e, 0x9d, 0xe0, 0x2a, 0xa3, 0x6b, 0xe0, 0xf9, 0x69, 0x49, 0x82, 0x1c, 0xb4, 0xc7, 0x15, 0xec, 0xd7, 0x8d, 0x51, 0xeb, 0xc7, 0x11, 0x80, 0x29, 0xbc, 0xd7, 0xbc, 0xed, 0xf6, 0x4d, 0xa8}, ValidationOpt: Enforce}, 11: {Expected: []byte{0xb4, 0x4e, 0x33, 0xd0, 0xc2, 0x57, 0x9f, 0xa0, 0x9a, 0x73, 0x40, 0x8b, 0x2a, 0x2f, 0xfd, 0x31, 0x30, 0x92, 0x76, 0x6e, 0x1d, 0xb0, 0x7a, 0xfd, 0x40, 0xc9, 0x50, 0xb7, 0xe4, 0xff, 0x25, 0x4c}, ValidationOpt: Enforce}, 12: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 13: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 14: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: WarnOnly}, 15: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}}
	qemu_QEMUTDX             M
	qemu_QEMUVTPM            = M{4: {Expected: []byte{0x20, 0x4a, 0x80, 0xbf, 0x33, 0xf5, 0xa4, 0x1e, 0x7c, 0x02, 0x30, 0x73, 0xb5, 0xc1, 0xae, 0xfc, 0x92, 0x75, 0x69, 0x0f, 0x07, 0x9b, 0xf3, 0xf5, 0x1b, 0xf7, 0xd2, 0xa4, 0x89, 0xce, 0x68, 0x87}, ValidationOpt: Enforce}, 8: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 9: {Expected: []byte{0x9e, 0xeb, 0x56, 0x6a, 0xb6, 0xf4, 0x84, 0x5d, 0xae, 0xb2, 0xc8, 0xed, 0x78, 0x2c, 0xb3, 0x00, 0xac, 0x22, 0xe2, 0x05, 0x10, 0xc9, 0x9f, 0x08, 0xfe, 0x2e, 0xf2, 0x78, 0xb8, 0xd7, 0x85, 0xd9}, ValidationOpt: Enforce}, 11: {Expected: []byte{0x8d, 0x1b, 0x12, 0xf2, 0x55, 0x93, 0xb7, 0x22, 0xca, 0xc0, 0xed, 0xb3, 0xa6, 0xd0, 0x86, 0x69, 0xc0, 0x06, 0xd5, 0x71, 0x09, 0xed, 0x47, 0xbd, 0x74, 0xcf, 0x12, 0x5e, 0x95, 0xeb, 0x8d, 0xb3}, ValidationOpt: Enforce}, 12: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 13: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}, 15: {Expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, ValidationOpt: Enforce}}
)
