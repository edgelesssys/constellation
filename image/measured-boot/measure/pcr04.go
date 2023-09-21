/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package measure

import (
	"fmt"
	"io"
)

// EFIBootStage is a stage (bootloader) of the EFI boot process.
type EFIBootStage struct {
	Name   string
	Digest [32]byte
}

// DescribeBootStages prints a description of the EFIBootStages to a writer.
func DescribeBootStages(w io.Writer, bootStages []EFIBootStage) error {
	if _, err := fmt.Fprintf(w, "EFI Boot Stages:\n"); err != nil {
		return err
	}
	var maxNameLen int
	for _, bootStage := range bootStages {
		if len(bootStage.Name) > maxNameLen {
			maxNameLen = len(bootStage.Name)
		}
	}
	for i, bootStage := range bootStages {
		if _, err := fmt.Fprintf(w, "  Stage %d - %-*s:\t%x\n", i+1, maxNameLen, bootStage.Name, bootStage.Digest); err != nil {
			return err
		}
	}
	return nil
}

// PredictPCR4 predicts the PCR4 value based on the EFIBootStages.
func PredictPCR4(simulator *Simulator, efiBootStages []EFIBootStage) error {
	// TCG PC Client Platform Firmware Profile Family "2.0 Section" 7.2.4.4.a
	if err := simulator.ExtendPCR(4, EVEFIActionPCR256(), nil, "EV_EFI_ACTION: Calling EFI Application from Boot Option"); err != nil {
		return err
	}
	// TCG PC Client Platform Firmware Profile Family "2.0 Section" 7.2.4.4.b
	if err := simulator.ExtendPCR(4, EVSeparatorPCR256(), []byte{0x00, 0x00, 0x00, 0x00}, "EV_SEPARATOR"); err != nil {
		return err
	}

	for i, efiBootStage := range efiBootStages {
		// TCG PC Client Platform Firmware Profile Family "2.0 Section" 7.2.4.4.e
		err := simulator.ExtendPCR(4, efiBootStage.Digest, nil, fmt.Sprintf("Boot Stage %d: %s", i+1, efiBootStage.Name))
		if err != nil {
			return err
		}
	}

	return nil
}
