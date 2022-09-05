/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package gcp

import (
	"fmt"

	"github.com/spf13/afero"
)

// Writer implements ConfigWriter.
type Writer struct {
	fs afero.Afero
}

// WriteGCEConf persists the GCE config on disk.
func (w *Writer) WriteGCEConf(config string) error {
	if err := w.fs.WriteFile("/etc/gce.conf", []byte(config), 0o644); err != nil {
		return fmt.Errorf("writing gce config: %w", err)
	}
	return nil
}
