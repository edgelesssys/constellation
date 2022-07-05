package kubeadm

import (
	"errors"
	"fmt"
	"time"

	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/edgelesssys/constellation/internal/file"
	"github.com/edgelesssys/constellation/internal/logger"
	"github.com/spf13/afero"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	certutil "k8s.io/client-go/util/cert"
	bootstraputil "k8s.io/cluster-bootstrap/token/util"
	bootstraptoken "k8s.io/kubernetes/cmd/kubeadm/app/apis/bootstraptoken/v1"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	kubeadm "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta3"
	tokenphase "k8s.io/kubernetes/cmd/kubeadm/app/phases/bootstraptoken/node"
	"k8s.io/kubernetes/cmd/kubeadm/app/phases/copycerts"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/kubeconfig"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/pubkeypin"
)

// Kubeadm manages joining of new nodes.
type Kubeadm struct {
	apiServerEndpoint string
	log               *logger.Logger
	client            clientset.Interface
	file              file.Handler
}

// New creates a new Kubeadm instance.
func New(apiServerEndpoint string, log *logger.Logger) (*Kubeadm, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get in-cluster config: %w", err)
	}
	client, err := clientset.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}
	file := file.NewHandler(afero.NewOsFs())

	return &Kubeadm{
		apiServerEndpoint: apiServerEndpoint,
		log:               log,
		client:            client,
		file:              file,
	}, nil
}

// GetJoinToken creates a new bootstrap (join) token, which a node can use to join the cluster.
func (k *Kubeadm) GetJoinToken(ttl time.Duration) (*kubeadm.BootstrapTokenDiscovery, error) {
	k.log.Infof("Generating new random bootstrap token")
	rawToken, err := bootstraputil.GenerateBootstrapToken()
	if err != nil {
		return nil, fmt.Errorf("couldn't generate random token: %w", err)
	}
	tokenStr, err := bootstraptoken.NewBootstrapTokenString(rawToken)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}
	token := bootstraptoken.BootstrapToken{
		Token:       tokenStr,
		Description: "Bootstrap token generated by Constellation's Join service",
		TTL:         &metav1.Duration{Duration: ttl},
		Usages:      []string{"signing", "authentication"},
		Groups:      []string{"system:bootstrappers:kubeadm:default-node-token"},
	}

	// create the token in Kubernetes
	k.log.Infof("Creating bootstrap token in Kubernetes")
	if err := tokenphase.CreateNewTokens(k.client, []bootstraptoken.BootstrapToken{token}); err != nil {
		return nil, fmt.Errorf("creating bootstrap token: %w", err)
	}

	// parse Kubernetes CA certs
	k.log.Infof("Preparing join token for new node")
	rawConfig, err := k.file.Read(constants.CoreOSAdminConfFilename)
	if err != nil {
		return nil, fmt.Errorf("loading kubeconfig file: %w", err)
	}
	config, err := clientcmd.Load(rawConfig)
	if err != nil {
		return nil, fmt.Errorf("loading kubeconfig file: %w", err)
	}
	clusterConfig := kubeconfig.GetClusterFromKubeConfig(config)
	if clusterConfig == nil {
		return nil, errors.New("couldn't get cluster config from kubeconfig file")
	}
	caCerts, err := certutil.ParseCertsPEM(clusterConfig.CertificateAuthorityData)
	if err != nil {
		return nil, fmt.Errorf("parsing CA certs: %w", err)
	}
	publicKeyPins := make([]string, 0, len(caCerts))
	for _, caCert := range caCerts {
		publicKeyPins = append(publicKeyPins, pubkeypin.Hash(caCert))
	}

	k.log.Infof("Join token creation successful")
	return &kubeadm.BootstrapTokenDiscovery{
		Token:             tokenStr.String(),
		APIServerEndpoint: k.apiServerEndpoint,
		CACertHashes:      publicKeyPins,
	}, nil
}

// GetControlPlaneCertificateKey uploads Kubernetes encrypted CA certificates to Kubernetes and returns the decryption key.
// The key can be used by new nodes to join the cluster as a control plane node.
func (k *Kubeadm) GetControlPlaneCertificateKey() (string, error) {
	k.log.Infof("Creating new random control plane certificate key")
	key, err := copycerts.CreateCertificateKey()
	if err != nil {
		return "", fmt.Errorf("couldn't create control plane certificate key: %w", err)
	}

	k.log.Infof("Uploading certs to Kubernetes")
	cfg := &kubeadmapi.InitConfiguration{
		ClusterConfiguration: kubeadmapi.ClusterConfiguration{
			CertificatesDir: constants.KubeadmCertificateDir,
		},
	}
	if err := copycerts.UploadCerts(k.client, cfg, key); err != nil {
		return "", fmt.Errorf("uploading certs: %w", err)
	}

	return key, nil
}
