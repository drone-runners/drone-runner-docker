// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package compiler

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/antonmedv/expr"
	"github.com/drone/drone-go/drone"
)

// helper function a when clause.
func evalif(code string, inputs map[string]interface{}) (bool, bool, error) {
	var evalFailure, evalAlways, onFailure, onSuccess bool

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

	// trim trailing whitespace
	code = strings.TrimSpace(code)

	// trim ${{ }} wrappers if they exist
	if strings.HasPrefix(code, "${{") {
		code = strings.TrimPrefix(code, "${{")
		code = strings.TrimSuffix(code, "}}")
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
	output1, err1 := expr.Run(program, builtins)
	if err1 != nil {
		return false, false, err1
	}

	// then evaluate the expression assuming
	// the pipeline is in a failing state.
	isSuccess = false
	output2, err2 := expr.Run(program, builtins)
	if err1 != nil {
		return false, false, err2
	}

	// if the user invoked always() we should
	// return true for both success and failure.
	if evalAlways {
		return true, true, nil
	}

	// coerce the outputs to bool
	onSuccess = coerceBool(output1)
	onFailure = coerceBool(output2)

	return onSuccess, (onFailure && evalFailure), nil
}

// expand string
func expand(code string, inputs map[string]interface{}) string {
	pattern := regexp.MustCompile(`\${{(.*)}}`)
	return pattern.ReplaceAllStringFunc(code, func(s string) string {
		s = strings.TrimSpace(s)
		s = strings.TrimPrefix(s, "${{")
		s = strings.TrimSuffix(s, "}}")
		out, _ := evalstr(s, inputs)
		return out
	})
}

// helper fucntion evaluates a string expression
func evalstr(code string, inputs map[string]interface{}) (string, error) {
	v, err := eval(code, inputs)
	return coerceString(v), err
}

// helper function evaluates an expression
func eval(code string, inputs map[string]interface{}) (interface{}, error) {
	builtins := map[string]interface{}{}

	// add input args to the predefined
	// variables list.
	for k, v := range inputs {
		builtins[k] = v
	}

	// compile the program
	program, err := expr.Compile(code, expr.Env(builtins))
	if err != nil {
		return nil, err
	}

	// evaluate the expression
	return expr.Run(program, builtins)
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

//
// coercion
//

// helper function to coerce a value to true
func coerceBool(v interface{}) bool {
	if v == nil || v == "<nil>" {
		return false
	}
	switch vv := v.(type) {
	case string:
		return vv != ""
	case int:
		return vv != 0
	case int8:
		return vv != 0
	case int16:
		return vv != 0
	case int32:
		return vv != 0
	case int64:
		return vv != 0
	case uint:
		return vv != 0
	case uint8:
		return vv != 0
	case uint16:
		return vv != 0
	case uint32:
		return vv != 0
	case uint64:
		return vv != 0
	case float32:
		return vv != 0
	case float64:
		return vv != 0
	case complex128:
		return vv != 0
	case complex64:
		return vv != 0
	case bool:
		return vv
	case []string:
		return len(vv) != 0
	case map[string]string:
		return len(vv) != 0
	case map[string]interface{}:
		return len(vv) != 0
	}
	return v != nil
}

// helper function to coerce a value to string
func coerceString(v interface{}) string {
	if v == nil || v == "<nil>" {
		return ""
	}
	switch vv := v.(type) {
	case string:
		return vv
	case int:
		return strconv.Itoa(vv)
	case int8:
		return fmt.Sprint(vv)
	case int16:
		return fmt.Sprint(vv)
	case int32:
		return fmt.Sprint(vv)
	case int64:
		return fmt.Sprint(vv)
	case uint:
		return fmt.Sprint(vv)
	case uint8:
		return fmt.Sprint(vv)
	case uint16:
		return fmt.Sprint(vv)
	case uint32:
		return fmt.Sprint(vv)
	case uint64:
		return fmt.Sprint(vv)
	case float32:
		return fmt.Sprint(vv)
	case float64:
		return fmt.Sprint(vv)
	case complex128:
		return fmt.Sprint(vv)
	case complex64:
		return fmt.Sprint(vv)
	case bool:
		return fmt.Sprint(vv)
	case []string:
		if len(vv) != 0 {
			return fmt.Sprint(vv)
		}
	case map[string]string:
		if len(vv) != 0 {
			return fmt.Sprint(vv)
		}
	case map[string]interface{}:
		if len(vv) != 0 {
			return fmt.Sprint(vv)
		}
	}
	return ""
}
