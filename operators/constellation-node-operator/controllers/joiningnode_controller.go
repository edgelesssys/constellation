/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package controllers

import (
	"context"
	"time"

	mainconstants "github.com/edgelesssys/constellation/v2/internal/constants"
	updatev1alpha1 "github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"k8s.io/utils/clock"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	joiningNodeNameKey = ".spec.name"
)

// JoiningNodesReconciler reconciles a JoiningNode object.
type JoiningNodesReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	clock.Clock
}

// NewJoiningNodesReconciler creates a new JoiningNodesReconciler.
func NewJoiningNodesReconciler(client client.Client, scheme *runtime.Scheme) *JoiningNodesReconciler {
	return &JoiningNodesReconciler{
		Client: client,
		Scheme: scheme,
		Clock:  clock.RealClock{},
	}
}

//+kubebuilder:rbac:groups=update.edgeless.systems,resources=joiningnodes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=update.edgeless.systems,resources=joiningnodes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=update.edgeless.systems,resources=joiningnodes/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=nodes,verbs=get;list;watch

// Reconcile annotates the node with the components hash.
func (r *JoiningNodesReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logr := log.FromContext(ctx)

	var joiningNode updatev1alpha1.JoiningNode
	if err := r.Get(ctx, req.NamespacedName, &joiningNode); err != nil {
		if !errors.IsNotFound(err) {
			logr.Error(err, "Unable to fetch JoiningNode")
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		var node corev1.Node
		if err := r.Get(ctx, types.NamespacedName{Name: joiningNode.Spec.Name}, &node); err != nil {
			logr.Info("unable to fetch Node", "err", err)
			return err
		}

		// add annotations to node
		if node.Annotations == nil {
			node.Annotations = map[string]string{}
		}
		node.Annotations[mainconstants.NodeKubernetesComponentsAnnotationKey] = joiningNode.Spec.ComponentsReference
		return r.Update(ctx, &node)
	})
	if err != nil {
		// check if the deadline has been reached
		// requeue if not
		if joiningNode.Spec.Deadline == nil || r.Now().Before(joiningNode.Spec.Deadline.Time) {
			var requeueAfter time.Duration
			if joiningNode.Spec.Deadline == nil {
				requeueAfter = defaultCheckInterval
			} else {
				requeueAfter = joiningNode.Spec.Deadline.Time.Sub(r.Now())
			}
			return ctrl.Result{
				RequeueAfter: requeueAfter,
			}, nil
		}
	}

	// if the joining node is too old or the annotation succeeded, delete it.
	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if err := r.Get(ctx, req.NamespacedName, &joiningNode); err != nil {
			return client.IgnoreNotFound(err)
		}

		return client.IgnoreNotFound(r.Delete(ctx, &joiningNode))
	})
	if err != nil {
		logr.Error(err, "unable to delete JoiningNode")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *JoiningNodesReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// index joining nodes by nodename.
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &updatev1alpha1.JoiningNode{}, joiningNodeNameKey, func(rawObj client.Object) []string {
		joiningNode := rawObj.(*updatev1alpha1.JoiningNode)
		return []string{joiningNode.Spec.Name}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&updatev1alpha1.JoiningNode{}).
		Watches(
			client.Object(&corev1.Node{}),
			handler.EnqueueRequestsFromMapFunc(r.findAllJoiningNodes),
		).
		Complete(r)
}

func (r *JoiningNodesReconciler) findAllJoiningNodes(ctx context.Context, obj client.Object) []reconcile.Request {
	var joiningNodesList updatev1alpha1.JoiningNodeList
	err := r.List(ctx, &joiningNodesList, client.MatchingFields{joiningNodeNameKey: obj.GetName()})
	if err != nil {
		return []reconcile.Request{}
	}
	requests := make([]reconcile.Request, len(joiningNodesList.Items))
	for i, item := range joiningNodesList.Items {
		requests[i] = reconcile.Request{
			NamespacedName: types.NamespacedName{Name: item.GetName()},
		}
	}
	return requests
}
