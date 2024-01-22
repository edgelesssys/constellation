//go:build e2e

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package upgrade

import (
	"fmt"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/semver"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
)

func servicesVersion(t *testing.T) (semver.Semver, error) {
	t.Helper()
	log := logger.NewTest(t)
	settings := cli.New()
	settings.KubeConfig = "constellation-admin.conf"
	actionConfig := &action.Configuration{}
	if err := actionConfig.Init(settings.RESTClientGetter(), constants.HelmNamespace, "secret", log.Info); err != nil {
		return semver.Semver{}, fmt.Errorf("initializing config: %w", err)
	}

	return currentVersion(actionConfig, "constellation-services")
}

func currentVersion(cfg *action.Configuration, release string) (semver.Semver, error) {
	action := action.NewList(cfg)
	action.Filter = release
	rel, err := action.Run()
	if err != nil {
		return semver.Semver{}, err
	}

	if len(rel) == 0 {
		return semver.Semver{}, fmt.Errorf("release %s not found", release)
	}
	if len(rel) > 1 {
		return semver.Semver{}, fmt.Errorf("multiple releases found for %s", release)
	}

	if rel[0] == nil || rel[0].Chart == nil || rel[0].Chart.Metadata == nil {
		return semver.Semver{}, fmt.Errorf("received invalid release %s", release)
	}

	return semver.New(rel[0].Chart.Metadata.Version)
}
