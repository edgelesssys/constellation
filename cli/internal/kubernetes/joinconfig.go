/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package kubernetes

import (
	"context"
	"fmt"

	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	corev1 "k8s.io/api/core/v1"
)

type configMapGetterAndCreater interface {
	GetConfigMap(ctx context.Context, name string) (*corev1.ConfigMap, error)
	CreateConfigMap(ctx context.Context, cm *corev1.ConfigMap) (*corev1.ConfigMap, error)
}

// NewJoinConfigMapClient returns a new JoinConfigMapClient.
func NewJoinConfigMapClient(stableClient configMapGetterAndCreater) *JoinConfigMapClient {
	return &JoinConfigMapClient{stableClient: stableClient}
}

// JoinConfigMapClient is a client for the JoinConfigMap.
type JoinConfigMapClient struct {
	stableClient configMapGetterAndCreater
}

// Get returns the JoinConfigMap.
func (c *JoinConfigMapClient) Get(ctx context.Context) (*JoinConfigMap, error) {
	cm, err := c.stableClient.GetConfigMap(ctx, constants.JoinConfigMap)
	if err != nil {
		return nil, err
	}
	return &JoinConfigMap{cm}, nil
}

// Backup creates a backup of the JoinConfigMap.
func (c *JoinConfigMapClient) Backup(ctx context.Context, clusterJoinConfig *JoinConfigMap) error {
	joinConfigBackup := clusterJoinConfig.ConfigMap.DeepCopy()
	joinConfigBackup.Name = fmt.Sprintf("%s-backup", joinConfigBackup.Name)
	joinConfigBackup.ResourceVersion = "" // reset resource version to create a new resource
	if _, err := c.stableClient.CreateConfigMap(ctx, joinConfigBackup); err != nil {
		return fmt.Errorf("setting new attestation config: %w", err)
	}
	return nil
}

// JoinConfigMap is a wrapper around a ConfigMap.
type JoinConfigMap struct {
	*corev1.ConfigMap
}

// GetAttestationConfig returns the attestation config from the JoinConfigMap.
func (j *JoinConfigMap) GetAttestationConfig(attestVariant variant.Variant) (config.AttestationCfg, error) {
	rawAttestationConfig, ok := j.Data[constants.AttestationConfigFilename]
	if !ok {
		return nil, fmt.Errorf("attestationConfig not found in %s", constants.JoinConfigMap)
	}
	attestationConfig, err := config.UnmarshalAttestationConfig([]byte(rawAttestationConfig), attestVariant)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling attestation config: %w", err)
	}
	return attestationConfig, nil
}
