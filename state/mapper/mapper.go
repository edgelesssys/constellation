package mapper

import (
	"errors"
	"fmt"
	"path/filepath"

	cryptsetup "github.com/martinjungblut/go-cryptsetup"
)

const (
	gcpStateDiskPath   = "/dev/disk/by-id/google-state-disk"
	azureStateDiskPath = "/dev/disk/azure/scsi1/lun0"
	fallBackPath       = "/dev/disk/by-id/state-disk"
)

// Mapper handles actions for formating and mapping crypt devices.
type Mapper struct {
	device cryptDevice
}

// New creates a new crypt device for the device at path.
func New(path string) (*Mapper, error) {
	device, err := cryptsetup.Init(path)
	if err != nil {
		return nil, fmt.Errorf("initializing crypt device for disk %q: %w", path, err)
	}
	return &Mapper{device: device}, nil
}

// Close closes and frees memory allocated for the crypt device.
func (m *Mapper) Close() error {
	if m.device.Free() {
		return nil
	}
	return errors.New("unable to close crypt device")
}

// IsLUKSDevice returns true if the device is formatted as a LUKS device.
func (m *Mapper) IsLUKSDevice() bool {
	return m.device.Load(cryptsetup.LUKS2{}) == nil
}

// GetUUID gets the device's UUID.
func (m *Mapper) DiskUUID() string {
	return m.device.GetUUID()
}

// FormatDisk formats the disk and adds passphrase in keyslot 0.
func (m *Mapper) FormatDisk(passphrase string) error {
	luksParams := cryptsetup.LUKS2{
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
		VolumeKeySize: 64,
	}

	if err := m.device.Format(luksParams, genericParams); err != nil {
		return fmt.Errorf("formating disk: %w", err)
	}

	if err := m.device.KeyslotAddByVolumeKey(0, "", passphrase); err != nil {
		return fmt.Errorf("adding keyslot: %w", err)
	}

	return nil
}

// MapDisk maps a crypt device to /dev/mapper/target using the provided passphrase.
func (m *Mapper) MapDisk(target, passphrase string) error {
	if err := m.device.ActivateByPassphrase(target, 0, passphrase, 0); err != nil {
		return fmt.Errorf("mapping disk as %q: %w", target, err)
	}
	return nil
}

// UnmapDisk removes the mapping of target.
func (m *Mapper) UnmapDisk(target string) error {
	return m.device.Deactivate(target)
}

// GetDiskPath returns the device path of the data disk by cloud provider.
//
// For GCP a symlink to the disk is expected at /dev/disk/by-id/google-state-disk
// For Azure a symlink to the disk is expected at /dev/disk/azure/scsi1/lun0
// If no symlink can be found at the given path, or if no known cloud provider is supplied,
// we instead return the device path of the os-disk stateful partition at /dev/disk/by-partlabel/stateful.
func GetDiskPath(csp string) (string, error) {
	var diskPath string
	var err error

	switch csp {
	case "gcp":
		diskPath, err = filepath.EvalSymlinks(gcpStateDiskPath)
	case "azure":
		diskPath, err = filepath.EvalSymlinks(azureStateDiskPath)
	default:
		diskPath = fallBackPath
	}

	if err != nil {
		return filepath.EvalSymlinks(fallBackPath)
	}
	return diskPath, nil
}
