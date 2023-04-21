/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package config provides configuration constants for the KeyService.
package config

const (
	// SymmetricKeyLength is the length of symmetric encryption keys in bytes. We use AES256, therefore this is 32 Bytes.
	SymmetricKeyLength = 32
)

var (
	// KmsTags are the default tags for kms client created KMS solutions.
	KmsTags = map[string]string{
		"createdBy": "constellation-kms-client",
		"component": "constellation-kek",
	}
	// StorageTags are the default tags for kms client created storage solutions.
	StorageTags = map[string]*string{
		"createdBy": toPtr("constellation-kms-client"),
		"component": toPtr("constellation-dek-store"),
	}
	// AWSS3Tag is the default tag string for kms client created AWS S3 storage solutions.
	AWSS3Tag = "createdBy=constellation-kms-client&component=constellation-dek-store"
)

func toPtr[T any](v T) *T {
	return &v
}
