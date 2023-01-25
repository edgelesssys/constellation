//go:build integration

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package test

import (
	"context"
	"io"
	"os"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/edgelesssys/constellation/v2/internal/kms/storage"
	"github.com/edgelesssys/constellation/v2/internal/kms/storage/gcs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/option"
)

const storageEmulator = "gcr.io/cloud-devrel-public-resources/storage-testbench"

func TestGoogleCloudStorage(t *testing.T) {
	if !*runGcsStorageTestBench {
		t.Skip("Skipping GCS storage-testbench test")
	}
	assert := assert.New(t)
	require := require.New(t)

	containerCtx := context.Background()

	// Set up the Storage Emulator
	t.Log("Creating storage emulator...")
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	require.NoError(err)
	emulator, err := setupEmulator(containerCtx, cli, storageEmulator)
	require.NoError(err)
	defer func() { _ = cli.ContainerStop(containerCtx, emulator.ID, nil) }()

	// Run the actual test
	t.Setenv("STORAGE_EMULATOR_HOST", "localhost:9000")

	bucketName := "test-bucket"
	projectName := "test-project"

	t.Log("Running test...")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*50)
	defer cancel()
	store, err := gcs.New(ctx, projectName, bucketName, nil, option.WithoutAuthentication())
	require.NoError(err)

	testDEK1 := []byte("test DEK")
	testDEK2 := []byte("more test DEK")

	// request unset value
	_, err = store.Get(ctx, "test:input")
	assert.Error(err)

	// test Put method
	assert.NoError(store.Put(ctx, "volume01", testDEK1))
	assert.NoError(store.Put(ctx, "volume02", testDEK2))

	// make sure values have been set
	val, err := store.Get(ctx, "volume01")
	assert.NoError(err)
	assert.Equal(testDEK1, val)
	val, err = store.Get(ctx, "volume02")
	assert.NoError(err)
	assert.Equal(testDEK2, val)

	_, err = store.Get(ctx, "invalid:key")
	assert.Error(err)
	assert.ErrorIs(err, storage.ErrDEKUnset)
}

func setupEmulator(ctx context.Context, cli *client.Client, imageName string) (container.ContainerCreateCreatedBody, error) {
	reader, err := cli.ImagePull(ctx, imageName, types.ImagePullOptions{})
	if err != nil {
		return container.ContainerCreateCreatedBody{}, err
	}
	if _, err := io.Copy(os.Stdout, reader); err != nil {
		return container.ContainerCreateCreatedBody{}, err
	}
	if err := reader.Close(); err != nil {
		return container.ContainerCreateCreatedBody{}, err
	}

	// the 3 true statements are necessary to attach later to the container log
	containerConfig := &container.Config{
		Image:        storageEmulator,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
	}
	emulator, err := cli.ContainerCreate(ctx, containerConfig, &container.HostConfig{NetworkMode: container.NetworkMode("host"), AutoRemove: true}, nil, nil, "google-cloud-storage-test")
	if err != nil {
		return emulator, err
	}
	if err := cli.ContainerStart(ctx, emulator.ID, types.ContainerStartOptions{}); err != nil {
		return emulator, err
	}

	logs, err := cli.ContainerLogs(ctx, emulator.ID, types.ContainerLogsOptions{
		ShowStdout: true,
		Follow:     true,
	})
	if err != nil {
		return emulator, err
	}
	go func() { _, _ = io.Copy(os.Stdout, logs) }()
	return emulator, nil
}
