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

	"github.com/edgelesssys/constellation/internal/constants"
	cryptsetup "github.com/martinjungblut/go-cryptsetup"
	"k8s.io/klog/v2"
	mount "k8s.io/mount-utils"
	utilexec "k8s.io/utils/exec"
)

const (
	// LUKSHeaderSize is the amount of bytes taken up by the header of a LUKS2 partition.
	// The header is 16MiB (1048576 Bytes * 16).
	LUKSHeaderSize    = 16777216
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

func init() {
	cryptsetup.SetDebugLevel(cryptsetup.CRYPT_LOG_NORMAL)
	cryptsetup.SetLogCallback(func(level int, message string) { klog.V(4).Infof("libcryptsetup: %s", message) })
}

// KeyCreator is an interface to create data encryption keys.
type KeyCreator interface {
	GetDEK(ctx context.Context, dekID string, dekSize int) ([]byte, error)
}

// DeviceMapper is an interface for device mapper methods.
type DeviceMapper interface {
	// Init initializes a crypt device backed by 'devicePath'.
	// Sets the deviceMapper to the newly allocated Device or returns any error encountered.
	Init(devicePath string) error
	// InitByName initializes a crypt device from provided active device 'name'.
	// Sets the deviceMapper to the newly allocated Device or returns any error encountered.
	InitByName(name string) error
	// ActivateByPassphrase activates a device by using a passphrase from a specific keyslot.
	// Returns nil on success, or an error otherwise.
	ActivateByPassphrase(deviceName string, keyslot int, passphrase string, flags int) error
	// ActivateByVolumeKey activates a device by using a volume key.
	// Returns nil on success, or an error otherwise.
	ActivateByVolumeKey(deviceName string, volumeKey string, volumeKeySize int, flags int) error
	// Deactivate deactivates a device.
	// Returns nil on success, or an error otherwise.
	Deactivate(deviceName string) error
	// Format formats a Device, using a specific device type, and type-independent parameters.
	// Returns nil on success, or an error otherwise.
	Format(deviceType cryptsetup.DeviceType, genericParams cryptsetup.GenericParams) error
	// Free releases crypt device context and used memory.
	Free() bool
	// GetDeviceName gets the path to the underlying device.
	GetDeviceName() string
	// GetUUID gets the devices UUID
	GetUUID() string
	// Load loads crypt device parameters from the on-disk header.
	// Returns nil on success, or an error otherwise.
	Load(cryptsetup.DeviceType) error
	// KeyslotAddByVolumeKey adds a key slot using a volume key to perform the required security check.
	// Returns nil on success, or an error otherwise.
	KeyslotAddByVolumeKey(keyslot int, volumeKey string, passphrase string) error
	// Wipe removes existing data and clears the device for use with dm-integrity.
	// Returns nil on success, or an error otherwise.
	Wipe(devicePath string, pattern int, offset, length uint64, wipeBlockSize int, flags int, progress func(size, offset uint64) int) error
	// Resize the crypt device.
	// Returns nil on success, or an error otherwise.
	Resize(name string, newSize uint64) error
}

// cryptDevice is a wrapper for cryptsetup.Device.
type CryptDevice struct {
	*cryptsetup.Device
}

// Init initializes a crypt device backed by 'devicePath'.
// Sets the cryptDevice's deviceMapper to the newly allocated Device or returns any error encountered.
func (c *CryptDevice) Init(devicePath string) error {
	device, err := cryptsetup.Init(devicePath)
	if err != nil {
		return err
	}
	c.Device = device
	return nil
}

// InitByName initializes a crypt device from provided active device 'name'.
// Sets the deviceMapper to the newly allocated Device or returns any error encountered.
func (c *CryptDevice) InitByName(name string) error {
	device, err := cryptsetup.InitByName(name)
	if err != nil {
		return err
	}
	c.Device = device
	return nil
}

// Free releases crypt device context and used memory.
func (c *CryptDevice) Free() bool {
	res := c.Device.Free()
	c.Device = nil
	return res
}

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
		return fmt.Errorf("getting device path for disk %q: %w", cryptPrefix+volumeID, err)
	}
	if err := closeCryptDevice(c.mapper, source, volumeID, "crypt"); err != nil {
		return fmt.Errorf("closing crypt device: %w", err)
	}

	integrity, err := filepath.EvalSymlinks(cryptPrefix + volumeID + integritySuffix)
	if err == nil {
		// If device was created with integrity, we need to also close the integrity device
		integrityErr := closeCryptDevice(c.mapper, integrity, volumeID+integritySuffix, "integrity")
		if integrityErr != nil {
			klog.Errorf("Failed to close integrity device: %s", integrityErr)
			return integrityErr
		}
	}
	if err != nil {
		var pathErr *fs.PathError
		if errors.As(err, &pathErr) {
			// integrity device does not exist
			return nil
		}
		return fmt.Errorf("getting device path for disk %q: %w", cryptPrefix+volumeID, err)
	}

	return nil
}

// OpenCryptDevice maps the volume at source to the crypt device identified by volumeID.
// The key used to encrypt the volume is fetched using CryptMapper's kms client.
func (c *CryptMapper) OpenCryptDevice(ctx context.Context, source, volumeID string, integrity bool) (string, error) {
	m := &mount.SafeFormatAndMount{Exec: utilexec.New()}
	return openCryptDevice(ctx, c.mapper, source, volumeID, integrity, c.kms.GetDEK, m.GetDiskFormat)
}

// ResizeCryptDevice resizes the underlying crypt device and returns the mapped device path.
func (c *CryptMapper) ResizeCryptDevice(ctx context.Context, volumeID string) (string, error) {
	dek, err := c.kms.GetDEK(ctx, volumeID, constants.StateDiskKeyLength)
	if err != nil {
		return "", err
	}
	klog.V(4).Infof("Resizing LUKS2 partition %q", cryptPrefix+volumeID)

	if err := resizeCryptDevice(c.mapper, volumeID, string(dek)); err != nil {
		return "", err
	}

	return cryptPrefix + volumeID, nil
}

// GetDeviceName returns the real device name of a mapped crypt device.
func (c *CryptMapper) GetDevicePath(volumeID string) (string, error) {
	return getDevicePath(c.mapper, strings.TrimPrefix(volumeID, cryptPrefix))
}

// closeCryptDevice closes the crypt device mapped for volumeID.
func closeCryptDevice(device DeviceMapper, source, volumeID, deviceType string) error {
	packageLock.Lock()
	defer packageLock.Unlock()

	klog.V(4).Infof("Unmapping dm-%s volume %q for device %q", deviceType, cryptPrefix+volumeID, source)

	if err := device.InitByName(volumeID); err != nil {
		klog.Errorf("Failed to initialize dm-%s to unmap device %q: %s", deviceType, source, err)
		return fmt.Errorf("initializing dm-%s to unmap device %q: %w", deviceType, source, err)
	}
	defer device.Free()

	if err := device.Deactivate(volumeID); err != nil {
		klog.Errorf("Failed to deactivate dm-%s volume %q for device %q: %s", deviceType, cryptPrefix+volumeID, source, err)
		return fmt.Errorf("deactivating dm-%s volume %q for device %q: %w", deviceType, cryptPrefix+volumeID, source, err)
	}

	klog.V(4).Infof("Successfully unmapped dm-%s volume %q for device %q", deviceType, cryptPrefix+volumeID, source)
	return nil
}

// openCryptDevice maps the volume at source to the crypt device identified by volumeID.
func openCryptDevice(ctx context.Context, device DeviceMapper, source, volumeID string, integrity bool,
	getKey func(ctx context.Context, keyID string, keySize int) ([]byte, error), diskInfo func(disk string) (string, error),
) (string, error) {
	packageLock.Lock()
	defer packageLock.Unlock()

	var integrityType string
	keySize := keySizeCrypt
	if integrity {
		integrityType = "hmac(sha256)"
		keySize = keySizeIntegrity
	}

	klog.V(4).Infof("Mapping device %q to dm-crypt volume %q", source, cryptPrefix+volumeID)

	// Initialize the block device
	if err := device.Init(source); err != nil {
		klog.Errorf("Initializing dm-crypt to map device %q: %s", source, err)
		return "", fmt.Errorf("initializing dm-crypt to map device %q: %w", source, err)
	}
	defer device.Free()

	var passphrase []byte
	// Try to load LUKS headers
	// If this fails, the device is either not formatted at all, or already formatted with a different FS
	if err := device.Load(cryptsetup.LUKS2{}); err != nil {
		klog.V(4).Infof("Device %q is not formatted as LUKS2 partition, checking for existing format...", source)
		format, err := diskInfo(source)
		if err != nil {
			return "", fmt.Errorf("determining if disk is formatted: %w", err)
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
				VolumeKeySize: keySize,
			}); err != nil {
			klog.Errorf("Failed to format device %q: %s", source, err)
			return "", fmt.Errorf("formatting device %q: %w", source, err)
		}

		uuid := device.GetUUID()
		klog.V(4).Infof("Fetching data encryption key for volume %q", volumeID)
		passphrase, err = getKey(ctx, uuid, constants.StateDiskKeyLength)
		if err != nil {
			return "", err
		}
		if len(passphrase) != constants.StateDiskKeyLength {
			return "", fmt.Errorf("expected key length to be [%d] but got [%d]", constants.StateDiskKeyLength, len(passphrase))
		}

		// Add a new keyslot using the internal volume key
		if err := device.KeyslotAddByVolumeKey(0, "", string(passphrase)); err != nil {
			return "", fmt.Errorf("adding keyslot: %w", err)
		}

		if integrity {
			if err := performWipe(device, volumeID); err != nil {
				return "", fmt.Errorf("wiping device: %w", err)
			}
		}
	} else {
		uuid := device.GetUUID()
		klog.V(4).Infof("Fetching data encryption key for volume %q", volumeID)
		passphrase, err = getKey(ctx, uuid, constants.StateDiskKeyLength)
		if err != nil {
			return "", err
		}
		if len(passphrase) != constants.StateDiskKeyLength {
			return "", fmt.Errorf("expected key length to be [%d] but got [%d]", constants.StateDiskKeyLength, len(passphrase))
		}
	}

	klog.V(4).Infof("Activating LUKS2 device %q", cryptPrefix+volumeID)

	if err := device.ActivateByPassphrase(volumeID, 0, string(passphrase), 0); err != nil {
		klog.Errorf("Trying to activate dm-crypt volume: %s", err)
		return "", fmt.Errorf("trying to activate dm-crypt volume: %w", err)
	}

	klog.V(4).Infof("Device %q successfully mapped to dm-crypt volume %q", source, cryptPrefix+volumeID)

	return cryptPrefix + volumeID, nil
}

// performWipe handles setting up parameters and clearing the device for dm-integrity.
func performWipe(device DeviceMapper, volumeID string) error {
	klog.V(4).Infof("Preparing device for dm-integrity. This may take while...")
	tmpDevice := "temporary-cryptsetup-" + volumeID

	// Active as temporary device
	if err := device.ActivateByVolumeKey(tmpDevice, "", 0, (cryptsetup.CRYPT_ACTIVATE_PRIVATE | cryptsetup.CRYPT_ACTIVATE_NO_JOURNAL)); err != nil {
		klog.Errorf("Trying to activate temporary dm-crypt volume: %s", err)
		return fmt.Errorf("trying to activate temporary dm-crypt volume: %w", err)
	}

	// Set progress callbacks
	var progressCallback func(size, offset uint64) int
	if fileInfo, _ := os.Stdout.Stat(); (fileInfo.Mode() & os.ModeCharDevice) != 0 {
		// If we are printing to a terminal we can show continues updates
		progressCallback = func(size, offset uint64) int {
			prog := (float64(offset) / float64(size)) * 100
			fmt.Printf("\033[1A\033[2K\rWipe in progress: %.2f%%\n", prog)
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

	// Wipe the device using the same options as used in cryptsetup: https://gitlab.com/cryptsetup/cryptsetup/-/blob/v2.4.3/src/cryptsetup.c#L1345
	if err := device.Wipe(cryptPrefix+tmpDevice, cryptsetup.CRYPT_WIPE_ZERO, 0, 0, 1024*1024, 0, progressCallback); err != nil {
		return err
	}

	// Deactivate the temporary device
	if err := device.Deactivate(tmpDevice); err != nil {
		klog.Errorf("Deactivating temporary volume: %s", err)
		return fmt.Errorf("deactivating temporary volume: %w", err)
	}

	klog.V(4).Info("dm-integrity successfully initiated")
	return nil
}

func resizeCryptDevice(device DeviceMapper, name, passphrase string) error {
	packageLock.Lock()
	defer packageLock.Unlock()

	if err := device.InitByName(name); err != nil {
		return fmt.Errorf("initializing device: %w", err)
	}
	defer device.Free()

	if err := device.Load(cryptsetup.LUKS2{}); err != nil {
		return fmt.Errorf("loading device: %w", err)
	}

	if err := device.ActivateByPassphrase("", 0, passphrase, cryptsetup.CRYPT_ACTIVATE_KEYRING_KEY); err != nil {
		klog.Errorf("Unable to activate keyring for crypt device %q with passphrase: %s", name, err)
		return fmt.Errorf("activating keyrung for crypt device %q with passphrase: %w", name, err)
	}

	if err := device.Resize(name, 0); err != nil {
		klog.Errorf("Unable to resize crypt device: %s", err)
		return fmt.Errorf("resizing device: %w", err)
	}
	klog.V(4).Infof("Successfully resized LUKS2 partition for %q", cryptPrefix+name)

	return nil
}

func getDevicePath(device DeviceMapper, name string) (string, error) {
	packageLock.Lock()
	defer packageLock.Unlock()

	if err := device.InitByName(name); err != nil {
		return "", fmt.Errorf("initializing device: %w", err)
	}
	defer device.Free()

	deviceName := device.GetDeviceName()
	if deviceName == "" {
		return "", errors.New("unable to determine device name")
	}
	return deviceName, nil
}

// IsIntegrityFS checks if the fstype string contains an integrity suffix.
// If yes, returns the trimmed fstype and true, fstype and false otherwise.
func IsIntegrityFS(fstype string) (string, bool) {
	if strings.HasSuffix(fstype, integrityFSSuffix) {
		return strings.TrimSuffix(fstype, integrityFSSuffix), true
	}
	return fstype, false
}
