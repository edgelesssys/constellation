/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// package nop implements a no-op for CSPs that don't require custom image upload functionality.
package nop

import (
	"context"
	"log/slog"
  "fmt"

	"github.com/edgelesssys/constellation/v2/internal/api/versionsapi"
	"github.com/edgelesssys/constellation/v2/internal/osimage"
)

// Uploader is a no-op uploader.
type Uploader struct {
	log *slog.Logger
}

// New creates a new Uploader.
func New(log *slog.Logger) *Uploader {
	return &Uploader{log: log}
}

// Upload pretends to upload images to a csp.
func (u *Uploader) Upload(_ context.Context, req *osimage.UploadRequest) ([]versionsapi.ImageInfoEntry, error) {
	u.log.Debug(fmt.Sprintf("Skipping image upload of %s since this CSP does not require images to be uploaded in advance.", req.Version.ShortPath()))
	return nil, nil
}
