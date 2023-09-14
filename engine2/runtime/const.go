// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package runtime

// import (
// 	"bytes"
// 	"encoding/json"
// )

// //
// // run policy
// //

// // RunPolicy defines the policy for starting containers
// // based on the point-in-time pass or fail state of
// // the pipeline.
// type RunPolicy int

// // RunPolicy enumeration.
// const (
// 	RunOnSuccess RunPolicy = iota
// 	RunOnFailure
// 	RunAlways
// 	RunNever
// )

// func (r RunPolicy) String() string {
// 	return runPolicyID[r]
// }

// var runPolicyID = map[RunPolicy]string{
// 	RunOnSuccess: "on-success",
// 	RunOnFailure: "on-failure",
// 	RunAlways:    "always",
// 	RunNever:     "never",
// }

// var runPolicyName = map[string]RunPolicy{
// 	"":           RunOnSuccess,
// 	"on-success": RunOnSuccess,
// 	"on-failure": RunOnFailure,
// 	"always":     RunAlways,
// 	"never":      RunNever,
// }

// // MarshalJSON marshals the string representation of the
// // run type to JSON.
// func (r *RunPolicy) MarshalJSON() ([]byte, error) {
// 	buffer := bytes.NewBufferString(`"`)
// 	buffer.WriteString(runPolicyID[*r])
// 	buffer.WriteString(`"`)
// 	return buffer.Bytes(), nil
// }

// // UnmarshalJSON unmarshals the json representation of the
// // run type from a string value.
// func (r *RunPolicy) UnmarshalJSON(b []byte) error {
// 	// unmarshal as string
// 	var s string
// 	err := json.Unmarshal(b, &s)
// 	if err != nil {
// 		return err
// 	}
// 	// lookup value
// 	*r = runPolicyName[s]
// 	return nil
// }

// //
// // failure policy
// //

// // ErrPolicy defines the step error policy
// type ErrPolicy int

// // ErrPolicy enumeration.
// const (
// 	ErrFail ErrPolicy = iota
// 	ErrFailFast
// 	ErrIgnore
// )

// func (p ErrPolicy) String() string {
// 	return errPolicyID[p]
// }

// var errPolicyID = map[ErrPolicy]string{
// 	ErrFail:     "fail",
// 	ErrFailFast: "fail-fast",
// 	ErrIgnore:   "ignore",
// }

// var errPolicyName = map[string]ErrPolicy{
// 	"":          ErrFail,
// 	"fail":      ErrFail,
// 	"fail-fast": ErrFailFast,
// 	"fast":      ErrFailFast,
// 	"always":    ErrFail,
// 	"ignore":    ErrIgnore,
// }

// // MarshalJSON marshals the string representation of the
// // pull type to JSON.
// func (p *ErrPolicy) MarshalJSON() ([]byte, error) {
// 	buffer := bytes.NewBufferString(`"`)
// 	buffer.WriteString(errPolicyID[*p])
// 	buffer.WriteString(`"`)
// 	return buffer.Bytes(), nil
// }

// // UnmarshalJSON unmarshals the json representation of the
// // pull type from a string value.
// func (p *ErrPolicy) UnmarshalJSON(b []byte) error {
// 	// unmarshal as string
// 	var s string
// 	err := json.Unmarshal(b, &s)
// 	if err != nil {
// 		return err
// 	}
// 	// lookup value
// 	*p = errPolicyName[s]
// 	return nil
// }
