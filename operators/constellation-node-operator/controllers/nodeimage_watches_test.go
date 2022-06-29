package controllers

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	updatev1alpha1 "github.com/edgelesssys/constellation/operators/constellation-node-operator/api/v1alpha1"
	nodemaintenancev1beta1 "github.com/medik8s/node-maintenance-operator/api/v1beta1"
)

func TestScalingGroupImageChangedPredicate(t *testing.T) {
	testCases := map[string]struct {
		event          event.UpdateEvent
		wantProcessing bool
	}{
		"old object is not a scaling group": {
			event: event.UpdateEvent{
				ObjectNew: &updatev1alpha1.ScalingGroup{},
			},
		},
		"new object is not a scaling group": {
			event: event.UpdateEvent{
				ObjectOld: &updatev1alpha1.ScalingGroup{},
			},
		},
		"image reference is unchanged": {
			event: event.UpdateEvent{
				ObjectOld: &updatev1alpha1.ScalingGroup{
					Status: updatev1alpha1.ScalingGroupStatus{ImageReference: "image-reference"},
				},
				ObjectNew: &updatev1alpha1.ScalingGroup{
					Status: updatev1alpha1.ScalingGroupStatus{ImageReference: "image-reference"},
				},
			},
		},
		"image reference has changed": {
			event: event.UpdateEvent{
				ObjectOld: &updatev1alpha1.ScalingGroup{
					Status: updatev1alpha1.ScalingGroupStatus{ImageReference: "old-image-reference"},
				},
				ObjectNew: &updatev1alpha1.ScalingGroup{
					Status: updatev1alpha1.ScalingGroupStatus{ImageReference: "new-image-reference"},
				},
			},
			wantProcessing: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			predicate := scalingGroupImageChangedPredicate()
			assert.Equal(tc.wantProcessing, predicate.Update(tc.event))
		})
	}
}

func TestAutoscalerEnabledStatusChangedPredicate(t *testing.T) {
	testCases := map[string]struct {
		event          event.UpdateEvent
		wantProcessing bool
	}{
		"old object is not an autoscaling strategy": {
			event: event.UpdateEvent{
				ObjectNew: &updatev1alpha1.AutoscalingStrategy{},
			},
		},
		"new object is not an autoscaling strategy": {
			event: event.UpdateEvent{
				ObjectOld: &updatev1alpha1.AutoscalingStrategy{},
			},
		},
		"status is unchanged": {
			event: event.UpdateEvent{
				ObjectOld: &updatev1alpha1.AutoscalingStrategy{
					Status: updatev1alpha1.AutoscalingStrategyStatus{Enabled: true},
				},
				ObjectNew: &updatev1alpha1.AutoscalingStrategy{
					Status: updatev1alpha1.AutoscalingStrategyStatus{Enabled: true},
				},
			},
		},
		"status has changed": {
			event: event.UpdateEvent{
				ObjectOld: &updatev1alpha1.AutoscalingStrategy{
					Status: updatev1alpha1.AutoscalingStrategyStatus{},
				},
				ObjectNew: &updatev1alpha1.AutoscalingStrategy{
					Status: updatev1alpha1.AutoscalingStrategyStatus{Enabled: true},
				},
			},
			wantProcessing: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			predicate := autoscalerEnabledStatusChangedPredicate()
			assert.Equal(tc.wantProcessing, predicate.Update(tc.event))
		})
	}
}

func TestNodeReadyPredicate(t *testing.T) {
	testCases := map[string]struct {
		event          event.UpdateEvent
		wantProcessing bool
	}{
		"old object is not a node": {
			event: event.UpdateEvent{
				ObjectNew: &corev1.Node{},
			},
		},
		"new object is not a node": {
			event: event.UpdateEvent{
				ObjectOld: &corev1.Node{},
			},
		},
		"status is unchanged": {
			event: event.UpdateEvent{
				ObjectOld: &corev1.Node{},
				ObjectNew: &corev1.Node{},
			},
		},
		"node became ready": {
			event: event.UpdateEvent{
				ObjectOld: &corev1.Node{
					Status: corev1.NodeStatus{
						Conditions: []corev1.NodeCondition{
							{Type: corev1.NodeReady, Status: corev1.ConditionFalse},
						},
					},
				},
				ObjectNew: &corev1.Node{
					Status: corev1.NodeStatus{
						Conditions: []corev1.NodeCondition{
							{Type: corev1.NodeReady, Status: corev1.ConditionTrue},
						},
					},
				},
			},
			wantProcessing: true,
		},
		"node acquired provider id": {
			event: event.UpdateEvent{
				ObjectOld: &corev1.Node{},
				ObjectNew: &corev1.Node{
					Spec: corev1.NodeSpec{
						ProviderID: "provider-id",
					},
				},
			},
			wantProcessing: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			predicate := nodeReadyPredicate()
			assert.Equal(tc.wantProcessing, predicate.Update(tc.event))
		})
	}
}

func TestNodeMaintenanceSucceededPredicate(t *testing.T) {
	testCases := map[string]struct {
		event          event.UpdateEvent
		wantProcessing bool
	}{
		"old object is not a node maintenance resource": {
			event: event.UpdateEvent{
				ObjectNew: &nodemaintenancev1beta1.NodeMaintenance{},
			},
		},
		"new object is not a node maintenance resource": {
			event: event.UpdateEvent{
				ObjectOld: &nodemaintenancev1beta1.NodeMaintenance{},
			},
		},
		"status is unchanged": {
			event: event.UpdateEvent{
				ObjectOld: &nodemaintenancev1beta1.NodeMaintenance{
					Status: nodemaintenancev1beta1.NodeMaintenanceStatus{
						Phase: nodemaintenancev1beta1.MaintenanceRunning,
					},
				},
				ObjectNew: &nodemaintenancev1beta1.NodeMaintenance{
					Status: nodemaintenancev1beta1.NodeMaintenanceStatus{
						Phase: nodemaintenancev1beta1.MaintenanceRunning,
					},
				},
			},
		},
		"status has changed": {
			event: event.UpdateEvent{
				ObjectOld: &nodemaintenancev1beta1.NodeMaintenance{
					Status: nodemaintenancev1beta1.NodeMaintenanceStatus{
						Phase: nodemaintenancev1beta1.MaintenanceRunning,
					},
				},
				ObjectNew: &nodemaintenancev1beta1.NodeMaintenance{
					Status: nodemaintenancev1beta1.NodeMaintenanceStatus{
						Phase: nodemaintenancev1beta1.MaintenanceSucceeded,
					},
				},
			},
			wantProcessing: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			predicate := nodeMaintenanceSucceededPredicate()
			assert.Equal(tc.wantProcessing, predicate.Update(tc.event))
		})
	}
}

func TestFindObjectsForScalingGroup(t *testing.T) {
	scalingGroup := updatev1alpha1.ScalingGroup{
		Spec: updatev1alpha1.ScalingGroupSpec{
			NodeImage: "nodeimage",
		},
	}
	wantRequests := []reconcile.Request{
		{
			NamespacedName: types.NamespacedName{
				Name: "nodeimage",
			},
		},
	}
	assert := assert.New(t)
	reconciler := NodeImageReconciler{}
	requests := reconciler.findObjectsForScalingGroup(&scalingGroup)
	assert.ElementsMatch(wantRequests, requests)
}

func TestFindAllNodeImages(t *testing.T) {
	testCases := map[string]struct {
		nodeImage         client.Object
		listNodeImagesErr error
		wantRequests      []reconcile.Request
	}{
		"getting the corresponding node images fails": {
			listNodeImagesErr: errors.New("get-node-images-err"),
		},
		"node image reconcile request is returned": {
			nodeImage: &updatev1alpha1.NodeImage{
				ObjectMeta: metav1.ObjectMeta{Name: "nodeimage"},
			},
			wantRequests: []reconcile.Request{
				{
					NamespacedName: types.NamespacedName{
						Name: "nodeimage",
					},
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			reconciler := NodeImageReconciler{
				Client: newStubReaderClient(t, []runtime.Object{tc.nodeImage}, nil, tc.listNodeImagesErr),
			}
			requests := reconciler.findAllNodeImages(nil)
			assert.ElementsMatch(tc.wantRequests, requests)
		})
	}
}
