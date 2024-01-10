/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// package azure implements uploading os images to azure.
package azure

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	armcomputev5 "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/pageblob"
	"github.com/edgelesssys/constellation/v2/internal/api/versionsapi"
	"github.com/edgelesssys/constellation/v2/internal/osimage"
)

// Uploader can upload and remove os images on Azure.
type Uploader struct {
	subscription      string
	location          string
	resourceGroup     string
	pollingFrequency  time.Duration
	disks             azureDiskAPI
	managedImages     azureManagedImageAPI
	blob              sasBlobUploader
	galleries         azureGalleriesAPI
	image             azureGalleriesImageAPI
	imageVersions     azureGalleriesImageVersionAPI
	communityVersions azureCommunityGalleryImageVersionAPI

	log *slog.Logger
}

// New creates a new Uploader.
func New(subscription, location, resourceGroup string, log *slog.Logger) (*Uploader, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}
	diskClient, err := armcomputev5.NewDisksClient(subscription, cred, nil)
	if err != nil {
		return nil, err
	}
	managedImagesClient, err := armcomputev5.NewImagesClient(subscription, cred, nil)
	if err != nil {
		return nil, err
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
	communityImageVersionClient, err := armcomputev5.NewCommunityGalleryImageVersionsClient(subscription, cred, nil)
	if err != nil {
		return nil, err
	}

	return &Uploader{
		subscription:     subscription,
		location:         location,
		resourceGroup:    resourceGroup,
		pollingFrequency: pollingFrequency,
		disks:            diskClient,
		managedImages:    managedImagesClient,
		blob: func(sasBlobURL string) (azurePageblobAPI, error) {
			return pageblob.NewClientWithNoCredential(sasBlobURL, nil)
		},
		galleries:         galleriesClient,
		image:             galleriesImageClient,
		imageVersions:     galleriesImageVersionClient,
		communityVersions: communityImageVersionClient,
		log:               log,
	}, nil
}

// Upload uploads an OS image to Azure.
func (u *Uploader) Upload(ctx context.Context, req *osimage.UploadRequest) ([]versionsapi.ImageInfoEntry, error) {
	formattedTime := req.Timestamp.Format(timestampFormat)
	diskName := fmt.Sprintf("constellation-%s-%s-%s", req.Version.Stream(), formattedTime, req.AttestationVariant)
	var sigName string
	switch req.Version.Stream() {
	case "stable":
		sigName = sigNameStable
	case "debug":
		sigName = sigNameDebug
	default:
		sigName = sigNameDefault
	}
	definitionName := imageOffer(req.Version)
	versionName, err := imageVersion(req.Version, req.Timestamp)
	if err != nil {
		return nil, fmt.Errorf("determining image version name: %w", err)
	}

	// ensure new image can be uploaded by deleting existing resources using the same name
	if err := u.ensureImageVersionDeleted(ctx, sigName, definitionName, versionName); err != nil {
		return nil, fmt.Errorf("pre-cleaning: ensuring no image version using the same name exists: %w", err)
	}
	if err := u.ensureManagedImageDeleted(ctx, diskName); err != nil {
		return nil, fmt.Errorf("pre-cleaning: ensuring no managed image using the same name exists: %w", err)
	}
	if err := u.ensureDiskDeleted(ctx, diskName); err != nil {
		return nil, fmt.Errorf("pre-cleaning: ensuring no temporary disk using the same name exists: %w", err)
	}

	diskID, err := u.createDisk(ctx, diskName, DiskTypeNormal, req.Image, nil, req.Size)
	if err != nil {
		return nil, fmt.Errorf("creating disk: %w", err)
	}
	defer func() {
		// cleanup temp disk
		err := u.ensureDiskDeleted(ctx, diskName)
		if err != nil {
			u.log.Error("post-cleaning: deleting disk image: %v", err)
		}
	}()
	managedImageID, err := u.createManagedImage(ctx, diskName, diskID)
	if err != nil {
		return nil, fmt.Errorf("creating managed image: %w", err)
	}
	if err := u.ensureSIG(ctx, sigName); err != nil {
		return nil, fmt.Errorf("ensuring sig exists: %w", err)
	}
	if err := u.ensureImageDefinition(ctx, sigName, definitionName, req.Version, req.AttestationVariant); err != nil {
		return nil, fmt.Errorf("ensuring image definition exists: %w", err)
	}

	unsharedImageVersionID, err := u.createImageVersion(ctx, sigName, definitionName, versionName, managedImageID)
	if err != nil {
		return nil, fmt.Errorf("creating image version: %w", err)
	}

	imageReference, err := u.getImageReference(ctx, sigName, definitionName, versionName, unsharedImageVersionID)
	if err != nil {
		return nil, fmt.Errorf("getting image reference: %w", err)
	}

	return []versionsapi.ImageInfoEntry{
		{
			CSP:                "azure",
			AttestationVariant: req.AttestationVariant,
			Reference:          imageReference,
		},
	}, nil
}

// createDisk creates and initializes (uploads contents of) an azure disk.
func (u *Uploader) createDisk(ctx context.Context, diskName string, diskType DiskType, img io.ReadSeeker, vmgs io.ReadSeeker, size int64) (string, error) {
	u.log.Debug("Creating disk %s in %s", diskName, u.resourceGroup)
	if diskType == DiskTypeWithVMGS && vmgs == nil {
		return "", errors.New("cannot create disk with vmgs: vmgs reader is nil")
	}

	var createOption armcomputev5.DiskCreateOption
	var requestVMGSSAS bool
	switch diskType {
	case DiskTypeNormal:
		createOption = armcomputev5.DiskCreateOptionUpload
	case DiskTypeWithVMGS:
		createOption = armcomputev5.DiskCreateOptionUploadPreparedSecure
		requestVMGSSAS = true
	}
	disk := armcomputev5.Disk{
		Location: &u.location,
		Properties: &armcomputev5.DiskProperties{
			CreationData: &armcomputev5.CreationData{
				CreateOption:    &createOption,
				UploadSizeBytes: toPtr(size),
			},
			HyperVGeneration: toPtr(armcomputev5.HyperVGenerationV2),
			OSType:           toPtr(armcomputev5.OperatingSystemTypesLinux),
		},
	}
	createPoller, err := u.disks.BeginCreateOrUpdate(ctx, u.resourceGroup, diskName, disk, &armcomputev5.DisksClientBeginCreateOrUpdateOptions{})
	if err != nil {
		return "", fmt.Errorf("creating disk: %w", err)
	}
	createdDisk, err := createPoller.PollUntilDone(ctx, &runtime.PollUntilDoneOptions{Frequency: u.pollingFrequency})
	if err != nil {
		return "", fmt.Errorf("waiting for disk to be created: %w", err)
	}

	u.log.Debug("Granting temporary upload permissions via SAS token")
	accessGrant := armcomputev5.GrantAccessData{
		Access:                   toPtr(armcomputev5.AccessLevelWrite),
		DurationInSeconds:        toPtr(int32(uploadAccessDuration)),
		GetSecureVMGuestStateSAS: &requestVMGSSAS,
	}
	accessPoller, err := u.disks.BeginGrantAccess(ctx, u.resourceGroup, diskName, accessGrant, &armcomputev5.DisksClientBeginGrantAccessOptions{})
	if err != nil {
		return "", fmt.Errorf("generating disk sas token: %w", err)
	}
	accesPollerResp, err := accessPoller.PollUntilDone(ctx, &runtime.PollUntilDoneOptions{Frequency: u.pollingFrequency})
	if err != nil {
		return "", fmt.Errorf("waiting for sas token: %w", err)
	}

	if requestVMGSSAS {
		u.log.Debug("Uploading vmgs")
		vmgsSize, err := vmgs.Seek(0, io.SeekEnd)
		if err != nil {
			return "", err
		}
		if _, err := vmgs.Seek(0, io.SeekStart); err != nil {
			return "", err
		}
		if accesPollerResp.SecurityDataAccessSAS == nil {
			return "", errors.New("uploading vmgs: grant access returned no vmgs sas")
		}
		if err := uploadBlob(ctx, *accesPollerResp.SecurityDataAccessSAS, vmgs, vmgsSize, u.blob); err != nil {
			return "", fmt.Errorf("uploading vmgs: %w", err)
		}
	}
	u.log.Debug("Uploading os image")
	if accesPollerResp.AccessSAS == nil {
		return "", errors.New("uploading disk: grant access returned no disk sas")
	}
	if err := uploadBlob(ctx, *accesPollerResp.AccessSAS, img, size, u.blob); err != nil {
		return "", fmt.Errorf("uploading image: %w", err)
	}
	revokePoller, err := u.disks.BeginRevokeAccess(ctx, u.resourceGroup, diskName, &armcomputev5.DisksClientBeginRevokeAccessOptions{})
	if err != nil {
		return "", fmt.Errorf("revoking disk sas token: %w", err)
	}
	if _, err := revokePoller.PollUntilDone(ctx, &runtime.PollUntilDoneOptions{Frequency: u.pollingFrequency}); err != nil {
		return "", fmt.Errorf("waiting for sas token revocation: %w", err)
	}
	if createdDisk.ID == nil {
		return "", errors.New("created disk has no id")
	}
	return *createdDisk.ID, nil
}

func (u *Uploader) ensureDiskDeleted(ctx context.Context, diskName string) error {
	_, err := u.disks.Get(ctx, u.resourceGroup, diskName, &armcomputev5.DisksClientGetOptions{})
	if err != nil {
		u.log.Debug("Disk %s in %s doesn't exist. Nothing to clean up.", diskName, u.resourceGroup)
		return nil
	}
	u.log.Debug("Deleting disk %s in %s", diskName, u.resourceGroup)
	deletePoller, err := u.disks.BeginDelete(ctx, u.resourceGroup, diskName, &armcomputev5.DisksClientBeginDeleteOptions{})
	if err != nil {
		return fmt.Errorf("deleting disk: %w", err)
	}
	if _, err = deletePoller.PollUntilDone(ctx, &runtime.PollUntilDoneOptions{Frequency: u.pollingFrequency}); err != nil {
		return fmt.Errorf("waiting for disk to be deleted: %w", err)
	}
	return nil
}

func (u *Uploader) createManagedImage(ctx context.Context, imageName string, diskID string) (string, error) {
	u.log.Debug("Creating managed image %s in %s", imageName, u.resourceGroup)
	image := armcomputev5.Image{
		Location: &u.location,
		Properties: &armcomputev5.ImageProperties{
			HyperVGeneration: toPtr(armcomputev5.HyperVGenerationTypesV2),
			StorageProfile: &armcomputev5.ImageStorageProfile{
				OSDisk: &armcomputev5.ImageOSDisk{
					OSState: toPtr(armcomputev5.OperatingSystemStateTypesGeneralized),
					OSType:  toPtr(armcomputev5.OperatingSystemTypesLinux),
					ManagedDisk: &armcomputev5.SubResource{
						ID: &diskID,
					},
				},
			},
		},
	}
	createPoller, err := u.managedImages.BeginCreateOrUpdate(
		ctx, u.resourceGroup, imageName, image,
		&armcomputev5.ImagesClientBeginCreateOrUpdateOptions{},
	)
	if err != nil {
		return "", fmt.Errorf("creating managed image: %w", err)
	}
	createdImage, err := createPoller.PollUntilDone(ctx, &runtime.PollUntilDoneOptions{Frequency: u.pollingFrequency})
	if err != nil {
		return "", fmt.Errorf("waiting for image to be created: %w", err)
	}
	if createdImage.ID == nil {
		return "", errors.New("created image has no id")
	}
	return *createdImage.ID, nil
}

func (u *Uploader) ensureManagedImageDeleted(ctx context.Context, imageName string) error {
	_, err := u.managedImages.Get(ctx, u.resourceGroup, imageName, &armcomputev5.ImagesClientGetOptions{})
	if err != nil {
		u.log.Debug("Managed image %s in %s doesn't exist. Nothing to clean up.", imageName, u.resourceGroup)
		return nil
	}
	u.log.Debug("Deleting managed image %s in %s", imageName, u.resourceGroup)
	deletePoller, err := u.managedImages.BeginDelete(ctx, u.resourceGroup, imageName, &armcomputev5.ImagesClientBeginDeleteOptions{})
	if err != nil {
		return fmt.Errorf("deleting image: %w", err)
	}
	if _, err = deletePoller.PollUntilDone(ctx, &runtime.PollUntilDoneOptions{Frequency: u.pollingFrequency}); err != nil {
		return fmt.Errorf("waiting for image to be deleted: %w", err)
	}
	return nil
}

// ensureSIG creates a SIG if it does not exist yet.
func (u *Uploader) ensureSIG(ctx context.Context, sigName string) error {
	_, err := u.galleries.Get(ctx, u.resourceGroup, sigName, &armcomputev5.GalleriesClientGetOptions{})
	if err == nil {
		u.log.Debug("Image gallery %s in %s exists", sigName, u.resourceGroup)
		return nil
	}
	u.log.Debug("Creating image gallery %s in %s", sigName, u.resourceGroup)
	gallery := armcomputev5.Gallery{
		Location: &u.location,
	}
	createPoller, err := u.galleries.BeginCreateOrUpdate(ctx, u.resourceGroup, sigName, gallery,
		&armcomputev5.GalleriesClientBeginCreateOrUpdateOptions{},
	)
	if err != nil {
		return fmt.Errorf("creating image gallery: %w", err)
	}
	if _, err = createPoller.PollUntilDone(ctx, &runtime.PollUntilDoneOptions{Frequency: u.pollingFrequency}); err != nil {
		return fmt.Errorf("waiting for image gallery to be created: %w", err)
	}
	return nil
}

// ensureImageDefinition creates an image definition (component of a SIG) if it does not exist yet.
func (u *Uploader) ensureImageDefinition(ctx context.Context, sigName, definitionName string, version versionsapi.Version, attestationVariant string) error {
	_, err := u.image.Get(ctx, u.resourceGroup, sigName, definitionName, &armcomputev5.GalleryImagesClientGetOptions{})
	if err == nil {
		u.log.Debug(fmt.Sprintf("Image definition %s/%s in %s exists", sigName, definitionName, u.resourceGroup))
		return nil
	}
	u.log.Debug(fmt.Sprintf("Creating image definition  %s/%s in %s", sigName, definitionName, u.resourceGroup))
	var securityType string
	// TODO(malt3): This needs to allow the *Supported or the normal variant
	// based on wether a VMGS was provided or not.
	// VMGS provided: ConfidentialVM
	// No VMGS provided: ConfidentialVMSupported
	switch strings.ToLower(attestationVariant) {
	case "azure-sev-snp":
		securityType = string("ConfidentialVMSupported")
	case "azure-trustedlaunch":
		securityType = string(armcomputev5.SecurityTypesTrustedLaunch)
	}
	offer := imageOffer(version)

	galleryImage := armcomputev5.GalleryImage{
		Location: &u.location,
		Properties: &armcomputev5.GalleryImageProperties{
			Identifier: &armcomputev5.GalleryImageIdentifier{
				Offer:     &offer,
				Publisher: toPtr(imageDefinitionPublisher),
				SKU:       toPtr(imageDefinitionSKU),
			},
			OSState:      toPtr(armcomputev5.OperatingSystemStateTypesGeneralized),
			OSType:       toPtr(armcomputev5.OperatingSystemTypesLinux),
			Architecture: toPtr(armcomputev5.ArchitectureX64),
			Features: []*armcomputev5.GalleryImageFeature{
				{
					Name:  toPtr("SecurityType"),
					Value: &securityType,
				},
			},
			HyperVGeneration: toPtr(armcomputev5.HyperVGenerationV2),
		},
	}
	createPoller, err := u.image.BeginCreateOrUpdate(ctx, u.resourceGroup, sigName, definitionName, galleryImage,
		&armcomputev5.GalleryImagesClientBeginCreateOrUpdateOptions{},
	)
	if err != nil {
		return fmt.Errorf("creating image definition: %w", err)
	}
	if _, err = createPoller.PollUntilDone(ctx, &runtime.PollUntilDoneOptions{Frequency: u.pollingFrequency}); err != nil {
		return fmt.Errorf("waiting for image definition to be created: %w", err)
	}
	return nil
}

func (u *Uploader) createImageVersion(ctx context.Context, sigName, definitionName, versionName, imageID string) (string, error) {
	u.log.Debug("Creating image version %s/%s/%s in %s", sigName, definitionName, versionName, u.resourceGroup)
	imageVersion := armcomputev5.GalleryImageVersion{
		Location: &u.location,
		Properties: &armcomputev5.GalleryImageVersionProperties{
			StorageProfile: &armcomputev5.GalleryImageVersionStorageProfile{
				OSDiskImage: &armcomputev5.GalleryOSDiskImage{
					HostCaching: toPtr(armcomputev5.HostCachingReadOnly),
				},
				Source: &armcomputev5.GalleryArtifactVersionFullSource{
					ID: &imageID,
				},
			},
			PublishingProfile: &armcomputev5.GalleryImageVersionPublishingProfile{
				ReplicaCount:    toPtr[int32](1),
				ReplicationMode: toPtr(armcomputev5.ReplicationModeFull),
				TargetRegions:   targetRegions,
			},
		},
	}
	createPoller, err := u.imageVersions.BeginCreateOrUpdate(ctx, u.resourceGroup, sigName, definitionName, versionName, imageVersion,
		&armcomputev5.GalleryImageVersionsClientBeginCreateOrUpdateOptions{},
	)
	if err != nil {
		return "", fmt.Errorf("creating image version: %w", err)
	}
	createdImage, err := createPoller.PollUntilDone(ctx, &runtime.PollUntilDoneOptions{Frequency: u.pollingFrequency})
	if err != nil {
		return "", fmt.Errorf("waiting for image version to be created: %w", err)
	}
	if createdImage.ID == nil {
		return "", errors.New("created image has no id")
	}
	return *createdImage.ID, nil
}

func (u *Uploader) ensureImageVersionDeleted(ctx context.Context, sigName, definitionName, versionName string) error {
	_, err := u.imageVersions.Get(ctx, u.resourceGroup, sigName, definitionName, versionName, &armcomputev5.GalleryImageVersionsClientGetOptions{})
	if err != nil {
		u.log.Debug("Image version %s in %s/%s/%s doesn't exist. Nothing to clean up.", versionName, u.resourceGroup, sigName, definitionName)
		return nil
	}
	u.log.Debug("Deleting image version %s in %s/%s/%s", versionName, u.resourceGroup, sigName, definitionName)
	deletePoller, err := u.imageVersions.BeginDelete(ctx, u.resourceGroup, sigName, definitionName, versionName, &armcomputev5.GalleryImageVersionsClientBeginDeleteOptions{})
	if err != nil {
		return fmt.Errorf("deleting image version: %w", err)
	}
	if _, err = deletePoller.PollUntilDone(ctx, &runtime.PollUntilDoneOptions{Frequency: u.pollingFrequency}); err != nil {
		return fmt.Errorf("waiting for image version to be deleted: %w", err)
	}
	return nil
}

// getImageReference returns the image reference to use for the image version.
// If the shared image gallery is a community gallery, the community identifier is returned.
// Otherwise, the unshared identifier is returned.
func (u *Uploader) getImageReference(ctx context.Context, sigName, definitionName, versionName, unsharedID string) (string, error) {
	galleryResp, err := u.galleries.Get(ctx, u.resourceGroup, sigName, &armcomputev5.GalleriesClientGetOptions{})
	if err != nil {
		return "", fmt.Errorf("getting image gallery %s: %w", sigName, err)
	}
	if galleryResp.Properties == nil ||
		galleryResp.Properties.SharingProfile == nil ||
		galleryResp.Properties.SharingProfile.CommunityGalleryInfo == nil ||
		galleryResp.Properties.SharingProfile.CommunityGalleryInfo.CommunityGalleryEnabled == nil ||
		!*galleryResp.Properties.SharingProfile.CommunityGalleryInfo.CommunityGalleryEnabled {
		u.log.Warn("Image gallery %s in %s is not shared. Using private identifier", sigName, u.resourceGroup)
		return unsharedID, nil
	}
	if galleryResp.Properties == nil ||
		galleryResp.Properties.SharingProfile == nil ||
		galleryResp.Properties.SharingProfile.CommunityGalleryInfo == nil ||
		galleryResp.Properties.SharingProfile.CommunityGalleryInfo.PublicNames == nil ||
		len(galleryResp.Properties.SharingProfile.CommunityGalleryInfo.PublicNames) < 1 ||
		galleryResp.Properties.SharingProfile.CommunityGalleryInfo.PublicNames[0] == nil {
		return "", fmt.Errorf("image gallery %s in %s is a community gallery but has no public names", sigName, u.resourceGroup)
	}
	communityGalleryName := *galleryResp.Properties.SharingProfile.CommunityGalleryInfo.PublicNames[0]
	u.log.Debug(fmt.Sprintf("Image gallery %s in %s is shared. Using community identifier in %s", sigName, u.resourceGroup, communityGalleryName))
	communityVersionResp, err := u.communityVersions.Get(ctx, u.location, communityGalleryName,
		definitionName, versionName,
		&armcomputev5.CommunityGalleryImageVersionsClientGetOptions{},
	)
	if err != nil {
		return "", fmt.Errorf("getting community image version %s/%s/%s: %w", communityGalleryName, definitionName, versionName, err)
	}
	if communityVersionResp.Identifier == nil || communityVersionResp.Identifier.UniqueID == nil {
		return "", fmt.Errorf("community image version %s/%s/%s has no id", communityGalleryName, definitionName, versionName)
	}
	return *communityVersionResp.Identifier.UniqueID, nil
}

func uploadBlob(ctx context.Context, sasURL string, disk io.ReadSeeker, size int64, uploader sasBlobUploader) error {
	uploadClient, err := uploader(sasURL)
	if err != nil {
		return fmt.Errorf("uploading blob: %w", err)
	}
	var offset int64
	var chunksize int
	chunk := make([]byte, pageSizeMax)
	var readErr error
	for offset < size {
		chunksize, readErr = io.ReadAtLeast(disk, chunk, 1)
		if readErr != nil {
			return fmt.Errorf("reading from disk: %w", err)
		}
		if err := uploadChunk(ctx, uploadClient, bytes.NewReader(chunk[:chunksize]), offset, int64(chunksize)); err != nil {
			return fmt.Errorf("uploading chunk: %w", err)
		}
		offset += int64(chunksize)
	}
	return nil
}

func uploadChunk(ctx context.Context, uploader azurePageblobAPI, chunk io.ReadSeeker, offset, chunksize int64) error {
	_, err := uploader.UploadPages(ctx, &readSeekNopCloser{chunk}, blob.HTTPRange{
		Offset: offset,
		Count:  chunksize,
	}, nil)
	return err
}

func imageOffer(version versionsapi.Version) string {
	switch {
	case version.Stream() == "stable":
		return "constellation"
	case version.Stream() == "debug" && version.Ref() == "-":
		return version.Version()
	}
	return version.Ref() + "-" + version.Stream()
}

// imageVersion determines the semantic version string used inside a sig image.
// For releases, the actual semantic version of the image (without leading v) is used (major.minor.patch).
// Otherwise, the version is derived from the commit timestamp.
func imageVersion(version versionsapi.Version, timestamp time.Time) (string, error) {
	switch {
	case version.Stream() == "stable":
		fallthrough
	case version.Stream() == "debug" && version.Ref() == "-":
		return strings.TrimLeft(version.Version(), "v"), nil
	}

	formattedTime := timestamp.Format(timestampFormat)
	if len(formattedTime) != len(timestampFormat) {
		return "", errors.New("invalid timestamp")
	}
	// <year>.<month><day>.<time>
	return formattedTime[:4] + "." + formattedTime[4:8] + "." + formattedTime[8:], nil
}

type sasBlobUploader func(sasBlobURL string) (azurePageblobAPI, error)

type azureDiskAPI interface {
	Get(ctx context.Context, resourceGroupName string, diskName string,
		options *armcomputev5.DisksClientGetOptions,
	) (armcomputev5.DisksClientGetResponse, error)
	BeginCreateOrUpdate(ctx context.Context, resourceGroupName string, diskName string, disk armcomputev5.Disk,
		options *armcomputev5.DisksClientBeginCreateOrUpdateOptions,
	) (*runtime.Poller[armcomputev5.DisksClientCreateOrUpdateResponse], error)
	BeginDelete(ctx context.Context, resourceGroupName string, diskName string,
		options *armcomputev5.DisksClientBeginDeleteOptions,
	) (*runtime.Poller[armcomputev5.DisksClientDeleteResponse], error)
	BeginGrantAccess(ctx context.Context, resourceGroupName string, diskName string, grantAccessData armcomputev5.GrantAccessData,
		options *armcomputev5.DisksClientBeginGrantAccessOptions,
	) (*runtime.Poller[armcomputev5.DisksClientGrantAccessResponse], error)
	BeginRevokeAccess(ctx context.Context, resourceGroupName string, diskName string,
		options *armcomputev5.DisksClientBeginRevokeAccessOptions,
	) (*runtime.Poller[armcomputev5.DisksClientRevokeAccessResponse], error)
}

type azureManagedImageAPI interface {
	Get(ctx context.Context, resourceGroupName string, imageName string,
		options *armcomputev5.ImagesClientGetOptions,
	) (armcomputev5.ImagesClientGetResponse, error)
	BeginCreateOrUpdate(ctx context.Context, resourceGroupName string,
		imageName string, parameters armcomputev5.Image,
		options *armcomputev5.ImagesClientBeginCreateOrUpdateOptions,
	) (*runtime.Poller[armcomputev5.ImagesClientCreateOrUpdateResponse], error)
	BeginDelete(ctx context.Context, resourceGroupName string, imageName string,
		options *armcomputev5.ImagesClientBeginDeleteOptions,
	) (*runtime.Poller[armcomputev5.ImagesClientDeleteResponse], error)
}

type azurePageblobAPI interface {
	UploadPages(ctx context.Context, body io.ReadSeekCloser, contentRange blob.HTTPRange,
		options *pageblob.UploadPagesOptions,
	) (pageblob.UploadPagesResponse, error)
}

type azureGalleriesAPI interface {
	Get(ctx context.Context, resourceGroupName string, galleryName string,
		options *armcomputev5.GalleriesClientGetOptions,
	) (armcomputev5.GalleriesClientGetResponse, error)
	NewListPager(options *armcomputev5.GalleriesClientListOptions,
	) *runtime.Pager[armcomputev5.GalleriesClientListResponse]
	BeginCreateOrUpdate(ctx context.Context, resourceGroupName string,
		galleryName string, gallery armcomputev5.Gallery,
		options *armcomputev5.GalleriesClientBeginCreateOrUpdateOptions,
	) (*runtime.Poller[armcomputev5.GalleriesClientCreateOrUpdateResponse], error)
}

type azureGalleriesImageAPI interface {
	Get(ctx context.Context, resourceGroupName string, galleryName string,
		galleryImageName string, options *armcomputev5.GalleryImagesClientGetOptions,
	) (armcomputev5.GalleryImagesClientGetResponse, error)
	BeginCreateOrUpdate(ctx context.Context, resourceGroupName string, galleryName string,
		galleryImageName string, galleryImage armcomputev5.GalleryImage,
		options *armcomputev5.GalleryImagesClientBeginCreateOrUpdateOptions,
	) (*runtime.Poller[armcomputev5.GalleryImagesClientCreateOrUpdateResponse], error)
	BeginDelete(ctx context.Context, resourceGroupName string, galleryName string, galleryImageName string,
		options *armcomputev5.GalleryImagesClientBeginDeleteOptions,
	) (*runtime.Poller[armcomputev5.GalleryImagesClientDeleteResponse], error)
}

type azureGalleriesImageVersionAPI interface {
	Get(ctx context.Context, resourceGroupName string, galleryName string, galleryImageName string, galleryImageVersionName string,
		options *armcomputev5.GalleryImageVersionsClientGetOptions,
	) (armcomputev5.GalleryImageVersionsClientGetResponse, error)
	NewListByGalleryImagePager(resourceGroupName string, galleryName string, galleryImageName string,
		options *armcomputev5.GalleryImageVersionsClientListByGalleryImageOptions,
	) *runtime.Pager[armcomputev5.GalleryImageVersionsClientListByGalleryImageResponse]
	BeginCreateOrUpdate(ctx context.Context, resourceGroupName string, galleryName string, galleryImageName string,
		galleryImageVersionName string, galleryImageVersion armcomputev5.GalleryImageVersion,
		options *armcomputev5.GalleryImageVersionsClientBeginCreateOrUpdateOptions,
	) (*runtime.Poller[armcomputev5.GalleryImageVersionsClientCreateOrUpdateResponse], error)
	BeginDelete(ctx context.Context, resourceGroupName string, galleryName string, galleryImageName string,
		galleryImageVersionName string, options *armcomputev5.GalleryImageVersionsClientBeginDeleteOptions,
	) (*runtime.Poller[armcomputev5.GalleryImageVersionsClientDeleteResponse], error)
}

type azureCommunityGalleryImageVersionAPI interface {
	Get(ctx context.Context, location string,
		publicGalleryName, galleryImageName, galleryImageVersionName string,
		options *armcomputev5.CommunityGalleryImageVersionsClientGetOptions,
	) (armcomputev5.CommunityGalleryImageVersionsClientGetResponse, error)
}

const (
	pollingFrequency = 10 * time.Second
	// uploadAccessDuration is the time in seconds that
	// sas tokens should be valid for (24 hours).
	uploadAccessDuration     = 86400   // 24 hours
	pageSizeMax              = 4194304 // 4MiB
	pageSizeMin              = 512     // 512 bytes
	sigNameStable            = "Constellation_CVM"
	sigNameDebug             = "Constellation_Debug_CVM"
	sigNameDefault           = "Constellation_Testing_CVM"
	imageDefinitionPublisher = "edgelesssys"
	imageDefinitionSKU       = "constellation"
	timestampFormat          = "20060102150405"
)

var targetRegions = []*armcomputev5.TargetRegion{
	{
		Name:                 toPtr("northeurope"),
		RegionalReplicaCount: toPtr[int32](1),
	},
	{
		Name:                 toPtr("eastus"),
		RegionalReplicaCount: toPtr[int32](1),
	},
	{
		Name:                 toPtr("westeurope"),
		RegionalReplicaCount: toPtr[int32](1),
	},
	{
		Name:                 toPtr("westus"),
		RegionalReplicaCount: toPtr[int32](1),
	},
	{
		Name:                 toPtr("southeastasia"),
		RegionalReplicaCount: toPtr[int32](1),
	},
}

//go:generate stringer -type=DiskType -trimprefix=DiskType

// DiskType is the kind of disk created using the Azure API.
type DiskType uint32

// FromString converts a string into an DiskType.
func FromString(s string) DiskType {
	switch strings.ToLower(s) {
	case strings.ToLower(DiskTypeNormal.String()):
		return DiskTypeNormal
	case strings.ToLower(DiskTypeWithVMGS.String()):
		return DiskTypeWithVMGS
	default:
		return DiskTypeUnknown
	}
}

const (
	// DiskTypeUnknown is default value for DiskType.
	DiskTypeUnknown DiskType = iota
	// DiskTypeNormal creates a normal Azure disk (single block device).
	DiskTypeNormal
	// DiskTypeWithVMGS creates a disk with VMGS (also called secure disk)
	// that has an additional block device for the VMGS disk.
	DiskTypeWithVMGS
)

func toPtr[T any](v T) *T {
	return &v
}

type readSeekNopCloser struct {
	io.ReadSeeker
}

func (n *readSeekNopCloser) Close() error {
	return nil
}
