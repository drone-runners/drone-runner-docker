// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package compiler

import (
	"context"
	"fmt"

	"github.com/drone-runners/drone-runner-docker/engine"
	"github.com/drone-runners/drone-runner-docker/engine/auth"
	"github.com/drone-runners/drone-runner-docker/engine/compiler/image"
	"github.com/drone-runners/drone-runner-docker/engine/resource"

	"github.com/drone/drone-go/drone"
	"github.com/drone/runner-go/clone"
	"github.com/drone/runner-go/environ"
	"github.com/drone/runner-go/manifest"
	"github.com/drone/runner-go/secret"

	"github.com/dchest/uniuri"
)

// random generator function
var random = uniuri.New

// Privileged provides a list of plugins that execute
// with privileged capabilities in order to run Docker
// in Docker.
var Privileged = []string{
	"plugins/docker",
	"plugins/acr",
	"plugins/ecr",
	"plugins/gcr",
	"plugins/heroku",
}

// Compiler compiles the Yaml configuration file to an
// intermediate representation optimized for simple execution.
type Compiler struct {
	// Manifest provides the parsed manifest.
	Manifest *manifest.Manifest

	// Pipeline provides the parsed pipeline. This pipeline is
	// the compiler source and is converted to the intermediate
	// representation by the Compile method.
	Pipeline *resource.Pipeline

	// Build provides the compiler with stage information that
	// is converted to environment variable format and passed to
	// each pipeline step. It is also used to clone the commit.
	Build *drone.Build

	// Stage provides the compiler with stage information that
	// is converted to environment variable format and passed to
	// each pipeline step.
	Stage *drone.Stage

	// Repo provides the compiler with repo information. This
	// repo information is converted to environment variable
	// format and passed to each pipeline step. It is also used
	// to clone the repository.
	Repo *drone.Repo

	// System provides the compiler with system information that
	// is converted to environment variable format and passed to
	// each pipeline step.
	System *drone.System

	// Environ provides a set of environment variables that
	// should be added to each pipeline step by default.
	Environ map[string]string

	// Labels provides a set of labels that should be added
	// to each container by default.
	Labels map[string]string

	// Privileged provides a list of docker images that
	// are always privileged.
	Privileged []string

	// Networks provides a set of networks that should be
	// attached to each pipeline container.
	Networks []string

	// Volumes provides a set of volumes that should be
	// mounted to each pipeline container.
	Volumes map[string]string

	// Netrc provides netrc parameters that can be used by the
	// default clone step to authenticate to the remote
	// repository.
	Netrc *drone.Netrc

	// Secret returns a named secret value that can be injected
	// into the pipeline step.
	Secret secret.Provider
}

// Compile compiles the configuration file.
func (c *Compiler) Compile(ctx context.Context) *engine.Spec {
	os := c.Pipeline.Platform.OS

	// create the workspace paths
	base, path, full := createWorkspace(c.Pipeline)

	// create the workspace mount
	mount := &engine.VolumeMount{
		Name: "_workspace",
		Path: base,
	}

	// create the workspace volume
	volume := &engine.VolumeEmptyDir{
		ID:   random(),
		Name: mount.Name,
		Labels: environ.Combine(
			c.Labels,
			createLabels(c.Repo, c.Build, c.Stage),
		),
	}

	spec := &engine.Spec{
		Network: engine.Network{
			ID: random(),
			Labels: environ.Combine(
				c.Labels,
				createLabels(c.Repo, c.Build, c.Stage),
			),
		},
		Platform: engine.Platform{
			OS:      c.Pipeline.Platform.OS,
			Arch:    c.Pipeline.Platform.Arch,
			Variant: c.Pipeline.Platform.Variant,
			Version: c.Pipeline.Platform.Version,
		},
		Volumes: []*engine.Volume{
			{EmptyDir: volume},
		},
	}

	// create the default environment variables.
	envs := environ.Combine(
		c.Environ,
		c.Build.Params,
		c.Pipeline.Environment,
		environ.Proxy(),
		environ.System(c.System),
		environ.Repo(c.Repo),
		environ.Build(c.Build),
		environ.Stage(c.Stage),
		environ.Link(c.Repo, c.Build, c.System),
		clone.Environ(clone.Config{
			SkipVerify: c.Pipeline.Clone.SkipVerify,
			Trace:      c.Pipeline.Clone.Trace,
			User: clone.User{
				Name:  c.Build.AuthorName,
				Email: c.Build.AuthorEmail,
			},
		}),
	)

	// create docker reference variables
	envs["DRONE_DOCKER_VOLUME_ID"] = volume.ID
	envs["DRONE_DOCKER_NETWORK_ID"] = spec.Network.ID

	// create the workspace variables
	envs["DRONE_WORKSPACE"] = full
	envs["DRONE_WORKSPACE_BASE"] = base
	envs["DRONE_WORKSPACE_PATH"] = path

	// create the netrc environment variables
	if c.Netrc != nil && c.Netrc.Machine != "" {
		envs["DRONE_NETRC_MACHINE"] = c.Netrc.Machine
		envs["DRONE_NETRC_USERNAME"] = c.Netrc.Login
		envs["DRONE_NETRC_PASSWORD"] = c.Netrc.Password
		envs["DRONE_NETRC_FILE"] = fmt.Sprintf(
			"machine %s login %s password %s",
			c.Netrc.Machine,
			c.Netrc.Login,
			c.Netrc.Password,
		)
	}

	match := manifest.Match{
		Action:   c.Build.Action,
		Cron:     c.Build.Cron,
		Ref:      c.Build.Ref,
		Repo:     c.Repo.Slug,
		Instance: c.System.Host,
		Target:   c.Build.Deploy,
		Event:    c.Build.Event,
		Branch:   c.Build.Target,
	}

	// create the clone step
	if c.Pipeline.Clone.Disable == false {
		step := createClone(c.Pipeline)
		step.ID = random()
		step.Envs = environ.Combine(envs, step.Envs)
		step.WorkingDir = full
		step.Labels = environ.Combine(
			c.Labels,
			createLabels(c.Repo, c.Build, c.Stage),
		)
		step.Volumes = append(step.Volumes, mount)
		spec.Steps = append(spec.Steps, step)
	}

	// create steps
	for _, src := range c.Pipeline.Services {
		dst := createStep(c.Pipeline, src)
		dst.Detach = true
		dst.Envs = environ.Combine(envs, dst.Envs)
		dst.Volumes = append(dst.Volumes, mount)
		dst.Labels = environ.Combine(
			c.Labels,
			createLabels(c.Repo, c.Build, c.Stage),
		)
		setupScript(src, dst, os)
		setupWorkdir(src, dst, full)
		spec.Steps = append(spec.Steps, dst)

		// if the pipeline step has unmet conditions the step is
		// automatically skipped.
		if !src.When.Match(match) {
			dst.RunPolicy = engine.RunNever
		}
	}

	// create steps
	for _, src := range c.Pipeline.Steps {
		dst := createStep(c.Pipeline, src)
		dst.Envs = environ.Combine(envs, dst.Envs)
		dst.Volumes = append(dst.Volumes, mount)
		dst.Labels = environ.Combine(
			c.Labels,
			createLabels(c.Repo, c.Build, c.Stage),
		)
		setupScript(src, dst, full)
		setupWorkdir(src, dst, full)
		spec.Steps = append(spec.Steps, dst)

		// if the pipeline step has unmet conditions the step is
		// automatically skipped.
		if !src.When.Match(match) {
			dst.RunPolicy = engine.RunNever
		}

		// if the pipeline step has an approved image, it is
		// automatically defaulted to run with escalalated
		// privileges.
		if c.isPrivileged(src) {
			dst.Privileged = true
		}
	}

	if isGraph(spec) == false {
		configureSerial(spec)
	} else if c.Pipeline.Clone.Disable == false {
		configureCloneDeps(spec)
	} else if c.Pipeline.Clone.Disable == true {
		removeCloneDeps(spec)
	}

	for _, step := range spec.Steps {
		for _, s := range step.Secrets {
			secret, ok := c.findSecret(ctx, s.Name)
			if ok {
				s.Data = []byte(secret)
			}
		}
	}

	var auths []*engine.Auth
	for _, name := range c.Pipeline.PullSecrets {
		secret, ok := c.findSecret(ctx, name)
		if ok {
			parsed, err := auth.ParseString(secret)
			if err == nil {
				auths = append(auths, parsed...)
			}
		}
	}

	for _, step := range spec.Steps {
	STEPS:
		for _, auth := range auths {
			if image.MatchHostname(step.Image, auth.Address) {
				step.Auth = auth
				break STEPS
			}
		}
	}

	// append global networks to the steps.
	for _, step := range spec.Steps {
		step.Networks = append(step.Networks, c.Networks...)
	}

	// append global volumes to the steps.
	for k, v := range c.Volumes {
		id := random()
		volume := &engine.Volume{
			HostPath: &engine.VolumeHostPath{
				ID:   id,
				Name: id,
				Path: k,
			},
		}
		spec.Volumes = append(spec.Volumes, volume)
		for _, step := range spec.Steps {
			mount := &engine.VolumeMount{
				Name: id,
				Path: v,
			}
			step.Volumes = append(step.Volumes, mount)
		}
	}

	return spec
}

func (c *Compiler) isPrivileged(step *resource.Step) bool {
	// privileged-by-default containers are only
	// enabled for plugins steps that do not define
	// commands, command, or entrypoint.
	if len(step.Commands) > 0 {
		return false
	}
	if len(step.Command) > 0 {
		return false
	}
	if len(step.Entrypoint) > 0 {
		return false
	}
	// if the container image matches any image
	// in the whitelist, return true.
	for _, img := range c.Privileged {
		a := img
		b := step.Image
		if image.Match(a, b) {
			return true
		}
	}
	return false
}

// helper function attempts to find and return the named secret.
// from the secret provider.
func (c *Compiler) findSecret(ctx context.Context, name string) (s string, ok bool) {
	if name == "" {
		return
	}
	found, _ := c.Secret.Find(ctx, &secret.Request{
		Name:  name,
		Build: c.Build,
		Repo:  c.Repo,
		Conf:  c.Manifest,
	})
	if found == nil {
		return
	}
	return found.Data, true
}
