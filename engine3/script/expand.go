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

	schema "github.com/bradrydzewski/spec/yaml"
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
func ExpandConfig(in *schema.Schema, inputs map[string]interface{}) {
	switch {
	case in.Pipeline != nil:
		v := in.Pipeline
		for _, vv := range v.Stages {
			ExpandStage(vv, inputs)
		}

		if clone := in.Pipeline.Clone; clone != nil {
			if ref := clone.Ref; ref != nil {
				ref.Name = Expand(ref.Name, inputs)
				ref.Sha = Expand(ref.Sha, inputs)
				ref.Type = Expand(ref.Type, inputs)
			}
		}

		// TODO
		// in.Pipeline.Env

		// TODO
		// in.Pipeline.Service

		// TODO
		// in.Pipeline.Environment

		if repo := in.Pipeline.Repo; repo != nil {
			repo.Connector = Expand(repo.Connector, inputs)
			repo.Name = Expand(repo.Name, inputs)
		}

	case in.Template != nil:
	case in.Environment != nil:
	case in.Service != nil:
	case in.Inputset != nil:
	}
}

// ExpandStage expands scripts in the stage.
func ExpandStage(stage *schema.Stage, inputs map[string]interface{}) {

	// we copy template inputs and matrix inputs to the `inputs`
	// map. ideally we are making a copy instead of mutating directly
	if stage.Context != nil {
		if _, ok := inputs["matrix"]; !ok {
			inputs["matrix"] = map[string]interface{}{}
		}
		if _, ok := inputs["inputs"]; !ok {
			inputs["inputs"] = map[string]interface{}{}
		}
		for k, v := range stage.Context.Matrix { // TODO fix dangerous
			inputs["matrix"].(map[string]interface{})[k] = v
		}
		for k, v := range stage.Context.Inputs { // TODO fix dangerous
			if s, isstr := v.(string); isstr {
				// if expressions are used in the input
				// we should resolve before adding them
				inputs["inputs"].(map[string]interface{})[k] = Expand(s, inputs)
			} else {
				inputs["inputs"].(map[string]interface{})[k] = v
			}

		}
	}

	stage.Id = Expand(stage.Id, inputs)
	stage.Name = Expand(stage.Name, inputs)
	stage.Delegate = Expand(stage.Delegate, inputs)

	// TODO
	if vv := stage.Platform; vv != nil {
		vv.Os = Expand(vv.Os, inputs)
		vv.Arch = Expand(vv.Arch, inputs)
	}

	// TODO
	if clone := stage.Clone; clone != nil {
		if ref := clone.Ref; ref != nil {
			ref.Name = Expand(ref.Name, inputs)
			ref.Sha = Expand(ref.Sha, inputs)
			ref.Type = Expand(ref.Type, inputs)
		}
	}

	if vv := stage.Cache; vv != nil {
		vv.Key = Expand(vv.Key, inputs)
		vv.Policy = Expand(vv.Key, inputs)
		// for i, s := range vv.Paths {
		// 	vv.Paths[i] = Expand(s, inputs)
		// }
		vv.Key = Expand(vv.Key, inputs)
	}

	for k, v := range stage.Env {
		stage.Env[k] = Expand(v, inputs)
	}
	for _, vv := range stage.Steps {
		ExpandStep(vv, inputs)
	}

	if stage.Group != nil {
		for _, vv := range stage.Group.Stages {
			ExpandStage(vv, inputs)
		}
	}
	if stage.Parallel != nil {
		for _, vv := range stage.Group.Stages {
			ExpandStage(vv, inputs)
		}
	}
}

// EexpandStep expands scripts in the step.
func ExpandStep(step *schema.Step, inputs map[string]interface{}) {

	// we copy template inputs and matrix inputs to the `inputs`
	// map. ideally we are making a copy instead of mutating directly
	if step.Context != nil {
		if _, ok := inputs["matrix"]; !ok {
			inputs["matrix"] = map[string]interface{}{}
		}
		if _, ok := inputs["inputs"]; !ok {
			inputs["inputs"] = map[string]interface{}{}
		}
		for k, v := range step.Context.Matrix { // TODO fix dangerous
			inputs["matrix"].(map[string]interface{})[k] = v
		}
		for k, v := range step.Context.Inputs { // TODO fix dangerous
			if s, isstr := v.(string); isstr {
				// if expressions are used in the input
				// we should resolve before adding them
				inputs["inputs"].(map[string]interface{})[k] = Expand(s, inputs)
			} else {
				inputs["inputs"].(map[string]interface{})[k] = v
			}

		}
	}

	step.Id = Expand(step.Id, inputs)
	step.Name = Expand(step.Name, inputs)

	if v := step.Background; v != nil {
		for i, s := range v.Script {
			v.Script[i] = Expand(s, inputs)
		}
		for k, s := range v.Env {
			v.Env[k] = Expand(s, inputs)
		}
		if v.Container != nil {
			v.Container.Image = Expand(v.Container.Image, inputs)

			for i, s := range v.Container.Entrypoint {
				v.Container.Entrypoint[i] = Expand(s, inputs)
			}
			for i, s := range v.Container.Args {
				v.Container.Args[i] = Expand(s, inputs)
			}
			if v.Container.Credentials != nil {
				v.Container.Credentials.Username = Expand(v.Container.Credentials.Username, inputs)
				v.Container.Credentials.Password = Expand(v.Container.Credentials.Username, inputs)
			}
		}
	}

	if v := step.Run; v != nil {
		for i, s := range v.Script {
			v.Script[i] = Expand(s, inputs)
		}
		for k, s := range v.Env {
			v.Env[k] = Expand(s, inputs)
		}
		if v.Container != nil {
			v.Container.Image = Expand(v.Container.Image, inputs)

			for i, s := range v.Container.Entrypoint {
				v.Container.Entrypoint[i] = Expand(s, inputs)
			}
			for i, s := range v.Container.Args {
				v.Container.Args[i] = Expand(s, inputs)
			}
			if v.Container.Credentials != nil {
				v.Container.Credentials.Username = Expand(v.Container.Credentials.Username, inputs)
				v.Container.Credentials.Password = Expand(v.Container.Credentials.Username, inputs)
			}
			if v.Report != nil {
				// for _, report := range v.Report {
				// 	for i, s := range report.Path {
				// 		report.Path[i] = Expand(s, inputs)
				// 	}
				// }
			}
		}
	}

	if v := step.Group; v != nil {
		for _, vv := range v.Steps {
			ExpandStep(vv, inputs)
		}
	}

	if v := step.Parallel; v != nil {
		for _, vv := range v.Steps {
			ExpandStep(vv, inputs)
		}
	}
}
