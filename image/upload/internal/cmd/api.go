/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"context"
	"io"

	versionsapi "github.com/edgelesssys/constellation/v2/internal/api/versions"
	"github.com/edgelesssys/constellation/v2/internal/osimage"
)

type archivist interface {
	Archive(ctx context.Context,
		version versionsapi.Version, csp, attestationVariant string, img io.Reader,
	) (string, error)
}

type uploader interface {
	Upload(ctx context.Context, req *osimage.UploadRequest) ([]versionsapi.ImageInfoEntry, error)
}
