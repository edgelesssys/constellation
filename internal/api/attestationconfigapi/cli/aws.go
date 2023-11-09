/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package main

import (
	"context"

	"github.com/edgelesssys/constellation/v2/internal/api/attestationconfigapi"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/logger"
)

func uploadAWS(_ context.Context, _ *attestationconfigapi.Client, _ uploadConfig, _ file.Handler, _ *logger.Logger) error {
	return nil
}

func deleteAWS(_ context.Context, _ *attestationconfigapi.Client, _ deleteConfig) error {
	return nil
}
