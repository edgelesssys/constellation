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
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"

	"github.com/bazelbuild/rules_go/go/runfiles"
	"github.com/edgelesssys/constellation/v2/csi/cryptmapper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

const (
	devicePath string = "testDevice"
	deviceName string = "testDeviceName"
)

var toolsEnvs []string = []string{"CP", "DD", "RM", "FSCK_EXT4", "MKFS_EXT4", "BLKID", "FSCK", "MOUNT", "UMOUNT"}

// addToolsToPATH is used to update the PATH to contain necessary tool binaries for
// coreutils, util-linux and ext4.
func addToolsToPATH() error {
	path := ":" + os.Getenv("PATH") + ":"
	for _, tool := range toolsEnvs {
		toolPath := os.Getenv(tool)
		if toolPath == "" {
			continue
		}
		toolPath, err := runfiles.Rlocation(toolPath)
		if err != nil {
			return err
		}
		pathComponent := filepath.Dir(toolPath)
		if strings.Contains(path, ":"+pathComponent+":") {
			continue
		}
		path = ":" + pathComponent + path
	}
	path = strings.Trim(path, ":")
	os.Setenv("PATH", path)
	return nil
}

func setup(devicePath string) {
	if err := exec.Command("dd", "if=/dev/zero", fmt.Sprintf("of=%s", devicePath), "bs=64M", "count=1").Run(); err != nil {
		panic(err)
	}
}

func teardown(devicePath string) {
	if err := exec.Command("rm", "-f", devicePath).Run(); err != nil {
		panic(err)
	}
}

func cp(source, target string) error {
	return exec.Command("cp", source, target).Run()
}

func resize(devicePath string) {
	if err := exec.Command("dd", "if=/dev/zero", fmt.Sprintf("of=%s", devicePath), "bs=32M", "count=1", "oflag=append", "conv=notrunc").Run(); err != nil {
		panic(err)
	}
}

func TestMain(m *testing.M) {
	// try to become root (best effort)
	_ = syscall.Setuid(0)
	if os.Getuid() != 0 {
		fmt.Printf("This test suite requires root privileges, as libcryptsetup uses the kernel's device mapper.\n")
		os.Exit(1)
	}
	if err := addToolsToPATH(); err != nil {
		fmt.Printf("Failed to add tools to PATH: %v\n", err)
		os.Exit(1)
	}

	goleak.VerifyTestMain(m)

	result := m.Run()
	os.Exit(result)
}

func TestOpenAndClose(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	setup(devicePath)
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
	resize(devicePath)

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
	setup(devicePath)
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
	resize(devicePath)
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
	setup(devicePath)
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

func TestConcurrency(t *testing.T) {
	assert := assert.New(t)
	setup(devicePath)
	defer teardown(devicePath)

	device2 := devicePath + "-2"
	setup(device2)
	defer teardown(device2)

	mapper := cryptmapper.New(&fakeKMS{})

	wg := sync.WaitGroup{}
	runTest := func(path, name string) {
		newPath, err := mapper.OpenCryptDevice(context.Background(), path, name, false)
		assert.NoError(err)
		defer func() {
			_ = mapper.CloseCryptDevice(name)
		}()

		// assert crypt device got created
		_, err = os.Stat(newPath)
		assert.NoError(err)
		// assert no integrity device got created
		_, err = os.Stat(newPath + "_dif")
		assert.True(os.IsNotExist(err))
		assert.NoError(mapper.CloseCryptDevice(name))
		wg.Done()
	}

	wg.Add(2)
	go runTest(devicePath, deviceName)
	go runTest(device2, deviceName+"-2")
	wg.Wait()
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
