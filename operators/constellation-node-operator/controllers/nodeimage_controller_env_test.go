package controllers

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	updatev1alpha1 "github.com/edgelesssys/constellation/operators/constellation-node-operator/api/v1alpha1"
	nodemaintenancev1beta1 "github.com/medik8s/node-maintenance-operator/api/v1beta1"
)

var _ = Describe("NodeImage controller", func() {
	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		nodeImageResourceName = "nodeimage"
		firstNodeName         = "node-1"
		secondNodeName        = "node-2"
		firstImage            = "image-1"
		secondImage           = "image-2"
		scalingGroupID        = "scaling-group"

		timeout  = time.Second * 10
		duration = time.Second * 2
		interval = time.Millisecond * 250
	)

	firstNodeLookupKey := types.NamespacedName{Name: firstNodeName}
	secondNodeLookupKey := types.NamespacedName{Name: secondNodeName}
	nodeImageLookupKey := types.NamespacedName{Name: nodeImageResourceName}
	scalingGroupLookupKey := types.NamespacedName{Name: scalingGroupID}
	joiningPendingNodeLookupKey := types.NamespacedName{Name: secondNodeName}
	nodeMaintenanceLookupKey := types.NamespacedName{Name: firstNodeName}

	Context("When updating the cluster-wide node image", func() {
		It("Should update every node in the cluster", func() {
			By("creating a node image resource specifying the first node image")
			Expect(fakes.scalingGroupUpdater.SetScalingGroupImage(ctx, scalingGroupID, firstImage)).Should(Succeed())
			nodeImage := &updatev1alpha1.NodeImage{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "update.edgeless.systems/v1alpha1",
					Kind:       "NodeImage",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: nodeImageResourceName,
				},
				Spec: updatev1alpha1.NodeImageSpec{ImageReference: firstImage},
			}
			Expect(k8sClient.Create(ctx, nodeImage)).Should(Succeed())

			By("creating a node resource using the first node image")
			fakes.nodeReplacer.setNodeImage(firstNodeName, firstImage)
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
				},
				Spec: corev1.NodeSpec{
					ProviderID: firstNodeName,
				},
			}
			Expect(k8sClient.Create(ctx, firstNode)).Should(Succeed())

			By("creating a scaling group resource using the first node image")
			Expect(fakes.scalingGroupUpdater.SetScalingGroupImage(ctx, scalingGroupID, firstImage)).Should(Succeed())
			scalingGroup := &updatev1alpha1.ScalingGroup{
				ObjectMeta: metav1.ObjectMeta{
					Name: scalingGroupID,
				},
				Spec: updatev1alpha1.ScalingGroupSpec{
					NodeImage:   nodeImageResourceName,
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
				if err := k8sClient.Get(ctx, nodeImageLookupKey, nodeImage); err != nil {
					return 0
				}
				return len(nodeImage.Status.UpToDate)
			}, timeout, interval).Should(Equal(1))

			By("updating the node image to the second image")
			fakes.nodeStateGetter.setNodeState(updatev1alpha1.NodeStateReady)
			fakes.nodeReplacer.setCreatedNode(secondNodeName, secondNodeName, nil)
			nodeImage.Spec.ImageReference = secondImage
			Expect(k8sClient.Update(ctx, nodeImage)).Should(Succeed())

			By("checking that there is an outdated node in the status")
			Eventually(func() int {
				if err := k8sClient.Get(ctx, nodeImageLookupKey, nodeImage); err != nil {
					return 0
				}
				return len(nodeImage.Status.Outdated)
			}, timeout, interval).Should(Equal(1))

			By("checking that the scaling group is up to date")
			Eventually(func() string {
				if err := k8sClient.Get(ctx, scalingGroupLookupKey, scalingGroup); err != nil {
					return ""
				}
				return scalingGroup.Status.ImageReference
			}, timeout, interval).Should(Equal(secondImage))

			By("checking that a pending node is created")
			pendingNode := &updatev1alpha1.PendingNode{}
			Eventually(func() error {
				return k8sClient.Get(ctx, joiningPendingNodeLookupKey, pendingNode)
			}).Should(Succeed())
			Expect(pendingNode.Status.CSPNodeState).Should(Equal(updatev1alpha1.NodeStateReady))
			Eventually(func() int {
				if err := k8sClient.Get(ctx, nodeImageLookupKey, nodeImage); err != nil {
					return 0
				}
				return len(nodeImage.Status.Pending)
			}, timeout, interval).Should(Equal(1))

			By("creating a new node resource using the second node image")
			fakes.nodeReplacer.setNodeImage(secondNodeName, secondImage)
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

			By("checking that the new node is properly annotated")
			Eventually(func() map[string]string {
				if err := k8sClient.Get(ctx, secondNodeLookupKey, secondNode); err != nil {
					return nil
				}
				return secondNode.Annotations
			}, timeout, interval).Should(HaveKeyWithValue(scalingGroupAnnotation, scalingGroupID))
			Expect(secondNode.Annotations).Should(HaveKeyWithValue(nodeImageAnnotation, secondImage))

			By("checking that the nodes are paired as donor and heir")
			Eventually(func() map[string]string {
				if err := k8sClient.Get(ctx, firstNodeLookupKey, firstNode); err != nil {
					return nil
				}
				return firstNode.Annotations
			}, timeout, interval).Should(HaveKeyWithValue(heirAnnotation, secondNodeName))
			Expect(k8sClient.Get(ctx, secondNodeLookupKey, secondNode)).Should(Succeed())
			Expect(secondNode.Annotations).Should(HaveKeyWithValue(donorAnnotation, firstNodeName))
			Expect(k8sClient.Get(ctx, nodeImageLookupKey, nodeImage)).Should(Succeed())
			Expect(nodeImage.Status.Donors).Should(HaveLen(1))
			Expect(nodeImage.Status.Heirs).Should(HaveLen(1))
			Expect(k8sClient.Get(ctx, joiningPendingNodeLookupKey, pendingNode)).Should(Not(Succeed()))

			By("checking that node labels are copied to the heir")
			Eventually(func() map[string]string {
				if err := k8sClient.Get(ctx, firstNodeLookupKey, firstNode); err != nil {
					return nil
				}
				return firstNode.Labels
			}, timeout, interval).Should(HaveKeyWithValue("custom-node-label", "custom-node-label-value"))

			By("marking the new node as ready")
			secondNode.Status.Conditions = []corev1.NodeCondition{
				{
					Type:   corev1.NodeReady,
					Status: corev1.ConditionTrue,
				},
			}
			Expect(k8sClient.Status().Update(ctx, secondNode)).Should(Succeed())

			By("waiting for a NodeMaintenance resource to be created")
			nodeMaintenance := &nodemaintenancev1beta1.NodeMaintenance{}
			Eventually(func() error {
				return k8sClient.Get(ctx, nodeMaintenanceLookupKey, nodeMaintenance)
			}, timeout, interval).Should(Succeed())

			By("marking the NodeMaintenance as successful")
			fakes.nodeStateGetter.setNodeState(updatev1alpha1.NodeStateTerminated)
			nodeMaintenance.Status.Phase = nodemaintenancev1beta1.MaintenanceSucceeded
			Expect(k8sClient.Status().Update(ctx, nodeMaintenance)).Should(Succeed())
			Eventually(func() error {
				return k8sClient.Get(ctx, nodeMaintenanceLookupKey, nodeMaintenance)
			}, timeout, interval).Should(Succeed())

			By("checking that the outdated node is removed")
			Eventually(func() error {
				return k8sClient.Get(ctx, firstNodeLookupKey, firstNode)
			}, timeout, interval).Should(Not(Succeed()))

			By("checking that all nodes are up-to-date")
			Eventually(func() int {
				err := k8sClient.Get(ctx, nodeImageLookupKey, nodeImage)
				if err != nil {
					return 0
				}
				return len(nodeImage.Status.UpToDate)
			}, timeout, interval).Should(Equal(1))

			By("cleaning up all resources")
			Expect(k8sClient.Delete(ctx, nodeImage)).Should(Succeed())
			Expect(k8sClient.Delete(ctx, scalingGroup)).Should(Succeed())
			Expect(k8sClient.Delete(ctx, autoscalerDeployment)).Should(Succeed())
			Expect(k8sClient.Delete(ctx, strategy)).Should(Succeed())
			Expect(k8sClient.Delete(ctx, secondNode)).Should(Succeed())
		})
	})
})
