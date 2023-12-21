/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package controllers

import (
	"testing"

	"k8s.io/apimachinery/pkg/runtime"

	nodemaintenancev1beta1 "github.com/edgelesssys/constellation/v2/3rdparty/node-maintenance-operator/api/v1beta1"
	updatev1alpha1 "github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/api/v1alpha1"
	"github.com/stretchr/testify/require"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
)

func getScheme(t *testing.T) *runtime.Scheme {
	scheme := runtime.NewScheme()
	require.NoError(t, clientgoscheme.AddToScheme(scheme))
	require.NoError(t, updatev1alpha1.AddToScheme(scheme))
	require.NoError(t, nodemaintenancev1beta1.AddToScheme(scheme))
	return scheme
}
