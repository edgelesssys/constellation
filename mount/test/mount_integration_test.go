//go:build integration

package integration

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/edgelesssys/constellation/mount/cryptmapper"
	"github.com/edgelesssys/constellation/mount/kms"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/klog"
)

const (
	DevicePath string = "testDevice"
	DeviceName string = "testDeviceName"
)

func setup() {
	exec.Command("/bin/dd", "if=/dev/zero", fmt.Sprintf("of=%s", DevicePath), "bs=64M", "count=1").Run()
}

func teardown() {
	exec.Command("/bin/rm", "-f", DevicePath).Run()
}

func TestMain(m *testing.M) {
	if os.Getuid() != 0 {
		fmt.Printf("This test suite requires root privileges, as libcrypsetup uses the kernel's device mapper.\n")
		os.Exit(1)
	}

	klog.InitFlags(nil)
	defer klog.Flush()

	result := m.Run()
	os.Exit(result)
}

func TestOpenAndClose(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	setup()
	defer teardown()

	kms := kms.NewStaticKMS()
	mapper := cryptmapper.New(kms, &cryptmapper.CryptDevice{})

	newPath, err := mapper.OpenCryptDevice(context.Background(), DevicePath, DeviceName, false)
	require.NoError(err)
	assert.Equal(newPath, "/dev/mapper/"+DeviceName)

	// assert crypt device got created
	_, err = os.Stat(newPath)
	assert.NoError(err)
	// assert no integrity device got created
	_, err = os.Stat(newPath + "_dif")
	assert.True(os.IsNotExist(err))

	assert.NoError(mapper.CloseCryptDevice(DeviceName))

	// assert crypt device got removed
	_, err = os.Stat(newPath)
	assert.True(os.IsNotExist(err))
}

func TestOpenAndCloseIntegrity(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	setup()
	defer teardown()

	kms := kms.NewStaticKMS()
	mapper := cryptmapper.New(kms, &cryptmapper.CryptDevice{})

	newPath, err := mapper.OpenCryptDevice(context.Background(), DevicePath, DeviceName, true)
	require.NoError(err)
	assert.Equal(newPath, "/dev/mapper/"+DeviceName)

	// assert crypt device got created
	_, err = os.Stat(newPath)
	assert.NoError(err)
	// assert integrity device got created
	_, err = os.Stat(newPath + "_dif")
	assert.NoError(err)

	assert.NoError(mapper.CloseCryptDevice(DeviceName))

	// assert crypt device got removed
	_, err = os.Stat(newPath)
	assert.True(os.IsNotExist(err))
	// assert integrity device got removed
	_, err = os.Stat(newPath + "_dif")
	assert.True(os.IsNotExist(err))
}
