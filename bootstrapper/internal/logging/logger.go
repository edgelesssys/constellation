/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package logging provides an interface for logging information to a non-confidential destination
package logging

import "io"

// CloudLogger is used to log information to a **non-confidential** destination
// at cloud provider for early-boot debugging. Make sure to **NOT** include any
// sensitive information!
type CloudLogger interface {
	// Disclose is used to log information into a **non-confidential** destination at
	// cloud provider for early-boot debugging. Make sure to **NOT** Disclose any
	// sensitive information!
	Disclose(msg string)
	io.Closer
}

// NopLogger implements CloudLogger interface, but does nothing.
type NopLogger struct{}

// Disclose does nothing.
func (l *NopLogger) Disclose(_ string) {}

// Close does nothing.
func (l *NopLogger) Close() error {
	return nil
}
