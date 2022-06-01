//go:build integration

package integration

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/coordinator/core"
	"github.com/edgelesssys/constellation/internal/atls"
	"github.com/edgelesssys/constellation/state/keyservice"
	"github.com/edgelesssys/constellation/state/keyservice/keyproto"
	"github.com/edgelesssys/constellation/state/mapper"
	"github.com/martinjungblut/go-cryptsetup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	devicePath   = "testDevice"
	mappedDevice = "mappedDevice"
)

func setup() error {
	return exec.Command("/bin/dd", "if=/dev/zero", fmt.Sprintf("of=%s", devicePath), "bs=64M", "count=1").Run()
}

func teardown() error {
	return exec.Command("/bin/rm", "-f", devicePath).Run()
}

func TestMain(m *testing.M) {
	if os.Getuid() != 0 {
		fmt.Printf("This test suite requires root privileges, as libcrypsetup uses the kernel's device mapper.\n")
		os.Exit(1)
	}

	result := m.Run()
	os.Exit(result)
}

func TestMapper(t *testing.T) {
	cryptsetup.SetDebugLevel(cryptsetup.CRYPT_LOG_VERBOSE)
	cryptsetup.SetLogCallback(func(level int, message string) { fmt.Println(message) })
	assert := assert.New(t)
	require := require.New(t)
	require.NoError(setup(), "failed to setup test disk")
	defer func() { require.NoError(teardown(), "failed to delete test disk") }()

	mapper, err := mapper.New(devicePath)
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

	// get a free port on localhost to run the test on
	listener, err := net.Listen("tcp", "localhost:0")
	require.NoError(err)
	apiAddr := listener.Addr().String()
	listener.Close()

	api := keyservice.New(core.NewMockIssuer(), &core.ProviderMetadataFake{}, 20*time.Second)

	// send a key to the server
	go func() {
		// wait 2 seconds before sending the key
		time.Sleep(2 * time.Second)

		clientCfg, err := atls.CreateAttestationClientTLSConfig(nil, nil)
		require.NoError(err)
		conn, err := grpc.Dial(apiAddr, grpc.WithTransportCredentials(credentials.NewTLS(clientCfg)))
		require.NoError(err)
		defer conn.Close()

		client := keyproto.NewAPIClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		_, err = client.PushStateDiskKey(ctx, &keyproto.PushStateDiskKeyRequest{
			StateDiskKey: testKey,
		})
		require.NoError(err)
	}()

	key, err := api.WaitForDecryptionKey("12345678-1234-1234-1234-123456789ABC", apiAddr)
	assert.NoError(err)
	assert.Equal(testKey, key)
}
