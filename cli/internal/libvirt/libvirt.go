/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package libvirt

import (
	"context"
	"errors"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	docker "github.com/docker/docker/client"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/spf13/afero"
)

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
	if _, err := docker.ContainerCreate(ctx,
		&container.Config{
			Image: imageName,
		},
		&container.HostConfig{
			NetworkMode: container.NetworkMode("host"),
			Privileged:  true,
			AutoRemove:  true,
			Mounts: []mount.Mount{
				{
					Type:   mount.TypeBind,
					Source: "/dev/kvm",
					Target: "/dev/kvm",
				},
			},
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

	if err := r.file.Write(r.nameFile, []byte(containerName)); err != nil {
		_ = docker.ContainerRemove(ctx, containerName, types.ContainerRemoveOptions{Force: true})
		return err
	}

	time.Sleep(15 * time.Second)

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
