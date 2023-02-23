package main

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/edgelesssys/go-tdx-qpl/tdx"
)

func main() {
	tdxDevice, err := os.Open(tdx.GuestDevice)
	if err != nil {
		fmt.Println("error: cannot open TDX device:", err)
		os.Exit(1)
	}

	measurements, err := tdx.ReadMeasurements(tdxDevice)
	if err != nil {
		fmt.Println("error: cannot read TDX measurements", err)
		os.Exit(1)
	}

	// The intent is to make the output look the same as form tpm2_pcrread
	// %8s = space padding (2 spaces for sha384, 4 spaces for MRTD/RTMR[i])
	// %3s = alignment for "MRTD" to have the same length as "RTMR[i]"
	fmt.Printf("%8s:\n", "sha384")
	fmt.Printf("%8s%3s: 0x%s\n", "MRTD", "", hex.EncodeToString(measurements[0][:]))
	for i := 0; i <= 3; i++ {
		fmt.Printf("%8s[%d]: 0x%s\n", "RTMR", i, hex.EncodeToString(measurements[i+1][:]))
	}
}
