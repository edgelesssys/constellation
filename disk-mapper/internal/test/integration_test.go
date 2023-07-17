//go:build integration && linux && cgo

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package integration

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/edgelesssys/constellation/v2/disk-mapper/internal/diskencryption"
	ccryptsetup "github.com/edgelesssys/constellation/v2/internal/cryptsetup"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	cryptsetup "github.com/martinjungblut/go-cryptsetup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

const (
	devicePath   = "testDevice"
	mappedDevice = "mappedDevice"
)

var diskPath = flag.String("disk", "", "Path to the disk to use for the benchmark")

func setup(sizeGB int) error {
	return exec.Command("/bin/dd", "if=/dev/random", fmt.Sprintf("of=%s", devicePath), "bs=1G", fmt.Sprintf("count=%d", sizeGB)).Run()
}

func teardown() error {
	return exec.Command("/bin/rm", "-f", devicePath).Run()
}

func TestMain(m *testing.M) {
	flag.Parse()

	if os.Getuid() != 0 {
		fmt.Printf("This test suite requires root privileges, as libcrypsetup uses the kernel's device mapper.\n")
		os.Exit(1)
	}

	goleak.VerifyTestMain(m,
		// https://github.com/census-instrumentation/opencensus-go/issues/1262
		goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start"),
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

	mapper, free, err := diskencryption.New(devicePath, logger.NewTest(t))
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

	// Try to map disk with incorrect passphrase
	assert.Error(mapper.MapDisk(mappedDevice, "invalid-passphrase"), "was able to map disk with incorrect passphrase")

	// Disk can be reformatted without manually re-initializing a mapper
	passphrase2 := passphrase + "2"
	require.NoError(mapper.FormatDisk(passphrase2), "failed to format disk")
	require.NoError(mapper.MapDisk(mappedDevice, passphrase2), "failed to map disk")
	require.NoError(mapper.UnmapDisk(mappedDevice), "failed to remove disk mapping")
}
