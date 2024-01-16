/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package setup

import (
	"context"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/kms/uri"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m,
		// https://github.com/census-instrumentation/opencensus-go/issues/1262
		goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start"),
		goleak.IgnoreAnyFunction("github.com/bazelbuild/rules_go/go/tools/bzltestutil.RegisterTimeoutHandler.func1"),
	)
}

func TestSetUpKMS(t *testing.T) {
	assert := assert.New(t)

	kms, err := KMS(context.Background(), "storage://unknown", "kms://unknown")
	assert.Error(err)
	assert.Nil(kms)

	masterSecret := uri.MasterSecret{Key: []byte("key"), Salt: []byte("salt")}
	kms, err = KMS(context.Background(), "storage://no-store", masterSecret.EncodeToURI())
	assert.NoError(err)
	assert.NotNil(kms)
}
