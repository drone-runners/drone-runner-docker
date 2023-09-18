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
	"strings"

	"github.com/antonmedv/expr"
)

// Eval evaluates an expression.
func Eval(code string, inputs map[string]interface{}) (interface{}, error) {
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

// EvalStr evaluates an expression and returns the response
// in string format.
func EvalStr(code string, inputs map[string]interface{}) (string, error) {
	v, err := Eval(code, inputs)
	return coerceString(v), err
}

// EvalBool evaluates an expression and returns the response
// in string format.
func EvalBool(code string, inputs map[string]interface{}) (bool, error) {
	v, err := Eval(code, inputs)
	return coerceBool(v), err
}

// EvalWhen evalutes a when clause.
func EvalWhen(code string, inputs map[string]interface{}) (bool, bool, error) {
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
