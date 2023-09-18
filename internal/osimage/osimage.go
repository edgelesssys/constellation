/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// package osimage is used to handle osimages in the CI (uploading and maintenance).
package osimage

import (
	"io"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/api/versionsapi"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/osimage/secureboot"
)

// UploadRequest is a request to upload an os image.
type UploadRequest struct {
	Provider           cloudprovider.Provider
	Version            versionsapi.Version
	AttestationVariant string
	SecureBoot         bool
	SBDatabase         secureboot.Database
	UEFIVarStore       secureboot.UEFIVarStore
	Size               int64
	Timestamp          time.Time
	Image              io.ReadSeeker
}
