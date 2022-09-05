/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package client

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

// TerminateResourceGroupResources deletes all resources from the resource group.
func (c *Client) TerminateResourceGroupResources(ctx context.Context) error {
	const timeOut = 10 * time.Minute
	ctx, cancel := context.WithTimeout(ctx, timeOut)
	defer cancel()

	pollers := make(chan *runtime.Poller[armresources.ClientDeleteByIDResponse], 20)
	delete := make(chan struct{}, 1)
	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() { // This routine lists resources and starts their deletion, where possible.
		defer wg.Done()
		defer func() {
			close(pollers)
			for range delete { // drain channel
			}
		}()

		for {
			ids, err := c.getResourceIDList(ctx)
			if err != nil {
				time.Sleep(3 * time.Second)
				continue
			}

			if len(ids) == 0 {
				return
			}

			for _, id := range ids {
				poller, err := c.deleteResourceByID(ctx, id)
				if err != nil {
					continue
				}
				pollers <- poller
			}

			select {
			case <-ctx.Done():
				return
			case _, ok := <-delete:
				if !ok { // channel was closed
					return
				}
			}
		}
	}()

	go func() { // This routine polls for for the deletions to complete.
		defer wg.Done()
		defer close(delete)

		for poller := range pollers {
			_, err := poller.PollUntilDone(ctx, nil)
			if err != nil {
				continue
			}
			select {
			case delete <- struct{}{}:
			default:
			}
		}
	}()

	wg.Wait()

	return nil
}

func (c *Client) getResourceIDList(ctx context.Context) ([]string, error) {
	var ids []string
	pager := c.resourceAPI.NewListByResourceGroupPager(c.resourceGroup, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("getting next page of ListByResourceGroup: %w", err)
		}
		for _, resource := range page.Value {
			if resource.ID == nil {
				return nil, fmt.Errorf("resource %v has no ID", resource)
			}
			ids = append(ids, *resource.ID)
		}
	}
	return ids, nil
}

func (c *Client) deleteResourceByID(ctx context.Context, id string,
) (*runtime.Poller[armresources.ClientDeleteByIDResponse], error) {
	apiVersion := "2020-02-02"

	// First try, API version unknown, will fail.
	poller, err := c.resourceAPI.BeginDeleteByID(ctx, id, apiVersion, nil)
	if isVersionWrongErr(err) {
		// bad hack, but easiest way to get the right API version
		apiVersion = parseAPIVersionFromErr(err)
		poller, err = c.resourceAPI.BeginDeleteByID(ctx, id, apiVersion, nil)
	}
	return poller, err
}

func isVersionWrongErr(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "NoRegisteredProviderFound") &&
		strings.Contains(err.Error(), "The supported api-versions are")
}

var apiVersionRegex = regexp.MustCompile(` (\d\d\d\d-\d\d-\d\d)'`)

func parseAPIVersionFromErr(err error) string {
	if err == nil {
		return ""
	}
	matches := apiVersionRegex.FindStringSubmatch(err.Error())
	return matches[1]
}
