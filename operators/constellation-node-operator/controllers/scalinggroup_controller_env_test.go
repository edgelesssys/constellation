//go:build integration
// +build integration

package controllers

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	updatev1alpha1 "github.com/edgelesssys/constellation/operators/constellation-node-operator/api/v1alpha1"
)

var _ = Describe("ScalingGroup controller", func() {
	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		nodeImageName    = "node-image"
		scalingGroupName = "test-group"

		timeout  = time.Second * 20
		duration = time.Second * 2
		interval = time.Millisecond * 250
	)

	nodeImageLookupKey := types.NamespacedName{Name: nodeImageName}

	Context("When changing a node image resource spec", func() {
		It("Should update corresponding scaling group images", func() {
			By("creating a node image resource")
			ctx := context.Background()
			nodeImage := &updatev1alpha1.NodeImage{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "update.edgeless.systems/v1alpha1",
					Kind:       "NodeImage",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: nodeImageName,
				},
				Spec: updatev1alpha1.NodeImageSpec{
					ImageReference: "image-1",
				},
			}
			Expect(k8sClient.Create(ctx, nodeImage)).Should(Succeed())
			createdNodeImage := &updatev1alpha1.NodeImage{}
			Eventually(func() error {
				return k8sClient.Get(ctx, nodeImageLookupKey, createdNodeImage)
			}, timeout, interval).Should(Succeed())
			Expect(createdNodeImage.Spec.ImageReference).Should(Equal("image-1"))

			By("creating a scaling group")
			scalingGroup := &updatev1alpha1.ScalingGroup{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "update.edgeless.systems/v1alpha1",
					Kind:       "ScalingGroup",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: scalingGroupName,
				},
				Spec: updatev1alpha1.ScalingGroupSpec{
					NodeImage: nodeImageName,
					GroupID:   "group-id",
				},
			}
			Expect(k8sClient.Create(ctx, scalingGroup)).Should(Succeed())
			scalingGroupLookupKey := types.NamespacedName{Name: scalingGroupName}
			createdScalingGroup := &updatev1alpha1.ScalingGroup{}
			Eventually(func() error {
				return k8sClient.Get(ctx, scalingGroupLookupKey, createdScalingGroup)
			}, timeout, interval).Should(Succeed())

			By("checking the scaling group status shows the correct image")
			Eventually(func() string {
				image, _ := fakes.scalingGroupUpdater.GetScalingGroupImage(ctx, "group-id")
				return image
			}, timeout, interval).Should(Equal("image-1"))
			Eventually(func() string {
				err := k8sClient.Get(ctx, scalingGroupLookupKey, createdScalingGroup)
				if err != nil {
					return ""
				}
				return createdScalingGroup.Status.ImageReference
			}, timeout, interval).Should(Equal("image-1"))
			Consistently(func() (string, error) {
				err := k8sClient.Get(ctx, scalingGroupLookupKey, createdScalingGroup)
				if err != nil {
					return "", err
				}
				return createdScalingGroup.Status.ImageReference, nil
			}, duration, interval).Should(Equal("image-1"))

			By("updating the node image")
			Expect(k8sClient.Get(ctx, nodeImageLookupKey, nodeImage)).Should(Succeed())
			nodeImage.Spec.ImageReference = "image-2"
			Expect(k8sClient.Update(ctx, nodeImage)).Should(Succeed())

			By("checking the scaling group eventually uses the latest image")
			Eventually(func() string {
				image, _ := fakes.scalingGroupUpdater.GetScalingGroupImage(ctx, "group-id")
				return image
			}, timeout, interval).Should(Equal("image-2"))

			By("checking the scaling group status shows the latest image")
			Consistently(func() (string, error) {
				err := k8sClient.Get(ctx, scalingGroupLookupKey, createdScalingGroup)
				if err != nil {
					return "", err
				}
				return createdScalingGroup.Status.ImageReference, nil
			}, duration, interval).Should(Equal("image-2"))

			By("cleaning up all resources")
			Expect(k8sClient.Delete(ctx, createdNodeImage)).Should(Succeed())
			Expect(k8sClient.Delete(ctx, scalingGroup)).Should(Succeed())
		})
	})
})
