//go:build linux && cgo

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cryptmapper

import (
	"context"
	"errors"
	"testing"

	cryptsetup "github.com/martinjungblut/go-cryptsetup"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

type stubCryptDevice struct {
	deviceName       string
	uuid             string
	initErr          error
	initByNameErr    error
	activateErr      error
	activatePassErr  error
	deactivateErr    error
	formatErr        error
	loadErr          error
	keySlotAddCalled bool
	keySlotAddErr    error
	wipeErr          error
	resizeErr        error
}

func (c *stubCryptDevice) Init(string) error {
	return c.initErr
}

func (c *stubCryptDevice) InitByName(string) error {
	return c.initByNameErr
}

func (c *stubCryptDevice) ActivateByVolumeKey(string, string, int, int) error {
	return c.activateErr
}

func (c *stubCryptDevice) ActivateByPassphrase(string, int, string, int) error {
	return c.activatePassErr
}

func (c *stubCryptDevice) Deactivate(string) error {
	return c.deactivateErr
}

func (c *stubCryptDevice) Format(cryptsetup.DeviceType, cryptsetup.GenericParams) error {
	return c.formatErr
}

func (c *stubCryptDevice) Free() bool {
	return true
}

func (c *stubCryptDevice) GetDeviceName() string {
	return c.deviceName
}

func (c *stubCryptDevice) GetUUID() string {
	return c.uuid
}

func (c *stubCryptDevice) Load(cryptsetup.DeviceType) error {
	return c.loadErr
}

func (c *stubCryptDevice) KeyslotAddByVolumeKey(int, string, string) error {
	c.keySlotAddCalled = true
	return c.keySlotAddErr
}

func (c *stubCryptDevice) Wipe(string, int, uint64, uint64, int, int, func(size, offset uint64) int) error {
	return c.wipeErr
}

func (c *stubCryptDevice) Resize(string, uint64) error {
	return c.resizeErr
}

func TestCloseCryptDevice(t *testing.T) {
	testCases := map[string]struct {
		mapper  *stubCryptDevice
		wantErr bool
	}{
		"success": {
			mapper:  &stubCryptDevice{},
			wantErr: false,
		},
		"error on InitByName": {
			mapper:  &stubCryptDevice{initByNameErr: errors.New("error")},
			wantErr: true,
		},
		"error on Deactivate": {
			mapper:  &stubCryptDevice{deactivateErr: errors.New("error")},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			err := closeCryptDevice(tc.mapper, "/dev/some-device", "volume0", "test")
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}

	mapper := New(&fakeKMS{}, &stubCryptDevice{})
	err := mapper.CloseCryptDevice("volume01-unit-test")
	assert.NoError(t, err)
}

func TestOpenCryptDevice(t *testing.T) {
	someErr := errors.New("error")
	getKeyFunc := func(context.Context, string, int) ([]byte, error) {
		return []byte{
			0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA,
			0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA,
			0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA,
			0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA,
		}, nil
	}

	testCases := map[string]struct {
		source    string
		volumeID  string
		integrity bool
		mapper    *stubCryptDevice
		getKey    func(context.Context, string, int) ([]byte, error)
		diskInfo  func(disk string) (string, error)
		wantErr   bool
	}{
		"success with Load": {
			source:   "/dev/some-device",
			volumeID: "volume0",
			mapper:   &stubCryptDevice{},
			getKey:   getKeyFunc,
			diskInfo: func(disk string) (string, error) { return "", nil },
			wantErr:  false,
		},
		"success with error on Load": {
			source:   "/dev/some-device",
			volumeID: "volume0",
			mapper:   &stubCryptDevice{loadErr: someErr},
			getKey:   getKeyFunc,
			diskInfo: func(disk string) (string, error) { return "", nil },
			wantErr:  false,
		},
		"success with integrity": {
			source:    "/dev/some-device",
			volumeID:  "volume0",
			integrity: true,
			mapper:    &stubCryptDevice{loadErr: someErr},
			getKey:    getKeyFunc,
			diskInfo:  func(disk string) (string, error) { return "", nil },
			wantErr:   false,
		},
		"error on Init": {
			source:   "/dev/some-device",
			volumeID: "volume0",
			mapper:   &stubCryptDevice{initErr: someErr},
			getKey:   getKeyFunc,
			diskInfo: func(disk string) (string, error) { return "", nil },
			wantErr:  true,
		},
		"error on Format": {
			source:   "/dev/some-device",
			volumeID: "volume0",
			mapper:   &stubCryptDevice{loadErr: someErr, formatErr: someErr},
			getKey:   getKeyFunc,
			diskInfo: func(disk string) (string, error) { return "", nil },
			wantErr:  true,
		},
		"error on Activate": {
			source:   "/dev/some-device",
			volumeID: "volume0",
			mapper:   &stubCryptDevice{activatePassErr: someErr},
			getKey:   getKeyFunc,
			diskInfo: func(disk string) (string, error) { return "", nil },
			wantErr:  true,
		},
		"error on diskInfo": {
			source:   "/dev/some-device",
			volumeID: "volume0",
			mapper:   &stubCryptDevice{loadErr: someErr},
			getKey:   getKeyFunc,
			diskInfo: func(disk string) (string, error) { return "", someErr },
			wantErr:  true,
		},
		"disk is already formatted": {
			source:   "/dev/some-device",
			volumeID: "volume0",
			mapper:   &stubCryptDevice{loadErr: someErr},
			getKey:   getKeyFunc,
			diskInfo: func(disk string) (string, error) { return "ext4", nil },
			wantErr:  true,
		},
		"error with integrity on wipe": {
			source:    "/dev/some-device",
			volumeID:  "volume0",
			integrity: true,
			mapper:    &stubCryptDevice{loadErr: someErr, wipeErr: someErr},
			getKey:    getKeyFunc,
			diskInfo:  func(disk string) (string, error) { return "", nil },
			wantErr:   true,
		},
		"error with integrity on activate": {
			source:    "/dev/some-device",
			volumeID:  "volume0",
			integrity: true,
			mapper:    &stubCryptDevice{loadErr: someErr, activateErr: someErr},
			getKey:    getKeyFunc,
			diskInfo:  func(disk string) (string, error) { return "", nil },
			wantErr:   true,
		},
		"error with integrity on deactivate": {
			source:    "/dev/some-device",
			volumeID:  "volume0",
			integrity: true,
			mapper:    &stubCryptDevice{loadErr: someErr, deactivateErr: someErr},
			getKey:    getKeyFunc,
			diskInfo:  func(disk string) (string, error) { return "", nil },
			wantErr:   true,
		},
		"error on adding keyslot": {
			source:   "/dev/some-device",
			volumeID: "volume0",
			mapper: &stubCryptDevice{
				loadErr:       someErr,
				keySlotAddErr: someErr,
			},
			getKey:   getKeyFunc,
			diskInfo: func(disk string) (string, error) { return "", nil },
			wantErr:  true,
		},
		"incorrect key length": {
			source:   "/dev/some-device",
			volumeID: "volume0",
			mapper:   &stubCryptDevice{},
			getKey:   func(ctx context.Context, s string, i int) ([]byte, error) { return []byte{0x1, 0x2, 0x3}, nil },
			diskInfo: func(disk string) (string, error) { return "", nil },
			wantErr:  true,
		},
		"incorrect key length with error on Load": {
			source:   "/dev/some-device",
			volumeID: "volume0",
			mapper:   &stubCryptDevice{loadErr: someErr},
			getKey:   func(ctx context.Context, s string, i int) ([]byte, error) { return []byte{0x1, 0x2, 0x3}, nil },
			diskInfo: func(disk string) (string, error) { return "", nil },
			wantErr:  true,
		},
		"getKey fails": {
			source:   "/dev/some-device",
			volumeID: "volume0",
			mapper:   &stubCryptDevice{},
			getKey:   func(ctx context.Context, s string, i int) ([]byte, error) { return nil, someErr },
			diskInfo: func(disk string) (string, error) { return "", nil },
			wantErr:  true,
		},
		"getKey fails with error on Load": {
			source:   "/dev/some-device",
			volumeID: "volume0",
			mapper:   &stubCryptDevice{loadErr: someErr},
			getKey:   func(ctx context.Context, s string, i int) ([]byte, error) { return nil, someErr },
			diskInfo: func(disk string) (string, error) { return "", nil },
			wantErr:  true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			out, err := openCryptDevice(
				context.Background(),
				tc.mapper,
				tc.source,
				tc.volumeID,
				tc.integrity,
				tc.getKey,
				tc.diskInfo,
			)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(cryptPrefix+tc.volumeID, out)

				if tc.mapper.loadErr == nil {
					assert.False(tc.mapper.keySlotAddCalled)
				} else {
					assert.True(tc.mapper.keySlotAddCalled)
				}
			}
		})
	}

	mapper := New(&fakeKMS{}, &stubCryptDevice{})
	_, err := mapper.OpenCryptDevice(context.Background(), "/dev/some-device", "volume01", false)
	assert.NoError(t, err)
}

func TestResizeCryptDevice(t *testing.T) {
	volumeID := "pvc-123"
	someErr := errors.New("error")
	testCases := map[string]struct {
		volumeID string
		device   *stubCryptDevice
		wantErr  bool
	}{
		"success": {
			volumeID: volumeID,
			device:   &stubCryptDevice{},
		},
		"InitByName fails": {
			volumeID: volumeID,
			device:   &stubCryptDevice{initByNameErr: someErr},
			wantErr:  true,
		},
		"Load fails": {
			volumeID: volumeID,
			device:   &stubCryptDevice{loadErr: someErr},
			wantErr:  true,
		},
		"Resize fails": {
			volumeID: volumeID,
			device:   &stubCryptDevice{resizeErr: someErr},
			wantErr:  true,
		},
		"ActivateByPassphrase fails": {
			volumeID: volumeID,
			device:   &stubCryptDevice{activatePassErr: someErr},
			wantErr:  true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			mapper := &CryptMapper{
				kms:    &fakeKMS{},
				mapper: tc.device,
			}

			res, err := mapper.ResizeCryptDevice(context.Background(), tc.volumeID)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(cryptPrefix+tc.volumeID, res)
			}
		})
	}
}

func TestGetDevicePath(t *testing.T) {
	volumeID := "pvc-123"
	someErr := errors.New("error")
	testCases := map[string]struct {
		volumeID string
		device   *stubCryptDevice
		wantErr  bool
	}{
		"success": {
			volumeID: volumeID,
			device:   &stubCryptDevice{deviceName: volumeID},
		},
		"InitByName fails": {
			volumeID: volumeID,
			device:   &stubCryptDevice{initByNameErr: someErr},
			wantErr:  true,
		},
		"GetDeviceName returns nothing": {
			volumeID: volumeID,
			device:   &stubCryptDevice{},
			wantErr:  true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			mapper := &CryptMapper{
				mapper: tc.device,
			}

			res, err := mapper.GetDevicePath(tc.volumeID)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.device.deviceName, res)
			}
		})
	}
}

func TestIsIntegrityFS(t *testing.T) {
	testCases := map[string]struct {
		wantIntegrity bool
		fstype        string
	}{
		"plain ext4": {
			wantIntegrity: false,
			fstype:        "ext4",
		},
		"integrity ext4": {
			wantIntegrity: true,
			fstype:        "ext4",
		},
		"integrity fs": {
			wantIntegrity: false,
			fstype:        "integrity",
		},
		"double integrity": {
			wantIntegrity: true,
			fstype:        "ext4-integrity",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			request := tc.fstype
			if tc.wantIntegrity {
				request = tc.fstype + integrityFSSuffix
			}

			fstype, isIntegrity := IsIntegrityFS(request)

			if tc.wantIntegrity {
				assert.True(isIntegrity)
				assert.Equal(tc.fstype, fstype)
			} else {
				assert.False(isIntegrity)
				assert.Equal(tc.fstype, fstype)
			}
		})
	}
}

type fakeKMS struct{}

func (k *fakeKMS) GetDEK(_ context.Context, _ string, dekSize int) ([]byte, error) {
	key := make([]byte, dekSize)
	for i := range key {
		key[i] = 0x41
	}
	return key, nil
}
