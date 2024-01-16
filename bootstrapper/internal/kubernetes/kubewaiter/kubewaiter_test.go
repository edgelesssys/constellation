/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package kubewaiter

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
	corev1 "k8s.io/api/core/v1"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m, goleak.IgnoreAnyFunction("github.com/bazelbuild/rules_go/go/tools/bzltestutil.RegisterTimeoutHandler.func1"))
}

func TestCloudKubeAPIWaiter(t *testing.T) {
	testCases := map[string]struct {
		kubeClient KubernetesClient
		wantErr    bool
	}{
		"success": {
			kubeClient: &stubKubernetesClient{},
		},
		"error": {
			kubeClient: &stubKubernetesClient{listAllNamespacesErr: errors.New("error")},
			wantErr:    true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			require := require.New(t)

			waiter := &CloudKubeAPIWaiter{}
			ctx, cancel := context.WithTimeout(context.Background(), 0)
			defer cancel()
			err := waiter.Wait(ctx, tc.kubeClient)
			if tc.wantErr {
				require.Error(err)
			} else {
				require.NoError(err)
			}
		})
	}
}

type stubKubernetesClient struct {
	listAllNamespacesErr error
}

func (c *stubKubernetesClient) ListAllNamespaces(_ context.Context) (*corev1.NamespaceList, error) {
	return nil, c.listAllNamespacesErr
}
