//go:build linux && cgo

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: BUSL-1.1
*/
package cryptsetup

// #include <libcryptsetup.h>
import "C"

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/martinjungblut/go-cryptsetup"
	"golang.org/x/sys/unix"
)

const (
	// ReadWriteQueueBypass is a flag to disable the write and read workqueues for a crypt device.
	ReadWriteQueueBypass = C.CRYPT_ACTIVATE_NO_WRITE_WORKQUEUE | C.CRYPT_ACTIVATE_NO_READ_WORKQUEUE
	wipeFlags            = cryptsetup.CRYPT_ACTIVATE_PRIVATE | cryptsetup.CRYPT_ACTIVATE_NO_JOURNAL
	wipePattern          = cryptsetup.CRYPT_WIPE_ZERO
)

var errInvalidType = errors.New("device is not a *cryptsetup.Device")

func format(device cryptDevice, integrity bool) error {
	switch d := device.(type) {
	case cgoFormatter:
		luks2Params := cryptsetup.LUKS2{
			SectorSize: 4096,
			PBKDFType: &cryptsetup.PbkdfType{
				// Use low memory recommendation from https://datatracker.ietf.org/doc/html/rfc9106#section-7
				Type:            "argon2id",
				TimeMs:          2000,
				Iterations:      3,
				ParallelThreads: 4,
				MaxMemoryKb:     65536, // ~64MiB
			},
		}
		genericParams := cryptsetup.GenericParams{
			Cipher:        "aes",
			CipherMode:    "xts-plain64",
			VolumeKeySize: 64, // 32*2 bytes for aes-xts-plain64 encryption
		}

		if integrity {
			luks2Params.Integrity = "hmac(sha256)"
			genericParams.VolumeKeySize += 32 // 32 bytes for hmac(sha256) integrity
		}

		return d.Format(luks2Params, genericParams)
	default:
		return errInvalidType
	}
}

// headerRestore restores the header of the given device from the header in the given file.
// Reloading the device is required for the changes to be reflected in the active [cryptDevice] struct.
func headerRestore(device cryptDevice, headerFile string) error {
	switch d := device.(type) {
	case cgoRestorer:
		return d.HeaderRestore(cryptsetup.LUKS2{}, headerFile)
	default:
		return errInvalidType
	}
}

// headerBackup creates a backup of the cryptDevice's header to the given file.
func headerBackup(device cryptDevice, headerFile string) error {
	switch d := device.(type) {
	case cgoBackuper:
		if err := os.Remove(headerFile); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("removing existing header file %q: %w", headerFile, err)
		}
		if err := d.HeaderBackup(cryptsetup.LUKS2{}, headerFile); err != nil {
			return fmt.Errorf("creating header backup: %w", err)
		}
		return nil
	default:
		return errInvalidType
	}
}

func initByDevicePath(devicePath string) (deviceDetachedHeader, deviceAttachedHeader cryptDevice, headerDevice string, headerFile string, err error) {
	tmpDevice, err := cryptsetup.Init(devicePath)
	if err != nil {
		return nil, nil, "", "", fmt.Errorf("init device by path %s: %w", devicePath, err)
	}
	// If the device is not LUKS2 formatted, this is treated as a new device,
	// meaning no header exists yet
	if tmpDevice.Load(cryptsetup.LUKS2{}) != nil {
		return nil, tmpDevice, "", "", nil
	}
	defer tmpDevice.Free()

	deviceAttachedHeader, err = cryptsetup.Init(devicePath)
	if err != nil {
		return nil, nil, "", "", fmt.Errorf("init device by path %s: %w", devicePath, err)
	}
	defer func() {
		if err != nil && deviceAttachedHeader != nil {
			deviceAttachedHeader.Free()
		}
	}()

	headerDevice, headerFile, err = detachHeader(tmpDevice)
	if err != nil {
		return nil, nil, "", "", err
	}
	defer func() {
		if err != nil {
			_ = detachLoopbackDevice(headerDevice)
		}
	}()

	cryptDevice, err := cryptsetup.InitDataDevice(headerDevice, devicePath)
	return cryptDevice, deviceAttachedHeader, headerDevice, headerFile, err
}

func initByName(name string) (deviceDetachedHeader, deviceAttachedHeader cryptDevice, headerDevice string, headerFile string, err error) {
	tmpDevice, err := cryptsetup.InitByName(name)
	if err != nil {
		return nil, nil, "", "", fmt.Errorf("init device by name %s: %w", name, err)
	}
	// If the device is not LUKS2 formatted, this is treated as a new device,
	// meaning no header exists yet
	if tmpDevice.Load(cryptsetup.LUKS2{}) != nil {
		return nil, tmpDevice, "", "", nil
	}
	defer tmpDevice.Free()

	deviceAttachedHeader, err = cryptsetup.InitByName(name)
	if err != nil {
		return nil, nil, "", "", fmt.Errorf("init device by name %s: %w", name, err)
	}
	defer func() {
		if err != nil && deviceAttachedHeader != nil {
			deviceAttachedHeader.Free()
		}
	}()

	headerDevice, headerFile, err = detachHeader(tmpDevice)
	if err != nil {
		return nil, nil, "", "", err
	}
	defer func() {
		if err != nil {
			_ = detachLoopbackDevice(headerDevice)
		}
	}()

	cryptDevice, err := cryptsetup.InitByNameAndHeader(name, headerDevice)
	return cryptDevice, deviceAttachedHeader, headerDevice, headerFile, err
}

func loadLUKS2(device cryptDevice) error {
	switch d := device.(type) {
	case cgoLoader:
		return d.Load(cryptsetup.LUKS2{})
	default:
		return errInvalidType
	}
}

// detachHeader loads reads the header from the given cryptsetup device and returns a loopback device with just the header.
func detachHeader(device *cryptsetup.Device) (headerDevice, headerFile string, err error) {
	headerFile = filepath.Join(os.TempDir(), fmt.Sprintf("luks-header-%s", uuid.New().String()))
	if err = headerBackup(device, headerFile); err != nil {
		return "", "", err
	}

	headerDevice, err = createLoopbackDevice(headerFile)
	if err != nil {
		return "", "", fmt.Errorf("create loopback device: %w", err)
	}
	defer func() {
		if err != nil {
			_ = detachLoopbackDevice(headerDevice)
		}
	}()

	headerCryptDevice, err := cryptsetup.Init(headerDevice)
	if err != nil {
		return "", "", fmt.Errorf("init header device: %w", err)
	}
	defer headerCryptDevice.Free()
	if err := headerCryptDevice.Load(cryptsetup.LUKS2{}); err != nil {
		return "", "", fmt.Errorf("creating header backup: %w", err)
	}
	metadataJSON, err := headerCryptDevice.DumpJSON()
	if err != nil {
		return "", "", fmt.Errorf("dumping device metadata: %w", err)
	}

	var metadata cryptsetupMetadata
	decoder := json.NewDecoder(strings.NewReader(metadataJSON))
	decoder.DisallowUnknownFields() // Ensure no unknown fields are present in the JSON data
	if err := decoder.Decode(&metadata); err != nil {
		return "", "", fmt.Errorf("decoding LUKS header JSON from %s: %w", headerFile, err)
	}

	if err := verifyLUKS2Header(metadata); err != nil {
		return "", "", fmt.Errorf("verifying LUKS2 header: %w", err)
	}

	return headerDevice, headerFile, nil
}

// verifyLUKS2Header verifies a LUKS2 header contains the expected configuration for Constellation.
func verifyLUKS2Header(metadata cryptsetupMetadata) error {
	if len(metadata.KeySlots) == 0 {
		return errors.New("no key slots found in LUKS2 header")
	}
	for slotName, slot := range metadata.KeySlots {
		if slot.Type != "luks2" {
			return fmt.Errorf("unsupported key slot type %q for slot %q", slot.Type, slotName)
		}
		if slot.KeySize != 64 && slot.KeySize != 96 { // 64 for encryption, 96 if integrity is added
			return fmt.Errorf("unsupported key size %d for slot %q", slot.KeySize, slotName)
		}
		if slot.AntiForensicSplitter.Type != "luks1" {
			return fmt.Errorf("unsupported anti-forensic splitter type %q for slot %q", slot.AntiForensicSplitter.Type, slotName)
		}
		if slot.AntiForensicSplitter.Stripes != 4000 {
			return fmt.Errorf("unsupported anti-forensic splitter stripes %d for slot %q", slot.AntiForensicSplitter.Stripes, slotName)
		}
		if slot.AntiForensicSplitter.Hash != "sha256" {
			return fmt.Errorf("unsupported anti-forensic splitter hash %q for slot %q", slot.AntiForensicSplitter.Hash, slotName)
		}
		if slot.Area.Type != "raw" {
			return fmt.Errorf("unsupported area type %q for slot %q", slot.Area.Type, slotName)
		}
		if slot.Area.Encryption != "aes-xts-plain64" {
			return fmt.Errorf("unsupported area encryption %q for slot %q", slot.Area.Encryption, slotName)
		}
		if slot.Area.KeySize != 64 {
			return fmt.Errorf("unsupported area key size %d for slot %q", slot.Area.KeySize, slotName)
		}
		if slot.KDF.Type != "argon2id" {
			return fmt.Errorf("unsupported KDF type %q for slot %q", slot.KDF.Type, slotName)
		}
		if slot.KDF.Memory == 0 {
			return fmt.Errorf("unsupported KDF memory %d for slot %q", slot.KDF.Memory, slotName)
		}
		if slot.KDF.Salt == "" {
			return fmt.Errorf("unsupported KDF salt for slot %q", slotName)
		}
	}
	if len(metadata.Segments) == 0 {
		return errors.New("no segments found in LUKS2 header")
	}
	for segmentName, segment := range metadata.Segments {
		if segment.Type != "crypt" {
			return fmt.Errorf("unsupported segment type %q for segment %q", segment.Type, segmentName)
		}
		if segment.SectorSize != 4096 {
			return fmt.Errorf("unsupported segment sector size %d for segment %q", segment.SectorSize, segmentName)
		}
		if segment.IVTweak != "0" {
			return fmt.Errorf("unsupported segment IV tweak %q for segment %q", segment.IVTweak, segmentName)
		}
		if segment.Encryption != "aes-xts-plain64" {
			return fmt.Errorf("unsupported segment encryption %q for segment %q", segment.Encryption, segmentName)
		}
		switch segment.Integrity.Type {
		case "hmac(sha256)":
			if segment.Integrity.JournalEncryption != "none" {
				return fmt.Errorf("unsupported segment integrity journal encryption %q for segment %q", segment.Integrity.JournalEncryption, segmentName)
			}
			if segment.Integrity.JournalIntegrity != "none" {
				return fmt.Errorf("unsupported segment integrity journal integrity %q for segment %q", segment.Integrity.JournalIntegrity, segmentName)
			}
		case "":
			if segment.Integrity.JournalEncryption != "" {
				return fmt.Errorf("unsupported segment integrity journal encryption %q for segment %q", segment.Integrity.JournalEncryption, segmentName)
			}
			if segment.Integrity.JournalIntegrity != "" {
				return fmt.Errorf("unsupported segment integrity journal integrity %q for segment %q", segment.Integrity.JournalIntegrity, segmentName)
			}
		default:
			return fmt.Errorf("unsupported segment integrity type %q for segment %q", segment.Integrity.Type, segmentName)
		}
	}
	if len(metadata.Digests) == 0 {
		return errors.New("no digests found in LUKS2 header")
	}
	for digestName, digest := range metadata.Digests {
		if digest.Type != "pbkdf2" {
			return fmt.Errorf("unsupported digest type %q for digest %q", digest.Type, digestName)
		}
		if digest.Hash != "sha256" {
			return fmt.Errorf("unsupported digest hash %q for digest %q", digest.Hash, digestName)
		}
		if digest.Salt == "" {
			return fmt.Errorf("unsupported digest salt for digest %q", digestName)
		}
		if digest.Digest == "" {
			return fmt.Errorf("unsupported digest value for digest %q", digestName)
		}
	}
	return nil
}

// createLoopbackDevice sets up a loop device for the given file and returns the loop device path (e.g., /dev/loop0).
func createLoopbackDevice(filePath string) (string, error) {
	file, err := os.OpenFile(filePath, os.O_RDWR, 0)
	if err != nil {
		return "", fmt.Errorf("open backing file: %w", err)
	}
	defer file.Close()

	// Get a free loop device number
	ctrl, err := os.OpenFile("/dev/loop-control", os.O_RDWR, 0)
	if err != nil {
		return "", fmt.Errorf("open /dev/loop-control: %w", err)
	}
	defer ctrl.Close()
	loopNum, _, errno := unix.Syscall(unix.SYS_IOCTL, ctrl.Fd(), unix.LOOP_CTL_GET_FREE, 0)
	if errno != 0 {
		return "", fmt.Errorf("LOOP_CTL_GET_FREE: %v", errno)
	}

	// Open the loop device
	loopDev := fmt.Sprintf("/dev/loop%d", loopNum)
	loop, err := os.OpenFile(loopDev, os.O_RDWR, 0)
	if err != nil {
		return "", fmt.Errorf("open loop device: %w", err)
	}
	defer loop.Close()

	// Associate the file with the loop device
	if _, _, errno := unix.Syscall(unix.SYS_IOCTL, loop.Fd(), unix.LOOP_SET_FD, file.Fd()); errno != 0 {
		return "", fmt.Errorf("LOOP_SET_FD: %v", errno)
	}

	return loopDev, nil
}

// detachLoopbackDevice removes the specified loopback device.
func detachLoopbackDevice(loopDev string) error {
	loop, err := os.OpenFile(loopDev, os.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("open loop device: %w", err)
	}
	defer loop.Close()

	if _, _, errno := unix.Syscall(unix.SYS_IOCTL, loop.Fd(), unix.LOOP_CLR_FD, 0); errno != 0 {
		return fmt.Errorf("LOOP_CLR_FD: %v", errno)
	}
	return nil
}

type cryptsetupMetadata struct {
	KeySlots map[string]struct {
		Type                 string `json:"type"`
		KeySize              int    `json:"key_size"`
		AntiForensicSplitter struct {
			Type    string `json:"type"`
			Stripes int    `json:"stripes"`
			Hash    string `json:"hash"`
		} `json:"af"`
		Area struct {
			Type       string `json:"type"`
			Offset     string `json:"offset"`
			Size       string `json:"size"`
			Encryption string `json:"encryption"`
			KeySize    int    `json:"key_size"`
		} `json:"area"`
		KDF struct {
			Type   string `json:"type"`
			Time   int    `json:"time"`
			Memory int    `json:"memory"`
			CPUs   int    `json:"cpus"`
			Salt   string `json:"salt"`
		} `json:"kdf"`
	} `json:"keyslots"`
	Tokens   map[string]any `json:"tokens"`
	Segments map[string]struct {
		Type       string   `json:"type"`
		Offset     string   `json:"offset"`
		Size       string   `json:"size"`
		Flags      []string `json:"flags,omitempty"`
		IVTweak    string   `json:"iv_tweak"`
		Encryption string   `json:"encryption"`
		SectorSize int      `json:"sector_size"`
		Integrity  struct {
			Type              string `json:"type"`
			JournalEncryption string `json:"journal_encryption"`
			JournalIntegrity  string `json:"journal_integrity"`
		}
	} `json:"segments"`
	Digests map[string]struct {
		Type       string   `json:"type"`
		Keyslots   []string `json:"keyslots"`
		Segments   []string `json:"segments"`
		Hash       string   `json:"hash"`
		Iterations int      `json:"iterations"`
		Salt       string   `json:"salt"`
		Digest     string   `json:"digest"`
	} `json:"digests"`
	Config struct {
		JSONSize     string `json:"json_size"`
		KeyslotsSize string `json:"keyslots_size"`
	}
}

type cgoFormatter interface {
	Format(deviceType cryptsetup.DeviceType, genericParams cryptsetup.GenericParams) error
}

type cgoLoader interface {
	Load(deviceType cryptsetup.DeviceType) error
}

type cgoRestorer interface {
	HeaderRestore(deviceType cryptsetup.DeviceType, headerFile string) error
}
type cgoBackuper interface {
	HeaderBackup(deviceType cryptsetup.DeviceType, headerFile string) error
}
