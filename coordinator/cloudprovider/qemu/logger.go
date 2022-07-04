package qemu

import (
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

	_, _ = http.Post(url.String(), "application/json", strings.NewReader(msg))
}

// Close is a no-op.
func (l *Logger) Close() error {
	return nil
}
