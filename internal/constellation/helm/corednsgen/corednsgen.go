/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// corednsgen synthesizes a Helm chart from the resource templates embedded in
// kubeadm and writes it to the `charts` directory underneath the current
// working directory. This removes the existing `coredns` subdirectory!
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/edgelesssys/constellation/v2/internal/versions"
	"github.com/regclient/regclient"
	"github.com/regclient/regclient/types/ref"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	"k8s.io/kubernetes/cmd/kubeadm/app/images"
	kubedns "k8s.io/kubernetes/cmd/kubeadm/app/phases/addons/dns"
	kubeadmutil "k8s.io/kubernetes/cmd/kubeadm/app/util"
	"sigs.k8s.io/yaml"
)

const configMapName = "edg-coredns"

var chartDir = flag.String("charts", "./charts", "target directory to create charts in")

func main() {
	flag.Parse()

	if err := os.RemoveAll(filepath.Join(*chartDir, "coredns")); err != nil {
		log.Fatalf("Could not remove chart dir: %v", err)
	}

	writeFileRelativeToChartDir(chartYAML(), "Chart.yaml")
	writeFileRelativeToChartDir(valuesYAML(), "values.yaml")

	writeTemplate(kubedns.CoreDNSServiceAccount, "serviceaccount.yaml")
	writeTemplate(kubedns.CoreDNSClusterRole, "clusterrole.yaml")
	writeTemplate(kubedns.CoreDNSClusterRoleBinding, "clusterrolebinding.yaml")
	writeTemplate(kubedns.CoreDNSService, "service.yaml")

	writeFileRelativeToChartDir(patchedConfigMap(), "templates", "configmap.yaml")
	writeFileRelativeToChartDir(patchedDeployment(), "templates", "deployment.yaml")
}

func chartYAML() []byte {
	chart := map[string]string{
		"apiVersion": "v2",
		"name":       "kube-dns",
		"version":    "0.0.0",
	}
	data, err := yaml.Marshal(chart)
	if err != nil {
		log.Fatalf("Could not marshal Chart.yaml: %v", err)
	}
	return data
}

func valuesYAML() []byte {
	cfg := &kubeadm.ClusterConfiguration{
		KubernetesVersion: string(versions.Default),
		ImageRepository:   "registry.k8s.io",
	}
	img := images.GetDNSImage(cfg)
	ref, err := ref.New(img)
	if err != nil {
		log.Fatalf("Could not parse image reference: %v", err)
	}

	rc := regclient.New()
	m, err := rc.ManifestGet(context.Background(), ref)
	if err != nil {
		log.Fatalf("Could not get image manifest: %v", err)
	}

	values := map[string]string{
		"clusterIP": "10.96.0.10",
		"dnsDomain": "cluster.local",
		"image":     fmt.Sprintf("%s/%s:%s@%s", ref.Registry, ref.Repository, ref.Tag, m.GetDescriptor().Digest.String()),
	}
	data, err := yaml.Marshal(values)
	if err != nil {
		log.Fatalf("Could not marshal values.yaml: %v", err)
	}
	return data
}

// patchedConfigMap renames the CoreDNS ConfigMap such that kubeadm does not find it.
//
// See https://github.com/kubernetes/kubeadm/issues/2846#issuecomment-1899942683.
func patchedConfigMap() []byte {
	var cm corev1.ConfigMap
	if err := yaml.Unmarshal(parseTemplate(kubedns.CoreDNSConfigMap), &cm); err != nil {
		log.Fatalf("Could not parse configmap: %v", err)
	}

	cm.Name = configMapName

	out, err := yaml.Marshal(cm)
	if err != nil {
		log.Fatalf("Could not marshal patched deployment: %v", err)
	}
	return out
}

// patchedDeployment extracts the CoreDNS Deployment from kubeadm, adds necessary tolerations and updates the ConfigMap reference.
func patchedDeployment() []byte {
	var d appsv1.Deployment
	if err := yaml.Unmarshal(parseTemplate(kubedns.CoreDNSDeployment), &d); err != nil {
		log.Fatalf("Could not parse deployment: %v", err)
	}

	tolerations := []corev1.Toleration{
		{Key: "node.cloudprovider.kubernetes.io/uninitialized", Value: "true", Effect: corev1.TaintEffectNoSchedule},
		{Key: "node.kubernetes.io/unreachable", Operator: corev1.TolerationOpExists, Effect: corev1.TaintEffectNoExecute, TolerationSeconds: toPtr(int64(10))},
	}
	d.Spec.Template.Spec.Tolerations = append(d.Spec.Template.Spec.Tolerations, tolerations...)

	for i, vol := range d.Spec.Template.Spec.Volumes {
		if vol.ConfigMap != nil {
			vol.ConfigMap.Name = configMapName
		}
		d.Spec.Template.Spec.Volumes[i] = vol
	}

	out, err := yaml.Marshal(d)
	if err != nil {
		log.Fatalf("Could not marshal patched deployment: %v", err)
	}
	return out
}

func writeFileRelativeToChartDir(content []byte, pathElements ...string) {
	p := filepath.Join(append([]string{*chartDir, "coredns"}, pathElements...)...)
	d := filepath.Dir(p)
	if err := os.MkdirAll(d, 0o755); err != nil {
		log.Fatalf("Could not create dir %q: %v", d, err)
	}
	if err := os.WriteFile(p, content, 0o644); err != nil {
		log.Fatalf("Could not write file %q: %v", p, err)
	}
}

// parseTemplate replaces the Go template placeholders in kubeadm resources
// with fixed values or Helm value placeholders.
func parseTemplate(tmpl string) []byte {
	vars := struct {
		DeploymentName, Image, ControlPlaneTaintKey, DNSDomain, DNSIP string
		Replicas                                                      *int32
	}{
		DeploymentName:       "coredns",
		DNSDomain:            `{{ .Values.dnsDomain }}`,
		DNSIP:                `"{{ .Values.clusterIP }}"`,
		Image:                `"{{ .Values.image }}"`,
		ControlPlaneTaintKey: "node-role.kubernetes.io/control-plane",
		Replicas:             toPtr(int32(2)),
	}
	data, err := kubeadmutil.ParseTemplate(tmpl, vars)
	if err != nil {
		log.Fatalf("Could not interpolate template: %v", err)
	}
	return data
}

func writeTemplate(tmpl string, name string) {
	data := parseTemplate(tmpl)
	writeFileRelativeToChartDir(data, "templates", name)
}

func toPtr[T any](v T) *T {
	return &v
}
