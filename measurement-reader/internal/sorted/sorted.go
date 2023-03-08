/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Type definition for sorted measurements.
package sorted

// Measurement wraps a measurement custom index and value.
type Measurement struct {
	Index string
	Value []byte
}
