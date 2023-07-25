/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"context"
	"fmt"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
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
Service versions:
	Cilium: v1.0.0
	cert-manager: v1.0.0
	constellation-operators: v1.1.0
	constellation-services: v1.1.0
Cluster status: Node version of every node is up to date
` + attestationConfigOutput

const inProgressOutput = `Target versions:
	Image: v1.1.0
	Kubernetes: v1.2.3
Service versions:
	Cilium: v1.0.0
	cert-manager: v1.0.0
	constellation-operators: v1.1.0
	constellation-services: v1.1.0
Cluster status: Some node versions are out of date
	Image: 1/2
	Kubernetes: 1/2
` + attestationConfigOutput

const attestationConfigOutput = `Attestation config:
	measurements:
	    0:
	        expected: 737f767a12f54e70eecbc8684011323ae2fe2dd9f90785577969d7a2013e8c12
	        warnOnly: true
	    2:
	        expected: 3d458cfe55cc03ea1f443f1562beec8df51c75e14a9fcf9a7234a13f198e7969
	        warnOnly: true
	    3:
	        expected: 3d458cfe55cc03ea1f443f1562beec8df51c75e14a9fcf9a7234a13f198e7969
	        warnOnly: true
	    4:
	        expected: 55f7616b2c51dd7603f491c1c266373fe5c1e25e06a851d2090960172b03b27f
	        warnOnly: false
	    6:
	        expected: 3d458cfe55cc03ea1f443f1562beec8df51c75e14a9fcf9a7234a13f198e7969
	        warnOnly: true
	    7:
	        expected: fb71e5e55cefba9e2b396d17604de0fe6e1841a76758856a120833e3ad1c40a3
	        warnOnly: true
	    8:
	        expected: "0000000000000000000000000000000000000000000000000000000000000000"
	        warnOnly: false
	    9:
	        expected: f7480d37929bef4b61c32823cb7b3771aea19f7510db2e1478719a1d88f9775d
	        warnOnly: false
	    11:
	        expected: "0000000000000000000000000000000000000000000000000000000000000000"
	        warnOnly: false
	    12:
	        expected: b8038d11eade4cfee5fd41da04bf64e58bab15c42bfe01801e4c0f61376ba010
	        warnOnly: false
	    13:
	        expected: "0000000000000000000000000000000000000000000000000000000000000000"
	        warnOnly: false
	    14:
	        expected: d7c4cc7ff7933022f013e03bdee875b91720b5b86cf1753cad830f95e791926f
	        warnOnly: true
	    15:
	        expected: "0000000000000000000000000000000000000000000000000000000000000000"
	        warnOnly: false
`

// TestStatus checks that the status function produces the correct strings.
func TestStatus(t *testing.T) {
	testCases := map[string]struct {
		kubeClient     stubKubeClient
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
			configMapper := stubConfigMapperAWSNitro{}
			variant := variant.AWSNitroTPM{}
			output, err := status(
				context.Background(),
				tc.kubeClient,
				configMapper,
				stubGetVersions(successOutput),
				&stubDynamicInterface{data: unstructured.Unstructured{Object: raw}, err: tc.dynamicErr},
				variant,
			)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.expectedOutput, output)
		})
	}
}

type stubConfigMapperAWSNitro struct{}

func (s stubConfigMapperAWSNitro) GetCurrentConfigMap(_ context.Context, _ string) (*corev1.ConfigMap, error) {
	return &corev1.ConfigMap{
		Data: map[string]string{
			"attestationConfig": `{"measurements":{"0":{"expected":"737f767a12f54e70eecbc8684011323ae2fe2dd9f90785577969d7a2013e8c12","warnOnly":true},"11":{"expected":"0000000000000000000000000000000000000000000000000000000000000000","warnOnly":false},"12":{"expected":"b8038d11eade4cfee5fd41da04bf64e58bab15c42bfe01801e4c0f61376ba010","warnOnly":false},"13":{"expected":"0000000000000000000000000000000000000000000000000000000000000000","warnOnly":false},"14":{"expected":"d7c4cc7ff7933022f013e03bdee875b91720b5b86cf1753cad830f95e791926f","warnOnly":true},"15":{"expected":"0000000000000000000000000000000000000000000000000000000000000000","warnOnly":false},"2":{"expected":"3d458cfe55cc03ea1f443f1562beec8df51c75e14a9fcf9a7234a13f198e7969","warnOnly":true},"3":{"expected":"3d458cfe55cc03ea1f443f1562beec8df51c75e14a9fcf9a7234a13f198e7969","warnOnly":true},"4":{"expected":"55f7616b2c51dd7603f491c1c266373fe5c1e25e06a851d2090960172b03b27f","warnOnly":false},"6":{"expected":"3d458cfe55cc03ea1f443f1562beec8df51c75e14a9fcf9a7234a13f198e7969","warnOnly":true},"7":{"expected":"fb71e5e55cefba9e2b396d17604de0fe6e1841a76758856a120833e3ad1c40a3","warnOnly":true},"8":{"expected":"0000000000000000000000000000000000000000000000000000000000000000","warnOnly":false},"9":{"expected":"f7480d37929bef4b61c32823cb7b3771aea19f7510db2e1478719a1d88f9775d","warnOnly":false}}}`,
		},
	}, nil
}

type stubKubeClient struct {
	nodes []corev1.Node
	err   error
}

func (s stubKubeClient) GetNodes(_ context.Context) ([]corev1.Node, error) {
	return s.nodes, s.err
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
