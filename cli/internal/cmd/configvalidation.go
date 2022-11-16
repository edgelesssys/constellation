/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"errors"
	"fmt"
	"io"

	"go.uber.org/multierr"
)

func displayConfigValidationErrors(errWriter io.Writer, configError error) error {
	errs := multierr.Errors(configError)
	if errs != nil {
		fmt.Fprintln(errWriter, "Problems validating config file:")
		for _, err := range errs {
			fmt.Fprintln(errWriter, "\t"+err.Error())
		}
		fmt.Fprintln(errWriter, "Fix the invalid entries or generate a new configuration using `constellation config generate`")
		return errors.New("invalid configuration")
	}
	return nil
}
