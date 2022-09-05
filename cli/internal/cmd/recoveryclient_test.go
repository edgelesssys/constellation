/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"context"

	"github.com/edgelesssys/constellation/internal/atls"
)

type stubRecoveryClient struct {
	conn                bool
	connectErr          error
	closeErr            error
	pushStateDiskKeyErr error

	pushStateDiskKeyKey []byte
}

func (c *stubRecoveryClient) Connect(_ string, _ atls.Validator) error {
	c.conn = true
	return c.connectErr
}

func (c *stubRecoveryClient) Close() error {
	c.conn = false
	return c.closeErr
}

func (c *stubRecoveryClient) PushStateDiskKey(_ context.Context, stateDiskKey, _ []byte) error {
	c.pushStateDiskKeyKey = stateDiskKey
	return c.pushStateDiskKeyErr
}
