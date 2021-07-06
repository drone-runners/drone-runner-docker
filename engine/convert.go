// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package engine

import (
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
)

// returns a container configuration.
func toConfig(spec *Spec, step *Step) *container.Config {
	config := &container.Config{
		Image:        step.Image,
		Labels:       step.Labels,
		WorkingDir:   step.WorkingDir,
		User:         step.User,
		AttachStdin:  false,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          false,
		OpenStdin:    false,
		StdinOnce:    false,
		ArgsEscaped:  false,
	}

	if len(step.Envs) != 0 {
		config.Env = toEnv(step.Envs)
	}
	for _, sec := range step.Secrets {
		config.Env = append(config.Env, sec.Env+"="+string(sec.Data))
	}

	if len(step.Entrypoint) != 0 {
		config.Entrypoint = step.Entrypoint
	}
	if len(step.Command) != 0 {
		config.Cmd = step.Command
	}
	if len(step.Volumes) != 0 {
		config.Volumes = toVolumeSet(spec, step)
	}
	return config
}

// returns a container host configuration.
func toHostConfig(spec *Spec, step *Step) *container.HostConfig {
	config := &container.HostConfig{
		LogConfig: container.LogConfig{
			Type: "json-file",
		},
		Privileged: step.Privileged,
		ShmSize:    step.ShmSize,
	}
	// windows does not support privileged so we hard-code
	// this value to false.
	if spec.Platform.OS == "windows" {
		config.Privileged = false
	}
	if len(step.Network) > 0 {
		config.NetworkMode = container.NetworkMode(step.Network)
	}
	if len(step.DNS) > 0 {
		config.DNS = step.DNS
	}
	if len(step.DNSSearch) > 0 {
		config.DNSSearch = step.DNSSearch
	}
	if len(step.ExtraHosts) > 0 {
		config.ExtraHosts = step.ExtraHosts
	}
	if isUnlimited(step) == false {
		config.Resources = container.Resources{
			CPUPeriod:  step.CPUPeriod,
			CPUQuota:   step.CPUQuota,
			CpusetCpus: strings.Join(step.CPUSet, ","),
			CPUShares:  step.CPUShares,
			Memory:     step.MemLimit,
			MemorySwap: step.MemSwapLimit,
		}
	}

	if len(step.Devices) != 0 {
		config.Devices = toDeviceSlice(spec, step)
	}
	if len(step.Volumes) != 0 {
		config.Binds = toVolumeSlice(spec, step)
		config.Mounts = toVolumeMounts(spec, step)
	}
	return config
}

// helper function returns the container network configuration.
func toNetConfig(spec *Spec, proc *Step) *network.NetworkingConfig {
	// if the user overrides the default network we do not
	// attach to the user-defined network.
	if proc.Network != "" {
		return &network.NetworkingConfig{}
	}
	endpoints := map[string]*network.EndpointSettings{}
	endpoints[spec.Network.ID] = &network.EndpointSettings{
		NetworkID: spec.Network.ID,
		Aliases:   []string{proc.Name},
	}
	return &network.NetworkingConfig{
		EndpointsConfig: endpoints,
	}
}

// helper function that converts a slice of device paths to a slice of
// container.DeviceMapping.
func toDeviceSlice(spec *Spec, step *Step) []container.DeviceMapping {
	var to []container.DeviceMapping
	for _, mount := range step.Devices {
		device, ok := lookupVolume(spec, mount.Name)
		if !ok {
			continue
		}
		if isDevice(device) == false {
			continue
		}
		to = append(to, container.DeviceMapping{
			PathOnHost:        device.HostPath.Path,
			PathInContainer:   mount.DevicePath,
			CgroupPermissions: "rwm",
		})
	}
	if len(to) == 0 {
		return nil
	}
	return to
}

// helper function that converts a slice of volume paths to a set
// of unique volume names.
func toVolumeSet(spec *Spec, step *Step) map[string]struct{} {
	set := map[string]struct{}{}
	for _, mount := range step.Volumes {
		volume, ok := lookupVolume(spec, mount.Name)
		if !ok {
			continue
		}
		if isDevice(volume) {
			continue
		}
		if isNamedPipe(volume) {
			continue
		}
		if isBindMount(volume) == false {
			continue
		}
		set[mount.Path] = struct{}{}
	}
	return set
}

// helper function returns a slice of volume mounts.
func toVolumeSlice(spec *Spec, step *Step) []string {
	// this entire function should be deprecated in
	// favor of toVolumeMounts, however, I am unable
	// to get it working with data volumes.
	var to []string
	for _, mount := range step.Volumes {
		volume, ok := lookupVolume(spec, mount.Name)
		if !ok {
			continue
		}
		if isDevice(volume) {
			continue
		}
		if isDataVolume(volume) {
			path := volume.EmptyDir.ID + ":" + mount.Path
			to = append(to, path)
		}
		if isBindMount(volume) {
			path := volume.HostPath.Path + ":" + mount.Path
			to = append(to, path)
		}
	}
	return to
}

// helper function returns a slice of docker mount
// configurations.
func toVolumeMounts(spec *Spec, step *Step) []mount.Mount {
	var mounts []mount.Mount
	for _, target := range step.Volumes {
		source, ok := lookupVolume(spec, target.Name)
		if !ok {
			continue
		}

		if isBindMount(source) && !isDevice(source) {
			continue
		}

		// HACK: this condition can be removed once
		// toVolumeSlice has been fully replaced. at this
		// time, I cannot figure out how to get mounts
		// working with data volumes :(
		if isDataVolume(source) {
			continue
		}
		mounts = append(mounts, toMount(source, target))
	}
	if len(mounts) == 0 {
		return nil
	}
	return mounts
}

// helper function converts the volume declaration to a
// docker mount structure.
func toMount(source *Volume, target *VolumeMount) mount.Mount {
	to := mount.Mount{
		Target: target.Path,
		Type:   toVolumeType(source),
	}
	if isBindMount(source) || isNamedPipe(source) {
		to.Source = source.HostPath.Path
		to.ReadOnly = source.HostPath.ReadOnly
	}
	if isTempfs(source) {
		to.TmpfsOptions = &mount.TmpfsOptions{
			SizeBytes: source.EmptyDir.SizeLimit,
			Mode:      0700,
		}
	}
	return to
}

// helper function returns the docker volume enumeration
// for the given volume.
func toVolumeType(from *Volume) mount.Type {
	switch {
	case isDataVolume(from):
		return mount.TypeVolume
	case isTempfs(from):
		return mount.TypeTmpfs
	case isNamedPipe(from):
		return mount.TypeNamedPipe
	default:
		return mount.TypeBind
	}
}

// helper function that converts a key value map of
// environment variables to a string slice in key=value
// format.
func toEnv(env map[string]string) []string {
	var envs []string
	for k, v := range env {
		if v != "" {
			envs = append(envs, k+"="+v)
		}
	}
	return envs
}

// returns true if the container has no resource limits.
func isUnlimited(res *Step) bool {
	return len(res.CPUSet) == 0 &&
		res.CPUPeriod == 0 &&
		res.CPUQuota == 0 &&
		res.CPUShares == 0 &&
		res.MemLimit == 0 &&
		res.MemSwapLimit == 0
}

// returns true if the volume is a bind mount.
func isBindMount(volume *Volume) bool {
	return volume.HostPath != nil
}

// returns true if the volume is in-memory.
func isTempfs(volume *Volume) bool {
	return volume.EmptyDir != nil && volume.EmptyDir.Medium == "memory"
}

// returns true if the volume is a data-volume.
func isDataVolume(volume *Volume) bool {
	return volume.EmptyDir != nil && volume.EmptyDir.Medium != "memory"
}

// returns true if the volume is a device
func isDevice(volume *Volume) bool {
	return volume.HostPath != nil && strings.HasPrefix(volume.HostPath.Path, "/dev/")
}

// returns true if the volume is a named pipe.
func isNamedPipe(volume *Volume) bool {
	return volume.HostPath != nil &&
		strings.HasPrefix(volume.HostPath.Path, `\\.\pipe\`)
}

// helper function returns the named volume.
func lookupVolume(spec *Spec, name string) (*Volume, bool) {
	for _, v := range spec.Volumes {
		if v.HostPath != nil && v.HostPath.Name == name {
			return v, true
		}
		if v.EmptyDir != nil && v.EmptyDir.Name == name {
			return v, true
		}
	}
	return nil, false
}
