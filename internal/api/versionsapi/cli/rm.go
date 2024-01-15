/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"regexp"
	"strings"
	"time"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	armcomputev5 "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/smithy-go"
	apiclient "github.com/edgelesssys/constellation/v2/internal/api/client"
	"github.com/edgelesssys/constellation/v2/internal/api/versionsapi"
	gaxv2 "github.com/googleapis/gax-go/v2"
	"github.com/spf13/cobra"
)

func newRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove a version/ref",
		Long: `Remove a version/ref from the versions API.

Developers should not use this command directly. It is invoked by the CI/CD pipeline.
Most developers won't have the required permissions to use this command.

❗ If you use the command nevertheless, you better know what you do.
`,
		RunE: runRemove,
		Args: cobra.ExactArgs(0),
	}

	cmd.Flags().String("ref", "", "Ref to delete from.")
	cmd.Flags().String("stream", "", "Stream to delete from.")
	cmd.Flags().String("version", "", "Version to delete. The versioned objects are deleted.")
	cmd.Flags().String("version-path", "", "Short path of a single version to delete. The versioned objects are deleted.")
	cmd.Flags().Bool("all", false, "Delete the entire ref. All versions and versioned objects are deleted.")
	cmd.Flags().Bool("dryrun", false, "Whether to run in dry-run mode (no changes are made)")
	cmd.Flags().String("gcp-project", "constellation-images", "GCP project to use")
	cmd.Flags().String("az-subscription", "0d202bbb-4fa7-4af8-8125-58c269a05435", "Azure subscription to use")
	cmd.Flags().String("az-location", "northeurope", "Azure location to use")
	cmd.Flags().String("az-resource-group", "constellation-images", "Azure resource group to use")

	cmd.MarkFlagsRequiredTogether("stream", "version")
	cmd.MarkFlagsMutuallyExclusive("all", "stream")
	cmd.MarkFlagsMutuallyExclusive("all", "version")
	cmd.MarkFlagsMutuallyExclusive("all", "version-path")
	cmd.MarkFlagsMutuallyExclusive("version-path", "ref")
	cmd.MarkFlagsMutuallyExclusive("version-path", "stream")
	cmd.MarkFlagsMutuallyExclusive("version-path", "version")

	return cmd
}

func runRemove(cmd *cobra.Command, _ []string) (retErr error) {
	flags, err := parseRmFlags(cmd)
	if err != nil {
		return err
	}
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: flags.logLevel}))
	log.Debug(fmt.Sprintf("Parsed flags: %+v", flags))

	log.Debug("Validating flags")
	if err := flags.validate(); err != nil {
		return err
	}

	log.Debug("Creating GCP client")
	gcpClient, err := newGCPClient(cmd.Context(), flags.gcpProject)
	if err != nil {
		return fmt.Errorf("creating GCP client: %w", err)
	}

	log.Debug("Creating AWS client")
	awsClient, err := newAWSClient()
	if err != nil {
		return fmt.Errorf("creating AWS client: %w", err)
	}

	log.Debug("Creating Azure client")
	azClient, err := newAzureClient(flags.azSubscription, flags.azLocation, flags.azResourceGroup)
	if err != nil {
		return fmt.Errorf("creating Azure client: %w", err)
	}

	log.Debug("Creating versions API client")
	verclient, verclientClose, err := versionsapi.NewClient(cmd.Context(), flags.region, flags.bucket, flags.distributionID, flags.dryrun, log)
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}
	defer func() {
		err := verclientClose(cmd.Context())
		if err != nil {
			retErr = errors.Join(retErr, fmt.Errorf("failed to invalidate cache: %w", err))
		}
	}()

	imageClients := rmImageClients{
		version: verclient,
		gcp:     gcpClient,
		aws:     awsClient,
		az:      azClient,
	}

	if flags.all {
		log.Info(fmt.Sprintf("Deleting ref %s", flags.ref))
		if err := deleteRef(cmd.Context(), imageClients, flags.ref, flags.dryrun, log); err != nil {
			return fmt.Errorf("deleting ref: %w", err)
		}
		return nil
	}

	log.Info(fmt.Sprintf("Deleting single version %s", flags.ver.ShortPath()))
	if err := deleteSingleVersion(cmd.Context(), imageClients, flags.ver, flags.dryrun, log); err != nil {
		return fmt.Errorf("deleting single version: %w", err)
	}

	return nil
}

func deleteSingleVersion(ctx context.Context, clients rmImageClients, ver versionsapi.Version, dryrun bool, log *slog.Logger) error {
	var retErr error

	log.Debug(fmt.Sprintf("Deleting images for %s", ver.Version()))
	if err := deleteImage(ctx, clients, ver, dryrun, log); err != nil {
		retErr = errors.Join(retErr, fmt.Errorf("deleting images: %w", err))
	}

	log.Debug(fmt.Sprintf("Deleting version %s from versions API", ver.Version()))
	if err := clients.version.DeleteVersion(ctx, ver); err != nil {
		retErr = errors.Join(retErr, fmt.Errorf("deleting version from versions API: %w", err))
	}

	return retErr
}

func deleteRef(ctx context.Context, clients rmImageClients, ref string, dryrun bool, log *slog.Logger) error {
	var vers []versionsapi.Version
	for _, stream := range []string{"nightly", "console", "debug"} {
		log.Info(fmt.Sprintf("Listing versions of stream %s", stream))

		minorVersions, err := listMinorVersions(ctx, clients.version, ref, stream)
		var notFoundErr *apiclient.NotFoundError
		if errors.As(err, &notFoundErr) {
			log.Debug(fmt.Sprintf("No minor versions found for stream %s", stream))
			continue
		} else if err != nil {
			return fmt.Errorf("listing minor versions for stream %s: %w", stream, err)
		}

		patchVersions, err := listPatchVersions(ctx, clients.version, ref, stream, minorVersions)
		if errors.As(err, &notFoundErr) {
			log.Debug(fmt.Sprintf("No patch versions found for stream %s", stream))
			continue
		} else if err != nil {
			return fmt.Errorf("listing patch versions for stream %s: %w", stream, err)
		}

		vers = append(vers, patchVersions...)
	}
	log.Info(fmt.Sprintf("Found %d versions to delete", len(vers)))

	var retErr error

	for _, ver := range vers {
		if err := deleteImage(ctx, clients, ver, dryrun, log); err != nil {
			retErr = errors.Join(retErr, fmt.Errorf("deleting images for version %s: %w", ver.Version(), err))
		}
	}

	log.Info(fmt.Sprintf("Deleting ref %s from versions API", ref))
	if err := clients.version.DeleteRef(ctx, ref); err != nil {
		retErr = errors.Join(retErr, fmt.Errorf("deleting ref from versions API: %w", err))
	}

	return retErr
}

func deleteImage(ctx context.Context, clients rmImageClients, ver versionsapi.Version, dryrun bool, log *slog.Logger) error {
	var retErr error

	imageInfo := versionsapi.ImageInfo{
		Ref:     ver.Ref(),
		Stream:  ver.Stream(),
		Version: ver.Version(),
	}
	imageInfo, err := clients.version.FetchImageInfo(ctx, imageInfo)
	var notFound *apiclient.NotFoundError
	if errors.As(err, &notFound) {
		log.Warn(fmt.Sprintf("Image info for %s not found", ver.Version()))
		log.Warn("Skipping image deletion")
		return nil
	} else if err != nil {
		return fmt.Errorf("fetching image info: %w", err)
	}

	for _, entry := range imageInfo.List {
		switch entry.CSP {
		case "aws":
			log.Info(fmt.Sprintf("Deleting AWS images from %s", imageInfo.JSONPath()))
			if err := clients.aws.deleteImage(ctx, entry.Reference, entry.Region, dryrun, log); err != nil {
				retErr = errors.Join(retErr, fmt.Errorf("deleting AWS image %s: %w", entry.Reference, err))
			}
		case "gcp":
			log.Info(fmt.Sprintf("Deleting GCP images from %s", imageInfo.JSONPath()))
			if err := clients.gcp.deleteImage(ctx, entry.Reference, dryrun, log); err != nil {
				retErr = errors.Join(retErr, fmt.Errorf("deleting GCP image %s: %w", entry.Reference, err))
			}
		case "azure":
			log.Info(fmt.Sprintf("Deleting Azure images from %s", imageInfo.JSONPath()))
			if err := clients.az.deleteImage(ctx, entry.Reference, dryrun, log); err != nil {
				retErr = errors.Join(retErr, fmt.Errorf("deleting Azure image %s: %w", entry.Reference, err))
			}
		}
	}

	// TODO(katexochen): Implement versions API trash. In case of failure, we should
	// collect the resources that couldn't be deleted and store them in the trash, so
	// that we can retry deleting them later.

	return retErr
}

type rmImageClients struct {
	version *versionsapi.Client
	gcp     *gcpClient
	aws     *awsClient
	az      *azureClient
}

type rmFlags struct {
	ref             string
	stream          string
	version         string
	versionPath     string
	all             bool
	dryrun          bool
	region          string
	bucket          string
	distributionID  string
	gcpProject      string
	azSubscription  string
	azLocation      string
	azResourceGroup string
	logLevel        slog.Level

	ver versionsapi.Version
}

func (f *rmFlags) validate() error {
	if f.ref == versionsapi.ReleaseRef {
		return fmt.Errorf("cannot delete from release ref")
	}

	if f.all {
		if err := versionsapi.ValidateRef(f.ref); err != nil {
			return fmt.Errorf("invalid ref: %w", err)
		}

		if f.ref == "main" {
			return fmt.Errorf("cannot delete 'main' ref")
		}

		return nil
	}

	if f.versionPath != "" {
		ver, err := versionsapi.NewVersionFromShortPath(f.versionPath, versionsapi.VersionKindImage)
		if err != nil {
			return fmt.Errorf("invalid version path: %w", err)
		}
		f.ver = ver

		return nil
	}

	ver, err := versionsapi.NewVersion(f.ref, f.stream, f.version, versionsapi.VersionKindImage)
	if err != nil {
		return fmt.Errorf("creating version: %w", err)
	}
	f.ver = ver

	return nil
}

func parseRmFlags(cmd *cobra.Command) (*rmFlags, error) {
	ref, err := cmd.Flags().GetString("ref")
	if err != nil {
		return nil, err
	}
	ref = versionsapi.CanonicalizeRef(ref)
	stream, err := cmd.Flags().GetString("stream")
	if err != nil {
		return nil, err
	}
	version, err := cmd.Flags().GetString("version")
	if err != nil {
		return nil, err
	}
	versionPath, err := cmd.Flags().GetString("version-path")
	if err != nil {
		return nil, err
	}
	all, err := cmd.Flags().GetBool("all")
	if err != nil {
		return nil, err
	}
	dryrun, err := cmd.Flags().GetBool("dryrun")
	if err != nil {
		return nil, err
	}
	region, err := cmd.Flags().GetString("region")
	if err != nil {
		return nil, err
	}
	bucket, err := cmd.Flags().GetString("bucket")
	if err != nil {
		return nil, err
	}
	distributionID, err := cmd.Flags().GetString("distribution-id")
	if err != nil {
		return nil, err
	}
	gcpProject, err := cmd.Flags().GetString("gcp-project")
	if err != nil {
		return nil, err
	}
	azSubscription, err := cmd.Flags().GetString("az-subscription")
	if err != nil {
		return nil, err
	}
	azLocation, err := cmd.Flags().GetString("az-location")
	if err != nil {
		return nil, err
	}
	azResourceGroup, err := cmd.Flags().GetString("az-resource-group")
	if err != nil {
		return nil, err
	}
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return nil, err
	}
	logLevel := slog.LevelInfo
	if verbose {
		logLevel = slog.LevelDebug
	}

	return &rmFlags{
		ref:             ref,
		stream:          stream,
		version:         version,
		versionPath:     versionPath,
		all:             all,
		dryrun:          dryrun,
		region:          region,
		bucket:          bucket,
		distributionID:  distributionID,
		gcpProject:      gcpProject,
		azSubscription:  azSubscription,
		azLocation:      azLocation,
		azResourceGroup: azResourceGroup,
		logLevel:        logLevel,
	}, nil
}

type awsClient struct {
	ec2 ec2API
}

// newAWSClient creates a new awsClient.
// Requires IAM permission 'ec2:DeregisterImage'.
func newAWSClient() (*awsClient, error) {
	return &awsClient{}, nil
}

type ec2API interface {
	DeregisterImage(ctx context.Context, params *ec2.DeregisterImageInput, optFns ...func(*ec2.Options),
	) (*ec2.DeregisterImageOutput, error)
	DescribeImages(ctx context.Context, params *ec2.DescribeImagesInput, optFns ...func(*ec2.Options),
	) (*ec2.DescribeImagesOutput, error)
	DeleteSnapshot(ctx context.Context, params *ec2.DeleteSnapshotInput, optFns ...func(*ec2.Options),
	) (*ec2.DeleteSnapshotOutput, error)
}

func (a *awsClient) deleteImage(ctx context.Context, ami string, region string, dryrun bool, log *slog.Logger) error {
	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return err
	}
	a.ec2 = ec2.NewFromConfig(cfg)
	log.Debug(fmt.Sprintf("Deleting resources in AWS region %s", region))

	snapshotID, err := a.getSnapshotID(ctx, ami, log)
	if err != nil {
		log.Warn(fmt.Sprintf("Failed to get AWS snapshot ID for image %s: %v", ami, err))
	}

	if err := a.deregisterImage(ctx, ami, dryrun, log); err != nil {
		return fmt.Errorf("deregistering image %s: %w", ami, err)
	}

	if snapshotID != "" {
		if err := a.deleteSnapshot(ctx, snapshotID, dryrun, log); err != nil {
			return fmt.Errorf("deleting snapshot %s: %w", snapshotID, err)
		}
	}

	return nil
}

func (a *awsClient) deregisterImage(ctx context.Context, ami string, dryrun bool, log *slog.Logger) error {
	log.Debug(fmt.Sprintf("Deregistering image %s", ami))

	deregisterReq := ec2.DeregisterImageInput{
		ImageId: &ami,
		DryRun:  &dryrun,
	}
	_, err := a.ec2.DeregisterImage(ctx, &deregisterReq)
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) &&
		(apiErr.ErrorCode() == "InvalidAMIID.NotFound" ||
			apiErr.ErrorCode() == "InvalidAMIID.Unavailable") {
		log.Warn(fmt.Sprintf("AWS image %s not found", ami))
		return nil
	}

	return err
}

func (a *awsClient) getSnapshotID(ctx context.Context, ami string, log *slog.Logger) (string, error) {
	log.Debug(fmt.Sprintf("Describing image %s", ami))

	req := ec2.DescribeImagesInput{
		ImageIds: []string{ami},
	}
	resp, err := a.ec2.DescribeImages(ctx, &req)
	if err != nil {
		return "", fmt.Errorf("describing image %s: %w", ami, err)
	}

	if len(resp.Images) == 0 {
		return "", fmt.Errorf("image %s not found", ami)
	}

	if len(resp.Images) > 1 {
		return "", fmt.Errorf("found multiple images with ami %s", ami)
	}
	image := resp.Images[0]

	if len(image.BlockDeviceMappings) != 1 {
		return "", fmt.Errorf("found %d block device mappings for image %s, expected 1", len(image.BlockDeviceMappings), ami)
	}
	if image.BlockDeviceMappings[0].Ebs == nil {
		return "", fmt.Errorf("image %s does not have an EBS block device mapping", ami)
	}
	ebs := image.BlockDeviceMappings[0].Ebs

	if ebs.SnapshotId == nil {
		return "", fmt.Errorf("image %s does not have an EBS snapshot", ami)
	}
	snapshotID := *ebs.SnapshotId

	return snapshotID, nil
}

func (a *awsClient) deleteSnapshot(ctx context.Context, snapshotID string, dryrun bool, log *slog.Logger) error {
	log.Debug(fmt.Sprintf("Deleting AWS snapshot %s", snapshotID))

	req := ec2.DeleteSnapshotInput{
		SnapshotId: &snapshotID,
		DryRun:     &dryrun,
	}
	_, err := a.ec2.DeleteSnapshot(ctx, &req)
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) &&
		(apiErr.ErrorCode() == "InvalidSnapshot.NotFound" ||
			apiErr.ErrorCode() == "InvalidSnapshot.Unavailable") {
		log.Warn(fmt.Sprintf("AWS snapshot %s not found", snapshotID))
		return nil
	}

	return err
}

type gcpClient struct {
	project string
	compute gcpComputeAPI
}

func newGCPClient(ctx context.Context, project string) (*gcpClient, error) {
	compute, err := compute.NewImagesRESTClient(ctx)
	if err != nil {
		return nil, err
	}

	return &gcpClient{
		compute: compute,
		project: project,
	}, nil
}

type gcpComputeAPI interface {
	Delete(ctx context.Context, req *computepb.DeleteImageRequest, opts ...gaxv2.CallOption,
	) (*compute.Operation, error)
	io.Closer
}

func (g *gcpClient) deleteImage(ctx context.Context, imageURI string, dryrun bool, log *slog.Logger) error {
	// Extract image name from image URI
	// Expected input into function: "projects/constellation-images/global/images/v2-6-0-stable"
	// Required for computepb.DeleteImageRequest: "v2-6-0-stable"
	imageURIParts := strings.Split(imageURI, "/")
	image := imageURIParts[len(imageURIParts)-1] // Don't need to check if len(imageURIParts) == 0 since sep is not empty and thus length must be ≥ 1

	req := &computepb.DeleteImageRequest{
		Image:   image,
		Project: g.project,
	}

	if dryrun {
		log.Debug(fmt.Sprintf("DryRun: delete image request: %v", req))
		return nil
	}

	log.Debug(fmt.Sprintf("Deleting image %s", image))
	op, err := g.compute.Delete(ctx, req)
	if err != nil && strings.Contains(err.Error(), "404") {
		log.Warn(fmt.Sprintf("GCP image %s not found", image))
		return nil
	} else if err != nil {
		return fmt.Errorf("deleting image %s: %w", image, err)
	}

	log.Debug("Waiting for operation to finish")
	if err := op.Wait(ctx); err != nil {
		return fmt.Errorf("waiting for operation: %w", err)
	}

	return nil
}

func (g *gcpClient) Close() error {
	return g.compute.Close()
}

type azureClient struct {
	subscription  string
	location      string
	resourceGroup string
	galleries     azureGalleriesAPI
	image         azureGalleriesImageAPI
	imageVersions azureGalleriesImageVersionAPI
}

func newAzureClient(subscription, location, resourceGroup string) (*azureClient, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatal(err)
	}
	galleriesClient, err := armcomputev5.NewGalleriesClient(subscription, cred, nil)
	if err != nil {
		return nil, err
	}
	galleriesImageClient, err := armcomputev5.NewGalleryImagesClient(subscription, cred, nil)
	if err != nil {
		return nil, err
	}
	galleriesImageVersionClient, err := armcomputev5.NewGalleryImageVersionsClient(subscription, cred, nil)
	if err != nil {
		return nil, err
	}

	return &azureClient{
		subscription:  subscription,
		location:      location,
		resourceGroup: resourceGroup,
		galleries:     galleriesClient,
		image:         galleriesImageClient,
		imageVersions: galleriesImageVersionClient,
	}, nil
}

type azureGalleriesAPI interface {
	NewListPager(options *armcomputev5.GalleriesClientListOptions,
	) *runtime.Pager[armcomputev5.GalleriesClientListResponse]
}

type azureGalleriesImageAPI interface {
	BeginDelete(ctx context.Context, resourceGroupName string, galleryName string, galleryImageName string,
		options *armcomputev5.GalleryImagesClientBeginDeleteOptions,
	) (*runtime.Poller[armcomputev5.GalleryImagesClientDeleteResponse], error)
}

type azureGalleriesImageVersionAPI interface {
	NewListByGalleryImagePager(resourceGroupName string, galleryName string, galleryImageName string,
		options *armcomputev5.GalleryImageVersionsClientListByGalleryImageOptions,
	) *runtime.Pager[armcomputev5.GalleryImageVersionsClientListByGalleryImageResponse]

	BeginDelete(ctx context.Context, resourceGroupName string, galleryName string, galleryImageName string,
		galleryImageVersionName string, options *armcomputev5.GalleryImageVersionsClientBeginDeleteOptions,
	) (*runtime.Poller[armcomputev5.GalleryImageVersionsClientDeleteResponse], error)
}

var (
	azImageRegex          = regexp.MustCompile("^/subscriptions/[[:alnum:]._-]+/resourceGroups/([[:alnum:]._-]+)/providers/Microsoft.Compute/galleries/([[:alnum:]._-]+)/images/([[:alnum:]._-]+)/versions/([[:alnum:]._-]+)$")
	azCommunityImageRegex = regexp.MustCompile("^/CommunityGalleries/([[:alnum:]-]+)/Images/([[:alnum:]._-]+)/Versions/([[:alnum:]._-]+)$")
)

func (a *azureClient) deleteImage(ctx context.Context, image string, dryrun bool, log *slog.Logger) error {
	azImage, err := a.parseImage(ctx, image, log)
	if err != nil {
		return err
	}

	if dryrun {
		log.Debug(fmt.Sprintf("DryRun: delete image %v", azImage))
		return nil
	}

	log.Debug(fmt.Sprintf("Deleting image %q, version %q", azImage.imageDefinition, azImage.version))
	poller, err := a.imageVersions.BeginDelete(ctx, azImage.resourceGroup, azImage.gallery,
		azImage.imageDefinition, azImage.version, nil)
	if err != nil {
		return fmt.Errorf("begin delete image version: %w", err)
	}

	log.Debug("Waiting for operation to finish")
	if _, err := poller.PollUntilDone(ctx, nil); err != nil {
		return fmt.Errorf("waiting for operation: %w", err)
	}

	log.Debug(fmt.Sprintf("Checking if image definition %q still has versions left", azImage.imageDefinition))
	pager := a.imageVersions.NewListByGalleryImagePager(azImage.resourceGroup, azImage.gallery,
		azImage.imageDefinition, nil)
	for pager.More() {
		nextResult, err := pager.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("listing image versions of image definition %s: %w", azImage.imageDefinition, err)
		}
		if len(nextResult.Value) != 0 {
			log.Debug(fmt.Sprintf("Image definition %q still has versions left, won't be deleted", azImage.imageDefinition))
			return nil
		}
	}

	time.Sleep(15 * time.Second) // Azure needs time understand that there is no version left...

	log.Debug(fmt.Sprintf("Deleting image definition %s", azImage.imageDefinition))
	op, err := a.image.BeginDelete(ctx, azImage.resourceGroup, azImage.gallery, azImage.imageDefinition, nil)
	if err != nil {
		return fmt.Errorf("deleting image definition %s: %w", azImage.imageDefinition, err)
	}

	log.Debug("Waiting for operation to finish")
	if _, err := op.PollUntilDone(ctx, nil); err != nil {
		return fmt.Errorf("waiting for operation: %w", err)
	}

	return nil
}

type azImage struct {
	resourceGroup   string
	gallery         string
	imageDefinition string
	version         string
}

func (a *azureClient) parseImage(ctx context.Context, image string, log *slog.Logger) (azImage, error) {
	if m := azImageRegex.FindStringSubmatch(image); len(m) == 5 {
		log.Debug(fmt.Sprintf(
			"Image matches local image format, resource group: %s, gallery: %s, image definition: %s, version: %s",
			m[1], m[2], m[3], m[4],
		))
		return azImage{
			resourceGroup:   m[1],
			gallery:         m[2],
			imageDefinition: m[3],
			version:         m[4],
		}, nil
	}

	if !azCommunityImageRegex.MatchString(image) {
		return azImage{}, fmt.Errorf("invalid image %s", image)
	}

	m := azCommunityImageRegex.FindStringSubmatch(image)
	galleryPublicName := m[1]
	imageDefinition := m[2]
	version := m[3]

	log.Debug(fmt.Sprintf(
		"Image matches community image format, gallery public name: %s, image definition: %s, version: %s",
		galleryPublicName, imageDefinition, version,
	))

	var galleryName string
	pager := a.galleries.NewListPager(nil)
	for pager.More() {
		nextResult, err := pager.NextPage(ctx)
		if err != nil {
			return azImage{}, fmt.Errorf("failed to advance page: %w", err)
		}
		for _, v := range nextResult.Value {
			if v.Name == nil {
				log.Debug("Skipping gallery with nil name")
				continue
			}
			if v.Properties.SharingProfile == nil {
				log.Debug(fmt.Sprintf("Skipping gallery %s with nil sharing profile", *v.Name))
				continue
			}
			if v.Properties.SharingProfile.CommunityGalleryInfo == nil {
				log.Debug(fmt.Sprintf("Skipping gallery %s with nil community gallery info", *v.Name))
				continue
			}
			if v.Properties.SharingProfile.CommunityGalleryInfo.PublicNames == nil {
				log.Debug(fmt.Sprintf("Skipping gallery %s with nil public names", *v.Name))
				continue
			}
			for _, publicName := range v.Properties.SharingProfile.CommunityGalleryInfo.PublicNames {
				if publicName == nil {
					log.Debug("Skipping nil public name")
					continue
				}
				if *publicName == galleryPublicName {
					galleryName = *v.Name
					break
				}

			}
			if galleryName != "" {
				break
			}
		}
	}

	if galleryName == "" {
		return azImage{}, fmt.Errorf("failed to find gallery for public name %s", galleryPublicName)
	}

	return azImage{
		resourceGroup:   a.resourceGroup,
		gallery:         galleryName,
		imageDefinition: imageDefinition,
		version:         version,
	}, nil
}
