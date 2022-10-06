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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	nodeMaintenanceOperatorNamespace        = "kube-system"
	nodeMaintenanceOperatorCatalogNamespace = "olm"
)

type NodeMaintenanceOperatorDeployment struct {
	CatalogSource operatorsv1alpha1.CatalogSource
	OperatorGroup operatorsv1.OperatorGroup
	Subscription  operatorsv1alpha1.Subscription
}

// NewNodeMaintenanceOperatorDeployment creates a new node maintenance operator (NMO) deployment.
// See https://github.com/medik8s/node-maintenance-operator for more information.
func NewNodeMaintenanceOperatorDeployment() *NodeMaintenanceOperatorDeployment {
	return &NodeMaintenanceOperatorDeployment{
		CatalogSource: operatorsv1alpha1.CatalogSource{
			TypeMeta: metav1.TypeMeta{APIVersion: "operators.coreos.com/v1alpha1", Kind: "CatalogSource"},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "node-maintenance-operator-catalog",
				Namespace: nodeMaintenanceOperatorCatalogNamespace,
			},
			Spec: operatorsv1alpha1.CatalogSourceSpec{
				SourceType:  "grpc",
				Image:       versions.NodeMaintenanceOperatorCatalogImage,
				DisplayName: "Node Maintenance Operator",
				Publisher:   "Medik8s Team",
				UpdateStrategy: &operatorsv1alpha1.UpdateStrategy{
					RegistryPoll: &operatorsv1alpha1.RegistryPoll{
						RawInterval: "1m0s",
						Interval: &metav1.Duration{
							Duration: time.Minute,
						},
					},
				},
			},
		},
		OperatorGroup: operatorsv1.OperatorGroup{
			TypeMeta: metav1.TypeMeta{APIVersion: "operators.coreos.com/v1", Kind: "OperatorGroup"},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "constellation-og",
				Namespace: nodeMaintenanceOperatorNamespace,
			},
			Spec: operatorsv1.OperatorGroupSpec{
				UpgradeStrategy: operatorsv1.UpgradeStrategyDefault,
			},
		},
		Subscription: operatorsv1alpha1.Subscription{
			TypeMeta: metav1.TypeMeta{APIVersion: "operators.coreos.com/v1alpha1", Kind: "Subscription"},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "node-maintenance-operator-sub",
				Namespace: nodeMaintenanceOperatorNamespace,
			},
			Spec: &operatorsv1alpha1.SubscriptionSpec{
				Channel:                "stable",
				Package:                "node-maintenance-operator",
				CatalogSource:          "node-maintenance-operator-catalog",
				CatalogSourceNamespace: "olm",
				InstallPlanApproval:    operatorsv1alpha1.ApprovalAutomatic,
				StartingCSV:            "node-maintenance-operator." + versions.NodeMaintenanceOperatorVersion,
			},
		},
	}
}

func (c *NodeMaintenanceOperatorDeployment) Marshal() ([]byte, error) {
	return kubernetes.MarshalK8SResources(c)
}
