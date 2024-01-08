/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package sgreconciler

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"

	mainconstants "github.com/edgelesssys/constellation/v2/internal/constants"
	updatev1alpha1 "github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/api/v1alpha1"
	cspapi "github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/internal/cloud/api"
	"github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/internal/constants"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestCreateScalingGroupIfNotExists(t *testing.T) {
	testCases := map[string]struct {
		createErr        error
		wantScalingGroup *updatev1alpha1.ScalingGroup
		wantErr          bool
	}{
		"create works": {
			wantScalingGroup: &updatev1alpha1.ScalingGroup{
				TypeMeta: metav1.TypeMeta{APIVersion: "update.edgeless.systems/v1alpha1", Kind: "ScalingGroup"},
				ObjectMeta: metav1.ObjectMeta{
					Name: "resource-name",
				},
				Spec: updatev1alpha1.ScalingGroupSpec{
					NodeVersion:         mainconstants.NodeVersionResourceName,
					GroupID:             "group-id",
					AutoscalerGroupName: "autoscaling-group-name",
					NodeGroupName:       "node-group-name",
					Min:                 1,
					Max:                 10,
					Role:                updatev1alpha1.WorkerRole,
				},
			},
		},
		"create fails": {
			createErr: errors.New("create failed"),
			wantErr:   true,
		},
		"scaling group exists": {
			createErr: k8sErrors.NewAlreadyExists(schema.GroupResource{}, constants.AutoscalingStrategyResourceName),
			wantScalingGroup: &updatev1alpha1.ScalingGroup{
				TypeMeta: metav1.TypeMeta{APIVersion: "update.edgeless.systems/v1alpha1", Kind: "ScalingGroup"},
				ObjectMeta: metav1.ObjectMeta{
					Name: "resource-name",
				},
				Spec: updatev1alpha1.ScalingGroupSpec{
					NodeVersion:         mainconstants.NodeVersionResourceName,
					GroupID:             "group-id",
					AutoscalerGroupName: "autoscaling-group-name",
					NodeGroupName:       "node-group-name",
					Min:                 1,
					Max:                 10,
					Role:                updatev1alpha1.WorkerRole,
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			k8sClient := &fakeK8sClient{createErr: tc.createErr}
			newScalingGroupConfig := newScalingGroupConfig{
				k8sClient:            k8sClient,
				resourceName:         "resource-name",
				groupID:              "group-id",
				nodeGroupName:        "node-group-name",
				autoscalingGroupName: "autoscaling-group-name",
				role:                 updatev1alpha1.WorkerRole,
			}
			err := createScalingGroupIfNotExists(context.Background(), newScalingGroupConfig)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Len(k8sClient.createdObjects, 1)
			assert.Equal(tc.wantScalingGroup, k8sClient.createdObjects[0])
		})
	}
}

func TestPatchNodeGroupName(t *testing.T) {
	testCases := map[string]struct {
		getRes     client.Object
		getErr     error
		updateErr  error
		wantExists bool
		wantErr    bool
	}{
		"patching works": {
			getRes: &updatev1alpha1.ScalingGroup{
				TypeMeta: metav1.TypeMeta{APIVersion: "update.edgeless.systems/v1alpha1", Kind: "ScalingGroup"},
				ObjectMeta: metav1.ObjectMeta{
					Name: "resource-name",
				},
				Spec: updatev1alpha1.ScalingGroupSpec{
					NodeVersion:         mainconstants.NodeVersionResourceName,
					GroupID:             "group-id",
					AutoscalerGroupName: "autoscaling-group-name",
					Autoscaling:         true,
					Min:                 1,
					Max:                 10,
					Role:                updatev1alpha1.ControlPlaneRole,
				},
			},
			wantExists: true,
		},
		"name already set": {
			getRes: &updatev1alpha1.ScalingGroup{
				TypeMeta: metav1.TypeMeta{APIVersion: "update.edgeless.systems/v1alpha1", Kind: "ScalingGroup"},
				ObjectMeta: metav1.ObjectMeta{
					Name: "resource-name",
				},
				Spec: updatev1alpha1.ScalingGroupSpec{
					NodeVersion:         mainconstants.NodeVersionResourceName,
					GroupID:             "group-id",
					NodeGroupName:       "node-group-name",
					AutoscalerGroupName: "autoscaling-group-name",
					Autoscaling:         true,
					Min:                 1,
					Max:                 10,
					Role:                updatev1alpha1.ControlPlaneRole,
				},
			},
			wantExists: true,
		},
		"does not exist": {
			getErr:     k8sErrors.NewNotFound(schema.GroupResource{}, "resource-name"),
			wantExists: false,
		},
		"getting fails": {
			getErr:  errors.New("get failed"),
			wantErr: true,
		},
		"patching fails": {
			getRes: &updatev1alpha1.ScalingGroup{
				TypeMeta: metav1.TypeMeta{APIVersion: "update.edgeless.systems/v1alpha1", Kind: "ScalingGroup"},
				ObjectMeta: metav1.ObjectMeta{
					Name: "resource-name",
				},
				Spec: updatev1alpha1.ScalingGroupSpec{
					NodeVersion:         mainconstants.NodeVersionResourceName,
					GroupID:             "group-id",
					AutoscalerGroupName: "autoscaling-group-name",
					Autoscaling:         true,
					Min:                 1,
					Max:                 10,
					Role:                updatev1alpha1.ControlPlaneRole,
				},
			},
			updateErr: errors.New("patch failed"),
			wantErr:   true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			k8sClient := &fakeK8sClient{
				getRes:    tc.getRes,
				getErr:    tc.getErr,
				updateErr: tc.updateErr,
			}
			gotExists, gotErr := patchNodeGroupName(context.Background(), k8sClient, "resource-name", "node-group-name")
			if tc.wantErr {
				assert.Error(gotErr)
				return
			}
			require.NoError(gotErr)
			assert.Equal(tc.wantExists, gotExists)
		})
	}
}

type fakeK8sClient struct {
	getRes         client.Object
	createdObjects []client.Object
	createErr      error
	listErr        error
	getErr         error
	updateErr      error
	client.Client
}

func (s *fakeK8sClient) Create(_ context.Context, obj client.Object, _ ...client.CreateOption) error {
	for _, o := range s.createdObjects {
		if obj.GetName() == o.GetName() {
			return k8sErrors.NewAlreadyExists(schema.GroupResource{}, obj.GetName())
		}
	}

	s.createdObjects = append(s.createdObjects, obj.DeepCopyObject().(client.Object))
	return s.createErr
}

func (s *fakeK8sClient) Get(_ context.Context, _ types.NamespacedName, out client.Object, _ ...client.GetOption) error {
	if s.getErr != nil {
		return s.getErr
	}
	obj := s.getRes.DeepCopyObject()
	outVal := reflect.ValueOf(out)
	objVal := reflect.ValueOf(obj)
	if !objVal.Type().AssignableTo(outVal.Type()) {
		return fmt.Errorf("fake had type %s, but %s was asked for", objVal.Type(), outVal.Type())
	}
	reflect.Indirect(outVal).Set(reflect.Indirect(objVal))
	return nil
}

func (s *fakeK8sClient) Update(_ context.Context, _ client.Object, _ ...client.UpdateOption) error {
	return s.updateErr
}

func (s *fakeK8sClient) List(_ context.Context, _ client.ObjectList, _ ...client.ListOption) error {
	return s.listErr
}

type stubScalingGroupDiscoverer struct {
	sync.RWMutex
	groups []cspapi.ScalingGroup
}

func (d *stubScalingGroupDiscoverer) ListScalingGroups(_ context.Context, _ string,
) ([]cspapi.ScalingGroup, error) {
	d.RLock()
	defer d.RUnlock()
	ret := make([]cspapi.ScalingGroup, len(d.groups))
	copy(ret, d.groups)
	return ret, nil
}

func (d *stubScalingGroupDiscoverer) set(groups []cspapi.ScalingGroup) {
	d.Lock()
	defer d.Unlock()
	d.groups = groups
}

func (d *stubScalingGroupDiscoverer) reset() {
	d.Lock()
	defer d.Unlock()
	d.groups = nil
}
