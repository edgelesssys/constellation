/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package measure

import (
	"crypto/sha256"
	"fmt"
	"io"

	"github.com/edgelesssys/constellation/v2/image/measured-boot/pesection"
)

// DescribeUKISections describes the expected measurements for the UKI sections.
func DescribeUKISections(w io.Writer, ukiSections []pesection.PESection) error {
	if _, err := fmt.Fprintf(w, "UKI sections:\n"); err != nil {
		return err
	}

	var maxNameLen int
	for _, ukiSection := range ukiSections {
		if len(ukiSection.Name) > maxNameLen {
			maxNameLen = len(ukiSection.Name)
		}
	}
	for i, ukiSection := range ukiSections {
		if ukiSection.Measure {
			if _, err := fmt.Fprintf(w, "  Section %2d - %-*s (%10d bytes):\t%x, %x\n", i+1, maxNameLen, ukiSection.Name, ukiSection.Size, sha256.Sum256(ukiSection.NullTerminatedName()), ukiSection.Digest); err != nil {
				return err
			}
			continue
		}
		if _, err := fmt.Fprintf(w, "  Section %2d - %-*s:\t%s\n", i+1, maxNameLen, ukiSection.Name, "not measured"); err != nil {
			return err
		}
	}
	return nil
}

// PredictPCR11 predicts the PCR11 value based on the components of unified kernel images.
func PredictPCR11(simulator *Simulator, ukiSections []pesection.PESection) error {
	for i, ukiSection := range ukiSections {
		// systemd-stub documentation TPM PCR Notes
		// https://github.com/systemd/systemd/blob/7c52d5236a3bc85db1755de6a458934be095cd1c/src/boot/efi/stub.c#L409-L441

		if !ukiSection.Measure {
			continue
		}

		// first, measure the name
		name := ukiSection.NullTerminatedName()
		err := simulator.ExtendPCR(11, sha256.Sum256(name), name, fmt.Sprintf("EV_IPL: UKI section %d name: %s", i+1, ukiSection.Name))
		if err != nil {
			return err
		}

		// then, measure the data
		err = simulator.ExtendPCR(11, ukiSection.Digest, nil, fmt.Sprintf("EV_IPL: UKI section %d data: %x", i+1, ukiSection.Digest))
		if err != nil {
			return err
		}
	}

	return nil
}
