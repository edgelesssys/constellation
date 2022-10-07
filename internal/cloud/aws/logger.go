/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package aws

// TODO: Implement for AWS.

// Logger is a Cloud Logger for AWS.
type Logger struct{}

// NewLogger creates a new Cloud Logger for AWS.
func NewLogger() *Logger {
	return &Logger{}
}

// Disclose is not implemented for AWS.
func (l *Logger) Disclose(msg string) {}

// Close is a no-op.
func (l *Logger) Close() error {
	return nil
}
