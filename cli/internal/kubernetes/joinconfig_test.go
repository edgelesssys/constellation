/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package kubernetes

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestJoinConfigMapClient_Backup(t *testing.T) {
	mockClient := &mockConfigMapGetterAndCreater{}
	mockClient.On("GetConfigMap", mock.Anything, mock.Anything).Return(&corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: "join-config",
		},
	}, nil)
	mockClient.On("CreateConfigMap", mock.Anything, mock.MatchedBy(func(cm *corev1.ConfigMap) bool {
		return cm.Name == "join-config-backup" && cm.ResourceVersion == ""
	})).Return(&corev1.ConfigMap{}, nil)

	sut := NewJoinConfigMapClient(mockClient)
	cm, err := sut.Get(context.Background())
	assert.NoError(t, err)
	err = sut.Backup(context.Background(), cm)
	assert.NoError(t, err)
}

type mockConfigMapGetterAndCreater struct {
	mock.Mock
}

func (m *mockConfigMapGetterAndCreater) GetConfigMap(ctx context.Context, name string) (*corev1.ConfigMap, error) {
	args := m.Called(ctx, name)
	return args.Get(0).(*corev1.ConfigMap), args.Error(1)
}

func (m *mockConfigMapGetterAndCreater) CreateConfigMap(ctx context.Context, cm *corev1.ConfigMap) (*corev1.ConfigMap, error) {
	args := m.Called(ctx, cm)
	return args.Get(0).(*corev1.ConfigMap), args.Error(1)
}
