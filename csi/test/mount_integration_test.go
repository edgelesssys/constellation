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
	devicePath string = "testDevice"
	deviceName string = "testdeviceName"
)

func setup() {
	if err := exec.Command("/bin/dd", "if=/dev/zero", fmt.Sprintf("of=%s", devicePath), "bs=64M", "count=1").Run(); err != nil {
		panic(err)
	}
}

func teardown(devicePath string) {
	if err := exec.Command("/bin/rm", "-f", devicePath).Run(); err != nil {
		panic(err)
	}
}

func cp(source, target string) error {
	return exec.Command("cp", source, target).Run()
}

func resize() {
	if err := exec.Command("/bin/dd", "if=/dev/zero", fmt.Sprintf("of=%s", devicePath), "bs=32M", "count=1", "oflag=append", "conv=notrunc").Run(); err != nil {
		panic(err)
	}
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
	defer teardown(devicePath)

	mapper := cryptmapper.New(&fakeKMS{})

	newPath, err := mapper.OpenCryptDevice(context.Background(), devicePath, deviceName, false)
	require.NoError(err)
	defer func() {
		_ = mapper.CloseCryptDevice(deviceName)
	}()

	// assert crypt device got created
	_, err = os.Stat(newPath)
	require.NoError(err)
	// assert no integrity device got created
	_, err = os.Stat(newPath + "_dif")
	assert.True(os.IsNotExist(err))

	// Opening the same device should return the same path and not error
	newPath2, err := mapper.OpenCryptDevice(context.Background(), devicePath, deviceName, false)
	require.NoError(err)
	assert.Equal(newPath, newPath2)

	// Resize the device
	resize()

	resizedPath, err := mapper.ResizeCryptDevice(context.Background(), deviceName)
	require.NoError(err)
	assert.Equal("/dev/mapper/"+deviceName, resizedPath)

	assert.NoError(mapper.CloseCryptDevice(deviceName))

	// assert crypt device got removed
	_, err = os.Stat(newPath)
	assert.True(os.IsNotExist(err))

	// check if we can reopen the device
	_, err = mapper.OpenCryptDevice(context.Background(), devicePath, deviceName, true)
	assert.NoError(err)
	assert.NoError(mapper.CloseCryptDevice(deviceName))
}

func TestOpenAndCloseIntegrity(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	setup()
	defer teardown(devicePath)

	mapper := cryptmapper.New(&fakeKMS{})

	newPath, err := mapper.OpenCryptDevice(context.Background(), devicePath, deviceName, true)
	require.NoError(err)
	assert.Equal("/dev/mapper/"+deviceName, newPath)

	// assert crypt device got created
	_, err = os.Stat(newPath)
	assert.NoError(err)
	// assert integrity device got created
	_, err = os.Stat(newPath + "_dif")
	assert.NoError(err)

	// Opening the same device should return the same path and not error
	newPath2, err := mapper.OpenCryptDevice(context.Background(), devicePath, deviceName, true)
	require.NoError(err)
	assert.Equal(newPath, newPath2)

	// integrity devices do not support resizing
	resize()
	_, err = mapper.ResizeCryptDevice(context.Background(), deviceName)
	assert.Error(err)

	assert.NoError(mapper.CloseCryptDevice(deviceName))

	// assert crypt device got removed
	_, err = os.Stat(newPath)
	assert.True(os.IsNotExist(err))
	// assert integrity device got removed
	_, err = os.Stat(newPath + "_dif")
	assert.True(os.IsNotExist(err))

	// check if we can reopen the device
	_, err = mapper.OpenCryptDevice(context.Background(), devicePath, deviceName, true)
	assert.NoError(err)
	assert.NoError(mapper.CloseCryptDevice(deviceName))
}

func TestDeviceCloning(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	setup()
	defer teardown(devicePath)

	mapper := cryptmapper.New(&dynamicKMS{})

	_, err := mapper.OpenCryptDevice(context.Background(), devicePath, deviceName, false)
	assert.NoError(err)

	require.NoError(cp(devicePath, devicePath+"-copy"))
	defer teardown(devicePath + "-copy")

	_, err = mapper.OpenCryptDevice(context.Background(), devicePath+"-copy", deviceName+"-copy", false)
	assert.NoError(err)

	assert.NoError(mapper.CloseCryptDevice(deviceName))
	assert.NoError(mapper.CloseCryptDevice(deviceName + "-copy"))
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
