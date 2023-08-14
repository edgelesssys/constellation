/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package kubernetes

import (
	"context"
	"errors"
	"io"
	"testing"

	kerrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/compatibility"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/versions"
	"github.com/edgelesssys/constellation/v2/internal/versions/components"
	updatev1alpha1 "github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/v2/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestUpgradeNodeVersion(t *testing.T) {
	someErr := errors.New("some error")
	testCases := map[string]struct {
		stable                *fakeStableClient
		conditions            []metav1.Condition
		currentImageVersion   string
		newImageReference     string
		badImageVersion       string
		currentClusterVersion string
		conf                  *config.Config
		force                 bool
		getErr                error
		wantErr               bool
		wantUpdate            bool
		assertCorrectError    func(t *testing.T, err error) bool
		customClientFn        func(nodeVersion updatev1alpha1.NodeVersion) dynamicInterface
	}{
		"success": {
			conf: func() *config.Config {
				conf := config.Default()
				conf.Image = "v1.2.3"
				conf.KubernetesVersion = versions.SupportedK8sVersions()[1]
				return conf
			}(),
			currentImageVersion:   "v1.2.2",
			currentClusterVersion: versions.SupportedK8sVersions()[0],
			stable: &fakeStableClient{
				configMaps: map[string]*corev1.ConfigMap{
					constants.JoinConfigMap: newJoinConfigMap(`{"0":{"expected":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA","warnOnly":false}}`),
				},
			},
			wantUpdate: true,
		},
		"only k8s upgrade": {
			conf: func() *config.Config {
				conf := config.Default()
				conf.Image = "v1.2.2"
				conf.KubernetesVersion = versions.SupportedK8sVersions()[1]
				return conf
			}(),
			currentImageVersion:   "v1.2.2",
			currentClusterVersion: versions.SupportedK8sVersions()[0],
			stable: &fakeStableClient{
				configMaps: map[string]*corev1.ConfigMap{
					constants.JoinConfigMap: newJoinConfigMap(`{"0":{"expected":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA","warnOnly":false}}`),
				},
			},
			wantUpdate: true,
			wantErr:    true,
			assertCorrectError: func(t *testing.T, err error) bool {
				var upgradeErr *compatibility.InvalidUpgradeError
				return assert.ErrorAs(t, err, &upgradeErr)
			},
		},
		"only image upgrade": {
			conf: func() *config.Config {
				conf := config.Default()
				conf.Image = "v1.2.3"
				conf.KubernetesVersion = versions.SupportedK8sVersions()[0]
				return conf
			}(),
			currentImageVersion:   "v1.2.2",
			currentClusterVersion: versions.SupportedK8sVersions()[0],
			stable: &fakeStableClient{
				configMaps: map[string]*corev1.ConfigMap{
					constants.JoinConfigMap: newJoinConfigMap(`{"0":{"expected":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA","warnOnly":false}}`),
				},
			},
			wantUpdate: true,
			wantErr:    true,
			assertCorrectError: func(t *testing.T, err error) bool {
				var upgradeErr *compatibility.InvalidUpgradeError
				return assert.ErrorAs(t, err, &upgradeErr)
			},
		},
		"not an upgrade": {
			conf: func() *config.Config {
				conf := config.Default()
				conf.Image = "v1.2.2"
				conf.KubernetesVersion = versions.SupportedK8sVersions()[0]
				return conf
			}(),
			currentImageVersion:   "v1.2.2",
			currentClusterVersion: versions.SupportedK8sVersions()[0],
			stable:                &fakeStableClient{},
			wantErr:               true,
			assertCorrectError: func(t *testing.T, err error) bool {
				var upgradeErr *compatibility.InvalidUpgradeError
				return assert.ErrorAs(t, err, &upgradeErr)
			},
		},
		"upgrade in progress": {
			conf: func() *config.Config {
				conf := config.Default()
				conf.Image = "v1.2.3"
				conf.KubernetesVersion = versions.SupportedK8sVersions()[1]
				return conf
			}(),
			conditions: []metav1.Condition{{
				Type:   updatev1alpha1.ConditionOutdated,
				Status: metav1.ConditionTrue,
			}},
			currentImageVersion:   "v1.2.2",
			currentClusterVersion: versions.SupportedK8sVersions()[0],
			stable:                &fakeStableClient{},
			wantErr:               true,
			assertCorrectError: func(t *testing.T, err error) bool {
				return assert.ErrorIs(t, err, ErrInProgress)
			},
		},
		"success with force and upgrade in progress": {
			conf: func() *config.Config {
				conf := config.Default()
				conf.Image = "v1.2.3"
				conf.KubernetesVersion = versions.SupportedK8sVersions()[1]
				return conf
			}(),
			conditions: []metav1.Condition{{
				Type:   updatev1alpha1.ConditionOutdated,
				Status: metav1.ConditionTrue,
			}},
			currentImageVersion:   "v1.2.2",
			currentClusterVersion: versions.SupportedK8sVersions()[0],
			stable:                &fakeStableClient{},
			force:                 true,
			wantUpdate:            true,
		},
		"get error": {
			conf: func() *config.Config {
				conf := config.Default()
				conf.Image = "v1.2.3"
				return conf
			}(),
			getErr:  someErr,
			wantErr: true,
			assertCorrectError: func(t *testing.T, err error) bool {
				return assert.ErrorIs(t, err, someErr)
			},
		},
		"image too new valid k8s": {
			conf: func() *config.Config {
				conf := config.Default()
				conf.Image = "v1.4.2"
				conf.KubernetesVersion = versions.SupportedK8sVersions()[1]
				return conf
			}(),
			newImageReference:     "path/to/image:v1.4.2",
			currentImageVersion:   "v1.2.2",
			currentClusterVersion: versions.SupportedK8sVersions()[0],
			stable: &fakeStableClient{
				configMaps: map[string]*corev1.ConfigMap{
					constants.JoinConfigMap: newJoinConfigMap(`{"0":{"expected":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA","warnOnly":true}}`),
				},
			},
			wantUpdate: true,
			wantErr:    true,
			assertCorrectError: func(t *testing.T, err error) bool {
				var upgradeErr *compatibility.InvalidUpgradeError
				return assert.ErrorAs(t, err, &upgradeErr)
			},
		},
		"success with force and image too new": {
			conf: func() *config.Config {
				conf := config.Default()
				conf.Image = "v1.4.2"
				conf.KubernetesVersion = versions.SupportedK8sVersions()[1]
				return conf
			}(),
			newImageReference:     "path/to/image:v1.4.2",
			currentImageVersion:   "v1.2.2",
			currentClusterVersion: versions.SupportedK8sVersions()[0],
			stable: &fakeStableClient{
				configMaps: map[string]*corev1.ConfigMap{
					constants.JoinConfigMap: newJoinConfigMap(`{"0":{"expected":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA","warnOnly":false}}`),
				},
			},
			wantUpdate: true,
			force:      true,
		},
		"apply returns bad object": {
			conf: func() *config.Config {
				conf := config.Default()
				conf.Image = "v1.2.3"
				conf.KubernetesVersion = versions.SupportedK8sVersions()[1]
				return conf
			}(),
			currentImageVersion:   "v1.2.2",
			currentClusterVersion: versions.SupportedK8sVersions()[0],
			badImageVersion:       "v3.2.1",
			stable: &fakeStableClient{
				configMaps: map[string]*corev1.ConfigMap{
					constants.JoinConfigMap: newJoinConfigMap(`{"0":{"expected":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA","warnOnly":false}}`),
				},
			},
			wantUpdate: true,
			wantErr:    true,
			assertCorrectError: func(t *testing.T, err error) bool {
				var target *applyError
				return assert.ErrorAs(t, err, &target)
			},
		},
		"outdated k8s version skips k8s upgrade": {
			conf: func() *config.Config {
				conf := config.Default()
				conf.Image = "v1.2.2"
				conf.KubernetesVersion = "v1.25.8"
				return conf
			}(),
			currentImageVersion:   "v1.2.2",
			currentClusterVersion: versions.SupportedK8sVersions()[0],
			stable: &fakeStableClient{
				configMaps: map[string]*corev1.ConfigMap{
					constants.JoinConfigMap: newJoinConfigMap(`{"0":{"expected":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA","warnOnly":false}}`),
				},
			},
			wantUpdate: false,
			wantErr:    true,
			assertCorrectError: func(t *testing.T, err error) bool {
				var upgradeErr *compatibility.InvalidUpgradeError
				return assert.ErrorAs(t, err, &upgradeErr)
			},
		},
		"succeed after update retry when the updated node object is outdated": {
			conf: func() *config.Config {
				conf := config.Default()
				conf.Image = "v1.2.3"
				conf.KubernetesVersion = versions.SupportedK8sVersions()[1]
				return conf
			}(),
			currentImageVersion:   "v1.2.2",
			currentClusterVersion: versions.SupportedK8sVersions()[0],
			stable: &fakeStableClient{
				configMaps: map[string]*corev1.ConfigMap{
					constants.JoinConfigMap: newJoinConfigMap(`{"0":{"expected":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA","warnOnly":false}}`),
				},
			},
			wantUpdate: false, // because customClient is used
			customClientFn: func(nodeVersion updatev1alpha1.NodeVersion) dynamicInterface {
				fakeClient := &fakeDynamicClient{}
				fakeClient.On("GetCurrent", mock.Anything, mock.Anything).Return(unstructedObjectWithGeneration(nodeVersion, 1), nil)
				fakeClient.On("Update", mock.Anything, mock.Anything).Return(nil, kerrors.NewConflict(schema.GroupResource{Resource: nodeVersion.Name}, nodeVersion.Name, nil)).Once()
				fakeClient.On("Update", mock.Anything, mock.Anything).Return(unstructedObjectWithGeneration(nodeVersion, 2), nil).Once()
				return fakeClient
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			nodeVersion := updatev1alpha1.NodeVersion{
				Spec: updatev1alpha1.NodeVersionSpec{
					ImageVersion:             tc.currentImageVersion,
					KubernetesClusterVersion: tc.currentClusterVersion,
				},
				Status: updatev1alpha1.NodeVersionStatus{
					Conditions: tc.conditions,
				},
			}

			var badUpdatedObject *unstructured.Unstructured
			if tc.badImageVersion != "" {
				nodeVersion.Spec.ImageVersion = tc.badImageVersion
				unstrBadNodeVersion, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&nodeVersion)
				require.NoError(err)
				badUpdatedObject = &unstructured.Unstructured{Object: unstrBadNodeVersion}
			}

			unstrNodeVersion, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&nodeVersion)
			require.NoError(err)
			dynamicClient := &stubDynamicClient{object: &unstructured.Unstructured{Object: unstrNodeVersion}, badUpdatedObject: badUpdatedObject, getErr: tc.getErr}
			upgrader := Upgrader{
				stableInterface:  tc.stable,
				dynamicInterface: dynamicClient,
				imageFetcher:     &stubImageFetcher{reference: tc.newImageReference},
				log:              logger.NewTest(t),
				outWriter:        io.Discard,
			}
			if tc.customClientFn != nil {
				upgrader.dynamicInterface = tc.customClientFn(nodeVersion)
			}

			err = upgrader.UpgradeNodeVersion(context.Background(), tc.conf, tc.force)
			// Check upgrades first because if we checked err first, UpgradeImage may error due to other reasons and still trigger an upgrade.
			if tc.wantUpdate {
				assert.NotNil(dynamicClient.updatedObject)
			} else {
				assert.Nil(dynamicClient.updatedObject)
			}

			if tc.wantErr {
				assert.Error(err)
				tc.assertCorrectError(t, err)
				return
			}
			assert.NoError(err)
		})
	}
}

func TestUpdateImage(t *testing.T) {
	someErr := errors.New("error")
	testCases := map[string]struct {
		newImageReference string
		newImageVersion   string
		oldImageReference string
		oldImageVersion   string
		updateErr         error
		wantUpdate        bool
		wantErr           bool
	}{
		"success": {
			oldImageReference: "old-image-ref",
			oldImageVersion:   "v0.0.0",
			newImageReference: "new-image-ref",
			newImageVersion:   "v0.1.0",
			wantUpdate:        true,
		},
		"same version fails": {
			oldImageVersion: "v0.0.0",
			newImageVersion: "v0.0.0",
			wantErr:         true,
		},
		"update error": {
			updateErr: someErr,
			wantErr:   true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			upgrader := &Upgrader{
				log: logger.NewTest(t),
			}

			nodeVersion := updatev1alpha1.NodeVersion{
				Spec: updatev1alpha1.NodeVersionSpec{
					ImageReference: tc.oldImageReference,
					ImageVersion:   tc.oldImageVersion,
				},
			}

			err := upgrader.updateImage(&nodeVersion, tc.newImageReference, tc.newImageVersion, false)

			if tc.wantErr {
				assert.Error(err)
				return
			}

			assert.NoError(err)
			if tc.wantUpdate {
				assert.Equal(tc.newImageReference, nodeVersion.Spec.ImageReference)
				assert.Equal(tc.newImageVersion, nodeVersion.Spec.ImageVersion)
			} else {
				assert.Equal(tc.oldImageReference, nodeVersion.Spec.ImageReference)
				assert.Equal(tc.oldImageVersion, nodeVersion.Spec.ImageVersion)
			}
		})
	}
}

func TestUpdateK8s(t *testing.T) {
	someErr := errors.New("error")
	testCases := map[string]struct {
		newClusterVersion string
		oldClusterVersion string
		updateErr         error
		wantUpdate        bool
		wantErr           bool
	}{
		"success": {
			oldClusterVersion: "v0.0.0",
			newClusterVersion: "v0.1.0",
			wantUpdate:        true,
		},
		"same version fails": {
			oldClusterVersion: "v0.0.0",
			newClusterVersion: "v0.0.0",
			wantErr:           true,
		},
		"update error": {
			updateErr: someErr,
			wantErr:   true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			upgrader := &Upgrader{
				log: logger.NewTest(t),
			}

			nodeVersion := updatev1alpha1.NodeVersion{
				Spec: updatev1alpha1.NodeVersionSpec{
					KubernetesClusterVersion: tc.oldClusterVersion,
				},
			}

			_, err := upgrader.updateK8s(&nodeVersion, tc.newClusterVersion, components.Components{}, false)

			if tc.wantErr {
				assert.Error(err)
				return
			}

			assert.NoError(err)
			if tc.wantUpdate {
				assert.Equal(tc.newClusterVersion, nodeVersion.Spec.KubernetesClusterVersion)
			} else {
				assert.Equal(tc.oldClusterVersion, nodeVersion.Spec.KubernetesClusterVersion)
			}
		})
	}
}

func newJoinConfigMap(data string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: constants.JoinConfigMap,
		},
		Data: map[string]string{
			constants.AttestationConfigFilename: data,
		},
	}
}

type fakeDynamicClient struct {
	mock.Mock
}

func (u *fakeDynamicClient) GetCurrent(ctx context.Context, str string) (*unstructured.Unstructured, error) {
	args := u.Called(ctx, str)
	return args.Get(0).(*unstructured.Unstructured), args.Error(1)
}

func (u *fakeDynamicClient) Update(ctx context.Context, updatedObject *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	args := u.Called(ctx, updatedObject)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return updatedObject, args.Error(1)
}

type stubDynamicClient struct {
	object           *unstructured.Unstructured
	updatedObject    *unstructured.Unstructured
	badUpdatedObject *unstructured.Unstructured
	getErr           error
	updateErr        error
}

func (u *stubDynamicClient) GetCurrent(_ context.Context, _ string) (*unstructured.Unstructured, error) {
	return u.object, u.getErr
}

func (u *stubDynamicClient) Update(_ context.Context, updatedObject *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	u.updatedObject = updatedObject
	if u.badUpdatedObject != nil {
		return u.badUpdatedObject, u.updateErr
	}
	return u.updatedObject, u.updateErr
}

type fakeStableClient struct {
	configMaps        map[string]*corev1.ConfigMap
	updatedConfigMaps map[string]*corev1.ConfigMap
	k8sVersion        string
	getErr            error
	updateErr         error
	createErr         error
	k8sErr            error
	nodes             []corev1.Node
	nodesErr          error
}

func (s *fakeStableClient) GetConfigMap(_ context.Context, name string) (*corev1.ConfigMap, error) {
	return s.configMaps[name], s.getErr
}

func (s *fakeStableClient) UpdateConfigMap(_ context.Context, configMap *corev1.ConfigMap) (*corev1.ConfigMap, error) {
	if s.updatedConfigMaps == nil {
		s.updatedConfigMaps = map[string]*corev1.ConfigMap{}
	}
	s.updatedConfigMaps[configMap.ObjectMeta.Name] = configMap
	return s.updatedConfigMaps[configMap.ObjectMeta.Name], s.updateErr
}

func (s *fakeStableClient) CreateConfigMap(_ context.Context, configMap *corev1.ConfigMap) (*corev1.ConfigMap, error) {
	if s.configMaps == nil {
		s.configMaps = map[string]*corev1.ConfigMap{}
	}
	s.configMaps[configMap.ObjectMeta.Name] = configMap
	return s.configMaps[configMap.ObjectMeta.Name], s.createErr
}

func (s *fakeStableClient) KubernetesVersion() (string, error) {
	return s.k8sVersion, s.k8sErr
}

func (s *fakeStableClient) GetNodes(_ context.Context) ([]corev1.Node, error) {
	return s.nodes, s.nodesErr
}

type stubImageFetcher struct {
	reference         string
	fetchReferenceErr error
}

func (f *stubImageFetcher) FetchReference(_ context.Context,
	_ cloudprovider.Provider, _ variant.Variant,
	_, _ string,
) (string, error) {
	return f.reference, f.fetchReferenceErr
}

func unstructedObjectWithGeneration(nodeVersion updatev1alpha1.NodeVersion, generation int64) *unstructured.Unstructured {
	unstrNodeVersion, _ := runtime.DefaultUnstructuredConverter.ToUnstructured(&nodeVersion)
	object := &unstructured.Unstructured{Object: unstrNodeVersion}
	object.SetGeneration(generation)
	return object
}
