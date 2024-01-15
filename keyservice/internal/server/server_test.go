/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package server

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/kms/kms"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/keyservice/keyserviceproto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m, goleak.IgnoreAnyFunction("github.com/bazelbuild/rules_go/go/tools/bzltestutil.RegisterTimeoutHandler.func1"))
}

func TestGetDataKey(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	log := slog.New(slog.NewTextHandler(logger.TestWriter{T: t}, nil))

	kms := &stubKMS{derivedKey: []byte{0x0, 0x1, 0x2, 0x3, 0x4, 0x5}}
	api := New(log, kms)

	res, err := api.GetDataKey(context.Background(), &keyserviceproto.GetDataKeyRequest{DataKeyId: "1", Length: 32})
	require.NoError(err)
	assert.Equal(kms.derivedKey, res.DataKey)

	// Test no data key id
	res, err = api.GetDataKey(context.Background(), &keyserviceproto.GetDataKeyRequest{Length: 32})
	require.Error(err)
	assert.Nil(res)

	// Test no / zero key length
	res, err = api.GetDataKey(context.Background(), &keyserviceproto.GetDataKeyRequest{DataKeyId: "1"})
	require.Error(err)
	assert.Nil(res)

	// Test derive key error
	api = New(log, &stubKMS{deriveKeyErr: errors.New("error")})
	res, err = api.GetDataKey(context.Background(), &keyserviceproto.GetDataKeyRequest{DataKeyId: "1", Length: 32})
	assert.Error(err)
	assert.Nil(res)
}

type stubKMS struct {
	kms.CloudKMS
	masterKey    []byte
	derivedKey   []byte
	deriveKeyErr error
}

func (c *stubKMS) CreateKEK(_ context.Context, _ string, kek []byte) error {
	c.masterKey = kek
	return nil
}

func (c *stubKMS) GetDEK(_ context.Context, _ string, _ int) ([]byte, error) {
	if c.deriveKeyErr != nil {
		return nil, c.deriveKeyErr
	}
	return c.derivedKey, nil
}
