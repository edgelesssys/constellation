/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package pesection

// PESection describes a PE section.
type PESection struct {
	Name         string
	Size         uint32
	Digest       [32]byte
	Measure      bool
	MeasureOrder int
}

// NullTerminatedName returns the name of the section with a null terminator.
func (u PESection) NullTerminatedName() []byte {
	if len(u.Name) > 0 && u.Name[len(u.Name)-1] == 0x00 {
		return []byte(u.Name)
	}
	return append([]byte(u.Name), 0x00)
}
