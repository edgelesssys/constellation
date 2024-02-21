/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestAskToConfirm(t *testing.T) {
	// errAborted is an error where the user aborted the action.
	errAborted := errors.New("user aborted")

	cmd := &cobra.Command{
		Use:  "test",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ok, err := askToConfirm(cmd, "777")
			if err != nil {
				return err
			}
			if !ok {
				return errAborted
			}
			return nil
		},
	}

	testCases := map[string]struct {
		input   string
		wantErr error
	}{
		"user confirms":                       {"y\n", nil},
		"user confirms long":                  {"yes\n", nil},
		"user disagrees":                      {"n\n", errAborted},
		"user disagrees long":                 {"no\n", errAborted},
		"user is first unsure, but agrees":    {"what?\ny\n", nil},
		"user is first unsure, but disagrees": {"wait.\nn\n", errAborted},
		"repeated invalid input":              {"h\nb\nq\n", ErrInvalidInput},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			out := &bytes.Buffer{}
			cmd.SetOut(out)
			cmd.SetErr(&bytes.Buffer{})
			in := bytes.NewBufferString(tc.input)
			cmd.SetIn(in)
			cmd.SetArgs([]string{})

			err := cmd.Execute()
			assert.ErrorIs(err, tc.wantErr)

			output, err := io.ReadAll(out)
			assert.NoError(err)
			assert.Contains(string(output), "777")
		})
	}
}
