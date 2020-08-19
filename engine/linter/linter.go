// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package linter

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/drone-runners/drone-runner-docker/engine/resource"
	"github.com/drone/drone-go/drone"
	"github.com/drone/runner-go/manifest"
)

// ErrDuplicateStepName is returned when two Pipeline steps
// have the same name.
var ErrDuplicateStepName = errors.New("linter: duplicate step names")

// Opts provides linting options.
type Opts struct {
	Trusted bool
}

// Linter evaluates the pipeline against a set of
// rules and returns an error if one or more of the
// rules are broken.
type Linter struct{}

// New returns a new Linter.
func New() *Linter {
	return new(Linter)
}

// Lint executes the linting rules for the pipeline
// configuration.
func (l *Linter) Lint(pipeline manifest.Resource, repo *drone.Repo) error {
	return checkPipeline(pipeline.(*resource.Pipeline), repo.Trusted)
}

func checkPipeline(pipeline *resource.Pipeline, trusted bool) error {
	// if err := checkNames(pipeline); err != nil {
	// 	return err
	// }
	if err := checkSteps(pipeline, trusted); err != nil {
		return err
	}
	if err := checkVolumes(pipeline, trusted); err != nil {
		return err
	}
	return nil
}

// func checkNames(pipeline *resource.Pipeline) error {
// 	names := map[string]struct{}{}
// 	if !pipeline.Clone.Disable {
// 		names["clone"] = struct{}{}
// 	}
// 	steps := append(pipeline.Services, pipeline.Steps...)
// 	for _, step := range steps {
// 		_, ok := names[step.Name]
// 		if ok {
// 			return ErrDuplicateStepName
// 		}
// 		names[step.Name] = struct{}{}
// 	}
// 	return nil
// }

func checkSteps(pipeline *resource.Pipeline, trusted bool) error {
	steps := append(pipeline.Services, pipeline.Steps...)
	names := map[string]struct{}{}
	if !pipeline.Clone.Disable {
		names["clone"] = struct{}{}
	}
	for _, step := range steps {
		if step == nil {
			return errors.New("linter: nil step")
		}

		// unique list of names
		_, ok := names[step.Name]
		if ok {
			return ErrDuplicateStepName
		}
		names[step.Name] = struct{}{}

		if err := checkStep(step, trusted); err != nil {
			return err
		}
		if err := checkDeps(step, names); err != nil {
			return err
		}
	}
	return nil
}

func checkStep(step *resource.Step, trusted bool) error {
	if step.Image == "" {
		return errors.New("linter: invalid or missing image")
	}
	// if step.Name == "" {
	// 	return errors.New("linter: invalid or missing name")
	// }
	// if len(step.Name) > 100 {
	// 	return errors.New("linter: name exceeds maximum length")
	// }
	if trusted == false && step.Privileged {
		return errors.New("linter: untrusted repositories cannot enable privileged mode")
	}
	if trusted == false && len(step.Devices) > 0 {
		return errors.New("linter: untrusted repositories cannot mount devices")
	}
	if trusted == false && len(step.DNS) > 0 {
		return errors.New("linter: untrusted repositories cannot configure dns")
	}
	if trusted == false && len(step.DNSSearch) > 0 {
		return errors.New("linter: untrusted repositories cannot configure dns_search")
	}
	if trusted == false && len(step.ExtraHosts) > 0 {
		return errors.New("linter: untrusted repositories cannot configure extra_hosts")
	}
	if trusted == false && len(step.Network) > 0 {
		return errors.New("linter: untrusted repositories cannot configure network_mode")
	}
	if trusted == false && int(step.ShmSize) > 0 {
		return errors.New("linter: untrusted repositories cannot configure shm_size")
	}
	for _, mount := range step.Volumes {
		switch mount.Name {
		case "workspace", "_workspace", "_docker_socket":
			return fmt.Errorf("linter: invalid volume name: %s", mount.Name)
		}
		if strings.HasPrefix(filepath.Clean(mount.MountPath), "/run/drone") {
			return fmt.Errorf("linter: cannot mount volume at /run/drone")
		}
	}
	return nil
}

func checkVolumes(pipeline *resource.Pipeline, trusted bool) error {
	for _, volume := range pipeline.Volumes {
		if volume.EmptyDir != nil {
			err := checkEmptyDirVolume(volume.EmptyDir, trusted)
			if err != nil {
				return err
			}
		}
		if volume.HostPath != nil {
			err := checkHostPathVolume(volume.HostPath, trusted)
			if err != nil {
				return err
			}
		}
		switch volume.Name {
		case "":
			return fmt.Errorf("linter: missing volume name")
		case "workspace", "_workspace", "_docker_socket":
			return fmt.Errorf("linter: invalid volume name: %s", volume.Name)
		}
	}
	return nil
}

func checkHostPathVolume(volume *resource.VolumeHostPath, trusted bool) error {
	if trusted == false {
		return errors.New("linter: untrusted repositories cannot mount host volumes")
	}
	return nil
}

func checkEmptyDirVolume(volume *resource.VolumeEmptyDir, trusted bool) error {
	if trusted == false && volume.Medium == "memory" {
		return errors.New("linter: untrusted repositories cannot mount in-memory volumes")
	}
	return nil
}

func checkDeps(step *resource.Step, deps map[string]struct{}) error {
	for _, dep := range step.DependsOn {
		_, ok := deps[dep]
		if !ok {
			return fmt.Errorf("linter: unknown step dependency detected: %s references %s", step.Name, dep)
		}
		if step.Name == dep {
			return fmt.Errorf("linter: cyclical step dependency detected: %s", dep)
		}
	}
	return nil
}
