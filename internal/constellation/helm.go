/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package constellation

import (
	"errors"

	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constellation/helm"
	"github.com/edgelesssys/constellation/v2/internal/kms/uri"
	"github.com/edgelesssys/constellation/v2/internal/state"
)

// PrepareHelmCharts loads Helm charts for Constellation and returns an executor to apply them.
func (a *Applier) PrepareHelmCharts(
	flags helm.Options, state *state.State, serviceAccURI string, masterSecret uri.MasterSecret, openStackCfg *config.OpenStackConfig,
) (helm.Applier, bool, error) {
	if a.helmClient == nil {
		return nil, false, errors.New("helm client not initialized")
	}

	return a.helmClient.PrepareApply(flags, state, serviceAccURI, masterSecret, openStackCfg)
}

type helmApplier interface {
	PrepareApply(
		flags helm.Options, stateFile *state.State, serviceAccURI string, masterSecret uri.MasterSecret, openStackCfg *config.OpenStackConfig,
	) (
		helm.Applier, bool, error)
}
