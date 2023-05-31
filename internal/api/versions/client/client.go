/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
Package client provides a versions API specific implementation of the general API client.
*/
package client

import (
	"context"
	"errors"
	"fmt"
	"path"

	"golang.org/x/mod/semver"

	apiclient "github.com/edgelesssys/constellation/v2/internal/api/client"
	apifetcher "github.com/edgelesssys/constellation/v2/internal/api/fetcher"
	versionsapi "github.com/edgelesssys/constellation/v2/internal/api/versions"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/logger"
)

// VersionsClient is a client for the versions API.
type VersionsClient struct {
	*apiclient.Client
	clientClose func(ctx context.Context) error
}

// NewClient creates a new client for the versions API.
func NewClient(ctx context.Context, region, bucket, distributionID string, dryRun bool,
	log *logger.Logger,
) (*VersionsClient, CloseFunc, error) {
	genericClient, genericClientClose, err := apiclient.NewClient(ctx, region, bucket, distributionID, dryRun, log)
	versionsClient := &VersionsClient{
		genericClient,
		genericClientClose,
	}
	versionsClientClose := func(ctx context.Context) error {
		return versionsClient.Close(ctx)
	}
	return versionsClient, versionsClientClose, err
}

// NewReadOnlyClient creates a new read-only client.
// This client can be used to fetch objects but cannot write updates.
func NewReadOnlyClient(ctx context.Context, region, bucket, distributionID string,
	log *logger.Logger,
) (*VersionsClient, CloseFunc, error) {
	genericClient, genericClientClose, err := apiclient.NewReadOnlyClient(ctx, region, bucket, distributionID, log)
	if err != nil {
		return nil, nil, err
	}
	versionsClient := &VersionsClient{
		genericClient,
		genericClientClose,
	}
	versionsClientClose := func(ctx context.Context) error {
		return versionsClient.Close(ctx)
	}
	return versionsClient, versionsClientClose, err
}

// Close closes the client.
func (c *VersionsClient) Close(ctx context.Context) error {
	if c.clientClose == nil {
		return nil
	}
	return c.clientClose(ctx)
}

// FetchVersionList fetches the given version list from the versions API.
func (c *VersionsClient) FetchVersionList(ctx context.Context, list versionsapi.List) (versionsapi.List, error) {
	return apiclient.Fetch(ctx, c.Client, list)
}

// UpdateVersionList updates the given version list in the versions API.
func (c *VersionsClient) UpdateVersionList(ctx context.Context, list versionsapi.List) error {
	semver.Sort(list.Versions)
	return apiclient.Update(ctx, c.Client, list)
}

// FetchVersionLatest fetches the latest version from the versions API.
func (c *VersionsClient) FetchVersionLatest(ctx context.Context, latest versionsapi.Latest) (versionsapi.Latest, error) {
	return apiclient.Fetch(ctx, c.Client, latest)
}

// UpdateVersionLatest updates the latest version in the versions API.
func (c *VersionsClient) UpdateVersionLatest(ctx context.Context, latest versionsapi.Latest) error {
	return apiclient.Update(ctx, c.Client, latest)
}

// FetchImageInfo fetches the given image info from the versions API.
func (c *VersionsClient) FetchImageInfo(ctx context.Context, imageInfo versionsapi.ImageInfo) (versionsapi.ImageInfo, error) {
	return apiclient.Fetch(ctx, c.Client, imageInfo)
}

// UpdateImageInfo updates the given image info in the versions API.
func (c *VersionsClient) UpdateImageInfo(ctx context.Context, imageInfo versionsapi.ImageInfo) error {
	return apiclient.Update(ctx, c.Client, imageInfo)
}

// FetchCLIInfo fetches the given CLI info from the versions API.
func (c *VersionsClient) FetchCLIInfo(ctx context.Context, cliInfo versionsapi.CLIInfo) (versionsapi.CLIInfo, error) {
	return apiclient.Fetch(ctx, c.Client, cliInfo)
}

// UpdateCLIInfo updates the given CLI info in the versions API.
func (c *VersionsClient) UpdateCLIInfo(ctx context.Context, cliInfo versionsapi.CLIInfo) error {
	return apiclient.Update(ctx, c.Client, cliInfo)
}

// DeleteRef deletes the given ref from the versions API.
func (c *VersionsClient) DeleteRef(ctx context.Context, ref string) error {
	if err := versionsapi.ValidateRef(ref); err != nil {
		return fmt.Errorf("validating ref: %w", err)
	}

	refPath := path.Join(constants.CDNAPIPrefix, "ref", ref)
	if err := c.Client.DeletePath(ctx, refPath); err != nil {
		return fmt.Errorf("deleting ref path: %w", err)
	}

	return nil
}

// DeleteVersion deletes the given version from the versions API.
// The version will be removed from version lists and latest versions, and the versioned
// objects are deleted.
// Notice that the versions API can get into an inconsistent state if the version is the latest
// version but there is no older version of the same minor version available.
// Manual update of latest versions is required in this case.
func (c *VersionsClient) DeleteVersion(ctx context.Context, ver versionsapi.Version) error {
	var retErr error

	c.Client.Log.Debugf("Deleting version %s from minor version list", ver.Version)
	possibleNewLatest, err := c.deleteVersionFromMinorVersionList(ctx, ver)
	if err != nil {
		retErr = errors.Join(retErr, fmt.Errorf("removing from minor version list: %w", err))
	}

	c.Client.Log.Debugf("Checking latest version for %s", ver.Version)
	if err := c.deleteVersionFromLatest(ctx, ver, possibleNewLatest); err != nil {
		retErr = errors.Join(retErr, fmt.Errorf("updating latest version: %w", err))
	}

	c.Client.Log.Debugf("Deleting artifact path %s for %s", ver.ArtifactPath(versionsapi.APIV1), ver.Version)
	if err := c.Client.DeletePath(ctx, ver.ArtifactPath(versionsapi.APIV1)); err != nil {
		retErr = errors.Join(retErr, fmt.Errorf("deleting artifact path: %w", err))
	}

	return retErr
}

func (c *VersionsClient) deleteVersionFromMinorVersionList(ctx context.Context, ver versionsapi.Version,
) (*versionsapi.Latest, error) {
	minorList := versionsapi.List{
		Ref:         ver.Ref,
		Stream:      ver.Stream,
		Granularity: versionsapi.GranularityMinor,
		Base:        ver.WithGranularity(versionsapi.GranularityMinor),
		Kind:        versionsapi.VersionKindImage,
	}
	c.Client.Log.Debugf("Fetching minor version list for version %s", ver.Version)
	minorList, err := c.FetchVersionList(ctx, minorList)
	var notFoundErr *apifetcher.NotFoundError
	if errors.As(err, &notFoundErr) {
		c.Client.Log.Warnf("Minor version list for version %s not found", ver.Version)
		c.Client.Log.Warnf("Skipping update of minor version list")
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("fetching minor version list for version %s: %w", ver.Version, err)
	}

	if !minorList.Contains(ver.Version) {
		c.Client.Log.Warnf("Version %s is not in minor version list %s", ver.Version, minorList.JSONPath())
		c.Client.Log.Warnf("Skipping update of minor version list")
		return nil, nil
	}

	semver.Sort(minorList.Versions)
	for i, v := range minorList.Versions {
		if v == ver.Version {
			minorList.Versions = append(minorList.Versions[:i], minorList.Versions[i+1:]...)
			break
		}
	}

	var latest *versionsapi.Latest
	if len(minorList.Versions) != 0 {
		latest = &versionsapi.Latest{
			Ref:     ver.Ref,
			Stream:  ver.Stream,
			Kind:    versionsapi.VersionKindImage,
			Version: minorList.Versions[len(minorList.Versions)-1],
		}
		c.Client.Log.Debugf("Possible latest version replacement %q", latest.Version)
	}

	if c.Client.DryRun {
		c.Client.Log.Debugf("DryRun: Updating minor version list %s to %v", minorList.JSONPath(), minorList)
		return latest, nil
	}

	c.Client.Log.Debugf("Updating minor version list %s", minorList.JSONPath())
	if err := c.UpdateVersionList(ctx, minorList); err != nil {
		return latest, fmt.Errorf("updating minor version list %s: %w", minorList.JSONPath(), err)
	}

	c.Client.Log.Debugf("Removed version %s from minor version list %s", ver.Version, minorList.JSONPath())
	return latest, nil
}

func (c *VersionsClient) deleteVersionFromLatest(ctx context.Context, ver versionsapi.Version, possibleNewLatest *versionsapi.Latest,
) error {
	latest := versionsapi.Latest{
		Ref:    ver.Ref,
		Stream: ver.Stream,
		Kind:   versionsapi.VersionKindImage,
	}
	c.Client.Log.Debugf("Fetching latest version from %s", latest.JSONPath())
	latest, err := c.FetchVersionLatest(ctx, latest)
	var notFoundErr *apifetcher.NotFoundError
	if errors.As(err, &notFoundErr) {
		c.Client.Log.Warnf("Latest version for %s not found", latest.JSONPath())
		return nil
	} else if err != nil {
		return fmt.Errorf("fetching latest version: %w", err)
	}

	if latest.Version != ver.Version {
		c.Client.Log.Debugf("Latest version is %s, not the deleted version %s", latest.Version, ver.Version)
		return nil
	}

	if possibleNewLatest == nil {
		c.Client.Log.Errorf("Latest version is %s, but no new latest version was found", latest.Version)
		c.Client.Log.Errorf("A manual update of latest at %s might be needed", latest.JSONPath())
		return fmt.Errorf("latest version is %s, but no new latest version was found", latest.Version)
	}

	if c.Client.DryRun {
		c.Client.Log.Debugf("Would update latest version from %s to %s", latest.Version, possibleNewLatest.Version)
		return nil
	}

	c.Client.Log.Infof("Updating latest version from %s to %s", latest.Version, possibleNewLatest.Version)
	if err := c.UpdateVersionLatest(ctx, *possibleNewLatest); err != nil {
		return fmt.Errorf("updating latest version: %w", err)
	}

	return nil
}

// CloseFunc is a function that closes the client.
type CloseFunc func(ctx context.Context) error
