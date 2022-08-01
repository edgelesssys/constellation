package azure

import (
	"context"
	"errors"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/applicationinsights/armapplicationinsights"
	"github.com/edgelesssys/constellation/internal/azureshared"
	"github.com/microsoft/ApplicationInsights-Go/appinsights"
)

type Logger struct {
	client appinsights.TelemetryClient
}

// NewLogger creates a new client to store information in Azure Application Insights
// https://github.com/Microsoft/ApplicationInsights-go
func NewLogger(ctx context.Context, metadata *Metadata) (*Logger, error) {
	providerID, err := metadata.providerID(ctx)
	if err != nil {
		return nil, err
	}

	_, resourceGroup, err := azureshared.BasicsFromProviderID(providerID)
	if err != nil {
		return nil, err
	}

	uid, err := azureshared.UIDFromProviderID(providerID)
	if err != nil {
		return nil, err
	}

	resourceName := "constellation-insights-" + uid
	resp, err := metadata.applicationInsightsAPI.Get(ctx, resourceGroup, resourceName, &armapplicationinsights.ComponentsClientGetOptions{})
	if err != nil {
		return nil, err
	}
	if resp.Properties == nil || resp.Properties.InstrumentationKey == nil {
		return nil, errors.New("unable to get instrumentation key")
	}
	client := appinsights.NewTelemetryClient(*resp.Properties.InstrumentationKey)

	instance, err := metadata.GetInstance(ctx, providerID)
	if err != nil {
		return nil, err
	}
	client.Context().CommonProperties["instance-name"] = instance.Name

	return &Logger{
		client: client,
	}, nil
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
