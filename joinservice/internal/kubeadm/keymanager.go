package kubeadm

import (
	"fmt"
	"sync"
	"time"

	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/edgelesssys/constellation/internal/logger"
	clientset "k8s.io/client-go/kubernetes"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	"k8s.io/kubernetes/cmd/kubeadm/app/phases/copycerts"
	"k8s.io/utils/clock"
)

// certificateKeyTTL is the time a certificate key is valid for.
const certificateKeyTTL = time.Hour

// keyManager handles creation of certificate encryption keys.
type keyManager struct {
	mux            sync.Mutex
	key            string
	expirationDate time.Time
	clock          clock.Clock
	client         clientset.Interface
	log            *logger.Logger
}

func newKeyManager(client clientset.Interface, log *logger.Logger) *keyManager {
	return &keyManager{
		clock:  clock.RealClock{},
		client: client,
		log:    log,
	}
}

// getCertificatetKey returns the encryption key to use for uploading PKI certificates to Kubernetes.
// A Key is cached for one hour, but its expiration date is extended by two minutes if a request is made
// within two minutes of the key expiring to avoid just-expired keys.
// This is necessary since uploading a certificate with a different key overwrites any others.
// This means we can no longer decrypt the certificates using an old key.
func (k *keyManager) getCertificatetKey() (string, error) {
	k.mux.Lock()
	defer k.mux.Unlock()

	switch {
	case k.key == "" || k.expirationDate.Before(k.clock.Now()):
		// key was not yet generated, or has expired
		// generate a new key and set TTL
		key, err := copycerts.CreateCertificateKey()
		if err != nil {
			return "", fmt.Errorf("couldn't create control plane certificate key: %w", err)
		}
		k.expirationDate = k.clock.Now().Add(certificateKeyTTL)
		k.key = key
		k.log.Infof("Uploading certs to Kubernetes")
		cfg := &kubeadmapi.InitConfiguration{
			ClusterConfiguration: kubeadmapi.ClusterConfiguration{
				CertificatesDir: constants.KubeadmCertificateDir,
			},
		}
		if err := copycerts.UploadCerts(k.client, cfg, key); err != nil {
			return "", fmt.Errorf("uploading certs: %w", err)
		}
	case k.expirationDate.After(k.clock.Now()):
		// key is still valid
		// if TTL is less than 2 minutes away, increase it by 2 minutes
		// this is to avoid the key expiring too soon when a node uses it to join the cluster
		if k.expirationDate.Sub(k.clock.Now()) < 2*time.Minute {
			k.expirationDate = k.expirationDate.Add(2 * time.Minute)
		}
	}

	return k.key, nil
}
