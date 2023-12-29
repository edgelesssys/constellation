/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"context"
	"log/slog"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMAAPatch(t *testing.T) {
	testCases := map[string]struct {
		attestationURL string
		patcher        *stubPolicyPatcher
		wantErr        bool
	}{
		"success": {
			attestationURL: "https://example.com",
			patcher:        &stubPolicyPatcher{},
			wantErr:        false,
		},
		"patch error": {
			attestationURL: "https://example.com",
			patcher:        &stubPolicyPatcher{patchErr: assert.AnError},
			wantErr:        true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			require := require.New(t)

      c := &maaPatchCmd{log: slog.New(slog.NewTextHandler(logger.TestWriter{T: t}, nil)), patcher: tc.patcher}
			err := c.patchMAA(&cobra.Command{}, tc.attestationURL)
			if tc.wantErr {
				require.Error(err)
			} else {
				require.NoError(err)
			}
		})
	}
}

type stubPolicyPatcher struct {
	patchErr error
}

// Patch implements the patcher interface.
func (p *stubPolicyPatcher) Patch(context.Context, string) error {
	return p.patchErr
}
