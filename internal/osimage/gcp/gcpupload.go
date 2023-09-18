/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// package gcp implements uploading os images to gcp.
package gcp

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/url"
	"path"
	"strings"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"cloud.google.com/go/storage"
	"github.com/edgelesssys/constellation/v2/internal/api/versionsapi"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/osimage"
	"github.com/edgelesssys/constellation/v2/internal/osimage/secureboot"
	gaxv2 "github.com/googleapis/gax-go/v2"
)

// Uploader can upload and remove os images on GCP.
type Uploader struct {
	project    string
	location   string
	bucketName string
	image      imagesAPI
	bucket     bucketAPI

	log *logger.Logger
}

// New creates a new Uploader.
func New(ctx context.Context, project, location, bucketName string, log *logger.Logger) (*Uploader, error) {
	image, err := compute.NewImagesRESTClient(ctx)
	if err != nil {
		return nil, err
	}
	storage, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	bucket := storage.Bucket(bucketName)

	return &Uploader{
		project:    project,
		location:   location,
		bucketName: bucketName,
		image:      image,
		bucket:     bucket,

		log: log,
	}, nil
}

// Upload uploads an OS image to GCP.
func (u *Uploader) Upload(ctx context.Context, req *osimage.UploadRequest) ([]versionsapi.ImageInfoEntry, error) {
	imageName := u.imageName(req.Version, req.AttestationVariant)
	blobName := imageName + ".tar.gz"
	if err := u.ensureBucket(ctx); err != nil {
		return nil, fmt.Errorf("setup: ensuring bucket exists: %w", err)
	}
	if err := u.ensureImageDeleted(ctx, imageName); err != nil {
		return nil, fmt.Errorf("pre-cleaning: ensuring no image using the same name exists: %w", err)
	}
	if err := u.ensureBlobDeleted(ctx, blobName); err != nil {
		return nil, fmt.Errorf("pre-cleaning: ensuring no blob using the same name exists: %w", err)
	}
	if err := u.uploadBlob(ctx, blobName, req.Image); err != nil {
		return nil, fmt.Errorf("uploading blob: %w", err)
	}
	defer func() {
		// cleanup temporary blob
		if err := u.ensureBlobDeleted(ctx, blobName); err != nil {
			u.log.Errorf("post-cleaning: deleting blob: %v", err)
		}
	}()
	imageRef, err := u.createImage(ctx, req.Version, imageName, blobName, req.SecureBoot, req.SBDatabase)
	if err != nil {
		return nil, fmt.Errorf("creating image: %w", err)
	}
	return []versionsapi.ImageInfoEntry{
		{
			CSP:                "gcp",
			AttestationVariant: req.AttestationVariant,
			Reference:          imageRef,
		},
	}, nil
}

func (u *Uploader) ensureBucket(ctx context.Context) error {
	_, err := u.bucket.Attrs(ctx)
	if err == nil {
		u.log.Debugf("Bucket %s exists", u.bucketName)
		return nil
	}
	if err != storage.ErrBucketNotExist {
		return err
	}
	u.log.Debugf("Creating bucket %s", u.bucketName)
	return u.bucket.Create(ctx, u.project, &storage.BucketAttrs{
		PublicAccessPrevention: storage.PublicAccessPreventionEnforced,
		Location:               u.location,
	})
}

func (u *Uploader) uploadBlob(ctx context.Context, blobName string, img io.Reader) error {
	u.log.Debugf("Uploading os image as %s", blobName)
	writer := u.bucket.Object(blobName).NewWriter(ctx)
	_, err := io.Copy(writer, img)
	if err != nil {
		return err
	}
	return writer.Close()
}

func (u *Uploader) ensureBlobDeleted(ctx context.Context, blobName string) error {
	_, err := u.bucket.Object(blobName).Attrs(ctx)
	if err == storage.ErrObjectNotExist {
		u.log.Debugf("Blob %s in %s doesn't exist. Nothing to clean up.", blobName, u.bucketName)
		return nil
	}
	if err != nil {
		return err
	}
	u.log.Debugf("Deleting blob %s", blobName)
	return u.bucket.Object(blobName).Delete(ctx)
}

func (u *Uploader) createImage(ctx context.Context, version versionsapi.Version, imageName, blobName string, enableSecureBoot bool, sbDatabase secureboot.Database) (string, error) {
	u.log.Debugf("Creating image %s", imageName)
	blobURL := u.blobURL(blobName)
	family := u.imageFamily(version)
	var initialState *computepb.InitialStateConfig
	if enableSecureBoot {
		initialState = &computepb.InitialStateConfig{
			Pk:   pk(&sbDatabase),
			Keks: keks(&sbDatabase),
			Dbs:  dbs(&sbDatabase),
		}
	}
	req := computepb.InsertImageRequest{
		ImageResource: &computepb.Image{
			Name: &imageName,
			RawDisk: &computepb.RawDisk{
				ContainerType: toPtr("TAR"),
				Source:        &blobURL,
			},
			Family:       &family,
			Architecture: toPtr("X86_64"),
			GuestOsFeatures: []*computepb.GuestOsFeature{
				{Type: toPtr("GVNIC")},
				{Type: toPtr("SEV_CAPABLE")},
				{Type: toPtr("SEV_SNP_CAPABLE")},
				{Type: toPtr("VIRTIO_SCSI_MULTIQUEUE")},
				{Type: toPtr("UEFI_COMPATIBLE")},
			},
			ShieldedInstanceInitialState: initialState,
		},
		Project: u.project,
	}
	op, err := u.image.Insert(ctx, &req)
	if err != nil {
		return "", fmt.Errorf("creating image: %w", err)
	}
	if err := op.Wait(ctx); err != nil {
		return "", fmt.Errorf("waiting for image to be created: %w", err)
	}
	policy := &computepb.Policy{
		Bindings: []*computepb.Binding{
			{
				Role:    toPtr("roles/compute.imageUser"),
				Members: []string{"allAuthenticatedUsers"},
			},
		},
	}
	if _, err = u.image.SetIamPolicy(ctx, &computepb.SetIamPolicyImageRequest{
		Resource: imageName,
		Project:  u.project,
		GlobalSetPolicyRequestResource: &computepb.GlobalSetPolicyRequest{
			Policy: policy,
		},
	}); err != nil {
		return "", fmt.Errorf("setting iam policy: %w", err)
	}
	image, err := u.image.Get(ctx, &computepb.GetImageRequest{
		Image:   imageName,
		Project: u.project,
	})
	if err != nil {
		return "", fmt.Errorf("created image doesn't exist: %w", err)
	}
	return strings.TrimPrefix(image.GetSelfLink(), "https://www.googleapis.com/compute/v1/"), nil
}

func (u *Uploader) ensureImageDeleted(ctx context.Context, imageName string) error {
	_, err := u.image.Get(ctx, &computepb.GetImageRequest{
		Image:   imageName,
		Project: u.project,
	})
	if err != nil {
		u.log.Debugf("Image %s doesn't exist. Nothing to clean up.", imageName)
		return nil
	}
	u.log.Debugf("Deleting image %s", imageName)
	op, err := u.image.Delete(ctx, &computepb.DeleteImageRequest{
		Image:   imageName,
		Project: u.project,
	})
	if err != nil {
		return err
	}
	return op.Wait(ctx)
}

func (u *Uploader) blobURL(blobName string) string {
	return (&url.URL{
		Scheme: "https",
		Host:   "storage.googleapis.com",
		Path:   path.Join(u.bucketName, blobName),
	}).String()
}

func (u *Uploader) imageName(version versionsapi.Version, attestationVariant string) string {
	return strings.ReplaceAll(version.Version(), ".", "-") + "-" + attestationVariant + "-" + version.Stream()
}

func (u *Uploader) imageFamily(version versionsapi.Version) string {
	if version.Stream() == "stable" {
		return "constellation"
	}
	truncatedRef := version.Ref()
	if len(version.Ref()) > 45 {
		truncatedRef = version.Ref()[:45]
	}
	return "constellation-" + truncatedRef
}

func pk(sbDatabase *secureboot.Database) *computepb.FileContentBuffer {
	encoded := base64.StdEncoding.EncodeToString(sbDatabase.PK)
	return &computepb.FileContentBuffer{
		Content:  toPtr(encoded),
		FileType: toPtr("X509"),
	}
}

func keks(sbDatabase *secureboot.Database) []*computepb.FileContentBuffer {
	keks := make([]*computepb.FileContentBuffer, 0, len(sbDatabase.Keks))
	for _, kek := range sbDatabase.Keks {
		encoded := base64.StdEncoding.EncodeToString(kek)
		keks = append(keks, &computepb.FileContentBuffer{
			Content:  toPtr(encoded),
			FileType: toPtr("X509"),
		})
	}
	return keks
}

func dbs(sbDatabase *secureboot.Database) []*computepb.FileContentBuffer {
	dbs := make([]*computepb.FileContentBuffer, 0, len(sbDatabase.DBs))
	for _, db := range sbDatabase.DBs {
		encoded := base64.StdEncoding.EncodeToString(db)
		dbs = append(dbs, &computepb.FileContentBuffer{
			Content:  toPtr(encoded),
			FileType: toPtr("X509"),
		})
	}
	return dbs
}

type imagesAPI interface {
	Get(ctx context.Context, req *computepb.GetImageRequest, opts ...gaxv2.CallOption,
	) (*computepb.Image, error)
	Insert(ctx context.Context, req *computepb.InsertImageRequest, opts ...gaxv2.CallOption,
	) (*compute.Operation, error)
	SetIamPolicy(ctx context.Context, req *computepb.SetIamPolicyImageRequest, opts ...gaxv2.CallOption,
	) (*computepb.Policy, error)
	Delete(ctx context.Context, req *computepb.DeleteImageRequest, opts ...gaxv2.CallOption,
	) (*compute.Operation, error)
	io.Closer
}

type bucketAPI interface {
	Attrs(ctx context.Context) (attrs *storage.BucketAttrs, err error)
	Create(ctx context.Context, projectID string, attrs *storage.BucketAttrs) (err error)
	Object(name string) *storage.ObjectHandle
}

func toPtr[T any](v T) *T {
	return &v
}
