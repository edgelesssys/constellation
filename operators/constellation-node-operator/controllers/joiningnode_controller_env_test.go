//go:build integration

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package controllers

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	updatev1alpha1 "github.com/edgelesssys/constellation/operators/constellation-node-operator/v2/api/v1alpha1"
)

var _ = Describe("JoiningNode controller", func() {
	const (
		nodeName1       = "node-name-1"
		nodeName2       = "node-name-2"
		componentsHash1 = "test-hash-1"
		componentsHash2 = "test-hash-2"

		timeout  = time.Second * 20
		duration = time.Second * 2
		interval = time.Millisecond * 250
	)
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
					Name:           nodeName1,
					ComponentsHash: componentsHash1,
				},
			}
			Expect(k8sClient.Create(ctx, joiningNode)).Should(Succeed())
			createdJoiningNode := &updatev1alpha1.JoiningNode{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{Name: nodeName1}, createdJoiningNode)
			}, timeout, interval).Should(Succeed())
			Expect(createdJoiningNode.Spec.Name).Should(Equal(nodeName1))
			Expect(createdJoiningNode.Spec.ComponentsHash).Should(Equal(componentsHash1))

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
				return createdNode.Annotations[NodeKubernetesComponentsHashAnnotationKey]
			}, timeout, interval).Should(Equal(componentsHash1))

			By("deleting the joining node resource")
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{Name: joiningNode.Name}, createdJoiningNode)
			}, timeout, interval).ShouldNot(Succeed())
		})
	})
	It("Should annotate the corresponding node when creating the node first", func() {
		ctx := context.Background()
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
				Name:           nodeName2,
				ComponentsHash: componentsHash2,
			},
		}
		Expect(k8sClient.Create(ctx, joiningNode)).Should(Succeed())
		createdJoiningNode := &updatev1alpha1.JoiningNode{}
		Eventually(func() error {
			return k8sClient.Get(ctx, types.NamespacedName{Name: joiningNode.Name}, createdJoiningNode)
		}, timeout, interval).Should(Succeed())
		Expect(createdJoiningNode.Spec.Name).Should(Equal(nodeName2))
		Expect(createdJoiningNode.Spec.ComponentsHash).Should(Equal(componentsHash2))

		By("annotating the node")
		Eventually(func() string {
			_ = k8sClient.Get(ctx, types.NamespacedName{Name: createdNode.Name}, createdNode)
			return createdNode.Annotations[NodeKubernetesComponentsHashAnnotationKey]
		}, timeout, interval).Should(Equal(componentsHash2))

		By("deleting the joining node resource")
		Eventually(func() error {
			return k8sClient.Get(ctx, types.NamespacedName{Name: joiningNode.Name}, createdJoiningNode)
		}, timeout, interval).ShouldNot(Succeed())
	})
})
