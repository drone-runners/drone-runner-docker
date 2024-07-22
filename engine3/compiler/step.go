// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package compiler

import (
	harness "github.com/bradrydzewski/spec/yaml"
	"github.com/drone-runners/drone-runner-docker/engine3/engine"
)

type state struct {
	// options   *harness.Default
	platform *harness.Platform
	// resources *harness.Resources
	labels map[string]string
	envs   map[string]string
}

func convertStep(stage *harness.Stage, step *harness.Step) []*engine.Step {

	switch {
	case step.Run != nil:
		dst := createRunStep(step, step.Run)
		dst.WorkingDir = "/gitness"
		if len(step.Run.Script) > 0 {
			setupScript(dst, step.Run.Script[0], "linux")
		}
		return []*engine.Step{dst}
	case step.Background != nil:
		dst := createRunStep(step, step.Background)
		dst.Detach = true
		if len(step.Background.Script) > 0 {
			setupScript(dst, step.Background.Script[0], "linux")
		}
		return []*engine.Step{dst}
	case step.Parallel != nil:
		var steps []*engine.Step
		for _, vv := range step.Parallel.Steps {
			steps = append(steps, convertStep(stage, vv)...)
		}
		return steps
	case step.Group != nil:
		var steps []*engine.Step
		for _, vv := range step.Group.Steps {
			steps = append(steps, convertStep(stage, vv)...)
		}
		return steps
		// case *harness.StepPlugin:
		// 	dst := createStepPlugin(step, v)
		// 	return []*engine.Step{dst}
		// case *harness.StepTemplate:
		// 	dst := createStepTemplate(step, v)
		// 	return []*engine.Step{dst}
	}
	return nil
}
