/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package libvirt is used to start and stop containerized libvirt instances.
package libvirt

import (
	"context"
	"errors"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	docker "github.com/docker/docker/client"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/spf13/afero"
)

// LibvirtTCPConnectURI is the default URI to connect to containerized libvirt.
// Non standard port to avoid conflict with host libvirt.
// Changes here should also be reflected in the Dockerfile in "cli/internal/libvirt/Dockerfile".
const LibvirtTCPConnectURI = "qemu+tcp://localhost:16599/system"

// Runner handles starting and stopping of containerized libvirt instances.
type Runner struct {
	nameFile string
	file     file.Handler
}

// New creates a new LibvirtRunner.
func New() *Runner {
	return &Runner{
		nameFile: "libvirt.name",
		file:     file.NewHandler(afero.NewOsFs()),
	}
}

// Start starts a containerized libvirt instance.
func (r *Runner) Start(ctx context.Context, name, imageName string) error {
	docker, err := docker.NewClientWithOpts(docker.FromEnv, docker.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	defer docker.Close()

	containerName := name + "-libvirt"

	// check if image exists locally, if not pull it
	// this allows us to use a custom image without having to push it to a registry
	images, err := docker.ImageList(ctx, types.ImageListOptions{
		Filters: filters.NewArgs(
			filters.KeyValuePair{
				Key:   "reference",
				Value: imageName,
			},
		),
	})
	if err != nil {
		return err
	}
	if len(images) == 0 {
		reader, err := docker.ImagePull(ctx, imageName, types.ImagePullOptions{})
		if err != nil {
			return err
		}
		defer reader.Close()
		if _, err := io.Copy(io.Discard, reader); err != nil {
			return err
		}
	}

	// create and start the libvirt container
	if _, err := docker.ContainerCreate(ctx,
		&container.Config{
			Image: imageName,
		},
		&container.HostConfig{
			NetworkMode: container.NetworkMode("host"),
			AutoRemove:  true,
			// container has to be "privileged" so libvirt has access to proc fs
			Privileged: true,
		},
		nil,
		nil,
		containerName,
	); err != nil {
		return err
	}
	if err := docker.ContainerStart(ctx, containerName, types.ContainerStartOptions{}); err != nil {
		_ = docker.ContainerRemove(ctx, containerName, types.ContainerRemoveOptions{Force: true})
		return err
	}

	// write the name of the container to a file so we can remove it later
	if err := r.file.Write(r.nameFile, []byte(containerName)); err != nil {
		_ = docker.ContainerRemove(ctx, containerName, types.ContainerRemoveOptions{Force: true})
		return err
	}

	return nil
}

// Stop stops a containerized libvirt instance.
func (r *Runner) Stop(ctx context.Context) error {
	name, err := r.file.Read(r.nameFile)
	if err != nil {
		if errors.Is(err, afero.ErrFileNotFound) {
			return nil
		}
		return err
	}

	docker, err := docker.NewClientWithOpts(docker.FromEnv, docker.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	defer docker.Close()
	if err := docker.ContainerRemove(ctx, string(name), types.ContainerRemoveOptions{Force: true}); err != nil {
		return err
	}

	if err := r.file.Remove(r.nameFile); err != nil && !errors.Is(err, afero.ErrFileNotFound) {
		return err
	}
	return nil
}
