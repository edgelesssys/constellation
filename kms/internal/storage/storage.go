package storage

import (
	"errors"
)

// ErrDEKUnset indicates if a key is not found in storage.
var ErrDEKUnset = errors.New("requested DEK not set")
