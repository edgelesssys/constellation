/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cloudcmd

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/versions/components"
	updatev1alpha1 "github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/v2/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestUpgradeK8s(t *testing.T) {
	someErr := errors.New("some error")
	testCases := map[string]struct {
		stable                      stubStableClient
		conditions                  []metav1.Condition
		activeClusterVersionUpgrade bool
		newClusterVersion           string
		currentClusterVersion       string
		components                  components.Components
		getErr                      error
		assertCorrectError          func(t *testing.T, err error) bool
		wantErr                     bool
	}{
		"success": {
			currentClusterVersion: "v1.2.2",
			newClusterVersion:     "v1.2.3",
		},
		"not an upgrade": {
			currentClusterVersion: "v1.2.3",
			newClusterVersion:     "v1.2.3",
			wantErr:               true,
			assertCorrectError: func(t *testing.T, err error) bool {
				target := &InvalidUpgradeError{}
				return assert.ErrorAs(t, err, &target)
			},
		},
		"downgrade": {
			currentClusterVersion: "v1.2.3",
			newClusterVersion:     "v1.2.2",
			wantErr:               true,
			assertCorrectError: func(t *testing.T, err error) bool {
				target := &InvalidUpgradeError{}
				return assert.ErrorAs(t, err, &target)
			},
		},
		"no constellation-version object": {
			getErr:  someErr,
			wantErr: true,
			assertCorrectError: func(t *testing.T, err error) bool {
				return assert.ErrorIs(t, err, someErr)
			},
		},
		"upgrade in progress": {
			currentClusterVersion: "v1.2.2",
			newClusterVersion:     "v1.2.3",
			conditions: []metav1.Condition{{
				Type:   updatev1alpha1.ConditionOutdated,
				Status: metav1.ConditionTrue,
			}},
			wantErr: true,
			assertCorrectError: func(t *testing.T, err error) bool {
				return assert.ErrorIs(t, err, ErrInProgress)
			},
		},
		"configmap create fails": {
			currentClusterVersion: "v1.2.2",
			newClusterVersion:     "v1.2.3",
			stable: stubStableClient{
				createErr: someErr,
			},
			wantErr: true,
			assertCorrectError: func(t *testing.T, err error) bool {
				return assert.ErrorIs(t, err, someErr)
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)
			nodeVersion := updatev1alpha1.NodeVersion{
				Spec: updatev1alpha1.NodeVersionSpec{
					KubernetesClusterVersion: tc.currentClusterVersion,
				},
				Status: updatev1alpha1.NodeVersionStatus{
					Conditions:                  tc.conditions,
					ActiveClusterVersionUpgrade: tc.activeClusterVersionUpgrade,
				},
			}

			unstrNodeVersion, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&nodeVersion)
			require.NoError(err)

			upgrader := Upgrader{
				stableInterface:  &tc.stable,
				dynamicInterface: &stubDynamicClient{object: &unstructured.Unstructured{Object: unstrNodeVersion}, getErr: tc.getErr},
				log:              logger.NewTest(t),
				outWriter:        io.Discard,
			}

			err = upgrader.UpgradeK8s(context.Background(), tc.newClusterVersion, tc.components)
			if tc.wantErr {
				tc.assertCorrectError(t, err)
				return
			}
			assert.NoError(err)
		})
	}
}

func TestUpgradeImage(t *testing.T) {
	someErr := errors.New("some error")
	testCases := map[string]struct {
		stable              *stubStableClient
		conditions          []metav1.Condition
		currentImageVersion string
		newImageVersion     string
		getErr              error
		wantErr             bool
		wantUpdate          bool
		assertCorrectError  func(t *testing.T, err error) bool
	}{
		"success": {
			currentImageVersion: "v1.2.2",
			newImageVersion:     "v1.2.3",
			stable: &stubStableClient{
				configMap: &corev1.ConfigMap{
					Data: map[string]string{
						constants.MeasurementsFilename: `{"0":{"expected":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA","warnOnly":false}}`,
					},
				},
			},
			wantUpdate: true,
		},
		"not an upgrade": {
			currentImageVersion: "v1.2.2",
			newImageVersion:     "v1.2.2",
			wantErr:             true,
			assertCorrectError: func(t *testing.T, err error) bool {
				target := &InvalidUpgradeError{}
				return assert.ErrorAs(t, err, &target)
			},
		},
		"downgrade": {
			currentImageVersion: "v1.2.2",
			newImageVersion:     "v1.2.1",
			wantErr:             true,
			assertCorrectError: func(t *testing.T, err error) bool {
				target := &InvalidUpgradeError{}
				return assert.ErrorAs(t, err, &target)
			},
		},
		"upgrade in progress": {
			currentImageVersion: "v1.2.2",
			newImageVersion:     "v1.2.3",
			conditions: []metav1.Condition{{
				Type:   updatev1alpha1.ConditionOutdated,
				Status: metav1.ConditionTrue,
			}},
			wantErr: true,
			assertCorrectError: func(t *testing.T, err error) bool {
				return assert.ErrorIs(t, err, ErrInProgress)
			},
		},
		"get error": {
			getErr:  someErr,
			wantErr: true,
			assertCorrectError: func(t *testing.T, err error) bool {
				return assert.ErrorIs(t, err, someErr)
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			nodeVersion := updatev1alpha1.NodeVersion{
				Spec: updatev1alpha1.NodeVersionSpec{
					ImageVersion: tc.currentImageVersion,
				},
				Status: updatev1alpha1.NodeVersionStatus{
					Conditions: tc.conditions,
				},
			}

			unstrNodeVersion, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&nodeVersion)
			require.NoError(err)

			dynamicClient := &stubDynamicClient{object: &unstructured.Unstructured{Object: unstrNodeVersion}, getErr: tc.getErr}
			upgrader := Upgrader{
				stableInterface:  tc.stable,
				dynamicInterface: dynamicClient,
				log:              logger.NewTest(t),
				outWriter:        io.Discard,
			}

			err = upgrader.UpgradeImage(context.Background(), "", tc.newImageVersion, nil)

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
		updater         *stubStableClient
		newMeasurements measurements.M
		wantUpdate      bool
		wantErr         bool
	}{
		"success": {
			updater: &stubStableClient{
				configMap: &corev1.ConfigMap{
					Data: map[string]string{
						constants.MeasurementsFilename: `{"0":{"expected":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA","warnOnly":false}}`,
					},
				},
			},
			newMeasurements: measurements.M{
				0: measurements.WithAllBytes(0xBB, false),
			},
			wantUpdate: true,
		},
		"measurements are the same": {
			updater: &stubStableClient{
				configMap: &corev1.ConfigMap{
					Data: map[string]string{
						constants.MeasurementsFilename: `{"0":{"expected":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA","warnOnly":false}}`,
					},
				},
			},
			newMeasurements: measurements.M{
				0: measurements.WithAllBytes(0xAA, false),
			},
		},
		"trying to set warnOnly to true results in error": {
			updater: &stubStableClient{
				configMap: &corev1.ConfigMap{
					Data: map[string]string{
						constants.MeasurementsFilename: `{"0":{"expected":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA","warnOnly":false}}`,
					},
				},
			},
			newMeasurements: measurements.M{
				0: measurements.WithAllBytes(0xAA, true),
			},
			wantErr: true,
		},
		"setting warnOnly to false is allowed": {
			updater: &stubStableClient{
				configMap: &corev1.ConfigMap{
					Data: map[string]string{
						constants.MeasurementsFilename: `{"0":{"expected":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA","warnOnly":true}}`,
					},
				},
			},
			newMeasurements: measurements.M{
				0: measurements.WithAllBytes(0xAA, false),
			},
			wantUpdate: true,
		},
		"getCurrent error": {
			updater: &stubStableClient{getErr: someErr},
			wantErr: true,
		},
		"update error": {
			updater: &stubStableClient{
				configMap: &corev1.ConfigMap{
					Data: map[string]string{
						constants.MeasurementsFilename: `{"0":{"expected":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA","warnOnly":false}}`,
					},
				},
				updateErr: someErr,
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

			err := upgrader.updateMeasurements(context.Background(), tc.newMeasurements)
			if tc.wantErr {
				assert.Error(err)
				return
			}

			assert.NoError(err)
			if tc.wantUpdate {
				newMeasurementsJSON, err := json.Marshal(tc.newMeasurements)
				require.NoError(t, err)
				assert.JSONEq(string(newMeasurementsJSON), tc.updater.updatedConfigMap.Data[constants.MeasurementsFilename])
			} else {
				assert.Nil(tc.updater.updatedConfigMap)
			}
		})
	}
}

func TestUpdateImage(t *testing.T) {
	someErr := errors.New("error")
	testCases := map[string]struct {
		nodeVersion       updatev1alpha1.NodeVersion
		newImageReference string
		newImageVersion   string
		oldImageVersion   string
		updateErr         error
		wantUpdate        bool
		wantErr           bool
	}{
		"success": {
			nodeVersion: updatev1alpha1.NodeVersion{
				Spec: updatev1alpha1.NodeVersionSpec{
					ImageReference: "old-image-ref",
					ImageVersion:   "old-image-ver",
				},
			},
			newImageReference: "new-image-ref",
			newImageVersion:   "new-image-ver",
			wantUpdate:        true,
		},
		"update error": {
			updateErr: someErr,
			wantErr:   true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			upgradeClient := &stubDynamicClient{updateErr: tc.updateErr}
			upgrader := &Upgrader{
				dynamicInterface: upgradeClient,
				outWriter:        io.Discard,
				log:              logger.NewTest(t),
			}

			err := upgrader.updateImage(context.Background(), tc.nodeVersion, tc.newImageReference, tc.newImageVersion)

			if tc.wantErr {
				assert.Error(err)
				return
			}

			assert.NoError(err)
			if tc.wantUpdate {
				assert.Equal(tc.newImageReference, upgradeClient.updatedObject.Object["spec"].(map[string]any)["image"])
				assert.Equal(tc.newImageVersion, upgradeClient.updatedObject.Object["spec"].(map[string]any)["imageVersion"])
			} else {
				assert.Nil(upgradeClient.updatedObject)
			}
		})
	}
}

type stubDynamicClient struct {
	object        *unstructured.Unstructured
	updatedObject *unstructured.Unstructured
	getErr        error
	updateErr     error
}

func (u *stubDynamicClient) getCurrent(ctx context.Context, name string) (*unstructured.Unstructured, error) {
	return u.object, u.getErr
}

func (u *stubDynamicClient) update(_ context.Context, updatedObject *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	u.updatedObject = updatedObject
	return u.updatedObject, u.updateErr
}

type stubStableClient struct {
	configMap        *corev1.ConfigMap
	updatedConfigMap *corev1.ConfigMap
	k8sVersion       string
	getErr           error
	updateErr        error
	createErr        error
	k8sErr           error
}

func (s *stubStableClient) getCurrentConfigMap(ctx context.Context, name string) (*corev1.ConfigMap, error) {
	return s.configMap, s.getErr
}

func (s *stubStableClient) updateConfigMap(ctx context.Context, configMap *corev1.ConfigMap) (*corev1.ConfigMap, error) {
	s.updatedConfigMap = configMap
	return nil, s.updateErr
}

func (s *stubStableClient) createConfigMap(ctx context.Context, configMap *corev1.ConfigMap) (*corev1.ConfigMap, error) {
	s.configMap = configMap
	return s.configMap, s.createErr
}

func (s *stubStableClient) kubernetesVersion() (string, error) {
	return s.k8sVersion, s.k8sErr
}
