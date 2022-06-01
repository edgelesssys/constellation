//go:build disable_tpm_simulator
// +build disable_tpm_simulator

package simulator

import (
	"io"
)

// OpenSimulatedTPM returns a simulated TPM device.
func OpenSimulatedTPM() (io.ReadWriteCloser, error) {
	panic("simulator not enabled")
}

// NewSimulatedTPMOpenFunc returns a TPMOpenFunc that opens a simulated TPM.
func NewSimulatedTPMOpenFunc() (func() (io.ReadWriteCloser, error), io.Closer) {
	panic("simulator not enabled")
}
