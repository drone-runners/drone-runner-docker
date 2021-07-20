// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package resource

import "github.com/drone/runner-go/manifest"

var (
	_ manifest.Resource          = (*Pipeline)(nil)
	_ manifest.TriggeredResource = (*Pipeline)(nil)
	_ manifest.DependantResource = (*Pipeline)(nil)
	_ manifest.PlatformResource  = (*Pipeline)(nil)
)

// Defines the Resource Kind and Type.
const (
	Kind = "pipeline"
	Type = "docker"
)

// Pipeline is a pipeline resource that executes pipelines
// on the host machine without any virtualization.
type Pipeline struct {
	Version string   `json:"version,omitempty"`
	Kind    string   `json:"kind,omitempty"`
	Type    string   `json:"type,omitempty"`
	Name    string   `json:"name,omitempty"`
	Deps    []string `json:"depends_on,omitempty" yaml:"depends_on"`

	Clone       manifest.Clone       `json:"clone,omitempty"`
	Concurrency manifest.Concurrency `json:"concurrency,omitempty"`
	Node        map[string]string    `json:"node,omitempty"`
	Platform    manifest.Platform    `json:"platform,omitempty"`
	Trigger     manifest.Conditions  `json:"conditions,omitempty"`

	Environment map[string]string `json:"environment,omitempty"`
	Services    []*Step           `json:"services,omitempty"`
	Steps       []*Step           `json:"steps,omitempty"`
	Volumes     []*Volume         `json:"volumes,omitempty"`
	PullSecrets []string          `json:"image_pull_secrets,omitempty" yaml:"image_pull_secrets"`
	Workspace   Workspace         `json:"workspace,omitempty"`
}

// GetVersion returns the resource version.
func (p *Pipeline) GetVersion() string { return p.Version }

// GetKind returns the resource kind.
func (p *Pipeline) GetKind() string { return p.Kind }

// GetType returns the resource type.
func (p *Pipeline) GetType() string { return p.Type }

// GetName returns the resource name.
func (p *Pipeline) GetName() string { return p.Name }

// GetDependsOn returns the resource dependencies.
func (p *Pipeline) GetDependsOn() []string { return p.Deps }

// GetTrigger returns the resource triggers.
func (p *Pipeline) GetTrigger() manifest.Conditions { return p.Trigger }

// GetNodes returns the resource node labels.
func (p *Pipeline) GetNodes() map[string]string { return p.Node }

// GetPlatform returns the resource platform.
func (p *Pipeline) GetPlatform() manifest.Platform { return p.Platform }

// GetConcurrency returns the resource concurrency limits.
func (p *Pipeline) GetConcurrency() manifest.Concurrency { return p.Concurrency }

// GetStep returns the named step. If no step exists with the
// given name, a nil value is returned.
func (p *Pipeline) GetStep(name string) *Step {
	for _, step := range p.Steps {
		if step.Name == name {
			return step
		}
	}
	return nil
}

type (
	// Step defines a Pipeline step.
	Step struct {
		Command      []string                       `json:"command,omitempty"`
		Commands     []string                       `json:"commands,omitempty"`
		Detach       bool                           `json:"detach,omitempty"`
		DependsOn    []string                       `json:"depends_on,omitempty" yaml:"depends_on"`
		Devices      []*VolumeDevice                `json:"devices,omitempty"`
		DNS          []string                       `json:"dns,omitempty"`
		DNSSearch    []string                       `json:"dns_search,omitempty" yaml:"dns_search"`
		Entrypoint   []string                       `json:"entrypoint,omitempty"`
		Environment  map[string]*manifest.Variable  `json:"environment,omitempty"`
		ExtraHosts   []string                       `json:"extra_hosts,omitempty" yaml:"extra_hosts"`
		Failure      string                         `json:"failure,omitempty"`
		Image        string                         `json:"image,omitempty"`
		MemLimit     manifest.BytesSize             `json:"mem_limit,omitempty" yaml:"mem_limit"`
		MemSwapLimit manifest.BytesSize             `json:"memswap_limit,omitempty" yaml:"memswap_limit"`
		Network      string                         `json:"network_mode,omitempty" yaml:"network_mode"`
		Name         string                         `json:"name,omitempty"`
		Privileged   bool                           `json:"privileged,omitempty"`
		Pull         string                         `json:"pull,omitempty"`
		Settings     map[string]*manifest.Parameter `json:"settings,omitempty"`
		Shell        string                         `json:"shell,omitempty"`
		ShmSize      manifest.BytesSize             `json:"shm_size,omitempty" yaml:"shm_size"`
		User         string                         `json:"user,omitempty"`
		Volumes      []*VolumeMount                 `json:"volumes,omitempty"`
		When         manifest.Conditions            `json:"when,omitempty"`
		WorkingDir   string                         `json:"working_dir,omitempty" yaml:"working_dir"`
	}

	// Volume that can be mounted by containers.
	Volume struct {
		Name     string          `json:"name,omitempty"`
		EmptyDir *VolumeEmptyDir `json:"temp,omitempty" yaml:"temp"`
		HostPath *VolumeHostPath `json:"host,omitempty" yaml:"host"`
	}

	// VolumeDevice describes a mapping of a raw block
	// device within a container.
	VolumeDevice struct {
		Name       string `json:"name,omitempty"`
		DevicePath string `json:"path,omitempty" yaml:"path"`
	}

	// VolumeMount describes a mounting of a Volume
	// within a container.
	VolumeMount struct {
		Name      string `json:"name,omitempty"`
		MountPath string `json:"path,omitempty" yaml:"path"`
	}

	// VolumeEmptyDir mounts a temporary directory from the
	// host node's filesystem into the container. This can
	// be used as a shared scratch space.
	VolumeEmptyDir struct {
		Medium    string             `json:"medium,omitempty"`
		SizeLimit manifest.BytesSize `json:"size_limit,omitempty" yaml:"size_limit"`
	}

	// VolumeHostPath mounts a file or directory from the
	// host node's filesystem into your container.
	VolumeHostPath struct {
		Path string `json:"path,omitempty"`
	}

	// Workspace represents the pipeline workspace configuration.
	Workspace struct {
		Base string `json:"base,omitempty"`
		Path string `json:"path,omitempty"`
	}
)
