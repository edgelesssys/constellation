package vtpm

import (
	"errors"
	"io"
	"testing"

	"github.com/google/go-tpm-tools/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// simTPMNOPCloser is a wrapper for the generic TPM simulator with a NOP Close() method.
type simTPMNOPCloser struct {
	io.ReadWriteCloser
}

func (s simTPMNOPCloser) Close() error {
	return nil
}

func TestMarkNodeAsInitialized(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	tpm, err := OpenSimulatedTPM()
	require.NoError(err)
	defer tpm.Close()
	pcrs, err := client.ReadAllPCRs(tpm)
	require.NoError(err)

	assert.NoError(MarkNodeAsInitialized(func() (io.ReadWriteCloser, error) {
		return &simTPMNOPCloser{tpm}, nil
	}, []byte{0x0, 0x1, 0x2, 0x3}, []byte{0x4, 0x5, 0x6, 0x7}))

	pcrsInitialized, err := client.ReadAllPCRs(tpm)
	require.NoError(err)

	for i := range pcrs {
		assert.NotEqual(pcrs[i].Pcrs[uint32(PCRIndexOwnerID)], pcrsInitialized[i].Pcrs[uint32(PCRIndexOwnerID)])
		assert.NotEqual(pcrs[i].Pcrs[uint32(PCRIndexClusterID)], pcrsInitialized[i].Pcrs[uint32(PCRIndexClusterID)])
	}
}

func TestFailOpener(t *testing.T) {
	assert := assert.New(t)

	assert.Error(MarkNodeAsInitialized(func() (io.ReadWriteCloser, error) { return nil, errors.New("failed") }, []byte{0x0, 0x1, 0x2, 0x3}, []byte{0x0, 0x1, 0x2, 0x3}))
}
