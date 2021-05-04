// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package engine

import (
	"context"
	goerrors "errors"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"time"

	"github.com/drone-runners/drone-runner-docker/internal/docker/errors"
	"github.com/drone-runners/drone-runner-docker/internal/docker/image"
	"github.com/drone-runners/drone-runner-docker/internal/docker/jsonmessage"
	"github.com/drone-runners/drone-runner-docker/internal/docker/stdcopy"
	"github.com/drone/runner-go/logger"
	"github.com/drone/runner-go/pipeline/runtime"
	"github.com/drone/runner-go/registry/auths"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
)

// Opts configures the Docker engine.
type Opts struct {
	HidePull bool
}

// Docker implements a Docker pipeline engine.
type Docker struct {
	client   client.APIClient
	hidePull bool
}

// New returns a new engine.
func New(client client.APIClient, opts Opts) *Docker {
	return &Docker{
		client:   client,
		hidePull: opts.HidePull,
	}
}

// NewEnv returns a new Engine from the environment.
func NewEnv(opts Opts) (*Docker, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}
	return New(cli, opts), nil
}

// Ping pings the Docker daemon.
func (e *Docker) Ping(ctx context.Context) error {
	_, err := e.client.Ping(ctx)
	return err
}

// Setup the pipeline environment.
func (e *Docker) Setup(ctx context.Context, specv runtime.Spec) error {
	spec := specv.(*Spec)

	// creates the default temporary (local) volumes
	// that are mounted into each container step.
	for _, vol := range spec.Volumes {
		if vol.EmptyDir == nil {
			continue
		}
		_, err := e.client.VolumeCreate(ctx, volume.VolumeCreateBody{
			Name:   vol.EmptyDir.ID,
			Driver: "local",
			Labels: vol.EmptyDir.Labels,
		})
		if err != nil {
			return errors.TrimExtraInfo(err)
		}
	}

	// creates the default pod network. All containers
	// defined in the pipeline are attached to this network.
	driver := "bridge"
	if spec.Platform.OS == "windows" {
		driver = "nat"
	}
	_, err := e.client.NetworkCreate(ctx, spec.Network.ID, types.NetworkCreate{
		Driver:  driver,
		Options: spec.Network.Options,
		Labels:  spec.Network.Labels,
	})

	// launches the inernal setup steps
	for _, step := range spec.Internal {
		if err := e.create(ctx, spec, step, ioutil.Discard); err != nil {
			logger.FromContext(ctx).
				WithError(err).
				WithField("container", step.ID).
				Errorln("cannot create tmate container")
			return err
		}
		if err := e.start(ctx, step.ID); err != nil {
			logger.FromContext(ctx).
				WithError(err).
				WithField("container", step.ID).
				Errorln("cannot start tmate container")
			return err
		}
		if !step.Detach {
			// the internal containers perform short-lived tasks
			// and should not require > 1 minute to execute.
			//
			// just to be on the safe side we apply a timeout to
			// ensure we never block pipeline execution because we
			// are waiting on an internal task.
			ctx, cancel := context.WithTimeout(ctx, time.Minute)
			defer cancel()
			e.wait(ctx, step.ID)
		}
	}

	return errors.TrimExtraInfo(err)
}

// Destroy the pipeline environment.
func (e *Docker) Destroy(ctx context.Context, specv runtime.Spec) error {
	spec := specv.(*Spec)

	removeOpts := types.ContainerRemoveOptions{
		Force:         true,
		RemoveLinks:   false,
		RemoveVolumes: true,
	}

	// stop all containers
	for _, step := range append(spec.Steps, spec.Internal...) {
		e.client.ContainerKill(ctx, step.ID, "9")
	}

	// cleanup all containers
	for _, step := range append(spec.Steps, spec.Internal...) {
		e.client.ContainerRemove(ctx, step.ID, removeOpts)
	}

	// cleanup all volumes
	for _, vol := range spec.Volumes {
		if vol.EmptyDir == nil {
			continue
		}
		// tempfs volumes do not have a volume entry,
		// and therefore do not require removal.
		if vol.EmptyDir.Medium == "memory" {
			continue
		}
		e.client.VolumeRemove(ctx, vol.EmptyDir.ID, true)
	}

	// cleanup the network
	e.client.NetworkRemove(ctx, spec.Network.ID)

	// notice that we never collect or return any errors.
	// this is because we silently ignore cleanup failures
	// and instead ask the system admin to periodically run
	// `docker prune` commands.
	return nil
}

// Run runs the pipeline step.
func (e *Docker) Run(ctx context.Context, specv runtime.Spec, stepv runtime.Step, output io.Writer) (*runtime.State, error) {
	spec := specv.(*Spec)
	step := stepv.(*Step)

	// create the container
	err := e.create(ctx, spec, step, output)
	if err != nil {
		return nil, errors.TrimExtraInfo(err)
	}
	// start the container
	err = e.start(ctx, step.ID)
	if err != nil {
		return nil, errors.TrimExtraInfo(err)
	}
	// tail the container
	err = e.tail(ctx, step.ID, output)
	if err != nil {
		return nil, errors.TrimExtraInfo(err)
	}
	// wait for the response
	return e.waitRetry(ctx, step.ID)
}

//
// emulate docker commands
//

func (e *Docker) create(ctx context.Context, spec *Spec, step *Step, output io.Writer) error {
	// create pull options with encoded authorization credentials.
	pullopts := types.ImagePullOptions{}
	if step.Auth != nil {
		pullopts.RegistryAuth = auths.Header(
			step.Auth.Username,
			step.Auth.Password,
		)
	}

	// automatically pull the latest version of the image if requested
	// by the process configuration, or if the image is :latest
	if step.Pull == PullAlways ||
		(step.Pull == PullDefault && image.IsLatest(step.Image)) {
		rc, pullerr := e.client.ImagePull(ctx, step.Image, pullopts)
		if pullerr == nil {
			if e.hidePull {
				io.Copy(ioutil.Discard, rc)
			} else {
				jsonmessage.Copy(rc, output)
			}
			rc.Close()
		}
		if pullerr != nil {
			return pullerr
		}
	}

	var secretErrors []string
	for _, secret := range step.Secrets {
		if len(secret.Data) == 0 {
			secretErrors = append(secretErrors, fmt.Sprintf("Could not find secret `%s` to be mounted on env variable `%s`. ", secret.Name, secret.Env))
		}
	}
	if len(secretErrors) > 0 {
		message := "some secrets could not be mounted. Check Drone's (and your external secret provider's) logs for more info. \n" + strings.Join(secretErrors, "\n")
		if spec.SecretsRequired {
			return goerrors.New(message)
		}
		logger.FromContext(ctx).Info(message)
	}

	_, err := e.client.ContainerCreate(ctx,
		toConfig(spec, step),
		toHostConfig(spec, step),
		toNetConfig(spec, step),
		step.ID,
	)

	// automatically pull and try to re-create the image if the
	// failure is caused because the image does not exist.
	if client.IsErrNotFound(err) && step.Pull != PullNever {
		rc, pullerr := e.client.ImagePull(ctx, step.Image, pullopts)
		if pullerr != nil {
			return pullerr
		}

		if e.hidePull {
			io.Copy(ioutil.Discard, rc)
		} else {
			jsonmessage.Copy(rc, output)
		}
		rc.Close()

		// once the image is successfully pulled we attempt to
		// re-create the container.
		_, err = e.client.ContainerCreate(ctx,
			toConfig(spec, step),
			toHostConfig(spec, step),
			toNetConfig(spec, step),
			step.ID,
		)
	}
	if err != nil {
		return err
	}

	// attach the container to user-defined networks.
	// primarily used to attach global user-defined networks.
	if step.Network == "" {
		for _, net := range step.Networks {
			err = e.client.NetworkConnect(ctx, net, step.ID, &network.EndpointSettings{
				Aliases: []string{net},
			})
			if err != nil {
				return nil
			}
		}
	}

	return nil
}

// helper function emulates the `docker start` command.
func (e *Docker) start(ctx context.Context, id string) error {
	return e.client.ContainerStart(ctx, id, types.ContainerStartOptions{})
}

// helper function emulates the `docker wait` command, blocking
// until the container stops and returning the exit code.
func (e *Docker) waitRetry(ctx context.Context, id string) (*runtime.State, error) {
	for {
		// if the context is canceled, meaning the
		// pipeline timed out or was killed by the
		// end-user, we should exit with an error.
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		state, err := e.wait(ctx, id)
		if err != nil {
			return nil, err
		}
		if state.Exited {
			return state, err
		}
		logger.FromContext(ctx).
			WithField("container", id).
			Trace("docker wait exited unexpectedly")
	}
}

// helper function emulates the `docker wait` command, blocking
// until the container stops and returning the exit code.
func (e *Docker) wait(ctx context.Context, id string) (*runtime.State, error) {
	wait, errc := e.client.ContainerWait(ctx, id, container.WaitConditionNotRunning)
	select {
	case <-wait:
	case <-errc:
	}

	info, err := e.client.ContainerInspect(ctx, id)
	if err != nil {
		return nil, err
	}

	return &runtime.State{
		Exited:    !info.State.Running,
		ExitCode:  info.State.ExitCode,
		OOMKilled: info.State.OOMKilled,
	}, nil
}

// helper function emulates the `docker logs -f` command, streaming
// all container logs until the container stops.
func (e *Docker) tail(ctx context.Context, id string, output io.Writer) error {
	opts := types.ContainerLogsOptions{
		Follow:     true,
		ShowStdout: true,
		ShowStderr: true,
		Details:    false,
		Timestamps: false,
	}

	logs, err := e.client.ContainerLogs(ctx, id, opts)
	if err != nil {
		return err
	}

	go func() {
		stdcopy.StdCopy(output, output, logs)
		logs.Close()
	}()
	return nil
}
