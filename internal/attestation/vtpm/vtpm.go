/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package vtpm

import (
	"io"

	"github.com/google/go-tpm/tpm2"
)

// TPMOpenFunc opens a TPM device.
type TPMOpenFunc func() (io.ReadWriteCloser, error)

// OpenVTPM opens the vTPM at `TPMPath`.
func OpenVTPM() (io.ReadWriteCloser, error) {
	return tpm2.OpenTPM()
}

type nopTPM struct{}

// OpenNOPTPM returns a NOP io.ReadWriteCloser that can be used as a TPM.
func OpenNOPTPM() (io.ReadWriteCloser, error) {
	return &nopTPM{}, nil
}

func (t nopTPM) Read(p []byte) (int, error) {
	return len(p), nil
}

func (t nopTPM) Write(p []byte) (int, error) {
	return len(p), nil
}

func (t nopTPM) Close() error {
	return nil
}
