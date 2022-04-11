package vtpm

import (
	"io"

	"github.com/google/go-tpm-tools/simulator"
	"github.com/google/go-tpm/tpm2"
)

const (
	// tpmPath is the path to the vTPM.
	tpmPath = "/dev/tpmrm0"
)

// TPMOpenFunc opens a TPM device.
type TPMOpenFunc func() (io.ReadWriteCloser, error)

// OpenVTPM opens the vTPM at `TPMPath`.
func OpenVTPM() (io.ReadWriteCloser, error) {
	return tpm2.OpenTPM(tpmPath)
}

// OpenSimulatedTPM returns a simulated TPM device.
func OpenSimulatedTPM() (io.ReadWriteCloser, error) {
	return simulator.Get()
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

type simulatedTPM struct {
	openSimulatedTPM io.ReadWriteCloser
}

// NewSimulatedTPMOpenFunc returns a TPMOpenFunc that opens a simulated TPM.
func NewSimulatedTPMOpenFunc() (TPMOpenFunc, io.Closer) {
	tpm, err := OpenSimulatedTPM()
	if err != nil {
		panic(err)
	}
	return func() (io.ReadWriteCloser, error) {
		return &simulatedTPM{tpm}, nil
	}, tpm
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
