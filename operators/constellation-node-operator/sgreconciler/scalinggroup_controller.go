/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package sgreconciler

import (
	"context"
	"strings"

	mainconstants "github.com/edgelesssys/constellation/v2/internal/constants"
	updatev1alpha1 "github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/api/v1alpha1"
	cspapi "github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/internal/cloud/api"
	"github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/internal/executor"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	// defaultScalingGroupMin is the default minimum number of nodes in a scaling group.
	// This value is used if the scaling group is created by the operator.
	// If a user modifies the scaling group, the operator will not overwrite the user's configuration.
	defaultScalingGroupMin = 1
	// defaultScalingGroupMax is the default maximum number of nodes in a scaling group.
	// This value is used if the scaling group is created by the operator.
	// If a user modifies the scaling group, the operator will not overwrite the user's configuration.
	defaultScalingGroupMax = 10
)

// ExternalScalingGroupReconciler reconciles on scaling groups in CSP infrastructure.
// It does NOT reconcile on k8s resources.
// Instead, it scans the cloud provider infrastructure and changes k8s resources accordingly.
type ExternalScalingGroupReconciler struct {
	// uid is the unique identifier of the Constellation cluster.
	uid                    string
	scalingGroupDiscoverer scalingGroupDiscoverer
	k8sClient              k8sReadWriter
}

// NewExternalScalingGroupReconciler creates a new InfrastructureReconciler.
func NewExternalScalingGroupReconciler(uid string, discoverer scalingGroupDiscoverer, k8sClient k8sReadWriter) *ExternalScalingGroupReconciler {
	return &ExternalScalingGroupReconciler{
		uid:                    uid,
		scalingGroupDiscoverer: discoverer,
		k8sClient:              k8sClient,
	}
}

// Reconcile reconciles on scaling groups in CSP infrastructure.
func (r *ExternalScalingGroupReconciler) Reconcile(ctx context.Context) (executor.Result, error) {
	logr := log.FromContext(ctx)
	logr.Info("reconciling external scaling groups")

	nodeGroups, err := r.scalingGroupDiscoverer.ListScalingGroups(ctx, r.uid)
	if err != nil {
		return executor.Result{}, err
	}

	existingNodeGroups := map[string]struct{}{}

	// create all scaling groups that are newly discovered
	for _, group := range nodeGroups {
		exists, err := patchNodeGroupName(ctx, r.k8sClient, group.Name, group.NodeGroupName)
		if err != nil {
			return executor.Result{}, err
		}
		if exists {
			// scaling group already exists
			existingNodeGroups[group.Name] = struct{}{}
			continue
		}
		err = createScalingGroupIfNotExists(ctx, newScalingGroupConfig{
			k8sClient:            r.k8sClient,
			resourceName:         group.Name,
			groupID:              group.GroupID,
			nodeGroupName:        group.NodeGroupName,
			autoscalingGroupName: group.AutoscalingGroupName,
			role:                 group.Role,
		})
		if err != nil {
			return executor.Result{}, err
		}
		existingNodeGroups[group.Name] = struct{}{}
	}

	logr.Info("ensured scaling groups are created", "count", len(nodeGroups))

	// delete all scaling groups that no longer exist
	var scalingGroups updatev1alpha1.ScalingGroupList
	if err := r.k8sClient.List(ctx, &scalingGroups); err != nil {
		return executor.Result{}, err
	}
	for _, group := range scalingGroups.Items {
		if _, ok := existingNodeGroups[group.Name]; !ok {
			logr.Info("deleting scaling group", "name", group.Name)
			err := r.k8sClient.Delete(ctx, &group)
			if err != nil {
				return executor.Result{}, err
			}
		}
	}

	logr.Info("external scaling groups reconciled")

	return executor.Result{}, nil
}

// patchNodeGroupName patches the node group name of a scaling group resource (if necessary and it exists).
func patchNodeGroupName(ctx context.Context, k8sClient k8sReadWriter, resourceName, nodeGroupName string) (exists bool, err error) {
	logr := log.FromContext(ctx)
	var scalingGroup updatev1alpha1.ScalingGroup
	err = k8sClient.Get(ctx, client.ObjectKey{Name: resourceName}, &scalingGroup)
	if k8sErrors.IsNotFound(err) {
		// scaling group does not exist
		// no need to patch
		return false /* doesn't exist */, nil
	}
	if err != nil {
		return false, err
	}
	if scalingGroup.Spec.NodeGroupName == nodeGroupName {
		// scaling group already has the correct node group name
		return true /* exists */, nil
	}
	logr.Info("patching node group name", "resourceName", resourceName, "nodeGroupName", nodeGroupName)
	return true, retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if err := k8sClient.Get(ctx, client.ObjectKey{Name: resourceName}, &scalingGroup); err != nil {
			return err
		}
		scalingGroup.Spec.NodeGroupName = nodeGroupName
		return k8sClient.Update(ctx, &scalingGroup)
	})
}

func createScalingGroupIfNotExists(ctx context.Context, config newScalingGroupConfig) error {
	logr := log.FromContext(ctx)
	err := config.k8sClient.Create(ctx, &updatev1alpha1.ScalingGroup{
		TypeMeta: metav1.TypeMeta{APIVersion: "update.edgeless.systems/v1alpha1", Kind: "ScalingGroup"},
		ObjectMeta: metav1.ObjectMeta{
			Name: strings.ToLower(config.resourceName),
		},
		Spec: updatev1alpha1.ScalingGroupSpec{
			NodeVersion:         mainconstants.NodeVersionResourceName,
			GroupID:             config.groupID,
			AutoscalerGroupName: config.autoscalingGroupName,
			NodeGroupName:       config.nodeGroupName,
			Min:                 defaultScalingGroupMin,
			Max:                 defaultScalingGroupMax,
			Role:                config.role,
		},
	})
	if k8sErrors.IsAlreadyExists(err) {
		return nil
	} else if err == nil {
		logr.Info("created scaling group", "name", config.resourceName, "nodeGroupName", config.nodeGroupName)
	} else {
		logr.Error(err, "failed to create scaling group", "name", config.resourceName, "nodeGroupName", config.nodeGroupName)
	}
	return err
}

// scalingGroupDiscoverer is used to discover scaling groups in the cloud provider infrastructure.
type scalingGroupDiscoverer interface {
	ListScalingGroups(ctx context.Context, uid string,
	) ([]cspapi.ScalingGroup, error)
}

type k8sReadWriter interface {
	client.Reader
	client.Writer
}

type newScalingGroupConfig struct {
	k8sClient            client.Writer
	resourceName         string
	groupID              string
	nodeGroupName        string
	autoscalingGroupName string
	role                 updatev1alpha1.NodeRole
}
