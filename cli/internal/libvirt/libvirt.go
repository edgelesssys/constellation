/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
Package libvirt is used to start and stop containerized libvirt instances.

The code in this package should be kept minimal, and likely won't need to be changed unless we do a major refactoring of our QEMU/libvirt installation.
*/
package libvirt

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
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
		return fmt.Errorf("failed to create docker client: %w", err)
	}
	defer docker.Close()

	// check for an existing container
	if containerName, err := r.file.Read(r.nameFile); err == nil {
		// check if a container with the same name already exists
		containers, err := docker.ContainerList(ctx, container.ListOptions{
			Filters: filters.NewArgs(
				filters.KeyValuePair{
					Key:   "name",
					Value: fmt.Sprintf("^%s$", containerName),
				},
			),
			All: true,
		})
		if err != nil {
			return err
		}
		if len(containers) > 1 {
			return fmt.Errorf("more than one container with name %q found", containerName)
		}

		// if a container with the same name exists,
		// check if it is using the correct image and if it is running
		if len(containers) == 1 {
			// make sure the container we listed is using the correct image
			if containers[0].Image != imageName {
				return fmt.Errorf("existing libvirt container %q is using a different image: expected %q, got %q", containerName, imageName, containers[0].Image)
			}

			// container already exists, check if its running
			if containers[0].State == "running" {
				// container is up, nothing to do
				return nil
			}
			// container exists but is not running, remove it
			// so we can start a new one
			if err := docker.ContainerRemove(ctx, containers[0].ID, container.RemoveOptions{Force: true}); err != nil {
				return err
			}
		}
	} else if !errors.Is(err, afero.ErrFileNotFound) {
		return err
	}

	return r.startNewContainer(ctx, docker, name+"-libvirt", imageName)
}

// startNewContainer starts a new libvirt container using the given image.
func (r *Runner) startNewContainer(ctx context.Context, docker *docker.Client, containerName, imageName string) error {
	// check if image exists locally, if not pull it
	// this allows us to use a custom image without having to push it to a registry
	images, err := docker.ImageList(ctx, image.ListOptions{
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
		reader, err := docker.ImagePull(ctx, imageName, image.PullOptions{})
		if err != nil {
			return fmt.Errorf("failed to pull image %q: %w", imageName, err)
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
		return fmt.Errorf("failed to create container: %w", err)
	}
	if err := docker.ContainerStart(ctx, containerName, container.StartOptions{}); err != nil {
		_ = docker.ContainerRemove(ctx, containerName, container.RemoveOptions{Force: true})
		return fmt.Errorf("failed to start container: %w", err)
	}

	// write the name of the container to a file so we can remove it later
	if err := r.file.Write(r.nameFile, []byte(containerName)); err != nil {
		_ = docker.ContainerRemove(ctx, containerName, container.RemoveOptions{Force: true})
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
	if err := docker.ContainerRemove(ctx, string(name), container.RemoveOptions{Force: true}); err != nil {
		return err
	}

	if err := r.file.Remove(r.nameFile); err != nil && !errors.Is(err, afero.ErrFileNotFound) {
		return err
	}
	return nil
}
