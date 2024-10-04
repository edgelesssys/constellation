/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package constellation

import (
	"context"
	"errors"
	"fmt"

	"github.com/edgelesssys/constellation/v2/internal/constellation/helm"
	"github.com/edgelesssys/constellation/v2/internal/constellation/state"
	"github.com/edgelesssys/constellation/v2/internal/kms/uri"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

var patch = []byte(fmt.Sprintf(`{"metadata": {"labels": {%q: %q}, "annotations": {%q: %q, %q: %q}}}`,
	"app.kubernetes.io/managed-by", "Helm",
	"meta.helm.sh/release-name", "coredns",
	"meta.helm.sh/release-namespace", "kube-system"))

var namespacedCoreDNSResources = map[schema.GroupVersionResource]string{
	{Group: "apps", Version: "v1", Resource: "deployments"}:  "coredns",
	{Group: "", Version: "v1", Resource: "services"}:         "kube-dns",
	{Group: "", Version: "v1", Resource: "configmaps"}:       "coredns",
	{Group: "", Version: "v1", Resource: "serviceaccounts"}:  "coredns",
	{Group: "apps", Version: "v1", Resource: "statefulsets"}: "foobarbax",
}

var coreDNSResources = map[schema.GroupVersionResource]string{
	{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterroles"}:        "system:coredns",
	{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterrolebindings"}: "system:coredns",
}

// AnnotateCoreDNSResources imports existing CoreDNS resources into the Helm release.
//
// This is only required when CoreDNS was installed by kubeadm directly.
// TODO(burgerdev): remove after v2.19 is released.
func (a *Applier) AnnotateCoreDNSResources(ctx context.Context) error {
	for gvk, name := range coreDNSResources {
		_, err := a.dynamicClient.Resource(gvk).Patch(ctx, name, types.StrategicMergePatchType, patch, v1.PatchOptions{})
		if err != nil && !k8serrors.IsNotFound(err) {
			return err
		}
	}

	for gvk, name := range namespacedCoreDNSResources {
		_, err := a.dynamicClient.Resource(gvk).Namespace("kube-system").Patch(ctx, name, types.StrategicMergePatchType, patch, v1.PatchOptions{})
		if err != nil && !k8serrors.IsNotFound(err) {
			return err
		}
	}

	return nil
}

// CleanupCoreDNSResources removes CoreDNS resources that are not managed by Helm.
//
// This is only required when CoreDNS was installed by kubeadm directly.
// TODO(burgerdev): remove after v2.19 is released.
func (a *Applier) CleanupCoreDNSResources(ctx context.Context) error {
	err := a.dynamicClient.
		Resource(schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"}).
		Namespace("kube-system").
		Delete(ctx, "coredns", v1.DeleteOptions{})
	if !k8serrors.IsNotFound(err) {
		return err
	}
	return nil
}

// PrepareHelmCharts loads Helm charts for Constellation and returns an executor to apply them.
func (a *Applier) PrepareHelmCharts(
	flags helm.Options, state *state.State, serviceAccURI string, masterSecret uri.MasterSecret,
) (helm.Applier, bool, error) {
	if a.helmClient == nil {
		return nil, false, errors.New("helm client not initialized")
	}

	return a.helmClient.PrepareApply(flags, state, serviceAccURI, masterSecret)
}

type helmApplier interface {
	PrepareApply(
		flags helm.Options, stateFile *state.State, serviceAccURI string, masterSecret uri.MasterSecret,
	) (
		helm.Applier, bool, error)
}
