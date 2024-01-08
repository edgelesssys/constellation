/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package controllers

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	updatev1alpha1 "github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/api/v1alpha1"
)

func TestNodeStateChangePredicate(t *testing.T) {
	updateTestCases := map[string]struct {
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

	for name, tc := range updateTestCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			predicate := nodeStateChangePredicate()
			assert.Equal(tc.wantProcessing, predicate.Update(tc.event))
		})
	}

	t.Run("create", func(t *testing.T) {
		assert := assert.New(t)
		predicate := nodeStateChangePredicate()
		assert.True(predicate.Create(event.CreateEvent{}))
	})

	t.Run("delete", func(t *testing.T) {
		assert := assert.New(t)
		predicate := nodeStateChangePredicate()
		assert.True(predicate.Delete(event.DeleteEvent{}))
	})

	t.Run("generic", func(t *testing.T) {
		assert := assert.New(t)
		predicate := nodeStateChangePredicate()
		assert.False(predicate.Generic(event.GenericEvent{}))
	})
}

func TestFindObjectsForNode(t *testing.T) {
	testCases := map[string]struct {
		pendingNode         client.Object
		listPendingNodesErr error
		wantRequests        []reconcile.Request
	}{
		"getting the corresponding pending nodes fails": {
			listPendingNodesErr: errors.New("get-pending-nodes-err"),
		},
		"pending nodes reconcile request is returned": {
			pendingNode: &updatev1alpha1.PendingNode{
				ObjectMeta: metav1.ObjectMeta{Name: "pending-node"},
			},
			wantRequests: []reconcile.Request{
				{
					NamespacedName: types.NamespacedName{
						Name: "pending-node",
					},
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			reconciler := PendingNodeReconciler{
				Client: newStubReaderClient(t, []runtime.Object{tc.pendingNode}, nil, tc.listPendingNodesErr),
			}
			requests := reconciler.findObjectsForNode(context.TODO(), &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "pending-node",
				},
			})
			assert.ElementsMatch(tc.wantRequests, requests)
		})
	}
}

func TestReachedGoal(t *testing.T) {
	testCases := map[string]struct {
		pendingNode       updatev1alpha1.PendingNode
		nodeState         updatev1alpha1.CSPNodeState
		getPendingNodeErr error
		wantErr           bool
		wantGoalReached   bool
	}{
		"join: getting the corresponding k8s node fails": {
			pendingNode: updatev1alpha1.PendingNode{
				ObjectMeta: metav1.ObjectMeta{Name: "pending-node"},
				Spec:       updatev1alpha1.PendingNodeSpec{Goal: updatev1alpha1.NodeGoalJoin},
			},
			nodeState:         updatev1alpha1.NodeStateReady,
			getPendingNodeErr: errors.New("get-pending-node-err"),
			wantErr:           true,
		},
		"join: node not found": {
			pendingNode: updatev1alpha1.PendingNode{
				ObjectMeta: metav1.ObjectMeta{Name: "pending-node"},
				Spec:       updatev1alpha1.PendingNodeSpec{Goal: updatev1alpha1.NodeGoalJoin},
			},
			nodeState: updatev1alpha1.NodeStateReady,
			getPendingNodeErr: &apierrors.StatusError{
				ErrStatus: metav1.Status{
					Status: "Failure",
					Reason: "NotFound",
					Code:   http.StatusNotFound,
				},
			},
		},
		"join: csp states node is not ready": {
			pendingNode: updatev1alpha1.PendingNode{
				ObjectMeta: metav1.ObjectMeta{Name: "pending-node"},
				Spec:       updatev1alpha1.PendingNodeSpec{Goal: updatev1alpha1.NodeGoalJoin},
			},
			nodeState: updatev1alpha1.NodeStateFailed,
		},
		"join: node joined": {
			pendingNode: updatev1alpha1.PendingNode{
				ObjectMeta: metav1.ObjectMeta{Name: "pending-node"},
				Spec:       updatev1alpha1.PendingNodeSpec{Goal: updatev1alpha1.NodeGoalJoin},
			},
			nodeState:       updatev1alpha1.NodeStateReady,
			wantGoalReached: true,
		},
		"leave: node still exists": {
			pendingNode: updatev1alpha1.PendingNode{
				ObjectMeta: metav1.ObjectMeta{Name: "pending-node"},
				Spec:       updatev1alpha1.PendingNodeSpec{Goal: updatev1alpha1.NodeGoalLeave},
			},
			nodeState: updatev1alpha1.NodeStateReady,
		},
		"leave: node terminated": {
			pendingNode: updatev1alpha1.PendingNode{
				ObjectMeta: metav1.ObjectMeta{Name: "pending-node"},
				Spec:       updatev1alpha1.PendingNodeSpec{Goal: updatev1alpha1.NodeGoalLeave},
			},
			nodeState:       updatev1alpha1.NodeStateTerminated,
			wantGoalReached: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			reconciler := PendingNodeReconciler{
				Client: newStubReaderClient(t, []runtime.Object{&tc.pendingNode}, tc.getPendingNodeErr, nil),
			}
			reachedGoal, err := reconciler.reachedGoal(context.Background(), tc.pendingNode, tc.nodeState)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantGoalReached, reachedGoal)
		})
	}
}

type stubNodeStateGetter struct {
	sync.RWMutex
	nodeState     updatev1alpha1.CSPNodeState
	nodeStateErr  error
	deleteNodeErr error
}

func (g *stubNodeStateGetter) GetNodeState(_ context.Context, _ string) (updatev1alpha1.CSPNodeState, error) {
	g.RLock()
	defer g.RUnlock()
	return g.nodeState, g.nodeStateErr
}

func (g *stubNodeStateGetter) DeleteNode(_ context.Context, _ string) error {
	g.RLock()
	defer g.RUnlock()
	return g.deleteNodeErr
}

// thread safe methods to update the stub while in use

func (g *stubNodeStateGetter) setNodeState(nodeState updatev1alpha1.CSPNodeState) {
	g.Lock()
	defer g.Unlock()
	g.nodeState = nodeState
}
