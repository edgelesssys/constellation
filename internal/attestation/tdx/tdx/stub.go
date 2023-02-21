// Package tdx contains stubs until we merge open PRs in the TDX QPL repo.
package tdx

import (
	"io"
	"os"
)

// GuestDevice is the path to the TDX guest device.
const GuestDevice = "/dev/tdx-guest"

// Device is a handle to the TDX guest device.
type Device interface {
	io.ReadWriteCloser
	Fd() uintptr
}

// Open opens the TDX guest device and returns a handle to it.
func Open() (Device, error) {
	return &os.File{}, nil
}

// IsTDXDevice checks if the given device is a TDX guest device.
func IsTDXDevice(device Device) bool {
	return false
}

// GenerateQuote generates a TDX quote for the given user data.
func GenerateQuote(tdx Device, userData []byte) ([]byte, error) {
	return nil, nil
}

// ExtendRTMR extends the RTMR with the given data.
func ExtendRTMR(tdx Device, extendData []byte, index uint8) error {
	return nil
}

// ReadRTMRs reads the RTMR values of a TDX guest.
func ReadRTMRs(tdx Device) ([4][48]byte, error) {
	return [4][48]byte{}, nil
}
