//go:build integration

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package controllers

import (
	"context"
	"fmt"
	"time"

	//revive:disable:dot-imports
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	//revive:enable:dot-imports

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	nodemaintenancev1beta1 "github.com/edgelesssys/constellation/v2/3rdparty/node-maintenance-operator/api/v1beta1"
	mainconstants "github.com/edgelesssys/constellation/v2/internal/constants"
	updatev1alpha1 "github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/api/v1alpha1"
)

var _ = Describe("NodeVersion controller", func() {
	AfterEach(func() {
		Eventually(func() error {
			return resetEnv()
		}, 30*time.Second, 1*time.Second).Should(Succeed())
	})

	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		nodeVersionResourceName = "nodeversion"
		firstNodeName           = "node-1"
		secondNodeName          = "node-2"
		scalingGroupID          = "scaling-group"

		timeout  = time.Second * 20
		duration = time.Second * 2
		interval = time.Millisecond * 250
	)
	firstVersionSpec := updatev1alpha1.NodeVersionSpec{
		ImageReference:                "version-1",
		KubernetesComponentsReference: "ref-1",
	}

	firstNodeLookupKey := types.NamespacedName{Name: firstNodeName}
	secondNodeLookupKey := types.NamespacedName{Name: secondNodeName}
	nodeVersionLookupKey := types.NamespacedName{Name: nodeVersionResourceName}
	scalingGroupLookupKey := types.NamespacedName{Name: scalingGroupID}
	joiningPendingNodeLookupKey := types.NamespacedName{Name: secondNodeName}
	nodeMaintenanceLookupKey := types.NamespacedName{Name: firstNodeName}

	Context("When updating the cluster-wide node version", func() {
		testNodeVersionUpdate := func(newNodeVersionSpec updatev1alpha1.NodeVersionSpec) {
			By("creating a node version resource specifying the first node version")
			Expect(fakes.scalingGroupUpdater.SetScalingGroupImage(ctx, scalingGroupID, firstVersionSpec.ImageReference)).Should(Succeed())
			nodeVersion := &updatev1alpha1.NodeVersion{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "update.edgeless.systems/v1alpha1",
					Kind:       "NodeVersion",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: nodeVersionResourceName,
				},
				Spec: firstVersionSpec,
			}
			Expect(k8sClient.Create(ctx, nodeVersion)).Should(Succeed())

			By("creating a node resource using the first node image")
			fakes.nodeReplacer.setNodeImage(firstNodeName, firstVersionSpec.ImageReference)
			fakes.nodeReplacer.setScalingGroupID(firstNodeName, scalingGroupID)
			firstNode := &corev1.Node{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Node",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: firstNodeName,
					Labels: map[string]string{
						"custom-node-label": "custom-node-label-value",
					},
					Annotations: map[string]string{
						mainconstants.NodeKubernetesComponentsAnnotationKey: firstVersionSpec.KubernetesComponentsReference,
					},
				},
				Spec: corev1.NodeSpec{
					ProviderID: firstNodeName,
				},
			}
			Expect(k8sClient.Create(ctx, firstNode)).Should(Succeed())

			By("creating a scaling group resource using the first node image")
			Expect(fakes.scalingGroupUpdater.SetScalingGroupImage(ctx, scalingGroupID, firstVersionSpec.ImageReference)).Should(Succeed())
			scalingGroup := &updatev1alpha1.ScalingGroup{
				ObjectMeta: metav1.ObjectMeta{
					Name: scalingGroupID,
				},
				Spec: updatev1alpha1.ScalingGroupSpec{
					NodeVersion: nodeVersionResourceName,
					GroupID:     scalingGroupID,
					Autoscaling: true,
				},
			}
			Expect(k8sClient.Create(ctx, scalingGroup)).Should(Succeed())

			By("creating a cluster-autoscaler deployment")
			ctx := context.Background()
			autoscalerDeployment := &appsv1.Deployment{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "apps/v1",
					Kind:       "Deployment",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cluster-autoscaler",
					Namespace: "kube-system",
				},
				Spec: appsv1.DeploymentSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app.kubernetes.io/name": "cluster-autoscaler",
						},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"app.kubernetes.io/name": "cluster-autoscaler",
							},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{Image: "cluster-autoscaler", Name: "cluster-autoscaler"},
							},
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, autoscalerDeployment)).Should(Succeed())

			By("creating an autoscaling strategy")
			strategy := &updatev1alpha1.AutoscalingStrategy{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "update.edgeless.systems/v1alpha1",
					Kind:       "AutoscalingStrategy",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "autoscaling-strategy",
				},
				Spec: updatev1alpha1.AutoscalingStrategySpec{
					DeploymentName:      "cluster-autoscaler",
					DeploymentNamespace: "kube-system",
				},
			}
			Expect(k8sClient.Create(ctx, strategy)).Should(Succeed())

			By("checking that all nodes are up-to-date")
			Eventually(func() int {
				if err := k8sClient.Get(ctx, nodeVersionLookupKey, nodeVersion); err != nil {
					return 0
				}
				return len(nodeVersion.Status.UpToDate)
			}, timeout, interval).Should(Equal(1))

			By("updating the node version")
			fakes.nodeStateGetter.setNodeState(updatev1alpha1.NodeStateReady)
			fakes.nodeReplacer.setCreatedNode(secondNodeName, secondNodeName, nil)
			// Eventually the node version with the new NodeVersion spec.
			Eventually(func() error {
				if err := k8sClient.Get(ctx, nodeVersionLookupKey, nodeVersion); err != nil {
					return err
				}
				nodeVersion.Spec = newNodeVersionSpec
				return k8sClient.Update(ctx, nodeVersion)
			}, timeout, interval).Should(Succeed())

			By("checking that there is an outdated node in the status")
			Eventually(func() int {
				if err := k8sClient.Get(ctx, nodeVersionLookupKey, nodeVersion); err != nil {
					return 0
				}
				return len(nodeVersion.Status.Outdated)
			}, timeout, interval).Should(Equal(1), "outdated nodes should be 1")

			By("checking that the scaling group is up to date")
			Eventually(func() string {
				if err := k8sClient.Get(ctx, scalingGroupLookupKey, scalingGroup); err != nil {
					return ""
				}
				return scalingGroup.Status.ImageReference
			}, timeout, interval).Should(Equal(newNodeVersionSpec.ImageReference))

			By("checking that a pending node is created")
			pendingNode := &updatev1alpha1.PendingNode{}
			Eventually(func() error {
				return k8sClient.Get(ctx, joiningPendingNodeLookupKey, pendingNode)
			}, timeout, interval).Should(Succeed())
			Eventually(func() updatev1alpha1.CSPNodeState {
				_ = k8sClient.Get(ctx, joiningPendingNodeLookupKey, pendingNode)
				return pendingNode.Status.CSPNodeState
			}, timeout, interval).Should(Equal(updatev1alpha1.NodeStateReady))
			Eventually(func() int {
				if err := k8sClient.Get(ctx, nodeVersionLookupKey, nodeVersion); err != nil {
					return 0
				}
				return len(nodeVersion.Status.Pending)
			}, timeout, interval).Should(Equal(1))

			By("creating a new node resource using the image from the new node version")
			fakes.nodeReplacer.setNodeImage(secondNodeName, newNodeVersionSpec.ImageReference)
			fakes.nodeReplacer.setScalingGroupID(secondNodeName, scalingGroupID)
			secondNode := &corev1.Node{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Node",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: secondNodeName,
				},
				Spec: corev1.NodeSpec{
					ProviderID: secondNodeName,
				},
			}
			Expect(k8sClient.Create(ctx, secondNode)).Should(Succeed())

			By("marking the new node as AwaitingAnnotation")
			Eventually(func() int {
				err := k8sClient.Get(ctx, nodeVersionLookupKey, nodeVersion)
				if err != nil {
					return 0
				}
				return len(nodeVersion.Status.AwaitingAnnotation)
			}, timeout, interval).Should(Equal(1))
			// add a JoiningNode CR for the new node
			joiningNode := &updatev1alpha1.JoiningNode{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "update.edgeless.systems/v1alpha1",
					Kind:       "JoiningNode",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: secondNodeName,
				},
				Spec: updatev1alpha1.JoiningNodeSpec{
					Name:                secondNodeName,
					ComponentsReference: newNodeVersionSpec.KubernetesComponentsReference,
				},
			}
			Expect(k8sClient.Create(ctx, joiningNode)).Should(Succeed())
			Eventually(func() int {
				err := k8sClient.Get(ctx, nodeVersionLookupKey, nodeVersion)
				if err != nil {
					return 1
				}
				return len(nodeVersion.Status.AwaitingAnnotation)
			}, timeout, interval).Should(Equal(0))

			By("checking that the new node is properly annotated")
			Eventually(func() error {
				if err := k8sClient.Get(ctx, secondNodeLookupKey, secondNode); err != nil {
					return err
				}
				// check nodeImageAnnotation annotation
				if _, ok := secondNode.Annotations[nodeImageAnnotation]; !ok {
					return fmt.Errorf("node %s is missing %s annotation", secondNode.Name, nodeImageAnnotation)
				}
				// check mainconstants.NodeKubernetesComponentsAnnotationKey annotation
				if _, ok := secondNode.Annotations[mainconstants.NodeKubernetesComponentsAnnotationKey]; !ok {
					return fmt.Errorf("node %s is missing %s annotation", secondNode.Name, mainconstants.NodeKubernetesComponentsAnnotationKey)
				}
				return nil
			}, timeout, interval).Should(Succeed())

			By("checking that the nodes are paired as donor and heir")
			Eventually(func() map[string]string {
				if err := k8sClient.Get(ctx, firstNodeLookupKey, firstNode); err != nil {
					return nil
				}
				return firstNode.Annotations
			}, timeout, interval).Should(HaveKeyWithValue(heirAnnotation, secondNodeName))
			Eventually(func() map[string]string {
				if err := k8sClient.Get(ctx, secondNodeLookupKey, secondNode); err != nil {
					return nil
				}
				return secondNode.Annotations
			}, timeout, interval).Should(HaveKeyWithValue(donorAnnotation, firstNodeName))

			Eventually(func() error {
				if err := k8sClient.Get(ctx, nodeVersionLookupKey, nodeVersion); err != nil {
					return err
				}
				if len(nodeVersion.Status.Donors) != 1 {
					return fmt.Errorf("node version %s has %d donors, expected 1", nodeVersion.Name, len(nodeVersion.Status.Donors))
				}
				if len(nodeVersion.Status.Heirs) != 1 {
					return fmt.Errorf("node version %s has %d heirs, expected 1", nodeVersion.Name, len(nodeVersion.Status.Heirs))
				}
				return nil
			}, timeout, interval).Should(Succeed())
			Expect(k8sClient.Get(ctx, joiningPendingNodeLookupKey, pendingNode)).Should(Not(Succeed()))

			By("checking that node labels are copied to the heir")
			Eventually(func() map[string]string {
				if err := k8sClient.Get(ctx, firstNodeLookupKey, firstNode); err != nil {
					return nil
				}
				return firstNode.Labels
			}, timeout, interval).Should(HaveKeyWithValue("custom-node-label", "custom-node-label-value"))

			By("marking the new node as ready")
			Eventually(func() error {
				if err := k8sClient.Get(ctx, secondNodeLookupKey, secondNode); err != nil {
					return err
				}
				secondNode.Status.Conditions = []corev1.NodeCondition{
					{
						Type:   corev1.NodeReady,
						Status: corev1.ConditionTrue,
					},
				}
				return k8sClient.Status().Update(ctx, secondNode)
			}, timeout, interval).Should(Succeed())

			By("waiting for a NodeMaintenance resource to be created")
			nodeMaintenance := &nodemaintenancev1beta1.NodeMaintenance{}
			Eventually(func() error {
				return k8sClient.Get(ctx, nodeMaintenanceLookupKey, nodeMaintenance)
			}, timeout, interval).Should(Succeed())

			By("marking the NodeMaintenance as successful")
			fakes.nodeStateGetter.setNodeState(updatev1alpha1.NodeStateTerminated)
			Eventually(func() error {
				if err := k8sClient.Get(ctx, nodeMaintenanceLookupKey, nodeMaintenance); err != nil {
					return err
				}
				nodeMaintenance.Status.Phase = nodemaintenancev1beta1.MaintenanceSucceeded
				return k8sClient.Status().Update(ctx, nodeMaintenance)
			}, timeout, interval).Should(Succeed())

			By("checking that the outdated node is removed")
			Eventually(func() error {
				return k8sClient.Get(ctx, firstNodeLookupKey, firstNode)
			}, timeout, interval).Should(Not(Succeed()))

			By("checking that all nodes are up-to-date")
			Eventually(func() int {
				err := k8sClient.Get(ctx, nodeVersionLookupKey, nodeVersion)
				if err != nil {
					return 0
				}
				return len(nodeVersion.Status.UpToDate)
			}, timeout, interval).Should(Equal(1))

			By("cleaning up all resources")
			Expect(k8sClient.Delete(ctx, nodeVersion)).Should(Succeed())
			Expect(k8sClient.Delete(ctx, scalingGroup)).Should(Succeed())
			Expect(k8sClient.Delete(ctx, autoscalerDeployment)).Should(Succeed())
			Expect(k8sClient.Delete(ctx, strategy)).Should(Succeed())
			Expect(k8sClient.Delete(ctx, secondNode)).Should(Succeed())
		}
		When("Updating the image reference", func() {
			It("Should update every node in the cluster", func() {
				testNodeVersionUpdate(updatev1alpha1.NodeVersionSpec{
					ImageReference:                "version-2",
					KubernetesComponentsReference: "ref-1",
				})
			})
		})
		When("Updating the Kubernetes components reference", func() {
			It("Should update every node in the cluster", func() {
				testNodeVersionUpdate(updatev1alpha1.NodeVersionSpec{
					ImageReference:                "version-1",
					KubernetesComponentsReference: "ref-2",
				})
			})
		})
	})
})
