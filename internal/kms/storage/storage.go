/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package storage

import (
	"errors"
)

// ErrDEKUnset indicates if a key is not found in storage.
var ErrDEKUnset = errors.New("requested DEK not set")
