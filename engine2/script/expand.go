// Copyright 2022 Harness, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package script

import (
	"regexp"
	"strings"

	schema "github.com/drone/spec/dist/go"
)

var pattern = regexp.MustCompile(`\${{(.*?)}}`)

// Expand expands the script inside the text snippet.
func Expand(code string, inputs map[string]interface{}) string {
	if !strings.Contains(code, "${{") {
		return code
	}
	return pattern.ReplaceAllStringFunc(code, func(s string) string {
		s = strings.TrimSpace(s)
		s = strings.TrimPrefix(s, "${{")
		s = strings.TrimSuffix(s, "}}")
		out, _ := EvalStr(s, inputs)
		return out
	})
}

// ExpandConfig expands scripts in the stage.
func ExpandConfig(in *schema.Config, inputs map[string]interface{}) {
	switch v := in.Spec.(type) {
	case *schema.Pipeline:
		for _, vv := range v.Stages {
			ExpandStage(vv, inputs)
		}
	case *schema.PluginStep:
	case *schema.PluginStage:
	case *schema.TemplateStage:
	case *schema.TemplateStep:
	}
}

// ExpandStage expands scripts in the stage.
func ExpandStage(stage *schema.Stage, inputs map[string]interface{}) {
	stage.Id = Expand(stage.Id, inputs)
	stage.Name = Expand(stage.Name, inputs)
	for i, s := range stage.Delegate {
		stage.Delegate[i] = Expand(s, inputs)
	}

	if stage.Strategy != nil && stage.Strategy.Spec != nil {
		if matrix, ok := stage.Strategy.Spec.(*schema.Matrix); ok {
			for _, axis := range matrix.Include {
				// FIXME this prevents matrix inside a matrix
				// FIXME ideally we aren't mutating the inputs here and make a copy instead
				inputs["matrix"] = axis
			}
		}
	}

	// FIXME ideally we aren't mutating the inputs and are making a copy instead
	for k, v := range stage.Inputs {
		m, ok := inputs["inputs"]
		if ok {
			mm, ok := m.(map[string]interface{})
			if ok {
				mm[k] = v
			}
		}
	}

	switch spec := stage.Spec.(type) {
	case *schema.StageCI:

		if vv := spec.Runtime; vv != nil {
			switch vv.Spec.(type) {
			case *schema.RuntimeCloud: // TODO
			case *schema.RuntimeMachine: // TODO
			case *schema.RuntimeVM: // TODO
			case *schema.RuntimeKube: // TODO
			}
		}

		// TODO
		if vv := spec.Platform; vv != nil {
			// vv.Os
			// vv.Arch
		}

		// TODO
		if vv := spec.Clone; vv != nil {
			// TODO
		}

		if vv := spec.Cache; vv != nil {
			vv.Key = Expand(vv.Key, inputs)
			vv.Policy = Expand(vv.Key, inputs)
			for i, s := range vv.Paths {
				vv.Paths[i] = Expand(s, inputs)
			}
			vv.Key = Expand(vv.Key, inputs)
		}

		for k, v := range spec.Envs {
			spec.Envs[k] = Expand(v, inputs)
		}
		for _, vv := range spec.Steps {
			ExpandStep(vv, inputs)
		}
	case *schema.StageGroup:
		for _, vv := range spec.Stages {
			ExpandStage(vv, inputs)
		}
	case *schema.StageParallel:
		for _, vv := range spec.Stages {
			ExpandStage(vv, inputs)
		}
	}
}

// EexpandStep expands scripts in the step.
func ExpandStep(step *schema.Step, inputs map[string]interface{}) {
	step.Id = Expand(step.Id, inputs)
	step.Name = Expand(step.Name, inputs)

	if step.Strategy != nil && step.Strategy.Spec != nil {
		if matrix, ok := step.Strategy.Spec.(*schema.Matrix); ok {
			for _, axis := range matrix.Include {
				// FIXME this prevents matrix inside a matrix
				// FIXME ideally we aren't mutating the inputs here and make a copy instead
				inputs["matrix"] = axis
			}
		}
	}

	// FIXME ideally we aren't mutating the inputs and are making a copy instead
	for k, v := range step.Inputs {
		m, ok := inputs["inputs"]
		if ok {
			mm, ok := m.(map[string]interface{})
			if ok {
				if s, isstr := v.(string); isstr {
					// if the pipeline defines a template or plugin,
					// and uses an expression as the input, expand now
					mm[k] = Expand(s, inputs)
				} else {
					mm[k] = v
				}
			}
		}
	}

	switch spec := step.Spec.(type) {
	case *schema.StepAction:
	case *schema.StepBackground:
		spec.Run = Expand(spec.Run, inputs)
		spec.Image = Expand(spec.Image, inputs)
		spec.Entrypoint = Expand(spec.Entrypoint, inputs)
		for i, s := range spec.Args {
			spec.Args[i] = Expand(s, inputs)
		}
	case *schema.StepBitrise:
	case *schema.StepExec:
		spec.Run = Expand(spec.Run, inputs)
		spec.Image = Expand(spec.Image, inputs)
		spec.Connector = Expand(spec.Connector, inputs)
		spec.Entrypoint = Expand(spec.Entrypoint, inputs)
		for i, s := range spec.Args {
			spec.Args[i] = Expand(s, inputs)
		}
		for k, v := range spec.Envs {
			spec.Envs[k] = Expand(v, inputs)
		}
		if spec.Reports != nil {
			for _, report := range spec.Reports {
				for i, s := range report.Path {
					report.Path[i] = Expand(s, inputs)
				}
			}
		}

	case *schema.StepGroup:
		for _, vv := range spec.Steps {
			ExpandStep(vv, inputs)
		}
	case *schema.StepParallel:
		for _, vv := range spec.Steps {
			ExpandStep(vv, inputs)
		}
	case *schema.StepRun:
		for i, s := range spec.Script {
			spec.Script[i] = Expand(s, inputs)
		}
		for k, v := range spec.Envs {
			spec.Envs[k] = Expand(v, inputs)
		}
		if spec.Reports != nil {
			for _, report := range spec.Reports {
				for i, s := range report.Path {
					report.Path[i] = Expand(s, inputs)
				}
			}
		}
		if container := spec.Container; container != nil {
			container.Image = Expand(container.Image, inputs)
			container.Connector = Expand(container.Connector, inputs)
			container.Entrypoint = Expand(container.Entrypoint, inputs)
			for i, s := range container.Args {
				container.Args[i] = Expand(s, inputs)
			}
		}
	case *schema.StepPlugin:
		// NOTE
		// this case should never happen. we find and replace
		// all plugins with the plugin steps.
	case *schema.StepTemplate:
		// NOTE
		// this case should never happen. we find and replace
		// all templates with the template steps.
	case *schema.StepTest:
	}
}
