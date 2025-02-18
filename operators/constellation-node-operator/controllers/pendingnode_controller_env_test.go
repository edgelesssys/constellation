//go:build integration

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package controllers

import (
	"context"
	"net/http"
	"time"

	//revive:disable:dot-imports
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	//revive:enable:dot-imports

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	updatev1alpha1 "github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/api/v1alpha1"
)

var _ = Describe("PendingNode controller", func() {
	AfterEach(func() {
		Eventually(func() error {
			return resetEnv()
		}, 30*time.Second, 1*time.Second).Should(Succeed())
	})

	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		pendingNodeName = "pending-node"

		timeout  = time.Second * 20
		duration = time.Second * 2
		interval = time.Millisecond * 250
	)

	beGone := func() OmegaMatcher {
		return MatchError(&errors.StatusError{
			ErrStatus: metav1.Status{
				Status:  "Failure",
				Message: `pendingnodes.update.edgeless.systems "pending-node" not found`,
				Reason:  "NotFound",
				Details: &metav1.StatusDetails{
					Name:  pendingNodeName,
					Group: "update.edgeless.systems",
					Kind:  "pendingnodes",
				},
				Code: http.StatusNotFound,
			},
		})
	}

	pendingNodeLookupKey := types.NamespacedName{Name: pendingNodeName}

	Context("When creating pending node with goal join", func() {
		It("Should terminate the node after failing to join by the deadline", func() {
			By("setting the CSP node state to creating")
			fakes.nodeStateGetter.setNodeState(pendingNodeName, updatev1alpha1.NodeStateCreating)

			By("creating a pending node resource")
			ctx := context.Background()
			pendingNode := &updatev1alpha1.PendingNode{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "update.edgeless.systems/v1alpha1",
					Kind:       "PendingNode",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: pendingNodeName,
				},
				Spec: updatev1alpha1.PendingNodeSpec{
					ProviderID:     pendingNodeName,
					ScalingGroupID: "scaling-group-id",
					NodeName:       "test-node",
					Goal:           updatev1alpha1.NodeGoalJoin,
					// create without deadline first
				},
			}
			Expect(k8sClient.Create(ctx, pendingNode)).Should(Succeed())
			createdPendingNode := &updatev1alpha1.PendingNode{}
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, pendingNodeLookupKey, createdPendingNode); err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())
			Expect(createdPendingNode.Spec.NodeName).Should(Equal("test-node"))

			By("checking the pending node state is creating")
			Eventually(func() updatev1alpha1.CSPNodeState {
				if err := k8sClient.Get(ctx, pendingNodeLookupKey, createdPendingNode); err != nil {
					return ""
				}
				return createdPendingNode.Status.CSPNodeState
			}, timeout, interval).Should(Equal(updatev1alpha1.NodeStateCreating))

			By("updating the deadline to be in the past")
			Eventually(func() error {
				if err := k8sClient.Get(ctx, pendingNodeLookupKey, pendingNode); err != nil {
					return err
				}
				pendingNode.Spec.Deadline = &metav1.Time{Time: fakes.clock.Now().Add(-time.Second)}
				return k8sClient.Update(ctx, pendingNode)
			}, timeout, interval).Should(Succeed())

			By("checking the pending node updates its goal")
			Eventually(func() updatev1alpha1.PendingNodeGoal {
				if err := k8sClient.Get(ctx, pendingNodeLookupKey, createdPendingNode); err != nil {
					return ""
				}
				return createdPendingNode.Spec.Goal
			}, timeout, interval).Should(Equal(updatev1alpha1.NodeGoalLeave))

			By("setting the CSP node state to terminated")
			fakes.nodeStateGetter.setNodeState(pendingNodeName, updatev1alpha1.NodeStateTerminated)
			// trigger reconciliation before regular check interval to speed up test by changing the spec
			Eventually(func() error {
				if err := k8sClient.Get(ctx, pendingNodeLookupKey, pendingNode); err != nil {
					return err
				}
				pendingNode.Spec.Deadline = &metav1.Time{Time: fakes.clock.Now().Add(time.Second)}
				return k8sClient.Update(ctx, pendingNode)
			}, timeout, interval).Should(Or(Succeed(), beGone()))

			By("checking if the pending node resource is deleted")
			Eventually(func() error {
				return k8sClient.Get(ctx, pendingNodeLookupKey, createdPendingNode)
			}, timeout, interval).Should(beGone())
		})

		It("Should should detect successful node join", func() {
			By("setting the CSP node state to creating")
			fakes.nodeStateGetter.setNodeState(pendingNodeName, updatev1alpha1.NodeStateCreating)

			By("creating a pending node resource")
			ctx := context.Background()
			pendingNode := &updatev1alpha1.PendingNode{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "update.edgeless.systems/v1alpha1",
					Kind:       "PendingNode",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: pendingNodeName,
				},
				Spec: updatev1alpha1.PendingNodeSpec{
					ProviderID:     pendingNodeName,
					ScalingGroupID: "scaling-group-id",
					NodeName:       "test-node",
					Goal:           updatev1alpha1.NodeGoalJoin,
					// deadline is always one second in the future
					Deadline: &metav1.Time{Time: fakes.clock.Now().Add(time.Second)},
				},
			}
			Expect(k8sClient.Create(ctx, pendingNode)).Should(Succeed())
			createdPendingNode := &updatev1alpha1.PendingNode{}
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, pendingNodeLookupKey, createdPendingNode); err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())
			Expect(createdPendingNode.Spec.NodeName).Should(Equal("test-node"))

			By("checking the pending node state is creating")
			Eventually(func() updatev1alpha1.CSPNodeState {
				if err := k8sClient.Get(ctx, pendingNodeLookupKey, createdPendingNode); err != nil {
					return ""
				}
				return createdPendingNode.Status.CSPNodeState
			}, timeout, interval).Should(Equal(updatev1alpha1.NodeStateCreating))

			By("setting the CSP node state to ready")
			fakes.nodeStateGetter.setNodeState(pendingNodeName, updatev1alpha1.NodeStateReady)

			By("creating a new node resource with the same node name and provider ID")
			node := &corev1.Node{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Node",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-node",
				},
				Spec: corev1.NodeSpec{
					ProviderID: pendingNodeName,
				},
			}
			Expect(k8sClient.Create(ctx, node)).Should(Succeed())

			By("checking the pending node goal has been reached")
			Eventually(func() updatev1alpha1.PendingNodeStatus {
				if err := k8sClient.Get(ctx, pendingNodeLookupKey, createdPendingNode); err != nil {
					return updatev1alpha1.PendingNodeStatus{}
				}
				return createdPendingNode.Status
			}, timeout, interval).Should(Equal(updatev1alpha1.PendingNodeStatus{
				CSPNodeState: updatev1alpha1.NodeStateReady,
				ReachedGoal:  true,
			}))

			By("cleaning up all resources")
			Expect(k8sClient.Delete(ctx, pendingNode)).Should(Succeed())
			Expect(k8sClient.Delete(ctx, node)).Should(Succeed())
		})
	})
})
