/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
Package cmd provides the Constellation CLI.

It is responsible for the interaction with the user.
*/
package cmd

import (
	"fmt"

	"github.com/edgelesssys/constellation/v2/internal/api/versionsapi"
	"github.com/edgelesssys/constellation/v2/internal/semver"
)

// Must panics if err is not nil.
func Must(err error) {
	if err != nil {
		panic(err)
	}
}

// ValidateCLIandConstellationVersionAreEqual checks if the image and microservice version are equal (down to patch level) to the CLI version.
func ValidateCLIandConstellationVersionAreEqual(cliVersion semver.Semver, imageVersion string, microserviceVersion semver.Semver) error {
	parsedImageVersion, err := versionsapi.NewVersionFromShortPath(imageVersion, versionsapi.VersionKindImage)
	if err != nil {
		return fmt.Errorf("parsing image version: %w", err)
	}

	semImage, err := semver.New(parsedImageVersion.Version())
	if err != nil {
		return fmt.Errorf("parsing image semantical version: %w", err)
	}

	if !cliVersion.MajorMinorEqual(semImage) {
		return fmt.Errorf("image version %q does not match the major and minor version of the cli version %q", semImage.String(), cliVersion.String())
	}
	if cliVersion.Compare(microserviceVersion) != 0 {
		return fmt.Errorf("cli version %q does not match microservice version %q", cliVersion.String(), microserviceVersion.String())
	}
	return nil
}
