package verify

import (
	"context"
	"fmt"

	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/kubernetes/kubectl"
)

func getCertChainCache(ctx context.Context, kubectl *kubectl.Kubectl, log debugLog) ([]byte, error) {
	log.Debugf("Retrieving certificate chain from cache")
	cm, err := kubectl.GetConfigMap(ctx, constants.ConstellationNamespace, constants.SevSnpCertCacheConfigMapName)
	if err != nil {
		return nil, fmt.Errorf("getting certificate chain cache configmap: %w", err)
	}

	var result []byte
	ask, ok := cm.Data[constants.CertCacheAskKey]
	if ok {
		result = append(result, ask...)
	}
	ark, ok := cm.Data[constants.CertCacheArkKey]
	if ok {
		result = append(result, ark...)
	}

	return result, nil
}
