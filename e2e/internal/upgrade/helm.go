package test

import (
	"fmt"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
)

func servicesVersion(t *testing.T) (string, error) {
	t.Helper()
	log := logger.NewTest(t)
	settings := cli.New()
	settings.KubeConfig = "constellation-admin.conf"
	actionConfig := &action.Configuration{}
	if err := actionConfig.Init(settings.RESTClientGetter(), constants.HelmNamespace, "secret", log.Infof); err != nil {
		return "", fmt.Errorf("initializing config: %w", err)
	}

	return currentVersion(actionConfig, "constellation-services")
}

func currentVersion(cfg *action.Configuration, release string) (string, error) {
	action := action.NewList(cfg)
	action.Filter = release
	rel, err := action.Run()
	if err != nil {
		return "", err
	}

	if len(rel) == 0 {
		return "", fmt.Errorf("release %s not found", release)
	}
	if len(rel) > 1 {
		return "", fmt.Errorf("multiple releases found for %s", release)
	}

	if rel[0] == nil || rel[0].Chart == nil || rel[0].Chart.Metadata == nil {
		return "", fmt.Errorf("received invalid release %s", release)
	}

	return rel[0].Chart.Metadata.Version, nil
}
