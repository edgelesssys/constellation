package azure

import (
	"github.com/microsoft/ApplicationInsights-Go/appinsights"
)

type Logger struct {
	client appinsights.TelemetryClient
}

// NewLogger creates a new client to store information in Azure Application Insights
// https://github.com/Microsoft/ApplicationInsights-go
func NewLogger(instrumentationKey string) *Logger {
	return &Logger{
		client: appinsights.NewTelemetryClient(instrumentationKey),
	}
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
