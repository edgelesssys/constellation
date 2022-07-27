package client

import (
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/applicationinsights/armapplicationinsights"
	"golang.org/x/net/context"
)

func (c *Client) CreateApplicationInsight(ctx context.Context) error {
	properties := armapplicationinsights.Component{
		Kind:     to.Ptr("web"),
		Location: to.Ptr(c.location),
		Properties: &armapplicationinsights.ComponentProperties{
			ApplicationType: to.Ptr(armapplicationinsights.ApplicationTypeWeb),
		},
	}

	_, err := c.applicationInsightsAPI.CreateOrUpdate(
		ctx,
		c.resourceGroup,
		"constellation-insights-"+c.uid,
		properties,
		nil,
	)
	return err
}
