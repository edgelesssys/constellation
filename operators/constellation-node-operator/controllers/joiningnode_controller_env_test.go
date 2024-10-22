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

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	mainconstants "github.com/edgelesssys/constellation/v2/internal/constants"
	updatev1alpha1 "github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/api/v1alpha1"
)

var _ = Describe("JoiningNode controller", func() {
	const (
		nodeName1            = "node-name-1"
		nodeName2            = "node-name-2"
		nodeName3            = "node-name-3"
		ComponentsReference1 = "test-ref-1"
		ComponentsReference2 = "test-ref-2"
		ComponentsReference3 = "test-ref-3"

		timeout  = time.Second * 20
		duration = time.Second * 2
		interval = time.Millisecond * 250
	)

	AfterEach(func() {
		Eventually(func() error {
			return resetEnv()
		}, 30*time.Second, 1*time.Second).Should(Succeed())
	})

	Context("When changing a joining node resource spec", func() {
		It("Should annotate the corresponding node when creating the CRD first", func() {
			By("creating a joining node resource")
			ctx := context.Background()
			joiningNode := &updatev1alpha1.JoiningNode{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "update.edgeless.systems/v1alpha1",
					Kind:       "JoiningNode",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: nodeName1,
				},
				Spec: updatev1alpha1.JoiningNodeSpec{
					Name:                nodeName1,
					ComponentsReference: ComponentsReference1,
				},
			}
			Expect(k8sClient.Create(ctx, joiningNode)).Should(Succeed())
			createdJoiningNode := &updatev1alpha1.JoiningNode{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{Name: nodeName1}, createdJoiningNode)
			}, timeout, interval).Should(Succeed())
			Expect(createdJoiningNode.Spec.Name).Should(Equal(nodeName1))
			Expect(createdJoiningNode.Spec.ComponentsReference).Should(Equal(ComponentsReference1))

			By("creating a node")
			node := &corev1.Node{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "update.edgeless.systems/v1alpha1",
					Kind:       "Node",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: nodeName1,
				},
				Spec: corev1.NodeSpec{},
			}
			Expect(k8sClient.Create(ctx, node)).Should(Succeed())
			createdNode := &corev1.Node{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{Name: nodeName1}, createdNode)
			}, timeout, interval).Should(Succeed())
			Expect(createdNode.ObjectMeta.Name).Should(Equal(nodeName1))

			By("annotating the node")
			Eventually(func() string {
				_ = k8sClient.Get(ctx, types.NamespacedName{Name: nodeName1}, createdNode)
				return createdNode.Annotations[mainconstants.NodeKubernetesComponentsAnnotationKey]
			}, timeout, interval).Should(Equal(ComponentsReference1))

			By("deleting the joining node resource")
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{Name: joiningNode.Name}, createdJoiningNode)
			}, timeout, interval).ShouldNot(Succeed())
		})
	})
	It("Should annotate the corresponding node when creating the node first", func() {
		ctx := context.Background()

		By("creating a joining node resource")
		joiningNode := &updatev1alpha1.JoiningNode{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "update.edgeless.systems/v1alpha1",
				Kind:       "JoiningNode",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: nodeName2,
			},
			Spec: updatev1alpha1.JoiningNodeSpec{
				Name:                nodeName2,
				ComponentsReference: ComponentsReference2,
			},
		}
		Expect(k8sClient.Create(ctx, joiningNode)).Should(Succeed())
		createdJoiningNode := &updatev1alpha1.JoiningNode{}
		Eventually(func() error {
			return k8sClient.Get(ctx, types.NamespacedName{Name: joiningNode.Name}, createdJoiningNode)
		}, timeout, interval).Should(Succeed())
		Expect(createdJoiningNode.Spec.Name).Should(Equal(nodeName2))
		Expect(createdJoiningNode.Spec.ComponentsReference).Should(Equal(ComponentsReference2))

		By("creating a node")
		node := &corev1.Node{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "update.edgeless.systems/v1alpha1",
				Kind:       "Node",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: nodeName2,
			},
			Spec: corev1.NodeSpec{},
		}
		Expect(k8sClient.Create(ctx, node)).Should(Succeed())
		createdNode := &corev1.Node{}
		Eventually(func() error {
			return k8sClient.Get(ctx, types.NamespacedName{Name: nodeName2}, createdNode)
		}, timeout, interval).Should(Succeed())
		Expect(createdNode.ObjectMeta.Name).Should(Equal(nodeName2))

		By("annotating the node")
		Eventually(func() string {
			_ = k8sClient.Get(ctx, types.NamespacedName{Name: createdNode.Name}, createdNode)
			return createdNode.Annotations[mainconstants.NodeKubernetesComponentsAnnotationKey]
		}, timeout, interval).Should(Equal(ComponentsReference2))

		By("deleting the joining node resource")
		Eventually(func() error {
			return k8sClient.Get(ctx, types.NamespacedName{Name: joiningNode.Name}, createdJoiningNode)
		}, timeout, interval).ShouldNot(Succeed())
	})
	It("Should clean up the joining node resource after the deadline is reached", func() {
		ctx := context.Background()
		By("creating a joining node resource")
		joiningNode := &updatev1alpha1.JoiningNode{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "update.edgeless.systems/v1alpha1",
				Kind:       "JoiningNode",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: nodeName3,
			},
			Spec: updatev1alpha1.JoiningNodeSpec{
				Name:                nodeName3,
				ComponentsReference: ComponentsReference3,
				// create without deadline first
			},
		}
		Expect(k8sClient.Create(ctx, joiningNode)).Should(Succeed())
		createdJoiningNode := &updatev1alpha1.JoiningNode{}
		Eventually(func() error {
			return k8sClient.Get(ctx, types.NamespacedName{Name: joiningNode.Name}, createdJoiningNode)
		}, timeout, interval).Should(Succeed())
		Expect(createdJoiningNode.Spec.Name).Should(Equal(nodeName3))
		Expect(createdJoiningNode.Spec.ComponentsReference).Should(Equal(ComponentsReference3))

		By("setting the deadline to the past")
		Eventually(func() error {
			if err := k8sClient.Get(ctx, types.NamespacedName{Name: joiningNode.Name}, createdJoiningNode); err != nil {
				return err
			}
			createdJoiningNode.Spec.Deadline = &metav1.Time{Time: fakes.clock.Now().Add(-time.Second)}
			return k8sClient.Update(ctx, createdJoiningNode)
		}, timeout, interval).Should(Succeed())

		By("deleting the joining node resource")
		Eventually(func() error {
			return k8sClient.Get(ctx, types.NamespacedName{Name: joiningNode.Name}, createdJoiningNode)
		}, timeout, interval).ShouldNot(Succeed())
	})
})
