/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package gcp

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/compute/metadata"
	"cloud.google.com/go/logging"
)

// Logger logs to GCP cloud logging. Do not use to log sensitive information.
type Logger struct {
	client *logging.Client
	logger *log.Logger
}

// NewLogger creates a new Cloud Logger for GCP.
// https://cloud.google.com/logging/docs/setup/go
func NewLogger(ctx context.Context, logName string) (*Logger, error) {
	projectID, err := metadata.NewClient(nil).ProjectID()
	if err != nil {
		return nil, fmt.Errorf("retrieving project ID from imds: %w", err)
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
