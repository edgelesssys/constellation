/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// package aws implements uploading os images to aws.
package aws

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	s3manager "github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"

	"github.com/edgelesssys/constellation/v2/internal/api/versionsapi"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/osimage"
	"github.com/edgelesssys/constellation/v2/internal/osimage/secureboot"
)

// Uploader can upload and remove os images on GCP.
type Uploader struct {
	region     string
	bucketName string
	ec2        func(ctx context.Context, region string) (ec2API, error)
	s3         func(ctx context.Context, region string) (s3API, error)
	s3uploader func(ctx context.Context, region string) (s3UploaderAPI, error)

	log *logger.Logger
}

// New creates a new Uploader.
func New(region, bucketName string, log *logger.Logger) (*Uploader, error) {
	return &Uploader{
		region:     region,
		bucketName: bucketName,
		ec2: func(ctx context.Context, region string) (ec2API, error) {
			cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
			if err != nil {
				return nil, err
			}
			return ec2.NewFromConfig(cfg), nil
		},
		s3: func(ctx context.Context, region string) (s3API, error) {
			cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
			if err != nil {
				return nil, err
			}
			return s3.NewFromConfig(cfg), nil
		},
		s3uploader: func(ctx context.Context, region string) (s3UploaderAPI, error) {
			cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
			if err != nil {
				return nil, err
			}
			return s3manager.NewUploader(s3.NewFromConfig(cfg)), nil
		},

		log: log,
	}, nil
}

// Upload uploads an OS image to AWS.
func (u *Uploader) Upload(ctx context.Context, req *osimage.UploadRequest) ([]versionsapi.ImageInfoEntry, error) {
	blobName := fmt.Sprintf("image-%s-%s-%d.raw", req.Version.Stream(), req.Version.Version(), req.Timestamp.Unix())
	imageName := imageName(req.Version, req.AttestationVariant, req.Timestamp)
	allRegions := []string{u.region}
	allRegions = append(allRegions, replicationRegions...)
	// TODO(malt3): make this configurable
	publish := true
	amiIDs := make(map[string]string, len(allRegions))
	if err := u.ensureBucket(ctx); err != nil {
		return nil, fmt.Errorf("ensuring bucket %s exists: %w", u.bucketName, err)
	}

	// pre-cleaning
	for _, region := range allRegions {
		if err := u.ensureImageDeleted(ctx, imageName, region); err != nil {
			return nil, fmt.Errorf("pre-cleaning: ensuring no image under the name %s in region %s: %w", imageName, region, err)
		}
	}
	if err := u.ensureSnapshotDeleted(ctx, imageName, u.region); err != nil {
		return nil, fmt.Errorf("pre-cleaning: ensuring no snapshot using the same name exists: %w", err)
	}
	if err := u.ensureBlobDeleted(ctx, blobName); err != nil {
		return nil, fmt.Errorf("pre-cleaning: ensuring no blob using the same name exists: %w", err)
	}

	// create primary image
	if err := u.uploadBlob(ctx, blobName, req.Image); err != nil {
		return nil, fmt.Errorf("uploading image to s3: %w", err)
	}
	defer func() {
		if err := u.ensureBlobDeleted(ctx, blobName); err != nil {
			u.log.Errorf("post-cleaning: deleting temporary blob from s3", err)
		}
	}()
	snapshotID, err := u.importSnapshot(ctx, blobName, imageName)
	if err != nil {
		return nil, fmt.Errorf("importing snapshot: %w", err)
	}
	primaryAMIID, err := u.createImageFromSnapshot(ctx, req.Version, imageName, snapshotID, req.SecureBoot, req.UEFIVarStore)
	if err != nil {
		return nil, fmt.Errorf("creating image from snapshot: %w", err)
	}
	amiIDs[u.region] = primaryAMIID
	if err := u.waitForImage(ctx, primaryAMIID, u.region); err != nil {
		return nil, fmt.Errorf("waiting for primary image to become available: %w", err)
	}

	// replicate image
	for _, region := range replicationRegions {
		amiID, err := u.replicateImage(ctx, imageName, primaryAMIID, region)
		if err != nil {
			return nil, fmt.Errorf("replicating image to region %s: %w", region, err)
		}
		amiIDs[region] = amiID
	}

	// wait for replication, tag, publish
	var imageInfo []versionsapi.ImageInfoEntry
	for _, region := range allRegions {
		if err := u.waitForImage(ctx, amiIDs[region], region); err != nil {
			return nil, fmt.Errorf("waiting for image to become available in region %s: %w", region, err)
		}
		if err := u.tagImageAndSnapshot(ctx, imageName, amiIDs[region], region); err != nil {
			return nil, fmt.Errorf("tagging image in region %s: %w", region, err)
		}
		if !publish {
			continue
		}
		if err := u.publishImage(ctx, amiIDs[region], region); err != nil {
			return nil, fmt.Errorf("publishing image in region %s: %w", region, err)
		}
		imageInfo = append(imageInfo, versionsapi.ImageInfoEntry{
			CSP:                "aws",
			AttestationVariant: req.AttestationVariant,
			Reference:          amiIDs[region],
			Region:             region,
		})
	}

	return imageInfo, nil
}

func (u *Uploader) ensureBucket(ctx context.Context) error {
	s3C, err := u.s3(ctx, u.region)
	if err != nil {
		return fmt.Errorf("determining if bucket %s exists: %w", u.bucketName, err)
	}
	_, err = s3C.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: &u.bucketName,
	})
	if err == nil {
		u.log.Debugf("Bucket %s exists", u.bucketName)
		return nil
	}
	var noSuchBucketErr *types.NoSuchBucket
	if !errors.As(err, &noSuchBucketErr) {
		return fmt.Errorf("determining if bucket %s exists: %w", u.bucketName, err)
	}
	u.log.Debugf("Creating bucket %s", u.bucketName)
	_, err = s3C.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: &u.bucketName,
	})
	if err != nil {
		return fmt.Errorf("creating bucket %s: %w", u.bucketName, err)
	}
	return nil
}

func (u *Uploader) uploadBlob(ctx context.Context, blobName string, img io.Reader) error {
	u.log.Debugf("Uploading os image as %s", blobName)
	uploadC, err := u.s3uploader(ctx, u.region)
	if err != nil {
		return err
	}
	_, err = uploadC.Upload(ctx, &s3.PutObjectInput{
		Bucket:            &u.bucketName,
		Key:               &blobName,
		Body:              img,
		ChecksumAlgorithm: s3types.ChecksumAlgorithmSha256,
	})
	return err
}

func (u *Uploader) ensureBlobDeleted(ctx context.Context, blobName string) error {
	s3C, err := u.s3(ctx, u.region)
	if err != nil {
		return err
	}
	_, err = s3C.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: &u.bucketName,
		Key:    &blobName,
	})
	var apiError smithy.APIError
	if errors.As(err, &apiError) && apiError.ErrorCode() == "NotFound" {
		u.log.Debugf("Blob %s in %s doesn't exist. Nothing to clean up.", blobName, u.bucketName)
		return nil
	}
	if err != nil {
		return err
	}
	u.log.Debugf("Deleting blob %s", blobName)
	_, err = s3C.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &u.bucketName,
		Key:    &blobName,
	})
	return err
}

func (u *Uploader) findSnapshots(ctx context.Context, snapshotName, region string) ([]string, error) {
	ec2C, err := u.ec2(ctx, region)
	if err != nil {
		return nil, fmt.Errorf("creating ec2 client: %w", err)
	}
	snapshots, err := ec2C.DescribeSnapshots(ctx, &ec2.DescribeSnapshotsInput{
		Filters: []ec2types.Filter{
			{
				Name:   toPtr("tag:Name"),
				Values: []string{snapshotName},
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("describing snapshots: %w", err)
	}
	var snapshotIDs []string
	for _, s := range snapshots.Snapshots {
		if s.SnapshotId == nil {
			continue
		}
		snapshotIDs = append(snapshotIDs, *s.SnapshotId)
	}
	return snapshotIDs, nil
}

func (u *Uploader) importSnapshot(ctx context.Context, blobName, snapshotName string) (string, error) {
	u.log.Debugf("Importing %s as snapshot %s", blobName, snapshotName)
	ec2C, err := u.ec2(ctx, u.region)
	if err != nil {
		return "", fmt.Errorf("creating ec2 client: %w", err)
	}
	importResp, err := ec2C.ImportSnapshot(ctx, &ec2.ImportSnapshotInput{
		ClientData: &ec2types.ClientData{
			Comment: &snapshotName,
		},
		Description: &snapshotName,
		DiskContainer: &ec2types.SnapshotDiskContainer{
			Description: &snapshotName,
			Format:      toPtr(string(ec2types.DiskImageFormatRaw)),
			UserBucket: &ec2types.UserBucket{
				S3Bucket: &u.bucketName,
				S3Key:    &blobName,
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("importing snapshot: %w", err)
	}
	if importResp.ImportTaskId == nil {
		return "", fmt.Errorf("importing snapshot: no import task ID returned")
	}
	u.log.Debugf("Waiting for snapshot %s to be ready", snapshotName)
	return waitForSnapshotImport(ctx, ec2C, *importResp.ImportTaskId)
}

func (u *Uploader) ensureSnapshotDeleted(ctx context.Context, snapshotName, region string) error {
	ec2C, err := u.ec2(ctx, region)
	if err != nil {
		return fmt.Errorf("creating ec2 client: %w", err)
	}
	snapshots, err := u.findSnapshots(ctx, snapshotName, region)
	if err != nil {
		return fmt.Errorf("finding snapshots: %w", err)
	}
	for _, snapshot := range snapshots {
		u.log.Debugf("Deleting snapshot %s in %s", snapshot, region)
		_, err = ec2C.DeleteSnapshot(ctx, &ec2.DeleteSnapshotInput{
			SnapshotId: toPtr(snapshot),
		})
		if err != nil {
			return fmt.Errorf("deleting snapshot %s: %w", snapshot, err)
		}
	}
	return nil
}

func (u *Uploader) createImageFromSnapshot(ctx context.Context, version versionsapi.Version, imageName, snapshotID string, enableSecureBoot bool, uefiVarStore secureboot.UEFIVarStore) (string, error) {
	u.log.Debugf("Creating image %s in %s", imageName, u.region)
	ec2C, err := u.ec2(ctx, u.region)
	if err != nil {
		return "", fmt.Errorf("creating ec2 client: %w", err)
	}
	var uefiData *string
	if enableSecureBoot {
		awsUEFIData, err := uefiVarStore.ToAWS()
		if err != nil {
			return "", fmt.Errorf("creating uefi data: %w", err)
		}
		uefiData = toPtr(awsUEFIData)
	}

	createReq, err := ec2C.RegisterImage(ctx, &ec2.RegisterImageInput{
		Name:         &imageName,
		Architecture: ec2types.ArchitectureValuesX8664,
		BlockDeviceMappings: []ec2types.BlockDeviceMapping{
			{
				DeviceName: toPtr("/dev/xvda"),
				Ebs: &ec2types.EbsBlockDevice{
					DeleteOnTermination: toPtr(true),
					SnapshotId:          &snapshotID,
				},
			},
		},
		BootMode:           ec2types.BootModeValuesUefi,
		Description:        toPtr("Constellation " + version.ShortPath()),
		EnaSupport:         toPtr(true),
		RootDeviceName:     toPtr("/dev/xvda"),
		TpmSupport:         ec2types.TpmSupportValuesV20,
		UefiData:           uefiData,
		VirtualizationType: toPtr("hvm"),
	})
	if err != nil {
		return "", fmt.Errorf("creating image: %w", err)
	}
	if createReq.ImageId == nil {
		return "", fmt.Errorf("creating image: no image ID returned")
	}
	return *createReq.ImageId, nil
}

func (u *Uploader) replicateImage(ctx context.Context, imageName, amiID string, region string) (string, error) {
	u.log.Debugf("Replicating image %s to %s", imageName, region)
	ec2C, err := u.ec2(ctx, region)
	if err != nil {
		return "", fmt.Errorf("creating ec2 client: %w", err)
	}
	replicateReq, err := ec2C.CopyImage(ctx, &ec2.CopyImageInput{
		Name:          &imageName,
		SourceImageId: &amiID,
		SourceRegion:  &u.region,
	})
	if err != nil {
		return "", fmt.Errorf("replicating image: %w", err)
	}
	if replicateReq.ImageId == nil {
		return "", fmt.Errorf("replicating image: no image ID returned")
	}
	return *replicateReq.ImageId, nil
}

func (u *Uploader) findImage(ctx context.Context, imageName, region string) (string, error) {
	ec2C, err := u.ec2(ctx, region)
	if err != nil {
		return "", fmt.Errorf("creating ec2 client: %w", err)
	}
	snapshots, err := ec2C.DescribeImages(ctx, &ec2.DescribeImagesInput{
		Filters: []ec2types.Filter{
			{
				Name:   toPtr("name"),
				Values: []string{imageName},
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("describing images: %w", err)
	}
	if len(snapshots.Images) == 0 {
		return "", errAMIDoesNotExist
	}
	if len(snapshots.Images) != 1 {
		return "", fmt.Errorf("expected 1 image, got %d", len(snapshots.Images))
	}
	if snapshots.Images[0].ImageId == nil {
		return "", fmt.Errorf("image ID is nil")
	}
	return *snapshots.Images[0].ImageId, nil
}

func (u *Uploader) waitForImage(ctx context.Context, amiID, region string) error {
	u.log.Debugf("Waiting for image %s in %s to be created", amiID, region)
	ec2C, err := u.ec2(ctx, region)
	if err != nil {
		return fmt.Errorf("creating ec2 client: %w", err)
	}
	waiter := ec2.NewImageAvailableWaiter(ec2C)
	err = waiter.Wait(ctx, &ec2.DescribeImagesInput{
		ImageIds: []string{amiID},
	}, maxWait)
	if err != nil {
		return fmt.Errorf("waiting for image: %w", err)
	}
	return nil
}

func (u *Uploader) tagImageAndSnapshot(ctx context.Context, imageName, amiID, region string) error {
	u.log.Debugf("Tagging backing snapshot of image %s in %s", amiID, region)
	ec2C, err := u.ec2(ctx, region)
	if err != nil {
		return fmt.Errorf("creating ec2 client: %w", err)
	}
	snapshotID, err := getBackingSnapshotID(ctx, ec2C, amiID)
	if err != nil {
		return fmt.Errorf("getting backing snapshot ID: %w", err)
	}
	_, err = ec2C.CreateTags(ctx, &ec2.CreateTagsInput{
		Resources: []string{amiID, snapshotID},
		Tags: []ec2types.Tag{
			{
				Key:   toPtr("Name"),
				Value: toPtr(imageName),
			},
		},
	})
	if err != nil {
		return fmt.Errorf("tagging ami and snapshot: %w", err)
	}
	return nil
}

func (u *Uploader) publishImage(ctx context.Context, imageName, region string) error {
	u.log.Debugf("Publishing image %s in %s", imageName, region)
	ec2C, err := u.ec2(ctx, region)
	if err != nil {
		return fmt.Errorf("creating ec2 client: %w", err)
	}
	_, err = ec2C.ModifyImageAttribute(ctx, &ec2.ModifyImageAttributeInput{
		ImageId: &imageName,
		LaunchPermission: &ec2types.LaunchPermissionModifications{
			Add: []ec2types.LaunchPermission{
				{
					Group: ec2types.PermissionGroupAll,
				},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("publishing image: %w", err)
	}
	return nil
}

func (u *Uploader) ensureImageDeleted(ctx context.Context, imageName, region string) error {
	ec2C, err := u.ec2(ctx, region)
	if err != nil {
		return fmt.Errorf("creating ec2 client: %w", err)
	}
	amiID, err := u.findImage(ctx, imageName, region)
	if err == errAMIDoesNotExist {
		u.log.Debugf("Image %s in %s doesn't exist. Nothing to clean up.", imageName, region)
		return nil
	}
	snapshotID, err := getBackingSnapshotID(ctx, ec2C, amiID)
	if err == errAMIDoesNotExist {
		u.log.Debugf("Image %s doesn't exist. Nothing to clean up.", amiID)
		return nil
	}
	u.log.Debugf("Deleting image %s in %s with backing snapshot", amiID, region)
	_, err = ec2C.DeregisterImage(ctx, &ec2.DeregisterImageInput{
		ImageId: &amiID,
	})
	if err != nil {
		return fmt.Errorf("deleting image: %w", err)
	}
	_, err = ec2C.DeleteSnapshot(ctx, &ec2.DeleteSnapshotInput{
		SnapshotId: &snapshotID,
	})
	if err != nil {
		return fmt.Errorf("deleting snapshot: %w", err)
	}
	return nil
}

func imageName(version versionsapi.Version, attestationVariant string, timestamp time.Time) string {
	if version.Stream() == "stable" {
		return fmt.Sprintf("constellation-%s-%s", version.Version(), attestationVariant)
	}
	return fmt.Sprintf("constellation-%s-%s-%s-%s", version.Stream(), version.Version(), attestationVariant, timestamp.Format(timestampFormat))
}

func waitForSnapshotImport(ctx context.Context, ec2C ec2API, importTaskID string) (string, error) {
	for {
		taskResp, err := ec2C.DescribeImportSnapshotTasks(ctx, &ec2.DescribeImportSnapshotTasksInput{
			ImportTaskIds: []string{importTaskID},
		})
		if err != nil {
			return "", fmt.Errorf("describing import snapshot task: %w", err)
		}
		if len(taskResp.ImportSnapshotTasks) == 0 {
			return "", fmt.Errorf("describing import snapshot task: no tasks returned")
		}
		if taskResp.ImportSnapshotTasks[0].SnapshotTaskDetail == nil {
			return "", fmt.Errorf("describing import snapshot task: no snapshot task detail returned")
		}
		if taskResp.ImportSnapshotTasks[0].SnapshotTaskDetail.Status == nil {
			return "", fmt.Errorf("describing import snapshot task: no status returned")
		}
		switch *taskResp.ImportSnapshotTasks[0].SnapshotTaskDetail.Status {
		case string(ec2types.SnapshotStateCompleted):
			return *taskResp.ImportSnapshotTasks[0].SnapshotTaskDetail.SnapshotId, nil
		case string(ec2types.SnapshotStateError):
			return "", fmt.Errorf("importing snapshot: task failed")
		}
		time.Sleep(waitInterval)
	}
}

func getBackingSnapshotID(ctx context.Context, ec2C ec2API, amiID string) (string, error) {
	describeResp, err := ec2C.DescribeImages(ctx, &ec2.DescribeImagesInput{
		ImageIds: []string{amiID},
	})
	if err != nil || len(describeResp.Images) == 0 {
		return "", errAMIDoesNotExist
	}
	if len(describeResp.Images) != 1 {
		return "", fmt.Errorf("describing image: expected 1 image, got %d", len(describeResp.Images))
	}
	image := describeResp.Images[0]
	if len(image.BlockDeviceMappings) != 1 {
		return "", fmt.Errorf("found %d block device mappings for image %s, expected 1", len(image.BlockDeviceMappings), amiID)
	}
	if image.BlockDeviceMappings[0].Ebs == nil {
		return "", fmt.Errorf("image %s does not have an EBS block device mapping", amiID)
	}
	ebs := image.BlockDeviceMappings[0].Ebs
	if ebs.SnapshotId == nil {
		return "", fmt.Errorf("image %s does not have an EBS snapshot", amiID)
	}
	return *ebs.SnapshotId, nil
}

type ec2API interface {
	DescribeImages(ctx context.Context, params *ec2.DescribeImagesInput,
		optFns ...func(*ec2.Options),
	) (*ec2.DescribeImagesOutput, error)
	ModifyImageAttribute(ctx context.Context, params *ec2.ModifyImageAttributeInput,
		optFns ...func(*ec2.Options),
	) (*ec2.ModifyImageAttributeOutput, error)
	RegisterImage(ctx context.Context, params *ec2.RegisterImageInput,
		optFns ...func(*ec2.Options),
	) (*ec2.RegisterImageOutput, error)
	CopyImage(ctx context.Context, params *ec2.CopyImageInput, optFns ...func(*ec2.Options),
	) (*ec2.CopyImageOutput, error)
	DeregisterImage(ctx context.Context, params *ec2.DeregisterImageInput,
		optFns ...func(*ec2.Options),
	) (*ec2.DeregisterImageOutput, error)
	ImportSnapshot(ctx context.Context, params *ec2.ImportSnapshotInput,
		optFns ...func(*ec2.Options),
	) (*ec2.ImportSnapshotOutput, error)
	DescribeImportSnapshotTasks(ctx context.Context, params *ec2.DescribeImportSnapshotTasksInput,
		optFns ...func(*ec2.Options),
	) (*ec2.DescribeImportSnapshotTasksOutput, error)
	DescribeSnapshots(ctx context.Context, params *ec2.DescribeSnapshotsInput,
		optFns ...func(*ec2.Options),
	) (*ec2.DescribeSnapshotsOutput, error)
	DeleteSnapshot(ctx context.Context, params *ec2.DeleteSnapshotInput, optFns ...func(*ec2.Options),
	) (*ec2.DeleteSnapshotOutput, error)
	CreateTags(ctx context.Context, params *ec2.CreateTagsInput, optFns ...func(*ec2.Options),
	) (*ec2.CreateTagsOutput, error)
}

type s3API interface {
	HeadBucket(ctx context.Context, params *s3.HeadBucketInput, optFns ...func(*s3.Options),
	) (*s3.HeadBucketOutput, error)
	CreateBucket(ctx context.Context, params *s3.CreateBucketInput, optFns ...func(*s3.Options),
	) (*s3.CreateBucketOutput, error)
	HeadObject(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options),
	) (*s3.HeadObjectOutput, error)
	DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options),
	) (*s3.DeleteObjectOutput, error)
}

type s3UploaderAPI interface {
	Upload(ctx context.Context, input *s3.PutObjectInput, opts ...func(*s3manager.Uploader),
	) (*s3manager.UploadOutput, error)
}

func toPtr[T any](v T) *T {
	return &v
}

const (
	waitInterval    = 15 * time.Second
	maxWait         = 30 * time.Minute
	timestampFormat = "20060102150405"
)

var (
	errAMIDoesNotExist = errors.New("ami does not exist")
	replicationRegions = []string{"eu-west-1", "eu-west-3", "us-east-2", "ap-south-1"}
)
