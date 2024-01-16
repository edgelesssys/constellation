/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cryptmapper

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m, goleak.IgnoreAnyFunction("github.com/bazelbuild/rules_go/go/tools/bzltestutil.RegisterTimeoutHandler.func1"))
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
			mapper:  &stubCryptDevice{initByNameErr: assert.AnError},
			wantErr: true,
		},
		"error on Deactivate": {
			mapper:  &stubCryptDevice{deactivateErr: assert.AnError},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			mapper := &CryptMapper{
				kms:    &fakeKMS{},
				mapper: testMapper(tc.mapper),
			}
			err := mapper.closeCryptDevice("/dev/mapper/volume01", "volume01-unit-test", "crypt")
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}

	mapper := &CryptMapper{
		mapper:        testMapper(&stubCryptDevice{}),
		kms:           &fakeKMS{},
		getDiskFormat: getDiskFormat,
	}
	err := mapper.CloseCryptDevice("volume01-unit-test")
	assert.NoError(t, err)
}

func TestOpenCryptDevice(t *testing.T) {
	testCases := map[string]struct {
		source    string
		volumeID  string
		integrity bool
		mapper    *stubCryptDevice
		kms       keyCreator
		diskInfo  func(disk string) (string, error)
		wantErr   bool
	}{
		"success with Load": {
			source:   "/dev/some-device",
			volumeID: "volume0",
			mapper:   &stubCryptDevice{},
			kms:      &fakeKMS{},
			diskInfo: func(disk string) (string, error) { return "", nil },
			wantErr:  false,
		},
		"success with error on Load": {
			source:   "/dev/some-device",
			volumeID: "volume0",
			mapper:   &stubCryptDevice{loadErr: assert.AnError},
			kms:      &fakeKMS{},
			diskInfo: func(disk string) (string, error) { return "", nil },
			wantErr:  false,
		},
		"success with integrity": {
			source:    "/dev/some-device",
			volumeID:  "volume0",
			integrity: true,
			mapper:    &stubCryptDevice{loadErr: assert.AnError},
			kms:       &fakeKMS{},
			diskInfo:  func(disk string) (string, error) { return "", nil },
			wantErr:   false,
		},
		"error on Init": {
			source:   "/dev/some-device",
			volumeID: "volume0",
			mapper:   &stubCryptDevice{initErr: assert.AnError},
			kms:      &fakeKMS{},
			diskInfo: func(disk string) (string, error) { return "", nil },
			wantErr:  true,
		},
		"error on Format": {
			source:   "/dev/some-device",
			volumeID: "volume0",
			mapper:   &stubCryptDevice{loadErr: assert.AnError, formatErr: assert.AnError},
			kms:      &fakeKMS{},
			diskInfo: func(disk string) (string, error) { return "", nil },
			wantErr:  true,
		},
		"error on Activate": {
			source:   "/dev/some-device",
			volumeID: "volume0",
			mapper:   &stubCryptDevice{activatePassErr: assert.AnError},
			kms:      &fakeKMS{},
			diskInfo: func(disk string) (string, error) { return "", nil },
			wantErr:  true,
		},
		"error on diskInfo": {
			source:   "/dev/some-device",
			volumeID: "volume0",
			mapper:   &stubCryptDevice{loadErr: assert.AnError},
			kms:      &fakeKMS{},
			diskInfo: func(disk string) (string, error) { return "", assert.AnError },
			wantErr:  true,
		},
		"disk is already formatted": {
			source:   "/dev/some-device",
			volumeID: "volume0",
			mapper:   &stubCryptDevice{loadErr: assert.AnError},
			kms:      &fakeKMS{},
			diskInfo: func(disk string) (string, error) { return "ext4", nil },
			wantErr:  true,
		},
		"error with integrity on wipe": {
			source:    "/dev/some-device",
			volumeID:  "volume0",
			integrity: true,
			mapper:    &stubCryptDevice{loadErr: assert.AnError, wipeErr: assert.AnError},
			kms:       &fakeKMS{},
			diskInfo:  func(disk string) (string, error) { return "", nil },
			wantErr:   true,
		},
		"error on adding keyslot": {
			source:   "/dev/some-device",
			volumeID: "volume0",
			mapper:   &stubCryptDevice{loadErr: assert.AnError, keySlotAddErr: assert.AnError},
			kms:      &fakeKMS{},
			diskInfo: func(disk string) (string, error) { return "", nil },
			wantErr:  true,
		},
		"incorrect key length": {
			source:   "/dev/some-device",
			volumeID: "volume0",
			mapper:   &stubCryptDevice{},
			kms:      &fakeKMS{presetKey: []byte{0x1, 0x2, 0x3}},
			diskInfo: func(disk string) (string, error) { return "", nil },
			wantErr:  true,
		},
		"incorrect key length with error on Load": {
			source:   "/dev/some-device",
			volumeID: "volume0",
			mapper:   &stubCryptDevice{loadErr: assert.AnError},
			kms:      &fakeKMS{presetKey: []byte{0x1, 0x2, 0x3}},
			diskInfo: func(disk string) (string, error) { return "", nil },
			wantErr:  true,
		},
		"getKey fails": {
			source:   "/dev/some-device",
			volumeID: "volume0",
			mapper:   &stubCryptDevice{},
			kms:      &fakeKMS{getDEKErr: assert.AnError},
			diskInfo: func(disk string) (string, error) { return "", nil },
			wantErr:  true,
		},
		"getKey fails with error on Load": {
			source:   "/dev/some-device",
			volumeID: "volume0",
			mapper:   &stubCryptDevice{loadErr: assert.AnError},
			kms:      &fakeKMS{getDEKErr: assert.AnError},
			diskInfo: func(disk string) (string, error) { return "", nil },
			wantErr:  true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			mapper := &CryptMapper{
				mapper:        testMapper(tc.mapper),
				kms:           tc.kms,
				getDiskFormat: tc.diskInfo,
			}

			out, err := mapper.OpenCryptDevice(context.Background(), tc.source, tc.volumeID, tc.integrity)
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

	mapper := &CryptMapper{
		mapper:        testMapper(&stubCryptDevice{}),
		kms:           &fakeKMS{},
		getDiskFormat: getDiskFormat,
	}
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
				mapper: testMapper(tc.device),
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
				mapper: testMapper(tc.device),
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

type fakeKMS struct {
	presetKey []byte
	getDEKErr error
}

func (k *fakeKMS) GetDEK(_ context.Context, _ string, dekSize int) ([]byte, error) {
	if k.getDEKErr != nil {
		return nil, k.getDEKErr
	}
	if k.presetKey != nil {
		return k.presetKey, nil
	}
	return bytes.Repeat([]byte{0xAA}, dekSize), nil
}

type stubCryptDevice struct {
	deviceName       string
	uuid             string
	uuidErr          error
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

func (c *stubCryptDevice) Init(_ string) (func(), error) {
	return func() {}, c.initErr
}

func (c *stubCryptDevice) InitByName(_ string) (func(), error) {
	return func() {}, c.initByNameErr
}

func (c *stubCryptDevice) ActivateByVolumeKey(_, _ string, _, _ int) error {
	return c.activateErr
}

func (c *stubCryptDevice) ActivateByPassphrase(_ string, _ int, _ string, _ int) error {
	return c.activatePassErr
}

func (c *stubCryptDevice) Deactivate(_ string) error {
	return c.deactivateErr
}

func (c *stubCryptDevice) Format(_ bool) error {
	return c.formatErr
}

func (c *stubCryptDevice) Free() {}

func (c *stubCryptDevice) GetDeviceName() string {
	return c.deviceName
}

func (c *stubCryptDevice) GetUUID() (string, error) {
	return c.uuid, c.uuidErr
}

func (c *stubCryptDevice) LoadLUKS2() error {
	return c.loadErr
}

func (c *stubCryptDevice) KeyslotAddByVolumeKey(_ int, _ string, _ string) error {
	c.keySlotAddCalled = true
	return c.keySlotAddErr
}

func (c *stubCryptDevice) Wipe(_ string, _ int, _ int, _ func(size, offset uint64), _ time.Duration) error {
	return c.wipeErr
}

func (c *stubCryptDevice) Resize(_ string, _ uint64) error {
	return c.resizeErr
}

func testMapper(stub *stubCryptDevice) func() deviceMapper {
	return func() deviceMapper {
		return stub
	}
}
