/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package controllers

import (
	"context"

	updatev1alpha1 "github.com/edgelesssys/constellation/operators/constellation-node-operator/v2/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	// NodeKubernetesComponentsHashAnnotationKey is the name of the annotation holding the hash of the installed components of this node.
	NodeKubernetesComponentsHashAnnotationKey = "updates.edgeless.systems/kubernetes-components-hash"

	joiningNodeNameKey = ".spec.name"
)

// JoiningNodesReconciler reconciles a JoiningNode object.
type JoiningNodesReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// NewJoiningNodesReconciler creates a new JoiningNodesReconciler.
func NewJoiningNodesReconciler(client client.Client, scheme *runtime.Scheme) *JoiningNodesReconciler {
	return &JoiningNodesReconciler{
		Client: client,
		Scheme: scheme,
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
		logr.Error(err, "unable to fetch JoiningNodes")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		var node corev1.Node
		if err := r.Get(ctx, types.NamespacedName{Name: joiningNode.Spec.Name}, &node); err != nil {
			logr.Error(err, "unable to fetch Node")
			return err
		}

		// add annotations to node
		if node.Annotations == nil {
			node.Annotations = map[string]string{}
		}
		node.Annotations[NodeKubernetesComponentsHashAnnotationKey] = joiningNode.Spec.ComponentsHash
		return r.Update(ctx, &node)
	})
	if err != nil {
		logr.Error(err, "unable to update Node")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if err := r.Delete(ctx, &joiningNode); err != nil {
			logr.Error(err, "unable to delete JoiningNode")
			return err
		}
		return nil
	})
	if err != nil {
		logr.Error(err, "unable to delete JoiningNode")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
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
			&source.Kind{Type: &corev1.Node{}},
			handler.EnqueueRequestsFromMapFunc(r.findAllJoiningNodes),
		).
		Complete(r)
}

func (r *JoiningNodesReconciler) findAllJoiningNodes(obj client.Object) []reconcile.Request {
	var joiningNodesList updatev1alpha1.JoiningNodeList
	err := r.List(context.TODO(), &joiningNodesList, client.MatchingFields{joiningNodeNameKey: obj.GetName()})
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
