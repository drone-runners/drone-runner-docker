// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package compiler

import (
	"github.com/antonmedv/expr"
	"github.com/drone/drone-go/drone"
)

// helper to eval whether a starlark script
// evaluates to true.
func evalif(code string, inputs map[string]interface{}) (bool, bool, error) {
	var evalFailure, evalAlways bool

	var isSuccess bool
	builtins := map[string]interface{}{
		"success": func() bool {
			return isSuccess
		},
		"failure": func() bool {
			evalFailure = true
			return !isSuccess
		},
		"always": func() bool {
			evalAlways = true
			return true
		},
	}

	// add input args to the predefined
	// variables list.
	for k, v := range inputs {
		builtins[k] = v
	}

	// compile the program
	program, err := expr.Compile(code, expr.Env(builtins))
	if err != nil {
		return false, false, err
	}

	// first evaluate the expression assuming
	// the pipeline is in a passing state.
	isSuccess = true
	onSuccess, err1 := expr.Run(program, builtins)
	if err1 != nil {
		return false, false, err1
	}

	// then evaluate the expression assuming
	// the pipeline is in a failing state.
	isSuccess = false
	onFailure, err2 := expr.Run(program, builtins)
	if err1 != nil {
		return false, false, err2
	}

	// if the user invoked always() we should
	// return true for both success and failure.
	if evalAlways {
		return true, true, nil
	}

	return onSuccess == true, (onFailure == true && evalFailure), nil
}

func fromBuild(v *drone.Build) map[string]interface{} {
	return map[string]interface{}{
		"event":         v.Event,
		"action":        v.Action,
		"cron":          v.Cron,
		"environment":   v.Deploy,
		"link":          v.Link,
		"branch":        v.Target,
		"source":        v.Source,
		"before":        v.Before,
		"after":         v.After,
		"target":        v.Target,
		"ref":           v.Ref,
		"commit":        v.After,
		"title":         v.Title,
		"message":       v.Message,
		"source_repo":   v.Fork,
		"author_login":  v.Author,
		"author_name":   v.AuthorName,
		"author_email":  v.AuthorEmail,
		"author_avatar": v.AuthorAvatar,
		"sender":        v.Sender,
		"debug":         v.Debug,
		"params":        v.Params,
	}
}

func fromRepo(v *drone.Repo) map[string]interface{} {
	return map[string]interface{}{
		"uid":                  v.UID,
		"name":                 v.Name,
		"namespace":            v.Namespace,
		"slug":                 v.Slug,
		"git_http_url":         v.HTTPURL,
		"git_ssh_url":          v.SSHURL,
		"link":                 v.Link,
		"branch":               v.Branch,
		"config":               v.Config,
		"private":              v.Private,
		"visibility":           v.Visibility,
		"active":               v.Active,
		"trusted":              v.Trusted,
		"protected":            v.Protected,
		"ignore_forks":         v.IgnoreForks,
		"ignore_pull_requests": v.IgnorePulls,
	}
}
