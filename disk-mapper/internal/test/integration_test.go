//go:build integration && linux && cgo

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package integration

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/edgelesssys/constellation/v2/disk-mapper/internal/diskencryption"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/martinjungblut/go-cryptsetup"
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

	assert.False(mapper.IsLUKSDevice())

	// Format and map disk
	passphrase := "unit-test"
	require.NoError(mapper.FormatDisk(passphrase), "failed to format disk")
	require.NoError(mapper.MapDisk(mappedDevice, passphrase), "failed to map disk")
	require.NoError(mapper.UnmapDisk(mappedDevice), "failed to remove disk mapping")

	assert.True(mapper.IsLUKSDevice())

	// Try to map disk with incorrect passphrase
	assert.Error(mapper.MapDisk(mappedDevice, "invalid-passphrase"), "was able to map disk with incorrect passphrase")
}

/*
type fakeMetadataAPI struct{}

func (f *fakeMetadataAPI) List(ctx context.Context) ([]metadata.InstanceMetadata, error) {
	return []metadata.InstanceMetadata{
		{
			Name:       "instanceName",
			ProviderID: "fake://instance-id",
			Role:       role.Unknown,
			VPCIP:      "192.0.2.1",
		},
	}, nil
}
*/
