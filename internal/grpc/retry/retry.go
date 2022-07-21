package retry

import (
	"errors"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ServiceIsUnavailable checks if the error is a grpc status with code Unavailable.
// In the special case of an authentication handshake failure, false is returned to prevent further retries.
func ServiceIsUnavailable(err error) bool {
	// taken from google.golang.org/grpc/status.FromError
	var targetErr interface {
		GRPCStatus() *status.Status
		Error() string
	}

	if !errors.As(err, &targetErr) {
		return false
	}

	statusErr, ok := status.FromError(targetErr)
	if !ok {
		return false
	}

	if statusErr.Code() != codes.Unavailable {
		return false
	}

	// ideally we would check the error type directly, but grpc only provides a string
	return !strings.HasPrefix(statusErr.Message(), `connection error: desc = "transport: authentication handshake failed`)
}
