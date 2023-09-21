/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package measure

import (
	"bytes"
	"fmt"
	"hash"
	"io"

	"github.com/foxboron/go-uefi/efi/pecoff"
)

// Authentihash returns the PE/COFF hash / Authentihash of a file.
func Authentihash(r io.Reader, h hash.Hash) ([]byte, error) {
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(r); err != nil {
		return nil, fmt.Errorf("failed to read pe file: %v", err)
	}

	signingCtx := pecoff.PECOFFChecksum(buf.Bytes())
	pecoff.PaddSigCtx(signingCtx)

	h.Write(signingCtx.SigData.Bytes())

	return h.Sum(nil), nil
}
