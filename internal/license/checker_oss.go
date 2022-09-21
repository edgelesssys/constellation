//go:build !enterprise

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package license

import (
	"context"

	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/file"
)

type Checker struct{}

func NewChecker(quotaChecker QuotaChecker, fileHandler file.Handler) *Checker {
	return &Checker{}
}

// CheckLicense is a no-op for open source version of Constellation.
func (c *Checker) CheckLicense(ctx context.Context, provider cloudprovider.Provider, providerCfg config.ProviderConfig, printer func(string, ...any)) error {
	return nil
}
