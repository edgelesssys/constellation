/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"bufio"
	"errors"
	"strings"

	"github.com/spf13/cobra"
)

// ErrInvalidInput is an error where user entered invalid input.
var ErrInvalidInput = errors.New("user made invalid input")

// AskToConfirm asks user to confirm an action.
// The user will be asked the handed question and can answer with
// yes or no.
func AskToConfirm(cmd *cobra.Command, question string) (bool, error) {
	reader := bufio.NewReader(cmd.InOrStdin())
	cmd.Printf("%s [y/n]: ", question)
	for i := 0; i < 3; i++ {
		resp, err := reader.ReadString('\n')
		if err != nil {
			return false, err
		}
		resp = strings.ToLower(strings.TrimSpace(resp))
		if resp == "n" || resp == "no" {
			return false, nil
		}
		if resp == "y" || resp == "yes" {
			return true, nil
		}
		cmd.Printf("Type 'y' or 'yes' to confirm, or abort action with 'n' or 'no': ")
	}
	return false, ErrInvalidInput
}
