/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/edgelesssys/constellation/v2/image/measured-boot/extract"
	"github.com/edgelesssys/constellation/v2/image/measured-boot/measure"
	"github.com/edgelesssys/constellation/v2/image/measured-boot/pesection"
	"github.com/spf13/afero"
)

const (
	ukiPath = "/efi/EFI/BOOT/BOOTX64.EFI"
)

func precalculatePCRs(fs afero.Fs, dissectToolchain, imageFile string) (*measure.Simulator, error) {
	dir, err := afero.TempDir(fs, "", "con-measure")
	if err != nil {
		return nil, err
	}
	defer func() { _ = fs.RemoveAll(dir) }()

	simulator := measure.NewDefaultSimulator()

	// extract UKI from raw image
	ukiFile := filepath.Join(dir, "uki.efi")
	if err := extract.CopyFrom(dissectToolchain, imageFile, ukiPath, ukiFile); err != nil {
		return nil, fmt.Errorf("failed to extract UKI: %v", err)
	}

	// extract section digests from UKI
	ukiReader, err := fs.Open(ukiFile)
	if err != nil {
		return nil, err
	}
	defer ukiReader.Close()

	ukiSections, err := extract.PeFileSectionDigests(ukiReader)
	if err != nil {
		return nil, fmt.Errorf("failed to extract UKI section digests: %v", err)
	}

	if err := precalculatePCR4(simulator, fs, ukiFile); err != nil {
		return nil, err
	}

	if err := precalculatePCR9(simulator, fs, ukiFile); err != nil {
		return nil, err
	}

	if err := precalculatePCR11(simulator, ukiSections); err != nil {
		return nil, err
	}

	fmt.Fprintf(os.Stderr, "PCR[ 4]: %x\n", simulator.Bank[4])
	fmt.Fprintf(os.Stderr, "PCR[ 9]: %x\n", simulator.Bank[9])
	fmt.Fprintf(os.Stderr, "PCR[11]: %x\n", simulator.Bank[11])
	// TODO(malt3): with systemd-stub >= 254, PCR[12] will
	// contain the "rendered" kernel command line,
	// credentials, and sysexts. We should measure these
	// values here.
	// For now, we expect the PCR to be zero.
	fmt.Fprintf(os.Stderr, "PCR[12]: %x\n", simulator.Bank[12])
	// PCR[13] would contain extension images for the initrd
	// We enforce the absence of extension images by
	// expecting PCR[13] to be zero.
	fmt.Fprintf(os.Stderr, "PCR[13]: %x\n", simulator.Bank[13])
	// PCR[15] can be used to measure from userspace (systemd-pcrphase and others)
	// We enforce the absence of userspace measurements by
	// expecting PCR[15] to be zero at boot.
	fmt.Fprintf(os.Stderr, "PCR[15]: %x\n", simulator.Bank[15])

	return simulator, nil
}

func measurePE(fs afero.Fs, peFile string) ([]byte, error) {
	f, err := fs.Open(peFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return measure.Authentihash(f, sha256.New())
}

func precalculatePCR4(simulator *measure.Simulator, fs afero.Fs, ukiFile string) error {
	ukiMeasurement, err := measurePE(fs, ukiFile)
	if err != nil {
		return fmt.Errorf("failed to measure UKI: %v", err)
	}

	ukiPe, err := fs.Open(ukiFile)
	if err != nil {
		return err
	}
	defer ukiPe.Close()
	linuxSectionReader, err := extract.PeSectionReader(ukiPe, ".linux")
	if err != nil {
		return fmt.Errorf("uki does not contain linux kernel image: %v", err)
	}
	linuxMeasurement, err := measure.Authentihash(linuxSectionReader, sha256.New())
	if err != nil {
		return fmt.Errorf("failed to measure linux kernel image: %v", err)
	}

	bootStages := []measure.EFIBootStage{
		{Name: "Unified Kernel Image (UKI)", Digest: measure.PCR256(ukiMeasurement)},
		{Name: "Linux", Digest: measure.PCR256(linuxMeasurement)},
	}

	if err := measure.DescribeBootStages(os.Stderr, bootStages); err != nil {
		return err
	}

	return measure.PredictPCR4(simulator, bootStages)
}

func precalculatePCR9(simulator *measure.Simulator, fs afero.Fs, ukiFile string) error {
	// load cmdline and initrd from UKI

	ukiPe, err := fs.Open(ukiFile)
	if err != nil {
		return err
	}
	defer ukiPe.Close()

	cmdlineSectionReader, err := extract.PeSectionReader(ukiPe, ".cmdline")
	if err != nil {
		return fmt.Errorf("uki does not contain cmdline: %v", err)
	}

	cmdline := new(bytes.Buffer)
	if _, err := cmdline.ReadFrom(cmdlineSectionReader); err != nil {
		return err
	}

	initrdSectionReader, err := extract.PeSectionReader(ukiPe, ".initrd")
	if err != nil {
		return fmt.Errorf("uki does not contain initrd: %v", err)
	}

	initrdDigest := sha256.New()
	if _, err := io.Copy(initrdDigest, initrdSectionReader); err != nil {
		return err
	}

	cmdlineBytes := cmdline.Bytes()
	initrdDigestBytes := [32]byte(initrdDigest.Sum(nil))

	if err := measure.DescribeLinuxLoad2(os.Stderr, cmdlineBytes, initrdDigestBytes); err != nil {
		return err
	}

	return measure.PredictPCR9(simulator, cmdlineBytes, initrdDigestBytes)
}

func precalculatePCR11(simulator *measure.Simulator, ukiSections []pesection.PESection) error {
	if err := measure.DescribeUKISections(os.Stderr, ukiSections); err != nil {
		return err
	}

	return measure.PredictPCR11(simulator, ukiSections)
}

func loadToolchain(key, fallback string) string {
	toolchain := os.Getenv(key)
	if toolchain == "" {
		toolchain = fallback
	}
	toolchain, err := exec.LookPath(toolchain)
	if err != nil {
		return ""
	}

	absolutePath, err := filepath.Abs(toolchain)
	if err != nil {
		return ""
	}
	return absolutePath
}

func writeOutput(fs afero.Fs, outputFile string, simulator *measure.Simulator) error {
	out, err := fs.Create(outputFile)
	if err != nil {
		return err
	}
	defer out.Close()

	return json.NewEncoder(out).Encode(simulator)
}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "Usage: measured-boot-precalc <image-file> <output-file>")
		os.Exit(1)
	}

	imageFile := os.Args[1]
	outputFile := os.Args[2]

	fs := afero.NewOsFs()
	dissectToolchain := loadToolchain("DISSECT_TOOLCHAIN", "systemd-dissect")

	simulator, err := precalculatePCRs(fs, dissectToolchain, imageFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := writeOutput(fs, outputFile, simulator); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
