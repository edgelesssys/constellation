package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"strings"

	"github.com/edgelesssys/constellation/coordinator/attestation/vtpm"
	"github.com/spf13/cobra"
)

var (
	// ErrInvalidInput is an error where user entered invalid input.
	ErrInvalidInput = errors.New("user made invalid input")
	warningStr      = "Warning: not verifying the Constellation's %s measurements\n"
)

// askToConfirm asks user to confirm an action.
// The user will be asked the handed question and can answer with
// yes or no.
func askToConfirm(cmd *cobra.Command, question string) (bool, error) {
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

// warnAboutPCRs displays warnings if specifc PCR values are not verified.
//
// PCR allocation inspired by https://link.springer.com/chapter/10.1007/978-1-4302-6584-9_12#Tab1
func warnAboutPCRs(cmd *cobra.Command, pcrs map[uint32][]byte, checkInit bool) error {
	for k, v := range pcrs {
		if len(v) != 32 {
			return fmt.Errorf("bad config: PCR[%d]: expected length: %d, but got: %d", k, 32, len(v))
		}
	}

	if pcrs[0] == nil || pcrs[1] == nil {
		cmd.PrintErrf(warningStr, "BIOS")
	}

	if pcrs[2] == nil || pcrs[3] == nil {
		cmd.PrintErrf(warningStr, "OPROM")
	}

	if pcrs[4] == nil || pcrs[5] == nil {
		cmd.PrintErrf(warningStr, "MBR")
	}

	// GRUB measures kernel command line and initrd into pcrs 8 and 9
	if pcrs[8] == nil {
		cmd.PrintErrf(warningStr, "kernel command line")
	}
	if pcrs[9] == nil {
		cmd.PrintErrf(warningStr, "initrd")
	}

	// Only warn about initialization PCRs if necessary
	if checkInit {
		if pcrs[uint32(vtpm.PCRIndexOwnerID)] == nil || pcrs[uint32(vtpm.PCRIndexClusterID)] == nil {
			cmd.PrintErrf(warningStr, "initialization status")
		}
	}

	return nil
}
