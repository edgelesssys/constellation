//go:build integration

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package controllers

import (
	"context"
	"time"

	//revive:disable:dot-imports
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	//revive:enable:dot-imports

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	updatev1alpha1 "github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/v2/api/v1alpha1"
)

var _ = Describe("ScalingGroup controller", func() {
	AfterEach(func() {
		Eventually(func() error {
			return resetEnv()
		}, 30*time.Second, 1*time.Second).Should(Succeed())
	})

	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		nodeVersionName  = "node-version"
		scalingGroupName = "test-group"

		timeout  = time.Second * 20
		interval = time.Millisecond * 250
	)

	nodeVersionLookupKey := types.NamespacedName{Name: nodeVersionName}

	Context("When changing a node image resource spec", func() {
		It("Should update corresponding scaling group images", func() {
			By("creating a node image resource")
			ctx := context.Background()
			nodeVersion := &updatev1alpha1.NodeVersion{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "update.edgeless.systems/v1alpha1",
					Kind:       "NodeVersion",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: nodeVersionName,
				},
				Spec: updatev1alpha1.NodeVersionSpec{
					ImageReference: "image-1",
				},
			}
			Expect(k8sClient.Create(ctx, nodeVersion)).Should(Succeed())
			createdNodeVersion := &updatev1alpha1.NodeVersion{}
			Eventually(func() error {
				return k8sClient.Get(ctx, nodeVersionLookupKey, createdNodeVersion)
			}, timeout, interval).Should(Succeed())
			Expect(createdNodeVersion.Spec.ImageReference).Should(Equal("image-1"))

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
					NodeVersion: nodeVersionName,
					GroupID:     "group-id",
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

			By("updating the node image")
			Eventually(func() error {
				if err := k8sClient.Get(ctx, nodeVersionLookupKey, nodeVersion); err != nil {
					return err
				}
				nodeVersion.Spec.ImageReference = "image-2"
				return k8sClient.Update(ctx, nodeVersion)
			}, timeout, interval).Should(Succeed())

			By("checking the scaling group eventually uses the latest image")
			Eventually(func() string {
				image, _ := fakes.scalingGroupUpdater.GetScalingGroupImage(ctx, "group-id")
				return image
			}, timeout, interval).Should(Equal("image-2"))

			By("checking the scaling group status shows the latest image")
			Eventually(func() (string, error) {
				err := k8sClient.Get(ctx, scalingGroupLookupKey, createdScalingGroup)
				if err != nil {
					return "", err
				}
				return createdScalingGroup.Status.ImageReference, nil
			}, timeout, interval).Should(Equal("image-2"))

			By("cleaning up all resources")
			Expect(k8sClient.Delete(ctx, createdNodeVersion)).Should(Succeed())
			Expect(k8sClient.Delete(ctx, scalingGroup)).Should(Succeed())
		})
	})
})
