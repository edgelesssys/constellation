/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package deploy

import (
	"context"
	"errors"
	"testing"
	"time"

	mainconstants "github.com/edgelesssys/constellation/v2/internal/constants"
	updatev1alpha1 "github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/v2/api/v1alpha1"
	"github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/v2/internal/constants"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestInitialResources(t *testing.T) {
	k8sComponentsReference := "k8s-components-sha256-ABC"
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

			k8sClient := &fakeK8sClient{
				createErr: tc.createErr,
				listConfigMaps: []corev1.ConfigMap{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: k8sComponentsReference,
						},
					},
				},
			}
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

			k8sClient := &fakeK8sClient{createErr: tc.createErr}
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

func TestCreateNodeVersion(t *testing.T) {
	k8sComponentsReference := "k8s-components-sha256-reference"
	k8sClusterVersion := "1.20.0"
	testCases := map[string]struct {
		createErr           error
		existingNodeVersion *updatev1alpha1.NodeVersion
		wantNodeVersion     *updatev1alpha1.NodeVersion
		wantErr             bool
	}{
		"create works": {
			wantNodeVersion: &updatev1alpha1.NodeVersion{
				TypeMeta: metav1.TypeMeta{APIVersion: "update.edgeless.systems/v1alpha1", Kind: "NodeVersion"},
				ObjectMeta: metav1.ObjectMeta{
					Name: mainconstants.NodeVersionResourceName,
				},
				Spec: updatev1alpha1.NodeVersionSpec{
					ImageReference:                "image-reference",
					ImageVersion:                  "image-version",
					KubernetesComponentsReference: k8sComponentsReference,
					KubernetesClusterVersion:      k8sClusterVersion,
				},
			},
		},
		"create fails": {
			createErr: errors.New("create failed"),
			wantErr:   true,
		},
		"version exists": {
			createErr: k8sErrors.NewAlreadyExists(schema.GroupResource{}, mainconstants.NodeVersionResourceName),
			existingNodeVersion: &updatev1alpha1.NodeVersion{
				TypeMeta: metav1.TypeMeta{APIVersion: "update.edgeless.systems/v1alpha1", Kind: "NodeVersion"},
				ObjectMeta: metav1.ObjectMeta{
					Name: mainconstants.NodeVersionResourceName,
				},
				Spec: updatev1alpha1.NodeVersionSpec{
					ImageReference:                "image-reference2",
					ImageVersion:                  "image-version2",
					KubernetesComponentsReference: "components-reference2",
					KubernetesClusterVersion:      "cluster-version2",
				},
			},
			wantNodeVersion: &updatev1alpha1.NodeVersion{
				TypeMeta: metav1.TypeMeta{APIVersion: "update.edgeless.systems/v1alpha1", Kind: "NodeVersion"},
				ObjectMeta: metav1.ObjectMeta{
					Name: mainconstants.NodeVersionResourceName,
				},
				Spec: updatev1alpha1.NodeVersionSpec{
					ImageReference:                "image-reference2",
					ImageVersion:                  "image-version2",
					KubernetesComponentsReference: "components-reference2",
					KubernetesClusterVersion:      "cluster-version2",
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			k8sClient := &fakeK8sClient{
				createErr: tc.createErr,
				listConfigMaps: []corev1.ConfigMap{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:              k8sComponentsReference,
							CreationTimestamp: metav1.Time{Time: time.Unix(1, 0)},
						},
						Data: map[string]string{
							mainconstants.K8sVersionFieldName: k8sClusterVersion,
						},
					},
				},
			}
			if tc.existingNodeVersion != nil {
				k8sClient.createdObjects = append(k8sClient.createdObjects, tc.existingNodeVersion)
			}
			err := createNodeVersion(context.Background(), k8sClient, "image-reference", "image-version")
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Len(k8sClient.createdObjects, 1)
			assert.Equal(tc.wantNodeVersion, k8sClient.createdObjects[0])
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
					NodeVersion:         mainconstants.NodeVersionResourceName,
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
					NodeVersion:         mainconstants.NodeVersionResourceName,
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

			k8sClient := &fakeK8sClient{createErr: tc.createErr}
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

type fakeK8sClient struct {
	createdObjects []client.Object
	createErr      error
	listConfigMaps []corev1.ConfigMap
	listErr        error
	getErr         error
	updateErr      error
	client.Client
}

func (s *fakeK8sClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	for _, o := range s.createdObjects {
		if obj.GetName() == o.GetName() {
			return k8sErrors.NewAlreadyExists(schema.GroupResource{}, obj.GetName())
		}
	}

	s.createdObjects = append(s.createdObjects, obj)
	return s.createErr
}

func (s *fakeK8sClient) Get(ctx context.Context, key types.NamespacedName, obj client.Object, opts ...client.GetOption) error {
	if ObjNodeVersion, ok := obj.(*updatev1alpha1.NodeVersion); ok {
		for _, o := range s.createdObjects {
			if createdNodeVersion, ok := o.(*updatev1alpha1.NodeVersion); ok && createdNodeVersion != nil {
				if createdNodeVersion.Name == key.Name {
					ObjNodeVersion.ObjectMeta = createdNodeVersion.ObjectMeta
					ObjNodeVersion.TypeMeta = createdNodeVersion.TypeMeta
					ObjNodeVersion.Spec = createdNodeVersion.Spec
					return nil
				}
			}
		}
	}

	return s.getErr
}

func (s *fakeK8sClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	if updatedObjectNodeVersion, ok := obj.(*updatev1alpha1.NodeVersion); ok {
		for i, o := range s.createdObjects {
			if createdObjectNodeVersion, ok := o.(*updatev1alpha1.NodeVersion); ok && createdObjectNodeVersion != nil {
				if createdObjectNodeVersion.Name == updatedObjectNodeVersion.Name {
					s.createdObjects[i] = obj
					return nil
				}
			}
		}
	}
	return s.updateErr
}

func (s *fakeK8sClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	if configMapList, ok := list.(*corev1.ConfigMapList); ok {
		configMapList.Items = append(configMapList.Items, s.listConfigMaps...)
	}
	return s.listErr
}

type stubImageInfo struct {
	imageVersion string
	err          error
}

func (s stubImageInfo) ImageVersion() (string, error) {
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
