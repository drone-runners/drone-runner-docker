// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package compiler

import (
	"context"
	"fmt"
	"strings"

	"github.com/drone-runners/drone-runner-docker/engine"
	"github.com/drone-runners/drone-runner-docker/engine/resource"
	"github.com/drone-runners/drone-runner-docker/internal/docker/image"

	"github.com/drone/runner-go/clone"
	"github.com/drone/runner-go/environ"
	"github.com/drone/runner-go/environ/provider"
	"github.com/drone/runner-go/labels"
	"github.com/drone/runner-go/manifest"
	"github.com/drone/runner-go/pipeline/runtime"
	"github.com/drone/runner-go/registry"
	"github.com/drone/runner-go/registry/auths"
	"github.com/drone/runner-go/secret"

	"github.com/dchest/uniuri"
)

// random generator function
var random = func() string {
	return "drone-" + uniuri.NewLen(20)
}

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

// Resources defines container resource constraints. These
// constraints are per-container, not per-pipeline.
type Resources struct {
	Memory     int64
	MemorySwap int64
	CPUQuota   int64
	CPUPeriod  int64
	CPUShares  int64
	CPUSet     []string
	ShmSize    int64
}

// Compiler compiles the Yaml configuration file to an
// intermediate representation optimized for simple execution.
type Compiler struct {
	// Environ provides a set of environment variables that
	// should be added to each pipeline step by default.
	Environ provider.Provider

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

	// Clone overrides the default plugin image used
	// when cloning a repository.
	Clone string

	// Resources provides global resource constraints
	// applies to pipeline containers.
	Resources Resources

	// Secret returns a named secret value that can be injected
	// into the pipeline step.
	Secret secret.Provider

	// Registry returns a list of registry credentials that can be
	// used to pull private container images.
	Registry registry.Provider

	// Mount is an optional field that overrides the default
	// workspace volume and mounts to the host path
	Mount string
}

// Compile compiles the configuration file.
func (c *Compiler) Compile(ctx context.Context, args runtime.CompilerArgs) runtime.Spec {
	pipeline := args.Pipeline.(*resource.Pipeline)
	os := pipeline.Platform.OS

	// create the workspace paths
	base, path, full := createWorkspace(pipeline)

	// if the source code is mounted from the host, the
	// target mount path inside the container must be the
	// full workspace path.
	if c.Mount != "" {
		base = full
		path = ""
	}

	// create system labels
	labels := labels.Combine(
		c.Labels,
		labels.FromRepo(args.Repo),
		labels.FromBuild(args.Build),
		labels.FromStage(args.Stage),
		labels.FromSystem(args.System),
		labels.WithTimeout(args.Repo),
	)

	// create the workspace mount
	mount := &engine.VolumeMount{
		Name: "_workspace",
		Path: base,
	}

	// create the workspace volume
	volume := &engine.Volume{
		EmptyDir: &engine.VolumeEmptyDir{
			ID:     random(),
			Name:   mount.Name,
			Labels: labels,
		},
	}

	// if the repository is mounted from a local volume,
	// we should replace the data volume with a host machine
	// volume declaration.
	if c.Mount != "" {
		volume.EmptyDir = nil
		volume.HostPath = &engine.VolumeHostPath{
			ID:     random(),
			Name:   mount.Name,
			Path:   c.Mount,
			Labels: labels,
		}
	}

	spec := &engine.Spec{
		Network: engine.Network{
			ID:     random(),
			Labels: labels,
		},
		Platform: engine.Platform{
			OS:      pipeline.Platform.OS,
			Arch:    pipeline.Platform.Arch,
			Variant: pipeline.Platform.Variant,
			Version: pipeline.Platform.Version,
		},
		Volumes: []*engine.Volume{volume},
	}

	// list the global environment variables
	globals, _ := c.Environ.List(ctx, &provider.Request{
		Build: args.Build,
		Repo:  args.Repo,
	})

	// create the default environment variables.
	envs := environ.Combine(
		provider.ToMap(
			provider.FilterUnmasked(globals),
		),
		args.Build.Params,
		pipeline.Environment,
		environ.Proxy(),
		environ.System(args.System),
		environ.Repo(args.Repo),
		environ.Build(args.Build),
		environ.Stage(args.Stage),
		environ.Link(args.Repo, args.Build, args.System),
		clone.Environ(clone.Config{
			SkipVerify: pipeline.Clone.SkipVerify,
			Trace:      pipeline.Clone.Trace,
			User: clone.User{
				Name:  args.Build.AuthorName,
				Email: args.Build.AuthorEmail,
			},
		}),
	)

	// create network reference variables
	envs["DRONE_DOCKER_NETWORK_ID"] = spec.Network.ID

	// create the workspace variables
	envs["DRONE_WORKSPACE"] = full
	envs["DRONE_WORKSPACE_BASE"] = base
	envs["DRONE_WORKSPACE_PATH"] = path

	// create volume reference variables
	if volume.EmptyDir != nil {
		envs["DRONE_DOCKER_VOLUME_ID"] = volume.EmptyDir.ID
	} else {
		envs["DRONE_DOCKER_VOLUME_PATH"] = volume.HostPath.Path
	}

	// create the netrc environment variables
	if args.Netrc != nil && args.Netrc.Machine != "" {
		envs["DRONE_NETRC_MACHINE"] = args.Netrc.Machine
		envs["DRONE_NETRC_USERNAME"] = args.Netrc.Login
		envs["DRONE_NETRC_PASSWORD"] = args.Netrc.Password
		envs["DRONE_NETRC_FILE"] = fmt.Sprintf(
			"machine %s login %s password %s",
			args.Netrc.Machine,
			args.Netrc.Login,
			args.Netrc.Password,
		)
	}

	match := manifest.Match{
		Action:   args.Build.Action,
		Cron:     args.Build.Cron,
		Ref:      args.Build.Ref,
		Repo:     args.Repo.Slug,
		Instance: args.System.Host,
		Target:   args.Build.Deploy,
		Event:    args.Build.Event,
		Branch:   args.Build.Target,
	}

	// create the clone step
	if pipeline.Clone.Disable == false {
		step := createClone(pipeline)
		step.ID = random()
		step.Envs = environ.Combine(envs, step.Envs)
		step.WorkingDir = full
		step.Labels = labels
		step.Pull = engine.PullIfNotExists
		step.Volumes = append(step.Volumes, mount)
		spec.Steps = append(spec.Steps, step)

		// if the clone image is customized, override
		// the default image.
		if c.Clone != "" {
			step.Image = c.Clone
		}

		// if the repository is mounted from a local
		// volume we should disable cloning.
		if c.Mount != "" {
			step.RunPolicy = runtime.RunNever
		}
	}

	// create steps
	for _, src := range pipeline.Services {
		dst := createStep(pipeline, src)
		dst.Detach = true
		dst.Envs = environ.Combine(envs, dst.Envs)
		dst.Volumes = append(dst.Volumes, mount)
		dst.Labels = labels
		setupScript(src, dst, os)
		setupWorkdir(src, dst, full)
		spec.Steps = append(spec.Steps, dst)

		// if the pipeline step has unmet conditions the step is
		// automatically skipped.
		if !src.When.Match(match) {
			dst.RunPolicy = runtime.RunNever
		}

		if c.isPrivileged(src) {
			dst.Privileged = true
		}
	}

	// create steps
	for _, src := range pipeline.Steps {
		dst := createStep(pipeline, src)
		dst.Envs = environ.Combine(envs, dst.Envs)
		dst.Volumes = append(dst.Volumes, mount)
		dst.Labels = labels
		setupScript(src, dst, os)
		setupWorkdir(src, dst, full)
		spec.Steps = append(spec.Steps, dst)

		// if the pipeline step has unmet conditions the step is
		// automatically skipped.
		if !src.When.Match(match) {
			dst.RunPolicy = runtime.RunNever
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
	} else if pipeline.Clone.Disable == false {
		configureCloneDeps(spec)
	} else if pipeline.Clone.Disable == true {
		removeCloneDeps(spec)
	}

	for _, step := range spec.Steps {
		for _, s := range step.Secrets {
			secret, ok := c.findSecret(ctx, args, s.Name)
			if ok {
				s.Data = []byte(secret)
			}
		}
	}

	// get registry credentials from registry plugins
	creds, err := c.Registry.List(ctx, &registry.Request{
		Repo:  args.Repo,
		Build: args.Build,
	})
	if err != nil {
		// TODO (bradrydzewski) return an error to the caller
		// if the provider returns an error.
	}

	// get registry credentials from secrets
	for _, name := range pipeline.PullSecrets {
		secret, ok := c.findSecret(ctx, args, name)
		if ok {
			parsed, err := auths.ParseString(secret)
			if err == nil {
				creds = append(parsed, creds...)
			}
		}
	}

	for _, step := range spec.Steps {
	STEPS:
		for _, cred := range creds {
			if image.MatchHostname(step.Image, cred.Address) {
				step.Auth = &engine.Auth{
					Address:  cred.Address,
					Username: cred.Username,
					Password: cred.Password,
				}
				break STEPS
			}
		}
	}

	// HACK: append masked global variables to secrets
	// this ensures the environment variable values are
	// masked when printed to the console.
	masked := provider.FilterMasked(globals)
	for _, step := range spec.Steps {
		for _, g := range masked {
			step.Secrets = append(step.Secrets, &engine.Secret{
				Name: g.Name,
				Data: []byte(g.Data),
				Mask: g.Mask,
				Env:  g.Name,
			})
		}
	}

	// append global resource limits to steps
	for _, step := range spec.Steps {
		// the resource limits defined in the yaml currently
		// take precedence over global values. This is something
		// we should re-think in a future release.
		if step.MemSwapLimit == 0 {
			step.MemSwapLimit = c.Resources.MemorySwap
		}
		if step.MemLimit == 0 {
			step.MemLimit = c.Resources.Memory
		}
		if step.ShmSize == 0 {
			step.ShmSize = c.Resources.ShmSize
		}
		step.CPUPeriod = c.Resources.CPUPeriod
		step.CPUQuota = c.Resources.CPUQuota
		step.CPUShares = c.Resources.CPUShares
		step.CPUSet = c.Resources.CPUSet
	}

	// append global networks to the steps.
	for _, step := range spec.Steps {
		step.Networks = append(step.Networks, c.Networks...)
	}

	// append global volumes to the steps.
	for k, v := range c.Volumes {
		id := random()
		ro := strings.HasSuffix(v, ":ro")
		v = strings.TrimSuffix(v, ":ro")
		volume := &engine.Volume{
			HostPath: &engine.VolumeHostPath{
				ID:       id,
				Name:     id,
				Path:     k,
				ReadOnly: ro,
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

	// append volumes
	for _, v := range pipeline.Volumes {
		id := random()
		src := new(engine.Volume)
		if v.EmptyDir != nil {
			src.EmptyDir = &engine.VolumeEmptyDir{
				ID:        id,
				Name:      v.Name,
				Medium:    v.EmptyDir.Medium,
				SizeLimit: int64(v.EmptyDir.SizeLimit),
				Labels:    labels,
			}
		} else if v.HostPath != nil {
			src.HostPath = &engine.VolumeHostPath{
				ID:   id,
				Name: v.Name,
				Path: v.HostPath.Path,
			}
		} else {
			continue
		}
		spec.Volumes = append(spec.Volumes, src)
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
func (c *Compiler) findSecret(ctx context.Context, args runtime.CompilerArgs, name string) (s string, ok bool) {
	if name == "" {
		return
	}

	// source secrets from the global secret provider
	// and the repository secret provider.
	provider := secret.Combine(
		args.Secret,
		c.Secret,
	)

	// TODO return an error to the caller if the provider
	// returns an error.
	found, _ := provider.Find(ctx, &secret.Request{
		Name:  name,
		Build: args.Build,
		Repo:  args.Repo,
		Conf:  args.Manifest,
	})
	if found == nil {
		return
	}
	return found.Data, true
}
