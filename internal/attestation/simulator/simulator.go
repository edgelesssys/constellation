//go:build !disable_tpm_simulator

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// TPM2 simulator used for unit tests.
package simulator

import (
	"io"

	"github.com/google/go-tpm-tools/simulator"
)

// OpenSimulatedTPM returns a simulated TPM device.
func OpenSimulatedTPM() (io.ReadWriteCloser, error) {
	return simulator.Get()
}

// NewSimulatedTPMOpenFunc returns a TPMOpenFunc that opens a simulated TPM.
func NewSimulatedTPMOpenFunc() (func() (io.ReadWriteCloser, error), io.Closer) {
	tpm, err := OpenSimulatedTPM()
	if err != nil {
		panic(err)
	}
	return func() (io.ReadWriteCloser, error) {
		return &simulatedTPM{tpm}, nil
	}, tpm
}

type simulatedTPM struct {
	openSimulatedTPM io.ReadWriteCloser
}

func (t *simulatedTPM) Read(p []byte) (int, error) {
	return t.openSimulatedTPM.Read(p)
}

func (t *simulatedTPM) Write(p []byte) (int, error) {
	return t.openSimulatedTPM.Write(p)
}

func (t *simulatedTPM) Close() error {
	// never close the underlying simulated TPM to allow calling the TPMOpenFunc again
	return nil
}

func (*simulatedTPM) EventLog() ([]byte, error) {
	return nil, nil
}
