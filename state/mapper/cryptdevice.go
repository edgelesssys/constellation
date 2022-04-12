package mapper

import cryptsetup "github.com/martinjungblut/go-cryptsetup"

type cryptDevice interface {
	// ActivateByPassphrase activates a device by using a passphrase from a specific keyslot.
	// Returns nil on success, or an error otherwise.
	// C equivalent: crypt_activate_by_passphrase
	ActivateByPassphrase(deviceName string, keyslot int, passphrase string, flags int) error
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
	// GetUUID gets the device's UUID.
	// C equivalent: crypt_get_uuid
	GetUUID() string
	// Load loads crypt device parameters from the on-disk header.
	// Returns nil on success, or an error otherwise.
	// C equivalent: crypt_load
	Load(cryptsetup.DeviceType) error
	// KeyslotAddByVolumeKey adds a key slot using a volume key to perform the required security check.
	// Returns nil on success, or an error otherwise.
	// C equivalent: crypt_keyslot_add_by_volume_key
	KeyslotAddByVolumeKey(keyslot int, volumeKey string, passphrase string) error
	// KeyslotChangeByPassphrase changes a defined a key slot using a previously added passphrase to perform the required security check.
	// Returns nil on success, or an error otherwise.
	// C equivalent: crypt_keyslot_change_by_passphrase
	KeyslotChangeByPassphrase(currentKeyslot int, newKeyslot int, currentPassphrase string, newPassphrase string) error
}
