// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package compiler

import (
	"testing"

	"github.com/drone/drone-go/drone"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

func TestEvalTrue(t *testing.T) {
	expr := `failure() or build.branch == "main"`
	repo := &drone.Repo{}
	build := &drone.Build{Source: "main", Target: "main"}

	inputs := starlark.StringDict{
		"repo":   starlarkstruct.FromStringDict(starlark.String("repo"), fromRepo(repo)),
		"build":  starlarkstruct.FromStringDict(starlark.String("build"), fromBuild(build)),
		"inputs": starlarkstruct.FromStringDict(starlark.String("build"), fromInputs(map[string]string{"username": "bradrydzewski"})),
	}

	onSuccess, onFailure, err := evalif(expr, inputs)
	if err != nil {
		t.Error(err)
	}
	if got, want := onSuccess, true; got != want {
		t.Errorf("want on_success %v, got %v", want, got)
	}
	if got, want := onFailure, true; got != want {
		t.Errorf("want on_success %v, got %v", want, got)
	}
}
