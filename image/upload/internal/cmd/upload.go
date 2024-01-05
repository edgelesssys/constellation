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
	"strings"

	"github.com/edgelesssys/constellation/v2/internal/api/versionsapi"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/osimage"
)

func uploadImage(ctx context.Context, archiveC archivist, uploadC uploader, req *osimage.UploadRequest, out io.Writer) error {
	// upload to S3 archive
	imageReader, err := req.ImageReader()
	if err != nil {
		return err
	}
	defer imageReader.Close()

	var archiveURL string
	if needsArchival(req.Provider, req.Version) {
		var err error
		archiveURL, err = archiveC.Archive(ctx, req.Version, strings.ToLower(req.Provider.String()), req.AttestationVariant, imageReader)
		if err != nil {
			return err
		}
	}

	// upload to CSP
	imageReferences, err := uploadC.Upload(ctx, req)
	if err != nil {
		return err
	}
	if len(imageReferences) == 0 {
		imageReferences = []versionsapi.ImageInfoEntry{
			{
				CSP:                req.Provider.String(),
				AttestationVariant: req.AttestationVariant,
				Reference:          archiveURL,
			},
		}
	}

	imageInfo := versionsapi.ImageInfo{
		Ref:     req.Version.Ref(),
		Stream:  req.Version.Stream(),
		Version: req.Version.Version(),
		List:    imageReferences,
	}

	if err := json.NewEncoder(out).Encode(imageInfo); err != nil {
		return fmt.Errorf("uploading image: marshaling output: %w", err)
	}

	return nil
}

func needsArchival(provider cloudprovider.Provider, version versionsapi.Version) bool {
	switch provider {
	case cloudprovider.OpenStack, cloudprovider.QEMU: // image upload for some CSPs only consists of this archival step
		return true
	}

	return version.Stream() == "stable" || version.Ref() == "-"
}
