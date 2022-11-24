/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package helm

import (
	"fmt"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
)

// Client handles interaction with helm.
type Client struct {
	config *action.Configuration
}

// NewClient returns a new initializes client for the namespace Client.
func NewClient(kubeConfigPath, helmNamespace string) (*Client, error) {
	settings := cli.New()
	settings.KubeConfig = kubeConfigPath // constants.AdminConfFilename

	actionConfig := &action.Configuration{}
	if err := actionConfig.Init(settings.RESTClientGetter(), helmNamespace, "secret", nil); err != nil {
		return nil, fmt.Errorf("initializing config: %w", err)
	}

	return &Client{config: actionConfig}, nil
}

// CurrentVersion returns the version of the currently installed helm release.
func (c *Client) CurrentVersion(release string) (string, error) {
	client := action.NewList(c.config)
	client.Filter = release
	rel, err := client.Run()
	if err != nil {
		return "", err
	}

	if len(rel) == 0 {
		return "", fmt.Errorf("release %s not found", release)
	}
	if len(rel) > 1 {
		return "", fmt.Errorf("multiple releases found for %s", release)
	}

	if rel[0] == nil || rel[0].Chart == nil || rel[0].Chart.Metadata == nil {
		return "", fmt.Errorf("received invalid release %s", release)
	}

	return rel[0].Chart.Metadata.Version, nil
}
