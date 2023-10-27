/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
imagefetch retrieves a CSP image reference from a Constellation config in the CWD.
This is especially useful when using self-managed infrastructure, where the image
reference needs to be chosen by the user, which would usually happen manually.
*/
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/edgelesssys/constellation/v2/internal/api/attestationconfigapi"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/imagefetcher"
	"github.com/spf13/afero"
)

var (
	caseInsensitiveCommunityGalleriesRegexp = regexp.MustCompile(`(?i)\/communitygalleries\/`)
	caseInsensitiveImagesRegExp             = regexp.MustCompile(`(?i)\/images\/`)
	caseInsensitiveVersionsRegExp           = regexp.MustCompile(`(?i)\/versions\/`)
)

func main() {
	cwd := os.Getenv("BUILD_WORKING_DIRECTORY") // set by Bazel, for bazel run compatibility
	ctx := context.Background()

	fh := file.NewHandler(afero.NewOsFs())
	attFetcher := attestationconfigapi.NewFetcher()
	conf, err := config.New(fh, filepath.Join(cwd, constants.ConfigFilename), attFetcher, true)
	var configValidationErr *config.ValidationError
	if errors.As(err, &configValidationErr) {
		fmt.Println(configValidationErr.LongMessage())
	}
	if err != nil {
		panic(err)
	}

	imgFetcher := imagefetcher.New()
	provider := conf.GetProvider()
	attestationVariant := conf.GetAttestationConfig().GetVariant()
	region := conf.GetRegion()
	image, err := imgFetcher.FetchReference(ctx, provider, attestationVariant, conf.Image, region)
	if err != nil {
		panic(err)
	}

	if provider == cloudprovider.Azure {
		image = caseInsensitiveCommunityGalleriesRegexp.ReplaceAllString(image, "/communityGalleries/")
		image = caseInsensitiveImagesRegExp.ReplaceAllString(image, "/images/")
		image = caseInsensitiveVersionsRegExp.ReplaceAllString(image, "/versions/")
	}

	fmt.Println(image)
}
