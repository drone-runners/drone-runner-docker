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
	"fmt"
	"strconv"
)

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
