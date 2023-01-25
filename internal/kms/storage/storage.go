/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
Package storage implements storage backends for DEKs.

If an unset DEK is requested, the backend MUST return [ErrDEKUnset].
*/
package storage

import (
	"context"
	"errors"
)

type Storage interface {
	Put(ctx context.Context, keyID string, dek []byte) error
	Get(ctx context.Context, keyID string) ([]byte, error)
}

// ErrDEKUnset indicates if a key is not found in storage.
var ErrDEKUnset = errors.New("requested DEK not set")
