//go:build !enterprise

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package config

const (
	// defaultImage is not set for OSS build.
	defaultImage = "feat-e2e-qemu-v2.3.0-pre.0.20221208104406-14fb80068017" // TODO: remove once tested
)
