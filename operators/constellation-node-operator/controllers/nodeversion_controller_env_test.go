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
	"github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/internal/constants"
)

var _ = Describe("NodeVersion controller", func() {
	AfterEach(func() {
		Eventually(func() error {
			return resetEnv()
		}, 30*time.Second, 1*time.Second).Should(Succeed())
	})

	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		nodeVersionResourceName    = "nodeversion"
		firstWorkerNodeName        = "worker-node-1"
		secondWorkerNodeName       = "worker-node-2"
		thirdWorkerNodeName        = "worker-node-3"
		firstControlPlaneNodeName  = "control-plane-node-1"
		secondControlPlaneNodeName = "control-plane-node-2"
		thirdControlPlaneNodeName  = "control-plane-node-3"
		scalingGroupIDWorker       = "scaling-group-worker"
		scalingGroupIDControlPlane = "scaling-group-control-plane"

		timeout  = time.Second * 20
		duration = time.Second * 2
		interval = time.Millisecond * 250
	)
	firstVersionSpec := updatev1alpha1.NodeVersionSpec{
		ImageReference:                "version-1",
		KubernetesComponentsReference: "ref-1",
	}

	firstNodeLookupKeyWorker := types.NamespacedName{Name: firstWorkerNodeName}
	secondNodeLookupKeyWorker := types.NamespacedName{Name: secondWorkerNodeName}
	firstNodeLookupKeyControlPlane := types.NamespacedName{Name: firstControlPlaneNodeName}
	secondNodeLookupKeyControlPlane := types.NamespacedName{Name: secondControlPlaneNodeName}
	nodeVersionLookupKey := types.NamespacedName{Name: nodeVersionResourceName}
	scalingGroupLookupKeyWorker := types.NamespacedName{Name: scalingGroupIDWorker}
	scalingGroupLookupKeyControlPlane := types.NamespacedName{Name: scalingGroupIDControlPlane}

	Context("When updating the cluster-wide node version", func() {
		testNodeVersionUpdate := func(newNodeVersionSpec updatev1alpha1.NodeVersionSpec, workersFirst bool) {
			By("creating a node version resource specifying the first node version")
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

			By("creating a control plane node resource using the first node image")
			fakes.nodeReplacer.setNodeImage(firstControlPlaneNodeName, firstVersionSpec.ImageReference)
			fakes.nodeReplacer.setScalingGroupID(firstControlPlaneNodeName, scalingGroupIDControlPlane)
			firstControlPlaneNode := &corev1.Node{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Node",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: firstControlPlaneNodeName,
					Labels: map[string]string{
						"custom-node-label":             "custom-node-label-value",
						constants.ControlPlaneRoleLabel: "",
					},
					Annotations: map[string]string{
						mainconstants.NodeKubernetesComponentsAnnotationKey: firstVersionSpec.KubernetesComponentsReference,
					},
				},
				Spec: corev1.NodeSpec{
					ProviderID: firstControlPlaneNodeName,
				},
				Status: corev1.NodeStatus{
					Addresses: []corev1.NodeAddress{
						{
							Type:    corev1.NodeInternalIP,
							Address: "192.0.2.1",
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, firstControlPlaneNode)).Should(Succeed())

			By("creating a worker node resource using the first node image")
			fakes.nodeReplacer.setNodeImage(firstWorkerNodeName, firstVersionSpec.ImageReference)
			fakes.nodeReplacer.setScalingGroupID(firstWorkerNodeName, scalingGroupIDWorker)
			firstWorkerNode := &corev1.Node{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Node",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: firstWorkerNodeName,
					Labels: map[string]string{
						"custom-node-label": "custom-node-label-value",
					},
					Annotations: map[string]string{
						mainconstants.NodeKubernetesComponentsAnnotationKey: firstVersionSpec.KubernetesComponentsReference,
					},
				},
				Spec: corev1.NodeSpec{
					ProviderID: firstWorkerNodeName,
				},
			}
			Expect(k8sClient.Create(ctx, firstWorkerNode)).Should(Succeed())

			By("creating worker scaling group resource using the first node image")
			Expect(fakes.scalingGroupUpdater.SetScalingGroupImage(ctx, scalingGroupIDWorker, firstVersionSpec.ImageReference)).Should(Succeed())
			scalingGroupWorker := &updatev1alpha1.ScalingGroup{
				ObjectMeta: metav1.ObjectMeta{
					Name: scalingGroupIDWorker,
				},
				Spec: updatev1alpha1.ScalingGroupSpec{
					NodeVersion: nodeVersionResourceName,
					GroupID:     scalingGroupIDWorker,
					Autoscaling: true,
					Role:        updatev1alpha1.WorkerRole,
				},
			}
			Expect(k8sClient.Create(ctx, scalingGroupWorker)).Should(Succeed())

			By("creating control plane scaling group resource using the first node image")
			Expect(fakes.scalingGroupUpdater.SetScalingGroupImage(ctx, scalingGroupIDControlPlane, firstVersionSpec.ImageReference)).Should(Succeed())
			scalingGroupControlPlane := &updatev1alpha1.ScalingGroup{
				ObjectMeta: metav1.ObjectMeta{
					Name: scalingGroupIDControlPlane,
				},
				Spec: updatev1alpha1.ScalingGroupSpec{
					NodeVersion: nodeVersionResourceName,
					GroupID:     scalingGroupIDControlPlane,
					Autoscaling: true,
					Role:        updatev1alpha1.ControlPlaneRole,
				},
			}
			Expect(k8sClient.Create(ctx, scalingGroupControlPlane)).Should(Succeed())

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
			}, timeout, interval).Should(Equal(2))

			By("updating the node version")
			fakes.nodeReplacer.addCreatedNode(scalingGroupIDControlPlane, secondControlPlaneNodeName, secondControlPlaneNodeName, nil)
			fakes.nodeReplacer.addCreatedNode(scalingGroupIDWorker, secondWorkerNodeName, secondWorkerNodeName, nil)
			fakes.nodeStateGetter.setNodeState(secondControlPlaneNodeName, updatev1alpha1.NodeStateReady)
			fakes.nodeStateGetter.setNodeState(secondWorkerNodeName, updatev1alpha1.NodeStateReady)
			// When the pending node CR that the nodeversion loop creates is not detected on the second iteration e.g., because of delay in the KubeAPI
			// it creates a second node via the cloud provider api. This is fine because the pending node/mint node is not matched if it's not needed
			// and the nodeversion loop will clean up the mint node.
			fakes.nodeStateGetter.setNodeState(thirdControlPlaneNodeName, updatev1alpha1.NodeStateCreating)
			fakes.nodeStateGetter.setNodeState(thirdWorkerNodeName, updatev1alpha1.NodeStateCreating)
			fakes.nodeReplacer.addCreatedNode(scalingGroupIDControlPlane, thirdControlPlaneNodeName, thirdControlPlaneNodeName, nil)
			fakes.nodeReplacer.addCreatedNode(scalingGroupIDWorker, thirdWorkerNodeName, thirdWorkerNodeName, nil)
			// Eventually the node version with the new NodeVersion spec.
			Eventually(func() error {
				if err := k8sClient.Get(ctx, nodeVersionLookupKey, nodeVersion); err != nil {
					return err
				}
				nodeVersion.Spec = newNodeVersionSpec
				return k8sClient.Update(ctx, nodeVersion)
			}, timeout, interval).Should(Succeed())

			By("checking that there are 2 outdated node in the status")
			Eventually(func() int {
				if err := k8sClient.Get(ctx, nodeVersionLookupKey, nodeVersion); err != nil {
					return 0
				}
				return len(nodeVersion.Status.Outdated)
			}, timeout, interval).Should(Equal(2), "outdated nodes should be 2")

			By("checking that the control plane scaling group is up to date")
			Eventually(func() string {
				if err := k8sClient.Get(ctx, scalingGroupLookupKeyControlPlane, scalingGroupControlPlane); err != nil {
					return ""
				}
				return scalingGroupControlPlane.Status.ImageReference
			}, timeout, interval).Should(Equal(newNodeVersionSpec.ImageReference))

			By("checking that the worker scaling group is up to date")
			Eventually(func() string {
				if err := k8sClient.Get(ctx, scalingGroupLookupKeyWorker, scalingGroupWorker); err != nil {
					return ""
				}
				return scalingGroupWorker.Status.ImageReference
			}, timeout, interval).Should(Equal(newNodeVersionSpec.ImageReference))

			replaceNodeTest := func(firstNodeLookupKey, secondNodeLookupKey types.NamespacedName, firstNodeName, secondNodeName, scalingGroupID string, firstNode *corev1.Node, expectPendingNode bool) *corev1.Node {
				By("checking that the pending node is created")
				pendingNode := &updatev1alpha1.PendingNode{}
				// If we try to upgrade the worker nodes first, we expect here that the worker node is not created
				if !expectPendingNode {
					Consistently(func() bool {
						if err := k8sClient.Get(ctx, secondNodeLookupKey, pendingNode); err != nil {
							return false
						}
						return true
					}, duration, interval).Should(BeFalse())
					return nil
				}
				Eventually(func() bool {
					if err := k8sClient.Get(ctx, secondNodeLookupKey, pendingNode); err != nil {
						return false
					}
					return true
				}, timeout, interval).Should(BeTrue())

				Eventually(func() updatev1alpha1.CSPNodeState {
					_ = k8sClient.Get(ctx, secondNodeLookupKey, pendingNode)
					return pendingNode.Status.CSPNodeState
				}, timeout, interval).Should(Equal(updatev1alpha1.NodeStateReady))
				Eventually(func() bool {
					if err := k8sClient.Get(ctx, nodeVersionLookupKey, nodeVersion); err != nil {
						return false
					}
					return len(nodeVersion.Status.Pending) >= 1
				}, timeout, interval).Should(BeTrue())

				By("creating a new node resource using the image from the new node version")
				fakes.nodeReplacer.setNodeImage(secondNodeName, newNodeVersionSpec.ImageReference)
				fakes.nodeReplacer.setScalingGroupID(secondNodeName, scalingGroupID)
				var labels map[string]string
				if _, ok := firstNode.Labels[constants.ControlPlaneRoleLabel]; ok {
					labels = map[string]string{
						constants.ControlPlaneRoleLabel: "",
					}
				}
				secondNode := &corev1.Node{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
						Kind:       "Node",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:   secondNodeName,
						Labels: labels,
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
				Expect(k8sClient.Get(ctx, secondNodeLookupKey, pendingNode)).Should(Not(Succeed()))

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
					return k8sClient.Get(ctx, firstNodeLookupKey, nodeMaintenance)
				}, timeout, interval).Should(Succeed())

				By("marking the NodeMaintenance as successful")
				fakes.nodeStateGetter.setNodeState(firstNodeName, updatev1alpha1.NodeStateTerminated)
				Eventually(func() error {
					if err := k8sClient.Get(ctx, firstNodeLookupKey, nodeMaintenance); err != nil {
						return err
					}
					nodeMaintenance.Status.Phase = nodemaintenancev1beta1.MaintenanceSucceeded
					return k8sClient.Status().Update(ctx, nodeMaintenance)
				}, timeout, interval).Should(Succeed())

				By("checking that the outdated node is removed")
				Eventually(func() error {
					return k8sClient.Get(ctx, firstNodeLookupKey, firstNode)
				}, timeout, interval).Should(Not(Succeed()))

				return secondNode
			}

			var createdControlPlane *corev1.Node
			var createdWorkerNode *corev1.Node
			if workersFirst {
				_ = replaceNodeTest(firstNodeLookupKeyWorker, secondNodeLookupKeyWorker, firstWorkerNodeName, secondWorkerNodeName, scalingGroupIDWorker, firstWorkerNode, false)
				createdControlPlane = replaceNodeTest(firstNodeLookupKeyControlPlane, secondNodeLookupKeyControlPlane, firstControlPlaneNodeName, secondControlPlaneNodeName, scalingGroupIDControlPlane, firstControlPlaneNode, true)
				createdWorkerNode = replaceNodeTest(firstNodeLookupKeyWorker, secondNodeLookupKeyWorker, firstWorkerNodeName, secondWorkerNodeName, scalingGroupIDWorker, firstWorkerNode, true)
			} else {
				createdControlPlane = replaceNodeTest(firstNodeLookupKeyControlPlane, secondNodeLookupKeyControlPlane, firstControlPlaneNodeName, secondControlPlaneNodeName, scalingGroupIDControlPlane, firstControlPlaneNode, true)
				createdWorkerNode = replaceNodeTest(firstNodeLookupKeyWorker, secondNodeLookupKeyWorker, firstWorkerNodeName, secondWorkerNodeName, scalingGroupIDWorker, firstWorkerNode, true)
			}

			By("checking that all nodes are up-to-date")
			Eventually(func() int {
				err := k8sClient.Get(ctx, nodeVersionLookupKey, nodeVersion)
				if err != nil {
					return 0
				}
				return len(nodeVersion.Status.UpToDate)
			}, timeout, interval).Should(Equal(2))

			By("cleaning up all resources")
			Expect(k8sClient.Delete(ctx, nodeVersion)).Should(Succeed())
			Expect(k8sClient.Delete(ctx, scalingGroupWorker)).Should(Succeed())
			Expect(k8sClient.Delete(ctx, autoscalerDeployment)).Should(Succeed())
			Expect(k8sClient.Delete(ctx, strategy)).Should(Succeed())
			Expect(k8sClient.Delete(ctx, createdControlPlane)).Should(Succeed())
			Expect(k8sClient.Delete(ctx, createdWorkerNode)).Should(Succeed())
		}
		When("Updating the image reference", func() {
			It("Should update every node in the cluster", func() {
				testNodeVersionUpdate(
					updatev1alpha1.NodeVersionSpec{
						ImageReference:                "version-2",
						KubernetesComponentsReference: "ref-1",
					},
					false,
				)
			})
		})
		When("Updating the Kubernetes components reference", func() {
			It("Should update every node in the cluster", func() {
				testNodeVersionUpdate(
					updatev1alpha1.NodeVersionSpec{
						ImageReference:                "version-1",
						KubernetesComponentsReference: "ref-2",
					},
					false,
				)
			})
		})
		When("Updating the Kubernetes components reference and wanting to upgrade the worker nodes first", func() {
			It("should fail to update the worker nodes before the control plane nodes", func() {
				testNodeVersionUpdate(
					updatev1alpha1.NodeVersionSpec{
						ImageReference:                "version-1",
						KubernetesComponentsReference: "ref-2",
						MaxNodeBudget:                 2,
					},
					true,
				)
			})
		})
	})
})
