/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package k8sapi

import (
	"context"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/edgelesssys/constellation/v2/bootstrapper/internal/certificate"
	"github.com/edgelesssys/constellation/v2/bootstrapper/internal/kubernetes/k8sapi/resources"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/versions/components"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apiserver/pkg/authentication/user"
	kubeconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"

	"github.com/edgelesssys/constellation/v2/internal/crypto"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/installer"
	"github.com/spf13/afero"
)

const (
	// kubeletStartTimeout is the maximum time given to the kubelet service to (re)start.
	kubeletStartTimeout = 10 * time.Minute

	kubeletServicePath = "/usr/lib/systemd/system/kubelet.service"
)

// Client provides the functions to talk to the k8s API.
type Client interface {
	Initialize(kubeconfig []byte) error
	CreateConfigMap(ctx context.Context, configMap *corev1.ConfigMap) error
	AddNodeSelectorsToDeployment(ctx context.Context, selectors map[string]string, name string, namespace string) error
	ListAllNamespaces(ctx context.Context) (*corev1.NamespaceList, error)
	AnnotateNode(ctx context.Context, nodeName, annotationKey, annotationValue string) error
	PatchFirstNodePodCIDR(ctx context.Context, firstNodePodCIDR string) error
}

type componentsInstaller interface {
	Install(
		ctx context.Context, kubernetesComponent *components.Component,
	) error
}

// KubernetesUtil provides low level management of the kubernetes cluster.
type KubernetesUtil struct {
	inst componentsInstaller
	file file.Handler
}

// NewKubernetesUtil creates a new KubernetesUtil.
func NewKubernetesUtil() *KubernetesUtil {
	return &KubernetesUtil{
		inst: installer.NewOSInstaller(),
		file: file.NewHandler(afero.NewOsFs()),
	}
}

// InstallComponents installs the kubernetes components passed from the CLI.
func (k *KubernetesUtil) InstallComponents(ctx context.Context, kubernetesComponents components.Components) error {
	for _, component := range kubernetesComponents {
		if err := k.inst.Install(ctx, component); err != nil {
			return fmt.Errorf("installing kubernetes component from URL %s: %w", component.Url, err)
		}
	}

	return enableSystemdUnit(ctx, kubeletServicePath)
}

// InitCluster instruments kubeadm to initialize the K8s cluster.
// On success an admin kubeconfig file is returned.
func (k *KubernetesUtil) InitCluster(
	ctx context.Context, initConfig []byte, nodeName, clusterName string, ips []net.IP, conformanceMode bool, log *slog.Logger,
) ([]byte, error) {
	// TODO(3u13r): audit policy should be user input
	auditPolicy, err := resources.NewDefaultAuditPolicy().Marshal()
	if err != nil {
		return nil, fmt.Errorf("generating default audit policy: %w", err)
	}
	if err := os.WriteFile(auditPolicyPath, auditPolicy, 0o644); err != nil {
		return nil, fmt.Errorf("writing default audit policy: %w", err)
	}

	initConfigFile, err := os.CreateTemp("", "kubeadm-init.*.yaml")
	if err != nil {
		return nil, fmt.Errorf("creating init config file %v: %w", initConfigFile.Name(), err)
	}

	if _, err := initConfigFile.Write(initConfig); err != nil {
		return nil, fmt.Errorf("writing kubeadm init yaml config %v: %w", initConfigFile.Name(), err)
	}

	// preflight
	log.Info("Running kubeadm preflight checks")
	cmd := exec.CommandContext(ctx, constants.KubeadmPath, "init", "phase", "preflight", "-v=5", "--config", initConfigFile.Name())
	out, err := cmd.CombinedOutput()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return nil, fmt.Errorf("kubeadm init phase preflight failed (code %v) with: %s", exitErr.ExitCode(), out)
		}
		return nil, fmt.Errorf("kubeadm init: %w", err)
	}

	// create CA certs
	log.Info("Creating Kubernetes control-plane certificates and keys")
	cmd = exec.CommandContext(ctx, constants.KubeadmPath, "init", "phase", "certs", "all", "-v=5", "--config", initConfigFile.Name())
	out, err = cmd.CombinedOutput()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return nil, fmt.Errorf("kubeadm init phase certs all failed (code %v) with: %s", exitErr.ExitCode(), out)
		}
		return nil, fmt.Errorf("kubeadm init: %w", err)
	}

	// create kubelet key and CA signed certificate for the node
	log.Info("Creating signed kubelet certificate")
	if err := k.createSignedKubeletCert(nodeName, ips); err != nil {
		return nil, fmt.Errorf("creating signed kubelete certificate: %w", err)
	}

	// Create static pods directory for all nodes (the Kubelets on the worker nodes also expect the path to exist)
	// If the node rebooted after the static pod directory was created,
	// the existing directory needs to be removed before we can
	// try to init the cluster again.
	if err := os.RemoveAll("/etc/kubernetes/manifests"); err != nil {
		return nil, fmt.Errorf("removing static pods directory: %w", err)
	}
	log.Info("Creating static Pod directory /etc/kubernetes/manifests")
	if err := os.MkdirAll("/etc/kubernetes/manifests", os.ModePerm); err != nil {
		return nil, fmt.Errorf("creating static pods directory: %w", err)
	}

	// initialize the cluster
	log.Info("Initializing the cluster using kubeadm init")
	skipPhases := "--skip-phases=preflight,certs,addon/coredns"
	if !conformanceMode {
		skipPhases += ",addon/kube-proxy"
	}

	cmd = exec.CommandContext(ctx, constants.KubeadmPath, "init", "-v=5", skipPhases, "--config", initConfigFile.Name())
	out, err = cmd.CombinedOutput()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return nil, fmt.Errorf("kubeadm init failed (code %v) with: %s", exitErr.ExitCode(), out)
		}
		return nil, fmt.Errorf("kubeadm init: %w", err)
	}
	log.With(slog.String("output", string(out))).Info("kubeadm init succeeded")

	userName := clusterName + "-admin"

	log.With(slog.String("userName", userName)).Info("Creating admin kubeconfig file")
	cmd = exec.CommandContext(
		ctx, constants.KubeadmPath, "kubeconfig", "user",
		"--client-name", userName, "--config", initConfigFile.Name(), "--org", user.SystemPrivilegedGroup,
	)
	out, err = cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return nil, fmt.Errorf("kubeadm kubeconfig user failed (code %v) with: %s", exitErr.ExitCode(), out)
		}
		return nil, fmt.Errorf("kubeadm kubeconfig user: %w", err)
	}
	log.Info("kubeadm kubeconfig user succeeded")
	return out, nil
}

// JoinCluster joins existing Kubernetes cluster using kubeadm join.
func (k *KubernetesUtil) JoinCluster(ctx context.Context, joinConfig []byte, log *slog.Logger) error {
	// TODO(3u13r): audit policy should be user input
	auditPolicy, err := resources.NewDefaultAuditPolicy().Marshal()
	if err != nil {
		return fmt.Errorf("generating default audit policy: %w", err)
	}
	if err := os.WriteFile(auditPolicyPath, auditPolicy, 0o644); err != nil {
		return fmt.Errorf("writing default audit policy: %w", err)
	}

	joinConfigFile, err := os.CreateTemp("", "kubeadm-join.*.yaml")
	if err != nil {
		return fmt.Errorf("creating join config file %v: %w", joinConfigFile.Name(), err)
	}

	if _, err := joinConfigFile.Write(joinConfig); err != nil {
		return fmt.Errorf("writing kubeadm init yaml config %v: %w", joinConfigFile.Name(), err)
	}

	// Create static pods directory for all nodes (the Kubelets on the worker nodes also expect the path to exist)
	// If the node rebooted after the static pod directory was created, for example
	// if a failure during an upgrade occurred, the existing directory needs to be
	// removed before we can try to join the cluster again.
	if err := os.RemoveAll("/etc/kubernetes/manifests"); err != nil {
		return fmt.Errorf("removing static pods directory: %w", err)
	}
	log.Info("Creating static Pod directory /etc/kubernetes/manifests")
	if err := os.MkdirAll("/etc/kubernetes/manifests", os.ModePerm); err != nil {
		return fmt.Errorf("creating static pods directory: %w", err)
	}

	// run `kubeadm join` to join a worker node to an existing Kubernetes cluster
	cmd := exec.CommandContext(ctx, constants.KubeadmPath, "join", "-v=5", "--config", joinConfigFile.Name())
	out, err := cmd.CombinedOutput()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return fmt.Errorf("kubeadm join failed (code %v) with: %s (full err: %s)", exitErr.ExitCode(), out, err)
		}
		return fmt.Errorf("kubeadm join: %w", err)
	}
	log.With(slog.String("output", string(out))).Info("kubeadm join succeeded")

	return nil
}

// StartKubelet enables and starts the kubelet systemd unit.
func (k *KubernetesUtil) StartKubelet() error {
	ctx, cancel := context.WithTimeout(context.TODO(), kubeletStartTimeout)
	defer cancel()
	if err := enableSystemdUnit(ctx, kubeletServicePath); err != nil {
		return fmt.Errorf("enabling kubelet systemd unit: %w", err)
	}
	return startSystemdUnit(ctx, "kubelet.service")
}

// createSignedKubeletCert manually creates a Kubernetes CA signed kubelet certificate for the bootstrapper node.
// This is necessary because this node does not request a certificate from the join service.
func (k *KubernetesUtil) createSignedKubeletCert(nodeName string, ips []net.IP) error {
	// Create CSR
	certRequestRaw, kubeletKey, err := certificate.GetKubeletCertificateRequest(nodeName, ips)
	if err != nil {
		return err
	}
	if err := k.file.Write(certificate.KeyFilename, kubeletKey, file.OptMkdirAll); err != nil {
		return err
	}

	certRequest, err := x509.ParseCertificateRequest(certRequestRaw)
	if err != nil {
		return err
	}

	// Prepare certificate signing
	serialNumber, err := crypto.GenerateCertificateSerialNumber()
	if err != nil {
		return err
	}

	now := time.Now()
	// Create the kubelet certificate
	// For a reference on the certificate fields, see: https://kubernetes.io/docs/setup/best-practices/certificates/
	certTmpl := &x509.Certificate{
		SerialNumber: serialNumber,
		NotBefore:    now.Add(-2 * time.Hour),
		NotAfter:     now.Add(24 * 365 * time.Hour),
		Subject:      certRequest.Subject,
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageClientAuth,
			x509.ExtKeyUsageServerAuth,
		},
		IsCA:                  false,
		BasicConstraintsValid: true,
		IPAddresses:           certRequest.IPAddresses,
	}

	parentCert, parentKey, err := k.getKubernetesCACertAndKey()
	if err != nil {
		return err
	}

	// Sign the certificate
	certRaw, err := x509.CreateCertificate(rand.Reader, certTmpl, parentCert, certRequest.PublicKey, parentKey)
	if err != nil {
		return err
	}

	// Write the certificate
	kubeletCert := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certRaw,
	})

	return k.file.Write(certificate.CertificateFilename, kubeletCert, file.OptMkdirAll)
}

// getKubernetesCACertAndKey returns the Kubernetes CA certificate and key.
// The key of type `any` can be consumed by `x509.CreateCertificate()`.
func (k *KubernetesUtil) getKubernetesCACertAndKey() (*x509.Certificate, any, error) {
	parentCertRaw, err := k.file.Read(filepath.Join(
		kubeconstants.KubernetesDir,
		kubeconstants.DefaultCertificateDir,
		kubeconstants.CACertName,
	))
	if err != nil {
		return nil, nil, err
	}
	parentCert, err := crypto.PemToX509Cert(parentCertRaw)
	if err != nil {
		return nil, nil, err
	}

	parentKeyRaw, err := k.file.Read(filepath.Join(
		kubeconstants.KubernetesDir,
		kubeconstants.DefaultCertificateDir,
		kubeconstants.CAKeyName,
	))
	if err != nil {
		return nil, nil, err
	}
	parentKeyPEM, _ := pem.Decode(parentKeyRaw)
	var parentKey any
	switch parentKeyPEM.Type {
	case "EC PRIVATE KEY":
		parentKey, err = x509.ParseECPrivateKey(parentKeyPEM.Bytes)
	case "RSA PRIVATE KEY":
		parentKey, err = x509.ParsePKCS1PrivateKey(parentKeyPEM.Bytes)
	case "PRIVATE KEY":
		parentKey, err = x509.ParsePKCS8PrivateKey(parentKeyPEM.Bytes)
	default:
		err = fmt.Errorf("unsupported key type %q", parentKeyPEM.Type)
	}
	if err != nil {
		return nil, nil, err
	}
	return parentCert, parentKey, nil
}
