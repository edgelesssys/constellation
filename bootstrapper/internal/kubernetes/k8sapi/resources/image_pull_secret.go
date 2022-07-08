package resources

import (
	"encoding/base64"
	"fmt"

	"github.com/edgelesssys/constellation/internal/secrets"
	k8s "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewImagePullSecret creates a new k8s.Secret from the config for authenticating when pulling images.
func NewImagePullSecret() k8s.Secret {
	base64EncodedSecret := base64.StdEncoding.EncodeToString(
		[]byte(fmt.Sprintf("%s:%s", secrets.PullSecretUser, secrets.PullSecretToken)),
	)

	pullSecretDockerCfgJSON := fmt.Sprintf(`{"auths":{"ghcr.io":{"auth":"%s"}}}`, base64EncodedSecret)

	return k8s.Secret{
		TypeMeta: meta.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: meta.ObjectMeta{
			Name:      secrets.PullSecretName,
			Namespace: "kube-system",
		},
		StringData: map[string]string{".dockerconfigjson": pullSecretDockerCfgJSON},
		Type:       "kubernetes.io/dockerconfigjson",
	}
}
