/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package extract

import (
	"crypto/sha256"
	"debug/pe"
	"fmt"
	"io"
	"os/exec"
	"sort"

	"github.com/edgelesssys/constellation/v2/image/measured-boot/pesection"
)

// CopyFrom is a wrapper for systemd-dissect --copy-from.
func CopyFrom(dissectToolchain, image, path, output string) error {
	if dissectToolchain == "" {
		dissectToolchain = "systemd-dissect"
	}
	out, err := exec.Command(dissectToolchain, "--copy-from", image, path, output).CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to extract %s from %s: %v\n%s", path, image, err, out)
	}
	return nil
}

// PeSectionReader returns a reader for the named section of a PE file.
func PeSectionReader(peFile io.ReaderAt, section string) (io.Reader, error) {
	f, err := pe.NewFile(peFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	for _, s := range f.Sections {
		if s.Name == section {
			return io.LimitReader(s.Open(), int64(s.VirtualSize)), nil
		}
	}

	return nil, fmt.Errorf("section %q not found", section)
}

// PeFileSectionDigests returns the section digests of a PE file.
func PeFileSectionDigests(peFile io.ReaderAt) ([]pesection.PESection, error) {
	f, err := pe.NewFile(peFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	sections := make([]pesection.PESection, len(f.Sections))
	for i, section := range f.Sections {
		sectionDigest := sha256.New()
		sectionReader := section.Open()
		_, err := io.CopyN(sectionDigest, sectionReader, int64(section.VirtualSize))
		if err != nil {
			return nil, err
		}

		sections[i].Name = section.Name
		sections[i].Size = section.VirtualSize
		sections[i].Digest = ([32]byte)(sectionDigest.Sum(nil))
		sections[i].Measure = shouldMeasureSection(section.Name)
		sections[i].MeasureOrder = sectionMeasureOrder(section.Name)
	}

	sort.Slice(sections, func(i, j int) bool {
		if sections[i].Measure != sections[j].Measure {
			return sections[i].Measure
		}
		if sections[i].MeasureOrder == sections[j].MeasureOrder {
			return sections[i].Name < sections[j].Name
		}
		return sections[i].MeasureOrder < sections[j].MeasureOrder
	})

	return sections, nil
}

var ukiSections = []string{
	".linux",
	".osrel",
	".cmdline",
	".initrd",
	".splash",
	".dtb",
	// uanme and sbat will be added in systemd-stub >= 254
	// ".uname",
	// ".sbat",
	".pcrsig",
	".pcrkey",
}

func shouldMeasureSection(name string) bool {
	for _, section := range ukiSections {
		if name == section && name != ".pcrsig" {
			return true
		}
	}
	return false
}

func sectionMeasureOrder(name string) int {
	for i, section := range ukiSections {
		if name == section {
			return i
		}
	}
	return -1
}
