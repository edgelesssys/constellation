//go:build integration

package store

import (
	"context"
	"io"
	"net"
	"os"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/require"
)

const (
	etcdImageName = "bitnami/etcd:3.5.2"
)

func TestEtcdStore(t *testing.T) {
	require := require.New(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	require.NoError(err)
	defer dockerClient.Close()

	pullReader, err := dockerClient.ImagePull(ctx, etcdImageName, types.ImagePullOptions{})
	require.NoError(err)
	_, err = io.Copy(os.Stdout, pullReader)
	require.NoError(err)
	require.NoError(pullReader.Close())

	etcdHostConfig := &container.HostConfig{AutoRemove: true}
	etcdContainerConfig := &container.Config{
		Image: etcdImageName,
		Env: []string{
			"ETCD_LISTEN_CLIENT_URLS=http://0.0.0.0:2379",
			"ETCD_ADVERTISE_CLIENT_URLS=http://127.0.0.1:2379",
			"ETCD_LOG_LEVEL=debug",
			"ETCD_DATA_DIR=/bitnami/etcd/data",
		},
		Entrypoint:   []string{"/opt/bitnami/etcd/bin/etcd"},
		AttachStdout: true, // necessary to attach to the container log
		AttachStderr: true, // necessary to attach to the container log
		Tty:          true, // necessary to attach to the container log
	}

	t.Log("create etcd container...")
	createResp, err := dockerClient.ContainerCreate(ctx, etcdContainerConfig, etcdHostConfig, nil, nil, "etcd-storage-unittest")
	require.NoError(err)
	require.NoError(dockerClient.ContainerStart(ctx, createResp.ID, types.ContainerStartOptions{}))

	logReader, err := dockerClient.ContainerLogs(ctx, createResp.ID, types.ContainerLogsOptions{ShowStdout: true, Follow: true})
	require.NoError(err)
	go io.Copy(os.Stdout, logReader)

	containerData, err := dockerClient.ContainerInspect(ctx, createResp.ID)
	require.NoError(err)
	t.Logf("etcd Docker IP-Addr %v", containerData.NetworkSettings.IPAddress)

	//
	// Run the store test.
	//
	store, err := NewEtcdStore(net.JoinHostPort(containerData.NetworkSettings.IPAddress, "2379"), false, nil)
	require.NoError(err)
	defer store.Close()

	// TODO: since the etcd store does network, it should be canceled with a timeout.
	testStore(t, func() (Store, error) {
		clearStore(require, store)
		return store, nil
	})

	// Usually call it with a defer statement. However this causes problems with the construct above
	require.NoError(dockerClient.ContainerStop(ctx, createResp.ID, nil))
}

func clearStore(require *require.Assertions, store Store) {
	iter, err := store.Iterator("")
	require.NoError(err)
	for iter.HasNext() {
		key, err := iter.GetNext()
		require.NoError(err)
		err = store.Delete(key)
		require.NoError(err)
	}
}
