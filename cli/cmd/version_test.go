package cmd

import (
	"bytes"
	"io"
	"testing"

	"github.com/edgelesssys/constellation/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestVersionCmd(t *testing.T) {
	assert := assert.New(t)

	cmd := newVersionCmd()
	b := bytes.NewBufferString("")
	cmd.SetOut(b)

	err := cmd.Execute()
	assert.NoError(err)

	s, err := io.ReadAll(b)
	assert.NoError(err)
	assert.Contains(string(s), config.Version)
}
