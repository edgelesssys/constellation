/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package gcp

import (
	"context"
	"log"

	"cloud.google.com/go/logging"
	"github.com/edgelesssys/constellation/v2/internal/gcpshared"
)

// Logger logs to GCP cloud logging. Do not use to log sensitive information.
type Logger struct {
	client *logging.Client
	logger *log.Logger
}

// NewLogger creates a new Cloud Logger for GCP.
// https://cloud.google.com/logging/docs/setup/go
func NewLogger(ctx context.Context, providerID string, logName string) (*Logger, error) {
	projectID, _, _, err := gcpshared.SplitProviderID(providerID)
	if err != nil {
		return nil, err
	}
	client, err := logging.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}

	logger := client.Logger(logName).StandardLogger(logging.Info)

	return &Logger{
		client: client,
		logger: logger,
	}, nil
}

// Disclose stores log information in GCP Cloud Logging! Do **NOT** log sensitive
// information!
func (l *Logger) Disclose(msg string) {
	l.logger.Println(msg)
}

// Close waits for all buffer to be written.
func (l *Logger) Close() error {
	return l.client.Close()
}
