/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package main

import (
	"context"
	"fmt"

	"github.com/edgelesssys/constellation/v2/internal/api/attestationconfigapi"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
)

func deleteAWS(ctx context.Context, client *attestationconfigapi.Client, cfg deleteConfig) error {
	if cfg.provider != cloudprovider.AWS || cfg.kind != snpReport {
		return fmt.Errorf("provider %s and kind %s not supported", cfg.provider, cfg.kind)
	}

	return client.DeleteSEVSNPVersion(ctx, variant.AWSSEVSNP{}, cfg.version)
}
