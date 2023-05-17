//go:build integration && linux && cgo

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package integration

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/edgelesssys/constellation/v2/csi/cryptmapper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

const (
	DevicePath string = "testDevice"
	DeviceName string = "testDeviceName"
)

func setup() {
	_ = exec.Command("/bin/dd", "if=/dev/zero", fmt.Sprintf("of=%s", DevicePath), "bs=64M", "count=1").Run()
}

func teardown(devicePath string) {
	_ = exec.Command("/bin/rm", "-f", devicePath).Run()
}

func cp(source, target string) error {
	return exec.Command("cp", source, target).Run()
}

func resize() {
	_ = exec.Command("/bin/dd", "if=/dev/zero", fmt.Sprintf("of=%s", DevicePath), "bs=32M", "count=1", "oflag=append", "conv=notrunc").Run()
}

func TestMain(m *testing.M) {
	if os.Getuid() != 0 {
		fmt.Printf("This test suite requires root privileges, as libcryptsetup uses the kernel's device mapper.\n")
		os.Exit(1)
	}

	goleak.VerifyTestMain(m)

	result := m.Run()
	os.Exit(result)
}

func TestOpenAndClose(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	setup()
	defer teardown(DevicePath)

	mapper := cryptmapper.New(&fakeKMS{}, &cryptmapper.CryptDevice{})

	newPath, err := mapper.OpenCryptDevice(context.Background(), DevicePath, DeviceName, false)
	require.NoError(err)
	assert.Equal("/dev/mapper/"+DeviceName, newPath)

	// assert crypt device got created
	_, err = os.Stat(newPath)
	assert.NoError(err)
	// assert no integrity device got created
	_, err = os.Stat(newPath + "_dif")
	assert.True(os.IsNotExist(err))

	// Resize the device
	resize()

	resizedPath, err := mapper.ResizeCryptDevice(context.Background(), DeviceName)
	require.NoError(err)
	assert.Equal("/dev/mapper/"+DeviceName, resizedPath)

	assert.NoError(mapper.CloseCryptDevice(DeviceName))

	// assert crypt device got removed
	_, err = os.Stat(newPath)
	assert.True(os.IsNotExist(err))

	// check if we can reopen the device
	_, err = mapper.OpenCryptDevice(context.Background(), DevicePath, DeviceName, true)
	assert.NoError(err)
	assert.NoError(mapper.CloseCryptDevice(DeviceName))
}

func TestOpenAndCloseIntegrity(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	setup()
	defer teardown(DevicePath)

	mapper := cryptmapper.New(&fakeKMS{}, &cryptmapper.CryptDevice{})

	newPath, err := mapper.OpenCryptDevice(context.Background(), DevicePath, DeviceName, true)
	require.NoError(err)
	assert.Equal("/dev/mapper/"+DeviceName, newPath)

	// assert crypt device got created
	_, err = os.Stat(newPath)
	assert.NoError(err)
	// assert integrity device got created
	_, err = os.Stat(newPath + "_dif")
	assert.NoError(err)

	// integrity devices do not support resizing
	resize()
	_, err = mapper.ResizeCryptDevice(context.Background(), DeviceName)
	assert.Error(err)

	assert.NoError(mapper.CloseCryptDevice(DeviceName))

	// assert crypt device got removed
	_, err = os.Stat(newPath)
	assert.True(os.IsNotExist(err))
	// assert integrity device got removed
	_, err = os.Stat(newPath + "_dif")
	assert.True(os.IsNotExist(err))

	// check if we can reopen the device
	_, err = mapper.OpenCryptDevice(context.Background(), DevicePath, DeviceName, true)
	assert.NoError(err)
	assert.NoError(mapper.CloseCryptDevice(DeviceName))
}

func TestDeviceCloning(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	setup()
	defer teardown(DevicePath)

	mapper := cryptmapper.New(&dynamicKMS{}, &cryptmapper.CryptDevice{})

	_, err := mapper.OpenCryptDevice(context.Background(), DevicePath, DeviceName, false)
	assert.NoError(err)

	require.NoError(cp(DevicePath, DevicePath+"-copy"))
	defer teardown(DevicePath + "-copy")

	_, err = mapper.OpenCryptDevice(context.Background(), DevicePath+"-copy", DeviceName+"-copy", false)
	assert.NoError(err)

	assert.NoError(mapper.CloseCryptDevice(DeviceName))
	assert.NoError(mapper.CloseCryptDevice(DeviceName + "-copy"))
}

type fakeKMS struct{}

func (k *fakeKMS) GetDEK(_ context.Context, _ string, dekSize int) ([]byte, error) {
	key := make([]byte, dekSize)
	for i := range key {
		key[i] = 0x41
	}
	return key, nil
}

type dynamicKMS struct{}

func (k *dynamicKMS) GetDEK(_ context.Context, dekID string, dekSize int) ([]byte, error) {
	key := make([]byte, dekSize)
	for i := range key {
		key[i] = 0x41 ^ dekID[i%len(dekID)]
	}
	return key, nil
}
