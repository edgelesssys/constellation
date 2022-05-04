package cryptmapper

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	cryptsetup "github.com/martinjungblut/go-cryptsetup"
	"k8s.io/klog/v2"
	mount "k8s.io/mount-utils"
	utilexec "k8s.io/utils/exec"
)

const (
	cryptPrefix       = "/dev/mapper/"
	integritySuffix   = "_dif"
	integrityFSSuffix = "-integrity"
	keySizeIntegrity  = 96
	keySizeCrypt      = 64
)

// packageLock is needed to block concurrent use of package functions, since libcryptsetup is not thread safe.
// See: https://gitlab.com/cryptsetup/cryptsetup/-/issues/710
// 		https://stackoverflow.com/questions/30553386/cryptsetup-backend-safe-with-multithreading
var packageLock = sync.Mutex{}

// CryptMapper manages dm-crypt volumes.
type CryptMapper struct {
	mapper DeviceMapper
	kms    KeyCreator
}

// New initializes a new CryptMapper with the given kms client and key-encryption-key ID.
// kms is used to fetch data encryption keys for the dm-crypt volumes.
func New(kms KeyCreator, mapper DeviceMapper) *CryptMapper {
	return &CryptMapper{
		mapper: mapper,
		kms:    kms,
	}
}

// KeyCreator is an interface to create data encryption keys.
type KeyCreator interface {
	GetDEK(ctx context.Context, dekID string, dekSize int) ([]byte, error)
}

// DeviceMapper is an interface for device mapper methods.
type DeviceMapper interface {
	// Init initializes a crypt device backed by 'devicePath'.
	// Sets the deviceMapper to the newly allocated Device or returns any error encountered.
	// C equivalent: crypt_init
	Init(devicePath string) error
	// ActivateByVolumeKey activates a device by using a volume key.
	// Returns nil on success, or an error otherwise.
	// C equivalent: crypt_activate_by_volume_key
	ActivateByVolumeKey(deviceName, volumeKey string, volumeKeySize, flags int) error
	// Deactivate deactivates a device.
	// Returns nil on success, or an error otherwise.
	// C equivalent: crypt_deactivate
	Deactivate(deviceName string) error
	// Format formats a Device, using a specific device type, and type-independent parameters.
	// Returns nil on success, or an error otherwise.
	// C equivalent: crypt_format
	Format(deviceType cryptsetup.DeviceType, genericParams cryptsetup.GenericParams) error
	// Free releases crypt device context and used memory.
	// C equivalent: crypt_free
	Free() bool
	// Load loads crypt device parameters from the on-disk header.
	// Returns nil on success, or an error otherwise.
	// C equivalent: crypt_load
	Load(cryptsetup.DeviceType) error
	// Wipe removes existing data and clears the device for use with dm-integrity.
	// Returns nil on success, or an error otherwise.
	// C equivalent: crypt_wipe
	Wipe(devicePath string, pattern int, offset, length uint64, wipeBlockSize int, flags int, progress func(size, offset uint64) int) error
}

// cryptDevice is a wrapper for cryptsetup.Device.
type CryptDevice struct {
	*cryptsetup.Device
}

// Init initializes a crypt device backed by 'devicePath'.
// Sets the cryptDevice's deviceMapper to the newly allocated Device or returns any error encountered.
// C equivalent: crypt_init.
func (c *CryptDevice) Init(devicePath string) error {
	device, err := cryptsetup.Init(devicePath)
	if err != nil {
		return err
	}
	c.Device = device
	return nil
}

// Free releases crypt device context and used memory.
// C equivalent: crypt_free.
func (c *CryptDevice) Free() bool {
	res := c.Device.Free()
	c.Device = nil
	return res
}

// CloseCryptDevice closes the crypt device mapped for volumeID.
// Returns nil if the volume does not exist.
func (c *CryptMapper) CloseCryptDevice(volumeID string) error {
	source, err := filepath.EvalSymlinks(cryptPrefix + volumeID)
	if err != nil {
		var pathErr *fs.PathError
		if errors.As(err, &pathErr) {
			klog.V(4).Infof("Skipping unmapping for disk %q: volume does not exist or is already unmapped", volumeID)
			return nil
		}
		return fmt.Errorf("failed to get device path for disk %q: %w", cryptPrefix+volumeID, err)
	}
	if err := closeCryptDevice(c.mapper, source, volumeID, "crypt"); err != nil {
		return fmt.Errorf("failed to close crypt device: %w", err)
	}

	integrity, err := filepath.EvalSymlinks(cryptPrefix + volumeID + integritySuffix)
	if err == nil {
		// If device was created with integrity, we need to also close the integrity device
		integrityErr := closeCryptDevice(c.mapper, integrity, volumeID+integritySuffix, "integrity")
		if integrityErr != nil {
			klog.Errorf("Could not close integrity device: %s", integrityErr)
			return integrityErr
		}
	}
	if err != nil {
		var pathErr *fs.PathError
		if errors.As(err, &pathErr) {
			// integrity device does not exist
			return nil
		}
		return fmt.Errorf("failed to get device path for disk %q: %w", cryptPrefix+volumeID, err)
	}

	return nil
}

// closeCryptDevice closes the crypt device mapped for volumeID.
func closeCryptDevice(device DeviceMapper, source, volumeID, deviceType string) error {
	packageLock.Lock()
	defer packageLock.Unlock()

	klog.V(4).Infof("Unmapping dm-%s volume %q for device %q", deviceType, cryptPrefix+volumeID, source)
	cryptsetup.SetLogCallback(func(level int, message string) { klog.V(4).Infof("libcryptsetup: %s", message) })

	if err := device.Init(source); err != nil {
		klog.Errorf("Could not initialize dm-%s to unmap device %q: %s", deviceType, source, err)
		return fmt.Errorf("could not initialize dm-%s to unmap device %q: %w", deviceType, source, err)
	}
	defer device.Free()

	if err := device.Deactivate(volumeID); err != nil {
		klog.Errorf("Could not deactivate dm-%s volume %q for device %q: %s", deviceType, cryptPrefix+volumeID, source, err)
		return fmt.Errorf("could not deactivate dm-%s volume %q for device %q: %w", deviceType, cryptPrefix+volumeID, source, err)
	}

	klog.V(4).Infof("Successfully unmapped dm-%s volume %q for device %q", deviceType, cryptPrefix+volumeID, source)
	return nil
}

// OpenCryptDevice maps the volume at source to the crypt device identified by volumeID.
// The key used to encrypt the volume is fetched using CryptMapper's kms client.
func (c *CryptMapper) OpenCryptDevice(ctx context.Context, source, volumeID string, integrity bool) (string, error) {
	klog.V(4).Infof("Fetching data encryption key for volume %q", volumeID)

	keySize := keySizeCrypt
	if integrity {
		keySize = keySizeIntegrity
	}
	dek, err := c.kms.GetDEK(ctx, volumeID, keySize)
	if err != nil {
		return "", err
	}

	m := &mount.SafeFormatAndMount{Exec: utilexec.New()}
	return openCryptDevice(c.mapper, source, volumeID, string(dek), integrity, m.GetDiskFormat)
}

// openCryptDevice maps the volume at source to the crypt device identified by volumeID.
func openCryptDevice(device DeviceMapper, source, volumeID, dek string, integrity bool, diskInfo func(disk string) (string, error)) (string, error) {
	packageLock.Lock()
	defer packageLock.Unlock()

	var integrityType string
	keySize := len(dek)

	if integrity {
		if len(dek) != keySizeIntegrity {
			return "", fmt.Errorf("invalid key size for crypt with integrity: expected [%d], got [%d]", keySizeIntegrity, len(dek))
		}
		integrityType = "hmac(sha256)"
	}

	if !integrity && (len(dek) != keySizeCrypt) {
		return "", fmt.Errorf("invalid key length for plain crypt: expected [%d], got [%d]", keySizeCrypt, len(dek))
	}

	klog.V(4).Infof("Mapping device %q to dm-crypt volume %q", source, cryptPrefix+volumeID)
	cryptsetup.SetLogCallback(func(level int, message string) { klog.V(4).Infof("libcryptsetup: %s", message) })

	// Initialize the block device
	if err := device.Init(source); err != nil {
		klog.Errorf("Initializing dm-crypt to map device %q: %s", source, err)
		return "", fmt.Errorf("initializing dm-crypt to map device %q: %w", source, err)
	}
	defer device.Free()

	needWipe := false
	// Try to load LUKS headers
	// If this fails, the device is either not formatted at all, or already formatted with a different FS
	if err := device.Load(nil); err != nil {
		klog.V(4).Infof("Device %q is not formatted as LUKS2 partition, checking for existing format...", source)
		format, err := diskInfo(source)
		if err != nil {
			return "", fmt.Errorf("could not determine if disk is formatted: %w", err)
		}
		if format != "" {
			// Device is already formated, return an error
			klog.Errorf("Disk %q is already formatted as: %s", source, format)
			return "", fmt.Errorf("disk %q is already formatted as: %s", source, format)
		}

		// Device is not formatted, so we can safely create a new LUKS2 partition
		klog.V(4).Infof("Device %q is not formatted. Creating new LUKS2 partition...", source)
		if err := device.Format(
			cryptsetup.LUKS2{
				SectorSize: 4096,
				Integrity:  integrityType,
				PBKDFType: &cryptsetup.PbkdfType{
					// Use low memory recommendation from https://datatracker.ietf.org/doc/html/rfc9106#section-7
					Type:            "argon2id",
					TimeMs:          2000,
					Iterations:      3,
					ParallelThreads: 4,
					MaxMemoryKb:     65536, // ~64MiB
				},
			},
			cryptsetup.GenericParams{
				Cipher:        "aes",
				CipherMode:    "xts-plain64",
				VolumeKey:     dek,
				VolumeKeySize: keySize,
			}); err != nil {
			klog.Errorf("Formatting device %q failed: %s", source, err)
			return "", fmt.Errorf("formatting device %q failed: %w", source, err)
		}
		needWipe = true
	}

	if integrity && needWipe {
		if err := performWipe(device, volumeID, dek); err != nil {
			return "", fmt.Errorf("wiping device: %w", err)
		}
	}

	klog.V(4).Infof("Activating LUKS2 device %q", cryptPrefix+volumeID)

	if err := device.ActivateByVolumeKey(volumeID, dek, keySize, 0); err != nil {
		klog.Errorf("Trying to activate dm-crypt volume: %s", err)
		return "", fmt.Errorf("trying to activate dm-crypt volume: %w", err)
	}

	klog.V(4).Infof("Device %q successfully mapped to dm-crypt volume %q", source, cryptPrefix+volumeID)

	return cryptPrefix + volumeID, nil
}

// performWipe handles setting up parameters and clearing the device for dm-integrity.
func performWipe(device DeviceMapper, volumeID, dek string) error {
	klog.V(4).Infof("Preparing device for dm-integrity. This may take while...")
	tmpDevice := "temporary-cryptsetup-" + volumeID

	// Active as temporary device
	if err := device.ActivateByVolumeKey(tmpDevice, dek, len(dek), (cryptsetup.CRYPT_ACTIVATE_PRIVATE | cryptsetup.CRYPT_ACTIVATE_NO_JOURNAL)); err != nil {
		klog.Errorf("Trying to activate temporary dm-crypt volume: %s", err)
		return fmt.Errorf("trying to activate temporary dm-crypt volume: %w", err)
	}

	// Set progress callbacks
	var progressCallback func(size, offset uint64) int
	if fileInfo, _ := os.Stdout.Stat(); (fileInfo.Mode() & os.ModeCharDevice) != 0 {
		// If we are printing to a terminal we can show continues updates
		progressCallback = func(size, offset uint64) int {
			prog := (float64(offset) / float64(size)) * 100
			fmt.Printf("\033[2K\rWipe in progress: %.2f%%", prog)
			return 0
		}
	} else {
		// No terminal available, limit callbacks to once every 30 seconds to not fill up logs with large amount of progress updates
		ticker := time.NewTicker(30 * time.Second)
		firstReq := make(chan struct{}, 1)
		firstReq <- struct{}{}
		defer ticker.Stop()

		logProgress := func(size, offset uint64) {
			prog := (float64(offset) / float64(size)) * 100
			klog.V(4).Infof("Wipe in progress: %.2f%%", prog)
		}

		progressCallback = func(size, offset uint64) int {
			select {
			case <-firstReq:
				logProgress(size, offset)
			case <-ticker.C:
				logProgress(size, offset)
			default:
			}

			return 0
		}
	}

	// Wipe the device using the same options as used in cryptsetup: https://gitlab.com/cryptsetup/cryptsetup/-/blob/master/src/cryptsetup.c#L1178
	if err := device.Wipe(cryptPrefix+tmpDevice, cryptsetup.CRYPT_WIPE_ZERO, 0, 0, 1024*1024, 0, progressCallback); err != nil {
		return err
	}
	fmt.Println()

	// Deactivate the temporary device
	if err := device.Deactivate(tmpDevice); err != nil {
		klog.Errorf("Deactivating temporary volume: %s", err)
		return fmt.Errorf("deactivating temporary volume: %w", err)
	}

	klog.V(4).Info("dm-integrity successfully initiated")
	return nil
}

// IsIntegrityFS checks if the fstype string contains an integrity suffix.
// If yes, returns the trimmed fstype and true, fstype and false otherwise.
func IsIntegrityFS(fstype string) (string, bool) {
	if strings.HasSuffix(fstype, integrityFSSuffix) {
		return strings.TrimSuffix(fstype, integrityFSSuffix), true
	}
	return fstype, false
}
