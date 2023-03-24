/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"context"
	"testing"

	"github.com/edgelesssys/constellation/v2/cli/internal/helm"
	updatev1alpha1 "github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/v2/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

const successOutput = `Target versions:
	Image: v1.1.0
	Kubernetes: v1.2.3
Installed service versions:
	Cilium: v1.0.0
	cert-manager: v1.0.0
	constellation-operators: v1.1.0
	constellation-services: v1.1.0
Cluster status: Node version of every node is up to date
`

const inProgressOutput = `Target versions:
	Image: v1.1.0
	Kubernetes: v1.2.3
Installed service versions:
	Cilium: v1.0.0
	cert-manager: v1.0.0
	constellation-operators: v1.1.0
	constellation-services: v1.1.0
Cluster status: Some node versions are out of date
	Image: 1/2
	Kubernetes: 1/2
`

// TestStatus checks that the status function produces the correct strings.
func TestStatus(t *testing.T) {
	testCases := map[string]struct {
		kubeClient     stubKubeClient
		helmClient     stubHelmClient
		nodeVersion    updatev1alpha1.NodeVersion
		dynamicErr     error
		expectedOutput string
		wantErr        bool
	}{
		"success": {
			kubeClient: stubKubeClient{
				nodes: []corev1.Node{
					{
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
					},
				},
			},
			helmClient: stubHelmClient{
				serviceVersions: helm.NewServiceVersions("v1.0.0", "v1.0.0", "v1.1.0", "v1.1.0"),
			},
			nodeVersion: updatev1alpha1.NodeVersion{
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
			},
			expectedOutput: successOutput,
		},
		"one of two nodes not upgraded": {
			kubeClient: stubKubeClient{
				nodes: []corev1.Node{
					{
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
					},
					{
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
					},
				},
			},
			helmClient: stubHelmClient{
				serviceVersions: helm.NewServiceVersions("v1.0.0", "v1.0.0", "v1.1.0", "v1.1.0"),
			},
			nodeVersion: updatev1alpha1.NodeVersion{
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
			},
			expectedOutput: inProgressOutput,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			require := require.New(t)
			assert := assert.New(t)

			raw, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&tc.nodeVersion)
			require.NoError(err)
			output, err := status(context.Background(), tc.kubeClient, tc.helmClient, &stubDynamicInterface{data: unstructured.Unstructured{Object: raw}, err: tc.dynamicErr})
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.expectedOutput, output)
		})
	}
}

type stubKubeClient struct {
	nodes []corev1.Node
	err   error
}

func (s stubKubeClient) GetNodes(_ context.Context) ([]corev1.Node, error) {
	return s.nodes, s.err
}

type stubHelmClient struct {
	serviceVersions helm.ServiceVersions
	err             error
}

func (s stubHelmClient) Versions() (helm.ServiceVersions, error) {
	return s.serviceVersions, s.err
}

type stubDynamicInterface struct {
	data unstructured.Unstructured
	err  error
}

func (s *stubDynamicInterface) GetCurrent(_ context.Context, _ string) (*unstructured.Unstructured, error) {
	return &s.data, s.err
}

func (s *stubDynamicInterface) Update(_ context.Context, _ *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	return &s.data, s.err
}
