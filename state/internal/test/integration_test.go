//go:build integration

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package integration

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/internal/atls"
	"github.com/edgelesssys/constellation/internal/cloud/metadata"
	"github.com/edgelesssys/constellation/internal/grpc/atlscredentials"
	"github.com/edgelesssys/constellation/internal/logger"
	"github.com/edgelesssys/constellation/internal/oid"
	"github.com/edgelesssys/constellation/internal/role"
	"github.com/edgelesssys/constellation/state/internal/keyservice"
	"github.com/edgelesssys/constellation/state/internal/mapper"
	"github.com/edgelesssys/constellation/state/keyproto"
	"github.com/martinjungblut/go-cryptsetup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
	"google.golang.org/grpc"
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

	mapper, err := mapper.New(devicePath, logger.NewTest(t))
	require.NoError(err, "failed to initialize crypt device")
	defer func() { require.NoError(mapper.Close(), "failed to close crypt device") }()

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

func TestKeyAPI(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	testKey := []byte("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	testSecret := []byte("BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB")

	// get a free port on localhost to run the test on
	listener, err := net.Listen("tcp", "localhost:0")
	require.NoError(err)
	apiAddr := listener.Addr().String()
	listener.Close()

	api := keyservice.New(
		logger.NewTest(t),
		atls.NewFakeIssuer(oid.Dummy{}),
		&fakeMetadataAPI{},
		20*time.Second,
		time.Second,
	)

	// send a key to the server
	go func() {
		// wait 2 seconds before sending the key
		time.Sleep(2 * time.Second)

		creds := atlscredentials.New(nil, nil)
		conn, err := grpc.Dial(apiAddr, grpc.WithTransportCredentials(creds))
		require.NoError(err)
		defer conn.Close()

		client := keyproto.NewAPIClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		_, err = client.PushStateDiskKey(ctx, &keyproto.PushStateDiskKeyRequest{
			StateDiskKey:      testKey,
			MeasurementSecret: testSecret,
		})
		require.NoError(err)
	}()

	key, measurementSecret, err := api.WaitForDecryptionKey("12345678-1234-1234-1234-123456789ABC", apiAddr)
	assert.NoError(err)
	assert.Equal(testKey, key)
	assert.Equal(testSecret, measurementSecret)
}

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
