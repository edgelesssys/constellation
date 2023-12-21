/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package controllers

import (
	"context"
	"fmt"
	"sort"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	updatev1alpha1 "github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/api/v1alpha1"
)

// AutoscalingStrategyReconciler reconciles a AutoscalingStrategy object.
type AutoscalingStrategyReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=update.edgeless.systems,resources=autoscalingstrategies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=update.edgeless.systems,resources=autoscalingstrategies/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=update.edgeless.systems,resources=autoscalingstrategies/finalizers,verbs=update
//+kubebuilder:rbac:groups=update.edgeless.systems,resources=scalinggroups,verbs=get;list;watch
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

	var scalingGroups updatev1alpha1.ScalingGroupList
	if err := r.List(ctx, &scalingGroups); err != nil {
		logr.Error(err, "Unable to fetch ScalingGroups")
		return ctrl.Result{}, err
	}

	autoscalerArgs := []string{"./cluster-autoscaler"}
	for key, val := range desiredAutoscalingStrategy.Spec.AutoscalerExtraArgs {
		autoscalerArgs = append(autoscalerArgs, "--"+key+"="+val)
	}
	const nodeGroupFmt = "--nodes=%d:%d:%s"
	for _, group := range scalingGroups.Items {
		// Don't autoscale control plane nodes for safety reasons.
		if group.Spec.Autoscaling && group.Spec.Role != updatev1alpha1.ControlPlaneRole {
			groupArg := fmt.Sprintf(nodeGroupFmt, group.Spec.Min, group.Spec.Max, group.Spec.AutoscalerGroupName)
			autoscalerArgs = append(autoscalerArgs, groupArg)
		}
	}
	sort.Strings(autoscalerArgs[1:])

	var autoscalerDeployment appsv1.Deployment
	deploymentKey := client.ObjectKey{
		Namespace: desiredAutoscalingStrategy.Spec.DeploymentNamespace,
		Name:      desiredAutoscalingStrategy.Spec.DeploymentName,
	}
	if err := r.Get(ctx, deploymentKey, &autoscalerDeployment); err != nil {
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

	var needUpdate bool
	if autoscalerDeployment.Spec.Replicas == nil || *autoscalerDeployment.Spec.Replicas != expectedReplicas {
		needUpdate = needUpdate || true
	}
	containers := autoscalerDeployment.Spec.Template.Spec.Containers
	if len(containers) != 0 && containers[0].Command == nil { // uninitialized
		needUpdate = needUpdate || true
	}
	if len(containers) != 0 && containers[0].Command != nil { // args updated
		if len(containers[0].Command) != len(autoscalerArgs) {
			needUpdate = needUpdate || true
		} else {
			for i, arg := range containers[0].Command {
				if arg != autoscalerArgs[i] {
					needUpdate = needUpdate || true
					break
				}
			}
		}
	}

	if !needUpdate {
		return ctrl.Result{}, nil
	}

	logr.Info("Updating autoscaling replicas and command", "expectedReplicas", expectedReplicas, "autoscalerArgs", autoscalerArgs)
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if err := r.Get(ctx, deploymentKey, &autoscalerDeployment); err != nil {
			return err
		}
		autoscalerDeployment.Spec.Replicas = &expectedReplicas
		if len(containers) != 0 {
			logr.Info("Updating autoscaler command", "old", containers[0].Command, "new", autoscalerArgs)
			autoscalerDeployment.Spec.Template.Spec.Containers[0].Command = autoscalerArgs
		}
		return r.Update(ctx, &autoscalerDeployment)
	})
	if err != nil {
		logr.Error(err, "Unable to update autoscaler Deployment")
		return ctrl.Result{}, err
	}

	return ctrl.Result{Requeue: true}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AutoscalingStrategyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&updatev1alpha1.AutoscalingStrategy{}).
		Watches(
			client.Object(&updatev1alpha1.ScalingGroup{}),
			handler.EnqueueRequestsFromMapFunc(r.findObjectsForDeployment),
			builder.WithPredicates(scalingGroupChangePredicate()),
		).
		Complete(r)
}

func (r *AutoscalingStrategyReconciler) findObjectsForDeployment(ctx context.Context, _ client.Object) []reconcile.Request {
	var autoscalingStrats updatev1alpha1.AutoscalingStrategyList
	err := r.List(ctx, &autoscalingStrats)
	if err != nil {
		return []reconcile.Request{}
	}

	requests := make([]reconcile.Request, len(autoscalingStrats.Items))
	for i, item := range autoscalingStrats.Items {
		requests[i] = reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name: item.GetName(),
			},
		}
	}
	return requests
}

// scalingGroupChangePredicate filters events on scaling group resources.
func scalingGroupChangePredicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return true
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldGroup, ok := e.ObjectOld.(*updatev1alpha1.ScalingGroup)
			if !ok {
				return false
			}
			newGroup, ok := e.ObjectNew.(*updatev1alpha1.ScalingGroup)
			if !ok {
				return false
			}
			switch {
			case oldGroup.Spec.Min != newGroup.Spec.Min:
				return true
			case oldGroup.Spec.Max != newGroup.Spec.Max:
				return true
			case oldGroup.Spec.Autoscaling != newGroup.Spec.Autoscaling:
				return true
			default:
				return false
			}
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return true
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return false
		},
	}
}

// tryUpdateStatus attempts to update the AutoscalingStrategy status field in a retry loop.
func (r *AutoscalingStrategyReconciler) tryUpdateStatus(ctx context.Context, name types.NamespacedName, status updatev1alpha1.AutoscalingStrategyStatus) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		var autoscalingStrategy updatev1alpha1.AutoscalingStrategy
		if err := r.Get(ctx, name, &autoscalingStrategy); err != nil {
			return err
		}
		autoscalingStrategy.Status = *status.DeepCopy()
		return r.Status().Update(ctx, &autoscalingStrategy)
	})
}
