/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cloudcmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestUpdateMeasurements(t *testing.T) {
	someErr := errors.New("error")
	testCases := map[string]struct {
		updater         *stubMeasurementsUpdater
		newMeasurements measurements.M
		wantUpdate      bool
		wantErr         bool
	}{
		"success": {
			updater: &stubMeasurementsUpdater{
				oldMeasurements: &corev1.ConfigMap{
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
			updater: &stubMeasurementsUpdater{
				oldMeasurements: &corev1.ConfigMap{
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
			updater: &stubMeasurementsUpdater{
				oldMeasurements: &corev1.ConfigMap{
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
			updater: &stubMeasurementsUpdater{
				oldMeasurements: &corev1.ConfigMap{
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
			updater: &stubMeasurementsUpdater{getErr: someErr},
			wantErr: true,
		},
		"update error": {
			updater: &stubMeasurementsUpdater{
				oldMeasurements: &corev1.ConfigMap{
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
				measurementsUpdater: tc.updater,
				outWriter:           &bytes.Buffer{},
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
				assert.JSONEq(string(newMeasurementsJSON), tc.updater.updatedMeasurements.Data[constants.MeasurementsFilename])
			} else {
				assert.Nil(tc.updater.updatedMeasurements)
			}
		})
	}
}

type stubMeasurementsUpdater struct {
	oldMeasurements     *corev1.ConfigMap
	updatedMeasurements *corev1.ConfigMap
	getErr              error
	updateErr           error
}

func (u *stubMeasurementsUpdater) getCurrent(context.Context, string) (*corev1.ConfigMap, error) {
	return u.oldMeasurements, u.getErr
}

func (u *stubMeasurementsUpdater) update(_ context.Context, updatedMeasurements *corev1.ConfigMap) (*corev1.ConfigMap, error) {
	u.updatedMeasurements = updatedMeasurements
	return nil, u.updateErr
}

func TestUpdateImage(t *testing.T) {
	someErr := errors.New("error")
	testCases := map[string]struct {
		updater    *stubImageUpdater
		newImage   string
		wantUpdate bool
		wantErr    bool
	}{
		"success": {
			updater: &stubImageUpdater{
				setImage: &unstructured.Unstructured{
					Object: map[string]any{
						"spec": map[string]any{
							"image": "old-image",
						},
					},
				},
			},
			newImage:   "new-image",
			wantUpdate: true,
		},
		"image is the same": {
			updater: &stubImageUpdater{
				setImage: &unstructured.Unstructured{
					Object: map[string]any{
						"spec": map[string]any{
							"image": "old-image",
						},
					},
				},
			},
			newImage: "old-image",
		},
		"getCurrent error": {
			updater: &stubImageUpdater{getErr: someErr},
			wantErr: true,
		},
		"update error": {
			updater: &stubImageUpdater{
				setImage: &unstructured.Unstructured{
					Object: map[string]any{
						"spec": map[string]any{
							"image": "old-image",
						},
					},
				},
				updateErr: someErr,
			},
			newImage: "new-image",
			wantErr:  true,
		},
		"no spec": {
			updater: &stubImageUpdater{
				setImage: &unstructured.Unstructured{
					Object: map[string]any{},
				},
			},
			newImage: "new-image",
			wantErr:  true,
		},
		"not a map": {
			updater: &stubImageUpdater{
				setImage: &unstructured.Unstructured{
					Object: map[string]any{
						"spec": "not a map",
					},
				},
			},
			newImage: "new-image",
			wantErr:  true,
		},
		"no spec.image": {
			updater: &stubImageUpdater{
				setImage: &unstructured.Unstructured{
					Object: map[string]any{
						"spec": map[string]any{},
					},
				},
			},
			newImage: "new-image",
			wantErr:  true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			upgrader := &Upgrader{
				imageUpdater: tc.updater,
				outWriter:    &bytes.Buffer{},
			}

			err := upgrader.updateImage(context.Background(), tc.newImage)

			if tc.wantErr {
				assert.Error(err)
				return
			}

			assert.NoError(err)
			if tc.wantUpdate {
				assert.Equal(tc.newImage, tc.updater.updatedImage.Object["spec"].(map[string]any)["image"])
			} else {
				assert.Nil(tc.updater.updatedImage)
			}
		})
	}
}

type stubImageUpdater struct {
	setImage     *unstructured.Unstructured
	updatedImage *unstructured.Unstructured
	getErr       error
	updateErr    error
}

func (u *stubImageUpdater) getCurrent(ctx context.Context, name string) (*unstructured.Unstructured, error) {
	return u.setImage, u.getErr
}

func (u *stubImageUpdater) update(_ context.Context, updatedImage *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	u.updatedImage = updatedImage
	return nil, u.updateErr
}
