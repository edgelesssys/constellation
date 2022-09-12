/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package controllers

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	updatev1alpha1 "github.com/edgelesssys/constellation/operators/constellation-node-operator/api/v1alpha1"
)

// AutoscalingStrategyReconciler reconciles a AutoscalingStrategy object.
type AutoscalingStrategyReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=update.edgeless.systems,resources=autoscalingstrategies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=update.edgeless.systems,resources=autoscalingstrategies/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=update.edgeless.systems,resources=autoscalingstrategies/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;delete

// Reconcile enabled or disables the cluster-autoscaler based on the AutoscalingStrategy spec
// by modifying the replicas of the Deployment.
func (r *AutoscalingStrategyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logr := log.FromContext(ctx)

	var desiredAutoscalingStrategy updatev1alpha1.AutoscalingStrategy
	if err := r.Get(ctx, req.NamespacedName, &desiredAutoscalingStrategy); err != nil {
		logr.Error(err, "Unable to fetch AutoscalingStrategy")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	var expectedReplicas int32
	if desiredAutoscalingStrategy.Spec.Enabled {
		expectedReplicas = 1
	}

	var autoscalerDeployment appsv1.Deployment
	if err := r.Get(ctx, client.ObjectKey{Namespace: desiredAutoscalingStrategy.Spec.DeploymentNamespace, Name: desiredAutoscalingStrategy.Spec.DeploymentName}, &autoscalerDeployment); err != nil {
		logr.Error(err, "Unable to fetch autoscaler Deployment")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if autoscalerDeployment.Spec.Replicas == nil || *autoscalerDeployment.Spec.Replicas == 0 {
		desiredAutoscalingStrategy.Status.Enabled = false
		desiredAutoscalingStrategy.Status.Replicas = 0
	} else {
		desiredAutoscalingStrategy.Status.Enabled = true
		desiredAutoscalingStrategy.Status.Replicas = *autoscalerDeployment.Spec.Replicas
	}

	if err := r.tryUpdateStatus(ctx, req.NamespacedName, desiredAutoscalingStrategy.Status); err != nil {
		logr.Error(err, "Unable to update AutoscalingStrategy status")
		return ctrl.Result{}, err
	}

	if autoscalerDeployment.Spec.Replicas == nil || *autoscalerDeployment.Spec.Replicas != expectedReplicas {
		logr.Info("Updating autoscaling replicas", "expectedReplicas", expectedReplicas)
		autoscalerDeployment.Spec.Replicas = &expectedReplicas
		if err := r.Update(ctx, &autoscalerDeployment); err != nil {
			logr.Error(err, "Unable to update autoscaler Deployment")
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AutoscalingStrategyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&updatev1alpha1.AutoscalingStrategy{}).
		Complete(r)
}

// tryUpdateStatus attempts to update the AutoscalingStrategy status field in a retry loop.
func (r *AutoscalingStrategyReconciler) tryUpdateStatus(ctx context.Context, name types.NamespacedName, status updatev1alpha1.AutoscalingStrategyStatus) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		var autoscalingStrategy updatev1alpha1.AutoscalingStrategy
		if err := r.Get(ctx, name, &autoscalingStrategy); err != nil {
			return err
		}
		autoscalingStrategy.Status = *status.DeepCopy()
		if err := r.Status().Update(ctx, &autoscalingStrategy); err != nil {
			return err
		}
		return nil
	})
}
