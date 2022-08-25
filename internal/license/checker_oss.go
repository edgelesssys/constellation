//go:build !enterprise

package license

import (
	"context"

	"github.com/edgelesssys/constellation/internal/file"
)

type Checker struct{}

func NewChecker(quotaChecker QuotaChecker, fileHandler file.Handler) *Checker {
	return &Checker{}
}

// CheckLicense is a no-op for open source version of Constellation.
func (c *Checker) CheckLicense(ctx context.Context, printer func(string, ...any)) error {
	return nil
}
