package azure

import (
	"context"
	"errors"
	"fmt"

	"github.com/microsoft/ApplicationInsights-Go/appinsights"
)

type Logger struct {
	client appinsights.TelemetryClient
}

// NewLogger creates a new client to store information in Azure Application Insights
// https://github.com/Microsoft/ApplicationInsights-go
func NewLogger(ctx context.Context, metadata *Metadata) (*Logger, error) {
	component, err := metadata.getAppInsights(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting app insights: %w", err)
	}

	if component.Properties == nil || component.Properties.InstrumentationKey == nil {
		return nil, errors.New("unable to get instrumentation key")
	}

	client := appinsights.NewTelemetryClient(*component.Properties.InstrumentationKey)

	self, err := metadata.Self(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting self: %w", err)
	}
	client.Context().CommonProperties["instance-name"] = self.Name

	return &Logger{client: client}, nil
}

// Disclose stores log information in Azure Application Insights!
// Do **NOT** log sensitive information!
func (l *Logger) Disclose(msg string) {
	l.client.Track(appinsights.NewTraceTelemetry(msg, appinsights.Information))
}

// Close blocks until all information are written to cloud API.
func (l *Logger) Close() error {
	<-l.client.Channel().Close()
	return nil
}
