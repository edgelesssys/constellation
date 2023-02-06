/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package snp

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/edgelesssys/go-azguestattestation/maa"
)

type maaClient struct {
	url    string // URL of the MAA service. Potentially configured at runtime.
	client *http.Client
}

func (m *maaClient) createToken(ctx context.Context, data []byte) (string, error) {
	return maa.Attest(ctx, data, m.url, m.client)
}

func (m *maaClient) validateToken(ctx context.Context, token string, extraData []byte) error {
	keySet, err := maa.GetKeySet(ctx, m.url, m.client)
	if err != nil {
		return fmt.Errorf("getting key set from MAA: %w", err)
	}
	claims, err := maa.ValidateToken(token, keySet)
	if err != nil {
		return fmt.Errorf("validating token: %w", err)
	}
	return m.validateClaims(claims, extraData)
}

func (m *maaClient) validateClaims(claims map[string]interface{}, extraData []byte) error {
	runtime, ok := claims["x-ms-runtime"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid claims: missing x-ms-runtime")
	}
	payload, ok := runtime["client-payload"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid claims: missing client-payload")
	}
	nonce, ok := payload["nonce"].(string)
	if !ok {
		return fmt.Errorf("invalid claims: missing nonce")
	}

	if nonce != base64.StdEncoding.EncodeToString(extraData) {
		return fmt.Errorf("invalid claims: nonce does not match extra data")
	}
	return nil
}
