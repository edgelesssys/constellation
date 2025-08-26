//go:build integration && linux && cgo

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: BUSL-1.1
*/

package integration

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/bazelbuild/rules_go/go/runfiles"
	"github.com/edgelesssys/constellation/v2/disk-mapper/internal/diskencryption"
	ccryptsetup "github.com/edgelesssys/constellation/v2/internal/cryptsetup"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	cryptsetup "github.com/martinjungblut/go-cryptsetup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

const (
	backingDisk  = "testDevice"
	mappedDevice = "mappedDevice"
)

var devicePath string

var diskPath = flag.String("disk", "", "Path to the disk to use for the benchmark")

var toolsEnvs = []string{"DD", "RM", "LOSETUP"}

// addToolsToPATH is used to update the PATH to contain necessary tool binaries for
// coreutils.
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

func setup(sizeGB int) error {
	if err := exec.Command("dd", "if=/dev/random", fmt.Sprintf("of=%s", backingDisk), "bs=1G", fmt.Sprintf("count=%d", sizeGB)).Run(); err != nil {
		return err
	}
	cmd := exec.Command("losetup", "-f", "--show", backingDisk)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("losetup failed: %w\nOutput: %s", err, out)
	}
	devicePath = strings.TrimSpace(string(out))
	return nil
}

func teardown() error {
	err := exec.Command("losetup", "-d", devicePath).Run()
	errors.Join(err, exec.Command("rm", "-f", backingDisk).Run())
	return err
}

func TestMain(m *testing.M) {
	flag.Parse()

	// try to become root (best effort)
	_ = syscall.Setuid(0)
	if os.Getuid() != 0 {
		fmt.Printf("This test suite requires root privileges, as libcrypsetup uses the kernel's device mapper.\n")
		os.Exit(1)
	}
	if err := addToolsToPATH(); err != nil {
		fmt.Printf("Failed to add tools to PATH: %v\n", err)
		os.Exit(1)
	}

	goleak.VerifyTestMain(m,
		// https://github.com/census-instrumentation/opencensus-go/issues/1262
		goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start"),
		goleak.IgnoreAnyFunction("github.com/bazelbuild/rules_go/go/tools/bzltestutil.RegisterTimeoutHandler.func1"),
	)

	result := m.Run()
	os.Exit(result)
}

func TestMapper(t *testing.T) {
	cryptsetup.SetDebugLevel(cryptsetup.CRYPT_LOG_ERROR)
	cryptsetup.SetLogCallback(func(_ int, message string) { fmt.Println(message) })

	assert := assert.New(t)
	require := require.New(t)
	require.NoError(setup(1), "failed to setup test disk")
	defer func() { require.NoError(teardown(), "failed to delete test disk") }()

	mapper, free, err := diskencryption.New(devicePath, logger.NewTextLogger(slog.LevelInfo))
	require.NoError(err, "failed to initialize crypt device")
	defer free()

	assert.False(mapper.IsInitialized())

	// Format and map disk
	passphrase := "unit-test"
	require.NoError(mapper.FormatDisk(passphrase), "failed to format disk")
	require.NoError(mapper.MapDisk(mappedDevice, passphrase), "failed to map disk")
	require.NoError(mapper.UnmapDisk(mappedDevice), "failed to remove disk mapping")

	// Make sure token was set
	ccrypt := ccryptsetup.New()
	freeDevice, err := ccrypt.Init(devicePath)
	require.NoError(err, "failed to initialize crypt device")
	defer freeDevice()
	require.NoError(ccrypt.LoadLUKS2(), "failed to load LUKS2")

	tokenJSON, err := ccrypt.TokenJSONGet(ccryptsetup.ConstellationStateDiskTokenID)
	require.NoError(err, "token should have been set")
	var token struct {
		Type              string   `json:"type"`
		Keyslots          []string `json:"keyslots"`
		DiskIsInitialized bool     `json:"diskIsInitialized"`
	}
	require.NoError(json.Unmarshal([]byte(tokenJSON), &token))
	assert.False(token.DiskIsInitialized, "disk should be marked as not initialized")
	assert.False(ccrypt.ConstellationStateDiskTokenIsInitialized(), "disk should be marked as not initialized")

	// Disk should still be marked as not initialized because token is set to false.
	assert.False(mapper.IsInitialized())

	// Set disk as initialized
	assert.NoError(ccrypt.SetConstellationStateDiskToken(ccryptsetup.SetDiskInitialized))

	// Set up a new client and check if the client still sees the disk as initialized
	ccrypt2 := ccryptsetup.New()
	freeDevice2, err := ccrypt2.Init(devicePath)
	require.NoError(err, "failed to initialize crypt device")
	defer freeDevice2()
	require.NoError(ccrypt2.LoadLUKS2(), "failed to load LUKS2")

	tokenJSON, err = ccrypt2.TokenJSONGet(ccryptsetup.ConstellationStateDiskTokenID)
	require.NoError(err, "token should have been set")
	var token2 struct {
		Type              string   `json:"type"`
		Keyslots          []string `json:"keyslots"`
		DiskIsInitialized bool     `json:"diskIsInitialized"`
	}
	require.NoError(json.Unmarshal([]byte(tokenJSON), &token2))
	assert.True(token2.DiskIsInitialized, "disk should be marked as initialized")
	assert.True(ccrypt2.ConstellationStateDiskTokenIsInitialized(), "disk should be marked as initialized")

	// Try to map disk with incorrect passphrase
	assert.Error(mapper.MapDisk(mappedDevice, "invalid-passphrase"), "was able to map disk with incorrect passphrase")

	// Disk can be reformatted without manually re-initializing a mapper
	passphrase2 := passphrase + "2"
	require.NoError(mapper.FormatDisk(passphrase2), "failed to format disk")
	require.NoError(mapper.MapDisk(mappedDevice, passphrase2), "failed to map disk")
	require.NoError(mapper.UnmapDisk(mappedDevice), "failed to remove disk mapping")
}
