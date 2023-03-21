/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package azure

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/applicationinsights/armapplicationinsights"
	"github.com/edgelesssys/constellation/v2/internal/cloud"
	"github.com/microsoft/ApplicationInsights-Go/appinsights"
)

// Logger implements CloudLogger interface for Azure to Disclose early boot
// logs into Azure's App Insights service.
type Logger struct {
	client appinsights.TelemetryClient
}

// NewLogger creates a new client to store information in Azure Application Insights
// https://github.com/Microsoft/ApplicationInsights-go
func NewLogger(ctx context.Context) (*Logger, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("loading credentials: %w", err)
	}
	imdsAPI := &IMDSClient{
		client: &http.Client{Transport: &http.Transport{Proxy: nil}},
	}
	subscriptionID, err := imdsAPI.subscriptionID(ctx)
	if err != nil {
		return nil, fmt.Errorf("retrieving subscription ID: %w", err)
	}
	appInsightAPI, err := armapplicationinsights.NewComponentsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("setting up insights API client. %w", err)
	}

	instrumentationKey, err := getAppInsightsKey(ctx, imdsAPI, appInsightAPI)
	if err != nil {
		return nil, fmt.Errorf("getting app insights instrumentation key: %w", err)
	}

	client := appinsights.NewTelemetryClient(instrumentationKey)

	name, err := imdsAPI.name(ctx)
	if err != nil {
		return nil, fmt.Errorf("retrieving instance name: %w", err)
	}
	client.Context().CommonProperties["instance-name"] = name

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

// getAppInsightsKey returns a instrumentation key needed to set up cloud logging on Azure.
// The key is retrieved from the resource group of the instance the function is called from.
func getAppInsightsKey(ctx context.Context, imdsAPI imdsAPI, appInsightAPI applicationInsightsAPI) (string, error) {
	resourceGroup, err := imdsAPI.resourceGroup(ctx)
	if err != nil {
		return "", err
	}
	uid, err := imdsAPI.uid(ctx)
	if err != nil {
		return "", err
	}

	pager := appInsightAPI.NewListByResourceGroupPager(resourceGroup, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return "", fmt.Errorf("retrieving application insights: %w", err)
		}

		for _, component := range page.Value {
			if component == nil || component.Tags == nil ||
				component.Tags[cloud.TagUID] == nil || *component.Tags[cloud.TagUID] != uid {
				continue
			}

			if component.Properties == nil || component.Properties.InstrumentationKey == nil {
				return "", errors.New("unable to get instrumentation key")
			}
			return *component.Properties.InstrumentationKey, nil
		}
	}
	return "", errors.New("could not find correctly tagged application insights")
}
