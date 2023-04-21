/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/osimage"
	"github.com/edgelesssys/constellation/v2/internal/versionsapi"
)

func uploadImage(ctx context.Context, archiveC archivist, uploadC uploader, req *osimage.UploadRequest, out io.Writer) error {
	// upload to S3 archive
	archiveURL, err := archiveC.Archive(ctx, req.Version, req.Provider.String(), req.Variant, req.Image)
	if err != nil {
		return err
	}
	// rewind reader so we can read again
	if _, err := req.Image.Seek(0, io.SeekStart); err != nil {
		return err
	}
	// upload to CSP
	imageReferences, err := uploadC.Upload(ctx, req)
	if err != nil {
		return err
	}
	if len(imageReferences) == 0 {
		imageReferences = map[string]string{
			req.Variant: archiveURL,
		}
	}

	imageInfo := versionsapi.ImageInfo{
		Ref:     req.Version.Ref,
		Stream:  req.Version.Stream,
		Version: req.Version.Version,
	}
	switch req.Provider {
	case cloudprovider.AWS:
		imageInfo.AWS = imageReferences
	case cloudprovider.Azure:
		imageInfo.Azure = imageReferences
	case cloudprovider.GCP:
		imageInfo.GCP = imageReferences
	case cloudprovider.OpenStack:
		imageInfo.OpenStack = imageReferences
	case cloudprovider.QEMU:
		imageInfo.QEMU = imageReferences
	default:
		return fmt.Errorf("uploading image: cloud provider %s is not yet supported", req.Provider.String())
	}

	if err := json.NewEncoder(out).Encode(imageInfo); err != nil {
		return fmt.Errorf("uploading image: marshaling output: %w", err)
	}

	return nil
}
