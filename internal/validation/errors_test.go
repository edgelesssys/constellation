package validation

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewValidationErrorSingleField(t *testing.T) {
	st := &ErrorTestDoc{
		ExportedField: "abc",
		OtherField:    42,
	}

	err := NewValidationError(st, &st.OtherField, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "validating otherField: <nil>")
}

type ErrorTestDoc struct {
	ExportedField string `json:"exportedField" yaml:"exportedField"`
	OtherField    int    `json:"otherField" yaml:"otherField"`
}
