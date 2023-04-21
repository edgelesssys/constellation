/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"context"
	"io"

	"github.com/edgelesssys/constellation/v2/internal/osimage"
	"github.com/edgelesssys/constellation/v2/internal/versionsapi"
)

type archivist interface {
	Archive(ctx context.Context,
		version versionsapi.Version, csp, variant string, img io.Reader,
	) (string, error)
}

type uploader interface {
	Upload(ctx context.Context, req *osimage.UploadRequest) (map[string]string, error)
}
