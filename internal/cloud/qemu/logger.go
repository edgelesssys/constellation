/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package qemu

import (
	"context"
	"net/http"
	"net/url"
	"strings"
)

// Logger is a Cloud Logger for QEMU.
type Logger struct{}

// NewLogger creates a new Cloud Logger for QEMU.
func NewLogger() *Logger {
	return &Logger{}
}

// Disclose writes log information to QEMU's cloud log.
// This is done by sending a POST request to the QEMU's metadata endpoint.
func (l *Logger) Disclose(msg string) {
	url := &url.URL{
		Scheme: "http",
		Host:   qemuMetadataEndpoint,
		Path:   "/log",
	}

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, url.String(), strings.NewReader(msg))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err == nil {
		defer resp.Body.Close()
	}
}

// Close is a no-op.
func (l *Logger) Close() error {
	return nil
}
