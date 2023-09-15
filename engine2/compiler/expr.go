// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package compiler

import (
	"github.com/drone/drone-go/drone"
	"go.starlark.net/starlark"
)

// helper to eval whether a starlark script
// evaluates to true.
func evalif(expr string, inputs starlark.StringDict) (bool, bool, error) {
	var onFailure, onSuccess bool
	var evalFailure, evalAlways bool

	thread1 := &starlark.Thread{
		Name: "when",
		Print: func(thread *starlark.Thread, msg string) {
			// capture the expression result and
			// test if truthy.
			switch msg {
			case "True", "true", "1":
				onSuccess = true
			}
		},
	}

	thread2 := &starlark.Thread{
		Name: "when",
		Print: func(thread *starlark.Thread, msg string) {
			// capture the expression result and
			// test if truthy.
			switch msg {
			case "True", "true", "1":
				onFailure = true
			}
		},
	}

	var isSuccess bool
	predeclared := starlark.StringDict{
		"success": starlark.NewBuiltin("success", func(*starlark.Thread, *starlark.Builtin, starlark.Tuple, []starlark.Tuple) (starlark.Value, error) {
			return starlark.Bool(isSuccess), nil
		}),
		"failure": starlark.NewBuiltin("failure", func(*starlark.Thread, *starlark.Builtin, starlark.Tuple, []starlark.Tuple) (starlark.Value, error) {
			evalFailure = true
			return starlark.Bool(!isSuccess), nil
		}),
		"always": starlark.NewBuiltin("always", func(*starlark.Thread, *starlark.Builtin, starlark.Tuple, []starlark.Tuple) (starlark.Value, error) {
			evalAlways = true
			return starlark.Bool(true), nil
		}),
	}

	// add input args to the predefined
	// variables list.
	for k, v := range inputs {
		predeclared[k] = v
	}

	// first evaluate the expression assuming
	// the pipeline is in a passing state.
	isSuccess = true
	_, err1 := starlark.ExecFile(thread1, "when.star", "print("+expr+")", predeclared)
	if err1 != nil {
		return false, false, err1
	}

	// then evaluate the expression assuming
	// the pipeline is in a failing state.
	isSuccess = false
	_, err2 := starlark.ExecFile(thread2, "when.star", "print("+expr+")", predeclared)
	if err2 != nil {
		return false, false, err2
	}
	return onSuccess || evalAlways, (onFailure && evalFailure) || evalAlways, nil
}

func fromBuild(v *drone.Build) starlark.StringDict {
	return starlark.StringDict{
		"event":         starlark.String(v.Event),
		"action":        starlark.String(v.Action),
		"cron":          starlark.String(v.Cron),
		"environment":   starlark.String(v.Deploy),
		"link":          starlark.String(v.Link),
		"branch":        starlark.String(v.Target),
		"source":        starlark.String(v.Source),
		"before":        starlark.String(v.Before),
		"after":         starlark.String(v.After),
		"target":        starlark.String(v.Target),
		"ref":           starlark.String(v.Ref),
		"commit":        starlark.String(v.After),
		"title":         starlark.String(v.Title),
		"message":       starlark.String(v.Message),
		"source_repo":   starlark.String(v.Fork),
		"author_login":  starlark.String(v.Author),
		"author_name":   starlark.String(v.AuthorName),
		"author_email":  starlark.String(v.AuthorEmail),
		"author_avatar": starlark.String(v.AuthorAvatar),
		"sender":        starlark.String(v.Sender),
		"debug":         starlark.Bool(v.Debug),
		"params":        fromMap(v.Params),
	}
}

func fromRepo(v *drone.Repo) starlark.StringDict {
	return starlark.StringDict{
		"uid":                  starlark.String(v.UID),
		"name":                 starlark.String(v.Name),
		"namespace":            starlark.String(v.Namespace),
		"slug":                 starlark.String(v.Slug),
		"git_http_url":         starlark.String(v.HTTPURL),
		"git_ssh_url":          starlark.String(v.SSHURL),
		"link":                 starlark.String(v.Link),
		"branch":               starlark.String(v.Branch),
		"config":               starlark.String(v.Config),
		"private":              starlark.Bool(v.Private),
		"visibility":           starlark.String(v.Visibility),
		"active":               starlark.Bool(v.Active),
		"trusted":              starlark.Bool(v.Trusted),
		"protected":            starlark.Bool(v.Protected),
		"ignore_forks":         starlark.Bool(v.IgnoreForks),
		"ignore_pull_requests": starlark.Bool(v.IgnorePulls),
	}
}

func fromInputs(m map[string]string) starlark.StringDict {
	out := map[string]starlark.Value{}
	for k, v := range m {
		out[k] = starlark.String(v)
	}
	return out
}

func fromMap(m map[string]string) *starlark.Dict {
	dict := new(starlark.Dict)
	for k, v := range m {
		dict.SetKey(
			starlark.String(k),
			starlark.String(v),
		)
	}
	return dict
}
