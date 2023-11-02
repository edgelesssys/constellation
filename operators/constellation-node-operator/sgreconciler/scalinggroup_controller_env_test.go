//go:build integration

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package sgreconciler

import (
	"context"
	"time"

	//revive:disable:dot-imports
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	//revive:enable:dot-imports

	cspapi "github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/v2/internal/cloud/api"
	"k8s.io/apimachinery/pkg/types"

	updatev1alpha1 "github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/v2/api/v1alpha1"
)

var _ = Describe("ExternalScalingGroup controller", func() {
	AfterEach(func() {
		Eventually(func() error {
			return resetEnv()
		}, 30*time.Second, 1*time.Second).Should(Succeed())
	})

	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		nodeVersionName      = "node-version"
		scalingGroupName     = "test-group"
		nodeGroupName        = "test-node-group"
		groupID              = "test-group-id"
		autoscalingGroupName = "test-autoscaling-group-name"

		timeout  = time.Second * 20
		interval = time.Millisecond * 250
	)

	Context("When creating a scaling group externally", func() {
		It("Should create a corresponding scaling group resource", func() {
			By("creating an external scaling group")
			ctx := context.Background()
			fakes.scalingGroupDiscoverer.set([]cspapi.ScalingGroup{
				{
					Name:                 scalingGroupName,
					NodeGroupName:        nodeGroupName,
					GroupID:              groupID,
					AutoscalingGroupName: autoscalingGroupName,
					Role:                 updatev1alpha1.ControlPlaneRole,
				},
			})

			By("checking that the scaling group resource was created")
			triggerReconcile()
			createdScalingGroup := &updatev1alpha1.ScalingGroup{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{Name: scalingGroupName}, createdScalingGroup)
			}, timeout, interval).Should(Succeed())
			Expect(createdScalingGroup.Spec.GroupID).Should(Equal(groupID))
			Expect(createdScalingGroup.Spec.AutoscalerGroupName).Should(Equal(autoscalingGroupName))
			Expect(createdScalingGroup.Spec.NodeGroupName).Should(Equal(nodeGroupName))
			Expect(createdScalingGroup.Spec.Role).Should(Equal(updatev1alpha1.ControlPlaneRole))
		})

		It("Should update if external scaling groups are added/removed", func() {
			By("changing the external scaling groups")
			fakes.scalingGroupDiscoverer.set([]cspapi.ScalingGroup{
				{
					Name:                 "other-scaling-group",
					NodeGroupName:        "other-node-group",
					GroupID:              "other-group-id",
					AutoscalingGroupName: "other-autoscaling-group-name",
					Role:                 updatev1alpha1.WorkerRole,
				},
			})

			By("checking that the scaling group resource was created")
			triggerReconcile()
			createdScalingGroup := &updatev1alpha1.ScalingGroup{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{Name: "other-scaling-group"}, createdScalingGroup)
			}, timeout, interval).Should(Succeed())
			Expect(createdScalingGroup.Spec.GroupID).Should(Equal("other-group-id"))
			Expect(createdScalingGroup.Spec.AutoscalerGroupName).Should(Equal("other-autoscaling-group-name"))
			Expect(createdScalingGroup.Spec.NodeGroupName).Should(Equal("other-node-group"))
			Expect(createdScalingGroup.Spec.Role).Should(Equal(updatev1alpha1.WorkerRole))

			By("checking that the old scaling group resource was deleted")
			deletedScalingGroup := &updatev1alpha1.ScalingGroup{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{Name: scalingGroupName}, deletedScalingGroup)
			}, timeout, interval).ShouldNot(Succeed())
		})
	})
})
