/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package kubecmd

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
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
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestUpgradeNodeVersion(t *testing.T) {
	testCases := map[string]struct {
		kubectl               *stubKubectl
		conditions            []metav1.Condition
		currentImageVersion   string
		newImageReference     string
		badImageVersion       string
		currentClusterVersion string
		conf                  *config.Config
		force                 bool
		getCRErr              error
		wantErr               bool
		wantUpdate            bool
		assertCorrectError    func(t *testing.T, err error) bool
		customClientFn        func(nodeVersion updatev1alpha1.NodeVersion) unstructuredInterface
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
			kubectl: &stubKubectl{
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
			kubectl: &stubKubectl{
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
			kubectl: &stubKubectl{
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
			kubectl:               &stubKubectl{},
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
			kubectl:               &stubKubectl{},
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
			kubectl:               &stubKubectl{},
			force:                 true,
			wantUpdate:            true,
		},
		"get error": {
			conf: func() *config.Config {
				conf := config.Default()
				conf.Image = "v1.2.3"
				conf.KubernetesVersion = versions.SupportedK8sVersions()[1]
				return conf
			}(),
			currentImageVersion:   "v1.2.2",
			currentClusterVersion: versions.SupportedK8sVersions()[0],
			kubectl: &stubKubectl{
				configMaps: map[string]*corev1.ConfigMap{
					constants.JoinConfigMap: newJoinConfigMap(`{"0":{"expected":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA","warnOnly":false}}`),
				},
			},
			getCRErr: assert.AnError,
			wantErr:  true,
			assertCorrectError: func(t *testing.T, err error) bool {
				return assert.ErrorIs(t, err, assert.AnError)
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
			kubectl: &stubKubectl{
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
			kubectl: &stubKubectl{
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
			kubectl: &stubKubectl{
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
			kubectl: &stubKubectl{
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
			kubectl: &stubKubectl{
				configMaps: map[string]*corev1.ConfigMap{
					constants.JoinConfigMap: newJoinConfigMap(`{"0":{"expected":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA","warnOnly":false}}`),
				},
			},
			wantUpdate: false, // because customClient is used
			customClientFn: func(nodeVersion updatev1alpha1.NodeVersion) unstructuredInterface {
				fakeClient := &fakeUnstructuredClient{}
				fakeClient.On("GetCR", mock.Anything, mock.Anything).Return(unstructedObjectWithGeneration(nodeVersion, 1), nil)
				fakeClient.On("UpdateCR", mock.Anything, mock.Anything).Return(nil, k8serrors.NewConflict(schema.GroupResource{Resource: nodeVersion.Name}, nodeVersion.Name, nil)).Once()
				fakeClient.On("UpdateCR", mock.Anything, mock.Anything).Return(unstructedObjectWithGeneration(nodeVersion, 2), nil).Once()
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
			unstructuredClient := &stubUnstructuredClient{
				object:           &unstructured.Unstructured{Object: unstrNodeVersion},
				badUpdatedObject: badUpdatedObject,
				getCRErr:         tc.getCRErr,
			}
			tc.kubectl.unstructuredInterface = unstructuredClient
			if tc.customClientFn != nil {
				tc.kubectl.unstructuredInterface = tc.customClientFn(nodeVersion)
			}

			upgrader := KubeCmd{
				kubectl:      tc.kubectl,
				imageFetcher: &stubImageFetcher{reference: tc.newImageReference},
				log:          logger.NewTest(t),
				outWriter:    io.Discard,
			}

			err = upgrader.UpgradeNodeVersion(context.Background(), tc.conf, tc.force, false, false)
			// Check upgrades first because if we checked err first, UpgradeImage may error due to other reasons and still trigger an upgrade.
			if tc.wantUpdate {
				assert.NotNil(unstructuredClient.updatedObject)
			} else {
				assert.Nil(unstructuredClient.updatedObject)
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

			upgrader := &KubeCmd{
				log: logger.NewTest(t),
			}

			nodeVersion := updatev1alpha1.NodeVersion{
				Spec: updatev1alpha1.NodeVersionSpec{
					ImageReference: tc.oldImageReference,
					ImageVersion:   tc.oldImageVersion,
				},
			}

			err := upgrader.isValidImageUpgrade(nodeVersion, tc.newImageVersion, false)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
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

			upgrader := &KubeCmd{
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

func TestApplyJoinConfig(t *testing.T) {
	mustMarshal := func(cfg config.AttestationCfg) string {
		data, err := json.Marshal(cfg)
		require.NoError(t, err)
		return string(data)
	}
	// repeatedErrors returns the given error multiple times
	// This is needed in tests, since the retry logic will retry multiple times
	// If the retry limit is raised in [KubeCmd.ApplyJoinConfig], it should also
	// be updated here
	repeatedErrors := func(err error) []error {
		var errs []error
		for i := 0; i < 20; i++ {
			errs = append(errs, err)
		}
		return errs
	}

	testCases := map[string]struct {
		newAttestationCfg config.AttestationCfg
		kubectl           *fakeConfigMapClient
		wantUpdate        bool
		wantErr           bool
	}{
		"success": {
			newAttestationCfg: &config.QEMUVTPM{
				Measurements: measurements.M{
					0: measurements.WithAllBytes(0x00, measurements.WarnOnly, measurements.PCRMeasurementLength),
				},
			},
			kubectl: &fakeConfigMapClient{
				configMaps: map[string]*corev1.ConfigMap{
					constants.JoinConfigMap: newJoinConfigMap(mustMarshal(&config.QEMUVTPM{
						Measurements: measurements.M{
							0: measurements.WithAllBytes(0xFF, measurements.WarnOnly, measurements.PCRMeasurementLength),
						},
					})),
				},
			},
			wantUpdate: true,
		},
		"Get ConfigMap error": {
			newAttestationCfg: &config.QEMUVTPM{
				Measurements: measurements.M{
					0: measurements.WithAllBytes(0x00, measurements.WarnOnly, measurements.PCRMeasurementLength),
				},
			},
			kubectl: &fakeConfigMapClient{
				getErrs: repeatedErrors(assert.AnError),
			},
			wantErr: true,
		},
		"Get ConfigMap fails then returns ConfigMap": {
			newAttestationCfg: &config.QEMUVTPM{
				Measurements: measurements.M{
					0: measurements.WithAllBytes(0x00, measurements.WarnOnly, measurements.PCRMeasurementLength),
				},
			},
			kubectl: &fakeConfigMapClient{
				configMaps: map[string]*corev1.ConfigMap{
					constants.JoinConfigMap: newJoinConfigMap(mustMarshal(&config.QEMUVTPM{
						Measurements: measurements.M{
							0: measurements.WithAllBytes(0xFF, measurements.WarnOnly, measurements.PCRMeasurementLength),
						},
					})),
				},
				getErrs: []error{assert.AnError, assert.AnError},
			},
			wantUpdate: true,
		},
		"Get ConfigMap fails then fails with NotFound": {
			newAttestationCfg: &config.QEMUVTPM{
				Measurements: measurements.M{
					0: measurements.WithAllBytes(0x00, measurements.WarnOnly, measurements.PCRMeasurementLength),
				},
			},
			kubectl: &fakeConfigMapClient{
				getErrs: []error{assert.AnError, assert.AnError, k8serrors.NewNotFound(schema.GroupResource{}, "")},
			},
		},
		"ConfigMap does not exist yet": {
			newAttestationCfg: &config.QEMUVTPM{
				Measurements: measurements.M{
					0: measurements.WithAllBytes(0x00, measurements.WarnOnly, measurements.PCRMeasurementLength),
				},
			},
			kubectl: &fakeConfigMapClient{
				getErrs: repeatedErrors(k8serrors.NewNotFound(schema.GroupResource{}, "")),
			},
		},
		"Create ConfigMap fails": {
			newAttestationCfg: &config.QEMUVTPM{
				Measurements: measurements.M{
					0: measurements.WithAllBytes(0x00, measurements.WarnOnly, measurements.PCRMeasurementLength),
				},
			},
			kubectl: &fakeConfigMapClient{
				getErrs:    repeatedErrors(k8serrors.NewNotFound(schema.GroupResource{}, "")),
				createErrs: repeatedErrors(assert.AnError),
			},
			wantErr: true,
		},
		"Create ConfigMap fails then succeeds": {
			newAttestationCfg: &config.QEMUVTPM{
				Measurements: measurements.M{
					0: measurements.WithAllBytes(0x00, measurements.WarnOnly, measurements.PCRMeasurementLength),
				},
			},
			kubectl: &fakeConfigMapClient{
				getErrs:    repeatedErrors(k8serrors.NewNotFound(schema.GroupResource{}, "")),
				createErrs: []error{assert.AnError, assert.AnError},
			},
		},
		"Update ConfigMap error": {
			newAttestationCfg: &config.QEMUVTPM{
				Measurements: measurements.M{
					0: measurements.WithAllBytes(0x00, measurements.WarnOnly, measurements.PCRMeasurementLength),
				},
			},
			kubectl: &fakeConfigMapClient{
				configMaps: map[string]*corev1.ConfigMap{
					constants.JoinConfigMap: newJoinConfigMap(mustMarshal(&config.QEMUVTPM{
						Measurements: measurements.M{
							0: measurements.WithAllBytes(0xFF, measurements.WarnOnly, measurements.PCRMeasurementLength),
						},
					})),
				},
				updateErrs: repeatedErrors(assert.AnError),
			},
			wantErr: true,
		},
		"Update ConfigMap fails then succeeds": {
			newAttestationCfg: &config.QEMUVTPM{
				Measurements: measurements.M{
					0: measurements.WithAllBytes(0x00, measurements.WarnOnly, measurements.PCRMeasurementLength),
				},
			},
			kubectl: &fakeConfigMapClient{
				configMaps: map[string]*corev1.ConfigMap{
					constants.JoinConfigMap: newJoinConfigMap(mustMarshal(&config.QEMUVTPM{
						Measurements: measurements.M{
							0: measurements.WithAllBytes(0xFF, measurements.WarnOnly, measurements.PCRMeasurementLength),
						},
					})),
				},
				updateErrs: []error{assert.AnError, assert.AnError},
			},
			wantUpdate: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			cmd := &KubeCmd{
				kubectl:       tc.kubectl,
				log:           logger.NewTest(t),
				retryInterval: time.Millisecond,
				outWriter:     io.Discard,
			}

			err := cmd.ApplyJoinConfig(context.Background(), tc.newAttestationCfg, []byte{0x11})
			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)

			var cfg *corev1.ConfigMap
			var ok bool
			if tc.wantUpdate {
				cfg, ok = tc.kubectl.updatedConfigMaps[constants.JoinConfigMap]
			} else {
				cfg, ok = tc.kubectl.configMaps[constants.JoinConfigMap]
			}
			require.True(ok)
			assert.Equal(mustMarshal(tc.newAttestationCfg), cfg.Data[constants.AttestationConfigFilename])
		})
	}
}

type fakeUnstructuredClient struct {
	mock.Mock
}

func (u *fakeUnstructuredClient) GetCR(ctx context.Context, _ schema.GroupVersionResource, str string) (*unstructured.Unstructured, error) {
	args := u.Called(ctx, str)
	return args.Get(0).(*unstructured.Unstructured), args.Error(1)
}

func (u *fakeUnstructuredClient) UpdateCR(ctx context.Context, _ schema.GroupVersionResource, updatedObject *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	args := u.Called(ctx, updatedObject)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return updatedObject, args.Error(1)
}

type stubUnstructuredClient struct {
	object           *unstructured.Unstructured
	updatedObject    *unstructured.Unstructured
	badUpdatedObject *unstructured.Unstructured
	getCRErr         error
	updateCRErr      error
}

func (u *stubUnstructuredClient) GetCR(_ context.Context, _ schema.GroupVersionResource, _ string) (*unstructured.Unstructured, error) {
	return u.object, u.getCRErr
}

func (u *stubUnstructuredClient) UpdateCR(_ context.Context, _ schema.GroupVersionResource, updatedObject *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	u.updatedObject = updatedObject
	if u.badUpdatedObject != nil {
		return u.badUpdatedObject, u.updateCRErr
	}
	return u.updatedObject, u.updateCRErr
}

type unstructuredInterface interface {
	GetCR(ctx context.Context, gvr schema.GroupVersionResource, name string) (*unstructured.Unstructured, error)
	UpdateCR(ctx context.Context, gvr schema.GroupVersionResource, obj *unstructured.Unstructured) (*unstructured.Unstructured, error)
}

type stubKubectl struct {
	unstructuredInterface
	configMaps        map[string]*corev1.ConfigMap
	updatedConfigMaps map[string]*corev1.ConfigMap
	k8sVersion        string
	getCMErr          error
	updateCMErr       error
	createCMErr       error
	k8sErr            error
	nodes             []corev1.Node
	nodesErr          error
	crds              []apiextensionsv1.CustomResourceDefinition
	getCRDsError      error
	crs               []unstructured.Unstructured
	getCRsError       error
}

func (s *stubKubectl) GetConfigMap(_ context.Context, _, name string) (*corev1.ConfigMap, error) {
	return s.configMaps[name], s.getCMErr
}

func (s *stubKubectl) UpdateConfigMap(_ context.Context, configMap *corev1.ConfigMap) (*corev1.ConfigMap, error) {
	if s.updatedConfigMaps == nil {
		s.updatedConfigMaps = map[string]*corev1.ConfigMap{}
	}
	s.updatedConfigMaps[configMap.ObjectMeta.Name] = configMap
	return s.updatedConfigMaps[configMap.ObjectMeta.Name], s.updateCMErr
}

func (s *stubKubectl) CreateConfigMap(_ context.Context, configMap *corev1.ConfigMap) error {
	if s.configMaps == nil {
		s.configMaps = map[string]*corev1.ConfigMap{}
	}
	s.configMaps[configMap.ObjectMeta.Name] = configMap
	return s.createCMErr
}

func (s *stubKubectl) KubernetesVersion() (string, error) {
	return s.k8sVersion, s.k8sErr
}

func (s *stubKubectl) GetNodes(_ context.Context) ([]corev1.Node, error) {
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

type fakeConfigMapClient struct {
	getErrs           []error
	updatedConfigMaps map[string]*corev1.ConfigMap
	updateErrs        []error
	configMaps        map[string]*corev1.ConfigMap
	createErrs        []error
	kubectlInterface
}

func (f *fakeConfigMapClient) GetConfigMap(_ context.Context, _, name string) (*corev1.ConfigMap, error) {
	if len(f.getErrs) > 0 {
		err := f.getErrs[0]
		if len(f.getErrs) > 1 {
			f.getErrs = f.getErrs[1:]
		} else {
			f.getErrs = nil
		}
		return nil, err
	}
	return f.configMaps[name], nil
}

func (f *fakeConfigMapClient) UpdateConfigMap(_ context.Context, configMap *corev1.ConfigMap) (*corev1.ConfigMap, error) {
	if len(f.updateErrs) > 0 {
		err := f.updateErrs[0]
		if len(f.updateErrs) > 1 {
			f.updateErrs = f.updateErrs[1:]
		} else {
			f.updateErrs = nil
		}
		return nil, err
	}

	if f.updatedConfigMaps == nil {
		f.updatedConfigMaps = map[string]*corev1.ConfigMap{}
	}
	f.updatedConfigMaps[configMap.ObjectMeta.Name] = configMap
	return f.updatedConfigMaps[configMap.ObjectMeta.Name], nil
}

func (f *fakeConfigMapClient) CreateConfigMap(_ context.Context, configMap *corev1.ConfigMap) error {
	if len(f.createErrs) > 0 {
		err := f.createErrs[0]
		if len(f.createErrs) > 1 {
			f.createErrs = f.createErrs[1:]
		} else {
			f.createErrs = nil
		}
		return err
	}

	if f.configMaps == nil {
		f.configMaps = map[string]*corev1.ConfigMap{}
	}
	f.configMaps[configMap.ObjectMeta.Name] = configMap
	return nil
}
