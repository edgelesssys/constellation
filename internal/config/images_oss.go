//go:build !enterprise

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package config

const (
	// DefaultImageAzure is not set for OSS build.
	DefaultImageAzure = ""
	// DefaultImageGCP is not set for OSS build.
	DefaultImageGCP = ""
)
