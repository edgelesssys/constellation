/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/
package sigstore

import (
	"bytes"
	"encoding/base64"

	"github.com/sigstore/cosign/v2/pkg/cosign"
)

// SignContent signs the content with the cosign encrypted private key and corresponding cosign password.
func SignContent(password, encryptedPrivateKey, content []byte) ([]byte, error) {
	sv, err := cosign.LoadPrivateKey(encryptedPrivateKey, password)
	if err != nil {
		return nil, err
	}
	sig, err := sv.SignMessage(bytes.NewReader(content))
    if err != nil {
        return nil, fmt.Errorf("signing message: %w", err)
    }
	return []byte(base64.StdEncoding.EncodeToString(sig)), nil
}
