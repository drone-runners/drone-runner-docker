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
	"testing"
)

func TestEvalTrue(t *testing.T) {
	var tests = []struct {
		onsuccess bool
		onfailure bool
		expr      string
		data      map[string]interface{}
	}{
		{
			expr:      `branch == "main"`,
			data:      map[string]interface{}{"branch": "main"},
			onsuccess: true,
			onfailure: false,
		},
		{
			expr:      `${{ branch == "main" }}`,
			data:      map[string]interface{}{"branch": "main"},
			onsuccess: true,
			onfailure: false,
		},
		{
			expr:      `branch != "main"`,
			data:      map[string]interface{}{"branch": "main"},
			onsuccess: false,
			onfailure: false,
		},
		{
			expr:      `pipeline.branch == "main"`, // field does not exist
			data:      map[string]interface{}{"pipeline": map[string]interface{}{}},
			onsuccess: false,
			onfailure: false,
		},
		{
			expr:      `branch != "main" || failure()`,
			data:      map[string]interface{}{"branch": "main"},
			onsuccess: false,
			onfailure: true,
		},
		{
			expr:      `always()`,
			data:      map[string]interface{}{"branch": "main"},
			onsuccess: true,
			onfailure: true,
		},
		{
			expr:      `failure()`,
			data:      map[string]interface{}{"branch": "main"},
			onsuccess: false,
			onfailure: true,
		},
		// returns non-empty string, which is coerced to true
		{
			expr:      `branch`,
			data:      map[string]interface{}{"branch": "main"},
			onsuccess: true,
			onfailure: false,
		},
		// returns empty string, which is coerced to false
		{
			expr:      `branch`,
			data:      map[string]interface{}{"branch": ""},
			onsuccess: false,
			onfailure: false,
		},
	}

	for _, test := range tests {
		t.Run(test.expr, func(t *testing.T) {
			t.Log(test.expr)
			onsuccess, onfailure, err := EvalWhen(test.expr, test.data)
			if err != nil {
				t.Error(err)
			}
			if got, want := onsuccess, test.onsuccess; got != want {
				t.Errorf("want on_success %v, got %v", want, got)
			}
			if got, want := onfailure, test.onfailure; got != want {
				t.Errorf("want on_failure %v, got %v", want, got)
			}
		})
	}
}
