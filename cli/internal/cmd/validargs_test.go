/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: BUSL-1.1
*/

package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestIsCloudProvider(t *testing.T) {
	testCases := map[string]struct {
		pos     int
		args    []string
		wantErr bool
	}{
		"gcp":     {0, []string{"gcp"}, false},
		"azure":   {1, []string{"foo", "azure"}, false},
		"foo":     {0, []string{"foo"}, true},
		"empty":   {0, []string{""}, true},
		"unknown": {0, []string{"unknown"}, true},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			testCmd := &cobra.Command{Args: isCloudProvider(tc.pos)}

			err := testCmd.ValidateArgs(tc.args)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}
