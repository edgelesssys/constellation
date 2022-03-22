package client

import (
	"errors"

	"github.com/aws/smithy-go"
)

// checkDryRunError error checks if an error is a DryRun error.
// If the error is nil, an error is returned, since a DryRun error
// is the expected result of a DryRun operation.
func checkDryRunError(err error) error {
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) && apiErr.ErrorCode() == "DryRunOperation" {
		return nil
	}
	if err != nil {
		return err
	}
	return errors.New("expected APIError: DryRunOperation, but got no error at all")
}
