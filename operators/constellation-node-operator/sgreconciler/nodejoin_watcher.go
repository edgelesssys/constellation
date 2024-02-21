/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package sgreconciler

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// NodeJoinWatcher watches node join / leave events.
type NodeJoinWatcher struct {
	trigger func()
	client.Client
	Scheme *runtime.Scheme
}

// NewNodeJoinWatcher creates a new NodeJoinWatcher.
func NewNodeJoinWatcher(trigger func(), client client.Client, scheme *runtime.Scheme) *NodeJoinWatcher {
	return &NodeJoinWatcher{
		trigger: trigger,
		Client:  client,
		Scheme:  scheme,
	}
}

//+kubebuilder:rbac:groups="",resources=nodes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=nodes/status,verbs=get

// Reconcile reconciles node join / leave events.
func (w *NodeJoinWatcher) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logr := log.FromContext(ctx)
	logr.Info("node has joined or left the cluster", "node", req.Name)
	w.trigger()

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (w *NodeJoinWatcher) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("node-join-watcher").
		Watches(
			client.Object(&corev1.Node{}),
			handler.EnqueueRequestsFromMapFunc(func(_ context.Context, obj client.Object) []ctrl.Request {
				return []ctrl.Request{{
					NamespacedName: types.NamespacedName{Name: obj.GetName()},
				}}
			}),
			builder.WithPredicates(nodeJoinLeavePredicate()),
		).
		Complete(w)
}

// nodeJoinLeavePredicate returns a predicate that returns true if a node joins or leaves the cluster.
func nodeJoinLeavePredicate() predicate.Predicate {
	return predicate.Funcs{
		// CreateFunc is not specified => never filter out create events
		// DeleteFunc is not specified => never filter out delete events
		UpdateFunc:  func(_ event.UpdateEvent) bool { return false },
		GenericFunc: func(_ event.GenericEvent) bool { return false },
	}
}
