/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package controllers

import (
	"context"
	"reflect"
	"strings"
	"time"

	nodeutil "github.com/edgelesssys/constellation/operators/constellation-node-operator/internal/node"
	"github.com/edgelesssys/constellation/operators/constellation-node-operator/internal/patch"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ref "k8s.io/client-go/tools/reference"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/source"

	updatev1alpha1 "github.com/edgelesssys/constellation/operators/constellation-node-operator/api/v1alpha1"
	nodemaintenancev1beta1 "github.com/medik8s/node-maintenance-operator/api/v1beta1"
)

const (
	// nodeOverprovisionLimit is the maximum number of extra nodes created during the update procedure at any point in time.
	nodeOverprovisionLimit = 1
	// nodeJoinTimeout is the time limit pending nodes have to join the cluster before being terminated.
	nodeJoinTimeout = time.Minute * 30
	// nodeLeaveTimeout is the time limit pending nodes have to leave the cluster and being terminated.
	nodeLeaveTimeout                   = time.Minute
	donorAnnotation                    = "constellation.edgeless.systems/donor"
	heirAnnotation                     = "constellation.edgeless.systems/heir"
	scalingGroupAnnotation             = "constellation.edgeless.systems/scaling-group-id"
	nodeImageAnnotation                = "constellation.edgeless.systems/node-image"
	obsoleteAnnotation                 = "constellation.edgeless.systems/obsolete"
	conditionNodeImageUpToDateReason   = "NodeImagesUpToDate"
	conditionNodeImageUpToDateMessage  = "Node image of every node is up to date"
	conditionNodeImageOutOfDateReason  = "NodeImagesOutOfDate"
	conditionNodeImageOutOfDateMessage = "Some node images are out of date"
)

// NodeImageReconciler reconciles a NodeImage object.
type NodeImageReconciler struct {
	nodeReplacer
	etcdRemover
	client.Client
	Scheme *runtime.Scheme
}

// NewNodeImageReconciler creates a new NodeImageReconciler.
func NewNodeImageReconciler(nodeReplacer nodeReplacer, etcdRemover etcdRemover, client client.Client, scheme *runtime.Scheme) *NodeImageReconciler {
	return &NodeImageReconciler{
		nodeReplacer: nodeReplacer,
		etcdRemover:  etcdRemover,
		Client:       client,
		Scheme:       scheme,
	}
}

//+kubebuilder:rbac:groups=update.edgeless.systems,resources=nodeimages,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=update.edgeless.systems,resources=nodeimages/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=update.edgeless.systems,resources=nodeimages/finalizers,verbs=update
//+kubebuilder:rbac:groups=nodemaintenance.medik8s.io,resources=nodemaintenances,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=nodes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=nodes/status,verbs=get

// Reconcile replaces outdated nodes (using an old image) with new nodes (using a new image) as specified in the NodeImage spec.
func (r *NodeImageReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logr := log.FromContext(ctx)
	logr.Info("Reconciling NodeImage")

	var desiredNodeImage updatev1alpha1.NodeImage
	if err := r.Get(ctx, req.NamespacedName, &desiredNodeImage); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	// get list of autoscaling strategies
	// there should be exactly one autoscaling strategy but we do not specify its name.
	// if there is no autoscaling strategy, it is assumed that autoscaling is disabled.
	var autoscalingStrategiesList updatev1alpha1.AutoscalingStrategyList
	if err := r.List(ctx, &autoscalingStrategiesList); err != nil {
		return ctrl.Result{}, err
	}
	var autoscalingEnabled bool
	for _, autoscalingStrategy := range autoscalingStrategiesList.Items {
		if autoscalingStrategy.Status.Enabled {
			autoscalingEnabled = true
			break
		}
	}

	// get list of all nodes
	var nodeList corev1.NodeList
	if err := r.List(ctx, &nodeList); err != nil {
		logr.Error(err, "Unable to list nodes")
		return ctrl.Result{}, err
	}
	// get list of all pending nodes
	var pendingNodeList updatev1alpha1.PendingNodeList
	if err := r.List(ctx, &pendingNodeList, client.InNamespace(req.Namespace)); err != nil {
		logr.Error(err, "Unable to list pending nodes")
		return ctrl.Result{}, err
	}
	// get list of all scaling groups
	var scalingGroupList updatev1alpha1.ScalingGroupList
	if err := r.List(ctx, &scalingGroupList, client.InNamespace(req.Namespace)); err != nil {
		logr.Error(err, "Unable to list scaling groups")
		return ctrl.Result{}, err
	}
	scalingGroupByID := make(map[string]updatev1alpha1.ScalingGroup, len(scalingGroupList.Items))
	for _, scalingGroup := range scalingGroupList.Items {
		scalingGroupByID[strings.ToLower(scalingGroup.Spec.GroupID)] = scalingGroup
	}
	annotatedNodes, invalidNodes := r.annotateNodes(ctx, nodeList.Items)
	groups := groupNodes(annotatedNodes, pendingNodeList.Items, desiredNodeImage.Spec.ImageReference)

	logr.Info("Grouped nodes",
		"outdatedNodes", len(groups.Outdated),
		"upToDateNodes", len(groups.UpToDate),
		"donorNodes", len(groups.Donors),
		"heirNodes", len(groups.Heirs),
		"mintNodes", len(groups.Mint),
		"pendingNodes", len(pendingNodeList.Items),
		"obsoleteNodes", len(groups.Obsolete),
		"invalidNodes", len(invalidNodes))

	// extraNodes are nodes that exist in the scaling group which cannot be used for regular workloads.
	// consists of nodes that are
	// - being created (joining)
	// - being destroyed (leaving)
	// - heirs to outdated nodes
	extraNodes := len(groups.Heirs) + len(pendingNodeList.Items)
	// newNodesBudget is the maximum number of new nodes that can be created in this Reconcile call.
	var newNodesBudget int
	if extraNodes < nodeOverprovisionLimit {
		newNodesBudget = nodeOverprovisionLimit - extraNodes
	}
	logr.Info("Budget for new nodes", "newNodesBudget", newNodesBudget)

	status := nodeImageStatus(r.Scheme, groups, pendingNodeList.Items, invalidNodes, newNodesBudget)
	if err := r.tryUpdateStatus(ctx, req.NamespacedName, status); err != nil {
		logr.Error(err, "Updating status")
	}

	allNodesUpToDate := len(groups.Outdated)+len(groups.Heirs)+len(pendingNodeList.Items)+len(groups.Obsolete) == 0
	if err := r.ensureAutoscaling(ctx, autoscalingEnabled, allNodesUpToDate); err != nil {
		logr.Error(err, "Ensure autoscaling", "autoscalingEnabledIs", autoscalingEnabled, "autoscalingEnabledWant", allNodesUpToDate)
		return ctrl.Result{}, err
	}

	if allNodesUpToDate {
		logr.Info("All node images up to date")
		return ctrl.Result{}, nil
	}

	// should requeue is set if a node is deleted
	var shouldRequeue bool
	// find pairs of mint nodes and outdated nodes in the same scaling group to become donor & heir
	replacementPairs := r.pairDonorsAndHeirs(ctx, &desiredNodeImage, groups.Outdated, groups.Mint)
	// extend replacement pairs to include existing pairs of donors and heirs
	replacementPairs = r.matchDonorsAndHeirs(ctx, replacementPairs, groups.Donors, groups.Heirs)
	// replace donor nodes by heirs
	for _, pair := range replacementPairs {
		logr.Info("Replacing node", "donorNode", pair.donor.Name, "heirNode", pair.heir.Name)
		done, err := r.replaceNode(ctx, &desiredNodeImage, pair)
		if err != nil {
			logr.Error(err, "Replacing node")
			return ctrl.Result{}, err
		}
		if done {
			shouldRequeue = true
			// remove donor annotation from heir
			if err := r.patchUnsetNodeAnnotations(ctx, pair.heir.Name, []string{donorAnnotation}); err != nil {
				logr.Error(err, "Unable to remove donor annotation from heir", "heirNode", pair.heir.Name)
			}
		}
	}

	// only create new nodes if the autoscaler is disabled.
	// otherwise, new nodes will also be created by the autoscaler
	if autoscalingEnabled {
		return ctrl.Result{Requeue: shouldRequeue}, nil
	}

	if err := r.createNewNodes(ctx, desiredNodeImage, groups.Outdated, pendingNodeList.Items, scalingGroupByID, newNodesBudget); err != nil {
		return ctrl.Result{Requeue: shouldRequeue}, nil
	}
	// cleanup obsolete nodes
	for _, node := range groups.Obsolete {
		done, err := r.deleteNode(ctx, &desiredNodeImage, node)
		if err != nil {
			logr.Error(err, "Unable to remove obsolete node")
		}
		if done {
			shouldRequeue = true
		}
	}

	return ctrl.Result{Requeue: shouldRequeue}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NodeImageReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&updatev1alpha1.NodeImage{}).
		Watches(
			&source.Kind{Type: &updatev1alpha1.ScalingGroup{}},
			handler.EnqueueRequestsFromMapFunc(r.findObjectsForScalingGroup),
			builder.WithPredicates(scalingGroupImageChangedPredicate()),
		).
		Watches(
			&source.Kind{Type: &updatev1alpha1.AutoscalingStrategy{}},
			handler.EnqueueRequestsFromMapFunc(r.findAllNodeImages),
			builder.WithPredicates(autoscalerEnabledStatusChangedPredicate()),
		).
		Watches(
			&source.Kind{Type: &corev1.Node{}},
			handler.EnqueueRequestsFromMapFunc(r.findAllNodeImages),
			builder.WithPredicates(nodeReadyPredicate()),
		).
		Watches(
			&source.Kind{Type: &nodemaintenancev1beta1.NodeMaintenance{}},
			handler.EnqueueRequestsFromMapFunc(r.findAllNodeImages),
			builder.WithPredicates(nodeMaintenanceSucceededPredicate()),
		).
		Owns(&updatev1alpha1.PendingNode{}).
		Complete(r)
}

// annotateNodes takes all nodes of the cluster and annotates them with the scaling group they are in and the image they are using.
func (r *NodeImageReconciler) annotateNodes(ctx context.Context, nodes []corev1.Node) (annotatedNodes, invalidNodes []corev1.Node) {
	logr := log.FromContext(ctx)
	for _, node := range nodes {
		annotations := make(map[string]string)
		if node.Spec.ProviderID == "" {
			logr.Info("Node is missing providerID", "invalidNode", node.Name)
			invalidNodes = append(invalidNodes, node)
			continue
		}
		if _, ok := node.Annotations[scalingGroupAnnotation]; !ok {
			scalingGroupID, err := r.nodeReplacer.GetScalingGroupID(ctx, node.Spec.ProviderID)
			if err != nil {
				logr.Error(err, "Unable to get node scaling group")
				invalidNodes = append(invalidNodes, node)
				continue
			}
			annotations[scalingGroupAnnotation] = scalingGroupID
		}
		if _, ok := node.Annotations[nodeImageAnnotation]; !ok {
			nodeImage, err := r.nodeReplacer.GetNodeImage(ctx, node.Spec.ProviderID)
			if err != nil {
				logr.Error(err, "Unable to get node image")
				invalidNodes = append(invalidNodes, node)
				continue
			}
			annotations[nodeImageAnnotation] = nodeImage
		}
		if len(annotations) > 0 {
			if err := r.patchNodeAnnotations(ctx, node.Name, annotations); err != nil {
				logr.Error(err, "Unable to patch node annotations")
				invalidNodes = append(invalidNodes, node)
				continue
			}
			if err := r.Get(ctx, types.NamespacedName{Name: node.Name}, &node); err != nil {
				logr.Error(err, "Unable to get patched node")
				invalidNodes = append(invalidNodes, node)
				continue
			}
		}
		annotatedNodes = append(annotatedNodes, node)
	}
	return annotatedNodes, invalidNodes
}

// pairDonorsAndHeirs takes a list of outdated nodes (that do not yet have a heir node) and a list of mint nodes (nodes using the latest image) and pairs matching nodes to become donor and heir.
// outdatedNodes is also updated with heir annotations.
func (r *NodeImageReconciler) pairDonorsAndHeirs(ctx context.Context, controller metav1.Object, outdatedNodes []corev1.Node, mintNodes []mintNode) []replacementPair {
	logr := log.FromContext(ctx)
	var pairs []replacementPair
	for _, mintNode := range mintNodes {
		var foundReplacement bool
		// find outdated node in the same group
		for i := range outdatedNodes {
			outdatedNode := &outdatedNodes[i]
			if !strings.EqualFold(outdatedNode.Annotations[scalingGroupAnnotation], mintNode.pendingNode.Spec.ScalingGroupID) || len(outdatedNode.Annotations[heirAnnotation]) != 0 {
				continue
			}
			// mark as donor <-> heir pair and delete "pending node" resource
			if err := r.patchNodeAnnotations(ctx, mintNode.node.Name, map[string]string{donorAnnotation: outdatedNode.Name}); err != nil {
				logr.Error(err, "Unable to update mint node donor annotation", "mintNode", mintNode.node.Name)
				break
			}
			if mintNode.node.Annotations == nil {
				mintNode.node.Annotations = make(map[string]string)
			}
			mintNode.node.Annotations[donorAnnotation] = outdatedNode.Name
			if err := r.patchNodeAnnotations(ctx, outdatedNode.Name, map[string]string{heirAnnotation: mintNode.node.Name}); err != nil {
				logr.Error(err, "Unable to update outdated node heir annotation", "outdatedNode", outdatedNode.Name)
				break
			}
			outdatedNode.Annotations[heirAnnotation] = mintNode.node.Name
			if err := r.Delete(ctx, &mintNode.pendingNode); err != nil {
				logr.Error(err, "Unable to delete pending node resource", "pendingNode", mintNode.pendingNode.Name)
				break
			}
			pairs = append(pairs, replacementPair{
				donor: *outdatedNode,
				heir:  mintNode.node,
			})
			logr.Info("New matched up pair", "donorNode", outdatedNode.Name, "heirNode", mintNode.node.Name)
			foundReplacement = true
			break
		}
		if !foundReplacement {
			logr.Info("No replacement found for mint node. Marking as outdated.", "mintNode", mintNode.node.Name, "scalingGroupID", mintNode.pendingNode.Spec.ScalingGroupID)
			// mint node was not needed as heir. Cleanup obsolete resources.
			if err := r.Delete(ctx, &mintNode.pendingNode); err != nil {
				logr.Error(err, "Unable to delete pending node resource", "pendingNode", mintNode.pendingNode.Name)
				break
			}
			if err := r.patchNodeAnnotations(ctx, mintNode.node.Name, map[string]string{obsoleteAnnotation: "true"}); err != nil {
				logr.Error(err, "Unable to update mint node obsolete annotation", "mintNode", mintNode.node.Name)
				break
			}
			if _, err := r.deleteNode(ctx, controller, mintNode.node); err != nil {
				logr.Error(err, "Unable to delete obsolete node", "obsoleteNode", mintNode.node.Name)
				break
			}
		}
	}
	return pairs
}

// matchDonorsAndHeirs takes separate lists of donors and heirs and matches each heir to its previously chosen donor.
// a list of replacement pairs is returned.
// donors and heirs with invalid pair references are cleaned up (the donor/heir annotations gets removed).
func (r *NodeImageReconciler) matchDonorsAndHeirs(ctx context.Context, pairs []replacementPair, donors, heirs []corev1.Node) []replacementPair {
	logr := log.FromContext(ctx)
	for _, heir := range heirs {
		var foundPair bool
		for _, donor := range donors {
			if heir.Annotations[donorAnnotation] == donor.Name {
				pairs = append(pairs, replacementPair{
					donor: donor,
					heir:  heir,
				})
				foundPair = true
				break
			}
		}
		if !foundPair {
			// remove donor annotation from heir
			if err := r.patchUnsetNodeAnnotations(ctx, heir.Name, []string{donorAnnotation}); err != nil {
				logr.Error(err, "Unable to remove donor annotation from heir", "heirNode", heir.Name)
			}
			delete(heir.Annotations, donorAnnotation)
		}
	}
	// iterate over all donors and remove donor annotation from nodes that are not in a pair
	// (cleanup)
	for _, donor := range donors {
		var foundPair bool
		for _, pair := range pairs {
			if pair.donor.Name == donor.Name {
				foundPair = true
				break
			}
		}
		if !foundPair {
			// remove heir annotation from donor
			if err := r.patchUnsetNodeAnnotations(ctx, donor.Name, []string{heirAnnotation}); err != nil {
				logr.Error(err, "Unable to remove heir annotation from donor", "donorNode", donor.Name)
			}
			delete(donor.Annotations, heirAnnotation)
		}
	}
	return pairs
}

// ensureAutoscaling will ensure that the autoscaling is enabled or disabled as needed.
func (r *NodeImageReconciler) ensureAutoscaling(ctx context.Context, autoscalingEnabled bool, wantAutoscalingEnabled bool) error {
	if autoscalingEnabled == wantAutoscalingEnabled {
		return nil
	}
	var autoscalingStrategiesList updatev1alpha1.AutoscalingStrategyList
	if err := r.List(ctx, &autoscalingStrategiesList); err != nil {
		return err
	}
	for i := range autoscalingStrategiesList.Items {
		if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			var autoscalingStrategy updatev1alpha1.AutoscalingStrategy
			if err := r.Get(ctx, types.NamespacedName{Name: autoscalingStrategiesList.Items[i].Name}, &autoscalingStrategy); err != nil {
				return err
			}
			autoscalingStrategy.Spec.Enabled = wantAutoscalingEnabled
			return r.Client.Update(ctx, &autoscalingStrategy)
		}); err != nil {
			return err
		}
	}
	return nil
}

// replaceNode take a donor and a heir node and then replaces the donor node by the heir node.
//
// Replacing nodes involves the following steps:
// Labels are copied from the donor node to the heir node.
// Readiness of the heir node is awaited.
// Deletion of the donor node is scheduled.
func (r *NodeImageReconciler) replaceNode(ctx context.Context, controller metav1.Object, pair replacementPair) (bool, error) {
	logr := log.FromContext(ctx)
	if !reflect.DeepEqual(nodeutil.FilterLabels(pair.donor.Labels), nodeutil.FilterLabels(pair.heir.Labels)) {
		if err := r.copyNodeLabels(ctx, pair.donor.Name, pair.heir.Name); err != nil {
			logr.Error(err, "Copy node labels")
			return false, err
		}
	}
	heirReady := nodeutil.Ready(&pair.heir)
	if !heirReady {
		return false, nil
	}
	return r.deleteNode(ctx, controller, pair.donor)
}

// deleteNode safely removes a node from the cluster and issues termination of the node by the CSP.
func (r *NodeImageReconciler) deleteNode(ctx context.Context, controller metav1.Object, node corev1.Node) (bool, error) {
	logr := log.FromContext(ctx)
	// cordon & drain node using node-maintenance-operator
	var foundNodeMaintenance nodemaintenancev1beta1.NodeMaintenance
	err := r.Get(ctx, types.NamespacedName{Name: node.Name}, &foundNodeMaintenance)
	if client.IgnoreNotFound(err) != nil {
		// unexpected error occurred
		return false, err
	}
	if err != nil {
		// NodeMaintenance resource does not exist yet
		nodeMaintenance := nodemaintenancev1beta1.NodeMaintenance{
			ObjectMeta: metav1.ObjectMeta{
				Name: node.Name,
			},
			Spec: nodemaintenancev1beta1.NodeMaintenanceSpec{
				NodeName: node.Name,
				Reason:   "node is replaced due to OS image update",
			},
		}
		return false, r.Create(ctx, &nodeMaintenance)
	}

	// NodeMaintenance resource already exists. Check cordon & drain status.
	if foundNodeMaintenance.Status.Phase != nodemaintenancev1beta1.MaintenanceSucceeded {
		logr.Info("Cordon & drain in progress", "maintenanceNode", node.Name, "nodeMaintenanceStatus", foundNodeMaintenance.Status.Phase)
		return false, nil
	}

	// node is unused & ready to be replaced
	if nodeutil.IsControlPlaneNode(&node) {
		nodeVPCIP, err := nodeutil.VPCIP(&node)
		if err != nil {
			logr.Error(err, "Unable to get node VPC IP")
			return false, err
		}
		if err := r.RemoveEtcdMemberFromCluster(ctx, nodeVPCIP); err != nil {
			logr.Error(err, "Unable to remove etcd member from cluster")
			return false, err
		}
	}
	if err := r.Delete(ctx, &node); err != nil {
		logr.Error(err, "Deleting node")
		return false, err
	}

	logr.Info("Deleted node", "deletedNode", node.Name)
	// schedule deletion of the node with the CSP
	if err := r.DeleteNode(ctx, node.Spec.ProviderID); err != nil {
		logr.Error(err, "Scheduling CSP node deletion", "providerID", node.Spec.ProviderID)
	}
	deadline := metav1.NewTime(time.Now().Add(nodeLeaveTimeout))
	pendingNode := updatev1alpha1.PendingNode{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: controller.GetNamespace(),
			Name:      node.Name,
		},
		Spec: updatev1alpha1.PendingNodeSpec{
			ProviderID:     node.Spec.ProviderID,
			ScalingGroupID: node.Annotations[scalingGroupAnnotation],
			NodeName:       node.Name,
			Goal:           updatev1alpha1.NodeGoalLeave,
			Deadline:       &deadline,
		},
	}
	if err := ctrl.SetControllerReference(controller, &pendingNode, r.Scheme); err != nil {
		return false, err
	}
	if err := r.Create(ctx, &pendingNode); err != nil {
		logr.Error(err, "Tracking CSP node deletion")
	}
	return true, nil
}

// createNewNodes creates new nodes using up to date images as replacement for outdated nodes.
func (r *NodeImageReconciler) createNewNodes(
	ctx context.Context, desiredNodeImage updatev1alpha1.NodeImage,
	outdatedNodes []corev1.Node, pendingNodes []updatev1alpha1.PendingNode,
	scalingGroupByID map[string]updatev1alpha1.ScalingGroup, newNodesBudget int,
) error {
	logr := log.FromContext(ctx)
	if newNodesBudget < 1 || len(outdatedNodes) == 0 {
		return nil
	}
	outdatedNodesPerScalingGroup := make(map[string]int)
	for _, node := range outdatedNodes {
		// skip outdated nodes that got assigned an heir in this Reconcile call
		if len(node.Annotations[heirAnnotation]) != 0 {
			continue
		}
		outdatedNodesPerScalingGroup[strings.ToLower(node.Annotations[scalingGroupAnnotation])]++
	}
	pendingJoiningNodesPerScalingGroup := make(map[string]int)
	for _, pendingNode := range pendingNodes {
		// skip pending nodes that are not joining
		if pendingNode.Spec.Goal != updatev1alpha1.NodeGoalJoin {
			continue
		}
		pendingJoiningNodesPerScalingGroup[strings.ToLower(pendingNode.Spec.ScalingGroupID)]++
	}
	requiredNodesPerScalingGroup := make(map[string]int, len(outdatedNodesPerScalingGroup))
	for scalingGroupID := range outdatedNodesPerScalingGroup {
		scalingGroupID := strings.ToLower(scalingGroupID)
		if pendingJoiningNodesPerScalingGroup[scalingGroupID] < outdatedNodesPerScalingGroup[scalingGroupID] {
			requiredNodesPerScalingGroup[scalingGroupID] = outdatedNodesPerScalingGroup[scalingGroupID] - pendingJoiningNodesPerScalingGroup[scalingGroupID]
		}
	}
	for scalingGroupID := range requiredNodesPerScalingGroup {
		scalingGroup, ok := scalingGroupByID[scalingGroupID]
		if !ok {
			logr.Info("Scaling group does not have matching resource", "scalingGroup", scalingGroupID, "scalingGroups", scalingGroupByID)
			continue
		}
		if !strings.EqualFold(scalingGroup.Status.ImageReference, desiredNodeImage.Spec.ImageReference) {
			logr.Info("Scaling group does not use latest image", "scalingGroup", scalingGroupID, "usedImage", scalingGroup.Status.ImageReference, "wantedImage", desiredNodeImage.Spec.ImageReference)
			continue
		}
		if requiredNodesPerScalingGroup[scalingGroupID] == 0 {
			continue
		}
		for {
			if newNodesBudget == 0 {
				return nil
			}
			if requiredNodesPerScalingGroup[scalingGroupID] == 0 {
				break
			}
			logr.Info("Creating new node", "scalingGroup", scalingGroupID)
			nodeName, providerID, err := r.CreateNode(ctx, scalingGroup.Spec.GroupID)
			if err != nil {
				return err
			}
			deadline := metav1.NewTime(time.Now().Add(nodeJoinTimeout))
			pendingNode := &updatev1alpha1.PendingNode{
				ObjectMeta: metav1.ObjectMeta{Name: nodeName},
				Spec: updatev1alpha1.PendingNodeSpec{
					ProviderID:     providerID,
					ScalingGroupID: scalingGroup.Spec.GroupID,
					NodeName:       nodeName,
					Goal:           updatev1alpha1.NodeGoalJoin,
					Deadline:       &deadline,
				},
			}
			if err := ctrl.SetControllerReference(&desiredNodeImage, pendingNode, r.Scheme); err != nil {
				return err
			}
			if err := r.Create(ctx, pendingNode); err != nil {
				return err
			}
			logr.Info("Created new node", "createdNode", nodeName, "scalingGroup", scalingGroupID)
			requiredNodesPerScalingGroup[scalingGroupID]--
			newNodesBudget--
		}
	}
	return nil
}

// patchNodeAnnotations attempts to patch node annotations in a retry loop.
func (r *NodeImageReconciler) patchNodeAnnotations(ctx context.Context, nodeName string, annotations map[string]string) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		var node corev1.Node
		if err := r.Get(ctx, types.NamespacedName{Name: nodeName}, &node); err != nil {
			return err
		}
		patchedNode := node.DeepCopy()
		patch := patch.SetAnnotations(&node, patchedNode, annotations)
		return r.Client.Patch(ctx, patchedNode, patch)
	})
}

// patchNodeAnnotations attempts to remove node annotations using a patch in a retry loop.
func (r *NodeImageReconciler) patchUnsetNodeAnnotations(ctx context.Context, nodeName string, annotationKeys []string) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		var node corev1.Node
		if err := r.Get(ctx, types.NamespacedName{Name: nodeName}, &node); err != nil {
			return err
		}
		patchedNode := node.DeepCopy()
		patch := patch.UnsetAnnotations(&node, patchedNode, annotationKeys)
		return r.Client.Patch(ctx, patchedNode, patch)
	})
}

// copyNodeLabels attempts to copy all node labels (except for reserved labels) from one node to another in a retry loop.
func (r *NodeImageReconciler) copyNodeLabels(ctx context.Context, oldNodeName, newNodeName string) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		var oldNode corev1.Node
		if err := r.Get(ctx, types.NamespacedName{Name: oldNodeName}, &oldNode); err != nil {
			return err
		}
		var newNode corev1.Node
		if err := r.Get(ctx, types.NamespacedName{Name: newNodeName}, &newNode); err != nil {
			return err
		}
		patchedNode := newNode.DeepCopy()
		patch := patch.SetLabels(&newNode, patchedNode, nodeutil.FilterLabels(oldNode.GetLabels()))
		return r.Client.Patch(ctx, patchedNode, patch)
	})
}

// tryUpdateStatus attempts to update the NodeImage status field in a retry loop.
func (r *NodeImageReconciler) tryUpdateStatus(ctx context.Context, name types.NamespacedName, status updatev1alpha1.NodeImageStatus) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		var nodeImage updatev1alpha1.NodeImage
		if err := r.Get(ctx, name, &nodeImage); err != nil {
			return err
		}
		nodeImage.Status = *status.DeepCopy()
		if err := r.Status().Update(ctx, &nodeImage); err != nil {
			return err
		}
		return nil
	})
}

// nodeImageStatus generates the NodeImage.Status field given node groups and the budget for new nodes.
func nodeImageStatus(scheme *runtime.Scheme, groups nodeGroups, pendingNodes []updatev1alpha1.PendingNode, invalidNodes []corev1.Node, newNodesBudget int) updatev1alpha1.NodeImageStatus {
	var status updatev1alpha1.NodeImageStatus
	outdatedCondition := metav1.Condition{
		Type: updatev1alpha1.ConditionOutdated,
	}
	if len(groups.Outdated)+len(groups.Heirs)+len(pendingNodes)+len(groups.Obsolete) == 0 {
		outdatedCondition.Status = metav1.ConditionFalse
		outdatedCondition.Reason = conditionNodeImageUpToDateReason
		outdatedCondition.Message = conditionNodeImageUpToDateMessage
	} else {
		outdatedCondition.Status = metav1.ConditionTrue
		outdatedCondition.Reason = conditionNodeImageOutOfDateReason
		outdatedCondition.Message = conditionNodeImageOutOfDateMessage
	}
	meta.SetStatusCondition(&status.Conditions, outdatedCondition)
	for _, node := range groups.Outdated {
		nodeRef, err := ref.GetReference(scheme, &node)
		if err != nil {
			continue
		}
		status.Outdated = append(status.Outdated, *nodeRef)
	}
	for _, node := range groups.UpToDate {
		nodeRef, err := ref.GetReference(scheme, &node)
		if err != nil {
			continue
		}
		status.UpToDate = append(status.UpToDate, *nodeRef)
	}
	for _, node := range groups.Donors {
		nodeRef, err := ref.GetReference(scheme, &node)
		if err != nil {
			continue
		}
		status.Donors = append(status.Donors, *nodeRef)
	}
	for _, node := range groups.Heirs {
		nodeRef, err := ref.GetReference(scheme, &node)
		if err != nil {
			continue
		}
		status.Heirs = append(status.Heirs, *nodeRef)
	}
	for _, node := range groups.Obsolete {
		nodeRef, err := ref.GetReference(scheme, &node)
		if err != nil {
			continue
		}
		status.Obsolete = append(status.Obsolete, *nodeRef)
	}
	for _, node := range invalidNodes {
		nodeRef, err := ref.GetReference(scheme, &node)
		if err != nil {
			continue
		}
		status.Invalid = append(status.Invalid, *nodeRef)
	}
	for _, mintN := range groups.Mint {
		nodeRef, err := ref.GetReference(scheme, &mintN.node)
		if err != nil {
			continue
		}
		status.Mints = append(status.Mints, *nodeRef)
	}
	for _, pending := range pendingNodes {
		pendingRef, err := ref.GetReference(scheme, &pending)
		if err != nil {
			continue
		}
		status.Pending = append(status.Pending, *pendingRef)
	}
	status.Budget = uint32(newNodesBudget)
	return status
}

// mintNode is a pair of a freshly joined kubernetes nodes
// and the corresponding (left over) pending node resource.
type mintNode struct {
	node        corev1.Node
	pendingNode updatev1alpha1.PendingNode
}

// replacementPair is a pair of a donor (outdated node that should be replaced)
// and a heir (up to date node that inherits node labels)
type replacementPair struct {
	donor corev1.Node
	heir  corev1.Node
}

// nodeGroups is a collection of disjoint sets of nodes.
// every properly annotated kubernetes node can be placed in exactly one of the sets.
type nodeGroups struct {
	// Outdated nodes are nodes that
	// do not use the most recent image AND
	// are not yet a donor to an up to date heir node
	Outdated,
	// UpToDate nodes are nodes that
	// use the most recent image,
	// are not an heir to an outdated donor node AND
	// are not mint nodes
	UpToDate,
	// Donors are nodes that
	// do not use the most recent image AND
	// are paired up with an up to date heir node
	Donors,
	// Heirs are nodes that
	// use the most recent image AND
	// are paired up with an outdated donor node
	Heirs,
	// Obsolete nodes are nodes that
	// were created by the operator as replacements (heirs)
	// but could not get paired up with a donor node.
	// They will be cleaned up by the operator.
	Obsolete []corev1.Node
	// Mint nodes are nodes that
	// use the most recent image AND
	// were created by the operator as replacements (heirs)
	// and are awaiting pairing up with a donor node.
	Mint []mintNode
}

// groupNodes classifies nodes by placing each into exactly one group.
func groupNodes(nodes []corev1.Node, pendingNodes []updatev1alpha1.PendingNode, latestImageReference string) nodeGroups {
	groups := nodeGroups{}
	for _, node := range nodes {
		if node.Annotations[obsoleteAnnotation] == "true" {
			groups.Obsolete = append(groups.Obsolete, node)
			continue
		}
		if !strings.EqualFold(node.Annotations[nodeImageAnnotation], latestImageReference) {
			if heir := node.Annotations[heirAnnotation]; heir != "" {
				groups.Donors = append(groups.Donors, node)
			} else {
				groups.Outdated = append(groups.Outdated, node)
			}
			continue
		}
		if donor := node.Annotations[donorAnnotation]; donor != "" {
			groups.Heirs = append(groups.Heirs, node)
			continue
		}
		if pendingNode := nodeutil.FindPending(pendingNodes, &node); pendingNode != nil {
			groups.Mint = append(groups.Mint, mintNode{
				node:        node,
				pendingNode: *pendingNode,
			})
			continue
		}
		groups.UpToDate = append(groups.UpToDate, node)
	}
	return groups
}

type nodeReplacer interface {
	// GetNodeImage retrieves the image currently used by a node.
	GetNodeImage(ctx context.Context, providerID string) (string, error)
	// GetScalingGroupID retrieves the scaling group that a node is part of.
	GetScalingGroupID(ctx context.Context, providerID string) (string, error)
	// CreateNode creates a new node inside a specified scaling group at the CSP and returns its future name and provider id.
	CreateNode(ctx context.Context, scalingGroupID string) (nodeName, providerID string, err error)
	// DeleteNode starts the termination of the node at the CSP.
	DeleteNode(ctx context.Context, providerID string) error
}

type etcdRemover interface {
	// RemoveEtcdMemberFromCluster removes an etcd member from the cluster.
	RemoveEtcdMemberFromCluster(ctx context.Context, vpcIP string) error
}
