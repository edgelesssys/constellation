//go:build disable_tpm_simulator
// +build disable_tpm_simulator

package simulator

import (
	"io"
	"testing"

	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

// OpenSimulatedTPM returns a simulated TPM device.
func OpenSimulatedTPM() (io.ReadWriteCloser, error) {
	panic("simulator not enabled")
}

// NewSimulatedTPMOpenFunc returns a TPMOpenFunc that opens a simulated TPM.
func NewSimulatedTPMOpenFunc() (func() (io.ReadWriteCloser, error), io.Closer) {
	panic("simulator not enabled")
}
