/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package versionsapi

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"path"

	"golang.org/x/mod/semver"

	apiclient "github.com/edgelesssys/constellation/v2/internal/api/client"
	"github.com/edgelesssys/constellation/v2/internal/constants"
)

// Client is a client for the versions API.
type Client struct {
	*apiclient.Client
	clientClose func(ctx context.Context) error
}

// NewClient creates a new client for the versions API.
func NewClient(ctx context.Context, region, bucket, distributionID string, dryRun bool,
	log *slog.Logger,
) (*Client, CloseFunc, error) {
	genericClient, genericClientClose, err := apiclient.NewClient(ctx, region, bucket, distributionID, dryRun, log)
	versionsClient := &Client{
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
	log *slog.Logger,
) (*Client, CloseFunc, error) {
	genericClient, genericClientClose, err := apiclient.NewReadOnlyClient(ctx, region, bucket, distributionID, log)
	if err != nil {
		return nil, nil, err
	}
	versionsClient := &Client{
		genericClient,
		genericClientClose,
	}
	versionsClientClose := func(ctx context.Context) error {
		return versionsClient.Close(ctx)
	}
	return versionsClient, versionsClientClose, err
}

// Close closes the client.
func (c *Client) Close(ctx context.Context) error {
	if c.clientClose == nil {
		return nil
	}
	return c.clientClose(ctx)
}

// FetchVersionList fetches the given version list from the versions API.
func (c *Client) FetchVersionList(ctx context.Context, list List) (List, error) {
	return apiclient.Fetch(ctx, c.Client, list)
}

// UpdateVersionList updates the given version list in the versions API.
func (c *Client) UpdateVersionList(ctx context.Context, list List) error {
	semver.Sort(list.Versions)
	return apiclient.Update(ctx, c.Client, list)
}

// FetchVersionLatest fetches the latest version from the versions API.
func (c *Client) FetchVersionLatest(ctx context.Context, latest Latest) (Latest, error) {
	return apiclient.Fetch(ctx, c.Client, latest)
}

// UpdateVersionLatest updates the latest version in the versions API.
func (c *Client) UpdateVersionLatest(ctx context.Context, latest Latest) error {
	return apiclient.Update(ctx, c.Client, latest)
}

// FetchImageInfo fetches the given image info from the versions API.
func (c *Client) FetchImageInfo(ctx context.Context, imageInfo ImageInfo) (ImageInfo, error) {
	return apiclient.Fetch(ctx, c.Client, imageInfo)
}

// UpdateImageInfo updates the given image info in the versions API.
func (c *Client) UpdateImageInfo(ctx context.Context, imageInfo ImageInfo) error {
	return apiclient.Update(ctx, c.Client, imageInfo)
}

// FetchCLIInfo fetches the given CLI info from the versions API.
func (c *Client) FetchCLIInfo(ctx context.Context, cliInfo CLIInfo) (CLIInfo, error) {
	return apiclient.Fetch(ctx, c.Client, cliInfo)
}

// UpdateCLIInfo updates the given CLI info in the versions API.
func (c *Client) UpdateCLIInfo(ctx context.Context, cliInfo CLIInfo) error {
	return apiclient.Update(ctx, c.Client, cliInfo)
}

// DeleteRef deletes the given ref from the versions API.
func (c *Client) DeleteRef(ctx context.Context, ref string) error {
	if err := ValidateRef(ref); err != nil {
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
func (c *Client) DeleteVersion(ctx context.Context, ver Version) error {
	var retErr error

	c.Client.Logger.Debug(fmt.Sprintf("Deleting version %s from minor version list", ver.version))
	possibleNewLatest, err := c.deleteVersionFromMinorVersionList(ctx, ver)
	if err != nil {
		retErr = errors.Join(retErr, fmt.Errorf("removing from minor version list: %w", err))
	}

	c.Client.Logger.Debug(fmt.Sprintf("Checking latest version for %s", ver.version))
	if err := c.deleteVersionFromLatest(ctx, ver, possibleNewLatest); err != nil {
		retErr = errors.Join(retErr, fmt.Errorf("updating latest version: %w", err))
	}

	c.Client.Logger.Debug(fmt.Sprintf("Deleting artifact path %s for %s", ver.ArtifactPath(APIV1), ver.version))
	if err := c.Client.DeletePath(ctx, ver.ArtifactPath(APIV1)); err != nil {
		retErr = errors.Join(retErr, fmt.Errorf("deleting artifact path: %w", err))
	}

	return retErr
}

func (c *Client) deleteVersionFromMinorVersionList(ctx context.Context, ver Version,
) (*Latest, error) {
	minorList := List{
		Ref:         ver.ref,
		Stream:      ver.stream,
		Granularity: GranularityMinor,
		Base:        ver.WithGranularity(GranularityMinor),
		Kind:        VersionKindImage,
	}
	c.Client.Logger.Debug(fmt.Sprintf("Fetching minor version list for version %s", ver.version))
	minorList, err := c.FetchVersionList(ctx, minorList)
	var notFoundErr *apiclient.NotFoundError
	if errors.As(err, &notFoundErr) {
		c.Client.Logger.Warn(fmt.Sprintf("Minor version list for version %s not found", ver.version))
		c.Client.Logger.Warn("Skipping update of minor version list")
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("fetching minor version list for version %s: %w", ver.version, err)
	}

	if !minorList.Contains(ver.version) {
		c.Client.Logger.Warn(fmt.Sprintf("Version %s is not in minor version list %s", ver.version, minorList.JSONPath()))
		c.Client.Logger.Warn("Skipping update of minor version list")
		return nil, nil
	}

	semver.Sort(minorList.Versions)
	for i, v := range minorList.Versions {
		if v == ver.version {
			minorList.Versions = append(minorList.Versions[:i], minorList.Versions[i+1:]...)
			break
		}
	}

	var latest *Latest
	if len(minorList.Versions) != 0 {
		latest = &Latest{
			Ref:     ver.ref,
			Stream:  ver.stream,
			Kind:    VersionKindImage,
			Version: minorList.Versions[len(minorList.Versions)-1],
		}
		c.Client.Logger.Debug(fmt.Sprintf("Possible latest version replacement %q", latest.Version))
	}

	if c.Client.DryRun {
		c.Client.Logger.Debug(fmt.Sprintf("DryRun: Updating minor version list %s to %v", minorList.JSONPath(), minorList))
		return latest, nil
	}

	c.Client.Logger.Debug(fmt.Sprintf("Updating minor version list %s", minorList.JSONPath()))
	if err := c.UpdateVersionList(ctx, minorList); err != nil {
		return latest, fmt.Errorf("updating minor version list %s: %w", minorList.JSONPath(), err)
	}

	c.Client.Logger.Debug(fmt.Sprintf("Removed version %s from minor version list %s", ver.version, minorList.JSONPath()))
	return latest, nil
}

func (c *Client) deleteVersionFromLatest(ctx context.Context, ver Version, possibleNewLatest *Latest,
) error {
	latest := Latest{
		Ref:    ver.ref,
		Stream: ver.stream,
		Kind:   VersionKindImage,
	}
	c.Client.Logger.Debug(fmt.Sprintf("Fetching latest version from %s", latest.JSONPath()))
	latest, err := c.FetchVersionLatest(ctx, latest)
	var notFoundErr *apiclient.NotFoundError
	if errors.As(err, &notFoundErr) {
		c.Client.Logger.Warn(fmt.Sprintf("Latest version for %s not found", latest.JSONPath()))
		return nil
	} else if err != nil {
		return fmt.Errorf("fetching latest version: %w", err)
	}

	if latest.Version != ver.version {
		c.Client.Logger.Debug(fmt.Sprintf("Latest version is %s, not the deleted version %s", latest.Version, ver.version))
		return nil
	}

	if possibleNewLatest == nil {
		c.Client.Logger.Error(fmt.Sprintf("Latest version is %s, but no new latest version was found", latest.Version))
		c.Client.Logger.Error(fmt.Sprintf("A manual update of latest at %s might be needed", latest.JSONPath()))
		return fmt.Errorf("latest version is %s, but no new latest version was found", latest.Version)
	}

	if c.Client.DryRun {
		c.Client.Logger.Debug(fmt.Sprintf("Would update latest version from %s to %s", latest.Version, possibleNewLatest.Version))
		return nil
	}

	c.Client.Logger.Info(fmt.Sprintf("Updating latest version from %s to %s", latest.Version, possibleNewLatest.Version))
	if err := c.UpdateVersionLatest(ctx, *possibleNewLatest); err != nil {
		return fmt.Errorf("updating latest version: %w", err)
	}

	return nil
}

// CloseFunc is a function that closes the client.
type CloseFunc func(ctx context.Context) error
