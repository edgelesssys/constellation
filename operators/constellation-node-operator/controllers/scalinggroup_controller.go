/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package controllers

import (
	"context"
	"strings"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	updatev1alpha1 "github.com/edgelesssys/constellation/operators/constellation-node-operator/api/v1alpha1"
)

const (
	nodeImageField                        = ".spec.nodeImage"
	conditionScalingGroupUpToDateReason   = "ScalingGroupNodeImageUpToDate"
	conditionScalingGroupUpToDateMessage  = "Scaling group will use the latest image when creating new nodes"
	conditionScalingGroupOutOfDateReason  = "ScalingGroupNodeImageOutOfDate"
	conditionScalingGroupOutOfDateMessage = "Scaling group will not use the latest image when creating new nodes"
)

// ScalingGroupReconciler reconciles a ScalingGroup object.
type ScalingGroupReconciler struct {
	scalingGroupUpdater
	client.Client
	Scheme *runtime.Scheme
}

// NewScalingGroupReconciler returns a new ScalingGroupReconciler.
func NewScalingGroupReconciler(scalingGroupUpdater scalingGroupUpdater, client client.Client, scheme *runtime.Scheme) *ScalingGroupReconciler {
	return &ScalingGroupReconciler{
		scalingGroupUpdater: scalingGroupUpdater,
		Client:              client,
		Scheme:              scheme,
	}
}

//+kubebuilder:rbac:groups=update.edgeless.systems,resources=scalinggroups,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=update.edgeless.systems,resources=scalinggroups/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=update.edgeless.systems,resources=scalinggroups/finalizers,verbs=update
//+kubebuilder:rbac:groups=update.edgeless.systems,resources=nodeimage,verbs=get;list;watch
//+kubebuilder:rbac:groups=update.edgeless.systems,resources=nodeimages/status,verbs=get

// Reconcile reads the latest node image from the referenced NodeImage spec and updates the scaling group to match.
func (r *ScalingGroupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logr := log.FromContext(ctx)

	var desiredScalingGroup updatev1alpha1.ScalingGroup
	if err := r.Get(ctx, req.NamespacedName, &desiredScalingGroup); err != nil {
		logr.Error(err, "Unable to fetch ScalingGroup")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	var desiredNodeImage updatev1alpha1.NodeImage
	if err := r.Get(ctx, client.ObjectKey{Name: desiredScalingGroup.Spec.NodeImage}, &desiredNodeImage); err != nil {
		logr.Error(err, "Unable to fetch NodeImage")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	nodeImage, err := r.scalingGroupUpdater.GetScalingGroupImage(ctx, desiredScalingGroup.Spec.GroupID)
	if err != nil {
		logr.Error(err, "Unable to get ScalingGroup NodeImage")
		return ctrl.Result{}, err
	}

	desiredScalingGroup.Status.ImageReference = nodeImage
	outdatedCondition := metav1.Condition{
		Type: updatev1alpha1.ConditionOutdated,
	}
	imagesMatch := strings.EqualFold(nodeImage, desiredNodeImage.Spec.ImageReference)
	if imagesMatch {
		outdatedCondition.Status = metav1.ConditionFalse
		outdatedCondition.Reason = conditionScalingGroupUpToDateReason
		outdatedCondition.Message = conditionScalingGroupUpToDateMessage
	} else {
		outdatedCondition.Status = metav1.ConditionTrue
		outdatedCondition.Reason = conditionScalingGroupOutOfDateReason
		outdatedCondition.Message = conditionScalingGroupOutOfDateMessage
	}
	meta.SetStatusCondition(&desiredScalingGroup.Status.Conditions, outdatedCondition)
	if err := r.Status().Update(ctx, &desiredScalingGroup); err != nil {
		logr.Error(err, "Unable to update AutoscalingStrategy status")
		return ctrl.Result{}, err
	}

	if !imagesMatch {
		logr.Info("ScalingGroup NodeImage is out of date")
		if err := r.scalingGroupUpdater.SetScalingGroupImage(ctx, desiredScalingGroup.Spec.GroupID, desiredNodeImage.Spec.ImageReference); err != nil {
			logr.Error(err, "Unable to set ScalingGroup NodeImage")
			return ctrl.Result{}, err
		}
		// requeue to update status
		return ctrl.Result{Requeue: true}, nil
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ScalingGroupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &updatev1alpha1.ScalingGroup{}, nodeImageField, func(rawObj client.Object) []string {
		// Extract the NodeImage name from the ScalingGroup Spec, if one is provided
		scalingGroup := rawObj.(*updatev1alpha1.ScalingGroup)
		if scalingGroup.Spec.NodeImage == "" {
			return nil
		}
		return []string{scalingGroup.Spec.NodeImage}
	}); err != nil {
		return err
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&updatev1alpha1.ScalingGroup{}).
		Watches(
			&source.Kind{Type: &updatev1alpha1.NodeImage{}},
			handler.EnqueueRequestsFromMapFunc(r.findObjectsForNodeImage),
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{}),
		).
		Complete(r)
}

// findObjectsForNodeImage requests reconcile calls for every scaling group referencing the node image.
func (r *ScalingGroupReconciler) findObjectsForNodeImage(nodeImage client.Object) []reconcile.Request {
	attachedScalingGroups := &updatev1alpha1.ScalingGroupList{}
	listOps := &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(nodeImageField, nodeImage.GetName()),
	}
	if err := r.List(context.TODO(), attachedScalingGroups, listOps); err != nil {
		return []reconcile.Request{}
	}

	requests := make([]reconcile.Request, len(attachedScalingGroups.Items))
	for i, item := range attachedScalingGroups.Items {
		requests[i] = reconcile.Request{
			NamespacedName: types.NamespacedName{Name: item.GetName()},
		}
	}
	return requests
}

type scalingGroupUpdater interface {
	GetScalingGroupImage(ctx context.Context, scalingGroupID string) (string, error)
	SetScalingGroupImage(ctx context.Context, scalingGroupID, imageURI string) error
}
