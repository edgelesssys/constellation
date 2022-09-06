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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	updatev1alpha1 "github.com/edgelesssys/constellation/operators/constellation-node-operator/api/v1alpha1"
)

var _ = Describe("AutoscalingStrategy controller", func() {
	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		ClusterAutoscalerDeploymentName      = "cluster-autoscaler"
		ClusterAutoscalerDeploymentNamespace = "kube-system"
		AutoscalingStrategyName              = "test-strategy"

		timeout  = time.Second * 20
		duration = time.Second * 2
		interval = time.Millisecond * 250
	)

	var (
		ClusterAutoscalerStartingReplicas int32 = 0
		DeploymentLookupKey                     = types.NamespacedName{Name: ClusterAutoscalerDeploymentName, Namespace: ClusterAutoscalerDeploymentNamespace}
	)

	checkDeploymentReplicas := func() int32 {
		deployment := &appsv1.Deployment{}
		err := k8sClient.Get(ctx, DeploymentLookupKey, deployment)
		if err != nil {
			return -1
		}
		if deployment.Spec.Replicas == nil {
			// replicas defaults to 1
			return 1
		}
		return *deployment.Spec.Replicas
	}

	Context("When enabling autoscaling", func() {
		It("Should increase and decrease the replicas of the cluster-autoscaler deployment", func() {
			By("creating a cluster-autoscaler deployment")
			ctx := context.Background()
			autoscalerDeployment := &appsv1.Deployment{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "apps/v1",
					Kind:       "Deployment",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      ClusterAutoscalerDeploymentName,
					Namespace: ClusterAutoscalerDeploymentNamespace,
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: &ClusterAutoscalerStartingReplicas,
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
			createdDeployment := &appsv1.Deployment{}
			Eventually(func() error {
				return k8sClient.Get(ctx, DeploymentLookupKey, createdDeployment)
			}, timeout, interval).Should(Succeed())
			Expect(createdDeployment.Spec.Replicas).NotTo(BeNil())
			Expect(*createdDeployment.Spec.Replicas).Should(Equal(ClusterAutoscalerStartingReplicas))

			By("creating an autoscaling strategy")
			strategy := &updatev1alpha1.AutoscalingStrategy{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "update.edgeless.systems/v1alpha1",
					Kind:       "AutoscalingStrategy",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: AutoscalingStrategyName,
				},
				Spec: updatev1alpha1.AutoscalingStrategySpec{
					DeploymentName:      ClusterAutoscalerDeploymentName,
					DeploymentNamespace: ClusterAutoscalerDeploymentNamespace,
				},
			}
			Expect(k8sClient.Create(ctx, strategy)).Should(Succeed())
			strategyLookupKey := types.NamespacedName{Name: AutoscalingStrategyName}
			createdStrategy := &updatev1alpha1.AutoscalingStrategy{}
			Eventually(func() error {
				return k8sClient.Get(ctx, strategyLookupKey, createdStrategy)
			}, timeout, interval).Should(Succeed())

			By("checking the autoscaling strategy status shows zero replicas")
			Eventually(func() (int32, error) {
				err := k8sClient.Get(ctx, strategyLookupKey, createdStrategy)
				if err != nil {
					return -1, err
				}
				return createdStrategy.Status.Replicas, nil
			}, timeout, interval).Should(Equal(int32(0)))

			By("enabling the autoscaler in the strategy")
			Expect(k8sClient.Get(ctx, strategyLookupKey, strategy)).Should(Succeed())
			strategy.Spec.Enabled = true
			Expect(k8sClient.Update(ctx, strategy)).Should(Succeed())

			By("checking the autoscaling deployment eventually has one replica")
			Eventually(checkDeploymentReplicas, timeout, interval).Should(Equal(int32(1)))

			By("checking the autoscaling strategy status shows one replica")
			Eventually(func() (int32, error) {
				err := k8sClient.Get(ctx, strategyLookupKey, createdStrategy)
				if err != nil {
					return -1, err
				}
				return createdStrategy.Status.Replicas, nil
			}, duration, interval).Should(Equal(int32(1)))
			Consistently(func() (int32, error) {
				err := k8sClient.Get(ctx, strategyLookupKey, createdStrategy)
				if err != nil {
					return -1, err
				}
				return createdStrategy.Status.Replicas, nil
			}, duration, interval).Should(Equal(int32(1)))

			By("disabling the autoscaler in the strategy")
			Expect(k8sClient.Get(ctx, strategyLookupKey, strategy)).Should(Succeed())
			strategy.Spec.Enabled = false
			Expect(k8sClient.Update(ctx, strategy)).Should(Succeed())

			By("checking the autoscaling deployment eventually has zero replicas")
			Eventually(checkDeploymentReplicas, timeout, interval).Should(Equal(int32(0)))

			By("cleaning up all resources")
			Expect(k8sClient.Delete(ctx, autoscalerDeployment)).Should(Succeed())
			Expect(k8sClient.Delete(ctx, strategy)).Should(Succeed())
		})
	})
})
