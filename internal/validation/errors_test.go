package validation

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewValidationErrorSingleField(t *testing.T) {
	st := &ErrorTestDoc{
		ExportedField: "abc",
		OtherField:    42,
	}

	err := NewValidationError(st, &st.OtherField, assert.AnError)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("validating otherField: %s", assert.AnError))
}

type ErrorTestDoc struct {
	ExportedField string `json:"exportedField" yaml:"exportedField"`
	OtherField    int    `json:"otherField" yaml:"otherField"`
}
