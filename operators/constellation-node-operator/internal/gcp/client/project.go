/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package client

import (
	"context"
	"errors"
	"regexp"

	"google.golang.org/genproto/googleapis/cloud/compute/v1"
)

var numericProjectIDRegex = regexp.MustCompile(`^\d+$`)

// canonicalProjectID returns the project id for a given project id or project number.
func (c *Client) canonicalProjectID(ctx context.Context, project string) (string, error) {
	if !numericProjectIDRegex.MatchString(project) {
		return project, nil
	}
	computeProject, err := c.projectAPI.Get(ctx, &compute.GetProjectRequest{Project: project})
	if err != nil {
		return "", err
	}
	if computeProject == nil || computeProject.Name == nil {
		return "", errors.New("invalid project")
	}
	return *computeProject.Name, nil
}
