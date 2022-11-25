/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package deploy

import (
	"context"
	"errors"
	"testing"

	updatev1alpha1 "github.com/edgelesssys/constellation/operators/constellation-node-operator/v2/api/v1alpha1"
	"github.com/edgelesssys/constellation/operators/constellation-node-operator/v2/internal/constants"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestInitialResources(t *testing.T) {
	testCases := map[string]struct {
		items         []scalingGroupStoreItem
		imageErr      error
		nameErr       error
		listErr       error
		createErr     error
		wantResources int
		wantErr       bool
	}{
		"creating initial resources works": {
			items: []scalingGroupStoreItem{
				{groupID: "control-plane", image: "image-1", name: "control-plane", isControlPlane: true},
				{groupID: "worker", image: "image-1", name: "worker"},
			},
			wantResources: 4,
		},
		"missing control planes": {
			items: []scalingGroupStoreItem{
				{groupID: "worker", image: "image-1", name: "worker"},
			},
			wantErr: true,
		},
		"missing workers": {
			items: []scalingGroupStoreItem{
				{groupID: "control-plane", image: "image-1", name: "control-plane", isControlPlane: true},
			},
			wantErr: true,
		},
		"listing groups fails": {
			listErr: errors.New("list failed"),
			wantErr: true,
		},
		"creating resources fails": {
			items: []scalingGroupStoreItem{
				{groupID: "control-plane", image: "image-1", name: "control-plane", isControlPlane: true},
				{groupID: "worker", image: "image-1", name: "worker"},
			},
			createErr: errors.New("create failed"),
			wantErr:   true,
		},
		"getting image fails": {
			items: []scalingGroupStoreItem{
				{groupID: "control-plane", image: "image-1", name: "control-plane", isControlPlane: true},
				{groupID: "worker", image: "image-1", name: "worker"},
			},
			imageErr: errors.New("getting image failed"),
			wantErr:  true,
		},
		"getting name fails": {
			items: []scalingGroupStoreItem{
				{groupID: "control-plane", image: "image-1", name: "control-plane", isControlPlane: true},
				{groupID: "worker", image: "image-1", name: "worker"},
			},
			nameErr: errors.New("getting name failed"),
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			k8sClient := &stubK8sClient{createErr: tc.createErr}
			scalingGroupGetter := newScalingGroupGetter(tc.items, tc.imageErr, tc.nameErr, tc.listErr)
			err := InitialResources(context.Background(), k8sClient, &stubImageInfo{}, scalingGroupGetter, "uid")
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Len(k8sClient.createdObjects, tc.wantResources)
		})
	}
}

func TestCreateAutoscalingStrategy(t *testing.T) {
	testCases := map[string]struct {
		createErr    error
		wantStrategy *updatev1alpha1.AutoscalingStrategy
		wantErr      bool
	}{
		"create works": {
			wantStrategy: &updatev1alpha1.AutoscalingStrategy{
				TypeMeta: metav1.TypeMeta{APIVersion: "update.edgeless.systems/v1alpha1", Kind: "AutoscalingStrategy"},
				ObjectMeta: metav1.ObjectMeta{
					Name: constants.AutoscalingStrategyResourceName,
				},
				Spec: updatev1alpha1.AutoscalingStrategySpec{
					Enabled:             true,
					DeploymentName:      "constellation-cluster-autoscaler",
					DeploymentNamespace: "kube-system",
					AutoscalerExtraArgs: map[string]string{
						"cloud-provider":  "stub",
						"logtostderr":     "true",
						"stderrthreshold": "info",
						"v":               "2",
						"namespace":       "kube-system",
					},
				},
			},
		},
		"create fails": {
			createErr: errors.New("create failed"),
			wantErr:   true,
		},
		"strategy exists": {
			createErr: k8sErrors.NewAlreadyExists(schema.GroupResource{}, constants.AutoscalingStrategyResourceName),
			wantStrategy: &updatev1alpha1.AutoscalingStrategy{
				TypeMeta: metav1.TypeMeta{APIVersion: "update.edgeless.systems/v1alpha1", Kind: "AutoscalingStrategy"},
				ObjectMeta: metav1.ObjectMeta{
					Name: constants.AutoscalingStrategyResourceName,
				},
				Spec: updatev1alpha1.AutoscalingStrategySpec{
					Enabled:             true,
					DeploymentName:      "constellation-cluster-autoscaler",
					DeploymentNamespace: "kube-system",
					AutoscalerExtraArgs: map[string]string{
						"cloud-provider":  "stub",
						"logtostderr":     "true",
						"stderrthreshold": "info",
						"v":               "2",
						"namespace":       "kube-system",
					},
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			k8sClient := &stubK8sClient{createErr: tc.createErr}
			err := createAutoscalingStrategy(context.Background(), k8sClient, "stub")
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Len(k8sClient.createdObjects, 1)
			assert.Equal(tc.wantStrategy, k8sClient.createdObjects[0])
		})
	}
}

func TestCreateNodeImage(t *testing.T) {
	testCases := map[string]struct {
		createErr     error
		wantNodeImage *updatev1alpha1.NodeImage
		wantErr       bool
	}{
		"create works": {
			wantNodeImage: &updatev1alpha1.NodeImage{
				TypeMeta: metav1.TypeMeta{APIVersion: "update.edgeless.systems/v1alpha1", Kind: "NodeImage"},
				ObjectMeta: metav1.ObjectMeta{
					Name: constants.NodeImageResourceName,
				},
				Spec: updatev1alpha1.NodeImageSpec{
					ImageReference: "image-reference",
					ImageVersion:   "image-version",
				},
			},
		},
		"create fails": {
			createErr: errors.New("create failed"),
			wantErr:   true,
		},
		"image exists": {
			createErr: k8sErrors.NewAlreadyExists(schema.GroupResource{}, constants.AutoscalingStrategyResourceName),
			wantNodeImage: &updatev1alpha1.NodeImage{
				TypeMeta: metav1.TypeMeta{APIVersion: "update.edgeless.systems/v1alpha1", Kind: "NodeImage"},
				ObjectMeta: metav1.ObjectMeta{
					Name: constants.NodeImageResourceName,
				},
				Spec: updatev1alpha1.NodeImageSpec{
					ImageReference: "image-reference",
					ImageVersion:   "image-version",
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			k8sClient := &stubK8sClient{createErr: tc.createErr}
			err := createNodeImage(context.Background(), k8sClient, "image-reference", "image-version")
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Len(k8sClient.createdObjects, 1)
			assert.Equal(tc.wantNodeImage, k8sClient.createdObjects[0])
		})
	}
}

func TestCreateScalingGroup(t *testing.T) {
	testCases := map[string]struct {
		createErr        error
		wantScalingGroup *updatev1alpha1.ScalingGroup
		wantErr          bool
	}{
		"create works": {
			wantScalingGroup: &updatev1alpha1.ScalingGroup{
				TypeMeta: metav1.TypeMeta{APIVersion: "update.edgeless.systems/v1alpha1", Kind: "ScalingGroup"},
				ObjectMeta: metav1.ObjectMeta{
					Name: "group-name",
				},
				Spec: updatev1alpha1.ScalingGroupSpec{
					NodeImage:           constants.NodeImageResourceName,
					GroupID:             "group-id",
					AutoscalerGroupName: "group-Name",
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
		"image exists": {
			createErr: k8sErrors.NewAlreadyExists(schema.GroupResource{}, constants.AutoscalingStrategyResourceName),
			wantScalingGroup: &updatev1alpha1.ScalingGroup{
				TypeMeta: metav1.TypeMeta{APIVersion: "update.edgeless.systems/v1alpha1", Kind: "ScalingGroup"},
				ObjectMeta: metav1.ObjectMeta{
					Name: "group-name",
				},
				Spec: updatev1alpha1.ScalingGroupSpec{
					NodeImage:           constants.NodeImageResourceName,
					GroupID:             "group-id",
					AutoscalerGroupName: "group-Name",
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

			k8sClient := &stubK8sClient{createErr: tc.createErr}
			newScalingGroupConfig := newScalingGroupConfig{k8sClient, "group-id", "group-Name", "group-Name", updatev1alpha1.WorkerRole}
			err := createScalingGroup(context.Background(), newScalingGroupConfig)
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

type stubK8sClient struct {
	createdObjects []client.Object
	createErr      error
	client.Writer
}

func (s *stubK8sClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	s.createdObjects = append(s.createdObjects, obj)
	return s.createErr
}

type stubImageInfo struct {
	imageVersion string
	err          error
}

func (s stubImageInfo) ImageVersion(_ string) (string, error) {
	return s.imageVersion, s.err
}

type stubScalingGroupGetter struct {
	store    map[string]scalingGroupStoreItem
	imageErr error
	nameErr  error
	listErr  error
}

func newScalingGroupGetter(items []scalingGroupStoreItem, imageErr, nameErr, listErr error) *stubScalingGroupGetter {
	store := make(map[string]scalingGroupStoreItem)
	for _, item := range items {
		store[item.groupID] = item
	}
	return &stubScalingGroupGetter{
		store:    store,
		imageErr: imageErr,
		nameErr:  nameErr,
		listErr:  listErr,
	}
}

func (g *stubScalingGroupGetter) GetScalingGroupImage(ctx context.Context, scalingGroupID string) (string, error) {
	return g.store[scalingGroupID].image, g.imageErr
}

func (g *stubScalingGroupGetter) GetScalingGroupName(scalingGroupID string) (string, error) {
	return g.store[scalingGroupID].name, g.nameErr
}

func (g *stubScalingGroupGetter) GetAutoscalingGroupName(scalingGroupID string) (string, error) {
	return g.store[scalingGroupID].name, g.nameErr
}

func (g *stubScalingGroupGetter) ListScalingGroups(ctx context.Context, uid string) (controlPlaneGroupIDs []string, workerGroupIDs []string, err error) {
	for _, item := range g.store {
		if item.isControlPlane {
			controlPlaneGroupIDs = append(controlPlaneGroupIDs, item.groupID)
		} else {
			workerGroupIDs = append(workerGroupIDs, item.groupID)
		}
	}
	return controlPlaneGroupIDs, workerGroupIDs, g.listErr
}

func (g *stubScalingGroupGetter) AutoscalingCloudProvider() string {
	return "stub"
}

type scalingGroupStoreItem struct {
	groupID        string
	name           string
	image          string
	isControlPlane bool
}
