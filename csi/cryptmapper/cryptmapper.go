/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package cryptmapper provides a wrapper around libcryptsetup to manage dm-crypt volumes for CSI drivers.
package cryptmapper

import (
	"context"
)

// KeyCreator is an interface to create data encryption keys.
type KeyCreator interface {
	GetDEK(ctx context.Context, dekID string, dekSize int) ([]byte, error)
}
