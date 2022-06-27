
package controllers

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	updatev1alpha1 "github.com/edgelesssys/constellation/operators/constellation-node-operator/api/v1alpha1"
)

// NodeImageReconciler reconciles a NodeImage object
type NodeImageReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=update.edgeless.systems,resources=nodeimages,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=update.edgeless.systems,resources=nodeimages/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=update.edgeless.systems,resources=nodeimages/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the NodeImage object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.2/pkg/reconcile
func (r *NodeImageReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO(user): your logic here

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NodeImageReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&updatev1alpha1.NodeImage{}).
		Complete(r)
}
