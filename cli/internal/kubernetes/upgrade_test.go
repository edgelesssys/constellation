/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package kubernetes

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/compatibility"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/versions"
	"github.com/edgelesssys/constellation/v2/internal/versions/components"
	updatev1alpha1 "github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/v2/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestUpgradeNodeVersion(t *testing.T) {
	someErr := errors.New("some error")
	testCases := map[string]struct {
		stable                *stubStableClient
		conditions            []metav1.Condition
		currentImageVersion   string
		newImageReference     string
		badImageVersion       string
		currentClusterVersion string
		conf                  *config.Config
		getErr                error
		wantErr               bool
		wantUpdate            bool
		assertCorrectError    func(t *testing.T, err error) bool
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
			stable: &stubStableClient{
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
			stable: &stubStableClient{
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
			stable: &stubStableClient{
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
			stable:                &stubStableClient{},
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
			stable:                &stubStableClient{},
			wantErr:               true,
			assertCorrectError: func(t *testing.T, err error) bool {
				return assert.ErrorIs(t, err, ErrInProgress)
			},
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
			stable: &stubStableClient{
				configMaps: map[string]*corev1.ConfigMap{
					constants.JoinConfigMap: newJoinConfigMap(`{"0":{"expected":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA","warnOnly":false}}`),
				},
			},
			wantUpdate: true,
			wantErr:    true,
			assertCorrectError: func(t *testing.T, err error) bool {
				upgradeErr := &compatibility.InvalidUpgradeError{}
				return assert.ErrorAs(t, err, &upgradeErr)
			},
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
			stable: &stubStableClient{
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

			unstrNodeVersion, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&nodeVersion)
			require.NoError(err)

			var badUpdatedObject *unstructured.Unstructured
			if tc.badImageVersion != "" {
				nodeVersion.Spec.ImageVersion = tc.badImageVersion
				unstrBadNodeVersion, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&nodeVersion)
				require.NoError(err)
				badUpdatedObject = &unstructured.Unstructured{Object: unstrBadNodeVersion}
			}

			dynamicClient := &stubDynamicClient{object: &unstructured.Unstructured{Object: unstrNodeVersion}, badUpdatedObject: badUpdatedObject, getErr: tc.getErr}
			upgrader := Upgrader{
				stableInterface:  tc.stable,
				dynamicInterface: dynamicClient,
				imageFetcher:     &stubImageFetcher{reference: tc.newImageReference},
				log:              logger.NewTest(t),
				outWriter:        io.Discard,
			}

			err = upgrader.UpgradeNodeVersion(context.Background(), tc.conf)

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

func TestUpdateMeasurements(t *testing.T) {
	someErr := errors.New("error")
	testCases := map[string]struct {
		updater    *stubStableClient
		newConfig  config.AttestationCfg
		wantUpdate bool
		wantErr    bool
	}{
		"success": {
			updater: &stubStableClient{
				configMaps: map[string]*corev1.ConfigMap{
					constants.JoinConfigMap: newJoinConfigMap(`{"measurements":{"0":{"expected":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA","warnOnly":false}}}`),
				},
			},
			newConfig: &config.GCPSEVES{
				Measurements: measurements.M{
					0: measurements.WithAllBytes(0xBB, measurements.Enforce),
				},
			},
			wantUpdate: true,
		},
		"measurements are the same": {
			updater: &stubStableClient{
				configMaps: map[string]*corev1.ConfigMap{
					constants.JoinConfigMap: newJoinConfigMap(`{"measurements":{"0":{"expected":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA","warnOnly":false}}}`),
				},
			},
			newConfig: &config.GCPSEVES{
				Measurements: measurements.M{
					0: measurements.WithAllBytes(0xAA, measurements.Enforce),
				},
			},
		},
		"setting warnOnly to true is allowed": {
			updater: &stubStableClient{
				configMaps: map[string]*corev1.ConfigMap{
					constants.JoinConfigMap: newJoinConfigMap(`{"measurements":{"0":{"expected":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA","warnOnly":false}}}`),
				},
			},
			newConfig: &config.GCPSEVES{
				Measurements: measurements.M{
					0: measurements.WithAllBytes(0xAA, measurements.WarnOnly),
				},
			},
			wantUpdate: true,
		},
		"setting warnOnly to false is allowed": {
			updater: &stubStableClient{
				configMaps: map[string]*corev1.ConfigMap{
					constants.JoinConfigMap: newJoinConfigMap(`{"measurements":{"0":{"expected":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA","warnOnly":true}}}`),
				},
			},
			newConfig: &config.GCPSEVES{
				Measurements: measurements.M{
					0: measurements.WithAllBytes(0xAA, measurements.Enforce),
				},
			},
			wantUpdate: true,
		},
		"getCurrent error": {
			updater: &stubStableClient{getErr: someErr},
			newConfig: &config.GCPSEVES{
				Measurements: measurements.M{
					0: measurements.WithAllBytes(0xBB, measurements.Enforce),
				},
			},
			wantErr: true,
		},
		"update error": {
			updater: &stubStableClient{
				configMaps: map[string]*corev1.ConfigMap{
					constants.JoinConfigMap: newJoinConfigMap(`{"measurements":{"0":{"expected":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA","warnOnly":false}}}`),
				},
				updateErr: someErr,
			},
			newConfig: &config.GCPSEVES{
				Measurements: measurements.M{
					0: measurements.WithAllBytes(0xBB, measurements.Enforce),
				},
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			upgrader := &Upgrader{
				stableInterface: tc.updater,
				outWriter:       io.Discard,
				log:             logger.NewTest(t),
			}

			err := upgrader.UpdateAttestationConfig(context.Background(), tc.newConfig)
			if tc.wantErr {
				assert.Error(err)
				return
			}

			assert.NoError(err)
			if tc.wantUpdate {
				newConfigJSON, err := json.Marshal(tc.newConfig)
				require.NoError(t, err)
				assert.JSONEq(string(newConfigJSON), tc.updater.updatedConfigMaps[constants.JoinConfigMap].Data[constants.AttestationConfigFilename])
			} else {
				assert.Nil(tc.updater.updatedConfigMaps)
			}
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

			err := upgrader.updateImage(&nodeVersion, tc.newImageReference, tc.newImageVersion)

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

			_, err := upgrader.updateK8s(&nodeVersion, tc.newClusterVersion, components.Components{})

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

type stubStableClient struct {
	configMaps        map[string]*corev1.ConfigMap
	updatedConfigMaps map[string]*corev1.ConfigMap
	k8sVersion        string
	getErr            error
	updateErr         error
	createErr         error
	k8sErr            error
}

func (s *stubStableClient) getCurrentConfigMap(_ context.Context, name string) (*corev1.ConfigMap, error) {
	return s.configMaps[name], s.getErr
}

func (s *stubStableClient) updateConfigMap(_ context.Context, configMap *corev1.ConfigMap) (*corev1.ConfigMap, error) {
	if s.updatedConfigMaps == nil {
		s.updatedConfigMaps = map[string]*corev1.ConfigMap{}
	}
	s.updatedConfigMaps[configMap.ObjectMeta.Name] = configMap
	return s.updatedConfigMaps[configMap.ObjectMeta.Name], s.updateErr
}

func (s *stubStableClient) createConfigMap(_ context.Context, configMap *corev1.ConfigMap) (*corev1.ConfigMap, error) {
	if s.configMaps == nil {
		s.configMaps = map[string]*corev1.ConfigMap{}
	}
	s.configMaps[configMap.ObjectMeta.Name] = configMap
	return s.configMaps[configMap.ObjectMeta.Name], s.createErr
}

func (s *stubStableClient) kubernetesVersion() (string, error) {
	return s.k8sVersion, s.k8sErr
}

type stubImageFetcher struct {
	reference         string
	fetchReferenceErr error
}

func (f *stubImageFetcher) FetchReference(_ context.Context, _ *config.Config) (string, error) {
	return f.reference, f.fetchReferenceErr
}
