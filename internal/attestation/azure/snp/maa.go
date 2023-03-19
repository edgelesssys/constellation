/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package snp

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/edgelesssys/go-azguestattestation/maa"
)

type maaClient struct {
	client *http.Client
}

func newMAAClient() *maaClient {
	return &maaClient{
		client: &http.Client{Transport: &http.Transport{Proxy: nil}},
	}
}

func (m *maaClient) newParameters(ctx context.Context, nonce []byte, tpmHandle io.ReadWriter) (maa.Parameters, error) {
	return maa.NewParameters(ctx, nonce, m.client, tpmHandle)
}

func (m *maaClient) createToken(
	ctx context.Context, tpm io.ReadWriter, maaURL string, data []byte, params maa.Parameters,
) (string, error) {
	tokenEnc, err := maa.GetEncryptedToken(ctx, params, data, maaURL, m.client)
	if err != nil {
		return "", fmt.Errorf("getting encrypted token: %w", err)
	}

	token, err := maa.DecryptToken(tokenEnc, tpm)
	if err != nil {
		return "", fmt.Errorf("decrypting token: %w", err)
	}

	return token, nil
}

func (m *maaClient) validateToken(ctx context.Context, maaURL, token string, extraData []byte) error {
	keySet, err := maa.GetKeySet(ctx, maaURL, m.client)
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
		return errors.New("invalid claims: missing x-ms-runtime")
	}
	payload, ok := runtime["client-payload"].(map[string]interface{})
	if !ok {
		return errors.New("invalid claims: missing client-payload")
	}
	nonce, ok := payload["nonce"].(string)
	if !ok {
		return errors.New("invalid claims: missing nonce")
	}

	if nonce != base64.StdEncoding.EncodeToString(extraData) {
		return errors.New("invalid claims: nonce does not match extra data")
	}
	return nil
}
