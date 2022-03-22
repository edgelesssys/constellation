package client

import (
	"errors"
	"testing"

	"github.com/aws/smithy-go"
	"github.com/stretchr/testify/assert"
)

func TestCheckDryRunError(t *testing.T) {
	assert := assert.New(t)

	someErr := errors.New("failed")
	assert.ErrorIs(checkDryRunError(someErr), someErr)

	dryRunErr := smithy.GenericAPIError{Code: "DryRunOperation"}
	assert.NoError(checkDryRunError(&dryRunErr))

	var nilErr error
	assert.Error(checkDryRunError(nilErr))
}
