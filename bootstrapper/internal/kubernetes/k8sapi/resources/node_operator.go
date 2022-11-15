/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package resources

import (
	"time"

	"github.com/edgelesssys/constellation/v2/internal/kubernetes"
	"github.com/edgelesssys/constellation/v2/internal/versions"
	operatorsv1 "github.com/operator-framework/api/pkg/operators/v1"
	operatorsv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	nodeOperatorNamespace        = "kube-system"
	nodeOperatorCatalogNamespace = "olm"
)

// NodeOperatorDeployment groups all deployments for node operator.
type NodeOperatorDeployment struct {
	CatalogSource operatorsv1alpha1.CatalogSource
	OperatorGroup operatorsv1.OperatorGroup
	Subscription  operatorsv1alpha1.Subscription
}

// NewNodeOperatorDeployment creates a new constellation node operator deployment.
// See /operators/constellation-node-operator for more information.
func NewNodeOperatorDeployment(cloudProvider string, uid string) *NodeOperatorDeployment {
	return &NodeOperatorDeployment{
		CatalogSource: operatorsv1alpha1.CatalogSource{
			TypeMeta: metav1.TypeMeta{APIVersion: "operators.coreos.com/v1alpha1", Kind: "CatalogSource"},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "constellation-node-operator-catalog",
				Namespace: nodeOperatorCatalogNamespace,
			},
			Spec: operatorsv1alpha1.CatalogSourceSpec{
				SourceType:  "grpc",
				Image:       versions.NodeOperatorCatalogImage,
				DisplayName: "Constellation Node Operator",
				Publisher:   "Edgeless Systems",
				UpdateStrategy: &operatorsv1alpha1.UpdateStrategy{
					RegistryPoll: &operatorsv1alpha1.RegistryPoll{
						RawInterval: "1m0s",
						Interval:    &metav1.Duration{Duration: 1 * time.Minute},
					},
				},
			},
		},
		OperatorGroup: operatorsv1.OperatorGroup{
			TypeMeta: metav1.TypeMeta{APIVersion: "operators.coreos.com/v1", Kind: "OperatorGroup"},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "constellation-og",
				Namespace: nodeOperatorNamespace,
			},
			Spec: operatorsv1.OperatorGroupSpec{
				UpgradeStrategy: operatorsv1.UpgradeStrategyDefault,
			},
		},
		Subscription: operatorsv1alpha1.Subscription{
			TypeMeta: metav1.TypeMeta{APIVersion: "operators.coreos.com/v1alpha1", Kind: "Subscription"},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "constellation-node-operator-sub",
				Namespace: nodeOperatorNamespace,
			},
			Spec: &operatorsv1alpha1.SubscriptionSpec{
				Channel:                "alpha",
				Package:                "node-operator",
				CatalogSource:          "constellation-node-operator-catalog",
				CatalogSourceNamespace: "olm",
				InstallPlanApproval:    operatorsv1alpha1.ApprovalAutomatic,
				StartingCSV:            "node-operator." + versions.NodeOperatorVersion,
				Config: &operatorsv1alpha1.SubscriptionConfig{
					Env: []corev1.EnvVar{
						{Name: "CONSTEL_CSP", Value: cloudProvider},
						{Name: "constellation-uid", Value: uid},
					},
				},
			},
		},
	}
}

// Marshal to Kubernetes YAML.
func (c *NodeOperatorDeployment) Marshal() ([]byte, error) {
	return kubernetes.MarshalK8SResources(c)
}
