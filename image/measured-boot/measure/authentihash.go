/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package measure

import (
	"bytes"
	"crypto"
	"fmt"
	"io"

	"github.com/foxboron/go-uefi/authenticode"
)

// Authentihash returns the PE/COFF hash / Authentihash of a file.
func Authentihash(r io.Reader, h crypto.Hash) ([]byte, error) {
	readerAt, err := getReaderAt(r)
	if err != nil {
		return nil, fmt.Errorf("failed to get readerAt: %v", err)
	}

	bin, err := authenticode.Parse(readerAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pe file: %v", err)
	}
	return bin.Hash(h), nil
}

func getReaderAt(r io.Reader) (io.ReaderAt, error) {
	if ra, ok := r.(io.ReaderAt); ok {
		return ra, nil
	}

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(r); err != nil {
		return nil, fmt.Errorf("failed to read pe file: %v", err)
	}
	return bytes.NewReader(buf.Bytes()), nil
}
