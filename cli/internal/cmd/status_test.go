/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/constellation/kubecmd"
	"github.com/edgelesssys/constellation/v2/internal/file"
	updatev1alpha1 "github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/v2/api/v1alpha1"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const successOutput = targetVersions + versionsOutput + nodesUpToDateOutput + attestationConfigOutput

const inProgressOutput = targetVersions + versionsOutput + nodesInProgressOutput + attestationConfigOutput

const targetVersions = `Target versions:
	Image: v1.1.0
	Kubernetes: v1.2.3
`

const nodesUpToDateOutput = `Cluster status: Node version of every node is up to date
`

const nodesInProgressOutput = `Cluster status: Some node versions are out of date
	Image: 1/2
	Kubernetes: 1/2
`

const versionsOutput = `Service versions:
	Cilium: v1.0.0
	cert-manager: v1.0.0
	constellation-operators: v1.1.0
	constellation-services: v1.1.0
`

const attestationConfigOutput = `Attestation config:
	measurements:
	    15:
	        expected: "0000000000000000000000000000000000000000000000000000000000000000"
	        warnOnly: false
`

// TestStatus checks that the status function produces the correct strings.
func TestStatus(t *testing.T) {
	mustParseNodeVersion := func(nV updatev1alpha1.NodeVersion) kubecmd.NodeVersion {
		nodeVersion, err := kubecmd.NewNodeVersion(nV)
		require.NoError(t, err)
		return nodeVersion
	}

	testCases := map[string]struct {
		kubeClient     stubKubeClient
		expectedOutput string
		wantErr        bool
	}{
		"success": {
			kubeClient: stubKubeClient{
				status: map[string]kubecmd.NodeStatus{
					"outdated": kubecmd.NewNodeStatus(corev1.Node{
						ObjectMeta: metav1.ObjectMeta{
							Name: "node1",
							Annotations: map[string]string{
								"constellation.edgeless.systems/node-image": "v1.1.0",
							},
						},
						Status: corev1.NodeStatus{
							NodeInfo: corev1.NodeSystemInfo{
								KubeletVersion: "v1.2.3",
							},
						},
					}),
				},
				version: mustParseNodeVersion(updatev1alpha1.NodeVersion{
					Spec: updatev1alpha1.NodeVersionSpec{
						ImageVersion:             "v1.1.0",
						ImageReference:           "v1.1.0",
						KubernetesClusterVersion: "v1.2.3",
					},
					Status: updatev1alpha1.NodeVersionStatus{
						Conditions: []metav1.Condition{
							{
								Message: "Node version of every node is up to date",
							},
						},
					},
				}),
				attestation: &config.QEMUVTPM{
					Measurements: measurements.M{
						15: measurements.WithAllBytes(0, measurements.Enforce, measurements.PCRMeasurementLength),
					},
				},
			},
			expectedOutput: successOutput,
		},
		"one of two nodes not upgraded": {
			kubeClient: stubKubeClient{
				status: map[string]kubecmd.NodeStatus{
					"outdated": kubecmd.NewNodeStatus(corev1.Node{
						ObjectMeta: metav1.ObjectMeta{
							Name: "outdated",
							Annotations: map[string]string{
								"constellation.edgeless.systems/node-image": "v1.0.0",
							},
						},
						Status: corev1.NodeStatus{
							NodeInfo: corev1.NodeSystemInfo{
								KubeletVersion: "v1.2.2",
							},
						},
					}),
					"uptodate": kubecmd.NewNodeStatus(corev1.Node{
						ObjectMeta: metav1.ObjectMeta{
							Name: "uptodate",
							Annotations: map[string]string{
								"constellation.edgeless.systems/node-image": "v1.1.0",
							},
						},
						Status: corev1.NodeStatus{
							NodeInfo: corev1.NodeSystemInfo{
								KubeletVersion: "v1.2.3",
							},
						},
					}),
				},
				version: mustParseNodeVersion(updatev1alpha1.NodeVersion{
					Spec: updatev1alpha1.NodeVersionSpec{
						ImageVersion:             "v1.1.0",
						ImageReference:           "v1.1.0",
						KubernetesClusterVersion: "v1.2.3",
					},
					Status: updatev1alpha1.NodeVersionStatus{
						Conditions: []metav1.Condition{
							{
								Message: "Some node versions are out of date",
							},
						},
					},
				}),
				attestation: &config.QEMUVTPM{
					Measurements: measurements.M{
						15: measurements.WithAllBytes(0, measurements.Enforce, measurements.PCRMeasurementLength),
					},
				},
			},
			expectedOutput: inProgressOutput,
		},
		"error getting node status": {
			kubeClient: stubKubeClient{
				statusErr: assert.AnError,
				version: mustParseNodeVersion(updatev1alpha1.NodeVersion{
					Spec: updatev1alpha1.NodeVersionSpec{
						ImageVersion:             "v1.1.0",
						ImageReference:           "v1.1.0",
						KubernetesClusterVersion: "v1.2.3",
					},
					Status: updatev1alpha1.NodeVersionStatus{
						Conditions: []metav1.Condition{
							{
								Message: "Node version of every node is up to date",
							},
						},
					},
				}),
				attestation: &config.QEMUVTPM{
					Measurements: measurements.M{
						15: measurements.WithAllBytes(0, measurements.Enforce, measurements.PCRMeasurementLength),
					},
				},
			},
			expectedOutput: successOutput,
			wantErr:        true,
		},
		"error getting node version": {
			kubeClient: stubKubeClient{
				status: map[string]kubecmd.NodeStatus{
					"outdated": kubecmd.NewNodeStatus(corev1.Node{
						ObjectMeta: metav1.ObjectMeta{
							Name: "node1",
							Annotations: map[string]string{
								"constellation.edgeless.systems/node-image": "v1.1.0",
							},
						},
						Status: corev1.NodeStatus{
							NodeInfo: corev1.NodeSystemInfo{
								KubeletVersion: "v1.2.3",
							},
						},
					}),
				},
				versionErr: assert.AnError,
				attestation: &config.QEMUVTPM{
					Measurements: measurements.M{
						15: measurements.WithAllBytes(0, measurements.Enforce, measurements.PCRMeasurementLength),
					},
				},
			},
			expectedOutput: successOutput,
			wantErr:        true,
		},
		"error getting attestation config": {
			kubeClient: stubKubeClient{
				status: map[string]kubecmd.NodeStatus{
					"outdated": kubecmd.NewNodeStatus(corev1.Node{
						ObjectMeta: metav1.ObjectMeta{
							Name: "node1",
							Annotations: map[string]string{
								"constellation.edgeless.systems/node-image": "v1.1.0",
							},
						},
						Status: corev1.NodeStatus{
							NodeInfo: corev1.NodeSystemInfo{
								KubeletVersion: "v1.2.3",
							},
						},
					}),
				},
				version: mustParseNodeVersion(updatev1alpha1.NodeVersion{
					Spec: updatev1alpha1.NodeVersionSpec{
						ImageVersion:             "v1.1.0",
						ImageReference:           "v1.1.0",
						KubernetesClusterVersion: "v1.2.3",
					},
					Status: updatev1alpha1.NodeVersionStatus{
						Conditions: []metav1.Condition{
							{
								Message: "Node version of every node is up to date",
							},
						},
					},
				}),
				attestationErr: assert.AnError,
			},
			expectedOutput: successOutput,
			wantErr:        true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			require := require.New(t)
			assert := assert.New(t)

			cmd := NewStatusCmd()
			var out bytes.Buffer
			cmd.SetOut(&out)
			var errOut bytes.Buffer
			cmd.SetErr(&errOut)

			fileHandler := file.NewHandler(afero.NewMemMapFs())
			cfg, err := createConfigWithAttestationVariant(cloudprovider.Azure, "", variant.AzureSEVSNP{})
			require.NoError(err)
			modifyConfigForAzureToPassValidate(cfg)
			require.NoError(fileHandler.WriteYAML(constants.ConfigFilename, cfg))
			s := statusCmd{fileHandler: fileHandler}

			err = s.status(
				cmd,
				stubGetVersions(versionsOutput),
				tc.kubeClient,
				stubAttestationFetcher{},
			)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.expectedOutput, out.String())
		})
	}
}

func modifyConfigForAzureToPassValidate(c *config.Config) {
	c.RemoveProviderAndAttestationExcept(cloudprovider.Azure)
	c.Image = constants.BinaryVersion().String()
	c.Provider.Azure.SubscriptionID = "11111111-1111-1111-1111-111111111111"
	c.Provider.Azure.TenantID = "11111111-1111-1111-1111-111111111111"
	c.Provider.Azure.Location = "westus"
	c.Provider.Azure.ResourceGroup = "test"
	c.Provider.Azure.UserAssignedIdentity = "/subscriptions/11111111-1111-1111-1111-111111111111/resourcegroups/constellation-identity/providers/Microsoft.ManagedIdentity/userAssignedIdentities/constellation-identity"
	c.Attestation.AzureSEVSNP.Measurements = measurements.M{
		0: measurements.WithAllBytes(0x00, measurements.Enforce, measurements.PCRMeasurementLength),
	}
	c.NodeGroups = map[string]config.NodeGroup{
		constants.ControlPlaneDefault: {
			Role:            "control-plane",
			Zone:            "",
			InstanceType:    "Standard_DC4as_v5",
			StateDiskSizeGB: 30,
			StateDiskType:   "StandardSSD_LRS",
			InitialCount:    3,
		},
		constants.WorkerDefault: {
			Role:            "worker",
			Zone:            "",
			InstanceType:    "Standard_DC4as_v5",
			StateDiskSizeGB: 30,
			StateDiskType:   "StandardSSD_LRS",
			InitialCount:    3,
		},
	}
}

type stubKubeClient struct {
	status         map[string]kubecmd.NodeStatus
	statusErr      error
	version        kubecmd.NodeVersion
	versionErr     error
	attestation    config.AttestationCfg
	attestationErr error
}

func (s stubKubeClient) ClusterStatus(_ context.Context) (map[string]kubecmd.NodeStatus, error) {
	return s.status, s.statusErr
}

func (s stubKubeClient) GetConstellationVersion(_ context.Context) (kubecmd.NodeVersion, error) {
	return s.version, s.versionErr
}

func (s stubKubeClient) GetClusterAttestationConfig(_ context.Context, _ variant.Variant) (config.AttestationCfg, error) {
	return s.attestation, s.attestationErr
}

func stubGetVersions(output string) func() (fmt.Stringer, error) {
	return func() (fmt.Stringer, error) {
		return stubServiceVersions{output}, nil
	}
}

type stubServiceVersions struct {
	output string
}

func (s stubServiceVersions) String() string {
	return s.output
}
