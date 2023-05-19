/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package controllers

import (
	"context"
	"time"

	node "github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/v2/internal/node"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"k8s.io/utils/clock"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	updatev1alpha1 "github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/v2/api/v1alpha1"
)

const (
	nodeNameKey          = ".spec.nodeName"
	defaultCheckInterval = 30 * time.Second
)

// PendingNodeReconciler reconciles a PendingNode object.
type PendingNodeReconciler struct {
	nodeStateGetter
	client.Client
	Scheme *runtime.Scheme
	clock.Clock
}

// NewPendingNodeReconciler creates a new PendingNodeReconciler.
func NewPendingNodeReconciler(nodeStateGetter nodeStateGetter, client client.Client, scheme *runtime.Scheme) *PendingNodeReconciler {
	return &PendingNodeReconciler{
		nodeStateGetter: nodeStateGetter,
		Client:          client,
		Scheme:          scheme,
		Clock:           clock.RealClock{},
	}
}

//+kubebuilder:rbac:groups=update.edgeless.systems,resources=pendingnodes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=update.edgeless.systems,resources=pendingnodes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=update.edgeless.systems,resources=pendingnodes/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=nodes,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=nodes/status,verbs=get

// Reconcile observes the state of a pending node that is either trying to join the cluster or is leaving the cluster (waiting to be destroyed).
// If the node is trying to join the cluster and fails to join within the deadline referenced in the PendingNode spec, the node is deleted.
func (r *PendingNodeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logr := log.FromContext(ctx)
	logr.Info("Reconciling PendingNode", "pendingNode", req.NamespacedName)

	var pendingNode updatev1alpha1.PendingNode
	if err := r.Get(ctx, req.NamespacedName, &pendingNode); err != nil {
		if !errors.IsNotFound(err) {
			logr.Error(err, "Unable to fetch PendingNode")
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	nodeState, err := r.GetNodeState(ctx, pendingNode.Spec.ProviderID)
	if err != nil {
		logr.Error(err, "Unable to get node state")
		return ctrl.Result{Requeue: true}, err
	}
	done, err := r.reachedGoal(ctx, pendingNode, nodeState)
	if err != nil {
		logr.Error(err, "Unable to check if node reached goal")
		return ctrl.Result{}, err
	}

	status := updatev1alpha1.PendingNodeStatus{
		CSPNodeState: nodeState,
		ReachedGoal:  done,
	}
	if err := r.tryUpdateStatus(ctx, req.NamespacedName, &pendingNode, status); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if done {
		logr.Info("Reached goal", "pendingNodeGoal", pendingNode.Spec.Goal, "cspNodeState", nodeState)
		if pendingNode.Spec.Goal == updatev1alpha1.NodeGoalLeave {
			// delete self after pending node has been terminated successfully
			if err := r.deletePendingNode(ctx, req.NamespacedName); err != nil {
				return ctrl.Result{}, err
			}
			logr.Info("Deleted pending node resource for terminated node", "pendingNode", req.Name)
		}
		return ctrl.Result{}, nil
	}

	// check if the deadline has been reached
	// requeue if not
	if pendingNode.Spec.Deadline == nil || r.Now().Before(pendingNode.Spec.Deadline.Time) {
		var requeueAfter time.Duration
		if pendingNode.Spec.Deadline == nil {
			requeueAfter = defaultCheckInterval
		} else {
			requeueAfter = pendingNode.Spec.Deadline.Time.Sub(r.Now())
		}
		return ctrl.Result{
			RequeueAfter: requeueAfter,
		}, nil
	}

	// deadline has been reached
	switch pendingNode.Spec.Goal {
	case updatev1alpha1.NodeGoalJoin:
		logr.Info("Node did not get ready in time or failed to join", "pendingNode", pendingNode.Spec.NodeName, "pendingNodeGoal", pendingNode.Spec.Goal, "cspNodeState", nodeState)
		if err := r.DeleteNode(ctx, pendingNode.Spec.ProviderID); err != nil {
			logr.Error(err, "Unable to delete node")
			return ctrl.Result{}, err
		}
		pendingNode.Spec.Goal = updatev1alpha1.NodeGoalLeave
		pendingNode.Spec.Deadline = nil
		if err := r.Update(ctx, &pendingNode); err != nil {
			logr.Error(err, "Unable to update PendingNode")
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	case updatev1alpha1.NodeGoalLeave:
		logr.Info("Node was not terminated on time", "pendingNodeGoal", pendingNode.Spec.Goal, "cspNodeState", nodeState)
		// TODO: decide if other actions should be taken here (e.g. send another request to delete node)
		return ctrl.Result{RequeueAfter: defaultCheckInterval}, nil
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PendingNodeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// index pending nodes by nodename.
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &updatev1alpha1.PendingNode{}, nodeNameKey, func(rawObj client.Object) []string {
		pendingNode := rawObj.(*updatev1alpha1.PendingNode)
		return []string{pendingNode.Spec.NodeName}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&updatev1alpha1.PendingNode{}).
		Watches(
			&source.Kind{Type: &corev1.Node{}},
			handler.EnqueueRequestsFromMapFunc(r.findObjectsForNode),
			builder.WithPredicates(nodeStateChangePredicate()),
		).
		Complete(r)
}

// nodeStateChangePredicate filters events on Node resources to those that indicate
// joining, leaving, node becoming ready or node receiving a provider id.
func nodeStateChangePredicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return true
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldNode, ok := e.ObjectOld.(*corev1.Node)
			if !ok {
				return false
			}
			newNode, ok := e.ObjectNew.(*corev1.Node)
			if !ok {
				return false
			}
			becameReady := !node.Ready(oldNode) && node.Ready(newNode)
			receivedProviderID := len(oldNode.Spec.ProviderID) == 0 && len(newNode.Spec.ProviderID) != 0
			return becameReady || receivedProviderID
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return true
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return false
		},
	}
}

// findObjectsForNode requests reconciliation for PendingNode whenever the corresponding Node state changes.
func (r *PendingNodeReconciler) findObjectsForNode(rawNode client.Object) []reconcile.Request {
	var pendingNodesList updatev1alpha1.PendingNodeList
	err := r.List(context.Background(), &pendingNodesList, client.MatchingFields{nodeNameKey: rawNode.GetName()})
	if err != nil {
		return []reconcile.Request{}
	}

	requests := make([]reconcile.Request, len(pendingNodesList.Items))
	for i, item := range pendingNodesList.Items {
		requests[i] = reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name: item.GetName(),
			},
		}
	}
	return requests
}

// reachedGoal checks if a pending node has reached its goal (joining or leaving the cluster).
// - joining node: CSP reports the node instance as running and node has joined kubernetes cluster.
// - leaving node: CSP reports node instance as terminated.
func (r *PendingNodeReconciler) reachedGoal(ctx context.Context, pendingNode updatev1alpha1.PendingNode, nodeState updatev1alpha1.CSPNodeState) (bool, error) {
	if pendingNode.Spec.Goal == updatev1alpha1.NodeGoalJoin {
		if err := r.Get(ctx, types.NamespacedName{Name: pendingNode.Spec.NodeName}, &corev1.Node{}); err != nil {
			return false, client.IgnoreNotFound(err)
		}
		return nodeState == updatev1alpha1.NodeStateReady, nil
	}
	return nodeState == updatev1alpha1.NodeStateTerminated, nil
}

// deletePendingNode deletes a PendingNode resource.
func (r *PendingNodeReconciler) deletePendingNode(ctx context.Context, name types.NamespacedName) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		var pendingNode updatev1alpha1.PendingNode
		if err := r.Get(ctx, name, &pendingNode); err != nil {
			return err
		}
		return r.Delete(ctx, &pendingNode)
	})
}

// tryUpdateStatus attempts to update the PendingNode status field in a retry loop.
func (r *PendingNodeReconciler) tryUpdateStatus(ctx context.Context, name types.NamespacedName, pendingNode *updatev1alpha1.PendingNode, status updatev1alpha1.PendingNodeStatus) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if err := r.Get(ctx, name, pendingNode); err != nil {
			return err
		}
		pendingNode.Status = *status.DeepCopy()
		return r.Status().Update(ctx, pendingNode)
	})
}

type nodeStateGetter interface {
	// GetNodeState retrieves the state of a pending node from a CSP.
	GetNodeState(ctx context.Context, providerID string) (updatev1alpha1.CSPNodeState, error)
	// DeleteNode starts the termination of the node at the CSP.
	DeleteNode(ctx context.Context, providerID string) error
}
